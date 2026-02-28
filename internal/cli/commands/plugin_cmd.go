// Package commands implements the plugin management commands.
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	tmpl "text/template"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/carlosinfantes/cio/internal/cli/output"
	"github.com/carlosinfantes/cio/internal/config"
	"github.com/carlosinfantes/cio/internal/plugins"
	"github.com/carlosinfantes/cio/internal/plugins/remote"
	"github.com/carlosinfantes/cio/internal/types"
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
	cmd.AddCommand(newPluginUpdateCmd())

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
		fmt.Println("Install a plugin with: cio plugin install <domain>")
		fmt.Println("Search available plugins: cio plugin search")
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

	// Show stars and downloads from registry if available
	registryStats := fetchRegistryStats(domain)
	if registryStats != nil {
		starsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
		downloadsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
		if registryStats.Stars > 0 {
			fmt.Printf("%s %s\n", labelStyle.Render("Stars:"), starsStyle.Render(fmt.Sprintf("★ %s", formatCount(registryStats.Stars))))
		}
		if registryStats.Downloads > 0 {
			fmt.Printf("%s %s\n", labelStyle.Render("Downloads:"), downloadsStyle.Render(fmt.Sprintf("↓ %s", formatCount(registryStats.Downloads))))
		}
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

	// Get registry URL from config
	cfg, _ := config.Load()
	registryURL := cfg.RegistryURL
	if registryURL == "" {
		registryURL = types.DefaultRegistryURL
	}

	// Get installed plugins directory
	installedDir, err := config.GetInstalledPluginsDir()
	if err != nil {
		return fmt.Errorf("getting plugins directory: %w", err)
	}

	// Download from registry (try configured URL, fall back to default)
	client := remote.NewClient(registryURL)
	downloader := remote.NewDownloader(client)

	// Show download info first
	info, err := downloader.GetDownloadInfo(domain, installedDir)
	if err != nil && registryURL != types.DefaultRegistryURL {
		// Retry with default registry URL
		client = remote.NewClient(types.DefaultRegistryURL)
		downloader = remote.NewDownloader(client)
		info, err = downloader.GetDownloadInfo(domain, installedDir)
	}
	if err != nil {
		output.PrintError(fmt.Sprintf("Plugin not found in registry: %s", domain))
		fmt.Println()
		fmt.Println("To manually install a plugin:")
		fmt.Printf("  1. Download the plugin package for '%s'\n", domain)
		fmt.Printf("  2. Extract to: %s/%s/\n", installedDir, domain)
		fmt.Println("  3. Ensure manifest.yaml exists in the plugin directory")
		return nil
	}

	// Show stars/downloads info
	if entry, _ := client.GetPlugin(domain); entry != nil {
		stats := formatStats(entry.Stars, entry.Downloads)
		if stats != "" {
			statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
			fmt.Printf("  %s\n", statsStyle.Render(stats))
		}
	}

	output.PrintInfo(fmt.Sprintf("Downloading %s v%s...", info.Domain, info.Version))

	if err := downloader.Download(domain, installedDir); err != nil {
		output.PrintError(fmt.Sprintf("Download failed: %v", err))
		return nil
	}

	// Track in config
	config.Update(func(cfg *types.Config) {
		for _, d := range cfg.InstalledDomains {
			if d == domain {
				return
			}
		}
		cfg.InstalledDomains = append(cfg.InstalledDomains, domain)
	})

	output.PrintSuccess(fmt.Sprintf("Plugin %s installed to %s", domain, info.InstalledDir))
	fmt.Printf("\nActivate with: cio plugin use %s\n", domain)

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

	// Update config: remove from installed list and clear active if needed
	config.Update(func(cfg *types.Config) {
		filtered := cfg.InstalledDomains[:0]
		for _, d := range cfg.InstalledDomains {
			if d != domain {
				filtered = append(filtered, d)
			}
		}
		cfg.InstalledDomains = filtered
		if cfg.ActiveDomain == domain {
			cfg.ActiveDomain = ""
		}
	})

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

	// Try to use templates
	templateDir := findTemplatesDir()
	if templateDir != "" {
		if err := createFromTemplates(templateDir, pluginDir, domain); err != nil {
			output.PrintInfo(fmt.Sprintf("Template processing failed (%v), using defaults", err))
			createFromDefaults(pluginDir, domain)
		}
	} else {
		createFromDefaults(pluginDir, domain)
	}

	output.PrintSuccess(fmt.Sprintf("Plugin created at: %s", pluginDir))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit manifest.yaml to customize advisors")
	fmt.Println("  2. Add persona files to personas/")
	fmt.Println("  3. Add cognitive processes to cognitive/")
	fmt.Printf("  4. Activate with: cio plugin use %s\n", domain)

	return nil
}

