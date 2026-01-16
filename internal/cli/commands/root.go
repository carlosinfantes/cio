// Package commands implements the CLI commands for the CTO Advisory Board.
package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/carlosinfantes/cto-advisory-board/internal/cli/output"
	"github.com/carlosinfantes/cto-advisory-board/internal/config"
	advisorsPkg "github.com/carlosinfantes/cto-advisory-board/internal/core/advisors"
	ctxLoader "github.com/carlosinfantes/cto-advisory-board/internal/core/context"
	"github.com/carlosinfantes/cto-advisory-board/internal/core/llm"
	"github.com/carlosinfantes/cto-advisory-board/internal/core/modes"
	"github.com/carlosinfantes/cto-advisory-board/internal/plugins"
	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

var (
	// Global flags
	outputFormat string
	mode         string
	advisors     []string
	include      []string
	verbose      bool
	noSave       bool
	pluginName   string

	rootCmd = &cobra.Command{
		Use:   "cto-advisory [question]",
		Short: "AI-powered executive committee for CTOs",
		Long: `CTO Advisory Board - Your AI-powered executive committee for technical decisions.

Get perspectives from a virtual CTO, CISO, VP Engineering, and Staff Architect
— all debating your specific situation.

Examples:
  cto-advisory "Should we adopt Kubernetes?"
  cto-advisory --mode socratic "We need to scale"
  cto-advisory --mode advocate "We've decided to use MongoDB"
  cto-advisory --advisors cto,architect "Platform architecture review"`,
		Args: cobra.MaximumNArgs(1),
		RunE: runAsk,
	}
)

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "terminal",
		"Output format: terminal, markdown, json")
	rootCmd.PersistentFlags().StringVarP(&mode, "mode", "m", string(types.ModePanel),
		"Interaction mode: panel, socratic, advocate, framework")
	rootCmd.PersistentFlags().StringSliceVarP(&advisors, "advisors", "a", nil,
		"Comma-separated advisor IDs (cto,ciso,vp-eng,architect)")
	rootCmd.PersistentFlags().StringSliceVar(&include, "include", nil,
		"Force-include specialist advisors (cfo,product,devops)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&noSave, "no-save", false,
		"Skip saving this decision to history")
	rootCmd.PersistentFlags().StringVar(&pluginName, "plugin", "",
		"Use a specific domain plugin (e.g., peluqueria, legal-advisory)")

	// Add subcommands
	// Note: initCmd, configCmd, contextCmd, historyCmd are added by their respective *_cmd.go files
	// via init() functions that replace placeholders with full implementations
	rootCmd.AddCommand(NewAskCmd())
	rootCmd.AddCommand(versionCmd)
}

func runAsk(cmd *cobra.Command, args []string) error {
	// If no args, launch interactive mode
	if len(args) == 0 {
		return RunInteractive()
	}

	question := args[0]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		output.PrintError(fmt.Sprintf("Loading config: %v", err))
		return err
	}

	// Load and activate plugin if specified
	if err := loadPlugin(); err != nil {
		output.PrintError(fmt.Sprintf("Loading plugin: %v", err))
		return err
	}

	if cfg.APIKey == "" {
		output.PrintError("No API key configured. Run: cto-advisory init")
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
		// Use core advisors from active plugin (or defaults)
		activeAdvisors = advisorsPkg.CoreAdvisors()
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

	return nil
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cto-advisory version 1.0.0")
		fmt.Println("https://github.com/carlosinfantes/cto-advisory-board")
	},
}

// loadPlugin loads and activates the specified plugin.
func loadPlugin() error {
	registry := plugins.GetRegistry()

	// Always load the default CTO plugin first
	defaultPlugin := plugins.DefaultCTOPlugin()
	defaultPluginWrapper := &plugins.Plugin{
		Manifest: defaultPlugin,
		Personas: defaultPlugin.ToPersonas(),
	}
	registry.RegisterPlugin("cto-advisory", defaultPluginWrapper)

	// Load plugins from the plugins directory
	if err := registry.LoadPluginsFromDir("plugins"); err != nil {
		return fmt.Errorf("loading plugins: %w", err)
	}

	// Also try loading from absolute path if we're in a different directory
	if cwd, err := filepath.Abs("."); err == nil {
		pluginsDir := filepath.Join(cwd, "plugins")
		_ = registry.LoadPluginsFromDir(pluginsDir)
	}

	// If a plugin is specified, activate it
	if pluginName != "" {
		if err := registry.SetActive(pluginName); err != nil {
			// List available plugins
			domains := registry.ListDomains()
			return fmt.Errorf("plugin '%s' not found. Available: %v", pluginName, domains)
		}
		output.PrintInfo(fmt.Sprintf("Plugin activo: %s", pluginName))
	}

	return nil
}
