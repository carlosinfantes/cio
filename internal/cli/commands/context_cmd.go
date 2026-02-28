// Package commands implements the context command for the CIO - Chief Intelligence Officer.
package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/carlosinfantes/cio/internal/cli/output"
	"github.com/carlosinfantes/cio/internal/config"
	ctxLoader "github.com/carlosinfantes/cio/internal/core/context"
	"github.com/carlosinfantes/cio/internal/types"
)

func init() {
	rootCmd.AddCommand(newContextCmd())
}

func newContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context [action] [type]",
		Short: "Manage project context",
		Long: `Manage project context (CRF entities).

Actions:
  show [type]   Show context summary or specific entity type (organization, team, system, capability, fact, policy)
  edit <type>   Open context file in $EDITOR
  check         Check for staleness and conflicts

Examples:
  cio context show                 Show all context entities
  cio context show organization    Show organization entity only
  cio context show team            Show team entities
  cio context edit organization    Edit organization.yaml in your editor
  cio context check                Check for outdated or conflicting context`,
		Args: cobra.MaximumNArgs(2),
		RunE: runContext,
	}

	return cmd
}

func runContext(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return handleContextShow("")
	}

	action := args[0]
	entityType := ""
	if len(args) > 1 {
		entityType = args[1]
	}

	switch action {
	case "show":
		return handleContextShow(entityType)
	case "edit":
		return handleContextEdit(entityType)
	case "check":
		return handleContextCheck()
	default:
		output.PrintError(fmt.Sprintf("Unknown action: %s", action))
		fmt.Println("Usage: cio context <show|edit|check> [type]")
		return nil
	}
}

func handleContextShow(entityType string) error {
	if !config.IsInitialized() {
		output.PrintError("Project not initialized. Run: cio init")
		return nil
	}

	// Load CRF context
	ctx, err := ctxLoader.LoadCRFContext()
	if err != nil {
		return err
	}
	if ctx == nil {
		output.PrintInfo("No context found. Run 'cio init' to create context.")
		return nil
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	fmt.Println()

	switch entityType {
	case "organization":
		return showOrganization(ctx, titleStyle, labelStyle, valueStyle)
	case "team":
		return showTeams(ctx, titleStyle, labelStyle, valueStyle)
	case "system":
		return showSystems(ctx, titleStyle, labelStyle, valueStyle)
	case "capability":
		return showCapabilities(ctx, titleStyle, labelStyle, valueStyle)
	case "fact":
		return showFacts(ctx, titleStyle, labelStyle, valueStyle)
	case "policy":
		return showPolicies(ctx, titleStyle, labelStyle, valueStyle)
	case "":
		// Show all
		if err := showOrganization(ctx, titleStyle, labelStyle, valueStyle); err != nil {
			return err
		}
		if err := showTeams(ctx, titleStyle, labelStyle, valueStyle); err != nil {
			return err
		}
		if err := showSystems(ctx, titleStyle, labelStyle, valueStyle); err != nil {
			return err
		}
		if err := showCapabilities(ctx, titleStyle, labelStyle, valueStyle); err != nil {
			return err
		}
		if err := showFacts(ctx, titleStyle, labelStyle, valueStyle); err != nil {
			return err
		}
		return showPolicies(ctx, titleStyle, labelStyle, valueStyle)
	default:
		output.PrintError(fmt.Sprintf("Unknown entity type: %s", entityType))
		fmt.Println("Available: organization, team, system, capability, fact, policy")
		return nil
	}
}

func showOrganization(ctx *types.CRFContext, titleStyle, labelStyle, valueStyle lipgloss.Style) error {
	org := ctx.GetOrganization()
	if org == nil {
		output.PrintInfo("No organization context found")
		return nil
	}

	fmt.Println(titleStyle.Render("Organization"))
	fmt.Println()
	fmt.Printf("  %s %s\n", labelStyle.Render("Name:"), valueStyle.Render(org.Name))

	if industry, ok := org.Attributes["industry"].(string); ok && industry != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Industry:"), valueStyle.Render(industry))
	}
	if stage, ok := org.Attributes["stage"].(string); ok && stage != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Stage:"), valueStyle.Render(stage))
	}
	if founded, ok := org.Attributes["founded"].(int); ok && founded > 0 {
		fmt.Printf("  %s %s\n", labelStyle.Render("Founded:"), valueStyle.Render(fmt.Sprintf("%d", founded)))
	}
	if businessModel, ok := org.Attributes["business_model"].(string); ok && businessModel != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Business Model:"), valueStyle.Render(businessModel))
	}
	if compliance, ok := org.Attributes["compliance"].([]interface{}); ok && len(compliance) > 0 {
		strs := make([]string, len(compliance))
		for i, c := range compliance {
			strs[i] = fmt.Sprintf("%v", c)
		}
		fmt.Printf("  %s %s\n", labelStyle.Render("Compliance:"), valueStyle.Render(strings.Join(strs, ", ")))
	}
	fmt.Println()

	return nil
}

