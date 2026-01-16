// Package output handles terminal rendering for the CTO Advisory Board.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

// Colors for each advisor
var advisorColors = map[types.AdvisorID]lipgloss.Color{
	types.AdvisorCTO:         lipgloss.Color("39"),  // Blue
	types.AdvisorCISO:        lipgloss.Color("196"), // Red
	types.AdvisorVPEng:       lipgloss.Color("46"),  // Green
	types.AdvisorArchitect:   lipgloss.Color("201"), // Magenta
	types.AdvisorCFO:         lipgloss.Color("226"), // Yellow
	types.AdvisorProduct:     lipgloss.Color("51"),  // Cyan
	types.AdvisorDevOps:      lipgloss.Color("255"), // White
	types.AdvisorFacilitator: lipgloss.Color("141"), // Light Purple
}

// Emojis for each advisor
var advisorEmojis = map[types.AdvisorID]string{
	types.AdvisorCTO:         "🎯",
	types.AdvisorCISO:        "🛡️",
	types.AdvisorVPEng:       "⚡",
	types.AdvisorArchitect:   "🏗️",
	types.AdvisorCFO:         "💰",
	types.AdvisorProduct:     "📱",
	types.AdvisorDevOps:      "🔧",
	types.AdvisorFacilitator: "💭",
}

// getTerminalWidth returns the current terminal width, or a default.
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width < 40 {
		return 80
	}
	if width > 120 {
		return 120
	}
	return width
}

// RenderTerminal outputs the response in styled terminal format.
func RenderTerminal(parsed types.ParsedResponse) {
	width := getTerminalWidth()
	boxWidth := width - 4 // Account for borders and padding

	// Render each advisor's response
	for _, advisor := range parsed.Advisors {
		renderAdvisorBox(advisor, boxWidth)
		fmt.Println()
	}

	// Render synthesis
	if parsed.Synthesis != "" {
		renderSynthesisBox(parsed.Synthesis, boxWidth)
	}
}

func renderAdvisorBox(advisor types.AdvisorResponse, width int) {
	color := advisorColors[advisor.AdvisorID]
	if color == "" {
		color = lipgloss.Color("250") // Default gray
	}

	emoji := advisorEmojis[advisor.AdvisorID]
	if emoji == "" {
		emoji = "💬"
	}

	// Title style
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(color)

	// Box style - no Width constraint to avoid rendering issues with unicode
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 1)

	// Build title
	title := fmt.Sprintf("%s %s — %s", emoji, advisor.Name, advisor.Role)

	// Wrap content
	content := wrapText(advisor.Response, width-4)

	// Render
	fmt.Println(titleStyle.Render(title))
	fmt.Println(boxStyle.Render(content))
}

func renderSynthesisBox(synthesis string, width int) {
	// Synthesis uses double border in green
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("46"))

	// No Width constraint to avoid rendering issues with unicode
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("46")).
		Padding(0, 1)

	title := "📋 SYNTHESIS"
	content := wrapText(synthesis, width-4)

	fmt.Println(titleStyle.Render(title))
	fmt.Println(boxStyle.Render(content))
}

// RenderMarkdown outputs the response in markdown format.
func RenderMarkdown(parsed types.ParsedResponse) {
	for _, advisor := range parsed.Advisors {
		fmt.Printf("## %s — %s\n\n", advisor.Name, advisor.Role)
		fmt.Println(advisor.Response)
		fmt.Println()
	}

	if parsed.Synthesis != "" {
		fmt.Println("## Synthesis")
		fmt.Println()
		fmt.Printf("> %s\n", strings.ReplaceAll(parsed.Synthesis, "\n", "\n> "))
	}
}

// RenderJSON outputs the response in JSON format.
func RenderJSON(parsed types.ParsedResponse) {
	output := map[string]interface{}{
		"advisors":  parsed.Advisors,
		"synthesis": parsed.Synthesis,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// wrapText wraps text to fit within the given width.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) <= width {
				currentLine += " " + word
			} else {
				result.WriteString(currentLine)
				result.WriteString("\n")
				currentLine = word
			}
		}
		result.WriteString(currentLine)
	}

	return result.String()
}

