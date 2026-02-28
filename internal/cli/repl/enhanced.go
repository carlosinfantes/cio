// Package repl implements the enhanced REPL with facilitation coordinator integration.
package repl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chzyer/readline"

	"github.com/carlosinfantes/cio/internal/cli/output"
	"github.com/carlosinfantes/cio/internal/core/facilitation"
	"github.com/carlosinfantes/cio/internal/core/llm"
	"github.com/carlosinfantes/cio/internal/types"
)

// EnhancedSession extends Session with facilitation coordinator.
type EnhancedSession struct {
	*Session
	Coordinator       *facilitation.Coordinator
	FacilitationState *facilitation.FacilitationState
	PendingSuggestion *ModeSuggestion
}

// ModeSuggestion represents a pending mode suggestion from Jordan.
type ModeSuggestion struct {
	Message string
	Options []string
	Mode    types.Mode
}

// NewEnhancedSession creates a new enhanced session with facilitation.
func NewEnhancedSession(client *llm.Client, projectCtx *types.CRFContext) *EnhancedSession {
	baseSession := NewSession()

	// Create facilitation callbacks
	var enhancedSession *EnhancedSession

	callbacks := facilitation.CoordinatorCallbacks{
		OnPhaseChange: func(from, to facilitation.FacilitationPhase, reason string) {
			output.PrintPhaseTransition(string(from), string(to), reason)
		},
		OnEscalation: func(reason string, brief *types.Brief) {
			output.PrintEscalationNotice(reason)
			if enhancedSession != nil {
				enhancedSession.Session.SetBrief(brief)
				enhancedSession.Session.SwitchToPanel()
			}
		},
		OnJordanMessage: func(message string) {
			output.PrintFacilitatorMessage(message)
		},
		OnSuggestion: func(suggestion string, options []string) {
			if enhancedSession != nil {
				enhancedSession.PendingSuggestion = &ModeSuggestion{
					Message: suggestion,
					Options: options,
				}
			}
			output.PrintJordanSuggestion(suggestion, options)
		},
		OnContextStale: func(staleDays int, missing []string) {
			output.PrintContextValidation(staleDays > 30, staleDays, missing)
		},
	}

	coordinator := facilitation.NewCoordinator(client, projectCtx, callbacks)

	enhancedSession = &EnhancedSession{
		Session:           baseSession,
		Coordinator:       coordinator,
		FacilitationState: coordinator.GetState(),
	}

	return enhancedSession
}

// ProcessWithCoordinator processes input through the facilitation coordinator.
func (es *EnhancedSession) ProcessWithCoordinator(ctx context.Context, input string) (*facilitation.ProcessResult, error) {
	result, err := es.Coordinator.ProcessMessage(ctx, input)
	if err != nil {
		return nil, err
	}

	es.FacilitationState = es.Coordinator.GetState()

	// Handle auto-escalation
	if result.Escalated {
		es.Session.SwitchToPanel()
		es.Session.SetBrief(result.Brief)

		// Update advisors based on suggestions
		if len(es.FacilitationState.SuggestedAdvisors) > 0 {
			es.Session.SetAdvisors(es.FacilitationState.SuggestedAdvisors)
		}
	}

	return result, nil
}

// HandleModeSwitchKey handles Ctrl+M key press for mode switching.
func (es *EnhancedSession) HandleModeSwitchKey() {
	fmt.Println()
	fmt.Println("┌─────────────────────────────────────┐")
	fmt.Println("│  Switch Mode                        │")
	fmt.Println("├─────────────────────────────────────┤")
	fmt.Println("│  [1] Discovery  - Talk to Jordan    │")
	fmt.Println("│  [2] Panel      - Advisory board    │")
	fmt.Println("│  [3] Socratic   - Q&A enrichment    │")
	fmt.Println("│  [4] Advocate   - Challenge mode    │")
	fmt.Println("│  [5] Framework  - Compare options   │")
	fmt.Println("│  [0] Cancel                         │")
	fmt.Println("└─────────────────────────────────────┘")
}

// SwitchMode switches to the specified mode.
func (es *EnhancedSession) SwitchMode(modeKey string) bool {
	switch modeKey {
	case "1":
		es.Session.SwitchToDiscovery()
		output.PrintSuccess("Switched to Discovery mode")
		return true
	case "2":
		es.Session.SwitchToPanel()
		es.Session.SetMode(types.ModePanel)
		output.PrintSuccess("Switched to Panel mode")
		return true
	case "3":
		es.Session.SwitchToPanel()
		es.Session.SetMode(types.ModeSocratic)
		output.PrintSuccess("Switched to Socratic mode")
		return true
	case "4":
		es.Session.SwitchToPanel()
		es.Session.SetMode(types.ModeAdvocate)
		output.PrintSuccess("Switched to Devil's Advocate mode")
		return true
	case "5":
		es.Session.SwitchToPanel()
		es.Session.SetMode(types.ModeFramework)
		output.PrintSuccess("Switched to Framework mode")
		return true
	case "0", "":
		return false
	default:
		fmt.Println("Invalid selection")
		return false
	}
}

