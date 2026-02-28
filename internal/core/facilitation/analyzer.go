// Package facilitation implements message analysis for facilitation state updates.
package facilitation

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/carlosinfantes/cio/internal/core/llm"
	"github.com/carlosinfantes/cio/internal/types"
)

// Analyzer evaluates user messages and updates facilitation state.
type Analyzer struct {
	client *llm.Client
}

// NewAnalyzer creates a new facilitation analyzer.
func NewAnalyzer(client *llm.Client) *Analyzer {
	return &Analyzer{client: client}
}

// AnalysisResult contains the analysis of a user message.
type AnalysisResult struct {
	// Content classification
	ContainsProblem    bool     `json:"contains_problem"`
	ContainsContext    bool     `json:"contains_context"`
	ContainsConstraint bool     `json:"contains_constraint"`
	ContainsGoal       bool     `json:"contains_goal"`
	ContainsQuestion   bool     `json:"contains_question"`

	// Extracted information
	ProblemStatement string   `json:"problem_statement,omitempty"`
	Constraints      []string `json:"constraints,omitempty"`
	Goals            []string `json:"goals,omitempty"`

	// Suggested follow-up
	NeedsClarification bool     `json:"needs_clarification"`
	ClarifyingQuestion string   `json:"clarifying_question,omitempty"`
	Topics             []string `json:"topics,omitempty"`

	// Readiness assessment
	ReadyForPanel bool   `json:"ready_for_panel"`
	ReadinessNote string `json:"readiness_note,omitempty"`

	// Mode suggestion
	SuggestedMode   string   `json:"suggested_mode,omitempty"`
	SuggestedReason string   `json:"suggested_reason,omitempty"`
	RelevantDomains []string `json:"relevant_domains,omitempty"`
}

// AnalyzeMessage analyzes a user message and extracts structured information.
func (a *Analyzer) AnalyzeMessage(ctx context.Context, message string, state *FacilitationState, crfContext *types.CRFContext) (*AnalysisResult, error) {
	// Build the analysis prompt
	systemPrompt := buildAnalysisSystemPrompt()
	userPrompt := buildAnalysisUserPrompt(message, state, crfContext)

	resp, err := a.client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    512,
	})
	if err != nil {
		// Fall back to rule-based analysis
		return a.ruleBasedAnalysis(message, state), nil
	}

	// Parse JSON response
	result, err := parseAnalysisResponse(resp.Content)
	if err != nil {
		return a.ruleBasedAnalysis(message, state), nil
	}

	return result, nil
}

// UpdateStateFromAnalysis updates the facilitation state based on analysis results.
func (a *Analyzer) UpdateStateFromAnalysis(state *FacilitationState, result *AnalysisResult, userMessage string) {
	// Update problem if articulated
	if result.ContainsProblem && result.ProblemStatement != "" {
		state.MarkProblemArticulated(result.ProblemStatement)
	}

	// Add constraints
	if len(result.Constraints) > 0 {
		existing := make(map[string]bool)
		for _, c := range state.KeyConstraints {
			existing[c] = true
		}
		for _, c := range result.Constraints {
			if !existing[c] {
				state.KeyConstraints = append(state.KeyConstraints, c)
			}
		}
	}

	// Add goals
	if len(result.Goals) > 0 {
		existing := make(map[string]bool)
		for _, g := range state.Goals {
			existing[g] = true
		}
		for _, g := range result.Goals {
			if !existing[g] {
				state.Goals = append(state.Goals, g)
			}
		}
	}

	// Check if ready for escalation
	if result.ReadyForPanel {
		state.MarkDiscoveryComplete()
	}

	// Suggest advisors based on domains
	if len(result.RelevantDomains) > 0 {
		state.SuggestAdvisors(strings.Join(result.RelevantDomains, " "))
	}

	state.TurnCount++
	state.UpdatePhase()
}

// ShouldEscalate determines if facilitation should escalate to panel.
func (a *Analyzer) ShouldEscalate(state *FacilitationState) (bool, string) {
	if state.ReadyForEscalation() {
		return true, "All required information has been gathered. Ready to consult the advisory panel."
	}

	// Check for sufficient turns with problem articulated
	if state.ProblemArticulated && state.TurnCount >= 3 {
		return true, "We have enough context. Let me bring in the advisory panel."
	}

	return false, ""
}

// GetNextQuestion generates the next clarifying question based on state.
func (a *Analyzer) GetNextQuestion(ctx context.Context, state *FacilitationState, crfContext *types.CRFContext) (string, error) {
	// Determine what information is missing
	missingInfo := []string{}

	if !state.ProblemArticulated {
		return "What specific technical decision or challenge are you facing?", nil
	}

	if len(state.KeyConstraints) == 0 {
		missingInfo = append(missingInfo, "constraints")
	}
	if len(state.Goals) == 0 {
		missingInfo = append(missingInfo, "goals")
	}
	if state.ContextStaleDays > 30 {
		return "I notice your context hasn't been updated recently. Has anything changed with your infrastructure or team?", nil
	}

	// Generate contextual question
	if len(missingInfo) > 0 {
		systemPrompt := "You are Jordan, a discovery facilitator. Generate a single, focused clarifying question to gather missing information. Be conversational and brief."
		userPrompt := "Problem: " + state.ProblemStatement + "\nMissing info: " + strings.Join(missingInfo, ", ") + "\nGenerate one clarifying question:"

		resp, err := a.client.Query(ctx, llm.Request{
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
			MaxTokens:    128,
		})
		if err == nil {
			return strings.TrimSpace(resp.Content), nil
		}
	}

	// Default question
	return "Is there anything else important I should know before consulting the panel?", nil
}

