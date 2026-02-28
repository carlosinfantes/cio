// Package llm provides prompt building for the CIO - Chief Intelligence Officer.
package llm

import (
	"fmt"
	"strings"

	"github.com/carlosinfantes/cio/internal/types"
)

// BuildSystemPrompt creates the system prompt for the advisory board.
func BuildSystemPrompt(advisors []types.Persona, context *types.CRFContext, mode types.Mode) string {
	var sb strings.Builder

	// Base instructions
	sb.WriteString(`You are simulating a CIO - Chief Intelligence Officer meeting. You will embody multiple expert advisors, each with distinct personalities and expertise.

CRITICAL INSTRUCTIONS:
1. Respond AS each advisor in turn, using their voice and perspective
2. Use the exact format: "## [Advisor Name] — [Role]" for each section
3. Keep each advisor response to 2-4 sentences - concise and impactful
4. End with "## Synthesis" summarizing key insights and recommended next steps
5. Do NOT break character or mention you are an AI

`)

	// Add advisor personalities
	sb.WriteString("THE ADVISORS:\n\n")
	for _, advisor := range advisors {
		sb.WriteString(formatAdvisorPrompt(advisor))
		sb.WriteString("\n")
	}

	// Add context if available
	if context != nil {
		sb.WriteString("PROJECT CONTEXT:\n\n")
		sb.WriteString(formatCRFContextPrompt(context))
		sb.WriteString("\n")
	}

	// Add mode-specific instructions
	sb.WriteString(formatModeInstructions(mode))

	return sb.String()
}

func formatAdvisorPrompt(advisor types.Persona) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### %s — %s\n", advisor.Name, advisor.Role))
	sb.WriteString(fmt.Sprintf("Thinking style: \"%s\"\n", advisor.ThinkingStyle))
	sb.WriteString(fmt.Sprintf("Background: %s\n", advisor.Background))
	sb.WriteString("Priorities: ")
	sb.WriteString(strings.Join(advisor.Priorities, ", "))
	sb.WriteString("\n")

	return sb.String()
}

// formatCRFContextPrompt renders CRF context entities into a prompt-friendly format.
func formatCRFContextPrompt(ctx *types.CRFContext) string {
	if ctx == nil {
		return ""
	}

	var sb strings.Builder

	// Organization info (company)
	if org := ctx.GetOrganization(); org != nil {
		sb.WriteString(fmt.Sprintf("Company: %s", org.Name))
		if industry, ok := org.Attributes["industry"].(string); ok && industry != "" {
			sb.WriteString(fmt.Sprintf(" (%s", industry))
			if size, ok := org.Attributes["size"].(string); ok && size != "" {
				sb.WriteString(fmt.Sprintf(", %s", size))
			}
			sb.WriteString(")")
		}
		sb.WriteString("\n")

		// Compliance frameworks
		if frameworks, ok := org.Attributes["compliance_frameworks"].([]interface{}); ok && len(frameworks) > 0 {
			var compList []string
			for _, f := range frameworks {
				if s, ok := f.(string); ok {
					compList = append(compList, s)
				}
			}
			if len(compList) > 0 {
				sb.WriteString(fmt.Sprintf("Compliance: %s\n", strings.Join(compList, ", ")))
			}
		}
	}

	// Team info from CRF organization entities of type "team"
	teams := ctx.GetTeams()
	if len(teams) > 0 {
		totalEngineers := 0
		var teamParts []string
		for _, team := range teams {
			if headcount, ok := team.Attributes["headcount"].(int); ok {
				totalEngineers += headcount
				teamParts = append(teamParts, fmt.Sprintf("%s: %d", team.Name, headcount))
			}
		}
		if totalEngineers > 0 {
			sb.WriteString(fmt.Sprintf("Team: %d engineers total\n", totalEngineers))
		}
		if len(teamParts) > 0 {
			sb.WriteString(fmt.Sprintf("Structure: %s\n", strings.Join(teamParts, ", ")))
		}
	}

	// Tech stack from CRF system entities
	if len(ctx.Systems) > 0 {
		var techStack []string
		var hosting string
		for _, doc := range ctx.Systems {
			if stack, ok := doc.Entity.Attributes["technology_stack"].([]interface{}); ok {
				for _, tech := range stack {
					if t, ok := tech.(string); ok {
						techStack = append(techStack, t)
					}
				}
			}
			if h, ok := doc.Entity.Attributes["hosting"].(string); ok && h != "" {
				hosting = h
			}
		}
		if len(techStack) > 0 {
			sb.WriteString(fmt.Sprintf("Tech Stack: %s\n", strings.Join(techStack, ", ")))
		}
		if hosting != "" {
			sb.WriteString(fmt.Sprintf("Cloud: %s\n", hosting))
		}
	}

	// Capabilities
	if len(ctx.Capabilities) > 0 {
		var caps []string
		for _, doc := range ctx.Capabilities {
			proficiency, _ := doc.Entity.Attributes["proficiency"].(string)
			if proficiency == "advanced" || proficiency == "expert" {
				caps = append(caps, doc.Entity.Name)
			}
		}
		if len(caps) > 0 {
			sb.WriteString(fmt.Sprintf("Key Capabilities: %s\n", strings.Join(caps, ", ")))
		}
	}

	// Constraints from facts
	for _, doc := range ctx.Facts {
		if factType, ok := doc.Entity.Attributes["fact_type"].(string); ok && factType == "constraint" {
			value := doc.Entity.Attributes["value"]
			unit, _ := doc.Entity.Attributes["unit"].(string)
			if value != nil {
				sb.WriteString(fmt.Sprintf("%s: %v %s\n", doc.Entity.Name, value, unit))
			}
		}
	}

	// Active policies
	if len(ctx.Policies) > 0 {
		sb.WriteString("Active Policies:\n")
		for _, doc := range ctx.Policies {
			enforcement, _ := doc.Entity.Attributes["enforcement"].(string)
			if enforcement == "mandatory" {
				sb.WriteString(fmt.Sprintf("  - %s: %s\n", doc.Entity.Name, doc.Entity.Description))
			}
		}
	}

	return sb.String()
}

