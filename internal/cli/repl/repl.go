// Package repl implements a readline-based REPL for the CIO - Chief Intelligence Officer.
package repl

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chzyer/readline"

	"github.com/carlosinfantes/cio/internal/cli/output"
	"github.com/carlosinfantes/cio/internal/config"
	advisorsPkg "github.com/carlosinfantes/cio/internal/core/advisors"
	ctxLoader "github.com/carlosinfantes/cio/internal/core/context"
	"github.com/carlosinfantes/cio/internal/core/decisions"
	"github.com/carlosinfantes/cio/internal/core/discovery"
	"github.com/carlosinfantes/cio/internal/core/llm"
	"github.com/carlosinfantes/cio/internal/core/modes"
	"github.com/carlosinfantes/cio/internal/types"
)

// Session maintains state for an interactive session.
type Session struct {
	Mode         types.Mode
	Advisors     []types.AdvisorID
	Summaries    []string
	Decisions    []string
	LastDecision string

	SessionMode              types.SessionMode
	DiscoverySession         *types.DiscoverySession
	CurrentBrief             *types.Brief
	PendingBriefConfirmation bool

	// Socratic mode state
	SocraticState *types.SocraticState

	// Framework mode state
	FrameworkState *types.FrameworkState
}

// REPL is the readline-based interactive session.
type REPL struct {
	rl         *readline.Instance
	session    *Session
	cfg        types.Config
	client     *llm.Client
	projectCtx *types.CRFContext
}

// NewSession creates a new interactive session with default settings.
func NewSession() *Session {
	cfg, err := config.Load()
	if err != nil {
		return &Session{
			Mode:        types.ModePanel,
			Advisors:    []types.AdvisorID{types.AdvisorCTO, types.AdvisorCISO, types.AdvisorVPEng, types.AdvisorArchitect},
			Summaries:   []string{},
			Decisions:   []string{},
			SessionMode: types.SessionModeDiscovery,
		}
	}

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

// Session helper methods
func (s *Session) IsDiscoveryMode() bool {
	return s.SessionMode == types.SessionModeDiscovery
}

func (s *Session) IsPendingBriefConfirmation() bool {
	return s.PendingBriefConfirmation
}

func (s *Session) SetPendingBriefConfirmation(pending bool) {
	s.PendingBriefConfirmation = pending
}

func (s *Session) SwitchToPanel() {
	s.SessionMode = types.SessionModePanel
	s.PendingBriefConfirmation = false
	if s.DiscoverySession != nil {
		s.DiscoverySession.Status = types.DiscoveryStatusConverted
	}
}

func (s *Session) SwitchToDiscovery() {
	s.SessionMode = types.SessionModeDiscovery
	s.PendingBriefConfirmation = false
}

func (s *Session) StartDiscovery() {
	s.DiscoverySession = types.NewDiscoverySession()
	s.CurrentBrief = nil
}

func (s *Session) SetBrief(brief *types.Brief) {
	s.CurrentBrief = brief
	if s.DiscoverySession != nil {
		s.DiscoverySession.GeneratedBrief = brief
	}
}

func (s *Session) AddSummary(question, synthesis string) {
	summary := fmt.Sprintf("Q: %s → %s", truncate(question, 60), truncate(synthesis, 100))
	s.Summaries = append(s.Summaries, summary)
	if len(s.Summaries) > 5 {
		s.Summaries = s.Summaries[len(s.Summaries)-5:]
	}
}

func (s *Session) AddDecision(id string) {
	s.Decisions = append(s.Decisions, id)
	s.LastDecision = id
}

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

func (s *Session) SetMode(mode types.Mode) {
	s.Mode = mode
}

func (s *Session) SetAdvisors(advisors []types.AdvisorID) {
	s.Advisors = advisors
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

// New creates a new REPL instance.
func New() (*REPL, error) {
	// Build tab completer
	completer := readline.NewPrefixCompleter(
		readline.PcItem("/panel"),
		readline.PcItem("/discuss"),
		readline.PcItem("/discovery"),
		readline.PcItem("/skip"),
		readline.PcItem("/confirm"),
		readline.PcItem("/edit-brief",
			readline.PcItem("problem"),
			readline.PcItem("context"),
			readline.PcItem("constraints"),
			readline.PcItem("goals"),
			readline.PcItem("questions"),
		),
		readline.PcItem("/regen"),
		readline.PcItem("/brief"),
		readline.PcItem("/mode",
			readline.PcItem("panel"),
			readline.PcItem("socratic"),
			readline.PcItem("advocate"),
			readline.PcItem("framework"),
		),
		readline.PcItem("/advisors"),
		readline.PcItem("/history"),
		readline.PcItem("/tag"),
		readline.PcItem("/track"),
		readline.PcItem("/context"),
		readline.PcItem("/save-discovery"),
		readline.PcItem("/resume"),
		readline.PcItem("/help"),
		readline.PcItem("/quit"),
	)

	// History file
	homeDir, _ := os.UserHomeDir()
	historyFile := filepath.Join(homeDir, ".cio", "history")

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(historyFile), 0755)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "discover> ",
		AutoComplete:    completer,
		HistoryFile:     historyFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "quit",
	})
	if err != nil {
		return nil, err
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		cfg = types.DefaultConfig()
	}

	if cfg.APIKey == "" {
		rl.Close()
		return nil, fmt.Errorf("no API key configured. Run: cio init")
	}

	// Initialize LLM client
	client, err := llm.NewClient(cfg.APIKey, cfg.Model)
	if err != nil {
		rl.Close()
		return nil, fmt.Errorf("failed to initialize LLM client: %w", err)
	}

	// Load project context
	projectCtx, _ := ctxLoader.LoadCRFContext()

	return &REPL{
		rl:         rl,
		session:    NewSession(),
		cfg:        cfg,
		client:     client,
		projectCtx: projectCtx,
	}, nil
}

