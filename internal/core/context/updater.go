// Package context handles loading and managing CRF context entities.
package context

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

// DetectUpdateSignals analyzes a question for signals that CRF context may have changed.
func DetectUpdateSignals(question string, ctx *types.CRFContext) []types.UpdateSuggestion {
	if ctx == nil {
		return nil
	}

	var suggestions []types.UpdateSuggestion
	questionLower := strings.ToLower(question)

	// Check for team size changes
	if teamSuggestion := detectTeamSizeChangeCRF(questionLower, ctx); teamSuggestion != nil {
		suggestions = append(suggestions, *teamSuggestion)
	}

	// Check for cloud provider changes
	if cloudSuggestion := detectCloudChangeCRF(questionLower, ctx); cloudSuggestion != nil {
		suggestions = append(suggestions, *cloudSuggestion)
	}

	// Check for runway changes
	if runwaySuggestion := detectRunwayChangeCRF(questionLower, ctx); runwaySuggestion != nil {
		suggestions = append(suggestions, *runwaySuggestion)
	}

	return suggestions
}

// detectTeamSizeChangeCRF looks for signals like "we now have X engineers"
func detectTeamSizeChangeCRF(question string, ctx *types.CRFContext) *types.UpdateSuggestion {
	patterns := []struct {
		regex   string
		trigger string
	}{
		{`(?:we|team)\s+(?:now|just|recently)\s+(?:have|hired|grew to|reached)\s+(\d+)\s*(?:engineers?|developers?|devs?|people)`, "you mentioned growing your team"},
		{`(?:grew|expanded|scaled)\s+(?:to|from)?\s*(\d+)\s*(?:engineers?|developers?|devs?)`, "you mentioned team growth"},
		{`(\d+)\s*(?:engineers?|developers?|devs?)\s+(?:now|currently)`, "you mentioned your current team size"},
	}

	// Get current team size from CRF
	currentTeamSize := 0
	var teamEntityID string
	for _, team := range ctx.GetTeams() {
		if headcount, ok := team.Attributes["headcount"].(int); ok {
			currentTeamSize += headcount
			if teamEntityID == "" {
				teamEntityID = team.ID
			}
		}
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.regex)
		matches := re.FindStringSubmatch(question)
		if len(matches) > 1 {
			newSize, err := strconv.Atoi(matches[1])
			if err != nil || newSize <= 0 || newSize > 10000 {
				continue
			}

			// Only suggest if significantly different
			if currentTeamSize > 0 && !withinThreshold(currentTeamSize, newSize, 0.1) {
				return &types.UpdateSuggestion{
					EntityID: teamEntityID,
					Field:    "headcount",
					OldValue: fmt.Sprintf("%d", currentTeamSize),
					NewValue: fmt.Sprintf("%d", newSize),
					Reason:   p.trigger,
				}
			}
		}
	}

	return nil
}

// detectCloudChangeCRF looks for signals like "we switched to AWS"
func detectCloudChangeCRF(question string, ctx *types.CRFContext) *types.UpdateSuggestion {
	triggers := []string{
		"switched to", "migrated to", "moved to", "now using", "now on",
	}

	providers := map[string][]string{
		"AWS":   {"aws", "amazon web services"},
		"GCP":   {"gcp", "google cloud"},
		"Azure": {"azure", "microsoft azure"},
	}

	// Get current cloud from CRF systems
	var currentCloud string
	var systemEntityID string
	for _, doc := range ctx.Systems {
		if hosting, ok := doc.Entity.Attributes["hosting"].(string); ok && hosting != "" {
			currentCloud = hosting
			systemEntityID = doc.Entity.ID
			break
		}
	}

	for _, trigger := range triggers {
		if strings.Contains(question, trigger) {
			for provider, keywords := range providers {
				for _, kw := range keywords {
					if strings.Contains(question, kw) {
						if currentCloud != "" && !strings.EqualFold(currentCloud, provider) {
							return &types.UpdateSuggestion{
								EntityID: systemEntityID,
								Field:    "hosting",
								OldValue: currentCloud,
								NewValue: provider,
								Reason:   fmt.Sprintf("you mentioned switching to %s", provider),
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// detectRunwayChangeCRF looks for signals like "we now have 18 months runway"
func detectRunwayChangeCRF(question string, ctx *types.CRFContext) *types.UpdateSuggestion {
	patterns := []struct {
		regex   string
		trigger string
	}{
		{`(?:now have|have|got)\s+(\d+)\s*months?\s*(?:of\s*)?runway`, "you mentioned your runway"},
		{`runway\s+(?:is|of)\s+(\d+)\s*months?`, "you mentioned your runway"},
		{`(\d+)\s*months?\s+(?:of\s*)?runway\s+(?:now|left|remaining)`, "you mentioned your runway"},
	}

	// Get current runway from CRF facts
	var currentRunway int
	var factEntityID string
	for _, doc := range ctx.Facts {
		if factType, ok := doc.Entity.Attributes["fact_type"].(string); ok && factType == "constraint" {
			if strings.Contains(strings.ToLower(doc.Entity.Name), "runway") {
				if val, ok := doc.Entity.Attributes["value"].(int); ok {
					currentRunway = val
					factEntityID = doc.Entity.ID
					break
				}
			}
		}
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.regex)
		matches := re.FindStringSubmatch(question)
		if len(matches) > 1 {
			newRunway, err := strconv.Atoi(matches[1])
			if err != nil || newRunway <= 0 || newRunway > 120 {
				continue
			}

			if currentRunway > 0 && !withinThreshold(currentRunway, newRunway, 0.2) {
				return &types.UpdateSuggestion{
					EntityID: factEntityID,
					Field:    "value",
					OldValue: fmt.Sprintf("%d", currentRunway),
					NewValue: fmt.Sprintf("%d", newRunway),
					Reason:   p.trigger,
				}
			}
		}
	}

	return nil
}
