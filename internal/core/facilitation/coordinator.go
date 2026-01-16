// Package facilitation implements the session coordinator with auto-escalation.
package facilitation

import (
	"context"
	"fmt"
	"time"

	"github.com/carlosinfantes/cto-advisory-board/internal/core/llm"
	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

// Coordinator manages the facilitation flow with auto-escalation.
type Coordinator struct {
	client    *llm.Client
	analyzer  *Analyzer
	state     *FacilitationState
	crfCtx    *types.CRFContext
	session   *types.DiscoverySession
	callbacks CoordinatorCallbacks
}

// CoordinatorCallbacks provides hooks for UI updates.
type CoordinatorCallbacks struct {
	OnPhaseChange   func(fromPhase, toPhase FacilitationPhase, reason string)
	OnEscalation    func(reason string, brief *types.Brief)
	OnJordanMessage func(message string)
	OnSuggestion    func(suggestion string, options []string)
	OnContextStale  func(staleDays int, missing []string)
}

// NewCoordinator creates a new facilitation coordinator.
func NewCoordinator(client *llm.Client, crfCtx *types.CRFContext, callbacks CoordinatorCallbacks) *Coordinator {
	state := NewFacilitationState()

	// Check context completeness
	if crfCtx != nil && len(crfCtx.AllEntities()) > 0 {
		// Check for staleness
		staleDays := 0
		var missing []string

		// Simple staleness check
		if org := crfCtx.GetOrganization(); org != nil {
			days := int(time.Since(org.Provenance.UpdatedAt).Hours() / 24)
			if org.Provenance.UpdatedAt.IsZero() {
				days = int(time.Since(org.Provenance.CreatedAt).Hours() / 24)
			}
			staleDays = days
		}

		// Check for missing essential context
		if crfCtx.GetOrganization() == nil {
			missing = append(missing, "organization")
		}
		if len(crfCtx.Systems) == 0 {
			missing = append(missing, "systems")
		}

		state.MarkContextComplete(staleDays, missing)

		if callbacks.OnContextStale != nil && (staleDays > 30 || len(missing) > 0) {
			callbacks.OnContextStale(staleDays, missing)
		}
	}

	return &Coordinator{
		client:    client,
		analyzer:  NewAnalyzer(client),
		state:     state,
		crfCtx:    crfCtx,
		session:   types.NewDiscoverySession(),
		callbacks: callbacks,
	}
}

// ProcessMessage processes a user message and returns Jordan's response.
func (c *Coordinator) ProcessMessage(ctx context.Context, message string) (*ProcessResult, error) {
	oldPhase := c.state.Phase

	// Analyze the message
	analysis, err := c.analyzer.AnalyzeMessage(ctx, message, c.state, c.crfCtx)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	// Update state from analysis
	c.analyzer.UpdateStateFromAnalysis(c.state, analysis, message)

	// Add message to session
	c.session.AddMessage("user", message)

	// Check for phase change
	if c.state.Phase != oldPhase && c.callbacks.OnPhaseChange != nil {
		c.callbacks.OnPhaseChange(oldPhase, c.state.Phase, c.getPhaseChangeReason(oldPhase, c.state.Phase))
	}

	// Check for auto-escalation
	shouldEscalate, escalationReason := c.analyzer.ShouldEscalate(c.state)
	if shouldEscalate {
		return c.handleEscalation(escalationReason)
	}

	// Generate Jordan's response
	response, err := c.generateJordanResponse(ctx, message, analysis)
	if err != nil {
		return nil, fmt.Errorf("response generation failed: %w", err)
	}

	// Add Jordan's response to session
	c.session.AddMessage("facilitator", response)

	// Check for suggestions
	if analysis.SuggestedMode != "" && analysis.SuggestedMode != string(types.ModePanel) {
		c.handleModeSuggestion(analysis)
	}

	return &ProcessResult{
		Response:     response,
		Phase:        c.state.Phase,
		ReadyForPanel: c.state.ReadyForEscalation(),
	}, nil
}

// ForceEscalate forces escalation to the panel.
func (c *Coordinator) ForceEscalate() (*ProcessResult, error) {
	return c.handleEscalation("User requested panel escalation")
}

// GetState returns the current facilitation state.
func (c *Coordinator) GetState() *FacilitationState {
	return c.state
}

// GetSession returns the current discovery session.
func (c *Coordinator) GetSession() *types.DiscoverySession {
	return c.session
}

// GetBrief generates a brief from the current state.
func (c *Coordinator) GetBrief() *types.Brief {
	return c.state.ToBrief()
}

// ProcessResult contains the result of processing a message.
type ProcessResult struct {
	Response      string
	Phase         FacilitationPhase
	ReadyForPanel bool
	Escalated     bool
	Brief         *types.Brief
	SuggestedMode types.Mode
}

func (c *Coordinator) handleEscalation(reason string) (*ProcessResult, error) {
	c.state.MarkEscalated()

	brief := c.state.ToBrief()

	if c.callbacks.OnEscalation != nil {
		c.callbacks.OnEscalation(reason, brief)
	}

	return &ProcessResult{
		Response:      reason,
		Phase:         PhaseEscalated,
		ReadyForPanel: true,
		Escalated:     true,
		Brief:         brief,
		SuggestedMode: c.state.SuggestedMode,
	}, nil
}

func (c *Coordinator) generateJordanResponse(ctx context.Context, userMessage string, analysis *AnalysisResult) (string, error) {
	// Build prompt based on state
	var responseType string
	switch c.state.Phase {
	case PhaseContextGathering:
		responseType = "gathering context"
	case PhaseProblemArticulation:
		responseType = "understanding the problem"
	case PhaseDiscovery:
		responseType = "clarifying details"
	case PhaseReadyForEscalation:
		responseType = "ready to escalate"
	default:
		responseType = "facilitating"
	}

	systemPrompt := fmt.Sprintf(`You are Jordan, a discovery facilitator for the CTO Advisory Board.
Your role is to help users articulate their technical decisions clearly.
You are currently %s.

Guidelines:
- Be conversational and supportive
- Ask one clarifying question at a time
- Acknowledge what the user has shared
- If you have enough information, indicate readiness to escalate to the panel
- Keep responses brief (2-4 sentences)
- Never make decisions for the user
- Focus on understanding the problem, constraints, and goals`, responseType)

	conversationContext := ""
	if len(c.session.Messages) > 0 {
		conversationContext = "Previous conversation:\n" + c.session.GetConversationText() + "\n\n"
	}

	stateContext := fmt.Sprintf(`Current state:
- Problem articulated: %v
- Constraints identified: %d
- Goals identified: %d
- Turn count: %d`,
		c.state.ProblemArticulated,
		len(c.state.KeyConstraints),
		len(c.state.Goals),
		c.state.TurnCount)

	userPrompt := fmt.Sprintf(`%s%s

User's latest message: %s

%s

Generate a response that either:
1. Asks a clarifying question if more info is needed
2. Summarizes understanding and offers to escalate to the panel if ready`,
		conversationContext,
		stateContext,
		userMessage,
		c.getGuidanceForPhase())

	resp, err := c.client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    256,
	})
	if err != nil {
		// Fallback response
		return c.getFallbackResponse(), nil
	}

	return resp.Content, nil
}

