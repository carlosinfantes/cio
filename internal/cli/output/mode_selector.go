// Package output implements the mode selector UI for unified entry.
package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/carlosinfantes/cio/internal/types"
)

// ModeOption represents a selectable mode in the unified entry.
type ModeOption struct {
	Key         string
	Emoji       string
	Name        string
	Description string
	Mode        types.Mode
}

// DefaultModeOptions returns the standard mode options for unified entry.
func DefaultModeOptions() []ModeOption {
	return []ModeOption{
		{
			Key:         "1",
			Emoji:       "🔍",
			Name:        "Discover",
			Description: "Explore a new question with Jordan",
			Mode:        "",
		},
		{
			Key:         "2",
			Emoji:       "🎯",
			Name:        "Decide",
			Description: "Get panel advice on a decision",
			Mode:        types.ModePanel,
		},
		{
			Key:         "3",
			Emoji:       "😈",
			Name:        "Challenge",
			Description: "Devil's advocate on your plan",
			Mode:        types.ModeAdvocate,
		},
		{
			Key:         "4",
			Emoji:       "⚖️",
			Name:        "Framework",
			Description: "Structured option evaluation",
			Mode:        types.ModeFramework,
		},
		{
			Key:         "5",
			Emoji:       "📋",
			Name:        "Context",
			Description: "Review/update your context",
			Mode:        "",
		},
		{
			Key:         "6",
			Emoji:       "📜",
			Name:        "History",
			Description: "Browse past decisions",
			Mode:        "",
		},
	}
}

// PrintModeSelector displays the unified mode selector interface.
func PrintModeSelector(ctx *types.CRFContext, lastDecision *types.DRFDocument) {
	width := getTerminalWidth()

	// Header style
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141")).
		MarginBottom(1)

	// Box style
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("141")).
		Padding(1, 2).
		Width(width - 4)

	// Info line style
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	// Build content
	var content strings.Builder
	content.WriteString(headerStyle.Render("💭 CIO - Chief Intelligence Officer"))
	content.WriteString("\n\n")

	// Context status
	if ctx != nil {
		org := ctx.GetOrganization()
		if org != nil {
			content.WriteString(infoStyle.Render(fmt.Sprintf("Current context: %s (loaded)", org.Name)))
			content.WriteString("\n")
		}
	} else {
		content.WriteString(infoStyle.Render("Current context: Not initialized"))
		content.WriteString("\n")
	}

	// Last decision
	if lastDecision != nil {
		days := daysSince(lastDecision.Meta.CreatedAt)
		content.WriteString(infoStyle.Render(fmt.Sprintf("Last decision: %d days ago (%s)", days, truncateText(lastDecision.Decision.Title, 40))))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString("Select mode:\n")

	// Mode options
	optionStyle := lipgloss.NewStyle().
		MarginLeft(2)

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("51"))

	options := DefaultModeOptions()
	for _, opt := range options {
		line := fmt.Sprintf("[%s] %s %s - %s",
			keyStyle.Render(opt.Key),
			opt.Emoji,
			opt.Name,
			opt.Description)
		content.WriteString(optionStyle.Render(line))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(infoStyle.Render("Or just type your question..."))

	fmt.Println(boxStyle.Render(content.String()))
	fmt.Println()
}

// PrintModeSwitchHint displays the mode switch hint.
func PrintModeSwitchHint() {
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Italic(true)

	fmt.Println(hintStyle.Render("Tip: Press Ctrl+M to switch modes, or type /help for commands"))
}

// PrintJordanSuggestion displays a mode suggestion from Jordan.
func PrintJordanSuggestion(suggestion string, options []string) {
	fmt.Println()

	// Facilitator style
	facilitatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141"))

	fmt.Println(facilitatorStyle.Render("💭 Jordan: " + suggestion))
	fmt.Println()

	if len(options) > 0 {
		optionStyle := lipgloss.NewStyle().
			MarginLeft(4).
			Foreground(lipgloss.Color("245"))

		for i, opt := range options {
			fmt.Println(optionStyle.Render(fmt.Sprintf("[%d] %s", i+1, opt)))
		}
		fmt.Println()
	}
}

// PrintEscalationNotice displays the auto-escalation message.
func PrintEscalationNotice(reason string) {
	fmt.Println()

	noticeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("46"))

	fmt.Println(noticeStyle.Render("💭 Jordan: " + reason))
	fmt.Println()
	fmt.Println("Bringing in the advisory panel now...")
	fmt.Println()
}

// PrintPhaseTransition displays a phase transition message.
func PrintPhaseTransition(fromPhase, toPhase, reason string) {
	transitionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Italic(true)

	fmt.Println(transitionStyle.Render(fmt.Sprintf("→ %s", reason)))
}

// PrintContextValidation displays context validation results.
func PrintContextValidation(stale bool, staleDays int, missing []string) {
	if !stale && len(missing) == 0 {
		return
	}

	fmt.Println()

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	if stale {
		fmt.Println(warningStyle.Render(fmt.Sprintf("⚠️  Context was last updated %d days ago", staleDays)))
	}

	if len(missing) > 0 {
		fmt.Println(warningStyle.Render(fmt.Sprintf("⚠️  Missing context: %s", strings.Join(missing, ", "))))
	}
}

// PrintQuickActions displays quick action shortcuts.
func PrintQuickActions() {
	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	fmt.Println(actionStyle.Render("Quick actions: /discover, /panel, /mode <name>, /context, /history, /help"))
}

// truncateText truncates text to maxLen and adds ellipsis.
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}

// daysSince calculates days since a given time.
func daysSince(t interface{}) int {
	// Type assertion for time.Time
	switch v := t.(type) {
	case interface{ Unix() int64 }:
		return int((currentUnix() - v.Unix()) / 86400)
	default:
		return 0
	}
}

func currentUnix() int64 {
	return time.Now().Unix()
}
