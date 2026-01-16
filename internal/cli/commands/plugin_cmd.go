// Package commands implements the plugin management commands.
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/carlosinfantes/cto-advisory-board/internal/cli/output"
	"github.com/carlosinfantes/cto-advisory-board/internal/config"
	"github.com/carlosinfantes/cto-advisory-board/internal/plugins"
	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

func init() {
	rootCmd.AddCommand(newPluginCmd())
}

func newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage advisory board plugins/domains",
		Long:  "Install, list, and manage domain-specific advisory board plugins.",
	}

	cmd.AddCommand(newPluginListCmd())
	cmd.AddCommand(newPluginInfoCmd())
	cmd.AddCommand(newPluginInstallCmd())
	cmd.AddCommand(newPluginUninstallCmd())
	cmd.AddCommand(newPluginCreateCmd())
	cmd.AddCommand(newPluginUseCmd())
	cmd.AddCommand(newPluginSearchCmd())

	return cmd
}

// plugin list
func newPluginListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		RunE:  runPluginList,
	}
}

func runPluginList(cmd *cobra.Command, args []string) error {
	packages, err := plugins.LoadInstalledPlugins()
	if err != nil {
		return err
	}

	// Get active domain from config
	cfg, _ := config.Load()
	activeDomain := cfg.ActiveDomain

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))

	fmt.Println(headerStyle.Render("Installed Plugins"))
	fmt.Println(strings.Repeat("─", 50))

	if len(packages) == 0 {
		fmt.Println()
		fmt.Println("No plugins installed.")
		fmt.Println()
		fmt.Println("Install a plugin with: cto plugin install <domain>")
		fmt.Println("Search available plugins: cto plugin search")
		return nil
	}

	for _, pkg := range packages {
		marker := "  "
		if pkg.Manifest.Domain == activeDomain {
			marker = activeStyle.Render("▸ ")
		}

		fmt.Printf("%s%s (%s)\n", marker, pkg.Manifest.Domain, pkg.Manifest.Version)
		if pkg.Manifest.DisplayName != "" {
			fmt.Printf("    %s\n", pkg.Manifest.DisplayName)
		}
		if pkg.Manifest.Description != "" {
			// Truncate long descriptions
			desc := pkg.Manifest.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Printf("    %s\n", desc)
		}
		fmt.Printf("    Advisors: %d core, %d specialists\n",
			len(pkg.Manifest.CoreAdvisors), len(pkg.Manifest.Specialists))
	}

	return nil
}

// plugin info
func newPluginInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <domain>",
		Short: "Show detailed information about a plugin",
		Args:  cobra.ExactArgs(1),
		RunE:  runPluginInfo,
	}
}

func runPluginInfo(cmd *cobra.Command, args []string) error {
	domain := args[0]

	pkg, err := plugins.FindPlugin(domain)
	if err != nil {
		output.PrintError(fmt.Sprintf("Plugin not found: %s", domain))
		return nil
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	fmt.Println(headerStyle.Render(pkg.Manifest.DisplayName))
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	fmt.Printf("%s %s\n", labelStyle.Render("Domain:"), pkg.Manifest.Domain)
	fmt.Printf("%s %s\n", labelStyle.Render("Version:"), pkg.Manifest.Version)
	if pkg.Manifest.Author != "" {
		fmt.Printf("%s %s\n", labelStyle.Render("Author:"), pkg.Manifest.Author)
	}
	if pkg.Manifest.License != "" {
		fmt.Printf("%s %s\n", labelStyle.Render("License:"), pkg.Manifest.License)
	}
	if pkg.Manifest.Homepage != "" {
		fmt.Printf("%s %s\n", labelStyle.Render("Homepage:"), pkg.Manifest.Homepage)
	}
	fmt.Println()

	if pkg.Manifest.Description != "" {
		fmt.Println(labelStyle.Render("Description:"))
		fmt.Printf("  %s\n", pkg.Manifest.Description)
		fmt.Println()
	}

	// Facilitator
	fmt.Println(labelStyle.Render("Facilitator:"))
	fmt.Printf("  %s %s - %s\n",
		pkg.Manifest.Facilitator.Emoji,
		pkg.Manifest.Facilitator.Name,
		pkg.Manifest.Facilitator.Role)
	fmt.Println()

	// Advisors
	fmt.Println(labelStyle.Render("Core Advisors:"))
	for _, advisor := range pkg.Manifest.CoreAdvisors {
		fmt.Printf("  %s %s - %s\n", advisor.Emoji, advisor.Name, advisor.Role)
	}
	fmt.Println()

	if len(pkg.Manifest.Specialists) > 0 {
		fmt.Println(labelStyle.Render("Specialists:"))
		for _, specialist := range pkg.Manifest.Specialists {
			fmt.Printf("  %s %s - %s\n", specialist.Emoji, specialist.Name, specialist.Role)
			if len(specialist.Keywords) > 0 {
				fmt.Printf("      Keywords: %s\n", strings.Join(specialist.Keywords, ", "))
			}
		}
		fmt.Println()
	}

	if len(pkg.Manifest.DecisionDomains) > 0 {
		fmt.Println(labelStyle.Render("Decision Domains:"))
		fmt.Printf("  %s\n", strings.Join(pkg.Manifest.DecisionDomains, ", "))
	}

	return nil
}

// plugin install
func newPluginInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <domain>",
		Short: "Install a plugin from the registry",
		Long:  "Download and install a domain plugin from the central registry.",
		Args:  cobra.ExactArgs(1),
		RunE:  runPluginInstall,
	}
}