// Close closes the REPL.
func (r *REPL) Close() {
	r.rl.Close()
}

// Run starts the REPL loop.
func (r *REPL) Run() error {
	// Print welcome
	if r.session.IsDiscoveryMode() {
		output.PrintDiscoveryWelcome()
	} else {
		output.PrintPanelModeActive()
	}

	// Show context if loaded
	if r.projectCtx != nil {
		output.PrintContextLoaded(r.projectCtx)

		// Check for stale context
		warning := ctxLoader.CheckStaleness(r.projectCtx, r.cfg.ContextRefreshDays)
		output.PrintStalenessWarning(warning)

		// Check for context conflicts with recent decisions
		conflicts := ctxLoader.DetectConflicts(r.projectCtx)
		output.PrintConflictWarning(conflicts)
	}

	// Generate initial greeting in discovery mode
	if r.session.IsDiscoveryMode() {
		r.generateGreeting()
	}

	// Main loop
	for {
		// Update prompt based on mode
		if r.session.IsDiscoveryMode() {
			if r.session.IsPendingBriefConfirmation() {
				r.rl.SetPrompt("confirm> ")
			} else {
				r.rl.SetPrompt("discover> ")
			}
		} else if r.session.SocraticState != nil && r.session.SocraticState.Phase == "answering" {
			// Socratic answering mode
			nextQ := len(r.session.SocraticState.Answers) + 1
			r.rl.SetPrompt(fmt.Sprintf("A%d> ", nextQ))
		} else if r.session.FrameworkState != nil && r.session.FrameworkState.Phase == "confirming" {
			// Framework criteria confirmation mode
			r.rl.SetPrompt("criteria> ")
		} else {
			r.rl.SetPrompt("cto> ")
		}

		line, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			// Empty input in confirmation mode = confirm
			if r.session.IsPendingBriefConfirmation() {
				r.confirmBriefAndTransition()
			}
			// Empty input in Socratic mode = skip question
			if r.session.SocraticState != nil && r.session.SocraticState.Phase == "answering" {
				r.handleSocraticAnswer("")
			}
			// Empty input in Framework criteria mode = confirm criteria
			if r.session.FrameworkState != nil && r.session.FrameworkState.Phase == "confirming" {
				r.confirmFrameworkCriteria()
			}
			continue
		}

		// Handle commands
		if strings.HasPrefix(line, "/") {
			if r.handleCommand(line) {
				break // quit
			}
			continue
		}

		// Handle pending confirmation
		if r.session.IsPendingBriefConfirmation() {
			fmt.Println()
			fmt.Println("Use /confirm to proceed, or /edit-brief <field> to modify")
			output.PrintBriefConfirmationPrompt()
			continue
		}

		// Handle Socratic answering mode
		if r.session.SocraticState != nil && r.session.SocraticState.Phase == "answering" {
			r.handleSocraticAnswer(line)
			continue
		}

		// Process input based on mode
		if r.session.IsDiscoveryMode() {
			r.processDiscovery(line)
		} else {
			r.processPanel(line)
		}
	}

	fmt.Println("\nSession ended. Goodbye!")
	return nil
}

// handleCommand processes slash commands. Returns true if should quit.
func (r *REPL) handleCommand(line string) bool {
	parts := strings.Fields(line)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/quit", "/exit", "/q":
		return true

	case "/help", "/?":
		r.printHelp()

	case "/panel", "/discuss":
		r.handlePanelTransition()

	case "/skip":
		r.handleSkip()

	case "/discovery":
		r.handleReturnToDiscovery()

	case "/brief":
		r.handleBrief(parts)

	case "/confirm":
		if r.session.IsPendingBriefConfirmation() {
			r.confirmBriefAndTransition()
		} else if r.session.FrameworkState != nil && r.session.FrameworkState.Phase == "confirming" {
			r.confirmFrameworkCriteria()
		} else {
			fmt.Println("\nNo brief or criteria pending confirmation")
		}

	case "/add":
		r.handleFrameworkAdd(parts)

	case "/remove":
		r.handleFrameworkRemove(parts)

	case "/weight":
		r.handleFrameworkWeight(parts)

	case "/cancel":
		r.handleFrameworkCancel()

	case "/edit-brief", "/eb":
		r.handleEditBrief(parts)

	case "/regen", "/regenerate":
		r.handleRegenerateBrief()

	case "/save-discovery":
		r.handleSaveDiscovery(parts)

	case "/resume":
		r.handleResume(parts)

	case "/mode", "/m":
		r.handleMode(parts)

	case "/history", "/hist", "/h":
		r.printSessionHistory()

	case "/tag", "/t":
		r.handleTag(parts)

	case "/advisors", "/a":
		r.handleAdvisors(parts)

	case "/context", "/ctx", "/c":
		r.handleContext()

	case "/track":
		r.handleTrackOutcome()

	default:
		fmt.Printf("\nUnknown command: %s\n", cmd)
		fmt.Println("Type /help for available commands")
	}

	return false
}