func formatModeInstructions(mode types.Mode) string {
	switch mode {
	case types.ModeSocratic:
		return `MODE: SOCRATIC
Before providing advice, first ask 3-5 clarifying questions to better understand the situation.
Format: Start with "## Clarifying Questions" section, then proceed with advisor responses.
`
	case types.ModeAdvocate:
		return `MODE: DEVIL'S ADVOCATE
Challenge the premise and decision. Each advisor should:
1. Identify potential blind spots
2. Argue the opposite position
3. Highlight risks that may be underestimated
Be constructively critical - the goal is to stress-test the decision.
`
	case types.ModeFramework:
		return `MODE: DECISION FRAMEWORK
Structure the response as an evaluation matrix:
1. First identify the options being compared (A vs B vs C)
2. Generate evaluation criteria relevant to this decision
3. Score each option on each criterion (1-5)
4. Provide weighted recommendation
`
	default: // ModePanel
		return `MODE: PANEL DISCUSSION
Each advisor gives their perspective, then synthesis combines insights.
Focus on actionable advice specific to their context.
`
	}
}

// BuildUserPrompt creates the user prompt with the question.
func BuildUserPrompt(question string) string {
	return fmt.Sprintf(`The CTO has brought the following question to the advisory board:

"%s"

Please have each advisor respond in character, then provide a synthesis.`, question)
}

// BuildUserPromptWithContext creates the user prompt with session context.
func BuildUserPromptWithContext(question string, sessionContext string) string {
	if sessionContext == "" {
		return BuildUserPrompt(question)
	}

	return fmt.Sprintf(`%s

The CTO has brought the following question to the advisory board:

"%s"

Please have each advisor respond in character, considering the previous discussion context. Then provide a synthesis.`, sessionContext, question)
}

// BuildSocraticQuestionsPrompt creates a prompt specifically for generating clarifying questions.
func BuildSocraticQuestionsPrompt(context *types.CRFContext) string {
	var sb strings.Builder

	sb.WriteString(`You are a senior technical advisor helping a CTO clarify their question before bringing it to an advisory board.

Your task is to generate 3-5 clarifying questions that will help the advisors give more relevant and actionable advice.

INSTRUCTIONS:
1. Generate exactly 3-5 questions
2. Focus on understanding context, constraints, and goals
3. Each question should be on its own line, starting with a number (1., 2., etc.)
4. Questions should be specific and actionable, not generic
5. Do NOT provide any advice yet - only questions

Good clarifying questions cover:
- Timeline and urgency
- Constraints (budget, team, tech)
- Success criteria
- Stakeholders affected
- Previous attempts or existing solutions

`)

	if context != nil {
		sb.WriteString("PROJECT CONTEXT:\n")
		sb.WriteString(formatCRFContextPrompt(context))
		sb.WriteString("\n")
	}

	return sb.String()
}

