package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/carlosinfantes/cio/internal/config"
	"github.com/carlosinfantes/cio/internal/plugins"
)

// DomainInfo contains information about an available domain.
type DomainInfo struct {
	Domain      string
	DisplayName string
	Description string
	Version     string
	Installed   bool
	Source      string // "installed", "custom", "registry"
}

// GetAvailableDomains returns all available domains from local plugins.
func GetAvailableDomains() ([]DomainInfo, error) {
	var domains []DomainInfo

	// Check installed plugins
	installedDir, err := config.GetInstalledPluginsDir()
	if err == nil {
		installed, _ := scanPluginDirectory(installedDir, "installed")
		domains = append(domains, installed...)
	}

	// Check custom plugins
	customDir, err := config.GetCustomPluginsDir()
	if err == nil {
		custom, _ := scanPluginDirectory(customDir, "custom")
		domains = append(domains, custom...)
	}

	return domains, nil
}

// scanPluginDirectory scans a directory for plugin manifests.
func scanPluginDirectory(dir, source string) ([]DomainInfo, error) {
	var domains []DomainInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return domains, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(dir, entry.Name(), "manifest.yaml")
		manifest, err := plugins.LoadManifest(manifestPath)
		if err != nil {
			continue // Skip invalid plugins
		}

		domains = append(domains, DomainInfo{
			Domain:      manifest.Domain,
			DisplayName: manifest.DisplayName,
			Description: manifest.Description,
			Version:     manifest.Version,
			Installed:   true,
			Source:      source,
		})
	}

	return domains, nil
}

// GetInstalledDomains returns only installed domain names.
func GetInstalledDomains() ([]string, error) {
	domains, err := GetAvailableDomains()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, d := range domains {
		if d.Installed {
			names = append(names, d.Domain)
		}
	}
	return names, nil
}

// IsDomainInstalled checks if a domain is installed locally.
func IsDomainInstalled(domain string) (bool, error) {
	domains, err := GetAvailableDomains()
	if err != nil {
		return false, err
	}

	for _, d := range domains {
		if d.Domain == domain && d.Installed {
			return true, nil
		}
	}
	return false, nil
}

// GetDomainPath returns the path to a domain's plugin directory.
func GetDomainPath(domain string) (string, error) {
	// Check custom first (higher priority)
	customDir, err := config.GetCustomPluginsDir()
	if err == nil {
		customPath := filepath.Join(customDir, domain)
		if _, err := os.Stat(filepath.Join(customPath, "manifest.yaml")); err == nil {
			return customPath, nil
		}
	}

	// Check installed
	installedDir, err := config.GetInstalledPluginsDir()
	if err == nil {
		installedPath := filepath.Join(installedDir, domain)
		if _, err := os.Stat(filepath.Join(installedPath, "manifest.yaml")); err == nil {
			return installedPath, nil
		}
	}

	return "", fmt.Errorf("domain %s not found", domain)
}

// FormatDomainList returns a formatted string of available domains.
func FormatDomainList(domains []DomainInfo) string {
	if len(domains) == 0 {
		return "No domains installed. Use 'cto plugin install <domain>' to install one."
	}

	var sb strings.Builder
	for _, d := range domains {
		sb.WriteString(fmt.Sprintf("  %s (%s)\n", d.Domain, d.Version))
		if d.Description != "" {
			sb.WriteString(fmt.Sprintf("    %s\n", d.Description))
		}
		sb.WriteString(fmt.Sprintf("    Source: %s\n", d.Source))
		sb.WriteString("\n")
	}
	return sb.String()
}