func (r *REPL) printHelp() {
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println()
	fmt.Println("  /panel          - Transition from discovery to panel")
	fmt.Println("  /discovery      - Return to discovery mode")
	fmt.Println("  /skip           - Skip discovery, go to panel")
	fmt.Println("  /confirm        - Confirm brief and proceed")
	fmt.Println("  /edit-brief     - Edit brief field")
	fmt.Println("  /regen          - Regenerate brief")
	fmt.Println("  /mode <m>       - Change mode (panel/socratic/advocate/framework)")
	fmt.Println("  /advisors       - Show/change advisors")
	fmt.Println("  /history        - Show session history")
	fmt.Println("  /tag <tag>      - Tag last decision")
	fmt.Println("  /track          - Record outcome for last decision")
	fmt.Println("  /context        - Show loaded context")
	fmt.Println("  /help           - Show this help")
	fmt.Println("  /quit           - Exit")
	fmt.Println()
}

func (r *REPL) generateGreeting() {
	fmt.Println("Generating greeting...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	greeting, err := modes.DiscoveryGreeting(ctx, r.client, r.projectCtx)
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to generate greeting: %v", err))
		return
	}

	fmt.Println()
	output.PrintFacilitatorMessage(greeting)
}

func (r *REPL) processDiscovery(message string) {
	if r.session.DiscoverySession == nil {
		r.session.StartDiscovery()
	}

	fmt.Println()
	output.PrintUserMessage(message)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	fmt.Println()
	fmt.Println("Thinking...")

	resp, err := modes.Discovery(ctx, r.client, r.session.DiscoverySession, message, r.projectCtx)
	if err != nil {
		output.PrintError(fmt.Sprintf("Error: %v", err))
		return
	}

	fmt.Println()
	output.PrintFacilitatorMessage(resp.Content)
}

func (r *REPL) processPanel(question string) {
	// Check for long questions and summarize if needed
	if modes.IsLongQuestion(question) {
		summarized, ok := r.handleLongQuestion(question)
		if !ok {
			return // User cancelled
		}
		if summarized != "" {
			question = summarized
		}
	}

	// Socratic mode: start two-turn flow
	if r.session.Mode == types.ModeSocratic {
		r.startSocraticFlow(question)
		return
	}

	// Framework mode: start multi-turn flow
	if r.session.Mode == types.ModeFramework {
		r.startFrameworkFlow(question)
		return
	}

	r.executePanelQuery(question)
}

// startSocraticFlow generates clarifying questions for Socratic mode.
func (r *REPL) startSocraticFlow(question string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	fmt.Println()
	fmt.Println("Generating clarifying questions...")

	questions, err := modes.SocraticQuestions(ctx, r.client, question, r.projectCtx)
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to generate questions: %v", err))
		// Fall back to regular panel
		r.session.Mode = types.ModePanel
		r.executePanelQuery(question)
		return
	}

	if len(questions) == 0 {
		// No questions generated, proceed to panel
		fmt.Println("No clarifying questions needed.")
		r.executePanelQuery(question)
		return
	}

	// Store state and display questions
	r.session.SocraticState = types.NewSocraticState(question)
	r.session.SocraticState.Questions = questions
	r.session.SocraticState.Phase = "answering"

	output.PrintSocraticQuestions(questions)

	// Show first question prompt
	output.PrintSocraticAnswerPrompt(1, questions[0])
}

// handleSocraticAnswer processes an answer to a Socratic question.
func (r *REPL) handleSocraticAnswer(answer string) {
	state := r.session.SocraticState
	if state == nil {
		return
	}

	currentIdx := len(state.Answers)
	if answer != "" {
		state.Answers[currentIdx] = answer
	} else {
		state.Answers[currentIdx] = "(skipped)"
	}

	// Check if more questions
	nextIdx := currentIdx + 1
	if nextIdx < len(state.Questions) {
		output.PrintSocraticAnswerPrompt(nextIdx+1, state.Questions[nextIdx])
		return
	}

	// All questions answered, proceed to panel
	state.Phase = "complete"
	r.completeSocraticFlow()
}

