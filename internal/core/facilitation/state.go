// Package facilitation implements Jordan's facilitation state machine.
// Jordan is a pure facilitator who ensures all necessary information is collected
// before escalating to the advisory panel.
package facilitation

import (
	"strings"
	"time"

	"github.com/carlosinfantes/cio/internal/types"
)

// FacilitationPhase represents the current phase of facilitation.
type FacilitationPhase string

const (
	// PhaseInit - Initial state, no context or problem yet
	PhaseInit FacilitationPhase = "init"
	// PhaseContextGathering - Gathering organizational context
	PhaseContextGathering FacilitationPhase = "context_gathering"
	// PhaseProblemArticulation - User is articulating the problem
	PhaseProblemArticulation FacilitationPhase = "problem_articulation"
	// PhaseDiscovery - Clarifying questions and discovery
	PhaseDiscovery FacilitationPhase = "discovery"
	// PhaseReadyForEscalation - All info gathered, ready for panel
	PhaseReadyForEscalation FacilitationPhase = "ready_for_escalation"
	// PhaseEscalated - Panel has been invoked
	PhaseEscalated FacilitationPhase = "escalated"
)

// FacilitationState tracks Jordan's facilitation progress.
type FacilitationState struct {
	Phase FacilitationPhase `yaml:"phase" json:"phase"`

	// Completeness flags
	ContextComplete    bool `yaml:"context_complete" json:"context_complete"`
	ProblemArticulated bool `yaml:"problem_articulated" json:"problem_articulated"`
	DiscoveryComplete  bool `yaml:"discovery_complete" json:"discovery_complete"`

	// Gathered information
	ProblemStatement string   `yaml:"problem_statement,omitempty" json:"problem_statement,omitempty"`
	KeyConstraints   []string `yaml:"key_constraints,omitempty" json:"key_constraints,omitempty"`
	Goals            []string `yaml:"goals,omitempty" json:"goals,omitempty"`
	ClarifyingQAs    []QAPair `yaml:"clarifying_qas,omitempty" json:"clarifying_qas,omitempty"`

	// Context validation
	ContextStaleDays    int      `yaml:"context_stale_days,omitempty" json:"context_stale_days,omitempty"`
	MissingContextTypes []string `yaml:"missing_context_types,omitempty" json:"missing_context_types,omitempty"`

	// Metadata
	StartedAt   time.Time `yaml:"started_at" json:"started_at"`
	UpdatedAt   time.Time `yaml:"updated_at" json:"updated_at"`
	TurnCount   int       `yaml:"turn_count" json:"turn_count"`
	EscalatedAt time.Time `yaml:"escalated_at,omitempty" json:"escalated_at,omitempty"`

	// Suggested next actions
	SuggestedMode     types.Mode        `yaml:"suggested_mode,omitempty" json:"suggested_mode,omitempty"`
	SuggestedAdvisors []types.AdvisorID `yaml:"suggested_advisors,omitempty" json:"suggested_advisors,omitempty"`
}

// QAPair represents a clarifying question and its answer.
type QAPair struct {
	Question string `yaml:"question" json:"question"`
	Answer   string `yaml:"answer" json:"answer"`
}

// NewFacilitationState creates a new facilitation state.
func NewFacilitationState() *FacilitationState {
	now := time.Now()
	return &FacilitationState{
		Phase:     PhaseInit,
		StartedAt: now,
		UpdatedAt: now,
	}
}

// ReadyForEscalation returns true if all required information is gathered.
func (fs *FacilitationState) ReadyForEscalation() bool {
	return fs.ContextComplete && fs.ProblemArticulated && fs.DiscoveryComplete
}

// UpdatePhase automatically transitions to the appropriate phase based on state.
func (fs *FacilitationState) UpdatePhase() {
	fs.UpdatedAt = time.Now()

	// Check for escalation readiness first
	if fs.ReadyForEscalation() {
		if fs.Phase != PhaseEscalated {
			fs.Phase = PhaseReadyForEscalation
		}
		return
	}

	// Determine current phase based on completeness
	if !fs.ContextComplete {
		fs.Phase = PhaseContextGathering
		return
	}

	if !fs.ProblemArticulated {
		fs.Phase = PhaseProblemArticulation
		return
	}

	if !fs.DiscoveryComplete {
		fs.Phase = PhaseDiscovery
		return
	}
}

// MarkContextComplete marks context gathering as complete.
func (fs *FacilitationState) MarkContextComplete(staleDays int, missingTypes []string) {
	fs.ContextComplete = true
	fs.ContextStaleDays = staleDays
	fs.MissingContextTypes = missingTypes
	fs.UpdatePhase()
}

// MarkProblemArticulated marks that the user has stated their problem.
func (fs *FacilitationState) MarkProblemArticulated(problem string) {
	fs.ProblemArticulated = true
	fs.ProblemStatement = problem
	fs.UpdatePhase()
}

