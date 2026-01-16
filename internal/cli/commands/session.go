// Package commands implements session state for interactive mode.
package commands

import (
	"fmt"
	"strings"

	"github.com/carlosinfantes/cto-advisory-board/internal/config"
	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

// Session maintains state for an interactive session.
type Session struct {
	Mode         types.Mode
	Advisors     []types.AdvisorID
	Summaries    []string // Q&A summaries for context
	Decisions    []string // Decision IDs from this session
	LastDecision string   // Most recent decision ID (for /tag)

	// Discovery mode fields
	SessionMode              types.SessionMode        // discovery or panel
	DiscoverySession         *types.DiscoverySession  // Current discovery conversation
	CurrentBrief             *types.Brief             // Generated brief from discovery
	PendingBriefConfirmation bool                     // True when awaiting brief confirmation
}

// NewSession creates a new interactive session with default settings.
func NewSession() *Session {
	cfg, err := config.Load()
	if err != nil {
		// Use defaults if config can't be loaded
		return &Session{
			Mode:        types.ModePanel,
			Advisors:    []types.AdvisorID{types.AdvisorCTO, types.AdvisorCISO, types.AdvisorVPEng, types.AdvisorArchitect},
			Summaries:   []string{},
			Decisions:   []string{},
			SessionMode: types.SessionModeDiscovery, // Default to discovery mode
		}
	}

	// Determine initial session mode based on config
	sessionMode := types.SessionModeDiscovery
	if !cfg.StartInDiscovery {
		sessionMode = types.SessionModePanel
	}

	return &Session{
		Mode:        cfg.DefaultMode,
		Advisors:    cfg.DefaultAdvisors,
		Summaries:   []string{},
		Decisions:   []string{},
		SessionMode: sessionMode,
	}
}

// IsDiscoveryMode returns true if in discovery mode.
func (s *Session) IsDiscoveryMode() bool {
	return s.SessionMode == types.SessionModeDiscovery
}

// IsPendingBriefConfirmation returns true if awaiting brief confirmation.
func (s *Session) IsPendingBriefConfirmation() bool {
	return s.PendingBriefConfirmation
}

// SetPendingBriefConfirmation sets the pending confirmation state.
func (s *Session) SetPendingBriefConfirmation(pending bool) {
	s.PendingBriefConfirmation = pending
}

// SwitchToPanel transitions from discovery to panel mode.
func (s *Session) SwitchToPanel() {
	s.SessionMode = types.SessionModePanel
	s.PendingBriefConfirmation = false
	if s.DiscoverySession != nil {
		s.DiscoverySession.Status = types.DiscoveryStatusConverted
	}
}

// SwitchToDiscovery transitions back to discovery mode.
func (s *Session) SwitchToDiscovery() {
	s.SessionMode = types.SessionModeDiscovery
	s.PendingBriefConfirmation = false
}

// StartDiscovery initializes a new discovery session.
func (s *Session) StartDiscovery() {
	s.DiscoverySession = types.NewDiscoverySession()
	s.CurrentBrief = nil
}

// SetBrief sets the current brief from discovery.
func (s *Session) SetBrief(brief *types.Brief) {
	s.CurrentBrief = brief
	if s.DiscoverySession != nil {
		s.DiscoverySession.GeneratedBrief = brief
	}
}

// AddSummary adds a Q&A summary to the session context.
func (s *Session) AddSummary(question, synthesis string) {
	// Create a brief summary (truncate if needed)
	summary := createSummary(question, synthesis)
	s.Summaries = append(s.Summaries, summary)

	// Keep only last 5 summaries to manage token usage
	if len(s.Summaries) > 5 {
		s.Summaries = s.Summaries[len(s.Summaries)-5:]
	}
}

// AddDecision records a decision ID from this session.
func (s *Session) AddDecision(id string) {
	s.Decisions = append(s.Decisions, id)
	s.LastDecision = id
}

// GetContextSummary returns the session context for prompt injection.
func (s *Session) GetContextSummary() string {
	if len(s.Summaries) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Previous discussion in this session:\n")
	for i, summary := range s.Summaries {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, summary))
	}
	return sb.String()
}

// SetMode changes the current interaction mode.
func (s *Session) SetMode(mode types.Mode) {
	s.Mode = mode
}

// SetAdvisors changes the active advisors.
func (s *Session) SetAdvisors(advisors []types.AdvisorID) {
	s.Advisors = advisors
}

// createSummary generates a brief summary from Q&A.
func createSummary(question, synthesis string) string {
	// Truncate question if too long
	q := question
	if len(q) > 60 {
		q = q[:57] + "..."
	}

	// Extract key point from synthesis (first sentence or truncate)
	synth := synthesis
	if idx := strings.Index(synth, "."); idx > 0 && idx < 150 {
		synth = synth[:idx+1]
	} else if len(synth) > 100 {
		synth = synth[:97] + "..."
	}

	return fmt.Sprintf("Q: %s → %s", q, synth)
}