// completeSocraticFlow executes the panel with enriched context.
func (r *REPL) completeSocraticFlow() {
	state := r.session.SocraticState
	if state == nil {
		return
	}

	activeAdvisors := advisorsPkg.GetByIDs(r.session.Advisors)

	if r.cfg.AutoSummonSpecialists {
		summonResults := advisorsPkg.SummonSpecialists(state.Question)
		if len(summonResults) > 0 {
			output.PrintSpecialistSummoned(summonResults)
			for _, sr := range summonResults {
				activeAdvisors = append(activeAdvisors, sr.Specialist)
			}
		}
	}

	if len(activeAdvisors) > r.cfg.MaxAdvisors {
		output.PrintAdvisorCapWarning(len(activeAdvisors), r.cfg.MaxAdvisors)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Println()
	fmt.Println("Consulting the advisory board with your clarifications...")

	parsed, err := modes.SocraticPanelWithAnswers(ctx, r.client, state, activeAdvisors, r.projectCtx)
	if err != nil {
		output.PrintError(fmt.Sprintf("Error: %v", err))
		r.session.SocraticState = nil
		return
	}

	fmt.Println()
	output.RenderTerminal(parsed)

	// Save decision
	advisorIDs := make([]types.AdvisorID, len(activeAdvisors))
	for i, a := range activeAdvisors {
		advisorIDs[i] = a.ID
	}

	doc := decisions.CreateDRFDocument(state.Question, types.ModeSocratic, advisorIDs, parsed, r.projectCtx)
	if err := decisions.SaveDRFDocument(doc); err != nil {
		output.PrintError(fmt.Sprintf("Failed to save decision: %v", err))
	} else {
		r.session.AddDecision(doc.Decision.ID)
		r.session.AddSummary(state.Question, parsed.Synthesis)
		output.PrintSuccess(fmt.Sprintf("Decision saved: %s", doc.Decision.ID))
	}

	// Clear state
	r.session.SocraticState = nil
}

// executePanelQuery runs the actual panel query (non-Socratic modes).
func (r *REPL) executePanelQuery(question string) {
	activeAdvisors := advisorsPkg.GetByIDs(r.session.Advisors)

	if r.cfg.AutoSummonSpecialists {
		summonResults := advisorsPkg.SummonSpecialists(question)
		if len(summonResults) > 0 {
			output.PrintSpecialistSummoned(summonResults)
			for _, sr := range summonResults {
				activeAdvisors = append(activeAdvisors, sr.Specialist)
			}
		}
	}

	// Warn if too many advisors
	if len(activeAdvisors) > r.cfg.MaxAdvisors {
		output.PrintAdvisorCapWarning(len(activeAdvisors), r.cfg.MaxAdvisors)
	}

	// Build session context with relevant past decisions
	sessionContext := r.session.GetContextSummary()

	// Include relevant past decisions
	relevantDecisions, _ := decisions.GetRelevantDecisions(question, 3)
	if len(relevantDecisions) > 0 {
		output.PrintRelevantDecisions(relevantDecisions)
		decisionContext := decisions.FormatDecisionContext(relevantDecisions)
		if sessionContext != "" {
			sessionContext += "\n" + decisionContext
		} else {
			sessionContext = decisionContext
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Println()
	fmt.Println("Consulting the advisory board...")

	var parsed types.ParsedResponse
	var err error

	if r.session.CurrentBrief != nil && len(r.session.Decisions) == 0 {
		parsed, err = modes.PanelWithBrief(ctx, r.client, r.session.CurrentBrief, activeAdvisors, r.projectCtx, r.session.Mode)
	} else {
		switch r.session.Mode {
		case types.ModeAdvocate:
			parsed, err = modes.AdvocateWithContext(ctx, r.client, question, activeAdvisors, r.projectCtx, sessionContext)
		case types.ModeFramework:
			parsed, err = modes.FrameworkWithContext(ctx, r.client, question, activeAdvisors, r.projectCtx, sessionContext)
		default:
			parsed, err = modes.PanelWithContext(ctx, r.client, question, activeAdvisors, r.projectCtx, sessionContext)
		}
	}

	if err != nil {
		output.PrintError(fmt.Sprintf("Error: %v", err))
		return
	}

	fmt.Println()
	output.RenderTerminal(parsed)

	advisorIDs := make([]types.AdvisorID, len(activeAdvisors))
	for i, a := range activeAdvisors {
		advisorIDs[i] = a.ID
	}

	doc := decisions.CreateDRFDocument(question, r.session.Mode, advisorIDs, parsed, r.projectCtx)
	if err := decisions.SaveDRFDocument(doc); err != nil {
		output.PrintError(fmt.Sprintf("Failed to save decision: %v", err))
	} else {
		r.session.AddDecision(doc.Decision.ID)
		r.session.AddSummary(question, parsed.Synthesis)
		output.PrintSuccess(fmt.Sprintf("Decision saved: %s", doc.Decision.ID))
	}

	// Check for context update suggestions based on question
	if r.projectCtx != nil {
		suggestions := ctxLoader.DetectUpdateSignals(question, r.projectCtx)
		for _, s := range suggestions {
			output.PrintUpdateSuggestion(&s)
		}
	}
}

func (r *REPL) handlePanelTransition() {
	if !r.session.IsDiscoveryMode() {
		fmt.Println("\nAlready in panel mode")
		return
	}

	if r.session.DiscoverySession == nil || len(r.session.DiscoverySession.Messages) < 2 {
		fmt.Println("\nPlease have at least one exchange with the facilitator before transitioning")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	fmt.Println()
	fmt.Println("Generating brief from discovery conversation...")

	brief, err := discovery.GenerateBrief(ctx, r.client, r.session.DiscoverySession)
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to generate brief: %v", err))
		return
	}

	r.session.SetBrief(brief)

	fmt.Println()
	output.PrintBrief(brief)

	suggestedAdvisors := discovery.SuggestAdvisorsFromBrief(brief)
	r.session.Advisors = suggestedAdvisors

	advisorNames := make([]string, len(suggestedAdvisors))
	for i, id := range suggestedAdvisors {
		advisorNames[i] = string(id)
	}

	fmt.Println()
	fmt.Printf("Suggested advisors: %s\n", strings.Join(advisorNames, ", "))
	fmt.Println()

	r.session.SetPendingBriefConfirmation(true)
	output.PrintBriefConfirmationPrompt()
}

func (r *REPL) confirmBriefAndTransition() {
	r.session.SetPendingBriefConfirmation(false)
	r.session.SwitchToPanel()

	fmt.Println()
	output.PrintSuccess("Brief confirmed. Transitioned to Panel Mode")
	fmt.Println("Ask questions directly, or use /discovery to return")
}

func (r *REPL) handleEditBrief(parts []string) {
	if !r.session.IsPendingBriefConfirmation() || r.session.CurrentBrief == nil {
		fmt.Println("\nNo brief pending confirmation")
		return
	}

	if len(parts) < 3 {
		fmt.Println("\nUsage: /edit-brief <field> \"new value\"")
		fmt.Println("Fields: problem, context, constraints, goals, questions")
		return
	}

	field := parts[1]
	value := strings.Join(parts[2:], " ")
	value = strings.Trim(value, "\"'")

	switch field {
	case "problem":
		r.session.CurrentBrief.ProblemStatement = value
	case "context":
		r.session.CurrentBrief.Context = value
	case "constraints":
		items := strings.Split(value, ",")
		for i := range items {
			items[i] = strings.TrimSpace(items[i])
		}
		r.session.CurrentBrief.Constraints = items
	case "goals":
		items := strings.Split(value, ",")
		for i := range items {
			items[i] = strings.TrimSpace(items[i])
		}
		r.session.CurrentBrief.Goals = items
	case "questions":
		items := strings.Split(value, ",")
		for i := range items {
			items[i] = strings.TrimSpace(items[i])
		}
		r.session.CurrentBrief.KeyQuestions = items
	default:
		fmt.Printf("\nUnknown field: %s\n", field)
		fmt.Println("Valid fields: problem, context, constraints, goals, questions")
		return
	}

	fmt.Println()
	output.PrintSuccess(fmt.Sprintf("Updated %s", field))
	fmt.Println()
	output.PrintBrief(r.session.CurrentBrief)
	fmt.Println()
	output.PrintBriefConfirmationPrompt()
}

func (r *REPL) handleRegenerateBrief() {
	if !r.session.IsPendingBriefConfirmation() || r.session.DiscoverySession == nil {
		fmt.Println("\nNo brief pending confirmation or no discovery session")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	fmt.Println()
	fmt.Println("Regenerating brief from discovery conversation...")

	brief, err := discovery.GenerateBrief(ctx, r.client, r.session.DiscoverySession)
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to regenerate brief: %v", err))
		return
	}

	r.session.SetBrief(brief)

	fmt.Println()
	output.PrintBrief(brief)
	fmt.Println()
	output.PrintBriefConfirmationPrompt()
}

func (r *REPL) handleSkip() {
	r.session.SwitchToPanel()
	r.session.DiscoverySession = nil
	r.session.CurrentBrief = nil

	fmt.Println()
	fmt.Println("Skipped discovery. Now in Panel mode.")
	fmt.Println("Ask questions directly, or use /discovery to start discovery.")
}

func (r *REPL) handleReturnToDiscovery() {
	if r.session.IsDiscoveryMode() && !r.session.IsPendingBriefConfirmation() {
		fmt.Println("\nAlready in discovery mode")
		return
	}

	wasPendingConfirmation := r.session.IsPendingBriefConfirmation()
	r.session.SwitchToDiscovery()

	if wasPendingConfirmation {
		fmt.Println("\nBrief confirmation cancelled. Returning to discovery mode...")
	}

	fmt.Println()
	output.PrintDiscoveryWelcome()

	if r.session.DiscoverySession != nil && len(r.session.DiscoverySession.Messages) > 0 {
		fmt.Println("Continuing previous discovery conversation...")
	} else {
		r.generateGreeting()
	}
}

func (r *REPL) handleBrief(parts []string) {
	if r.session.CurrentBrief == nil {
		fmt.Println("\nNo brief generated yet. Complete discovery first with /panel.")
		return
	}
	fmt.Println()
	output.PrintBrief(r.session.CurrentBrief)
}

func (r *REPL) handleMode(parts []string) {
	if len(parts) < 2 {
		fmt.Printf("\nCurrent mode: %s\n", r.session.Mode)
		fmt.Println("Usage: /mode <panel|socratic|advocate|framework>")
		return
	}

	newMode := types.Mode(parts[1])
	switch newMode {
	case types.ModePanel, types.ModeSocratic, types.ModeAdvocate, types.ModeFramework:
		r.session.SetMode(newMode)
		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Mode changed to: %s", newMode))
	default:
		fmt.Printf("\nInvalid mode: %s\n", parts[1])
		fmt.Println("Valid modes: panel, socratic, advocate, framework")
	}
}

func (r *REPL) printSessionHistory() {
	fmt.Println()
	if len(r.session.Decisions) == 0 {
		fmt.Println("No decisions in this session yet")
		return
	}

	fmt.Println("Session Decisions:")
	for _, id := range r.session.Decisions {
		doc, err := decisions.GetDRFDocument(id)
		if err != nil || doc == nil {
			fmt.Printf("  %s\n", id)
			continue
		}
		fmt.Printf("  %s [%s]\n", id, doc.Meta.Status)
	}
}

func (r *REPL) handleTag(parts []string) {
	if len(parts) < 2 {
		fmt.Println("\nUsage: /tag <tag>")
		return
	}
	if r.session.LastDecision == "" {
		fmt.Println("\nNo decision to tag yet")
		return
	}
	tag := strings.Join(parts[1:], " ")
	if err := decisions.AddTag(r.session.LastDecision, tag); err != nil {
		output.PrintError(fmt.Sprintf("Failed to add tag: %v", err))
	} else {
		output.PrintSuccess(fmt.Sprintf("Tagged %s with: %s", r.session.LastDecision, tag))
	}
}

func (r *REPL) handleAdvisors(parts []string) {
	if len(parts) < 2 {
		advisorStrs := make([]string, len(r.session.Advisors))
		for i, a := range r.session.Advisors {
			advisorStrs[i] = string(a)
		}
		fmt.Printf("\nCurrent advisors: %s\n", strings.Join(advisorStrs, ", "))
		fmt.Println("Usage: /advisors <ids> (e.g., /advisors cto,ciso,architect)")
		return
	}

	ids := strings.Split(parts[1], ",")
	advisorIDs := make([]types.AdvisorID, len(ids))
	for i, id := range ids {
		advisorIDs[i] = types.AdvisorID(strings.TrimSpace(id))
	}
	r.session.SetAdvisors(advisorIDs)
	output.PrintSuccess(fmt.Sprintf("Advisors set to: %s", parts[1]))
}

func (r *REPL) handleContext() {
	projectCtx, err := ctxLoader.LoadCRFContext()
	if err != nil {
		output.PrintError(fmt.Sprintf("Loading context: %v", err))
		return
	}
	fmt.Println()
	if projectCtx != nil {
		output.PrintContextLoaded(projectCtx)
	} else {
		fmt.Println("No context loaded")
	}
}

func (r *REPL) handleSaveDiscovery(parts []string) {
	if r.session.DiscoverySession == nil {
		fmt.Println("\nNo discovery session to save")
		return
	}

	name := ""
	if len(parts) > 1 {
		name = strings.Join(parts[1:], "-")
	}

	id, err := discovery.SaveSession(r.session.DiscoverySession, name)
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to save session: %v", err))
	} else {
		output.PrintSuccess(fmt.Sprintf("Discovery session saved: %s", id))
	}
}

func (r *REPL) handleResume(parts []string) {
	if len(parts) < 2 {
		sessions, err := discovery.ListSessions()
		if err != nil {
			output.PrintError(fmt.Sprintf("Failed to list sessions: %v", err))
			return
		}
		if len(sessions) == 0 {
			fmt.Println("\nNo saved discovery sessions")
			return
		}
		fmt.Println("\nSaved discovery sessions:")
		for _, s := range sessions {
			fmt.Printf("  %s (created: %s)\n", s.ID, s.CreatedAt.Format("2006-01-02 15:04"))
		}
		fmt.Println("\nUsage: /resume <session-id>")
		return
	}

	session, err := discovery.LoadSession(parts[1])
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to load session: %v", err))
		return
	}

	r.session.DiscoverySession = session
	r.session.SwitchToDiscovery()
	output.PrintSuccess(fmt.Sprintf("Resumed session: %s", parts[1]))

	// Show conversation history
	for _, msg := range session.Messages {
		if msg.Role == "facilitator" {
			output.PrintFacilitatorMessage(msg.Content)
		} else {
			output.PrintUserMessage(msg.Content)
		}
	}
}

// startFrameworkFlow begins the Framework decision flow.
func (r *REPL) startFrameworkFlow(question string) {
	// Parse options from question
	options := llm.ParseFrameworkOptions(question)
	if len(options) < 2 {
		// No options found, fall back to regular panel
		fmt.Println()
		output.PrintInfo("No comparison detected (e.g., 'A vs B'). Using regular panel mode.")
		r.session.Mode = types.ModePanel
		r.executePanelQuery(question)
		return
	}

	output.PrintFrameworkOptions(options)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	fmt.Println()
	fmt.Println("Generating evaluation criteria...")

	criteria, err := modes.FrameworkCriteria(ctx, r.client, question, options, r.projectCtx)
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to generate criteria: %v", err))
		r.session.Mode = types.ModePanel
		r.executePanelQuery(question)
		return
	}

	if len(criteria) == 0 {
		output.PrintError("No criteria generated. Falling back to panel mode.")
		r.session.Mode = types.ModePanel
		r.executePanelQuery(question)
		return
	}

	// Store state
	r.session.FrameworkState = types.NewFrameworkState(question, options)
	r.session.FrameworkState.Criteria = criteria
	r.session.FrameworkState.Phase = "confirming"

	output.PrintFrameworkCriteria(criteria)
	output.PrintFrameworkCriteriaPrompt()
}