// ParseSocraticQuestions extracts questions from LLM response.
func ParseSocraticQuestions(content string) []string {
	var questions []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for numbered questions (1., 2., etc.) or questions ending with ?
		if len(line) > 2 && (line[0] >= '1' && line[0] <= '9') && line[1] == '.' {
			// Remove the number prefix
			q := strings.TrimSpace(line[2:])
			if q != "" {
				questions = append(questions, q)
			}
		} else if strings.HasSuffix(line, "?") && !strings.HasPrefix(line, "#") {
			questions = append(questions, line)
		}
	}

	return questions
}

// ParseFrameworkOptions extracts options from a comparison question.
// e.g., "AWS vs GCP vs Azure" -> ["AWS", "GCP", "Azure"]
func ParseFrameworkOptions(question string) []string {
	// Common comparison patterns
	separators := []string{" vs ", " vs. ", " versus ", " or ", " VS "}

	for _, sep := range separators {
		if strings.Contains(strings.ToLower(question), strings.ToLower(sep)) {
			// Use case-insensitive split
			lower := strings.ToLower(question)
			lowerSep := strings.ToLower(sep)
			indices := []int{0}

			pos := 0
			for {
				idx := strings.Index(lower[pos:], lowerSep)
				if idx == -1 {
					break
				}
				indices = append(indices, pos+idx)
				pos = pos + idx + len(sep)
			}
			indices = append(indices, len(question))

			var options []string
			for i := 0; i < len(indices)-1; i++ {
				start := indices[i]
				if i > 0 {
					start += len(sep)
				}
				end := indices[i+1]
				opt := strings.TrimSpace(question[start:end])
				// Clean up - remove question marks, "should I use", etc.
				opt = strings.TrimSuffix(opt, "?")
				opt = strings.TrimPrefix(opt, "Should I use ")
				opt = strings.TrimPrefix(opt, "should I use ")
				opt = strings.TrimPrefix(opt, "Should we use ")
				opt = strings.TrimPrefix(opt, "should we use ")
				if opt != "" {
					options = append(options, opt)
				}
			}
			if len(options) >= 2 {
				return options
			}
		}
	}

	return nil
}

// BuildFrameworkCriteriaPrompt creates a prompt for generating evaluation criteria.
func BuildFrameworkCriteriaPrompt(question string, options []string, context *types.CRFContext) string {
	var sb strings.Builder

	sb.WriteString(`You are helping evaluate a technical decision. Generate 4-6 evaluation criteria for comparing the options.

INSTRUCTIONS:
1. Generate exactly 4-6 criteria relevant to this specific decision
2. Each criterion should be on its own line in this format:
   CRITERION: <name> | <description> | <weight 1-5>
3. Weight indicates importance (1=low, 5=critical)
4. Focus on criteria that differentiate the options
5. Include both technical and business factors

Example output:
CRITERION: Cost | Monthly operational cost including compute, storage, and networking | 4
CRITERION: Scalability | Ability to handle 10x growth without major re-architecture | 5
CRITERION: Learning Curve | Time for team to become productive with the technology | 3

`)

	sb.WriteString(fmt.Sprintf("DECISION: %s\n", question))
	sb.WriteString(fmt.Sprintf("OPTIONS: %s\n\n", strings.Join(options, ", ")))

	if context != nil {
		sb.WriteString("PROJECT CONTEXT:\n")
		sb.WriteString(formatCRFContextPrompt(context))
		sb.WriteString("\n")
	}

	return sb.String()
}

