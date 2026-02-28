// Package discovery handles discovery session storage and management.
package discovery

import (
	"context"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/carlosinfantes/cio/internal/core/advisors"
	"github.com/carlosinfantes/cio/internal/core/llm"
	"github.com/carlosinfantes/cio/internal/types"
)

// GenerateBrief creates a structured brief from a discovery conversation.
func GenerateBrief(ctx context.Context, client *llm.Client, session *types.DiscoverySession) (*types.Brief, error) {
	if session == nil || len(session.Messages) == 0 {
		return nil, nil
	}

	conversationText := session.GetConversationText()

	// Use the brief generation prompt
	prompt := llm.BuildBriefGenerationPrompt(conversationText)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: "You are an expert at analyzing conversations and extracting structured information. Output only valid YAML.",
		UserPrompt:   prompt,
		MaxTokens:    1024,
	})
	if err != nil {
		return nil, err
	}

	// Parse the YAML response
	brief, err := parseBriefFromYAML(resp.Content)
	if err != nil {
		// If parsing fails, create a basic brief from the conversation
		return createFallbackBrief(conversationText), nil
	}

	// Ensure suggested advisors includes the base advisors
	brief.SuggestedAdvisors = ensureBaseAdvisors(brief.SuggestedAdvisors)

	return brief, nil
}

// parseBriefFromYAML extracts a Brief from YAML content.
func parseBriefFromYAML(content string) (*types.Brief, error) {
	// Clean up the content - remove markdown code blocks if present
	content = strings.TrimPrefix(content, "```yaml")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var brief types.Brief
	if err := yaml.Unmarshal([]byte(content), &brief); err != nil {
		return nil, err
	}

	return &brief, nil
}

// createFallbackBrief creates a basic brief when parsing fails.
func createFallbackBrief(conversationText string) *types.Brief {
	return &types.Brief{
		ProblemStatement:  "Problem extracted from discovery conversation",
		Context:           conversationText,
		Constraints:       []string{},
		Goals:             []string{},
		KeyQuestions:      []string{"What approach should we take?"},
		SuggestedAdvisors: []types.AdvisorID{types.AdvisorCTO, types.AdvisorArchitect},
	}
}

// ensureBaseAdvisors makes sure CTO and Architect are always included.
func ensureBaseAdvisors(advisorIDs []types.AdvisorID) []types.AdvisorID {
	hasID := func(id types.AdvisorID) bool {
		for _, a := range advisorIDs {
			if a == id {
				return true
			}
		}
		return false
	}

	result := make([]types.AdvisorID, 0, len(advisorIDs)+2)

	// Add base advisors first if not present
	if !hasID(types.AdvisorCTO) {
		result = append(result, types.AdvisorCTO)
	}
	if !hasID(types.AdvisorArchitect) {
		result = append(result, types.AdvisorArchitect)
	}

	// Add all existing advisors
	result = append(result, advisorIDs...)

	return result
}

// SuggestAdvisorsFromBrief analyzes a brief and returns recommended advisors.
func SuggestAdvisorsFromBrief(brief *types.Brief) []types.AdvisorID {
	if brief == nil {
		return []types.AdvisorID{types.AdvisorCTO, types.AdvisorArchitect, types.AdvisorCISO, types.AdvisorVPEng}
	}

	// Use the suggested advisors from the brief if available
	if len(brief.SuggestedAdvisors) > 0 {
		return ensureBaseAdvisors(brief.SuggestedAdvisors)
	}

	// Otherwise, analyze the brief content for specialist keywords
	allText := strings.ToLower(strings.Join([]string{
		brief.ProblemStatement,
		brief.Context,
		strings.Join(brief.Constraints, " "),
		strings.Join(brief.Goals, " "),
		strings.Join(brief.KeyQuestions, " "),
	}, " "))

	result := []types.AdvisorID{types.AdvisorCTO, types.AdvisorArchitect}

	// Check for security-related content
	if containsAny(allText, []string{"security", "compliance", "risk", "breach", "vulnerability", "encryption", "auth"}) {
		result = append(result, types.AdvisorCISO)
	}

	// Check for team/execution content
	if containsAny(allText, []string{"team", "capacity", "hire", "velocity", "delivery", "process", "sprint"}) {
		result = append(result, types.AdvisorVPEng)
	}

	// Use existing specialist detection
	summonResults := advisors.SummonSpecialists(allText)
	for _, sr := range summonResults {
		// Check if not already in result
		found := false
		for _, r := range result {
			if r == sr.Specialist.ID {
				found = true
				break
			}
		}
		if !found {
			result = append(result, sr.Specialist.ID)
		}
	}

	return result
}

// containsAny checks if text contains any of the keywords.
func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