// confirmFrameworkCriteria proceeds to scoring after criteria are confirmed.
func (r *REPL) confirmFrameworkCriteria() {
	state := r.session.FrameworkState
	if state == nil || len(state.Criteria) == 0 {
		fmt.Println("\nNo criteria to confirm")
		return
	}

	state.Phase = "scoring"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Println()
	fmt.Println("Scoring options against criteria...")

	err := modes.FrameworkScoring(ctx, r.client, state, r.projectCtx)
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to score options: %v", err))
		r.session.FrameworkState = nil
		return
	}

	state.Phase = "complete"

	// Display results
	output.PrintFrameworkMatrix(state)
	output.PrintFrameworkRecommendation(state)

	// Save decision as DRF document
	frameworkParsed := types.ParsedResponse{
		Synthesis: state.Recommendation,
	}
	doc := decisions.CreateDRFDocument(state.Question, types.ModeFramework, r.session.Advisors, frameworkParsed, r.projectCtx)
	if err := decisions.SaveDRFDocument(doc); err != nil {
		output.PrintError(fmt.Sprintf("Failed to save decision: %v", err))
	} else {
		r.session.AddDecision(doc.Decision.ID)
		r.session.AddSummary(state.Question, "Framework: "+state.Recommendation)
		output.PrintSuccess(fmt.Sprintf("Decision saved: %s", doc.Decision.ID))
	}

	// Clear state
	r.session.FrameworkState = nil
}