// templateData holds variables for plugin templates.
type templateData struct {
	Domain      string
	DisplayName string
	Author      string
}

// findTemplatesDir locates the plugin-templates/default directory.
func findTemplatesDir() string {
	// Try relative to executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		for _, rel := range []string{"plugin-templates/default", "../plugin-templates/default"} {
			dir := filepath.Join(execDir, rel)
			if _, err := os.Stat(dir); err == nil {
				return dir
			}
		}
	}

	// Try current working directory
	cwd, err := os.Getwd()
	if err == nil {
		dir := filepath.Join(cwd, "plugin-templates", "default")
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}

	return ""
}

// createFromTemplates scaffolds a plugin using .tmpl files from the templates directory.
func createFromTemplates(templateDir, pluginDir, domain string) error {
	data := map[string]interface{}{
		"Domain":      domain,
		"DisplayName": titleCase(domain) + " Advisory Board",
		"Author":      "Custom",
	}

	return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Compute relative path and destination
		relPath, _ := filepath.Rel(templateDir, path)
		destPath := filepath.Join(pluginDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Process .tmpl files with text/template
		if strings.HasSuffix(path, ".tmpl") {
			destPath = strings.TrimSuffix(destPath, ".tmpl")
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			t, err := tmpl.New(filepath.Base(path)).Option("missingkey=zero").Parse(string(content))
			if err != nil {
				return fmt.Errorf("parsing template %s: %w", relPath, err)
			}

			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}

			f, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer f.Close()

			return t.Execute(f, data)
		}

		// Copy non-template files as-is
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(destPath, content, 0644)
	})
}

// createFromDefaults generates a plugin using hardcoded defaults (fallback).
func createFromDefaults(pluginDir, domain string) {
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
		os.MkdirAll(dir, 0755)
	}

	manifest := &plugins.Manifest{
		Domain:      domain,
		Version:     "0.1.0",
		DisplayName: fmt.Sprintf("%s Advisory Board", titleCase(domain)),
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
	plugins.SaveManifest(manifest, manifestPath)

	settingsContent := "# Plugin settings for " + domain + "\ndefault_mode: panel\nstart_in_discovery: true\nmax_advisors: 5\n"
	settingsPath := filepath.Join(pluginDir, "settings.yaml")
	os.WriteFile(settingsPath, []byte(settingsContent), 0644)
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
		fmt.Println("Install it first with: cio plugin install " + domain)
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

	// Get registry URL from config
	cfg, _ := config.Load()
	registryURL := cfg.RegistryURL
	if registryURL == "" {
		registryURL = types.DefaultRegistryURL
	}

	// Search the registry
	client := remote.NewClient(registryURL)
	results, err := client.Search(query)
	if err != nil {
		// Fall back to default index if registry is unreachable
		output.PrintInfo("Registry unreachable, showing cached results...")
		defaultIndex := remote.GetDefaultIndex()
		results = defaultIndex.Plugins
		if query != "" {
			var filtered []remote.PluginEntry
			for _, p := range results {
				if p.Matches(query) {
					filtered = append(filtered, p)
				}
			}
			results = filtered
		}
	}

	fmt.Println()
	if len(results) == 0 {
		fmt.Printf("No plugins found matching '%s'\n", query)
	} else {
		fmt.Println("Available plugins:")
		fmt.Println()

		featuredStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
		statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

		for _, p := range results {
			featured := ""
			if p.Featured {
				featured = featuredStyle.Render(" ★")
			}
			emoji := ""
			if p.Emoji != "" {
				emoji = p.Emoji + " "
			}
			fmt.Printf("  %s%s%s (%s)\n", emoji, p.Domain, featured, p.Version)
			if p.DisplayName != "" {
				fmt.Printf("    %s\n", p.DisplayName)
			}
			if p.Description != "" {
				desc := p.Description
				if len(desc) > 70 {
					desc = desc[:67] + "..."
				}
				fmt.Printf("    %s\n", desc)
			}
			// Stars and downloads
			stats := formatStats(p.Stars, p.Downloads)
			if stats != "" {
				fmt.Printf("    %s\n", statsStyle.Render(stats))
			}
			fmt.Println()
		}
	}

	fmt.Println("Install with: cio plugin install <domain>")

	return nil
}