// ruleBasedAnalysis performs simple rule-based analysis when LLM is unavailable.
func (a *Analyzer) ruleBasedAnalysis(message string, state *FacilitationState) *AnalysisResult {
	lower := strings.ToLower(message)
	result := &AnalysisResult{}

	// Detect problem statement
	problemPatterns := []string{"should we", "how do we", "what's the best", "need to decide", "considering", "thinking about", "challenge", "problem", "issue"}
	for _, pattern := range problemPatterns {
		if strings.Contains(lower, pattern) {
			result.ContainsProblem = true
			result.ProblemStatement = extractSentenceContaining(message, pattern)
			break
		}
	}

	// Detect constraints
	constraintPatterns := []string{"deadline", "budget", "must", "cannot", "limited", "constraint", "requirement", "compliance"}
	for _, pattern := range constraintPatterns {
		if strings.Contains(lower, pattern) {
			result.ContainsConstraint = true
			result.Constraints = append(result.Constraints, extractSentenceContaining(message, pattern))
		}
	}

	// Detect goals
	goalPatterns := []string{"want to", "goal", "objective", "aim", "trying to", "need to achieve", "success"}
	for _, pattern := range goalPatterns {
		if strings.Contains(lower, pattern) {
			result.ContainsGoal = true
			result.Goals = append(result.Goals, extractSentenceContaining(message, pattern))
		}
	}

	// Detect domains
	if containsAny(lower, []string{"security", "compliance", "auth", "encryption"}) {
		result.RelevantDomains = append(result.RelevantDomains, "security")
	}
	if containsAny(lower, []string{"kubernetes", "docker", "aws", "deploy", "infrastructure"}) {
		result.RelevantDomains = append(result.RelevantDomains, "infrastructure")
	}
	if containsAny(lower, []string{"team", "hire", "capacity", "velocity"}) {
		result.RelevantDomains = append(result.RelevantDomains, "team")
	}
	if containsAny(lower, []string{"cost", "budget", "roi", "pricing"}) {
		result.RelevantDomains = append(result.RelevantDomains, "financial")
	}

	// Determine if needs clarification
	result.NeedsClarification = !result.ContainsProblem || len(message) < 50
	if result.NeedsClarification {
		result.ClarifyingQuestion = "Could you tell me more about the specific situation?"
	}

	// Check readiness
	if state.ProblemArticulated && state.TurnCount >= 2 {
		result.ReadyForPanel = true
		result.ReadinessNote = "Sufficient context gathered"
	}

	return result
}

func buildAnalysisSystemPrompt() string {
	return `You are an expert at analyzing user messages in a technical advisory context.
Analyze the message and extract structured information.

Output ONLY valid JSON with this structure:
{
  "contains_problem": boolean,
  "contains_context": boolean,
  "contains_constraint": boolean,
  "contains_goal": boolean,
  "contains_question": boolean,
  "problem_statement": "extracted problem if present",
  "constraints": ["list of constraints mentioned"],
  "goals": ["list of goals mentioned"],
  "needs_clarification": boolean,
  "clarifying_question": "question to ask if clarification needed",
  "topics": ["technical topics mentioned"],
  "ready_for_panel": boolean,
  "readiness_note": "why ready or not ready",
  "suggested_mode": "panel|socratic|advocate|framework",
  "suggested_reason": "why this mode",
  "relevant_domains": ["security", "infrastructure", "team", "financial", "product", "architecture"]
}`
}

func buildAnalysisUserPrompt(message string, state *FacilitationState, crfContext *types.CRFContext) string {
	prompt := "Analyze this user message:\n\n" + message + "\n\n"

	// Add state context
	prompt += "Current facilitation state:\n"
	prompt += "- Problem articulated: " + boolToString(state.ProblemArticulated) + "\n"
	prompt += "- Turn count: " + string(rune('0'+state.TurnCount)) + "\n"
	if state.ProblemStatement != "" {
		prompt += "- Current problem: " + state.ProblemStatement + "\n"
	}

	// Add organizational context
	if crfContext != nil {
		if org := crfContext.GetOrganization(); org != nil {
			prompt += "- Organization: " + org.Name + "\n"
		}
	}

	return prompt
}

func parseAnalysisResponse(content string) (*AnalysisResult, error) {
	// Clean up response
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result AnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func extractSentenceContaining(text, pattern string) string {
	lower := strings.ToLower(text)
	idx := strings.Index(lower, pattern)
	if idx == -1 {
		return ""
	}

	// Find sentence boundaries
	start := idx
	for start > 0 && text[start-1] != '.' && text[start-1] != '!' && text[start-1] != '?' {
		start--
	}

	end := idx + len(pattern)
	for end < len(text) && text[end] != '.' && text[end] != '!' && text[end] != '?' {
		end++
	}
	if end < len(text) {
		end++
	}

	return strings.TrimSpace(text[start:end])
}

func boolToString(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