func showTeams(ctx *types.CRFContext, titleStyle, labelStyle, valueStyle lipgloss.Style) error {
	teams := ctx.GetTeams()
	if len(teams) == 0 {
		output.PrintInfo("No team context found")
		return nil
	}

	fmt.Println(titleStyle.Render("Teams"))
	fmt.Println()

	for _, team := range teams {
		fmt.Printf("  %s %s\n", labelStyle.Render("Name:"), valueStyle.Render(team.Name))
		if headcount, ok := team.Attributes["headcount"].(int); ok && headcount > 0 {
			fmt.Printf("    %s %s\n", labelStyle.Render("Headcount:"), valueStyle.Render(fmt.Sprintf("%d", headcount)))
		}
		if skills, ok := team.Attributes["skills"].([]interface{}); ok && len(skills) > 0 {
			strs := make([]string, len(skills))
			for i, s := range skills {
				strs[i] = fmt.Sprintf("%v", s)
			}
			fmt.Printf("    %s %s\n", labelStyle.Render("Skills:"), valueStyle.Render(strings.Join(strs, ", ")))
		}
		fmt.Println()
	}

	return nil
}

func showSystems(ctx *types.CRFContext, titleStyle, labelStyle, valueStyle lipgloss.Style) error {
	if len(ctx.Systems) == 0 {
		output.PrintInfo("No system context found")
		return nil
	}

	fmt.Println(titleStyle.Render("Systems"))
	fmt.Println()

	for _, doc := range ctx.Systems {
		fmt.Printf("  %s %s\n", labelStyle.Render("Name:"), valueStyle.Render(doc.Entity.Name))
		if hosting, ok := doc.Entity.Attributes["hosting"].(string); ok && hosting != "" {
			fmt.Printf("    %s %s\n", labelStyle.Render("Hosting:"), valueStyle.Render(hosting))
		}
		if primaryLanguage, ok := doc.Entity.Attributes["primary_language"].(string); ok && primaryLanguage != "" {
			fmt.Printf("    %s %s\n", labelStyle.Render("Primary Language:"), valueStyle.Render(primaryLanguage))
		}
		if languages, ok := doc.Entity.Attributes["languages"].([]interface{}); ok && len(languages) > 0 {
			strs := make([]string, len(languages))
			for i, l := range languages {
				strs[i] = fmt.Sprintf("%v", l)
			}
			fmt.Printf("    %s %s\n", labelStyle.Render("Languages:"), valueStyle.Render(strings.Join(strs, ", ")))
		}
		if deployment, ok := doc.Entity.Attributes["deployment"].(string); ok && deployment != "" {
			fmt.Printf("    %s %s\n", labelStyle.Render("Deployment:"), valueStyle.Render(deployment))
		}
		fmt.Println()
	}

	return nil
}

func showCapabilities(ctx *types.CRFContext, titleStyle, labelStyle, valueStyle lipgloss.Style) error {
	if len(ctx.Capabilities) == 0 {
		output.PrintInfo("No capability context found")
		return nil
	}

	fmt.Println(titleStyle.Render("Capabilities"))
	fmt.Println()

	for _, doc := range ctx.Capabilities {
		fmt.Printf("  %s %s\n", labelStyle.Render("Name:"), valueStyle.Render(doc.Entity.Name))
		if maturity, ok := doc.Entity.Attributes["maturity"].(string); ok && maturity != "" {
			fmt.Printf("    %s %s\n", labelStyle.Render("Maturity:"), valueStyle.Render(maturity))
		}
		fmt.Println()
	}

	return nil
}