// handleFrameworkAdd adds a criterion during criteria confirmation.
func (r *REPL) handleFrameworkAdd(parts []string) {
	state := r.session.FrameworkState
	if state == nil || state.Phase != "confirming" {
		fmt.Println("\nNot in criteria confirmation mode")
		return
	}

	if len(parts) < 2 {
		fmt.Println("\nUsage: /add <criterion name>")
		return
	}

	name := strings.Join(parts[1:], " ")
	state.Criteria = append(state.Criteria, types.FrameworkCriterion{
		Name:        name,
		Description: "User-added criterion",
		Weight:      3,
	})

	output.PrintSuccess(fmt.Sprintf("Added criterion: %s", name))
	output.PrintFrameworkCriteria(state.Criteria)
	output.PrintFrameworkCriteriaPrompt()
}

// handleFrameworkRemove removes a criterion during criteria confirmation.
func (r *REPL) handleFrameworkRemove(parts []string) {
	state := r.session.FrameworkState
	if state == nil || state.Phase != "confirming" {
		fmt.Println("\nNot in criteria confirmation mode")
		return
	}

	if len(parts) < 2 {
		fmt.Println("\nUsage: /remove <number>")
		return
	}

	var idx int
	fmt.Sscanf(parts[1], "%d", &idx)
	idx-- // Convert to 0-indexed

	if idx < 0 || idx >= len(state.Criteria) {
		fmt.Printf("\nInvalid criterion number: %s\n", parts[1])
		return
	}

	removed := state.Criteria[idx].Name
	state.Criteria = append(state.Criteria[:idx], state.Criteria[idx+1:]...)

	output.PrintSuccess(fmt.Sprintf("Removed criterion: %s", removed))
	output.PrintFrameworkCriteria(state.Criteria)
	output.PrintFrameworkCriteriaPrompt()
}