func runPluginInstall(cmd *cobra.Command, args []string) error {
	domain := args[0]

	output.PrintInfo(fmt.Sprintf("Installing plugin: %s", domain))

	// Check if already installed
	if pkg, _ := plugins.FindPlugin(domain); pkg != nil {
		output.PrintInfo(fmt.Sprintf("Plugin %s is already installed (version %s)", domain, pkg.Manifest.Version))
		return nil
	}

	// TODO: Implement actual download from registry
	// For now, show placeholder message
	output.PrintInfo("Registry download not yet implemented.")
	fmt.Println()
	fmt.Println("To manually install a plugin:")
	fmt.Printf("  1. Download the plugin package for '%s'\n", domain)

	installedDir, _ := config.GetInstalledPluginsDir()
	fmt.Printf("  2. Extract to: %s/%s/\n", installedDir, domain)
	fmt.Println("  3. Ensure manifest.yaml exists in the plugin directory")
	fmt.Println()
	fmt.Println("Registry URL: https://github.com/cto-advisory-board/plugin-registry")

	return nil
}

// plugin uninstall
func newPluginUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <domain>",
		Short: "Uninstall a plugin",
		Args:  cobra.ExactArgs(1),
		RunE:  runPluginUninstall,
	}
}

func runPluginUninstall(cmd *cobra.Command, args []string) error {
	domain := args[0]

	// Find plugin path
	pkg, err := plugins.FindPlugin(domain)
	if err != nil {
		output.PrintError(fmt.Sprintf("Plugin not found: %s", domain))
		return nil
	}

	output.PrintInfo(fmt.Sprintf("Uninstalling plugin: %s", domain))

	// Remove the plugin directory
	if err := os.RemoveAll(pkg.Path); err != nil {
		return fmt.Errorf("removing plugin: %w", err)
	}

	output.PrintSuccess(fmt.Sprintf("Plugin %s uninstalled", domain))

	// Update config if this was the active domain
	cfg, _ := config.Load()
	if cfg.ActiveDomain == domain {
		cfg.ActiveDomain = ""
		if err := config.Save(cfg); err != nil {
			output.PrintInfo("Note: You may want to set a new active domain with 'cto plugin use <domain>'")
		}
	}

	return nil
}

// plugin create
func newPluginCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <domain>",
		Short: "Create a new custom plugin",
		Long:  "Scaffold a new custom plugin in the local plugins directory.",
		Args:  cobra.ExactArgs(1),
		RunE:  runPluginCreate,
	}
}