func showFacts(ctx *types.CRFContext, titleStyle, labelStyle, valueStyle lipgloss.Style) error {
	if len(ctx.Facts) == 0 {
		output.PrintInfo("No fact context found")
		return nil
	}

	fmt.Println(titleStyle.Render("Facts"))
	fmt.Println()

	for _, doc := range ctx.Facts {
		fmt.Printf("  %s %s\n", labelStyle.Render("Name:"), valueStyle.Render(doc.Entity.Name))
		if factType, ok := doc.Entity.Attributes["fact_type"].(string); ok && factType != "" {
			fmt.Printf("    %s %s\n", labelStyle.Render("Type:"), valueStyle.Render(factType))
		}
		if value, ok := doc.Entity.Attributes["value"]; ok {
			fmt.Printf("    %s %v\n", labelStyle.Render("Value:"), valueStyle.Render(fmt.Sprintf("%v", value)))
		}
		fmt.Println()
	}

	return nil
}

func showPolicies(ctx *types.CRFContext, titleStyle, labelStyle, valueStyle lipgloss.Style) error {
	if len(ctx.Policies) == 0 {
		output.PrintInfo("No policy context found")
		return nil
	}

	fmt.Println(titleStyle.Render("Policies"))
	fmt.Println()

	for _, doc := range ctx.Policies {
		fmt.Printf("  %s %s\n", labelStyle.Render("Name:"), valueStyle.Render(doc.Entity.Name))
		if scope, ok := doc.Entity.Attributes["scope"].(string); ok && scope != "" {
			fmt.Printf("    %s %s\n", labelStyle.Render("Scope:"), valueStyle.Render(scope))
		}
		fmt.Println()
	}

	return nil
}

func handleContextEdit(entityType string) error {
	if entityType == "" {
		output.PrintError("Please specify an entity type to edit: organization, team, system, capability, fact, policy")
		return nil
	}

	// Map entity type to filename
	validFiles := map[string]string{
		"organization": "organization.yaml",
		"team":         "teams.yaml",
		"system":       "systems.yaml",
		"capability":   "capabilities.yaml",
		"fact":         "facts.yaml",
		"policy":       "policies.yaml",
	}

	filename, ok := validFiles[entityType]
	if !ok {
		output.PrintError(fmt.Sprintf("Unknown entity type: %s", entityType))
		fmt.Println("Available: organization, team, system, capability, fact, policy")
		return nil
	}

	// Get the context directory
	contextDir, err := config.GetContextDir()
	if err != nil {
		return err
	}

	filePath := contextDir + "/" + filename

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		output.PrintError(fmt.Sprintf("Context file not found: %s", filePath))
		fmt.Println("Run 'cio init' to create context files")
		return nil
	}

	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi" // Default fallback
	}

	// Open in editor
	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	output.PrintInfo(fmt.Sprintf("Opening %s in %s...", filename, editor))
	return cmd.Run()
}

func handleContextCheck() error {
	if !config.IsInitialized() {
		output.PrintError("Project not initialized. Run: cio init")
		return nil
	}

	// Load CRF context
	ctx, err := ctxLoader.LoadCRFContext()
	if err != nil {
		return err
	}

	fmt.Println()
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
	fmt.Println(titleStyle.Render("Context Health Check"))
	fmt.Println()

	hasIssues := false

	// Check staleness
	if warning := ctxLoader.CheckStaleness(ctx, 30); warning != nil {
		output.PrintStalenessWarning(warning)
		hasIssues = true
	}

	// Check for conflicts with recent decisions
	if conflicts := ctxLoader.DetectConflicts(ctx); len(conflicts) > 0 {
		output.PrintConflictWarning(conflicts)
		hasIssues = true
	}

	if !hasIssues {
		output.PrintSuccess("All context files are up to date with no conflicts detected")
	}

	fmt.Println()
	return nil
}
