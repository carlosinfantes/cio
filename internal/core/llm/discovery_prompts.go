// Package llm provides prompt building for the CIO - Chief Intelligence Officer.
package llm

import (
	"fmt"
	"strings"

	"github.com/carlosinfantes/cio/internal/types"
)

// BuildFacilitatorSystemPrompt creates the system prompt for discovery mode.
func BuildFacilitatorSystemPrompt(facilitator types.Persona, context *types.CRFContext) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You are %s, a skilled Socratic facilitator helping CTOs clarify their technical challenges before consulting an advisory board.

YOUR ROLE:
You help CTOs articulate and clarify their problems BEFORE bringing in expert advisors. You do NOT give advice or solutions - your job is purely to help them think through and express their challenge clearly.

YOUR APPROACH:
- Ask ONE focused follow-up question at a time
- Use Socratic questioning to uncover the real problem
- Be warm but direct - CTOs are busy
- Never give advice or solutions - clarification only
- Help surface implicit assumptions and constraints
- Keep responses brief (2-3 sentences max)

QUESTIONING PATTERNS:
1. Clarify vague terms: "When you say 'scaling issues', what specifically is happening?"
2. Explore context: "How long has this been a problem? What changed?"
3. Identify stakeholders: "Who else is affected by this?"
4. Surface constraints: "What's preventing the obvious solution?"
5. Understand goals: "What would success look like?"
6. Challenge assumptions: "What's making you lean toward that approach?"

CONVERSATION FLOW:
- If this is the start: Ask an open question like "What's on your mind today?"
- During conversation: Build on their responses with clarifying questions
- After 4-6 exchanges: Start summarizing: "So if I understand correctly..."
- When clarity is achieved: "I think we have enough to bring to the advisory board. Type /panel when you're ready."

`, facilitator.Name))

	sb.WriteString(fmt.Sprintf(`YOUR PERSONALITY:
Thinking style: "%s"
Background: %s
Priorities: %s

`, facilitator.ThinkingStyle, facilitator.Background, strings.Join(facilitator.Priorities, ", ")))

	// Add context if available
	if context != nil && context.GetOrganization() != nil {
		sb.WriteString("PROJECT CONTEXT (use this to ask more relevant questions):\n")
		sb.WriteString(formatCRFContextPrompt(context))
		sb.WriteString("\n")
	}

	sb.WriteString(`IMPORTANT RULES:
- Stay in character as the facilitator
- Do not break the fourth wall or mention you are an AI
- Focus purely on problem clarification, not solutions
- If they try to get advice, gently redirect: "Let's make sure we understand the full picture first..."
- Always respond with a question or a summary that invites confirmation
`)

	return sb.String()
}

// BuildFacilitatorUserPrompt creates the user prompt for a discovery turn.
func BuildFacilitatorUserPrompt(conversationHistory string, userMessage string, isFirstMessage bool) string {
	if isFirstMessage {
		return "Please greet the user warmly and ask what's on their mind today. Be brief and inviting."
	}

	var sb strings.Builder

	if conversationHistory != "" {
		sb.WriteString("CONVERSATION SO FAR:\n")
		sb.WriteString(conversationHistory)
		sb.WriteString("\n---\n\n")
	}

	sb.WriteString(fmt.Sprintf("USER'S LATEST MESSAGE:\n%s\n\n", userMessage))
	sb.WriteString("Respond with a clarifying question or, if you feel we have enough clarity, summarize what you've understood and suggest moving to the panel.")

	return sb.String()
}

// BuildBriefGenerationPrompt creates the prompt to generate a structured brief.
func BuildBriefGenerationPrompt(conversationHistory string) string {
	return fmt.Sprintf(`Analyze the following discovery conversation and generate a structured brief for an advisory board.

CONVERSATION:
%s

Generate a brief with the following sections. Be concise but comprehensive:

1. PROBLEM STATEMENT: One clear sentence describing the core challenge
2. CONTEXT: 2-3 sentences of relevant background
3. CONSTRAINTS: Bullet list of limitations (budget, time, team, tech, etc.)
4. GOALS: What success looks like (bullet list)
5. KEY QUESTIONS: 2-3 specific questions for the advisors to address

Also identify which specialist advisors should be included based on the conversation:
- cfo: Include if discussing budget, cost, ROI, pricing, investment, financial matters
- product: Include if discussing features, customers, MVP, roadmap, users, market
- devops: Include if discussing kubernetes, deployment, AWS/GCP/Azure, infrastructure, CI/CD

OUTPUT FORMAT (YAML - follow this exactly):
problem_statement: |
  [One sentence describing the core challenge]
context: |
  [2-3 sentences of background]
constraints:
  - [Constraint 1]
  - [Constraint 2]
goals:
  - [Goal 1]
  - [Goal 2]
key_questions:
  - [Question 1]
  - [Question 2]
suggested_advisors:
  - cto
  - architect
  - [other relevant advisors]`, conversationHistory)
}

// BuildPanelPromptWithBrief creates the panel prompt including the brief context.
func BuildPanelPromptWithBrief(brief *types.Brief) string {
	var sb strings.Builder

	sb.WriteString("The CTO has completed a discovery session and presents the following challenge:\n\n")

	sb.WriteString(fmt.Sprintf("PROBLEM: %s\n\n", brief.ProblemStatement))
	sb.WriteString(fmt.Sprintf("CONTEXT: %s\n\n", brief.Context))

	if len(brief.Constraints) > 0 {
		sb.WriteString("CONSTRAINTS:\n")
		for _, c := range brief.Constraints {
			sb.WriteString(fmt.Sprintf("- %s\n", c))
		}
		sb.WriteString("\n")
	}

	if len(brief.Goals) > 0 {
		sb.WriteString("GOALS:\n")
		for _, g := range brief.Goals {
			sb.WriteString(fmt.Sprintf("- %s\n", g))
		}
		sb.WriteString("\n")
	}

	if len(brief.KeyQuestions) > 0 {
		sb.WriteString("KEY QUESTIONS TO ADDRESS:\n")
		for _, q := range brief.KeyQuestions {
			sb.WriteString(fmt.Sprintf("- %s\n", q))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Please have each advisor respond to these questions in character, then provide a synthesis with recommended next steps.")

	return sb.String()
}