// ParseFrameworkCriteria extracts criteria from LLM response.
func ParseFrameworkCriteria(content string) []types.FrameworkCriterion {
	var criteria []types.FrameworkCriterion
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "CRITERION:") {
			continue
		}

		// Parse: CRITERION: name | description | weight
		parts := strings.Split(strings.TrimPrefix(line, "CRITERION:"), "|")
		if len(parts) < 3 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		desc := strings.TrimSpace(parts[1])
		weightStr := strings.TrimSpace(parts[2])

		var weight float64
		fmt.Sscanf(weightStr, "%f", &weight)
		if weight < 1 {
			weight = 3 // Default to medium
		}
		if weight > 5 {
			weight = 5
		}

		criteria = append(criteria, types.FrameworkCriterion{
			Name:        name,
			Description: desc,
			Weight:      weight,
		})
	}

	return criteria
}

// BuildFrameworkScoringPrompt creates a prompt for scoring options.
func BuildFrameworkScoringPrompt(state *types.FrameworkState, context *types.CRFContext) string {
	var sb strings.Builder

	sb.WriteString(`You are evaluating options for a technical decision. Score each option against each criterion.

INSTRUCTIONS:
1. Score each option on each criterion (1-5 scale)
2. Use this exact format for each score:
   SCORE: <option> | <criterion> | <score 1-5> | <brief rationale>
3. Be objective and specific in rationales
4. After all scores, provide:
   RECOMMENDATION: <option name>
   CONFIDENCE: <low|medium|high>
   RATIONALE: <1-2 sentence explanation>

Scoring scale:
1 = Poor/Weak
2 = Below Average
3 = Average/Acceptable
4 = Good/Strong
5 = Excellent/Best-in-class

`)

	sb.WriteString(fmt.Sprintf("DECISION: %s\n\n", state.Question))
	sb.WriteString("OPTIONS:\n")
	for _, opt := range state.Options {
		sb.WriteString(fmt.Sprintf("- %s\n", opt))
	}

	sb.WriteString("\nCRITERIA:\n")
	for _, crit := range state.Criteria {
		sb.WriteString(fmt.Sprintf("- %s (weight: %.0f): %s\n", crit.Name, crit.Weight, crit.Description))
	}

	if context != nil {
		sb.WriteString("\nPROJECT CONTEXT:\n")
		sb.WriteString(formatCRFContextPrompt(context))
	}

	return sb.String()
}

// LongQuestionThreshold is the character count above which questions are summarized.
const LongQuestionThreshold = 2000

// BuildSummarizeQuestionPrompt creates a prompt for summarizing long questions.
func BuildSummarizeQuestionPrompt() string {
	return `You are helping summarize a long question/problem statement for an advisory board.

INSTRUCTIONS:
1. Distill the key question or decision being asked
2. Preserve critical context, constraints, and requirements
3. Keep the summary under 500 characters
4. Maintain the original intent and urgency
5. Output ONLY the summarized question, nothing else

Focus on: What is being decided? What are the key constraints? What outcome is desired?`
}

// BuildSummarizeQuestionUserPrompt creates the user prompt with the long question.
func BuildSummarizeQuestionUserPrompt(question string) string {
	return fmt.Sprintf(`Please summarize this question/problem for the advisory board:

---
%s
---

Provide a concise summary (under 500 characters) that captures the core question and key context.`, question)
}

// ParseFrameworkScores extracts scores from LLM response.
func ParseFrameworkScores(content string, state *types.FrameworkState) {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "SCORE:") {
			// Parse: SCORE: option | criterion | score | rationale
			parts := strings.Split(strings.TrimPrefix(line, "SCORE:"), "|")
			if len(parts) < 4 {
				continue
			}

			option := strings.TrimSpace(parts[0])
			criterion := strings.TrimSpace(parts[1])
			scoreStr := strings.TrimSpace(parts[2])
			rationale := strings.TrimSpace(parts[3])

			var score int
			fmt.Sscanf(scoreStr, "%d", &score)
			if score < 1 {
				score = 1
			}
			if score > 5 {
				score = 5
			}

			state.Scores = append(state.Scores, types.FrameworkScore{
				Option:    option,
				Criterion: criterion,
				Score:     score,
				Rationale: rationale,
			})
		} else if strings.HasPrefix(line, "RECOMMENDATION:") {
			state.Recommendation = strings.TrimSpace(strings.TrimPrefix(line, "RECOMMENDATION:"))
		} else if strings.HasPrefix(line, "CONFIDENCE:") {
			state.Confidence = strings.TrimSpace(strings.TrimPrefix(line, "CONFIDENCE:"))
		}
	}
}