// Context loading message
func PrintContextLoaded(ctx *types.CRFContext) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46"))

	// Get organization info
	orgName := "Unknown"
	orgSize := ""
	teamCount := 0

	if org := ctx.GetOrganization(); org != nil {
		orgName = org.Name
		if size, ok := org.Attributes["size"].(string); ok {
			orgSize = size
		}
	}

	// Count engineers from teams
	for _, team := range ctx.GetTeams() {
		if headcount, ok := team.Attributes["headcount"].(int); ok {
			teamCount += headcount
		}
	}

	summary := fmt.Sprintf("✓ Context loaded (%s", orgName)
	if orgSize != "" {
		summary += fmt.Sprintf(", %s", orgSize)
	}
	if teamCount > 0 {
		summary += fmt.Sprintf(", %d engineers", teamCount)
	}
	summary += ")"
	fmt.Println(style.Render(summary))
}

// Specialist summoned message with reasons
func PrintSpecialistSummoned(results []types.SummonResult) {
	if len(results) == 0 {
		return
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	fmt.Println()
	for _, r := range results {
		msg := fmt.Sprintf("📎 Including %s — %s", r.Specialist.Name, r.Reason)
		fmt.Println(style.Render(msg))
	}
}

// PrintAdvisorCapWarning warns when too many advisors are selected.
func PrintAdvisorCapWarning(count, max int) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	fmt.Println()
	msg := fmt.Sprintf("⚠️  %d advisors selected (recommended max: %d)", count, max)
	fmt.Println(style.Render(msg))
	fmt.Println(style.Render("   More advisors = longer response time and higher cost"))
}

// PrintSocraticQuestions displays the clarifying questions.
func PrintSocraticQuestions(questions []string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141"))

	fmt.Println()
	fmt.Println(style.Render("Before consulting the advisory board, let me ask a few clarifying questions:"))
	fmt.Println()

	for i, q := range questions {
		fmt.Printf("  %d. %s\n", i+1, q)
	}
	fmt.Println()
	fmt.Println(style.Render("Answer each question (or press Enter to skip):"))
}

// PrintSocraticAnswerPrompt shows which question to answer.
func PrintSocraticAnswerPrompt(questionNum int, question string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141"))

	fmt.Println()
	fmt.Println(style.Render(fmt.Sprintf("Q%d: %s", questionNum, question)))
}

// Error message
func PrintError(msg string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)
	fmt.Println(style.Render("✖ " + msg))
}

// Success message
func PrintSuccess(msg string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46"))
	fmt.Println(style.Render("✓ " + msg))
}

// Info message
func PrintInfo(msg string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))
	fmt.Println(style.Render(msg))
}

// PrintFacilitatorMessage displays a message from the discovery facilitator.
func PrintFacilitatorMessage(msg string) {
	// Simple plain text output
	fmt.Println()
	fmt.Println("Jordan — Discovery Coach")
	fmt.Println(msg)
	fmt.Println()
}

// PrintUserMessage displays a user's message in the discovery conversation.
func PrintUserMessage(msg string) {
	// Simple plain text output
	fmt.Println()
	fmt.Println("You:")
	fmt.Println(msg)
	fmt.Println()
}

// PrintBrief displays a generated brief in a structured format.
func PrintBrief(brief *types.Brief) {
	if brief == nil {
		fmt.Println("No brief generated yet")
		return
	}

	fmt.Println()
	fmt.Println("=== GENERATED BRIEF ===")
	fmt.Println()

	fmt.Println("Problem:")
	fmt.Println(brief.ProblemStatement)
	fmt.Println()

	fmt.Println("Context:")
	fmt.Println(brief.Context)
	fmt.Println()

	if len(brief.Constraints) > 0 {
		fmt.Println("Constraints:")
		for _, c := range brief.Constraints {
			fmt.Println("  - " + c)
		}
		fmt.Println()
	}

	if len(brief.Goals) > 0 {
		fmt.Println("Goals:")
		for _, g := range brief.Goals {
			fmt.Println("  - " + g)
		}
		fmt.Println()
	}

	if len(brief.KeyQuestions) > 0 {
		fmt.Println("Key Questions:")
		for _, q := range brief.KeyQuestions {
			fmt.Println("  - " + q)
		}
		fmt.Println()
	}

	if len(brief.SuggestedAdvisors) > 0 {
		advisorNames := make([]string, len(brief.SuggestedAdvisors))
		for i, id := range brief.SuggestedAdvisors {
			advisorNames[i] = string(id)
		}
		fmt.Println("Suggested Advisors: " + strings.Join(advisorNames, ", "))
	}
	fmt.Println()
}