// handleFrameworkWeight changes a criterion's weight.
func (r *REPL) handleFrameworkWeight(parts []string) {
	state := r.session.FrameworkState
	if state == nil || state.Phase != "confirming" {
		fmt.Println("\nNot in criteria confirmation mode")
		return
	}

	if len(parts) < 3 {
		fmt.Println("\nUsage: /weight <criterion number> <1-5>")
		return
	}

	var idx, weight int
	fmt.Sscanf(parts[1], "%d", &idx)
	fmt.Sscanf(parts[2], "%d", &weight)
	idx-- // Convert to 0-indexed

	if idx < 0 || idx >= len(state.Criteria) {
		fmt.Printf("\nInvalid criterion number: %s\n", parts[1])
		return
	}

	if weight < 1 || weight > 5 {
		fmt.Println("\nWeight must be between 1 and 5")
		return
	}

	state.Criteria[idx].Weight = float64(weight)

	output.PrintSuccess(fmt.Sprintf("Updated weight for %s to %d", state.Criteria[idx].Name, weight))
	output.PrintFrameworkCriteria(state.Criteria)
	output.PrintFrameworkCriteriaPrompt()
}

// handleFrameworkCancel cancels the Framework flow.
func (r *REPL) handleFrameworkCancel() {
	if r.session.FrameworkState == nil {
		fmt.Println("\nNo Framework session to cancel")
		return
	}

	r.session.FrameworkState = nil
	fmt.Println("\nFramework mode cancelled")
}

