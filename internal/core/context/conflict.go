// Package context handles loading and managing CRF context entities.
package context

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/carlosinfantes/cio/internal/core/decisions"
	"github.com/carlosinfantes/cio/internal/types"
)

// DetectConflicts checks for contradictions between CRF context and recent decisions.
func DetectConflicts(ctx *types.CRFContext) []types.ContextConflict {
	if ctx == nil {
		return nil
	}

	var conflicts []types.ContextConflict

	// Get recent decisions to check against
	recentDocs, err := decisions.ListDRFDocuments(nil)
	if err != nil || len(recentDocs) == 0 {
		return nil
	}

	// Limit to last 10 decisions
	if len(recentDocs) > 10 {
		recentDocs = recentDocs[:10]
	}

	// Get team size from CRF
	teamSize := 0
	for _, team := range ctx.GetTeams() {
		if headcount, ok := team.Attributes["headcount"].(int); ok {
			teamSize += headcount
		}
	}

	// Check team size conflicts
	if teamSize > 0 {
		for _, d := range recentDocs {
			searchText := d.Decision.Intent + " " + d.Synthesis.Decision
			mentionedSize := extractTeamSize(searchText)
			if mentionedSize > 0 && !withinThreshold(teamSize, mentionedSize, 0.2) {
				conflicts = append(conflicts, types.ContextConflict{
					Field:         "team_size",
					ContextValue:  fmt.Sprintf("%d engineers", teamSize),
					DecisionValue: fmt.Sprintf("%d mentioned", mentionedSize),
					DecisionID:    d.Decision.ID,
					Severity:      "warning",
				})
				break // Only report first conflict per field
			}
		}
	}

	// Check runway conflicts from facts
	var runwayMonths int
	for _, doc := range ctx.Facts {
		if factType, ok := doc.Entity.Attributes["fact_type"].(string); ok && factType == "constraint" {
			if strings.Contains(strings.ToLower(doc.Entity.Name), "runway") {
				if val, ok := doc.Entity.Attributes["value"].(int); ok {
					runwayMonths = val
					break
				}
			}
		}
	}

	if runwayMonths > 0 {
		for _, d := range recentDocs {
			searchText := d.Decision.Intent + " " + d.Synthesis.Decision
			mentionedRunway := extractRunway(searchText)
			if mentionedRunway > 0 && !withinThreshold(runwayMonths, mentionedRunway, 0.25) {
				conflicts = append(conflicts, types.ContextConflict{
					Field:         "runway",
					ContextValue:  fmt.Sprintf("%d months", runwayMonths),
					DecisionValue: fmt.Sprintf("%d months mentioned", mentionedRunway),
					DecisionID:    d.Decision.ID,
					Severity:      "warning",
				})
				break
			}
		}
	}

	// Check cloud provider conflicts from systems
	var cloudProvider string
	for _, doc := range ctx.Systems {
		if hosting, ok := doc.Entity.Attributes["hosting"].(string); ok && hosting != "" {
			cloudProvider = hosting
			break
		}
	}

	if cloudProvider != "" {
		for _, d := range recentDocs {
			searchText := d.Decision.Intent + " " + d.Synthesis.Decision
			mentionedCloud := extractCloudProvider(searchText)
			if mentionedCloud != "" && !strings.EqualFold(cloudProvider, mentionedCloud) {
				// Only flag if it seems like a current state, not a comparison
				if strings.Contains(strings.ToLower(d.Decision.Intent), "switched to") ||
					strings.Contains(strings.ToLower(d.Decision.Intent), "migrated to") ||
					strings.Contains(strings.ToLower(d.Decision.Intent), "now using") {
					conflicts = append(conflicts, types.ContextConflict{
						Field:         "cloud",
						ContextValue:  cloudProvider,
						DecisionValue: mentionedCloud,
						DecisionID:    d.Decision.ID,
						Severity:      "info",
					})
					break
				}
			}
		}
	}

	return conflicts
}

// extractTeamSize looks for team size mentions in text.
func extractTeamSize(text string) int {
	text = strings.ToLower(text)

	patterns := []string{
		`(\d+)\s*(?:engineers?|developers?|devs?)`,
		`team\s*(?:of|size|is)?\s*(\d+)`,
		`(\d+)\s*(?:person|people)\s*(?:team|engineering)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			if n, err := strconv.Atoi(matches[1]); err == nil && n > 0 && n < 10000 {
				return n
			}
		}
	}

	return 0
}

// extractRunway looks for runway mentions in text.
func extractRunway(text string) int {
	text = strings.ToLower(text)

	patterns := []string{
		`(\d+)\s*months?\s*(?:of\s*)?runway`,
		`runway\s*(?:of|is)?\s*(\d+)\s*months?`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			if n, err := strconv.Atoi(matches[1]); err == nil && n > 0 && n < 120 {
				return n
			}
		}
	}

	return 0
}

// extractCloudProvider looks for cloud provider mentions.
func extractCloudProvider(text string) string {
	text = strings.ToLower(text)

	providers := map[string][]string{
		"AWS":   {"aws", "amazon web services"},
		"GCP":   {"gcp", "google cloud", "google cloud platform"},
		"Azure": {"azure", "microsoft azure"},
	}

	for provider, keywords := range providers {
		for _, kw := range keywords {
			if strings.Contains(text, kw) {
				return provider
			}
		}
	}

	return ""
}

// withinThreshold checks if two values are within a percentage of each other.
func withinThreshold(a, b int, threshold float64) bool {
	if a == 0 || b == 0 {
		return true
	}
	diff := float64(a-b) / float64(a)
	if diff < 0 {
		diff = -diff
	}
	return diff <= threshold
}