// PrintDiscoveryWelcome displays the discovery mode welcome banner.
func PrintDiscoveryWelcome() {
	fmt.Println("CTO Advisory Board - Discovery Mode")
	fmt.Println("Let's clarify your challenge before consulting the board.")
	fmt.Println("Type /skip to jump to panel mode, or /help for commands.")
	fmt.Println()
}

// PrintPanelModeActive displays the panel mode indicator.
func PrintPanelModeActive() {
	fmt.Println("Panel Mode - Ask questions directly to the advisory board")
	fmt.Println()
}

// PrintBriefConfirmationPrompt displays options for confirming or editing the brief.
func PrintBriefConfirmationPrompt() {
	fmt.Println("Review the brief above. Options:")
	fmt.Println("  Enter or /confirm  - Accept and proceed to panel")
	fmt.Println("  /edit-brief <field> - Modify a field (problem, context, constraints, goals, questions)")
	fmt.Println("  /regen             - Regenerate brief from conversation")
	fmt.Println("  /discovery         - Return to discovery mode")
	fmt.Println()
}

// PrintFrameworkOptions displays the parsed options for a framework decision.
func PrintFrameworkOptions(options []string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141"))

	fmt.Println()
	fmt.Println(style.Render("Decision Framework Mode"))
	fmt.Printf("Comparing: %s\n", strings.Join(options, " vs "))
}

// PrintFrameworkCriteria displays the evaluation criteria.
func PrintFrameworkCriteria(criteria []types.FrameworkCriterion) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	fmt.Println()
	fmt.Println(style.Render("Evaluation Criteria:"))
	fmt.Println()

	for i, crit := range criteria {
		weightBar := strings.Repeat("●", int(crit.Weight)) + strings.Repeat("○", 5-int(crit.Weight))
		fmt.Printf("  %d. %s [%s]\n", i+1, crit.Name, weightBar)
		fmt.Printf("     %s\n", crit.Description)
	}
}

// PrintFrameworkCriteriaPrompt shows options for confirming criteria.
func PrintFrameworkCriteriaPrompt() {
	fmt.Println()
	fmt.Println("Review criteria above. Options:")
	fmt.Println("  Enter or /confirm  - Accept and proceed to scoring")
	fmt.Println("  /add <name>        - Add a criterion")
	fmt.Println("  /remove <number>   - Remove a criterion")
	fmt.Println("  /weight <n> <1-5>  - Change weight of criterion n")
	fmt.Println("  /cancel            - Cancel framework mode")
	fmt.Println()
}

// PrintFrameworkMatrix displays the scoring matrix.
func PrintFrameworkMatrix(state *types.FrameworkState) {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("46"))

	fmt.Println()
	fmt.Println(titleStyle.Render("=== DECISION MATRIX ==="))
	fmt.Println()

	// Calculate column widths
	critWidth := 20
	optWidth := 12
	for _, crit := range state.Criteria {
		if len(crit.Name) > critWidth {
			critWidth = len(crit.Name)
		}
	}
	for _, opt := range state.Options {
		if len(opt) > optWidth {
			optWidth = len(opt)
		}
	}

	// Header row
	header := fmt.Sprintf("%-*s", critWidth, "Criterion")
	for _, opt := range state.Options {
		header += fmt.Sprintf(" | %-*s", optWidth, opt)
	}
	header += " | Weight"
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	// Score matrix
	matrix := state.GetScoreMatrix()

	// Data rows
	for _, crit := range state.Criteria {
		row := fmt.Sprintf("%-*s", critWidth, crit.Name)
		for _, opt := range state.Options {
			score := matrix[opt][crit.Name]
			scoreStr := fmt.Sprintf("%d/5", score)
			row += fmt.Sprintf(" | %-*s", optWidth, scoreStr)
		}
		row += fmt.Sprintf(" | %.0f", crit.Weight)
		fmt.Println(row)
	}

	fmt.Println(strings.Repeat("-", len(header)))

	// Weighted totals
	totals := state.GetWeightedScores()
	totalRow := fmt.Sprintf("%-*s", critWidth, "WEIGHTED TOTAL")
	for _, opt := range state.Options {
		totalRow += fmt.Sprintf(" | %-*.1f", optWidth, totals[opt])
	}
	fmt.Println(totalRow)
}