// AddClarifyingQA adds a clarifying question-answer pair.
func (fs *FacilitationState) AddClarifyingQA(question, answer string) {
	fs.ClarifyingQAs = append(fs.ClarifyingQAs, QAPair{
		Question: question,
		Answer:   answer,
	})
	fs.TurnCount++
	fs.UpdatedAt = time.Now()
}

// MarkDiscoveryComplete marks discovery as complete.
func (fs *FacilitationState) MarkDiscoveryComplete() {
	fs.DiscoveryComplete = true
	fs.UpdatePhase()
}

// MarkEscalated marks that the panel has been invoked.
func (fs *FacilitationState) MarkEscalated() {
	fs.Phase = PhaseEscalated
	fs.EscalatedAt = time.Now()
	fs.UpdatedAt = time.Now()
}

// SetConstraints sets the key constraints identified during facilitation.
func (fs *FacilitationState) SetConstraints(constraints []string) {
	fs.KeyConstraints = constraints
	fs.UpdatedAt = time.Now()
}

// SetGoals sets the goals identified during facilitation.
func (fs *FacilitationState) SetGoals(goals []string) {
	fs.Goals = goals
	fs.UpdatedAt = time.Now()
}

// SuggestMode suggests an appropriate mode based on the problem characteristics.
func (fs *FacilitationState) SuggestMode(problemText string) types.Mode {
	lower := strings.ToLower(problemText)

	// Framework mode for comparisons
	if containsComparison(lower) {
		fs.SuggestedMode = types.ModeFramework
		return types.ModeFramework
	}

	// Advocate mode for validation
	if containsValidation(lower) {
		fs.SuggestedMode = types.ModeAdvocate
		return types.ModeAdvocate
	}

	// Default to panel
	fs.SuggestedMode = types.ModePanel
	return types.ModePanel
}

// SuggestAdvisors suggests relevant advisors based on problem content.
func (fs *FacilitationState) SuggestAdvisors(problemText string) []types.AdvisorID {
	lower := strings.ToLower(problemText)
	advisors := []types.AdvisorID{types.AdvisorCTO, types.AdvisorArchitect}

	// Security concerns
	if containsAny(lower, []string{"security", "compliance", "risk", "breach", "vulnerability", "encryption", "auth", "soc2", "gdpr", "hipaa"}) {
		advisors = append(advisors, types.AdvisorCISO)
	}

	// Team/execution concerns
	if containsAny(lower, []string{"team", "capacity", "hire", "velocity", "delivery", "process", "sprint", "morale", "burnout"}) {
		advisors = append(advisors, types.AdvisorVPEng)
	}

	// Financial concerns
	if containsAny(lower, []string{"budget", "cost", "roi", "pricing", "revenue", "expense", "investment", "financial"}) {
		advisors = append(advisors, types.AdvisorCFO)
	}

	// Product concerns
	if containsAny(lower, []string{"feature", "user", "customer", "product", "roadmap", "mvp", "launch", "ux", "ui"}) {
		advisors = append(advisors, types.AdvisorProduct)
	}

	// Infrastructure concerns
	if containsAny(lower, []string{"kubernetes", "k8s", "deploy", "aws", "gcp", "azure", "docker", "ci/cd", "infrastructure", "devops"}) {
		advisors = append(advisors, types.AdvisorDevOps)
	}

	fs.SuggestedAdvisors = advisors
	return advisors
}

// GetPhasePrompt returns Jordan's prompt based on current phase.
func (fs *FacilitationState) GetPhasePrompt() string {
	switch fs.Phase {
	case PhaseInit:
		return "Let me learn about your situation. What's the technical decision or challenge you're facing?"
	case PhaseContextGathering:
		return "I'd like to understand your context better. Can you tell me about your organization's current technical landscape?"
	case PhaseProblemArticulation:
		return "What's the specific decision or problem you need guidance on?"
	case PhaseDiscovery:
		return "Let me ask a few clarifying questions to make sure I understand fully..."
	case PhaseReadyForEscalation:
		return "I have everything I need. Ready to bring in the advisory panel?"
	case PhaseEscalated:
		return "The advisory panel is now engaged."
	default:
		return "How can I help you today?"
	}
}

// ToBrief converts the facilitation state to a Brief for panel consumption.
func (fs *FacilitationState) ToBrief() *types.Brief {
	// Build context from Q&A pairs
	var contextParts []string
	for _, qa := range fs.ClarifyingQAs {
		contextParts = append(contextParts, qa.Question+": "+qa.Answer)
	}

	return &types.Brief{
		ProblemStatement:  fs.ProblemStatement,
		Context:           strings.Join(contextParts, "\n"),
		Constraints:       fs.KeyConstraints,
		Goals:             fs.Goals,
		SuggestedAdvisors: fs.SuggestedAdvisors,
	}
}

// Helper functions

func containsComparison(text string) bool {
	patterns := []string{" vs ", " versus ", " or ", "compare", "which one", "what should we choose", "option a", "option b"}
	return containsAny(text, patterns)
}

func containsValidation(text string) bool {
	patterns := []string{"validate", "review", "challenge", "is this right", "am i missing", "blind spot", "devil's advocate"}
	return containsAny(text, patterns)
}

func containsAny(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}