func runPluginCreate(cmd *cobra.Command, args []string) error {
	domain := args[0]

	// Check if already exists
	if pkg, _ := plugins.FindPlugin(domain); pkg != nil {
		output.PrintError(fmt.Sprintf("Plugin %s already exists", domain))
		return nil
	}

	customDir, err := config.GetCustomPluginsDir()
	if err != nil {
		return err
	}

	pluginDir := filepath.Join(customDir, domain)

	output.PrintInfo(fmt.Sprintf("Creating plugin: %s", domain))

	// Create directory structure
	dirs := []string{
		pluginDir,
		filepath.Join(pluginDir, "personas"),
		filepath.Join(pluginDir, "personas", "specialists"),
		filepath.Join(pluginDir, "cognitive"),
		filepath.Join(pluginDir, "cognitive", "frameworks"),
		filepath.Join(pluginDir, "prompts"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	// Create manifest template
	manifest := &plugins.Manifest{
		Domain:      domain,
		Version:     "0.1.0",
		DisplayName: fmt.Sprintf("%s Advisory Board", strings.Title(domain)),
		Description: fmt.Sprintf("Custom advisory board for %s domain", domain),
		Facilitator: plugins.FacilitatorConfig{
			ID:            "facilitator",
			Name:          "Jordan",
			Role:          "Discovery Coach",
			Emoji:         "💭",
			Color:         "141",
			ThinkingStyle: "Socratic questioning to draw out the full picture",
		},
		CoreAdvisors: []plugins.AdvisorConfig{
			{
				ID:            "advisor-1",
				Name:          "Expert One",
				Role:          "Domain Expert",
				Emoji:         "🎯",
				Color:         "39",
				ThinkingStyle: "Strategic, long-term perspective",
				Priorities:    []string{"Quality", "Efficiency"},
			},
			{
				ID:            "advisor-2",
				Name:          "Expert Two",
				Role:          "Technical Advisor",
				Emoji:         "🔧",
				Color:         "208",
				ThinkingStyle: "Practical, implementation-focused",
				Priorities:    []string{"Feasibility", "Best practices"},
			},
		},
		DecisionDomains: []string{"strategy", "operations", "technology"},
		Author:          "Custom",
		License:         "MIT",
	}

	manifestPath := filepath.Join(pluginDir, "manifest.yaml")
	if err := plugins.SaveManifest(manifest, manifestPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	// Create settings template
	settingsContent := `# Plugin settings for ` + domain + `

# Default interaction mode
default_mode: panel

# Default advisors to include
default_advisors:
  - advisor-1
  - advisor-2

# Whether to start with the facilitator
start_in_discovery: true

# Maximum advisors in a session
max_advisors: 5

# Custom domain-specific settings
custom: {}
`
	settingsPath := filepath.Join(pluginDir, "settings.yaml")
	if err := os.WriteFile(settingsPath, []byte(settingsContent), 0644); err != nil {
		return fmt.Errorf("saving settings: %w", err)
	}

	output.PrintSuccess(fmt.Sprintf("Plugin created at: %s", pluginDir))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit manifest.yaml to customize advisors")
	fmt.Println("  2. Add persona files to personas/")
	fmt.Println("  3. Add cognitive processes to cognitive/")
	fmt.Printf("  4. Activate with: cto plugin use %s\n", domain)

	return nil
}

// plugin use
func newPluginUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <domain>",
		Short: "Set the active plugin/domain",
		Args:  cobra.ExactArgs(1),
		RunE:  runPluginUse,
	}
}

func runPluginUse(cmd *cobra.Command, args []string) error {
	domain := args[0]

	// Verify plugin exists
	pkg, err := plugins.FindPlugin(domain)
	if err != nil {
		output.PrintError(fmt.Sprintf("Plugin not found: %s", domain))
		fmt.Println()
		fmt.Println("Install it first with: cto plugin install " + domain)
		return nil
	}

	// Update config
	if err := config.Update(func(cfg *types.Config) {
		cfg.ActiveDomain = domain
	}); err != nil {
		return fmt.Errorf("updating config: %w", err)
	}

	output.PrintSuccess(fmt.Sprintf("Active domain set to: %s (%s)", domain, pkg.Manifest.DisplayName))

	return nil
}

// plugin search
func newPluginSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search the plugin registry",
		RunE:  runPluginSearch,
	}
}

func runPluginSearch(cmd *cobra.Command, args []string) error {
	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	output.PrintInfo("Searching plugin registry...")

	// TODO: Implement actual registry search
	// For now, show available domains
	fmt.Println()
	fmt.Println("Available domains in registry:")
	fmt.Println()

	domains := []struct {
		name        string
		description string
	}{
		{"cto-advisory", "AI-powered executive committee for CTOs"},
		{"legal-advisory", "Legal counsel for business decisions"},
		{"medical-advisory", "Healthcare practice advisory"},
		{"architecture-advisory", "Building & construction advisory"},
		{"finance-advisory", "Financial planning and strategy"},
	}

	for _, d := range domains {
		if query == "" || strings.Contains(strings.ToLower(d.name), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(d.description), strings.ToLower(query)) {
			fmt.Printf("  %s\n", d.name)
			fmt.Printf("    %s\n", d.description)
			fmt.Println()
		}
	}

	fmt.Println("Install with: cto plugin install <domain>")
	fmt.Println()
	fmt.Println("Registry: https://github.com/cto-advisory-board/plugin-registry")

	return nil
}
