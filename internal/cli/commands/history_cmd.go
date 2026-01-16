// Package commands implements the history command for the CTO Advisory Board.
package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/carlosinfantes/cto-advisory-board/internal/cli/output"
	"github.com/carlosinfantes/cto-advisory-board/internal/core/decisions"
	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

var (
	historyStatusFilter string
	historyTagFilter    string
)

func init() {
	rootCmd.AddCommand(newHistoryCmd())
}

func newHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history <action> [args...]",
		Short: "Manage decision history",
		Long: `Manage decision history.

Actions:
  list                     List all decisions (supports --status and --tag filters)
  show <id>                Show decision details
  search <query>           Search decisions by keyword
  status <id> <status>     Update decision status (draft, review, approved, rejected, superseded, archived)
  tag <id> <tag>           Add a tag to a decision

Examples:
  cto-advisory history list
  cto-advisory history list --status approved
  cto-advisory history list --tag infrastructure
  cto-advisory history search kubernetes
  cto-advisory history show dec-2024-01-13-kubernetes
  cto-advisory history status dec-2024-01-13-kubernetes approved
  cto-advisory history tag dec-2024-01-13-kubernetes infrastructure`,
		Args: cobra.MinimumNArgs(1),
		RunE: runHistory,
	}

	cmd.Flags().StringVar(&historyStatusFilter, "status", "", "Filter by status (draft, review, approved, rejected, superseded, archived)")
	cmd.Flags().StringVar(&historyTagFilter, "tag", "", "Filter by tag")

	return cmd
}

func runHistory(cmd *cobra.Command, args []string) error {
	action := args[0]

	switch action {
	case "list":
		return handleHistoryList()
	case "show":
		if len(args) < 2 {
			output.PrintError("Usage: cto-advisory history show <id>")
			return nil
		}
		return handleHistoryShow(args[1])
	case "search":
		if len(args) < 2 {
			output.PrintError("Usage: cto-advisory history search <query>")
			return nil
		}
		return handleHistorySearch(strings.Join(args[1:], " "))
	case "status":
		if len(args) < 3 {
			output.PrintError("Usage: cto-advisory history status <id> <status>")
			fmt.Println("Status values: draft, review, approved, rejected, superseded, archived")
			return nil
		}
		return handleHistoryStatus(args[1], args[2])
	case "tag":
		if len(args) < 3 {
			output.PrintError("Usage: cto-advisory history tag <id> <tag>")
			return nil
		}
		return handleHistoryTag(args[1], args[2])
	default:
		output.PrintError(fmt.Sprintf("Unknown action: %s", action))
		fmt.Println("Usage: cto-advisory history <list|show|search|status|tag>")
		return nil
	}
}

func handleHistoryList() error {
	// Build filters
	var filters *decisions.ListFilters
	if historyStatusFilter != "" || historyTagFilter != "" {
		filters = &decisions.ListFilters{
			Status: types.DRFStatus(historyStatusFilter),
			Tag:    historyTagFilter,
		}
	}

	docs, err := decisions.ListDRFDocuments(filters)
	if err != nil {
		return err
	}

	if len(docs) == 0 {
		if filters != nil {
			output.PrintInfo("No decisions match the filters")
		} else {
			output.PrintInfo("No decisions recorded yet")
			fmt.Println("\nDecisions are automatically saved when you ask questions.")
			fmt.Println("Use --no-save to skip saving a decision.")
		}
		return nil
	}

	titleStyle := lipgloss.NewStyle().Bold(true)
	idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	tagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	fmt.Println()
	fmt.Println(titleStyle.Render("Decision History"))
	if filters != nil {
		filterInfo := ""
		if historyStatusFilter != "" {
			filterInfo += fmt.Sprintf(" status=%s", historyStatusFilter)
		}
		if historyTagFilter != "" {
			filterInfo += fmt.Sprintf(" tag=%s", historyTagFilter)
		}
		fmt.Printf("  Filters:%s\n", filterInfo)
	}
	fmt.Println()

	for _, doc := range docs {
		statusColor := getStatusColor(doc.Meta.Status)
		styledStatus := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(string(doc.Meta.Status))

		// Use title or intent as the display text
		displayText := doc.Decision.Title
		if displayText == "" {
			displayText = doc.Decision.Intent
		}
		if len(displayText) > 60 {
			displayText = displayText[:57] + "..."
		}

		fmt.Printf("  %s %s\n", idStyle.Render(doc.Decision.ID), styledStatus)
		fmt.Printf("    %s\n", displayText)

		// Show tags if present
		if len(doc.Meta.Tags) > 0 {
			fmt.Printf("    %s\n", tagStyle.Render("["+strings.Join(doc.Meta.Tags, ", ")+"]"))
		}

		// Get mode from cognitive state or reasoning patterns
		mode := "panel"
		if doc.Reasoning != nil && len(doc.Reasoning.PatternsApplied) > 0 {
			mode = doc.Reasoning.PatternsApplied[0]
		}
		fmt.Printf("    %s %s\n", dateStyle.Render(doc.Meta.CreatedAt.Format("2006-01-02")), statusStyle.Render(fmt.Sprintf("(%s mode)", mode)))
		fmt.Println()
	}

	return nil
}