// handleLongQuestion summarizes a long question and gets user confirmation.
// Returns the summarized question (or empty for original) and whether to proceed.
func (r *REPL) handleLongQuestion(question string) (string, bool) {
	output.PrintLongQuestionWarning(len(question))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	summary, err := modes.SummarizeLongQuestion(ctx, r.client, question)
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to summarize: %v", err))
		fmt.Println("Proceeding with original question...")
		return "", true
	}

	// Clean up summary (remove any trailing whitespace/newlines)
	summary = strings.TrimSpace(summary)

	output.PrintQuestionSummary(question, summary)
	output.PrintSummaryConfirmPrompt()

	// Get confirmation
	r.rl.SetPrompt("Use summary? [Y/n/edit]: ")
	response, err := r.rl.Readline()
	if err != nil {
		return "", false
	}

	response = strings.TrimSpace(strings.ToLower(response))

	switch response {
	case "", "y", "yes":
		output.PrintSuccess("Using summarized question")
		return summary, true
	case "n", "no":
		output.PrintInfo("Using original question")
		return "", true
	case "/edit", "edit", "e":
		// Let user edit the summary
		r.rl.SetPrompt("Edit summary: ")
		edited, err := r.rl.Readline()
		if err != nil {
			return "", false
		}
		edited = strings.TrimSpace(edited)
		if edited == "" {
			output.PrintInfo("Using original question")
			return "", true
		}
		output.PrintSuccess("Using edited summary")
		return edited, true
	default:
		// Treat any other input as using the summary
		output.PrintSuccess("Using summarized question")
		return summary, true
	}
}

// handleTrackOutcome records the outcome of the last decision.
func (r *REPL) handleTrackOutcome() {
	if r.session.LastDecision == "" {
		fmt.Println("\nNo decision to track. Ask a question first.")
		return
	}

	doc, err := decisions.GetDRFDocument(r.session.LastDecision)
	if err != nil || doc == nil {
		output.PrintError("Could not load decision")
		return
	}

	// In DRF, outcome is captured in Synthesis.Decision - check if already approved
	if doc.Meta.Status == types.DRFStatusApproved {
		fmt.Println("\nThis decision is already approved.")
		fmt.Printf("Decision: %s\n", doc.Synthesis.Decision)
		return
	}

	fmt.Println()
	fmt.Println("Recording outcome for:", doc.Decision.Title)
	fmt.Println()

	// Collect outcome details using readline
	r.rl.SetPrompt("Final decision (what did you decide?): ")
	choice, err := r.rl.Readline()
	if err != nil {
		return
	}
	choice = strings.TrimSpace(choice)
	if choice == "" {
		fmt.Println("Cancelled - no decision entered")
		return
	}

	r.rl.SetPrompt("Rationale (why this choice?): ")
	rationale, _ := r.rl.Readline()
	rationale = strings.TrimSpace(rationale)

	r.rl.SetPrompt("Confidence (0-100, Enter for 75): ")
	confStr, _ := r.rl.Readline()
	confidence := 75
	if confStr = strings.TrimSpace(confStr); confStr != "" {
		fmt.Sscanf(confStr, "%d", &confidence)
		if confidence < 0 {
			confidence = 0
		}
		if confidence > 100 {
			confidence = 100
		}
	}

	// Update the DRF document
	if err := decisions.UpdateSynthesis(r.session.LastDecision, choice, rationale); err != nil {
		output.PrintError(fmt.Sprintf("Failed to update synthesis: %v", err))
		return
	}

	if err := decisions.SetConfidence(r.session.LastDecision, confidence); err != nil {
		output.PrintError(fmt.Sprintf("Failed to set confidence: %v", err))
		return
	}

	if err := decisions.ApproveDecision(r.session.LastDecision); err != nil {
		output.PrintError(fmt.Sprintf("Failed to approve decision: %v", err))
		return
	}

	output.PrintOutcomeRecorded(r.session.LastDecision)
}
