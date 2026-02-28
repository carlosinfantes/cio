// Package wizard implements the interactive setup wizard for new projects.
package wizard

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/carlosinfantes/cio/internal/config"
	"github.com/carlosinfantes/cio/internal/types"
)

// Wizard orchestrates the interactive setup process.
type Wizard struct {
	reader *bufio.Reader
	config types.Config
}

// New creates a new Wizard instance.
func New() *Wizard {
	return &Wizard{
		reader: bufio.NewReader(os.Stdin),
		config: types.DefaultConfig(),
	}
}

// Run executes the full wizard flow and returns the configured settings.
func (w *Wizard) Run() (*types.Config, error) {
	w.printWelcome()

	// Step 1: API Key
	if err := w.stepAPIKey(); err != nil {
		return nil, err
	}

	// Step 2: Domain Selection
	if err := w.stepDomainSelection(); err != nil {
		return nil, err
	}

	// Step 3: Model Preference
	if err := w.stepModelPreference(); err != nil {
		return nil, err
	}

	// Step 4: Cognitive Style
	if err := w.stepCognitiveStyle(); err != nil {
		return nil, err
	}

	// Step 5: Organization Context
	if err := w.stepOrganizationContext(); err != nil {
		return nil, err
	}

	// Summary
	w.printSummary()

	// Save configuration
	if err := config.Save(w.config); err != nil {
		return nil, fmt.Errorf("saving config: %w", err)
	}

	return &w.config, nil
}

// printWelcome displays the welcome banner.
func (w *Wizard) printWelcome() {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                               ║")
	fmt.Println("║           Welcome to the CIO - Chief Intelligence Officer                   ║")
	fmt.Println("║                                                               ║")
	fmt.Println("║   AI-powered advisory boards for technical decision-making    ║")
	fmt.Println("║                                                               ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Let's set up your advisory board. Press Enter to accept defaults.")
	fmt.Println()
}

// printSummary displays the configuration summary.
func (w *Wizard) printSummary() {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Configuration Summary                      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  API Key:      %s\n", maskAPIKey(w.config.APIKey))
	fmt.Printf("  Domain:       %s\n", w.config.ActiveDomain)
	fmt.Printf("  Model:        %s\n", w.config.Model)
	fmt.Printf("  Mode:         %s\n", w.config.DefaultMode)
	fmt.Printf("  Discovery:    %v\n", w.config.StartInDiscovery)
	fmt.Println()
	fmt.Println("Configuration saved to .cio/config.yaml")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Install a domain plugin:  cto plugin install cio")
	fmt.Println("  2. Start an advisory session: cto ask \"Your question here\"")
	fmt.Println()
}

// promptWithDefault asks for input with a default value.
func (w *Wizard) promptWithDefault(prompt, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := w.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

// promptRequired asks for required input (no default).
func (w *Wizard) promptRequired(prompt string) (string, error) {
	for {
		fmt.Printf("%s: ", prompt)
		input, err := w.reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input = strings.TrimSpace(input)
		if input != "" {
			return input, nil
		}
		fmt.Println("  This field is required. Please enter a value.")
	}
}

// promptSelect displays options and returns selected index.
func (w *Wizard) promptSelect(prompt string, options []string, defaultIdx int) (int, error) {
	fmt.Println(prompt)
	for i, opt := range options {
		marker := "  "
		if i == defaultIdx {
			marker = "> "
		}
		fmt.Printf("  %s%d. %s\n", marker, i+1, opt)
	}

	defaultStr := ""
	if defaultIdx >= 0 && defaultIdx < len(options) {
		defaultStr = fmt.Sprintf("%d", defaultIdx+1)
	}

	for {
		input, err := w.promptWithDefault("Select option", defaultStr)
		if err != nil {
			return defaultIdx, err
		}

		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil {
			if idx >= 1 && idx <= len(options) {
				return idx - 1, nil
			}
		}
		fmt.Println("  Invalid selection. Please enter a number from the list.")
	}
}

// maskAPIKey returns a masked version of the API key.
func maskAPIKey(key string) string {
	if len(key) < 14 {
		return "****"
	}
	return key[:10] + "..." + key[len(key)-4:]
}