func handleHistoryShow(id string) error {
	doc, err := decisions.GetDRFDocument(id)
	if err != nil {
		return err
	}
	if doc == nil {
		output.PrintError(fmt.Sprintf("Decision not found: %s", id))
		return nil
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	fmt.Println()
	fmt.Println(titleStyle.Render("Decision Details"))
	fmt.Println()
	fmt.Printf("  %s %s\n", labelStyle.Render("ID:"), valueStyle.Render(doc.Decision.ID))
	fmt.Printf("  %s %s\n", labelStyle.Render("Title:"), valueStyle.Render(doc.Decision.Title))
	fmt.Printf("  %s %s\n", labelStyle.Render("Intent:"), valueStyle.Render(doc.Decision.Intent))

	statusColor := getStatusColor(doc.Meta.Status)
	styledStatus := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(string(doc.Meta.Status))
	fmt.Printf("  %s %s\n", labelStyle.Render("Status:"), styledStatus)

	// Get mode from reasoning patterns
	mode := "panel"
	if doc.Reasoning != nil && len(doc.Reasoning.PatternsApplied) > 0 {
		mode = doc.Reasoning.PatternsApplied[0]
	}
	fmt.Printf("  %s %s\n", labelStyle.Render("Mode:"), valueStyle.Render(mode))

	if len(doc.Meta.Tags) > 0 {
		fmt.Printf("  %s %s\n", labelStyle.Render("Tags:"), valueStyle.Render(strings.Join(doc.Meta.Tags, ", ")))
	}

	// Show advisors (from interventions)
	if len(doc.Interventions) > 0 {
		advisors := make([]string, 0)
		seen := make(map[string]bool)
		for _, intervention := range doc.Interventions {
			if !seen[intervention.Source] {
				advisors = append(advisors, intervention.Source)
				seen[intervention.Source] = true
			}
		}
		fmt.Printf("  %s %s\n", labelStyle.Render("Advisors:"), valueStyle.Render(strings.Join(advisors, ", ")))
	}

	fmt.Printf("  %s %s\n", labelStyle.Render("Created:"), valueStyle.Render(doc.Meta.CreatedAt.Format("2006-01-02 15:04")))
	fmt.Printf("  %s %s\n", labelStyle.Render("Updated:"), valueStyle.Render(doc.Meta.UpdatedAt.Format("2006-01-02 15:04")))

	// Show synthesis if available
	if doc.Synthesis.Decision != "" {
		fmt.Println()
		fmt.Println(titleStyle.Render("Synthesis"))
		synthesis := doc.Synthesis.Decision
		if len(synthesis) > 500 {
			synthesis = synthesis[:497] + "..."
		}
		fmt.Printf("  %s\n", valueStyle.Render(synthesis))

		if doc.CognitiveState.Confidence > 0 {
			fmt.Printf("  %s %d%%\n", labelStyle.Render("Confidence:"), doc.CognitiveState.Confidence)
		}
	}

	fmt.Println()
	return nil
}

func handleHistorySearch(query string) error {
	results, err := decisions.Search(query)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		output.PrintInfo(fmt.Sprintf("No decisions found matching: %s", query))
		return nil
	}

	titleStyle := lipgloss.NewStyle().Bold(true)
	idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	fmt.Println()
	fmt.Printf("%s \"%s\"\n", titleStyle.Render("Search Results:"), query)
	fmt.Println()

	for _, doc := range results {
		statusColor := getStatusColor(doc.Meta.Status)
		styledStatus := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(string(doc.Meta.Status))

		displayText := doc.Decision.Title
		if displayText == "" {
			displayText = doc.Decision.Intent
		}
		if len(displayText) > 60 {
			displayText = displayText[:57] + "..."
		}

		fmt.Printf("  %s %s\n", idStyle.Render(doc.Decision.ID), styledStatus)
		fmt.Printf("    %s\n", displayText)
		fmt.Printf("    %s\n", dateStyle.Render(doc.Meta.CreatedAt.Format("2006-01-02")))
		fmt.Println()
	}

	return nil
}

func handleHistoryStatus(id, status string) error {
	// Validate status
	validStatuses := map[string]types.DRFStatus{
		"draft":      types.DRFStatusDraft,
		"review":     types.DRFStatusReview,
		"approved":   types.DRFStatusApproved,
		"rejected":   types.DRFStatusRejected,
		"superseded": types.DRFStatusSuperseded,
		"archived":   types.DRFStatusArchived,
	}

	newStatus, ok := validStatuses[status]
	if !ok {
		output.PrintError(fmt.Sprintf("Invalid status: %s", status))
		fmt.Println("Valid values: draft, review, approved, rejected, superseded, archived")
		return nil
	}

	if err := decisions.UpdateStatus(id, newStatus); err != nil {
		if strings.Contains(err.Error(), "not found") {
			output.PrintError(fmt.Sprintf("Decision not found: %s", id))
			return nil
		}
		return err
	}

	output.PrintSuccess(fmt.Sprintf("Updated %s status to: %s", id, status))
	return nil
}

func handleHistoryTag(id, tag string) error {
	if err := decisions.AddTag(id, tag); err != nil {
		if strings.Contains(err.Error(), "not found") {
			output.PrintError(fmt.Sprintf("Decision not found: %s", id))
			return nil
		}
		return err
	}

	output.PrintSuccess(fmt.Sprintf("Added tag '%s' to %s", tag, id))
	return nil
}

func getStatusColor(status types.DRFStatus) string {
	switch status {
	case types.DRFStatusDraft:
		return "226" // Yellow
	case types.DRFStatusReview:
		return "51" // Cyan
	case types.DRFStatusApproved:
		return "82" // Green
	case types.DRFStatusRejected:
		return "196" // Red
	case types.DRFStatusSuperseded:
		return "213" // Pink
	case types.DRFStatusArchived:
		return "240" // Gray
	default:
		return "255" // White
	}
}
