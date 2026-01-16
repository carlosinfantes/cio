// Package commands implements the config command for the CTO Advisory Board.
package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/carlosinfantes/cto-advisory-board/internal/cli/output"
	"github.com/carlosinfantes/cto-advisory-board/internal/config"
	"github.com/carlosinfantes/cto-advisory-board/internal/core/llm"
	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

func init() {
	rootCmd.AddCommand(newConfigCmd())
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <action> [key] [value]",
		Short: "Manage configuration",
		Long: `Manage CTO Advisory Board configuration.

Actions:
  set <key> <value>  Set a configuration value
  get <key>          Get a configuration value
  list               List all configuration values

Keys:
  api-key            Your OpenRouter API key
  model              Model to use (claude-sonnet-4-20250514, claude-opus-4-20250514)
  default-mode       Default interaction mode (panel, socratic, advocate, framework)
  auto-summon        Auto-summon specialists (true/false)
  max-advisors       Maximum advisors per session (1-7)`,
		Args: cobra.MinimumNArgs(1),
		RunE: runConfig,
	}

	return cmd
}

func runConfig(cmd *cobra.Command, args []string) error {
	action := args[0]

	switch action {
	case "set":
		return handleConfigSet(args[1:])
	case "get":
		return handleConfigGet(args[1:])
	case "list":
		return handleConfigList()
	default:
		output.PrintError(fmt.Sprintf("Unknown action: %s", action))
		fmt.Println("Usage: cto-advisory config <set|get|list> [key] [value]")
		return nil
	}
}

func handleConfigSet(args []string) error {
	if len(args) < 2 {
		output.PrintError("Usage: cto-advisory config set <key> <value>")
		fmt.Println("\nAvailable keys:")
		fmt.Println("  api-key          Your OpenRouter API key")
		fmt.Println("  model            Model to use")
		fmt.Println("  default-mode     Default interaction mode")
		fmt.Println("  auto-summon      Auto-summon specialists (true/false)")
		fmt.Println("  max-advisors     Maximum advisors (1-7)")
		return nil
	}

	key := args[0]
	value := args[1]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch key {
	case "api-key":
		// Validate API key
		output.PrintInfo("Validating API key...")

		client, err := llm.NewClient(value, cfg.Model)
		if err != nil {
			output.PrintError("Invalid API key format")
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := client.ValidateAPIKey(ctx); err != nil {
			if err == llm.ErrInvalidKey {
				output.PrintError("Invalid API key. Please check and try again.")
				return nil
			}
			output.PrintInfo(fmt.Sprintf("Could not verify key (%v). Saving anyway.", err))
		}

		cfg.APIKey = value
		if err := config.Save(cfg); err != nil {
			return err
		}
		output.PrintSuccess("API key saved and validated")

	case "model":
		// OpenRouter supports many models - just save whatever they provide
		// Common models: anthropic/claude-sonnet-4-20250514, openai/gpt-4o, meta-llama/llama-3-70b-instruct
		if value == "" {
			output.PrintError("Model name cannot be empty")
			return nil
		}
		cfg.Model = value
		if err := config.Save(cfg); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Model set to: %s", value))
		output.PrintInfo("Popular models: anthropic/claude-3.5-sonnet, openai/gpt-4o, google/gemini-pro")

	case "default-mode":
		validModes := []string{"panel", "socratic", "advocate", "framework"}
		valid := false
		for _, m := range validModes {
			if value == m {
				valid = true
				break
			}
		}
		if !valid {
			output.PrintError(fmt.Sprintf("Invalid mode. Use: %s", strings.Join(validModes, ", ")))
			return nil
		}
		cfg.DefaultMode = types.Mode(value)
		if err := config.Save(cfg); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Default mode set to: %s", value))

	case "auto-summon":
		val := strings.ToLower(value)
		cfg.AutoSummonSpecialists = val == "true" || val == "1" || val == "yes"
		if err := config.Save(cfg); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Auto-summon specialists: %v", cfg.AutoSummonSpecialists))

	case "max-advisors":
		num, err := strconv.Atoi(value)
		if err != nil || num < 1 || num > 7 {
			output.PrintError("Invalid value. Use a number between 1 and 7")
			return nil
		}
		cfg.MaxAdvisors = num
		if err := config.Save(cfg); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Max advisors set to: %d", num))

	default:
		output.PrintError(fmt.Sprintf("Unknown config key: %s", key))
	}

	return nil
}

func handleConfigGet(args []string) error {
	if len(args) < 1 {
		output.PrintError("Usage: cto-advisory config get <key>")
		return nil
	}

	key := args[0]
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch key {
	case "api-key":
		if cfg.APIKey != "" {
			fmt.Println(config.MaskAPIKey(cfg.APIKey))
		} else {
			fmt.Println("Not set")
		}
	case "model":
		fmt.Println(cfg.Model)
	case "default-mode":
		fmt.Println(cfg.DefaultMode)
	case "default-advisors":
		advisorStrs := make([]string, len(cfg.DefaultAdvisors))
		for i, a := range cfg.DefaultAdvisors {
			advisorStrs[i] = string(a)
		}
		fmt.Println(strings.Join(advisorStrs, ", "))
	case "auto-summon":
		fmt.Println(cfg.AutoSummonSpecialists)
	case "max-advisors":
		fmt.Println(cfg.MaxAdvisors)
	default:
		output.PrintError(fmt.Sprintf("Unknown config key: %s", key))
	}

	return nil
}

func handleConfigList() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	titleStyle := lipgloss.NewStyle().Bold(true)
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	fmt.Println()
	fmt.Println(titleStyle.Render("CTO Advisory Board Configuration"))
	fmt.Println()

	if !config.IsInitialized() {
		fmt.Println(dimStyle.Render("⚠ Project not initialized. Run: cto-advisory init"))
		fmt.Println()
	}

	// API Key
	maskedKey := "Not set"
	if cfg.APIKey != "" {
		maskedKey = config.MaskAPIKey(cfg.APIKey)
	}
	fmt.Printf("  %s:          %s\n", keyStyle.Render("api-key"), maskedKey)

	// Model
	fmt.Printf("  %s:            %s\n", keyStyle.Render("model"), cfg.Model)

	// Default mode
	fmt.Printf("  %s:     %s\n", keyStyle.Render("default-mode"), cfg.DefaultMode)

	// Default advisors
	advisorStrs := make([]string, len(cfg.DefaultAdvisors))
	for i, a := range cfg.DefaultAdvisors {
		advisorStrs[i] = string(a)
	}
	fmt.Printf("  %s: %s\n", keyStyle.Render("default-advisors"), strings.Join(advisorStrs, ", "))

	// Auto-summon
	fmt.Printf("  %s:      %v\n", keyStyle.Render("auto-summon"), cfg.AutoSummonSpecialists)

	// Max advisors
	fmt.Printf("  %s:     %d\n", keyStyle.Render("max-advisors"), cfg.MaxAdvisors)

	fmt.Println()

	return nil
}