// GetPrompt returns the appropriate prompt based on current state.
func (es *EnhancedSession) GetPrompt() string {
	if es.Session.IsDiscoveryMode() {
		phase := es.FacilitationState.Phase
		switch phase {
		case facilitation.PhaseInit:
			return "start> "
		case facilitation.PhaseContextGathering:
			return "context> "
		case facilitation.PhaseProblemArticulation:
			return "problem> "
		case facilitation.PhaseDiscovery:
			return "discover> "
		case facilitation.PhaseReadyForEscalation:
			return "ready> "
		default:
			return "discover> "
		}
	}

	// Panel mode prompts
	if es.Session.SocraticState != nil && es.Session.SocraticState.Phase == "answering" {
		nextQ := len(es.Session.SocraticState.Answers) + 1
		return fmt.Sprintf("A%d> ", nextQ)
	}

	if es.Session.FrameworkState != nil && es.Session.FrameworkState.Phase == "confirming" {
		return "criteria> "
	}

	if es.Session.IsPendingBriefConfirmation() {
		return "confirm> "
	}

	// Mode-specific prompts
	switch es.Session.Mode {
	case types.ModeSocratic:
		return "socratic> "
	case types.ModeAdvocate:
		return "advocate> "
	case types.ModeFramework:
		return "framework> "
	default:
		return "cto> "
	}
}

// HasPendingSuggestion returns true if there's a pending mode suggestion.
func (es *EnhancedSession) HasPendingSuggestion() bool {
	return es.PendingSuggestion != nil
}

// ClearPendingSuggestion clears the pending suggestion.
func (es *EnhancedSession) ClearPendingSuggestion() {
	es.PendingSuggestion = nil
}

// AcceptSuggestion accepts the pending mode suggestion.
func (es *EnhancedSession) AcceptSuggestion() {
	if es.PendingSuggestion == nil {
		return
	}

	if es.PendingSuggestion.Mode != "" {
		es.Session.SetMode(es.PendingSuggestion.Mode)
		es.Session.SwitchToPanel()
	}

	es.ClearPendingSuggestion()
}

// ConfigureReadlineForModeSwitch configures readline to handle Ctrl+M.
func ConfigureReadlineForModeSwitch(rl *readline.Instance) {
	// Note: readline doesn't directly support Ctrl+M as it's typically Enter
	// We'll use Ctrl+] (0x1D) as the mode switch key instead
	// This is handled in the main loop
}

// PrintEnhancedWelcome prints the enhanced welcome message with mode selector.
func PrintEnhancedWelcome(projectCtx *types.CRFContext, lastDecision *types.DRFDocument) {
	output.PrintModeSelector(projectCtx, lastDecision)
	output.PrintModeSwitchHint()
}

// RunEnhancedDiscovery runs a single turn of enhanced discovery with the coordinator.
func (r *REPL) RunEnhancedDiscovery(es *EnhancedSession, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	fmt.Println()
	fmt.Println("Thinking...")

	result, err := es.ProcessWithCoordinator(ctx, message)
	if err != nil {
		output.PrintError(fmt.Sprintf("Error: %v", err))
		return err
	}

	fmt.Println()
	output.PrintFacilitatorMessage(result.Response)

	// Check if escalated
	if result.Escalated {
		fmt.Println()
		fmt.Println("─────────────────────────────────────────────────")
		fmt.Println("Discovery complete. Ready for panel discussion.")
		fmt.Println("Type your question or /panel to proceed.")
		fmt.Println("─────────────────────────────────────────────────")
	}

	return nil
}

// IntegrateEnhancedSession integrates the enhanced session into the REPL.
func (r *REPL) IntegrateEnhancedSession() *EnhancedSession {
	return NewEnhancedSession(r.client, r.projectCtx)
}

// HandleModeSwitch handles the mode switch command or key.
func HandleModeSwitch(rl *readline.Instance, es *EnhancedSession) {
	es.HandleModeSwitchKey()

	rl.SetPrompt("mode> ")
	line, err := rl.Readline()
	if err != nil {
		return
	}

	line = strings.TrimSpace(line)
	es.SwitchMode(line)
}
