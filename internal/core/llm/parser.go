// Package llm provides response parsing for the CTO Advisory Board.
package llm

import (
	"regexp"
	"strings"

	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

var (
	// Match headers like "## Victoria Chen — Fractional CTO" or "## Synthesis"
	headerRegex = regexp.MustCompile(`(?m)^##\s+(.+?)(?:\s*—\s*(.+?))?$`)
)

// ParseResponse extracts individual advisor responses from the LLM output.
func ParseResponse(content string, advisors []types.Persona) types.ParsedResponse {
	result := types.ParsedResponse{
		Advisors: []types.AdvisorResponse{},
	}

	// Find all headers and their positions
	matches := headerRegex.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		// No structured response, return as synthesis
		result.Synthesis = strings.TrimSpace(content)
		return result
	}

	// Extract sections between headers
	for i, match := range matches {
		// match[0:1] is full match start:end
		// match[2:3] is first group (name) start:end
		// match[4:5] is second group (role) start:end (may be -1 if not present)

		headerEnd := match[1]
		nameStart := match[2]
		nameEnd := match[3]

		name := strings.TrimSpace(content[nameStart:nameEnd])

		// Find content end (next header or end of string)
		var contentEnd int
		if i+1 < len(matches) {
			contentEnd = matches[i+1][0]
		} else {
			contentEnd = len(content)
		}

		sectionContent := strings.TrimSpace(content[headerEnd:contentEnd])

		// Check if this is the synthesis section
		nameLower := strings.ToLower(name)
		if nameLower == "synthesis" || strings.Contains(nameLower, "synthesis") {
			result.Synthesis = sectionContent
			continue
		}

		// Try to match to an advisor
		advisorID, advisorName, advisorRole := matchAdvisor(name, advisors)
		if advisorID != "" {
			result.Advisors = append(result.Advisors, types.AdvisorResponse{
				AdvisorID: advisorID,
				Name:      advisorName,
				Role:      advisorRole,
				Response:  sectionContent,
			})
		}
	}

	return result
}

// matchAdvisor tries to identify which advisor a header belongs to.
func matchAdvisor(headerName string, advisors []types.Persona) (types.AdvisorID, string, string) {
	headerLower := strings.ToLower(headerName)

	for _, advisor := range advisors {
		nameLower := strings.ToLower(advisor.Name)

		// Check for name match
		if strings.Contains(headerLower, nameLower) {
			return advisor.ID, advisor.Name, advisor.Role
		}

		// Check for first name match
		nameParts := strings.Fields(advisor.Name)
		if len(nameParts) > 0 {
			firstNameLower := strings.ToLower(nameParts[0])
			if strings.Contains(headerLower, firstNameLower) {
				return advisor.ID, advisor.Name, advisor.Role
			}
		}

		// Check for role match
		roleLower := strings.ToLower(advisor.Role)
		roleWords := strings.Fields(roleLower)
		for _, word := range roleWords {
			if len(word) > 3 && strings.Contains(headerLower, word) {
				return advisor.ID, advisor.Name, advisor.Role
			}
		}
	}

	return "", headerName, ""
}