// PrintFrameworkRecommendation displays the final recommendation.
func PrintFrameworkRecommendation(state *types.FrameworkState) {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("46"))

	fmt.Println()
	fmt.Println(titleStyle.Render("=== RECOMMENDATION ==="))
	fmt.Println()

	// Find winner based on weighted scores
	totals := state.GetWeightedScores()
	winner := ""
	maxScore := 0.0
	for opt, score := range totals {
		if score > maxScore {
			maxScore = score
			winner = opt
		}
	}

	if state.Recommendation != "" {
		winner = state.Recommendation
	}

	fmt.Printf("Recommended: %s\n", winner)

	if state.Confidence != "" {
		confidenceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
		fmt.Printf("Confidence: %s\n", confidenceStyle.Render(state.Confidence))
	}

	fmt.Println()
}

// PrintStalenessWarning displays a warning when context is outdated.
func PrintStalenessWarning(warning *types.StalenessWarning) {
	if warning == nil {
		return
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	fmt.Println()
	fmt.Println(style.Render(fmt.Sprintf("⚠️  Context outdated: %s.md last updated %d days ago",
		warning.OldestFile, warning.DaysSinceUpdate)))
	fmt.Println(style.Render("   Consider running: cto-advisory context update"))
	fmt.Println()
}

// PrintConflictWarning displays detected context conflicts.
func PrintConflictWarning(conflicts []types.ContextConflict) {
	if len(conflicts) == 0 {
		return
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	fmt.Println()
	fmt.Println(style.Render("⚠️  Context conflicts detected:"))
	for _, c := range conflicts {
		fmt.Printf("   • %s: context says %s, but decision %s mentions %s\n",
			c.Field, c.ContextValue, c.DecisionID, c.DecisionValue)
	}
	fmt.Println()
}

// PrintUpdateSuggestion displays a suggestion to update context.
func PrintUpdateSuggestion(suggestion *types.UpdateSuggestion) {
	if suggestion == nil {
		return
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141"))

	fmt.Println()
	fmt.Println(style.Render(fmt.Sprintf("💡 It sounds like %s", suggestion.Reason)))
	fmt.Printf("   Update entity %s field %s: %s → %s?\n",
		suggestion.EntityID, suggestion.Field, suggestion.OldValue, suggestion.NewValue)
	fmt.Println(style.Render("   Use /update-context to apply"))
	fmt.Println()
}

// PrintRelevantDecisions displays past DRF documents included in context.
func PrintRelevantDecisions(docs []types.DRFDocument) {
	if len(docs) == 0 {
		return
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	fmt.Println()
	fmt.Println(style.Render(fmt.Sprintf("📚 Including %d relevant past decisions in context", len(docs))))
}

// PrintOutcomePrompt shows the outcome recording prompt.
func PrintOutcomePrompt() {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141"))

	fmt.Println()
	fmt.Println(style.Render("📝 Track this decision's outcome?"))
	fmt.Println("   /track       - Record outcome now")
	fmt.Println("   /skip        - Skip (default)")
	fmt.Println("   Enter        - Continue without tracking")
}

// PrintOutcomeRecorded confirms outcome was saved.
func PrintOutcomeRecorded(decisionID string) {
	PrintSuccess(fmt.Sprintf("Outcome recorded for %s", decisionID))
}

// PrintLongQuestionWarning displays a warning about a long question.
func PrintLongQuestionWarning(length int) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	fmt.Println()
	fmt.Println(style.Render(fmt.Sprintf("⚠️  Long question detected (%d characters)", length)))
	fmt.Println(style.Render("   Summarizing to improve advisory board response quality..."))
	fmt.Println()
}

// PrintQuestionSummary displays the summarized question for confirmation.
func PrintQuestionSummary(original, summary string) {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("51"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")).
		Padding(0, 1)

	fmt.Println()
	fmt.Println(titleStyle.Render("Summarized Question:"))
	fmt.Println(boxStyle.Render(summary))
	fmt.Println()
	fmt.Printf("Original: %d chars → Summary: %d chars\n", len(original), len(summary))
}

// PrintSummaryConfirmPrompt displays options for confirming or rejecting the summary.
func PrintSummaryConfirmPrompt() {
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  Enter or y  - Use this summary")
	fmt.Println("  n           - Use original question instead")
	fmt.Println("  /edit       - Edit the summary manually")
	fmt.Println()
}