func (c *Coordinator) getGuidanceForPhase() string {
	switch c.state.Phase {
	case PhaseContextGathering:
		return "Focus on understanding the organizational context and current technical landscape."
	case PhaseProblemArticulation:
		return "Help the user clearly state the decision or problem they're facing."
	case PhaseDiscovery:
		return "Ask about constraints, timeline, goals, or any missing details."
	case PhaseReadyForEscalation:
		return "Summarize what you've learned and offer to bring in the advisory panel."
	default:
		return ""
	}
}

func (c *Coordinator) getFallbackResponse() string {
	switch c.state.Phase {
	case PhaseContextGathering:
		return "Could you tell me about your current technical setup and team?"
	case PhaseProblemArticulation:
		return "What specific decision or challenge are you facing?"
	case PhaseDiscovery:
		return "Are there any constraints or deadlines I should know about?"
	case PhaseReadyForEscalation:
		return "I think I have a good understanding. Would you like me to bring in the advisory panel?"
	default:
		return "Tell me more about your situation."
	}
}

func (c *Coordinator) getPhaseChangeReason(from, to FacilitationPhase) string {
	switch to {
	case PhaseContextGathering:
		return "Let me learn about your context"
	case PhaseProblemArticulation:
		return "Now help me understand the specific problem"
	case PhaseDiscovery:
		return "Let me ask a few clarifying questions"
	case PhaseReadyForEscalation:
		return "I have what I need to consult the panel"
	case PhaseEscalated:
		return "Bringing in the advisory panel"
	default:
		return ""
	}
}

func (c *Coordinator) handleModeSuggestion(analysis *AnalysisResult) {
	if c.callbacks.OnSuggestion == nil {
		return
	}

	var suggestion string
	var options []string

	switch analysis.SuggestedMode {
	case "framework":
		suggestion = "This looks like a comparison between options. Would you like to use the Framework mode for structured evaluation?"
		options = []string{"Yes, use Framework mode", "No, continue with Panel mode"}
	case "advocate":
		suggestion = "It sounds like you want to validate a decision. Would you like to use Devil's Advocate mode?"
		options = []string{"Yes, challenge my thinking", "No, continue with Panel mode"}
	}

	if suggestion != "" {
		c.callbacks.OnSuggestion(suggestion, options)
	}
}
