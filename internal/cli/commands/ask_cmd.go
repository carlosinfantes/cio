// Package commands implements the CLI commands for the CIO - Chief Intelligence Officer.
package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/carlosinfantes/cio/internal/cli/output"
	"github.com/carlosinfantes/cio/internal/config"
	advisorsPkg "github.com/carlosinfantes/cio/internal/core/advisors"
	ctxLoader "github.com/carlosinfantes/cio/internal/core/context"
	"github.com/carlosinfantes/cio/internal/core/decisions"
	"github.com/carlosinfantes/cio/internal/core/llm"
	"github.com/carlosinfantes/cio/internal/core/modes"
	"github.com/carlosinfantes/cio/internal/types"
)

// NewAskCmd creates the ask command.
func NewAskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ask [question]",
		Short: "Ask the advisory board a question",
		Long: `Ask a question to get perspectives from your AI advisory board.

The advisory board consists of virtual executives (CTO, CISO, VP Engineering,
Staff Architect) who will debate your specific situation.

Examples:
  cio ask "Should we adopt Kubernetes?"
  cio ask --mode socratic "We need to scale"
  cio ask --mode advocate "We've decided to use MongoDB"
  cio ask --advisors cto,architect "Platform architecture review"`,
		Args: cobra.ExactArgs(1),
		RunE: runAskCommand,
	}

	return cmd
}

func runAskCommand(cmd *cobra.Command, args []string) error {
	question := args[0]

	// Load plugins
	if err := loadPlugin(); err != nil {
		output.PrintError(fmt.Sprintf("Loading plugins: %v", err))
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		output.PrintError(fmt.Sprintf("Loading config: %v", err))
		return err
	}

	if cfg.APIKey == "" {
		output.PrintError("No API key configured. Run: cio init")
		return fmt.Errorf("no API key")
	}

	// Load CRF context
	crfCtx, err := ctxLoader.LoadCRFContext()
	if err != nil {
		output.PrintError(fmt.Sprintf("Loading context: %v", err))
	}
	if crfCtx != nil {
		output.PrintContextLoaded(crfCtx)
	}

	// Determine advisors
	var activeAdvisors []types.Persona
	if len(advisors) > 0 {
		// Use specified advisors
		ids := make([]types.AdvisorID, len(advisors))
		for i, a := range advisors {
			ids[i] = types.AdvisorID(a)
		}
		activeAdvisors = advisorsPkg.GetByIDs(ids)
	} else {
		// Use defaults
		ids := cfg.DefaultAdvisors
		activeAdvisors = advisorsPkg.GetByIDs(ids)
	}

	// Auto-summon specialists
	if cfg.AutoSummonSpecialists {
		summonResults := advisorsPkg.SummonSpecialists(question)
		if len(summonResults) > 0 {
			output.PrintSpecialistSummoned(summonResults)
			for _, sr := range summonResults {
				activeAdvisors = append(activeAdvisors, sr.Specialist)
			}
		}
	}

	// Force-include specialists
	if len(include) > 0 {
		ids := make([]types.AdvisorID, len(include))
		for i, a := range include {
			ids[i] = types.AdvisorID(a)
		}
		included := advisorsPkg.GetByIDs(ids)
		activeAdvisors = append(activeAdvisors, included...)
	}

	// Warn if too many advisors
	if len(activeAdvisors) > cfg.MaxAdvisors {
		output.PrintAdvisorCapWarning(len(activeAdvisors), cfg.MaxAdvisors)
	}

	// Create LLM client
	client, err := llm.NewClient(cfg.APIKey, cfg.Model)
	if err != nil {
		output.PrintError(fmt.Sprintf("Creating client: %v", err))
		return err
	}

	// Execute the appropriate mode
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	output.PrintInfo("Consulting the advisory board...")
	fmt.Println()

	modeType := types.Mode(mode)
	var parsed types.ParsedResponse

	switch modeType {
	case types.ModeSocratic:
		parsed, err = modes.Socratic(ctx, client, question, activeAdvisors, crfCtx)
	case types.ModeAdvocate:
		parsed, err = modes.Advocate(ctx, client, question, activeAdvisors, crfCtx)
	case types.ModeFramework:
		parsed, err = modes.Framework(ctx, client, question, activeAdvisors, crfCtx)
	default:
		parsed, err = modes.Panel(ctx, client, question, activeAdvisors, crfCtx)
	}

	if err != nil {
		output.PrintError(fmt.Sprintf("Query failed: %v", err))
		return err
	}

	// Render output
	fmt.Println()
	switch outputFormat {
	case "markdown":
		output.RenderMarkdown(parsed)
	case "json":
		output.RenderJSON(parsed)
	default:
		output.RenderTerminal(parsed)
	}

	// Save decision to history (unless --no-save flag is set)
	if !noSave {
		advisorIDs := make([]types.AdvisorID, len(activeAdvisors))
		for i, a := range activeAdvisors {
			advisorIDs[i] = a.ID
		}

		doc := decisions.CreateDRFDocument(question, modeType, advisorIDs, parsed, crfCtx)
		if err := decisions.SaveDRFDocument(doc); err != nil {
			output.PrintError(fmt.Sprintf("Failed to save decision: %v", err))
		} else {
			fmt.Println()
			output.PrintInfo(fmt.Sprintf("Decision saved: %s", doc.Decision.ID))
			fmt.Println("  Use 'cio history show " + doc.Decision.ID + "' to view details")
			fmt.Println("  Use 'cio history status " + doc.Decision.ID + " approved' to update status")
		}
	}

	return nil
}