// plugin update
func newPluginUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update [domain]",
		Short: "Update plugins to latest versions from the registry",
		Long:  "Check for and install newer versions of installed plugins.",
		RunE:  runPluginUpdate,
	}
}

func runPluginUpdate(cmd *cobra.Command, args []string) error {
	// Get registry client
	cfg, _ := config.Load()
	registryURL := cfg.RegistryURL
	if registryURL == "" {
		registryURL = types.DefaultRegistryURL
	}
	client := remote.NewClient(registryURL)

	// Fetch latest index
	output.PrintInfo("Checking for updates...")
	index, err := client.FetchIndex()
	if err != nil {
		output.PrintError(fmt.Sprintf("Could not reach registry: %v", err))
		return nil
	}

	// Build registry lookup
	registryVersions := make(map[string]string)
	for _, p := range index.Plugins {
		registryVersions[p.Domain] = p.Version
	}

	// Get installed plugins
	packages, err := plugins.LoadInstalledPlugins()
	if err != nil {
		return err
	}

	// Filter to specific domain if provided
	if len(args) > 0 {
		domain := args[0]
		var filtered []*plugins.PluginPackage
		for _, pkg := range packages {
			if pkg.Manifest.Domain == domain {
				filtered = append(filtered, pkg)
			}
		}
		if len(filtered) == 0 {
			output.PrintError(fmt.Sprintf("Plugin %s is not installed", domain))
			return nil
		}
		packages = filtered
	}

	installedDir, err := config.GetInstalledPluginsDir()
	if err != nil {
		return fmt.Errorf("getting plugins directory: %w", err)
	}

	updated := 0
	for _, pkg := range packages {
		domain := pkg.Manifest.Domain
		localVersion := pkg.Manifest.Version
		remoteVersion, exists := registryVersions[domain]

		if !exists {
			continue // not in registry, skip
		}

		if remoteVersion == localVersion {
			fmt.Printf("  %s %s — up to date (v%s)\n", pkg.Manifest.Facilitator.Emoji, domain, localVersion)
			continue
		}

		fmt.Printf("  %s %s — updating v%s → v%s\n", pkg.Manifest.Facilitator.Emoji, domain, localVersion, remoteVersion)

		// Remove old version
		os.RemoveAll(pkg.Path)

		// Download new version
		downloader := remote.NewDownloader(client)
		if err := downloader.Download(domain, installedDir); err != nil {
			output.PrintError(fmt.Sprintf("Failed to update %s: %v", domain, err))
			continue
		}
		updated++
	}

	fmt.Println()
	if updated == 0 {
		output.PrintSuccess("All plugins are up to date.")
	} else {
		output.PrintSuccess(fmt.Sprintf("Updated %d plugin(s).", updated))
	}

	return nil
}

// formatCount formats a number with K/M suffixes for display.
func formatCount(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

// formatStats builds a combined stars + downloads string.
func formatStats(stars, downloads int) string {
	var parts []string
	if stars > 0 {
		parts = append(parts, fmt.Sprintf("★ %s", formatCount(stars)))
	}
	if downloads > 0 {
		parts = append(parts, fmt.Sprintf("↓ %s", formatCount(downloads)))
	}
	return strings.Join(parts, "  ")
}

// fetchRegistryStats fetches stars/downloads for a plugin from the registry.
func fetchRegistryStats(domain string) *remote.PluginEntry {
	cfg, _ := config.Load()
	registryURL := cfg.RegistryURL
	if registryURL == "" {
		registryURL = types.DefaultRegistryURL
	}
	client := remote.NewClient(registryURL)
	entry, err := client.GetPlugin(domain)
	if err != nil {
		return nil
	}
	return entry
}

// titleCase capitalizes the first letter of a string.
func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	first := s[0]
	if first >= 'a' && first <= 'z' {
		return string(first-32) + s[1:]
	}
	return s
}
