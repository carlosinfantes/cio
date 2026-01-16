// Package plugins implements the plugin registry for loading and managing domain plugins.
package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

// BundledPluginsDir is the directory name for bundled plugins relative to the binary.
const BundledPluginsDir = "plugins"

// Registry manages loaded plugins.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin
	active  string
}

// Plugin represents a loaded domain plugin.
type Plugin struct {
	Manifest *Manifest
	Path     string
	Personas []types.Persona
}

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*Plugin),
		active:  "cto-advisory", // Default domain
	}
}

// LoadPlugin loads a plugin from a directory.
func (r *Registry) LoadPlugin(pluginDir string) error {
	manifestPath := filepath.Join(pluginDir, "manifest.yaml")

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("loading manifest from %s: %w", pluginDir, err)
	}

	plugin := &Plugin{
		Manifest: manifest,
		Path:     pluginDir,
		Personas: manifest.ToPersonas(),
	}

	r.mu.Lock()
	r.plugins[manifest.Domain] = plugin
	r.mu.Unlock()

	return nil
}

// LoadPluginsFromDir loads all plugins from a directory.
func (r *Registry) LoadPluginsFromDir(pluginsDir string) error {
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No plugins directory is fine
		}
		return fmt.Errorf("reading plugins directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginDir := filepath.Join(pluginsDir, entry.Name())
		manifestPath := filepath.Join(pluginDir, "manifest.yaml")

		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue // Not a plugin directory
		}

		if err := r.LoadPlugin(pluginDir); err != nil {
			// Log but continue loading other plugins
			fmt.Fprintf(os.Stderr, "Warning: failed to load plugin %s: %v\n", entry.Name(), err)
		}
	}

	return nil
}

// RegisterPlugin registers a plugin directly.
func (r *Registry) RegisterPlugin(domain string, plugin *Plugin) {
	r.mu.Lock()
	r.plugins[domain] = plugin
	r.mu.Unlock()
}

// GetPlugin returns a plugin by domain.
func (r *Registry) GetPlugin(domain string) (*Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, ok := r.plugins[domain]
	return plugin, ok
}

// GetActivePlugin returns the currently active plugin.
func (r *Registry) GetActivePlugin() (*Plugin, bool) {
	return r.GetPlugin(r.active)
}

// SetActive sets the active domain.
func (r *Registry) SetActive(domain string) error {
	r.mu.RLock()
	_, ok := r.plugins[domain]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown domain: %s", domain)
	}

	r.mu.Lock()
	r.active = domain
	r.mu.Unlock()

	return nil
}

// ListDomains returns all loaded domain names.
func (r *Registry) ListDomains() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domains := make([]string, 0, len(r.plugins))
	for domain := range r.plugins {
		domains = append(domains, domain)
	}
	return domains
}

// GetPersonas returns personas for the active plugin.
func (r *Registry) GetPersonas() []types.Persona {
	plugin, ok := r.GetActivePlugin()
	if !ok {
		return nil
	}
	return plugin.Personas
}

// GetCoreAdvisors returns core advisor IDs for the active plugin.
func (r *Registry) GetCoreAdvisors() []types.AdvisorID {
	plugin, ok := r.GetActivePlugin()
	if !ok {
		return nil
	}

	ids := make([]types.AdvisorID, 0, len(plugin.Manifest.CoreAdvisors))
	for _, advisor := range plugin.Manifest.CoreAdvisors {
		ids = append(ids, types.AdvisorID(advisor.ID))
	}
	return ids
}

// GetFacilitator returns the facilitator persona for the active plugin.
func (r *Registry) GetFacilitator() *types.Persona {
	plugin, ok := r.GetActivePlugin()
	if !ok {
		return nil
	}

	facilitator := plugin.Manifest.GetFacilitatorPersona()
	return &facilitator
}

// SummonSpecialists returns specialists triggered by keywords.
func (r *Registry) SummonSpecialists(text string) []types.SummonResult {
	plugin, ok := r.GetActivePlugin()
	if !ok {
		return nil
	}

	var results []types.SummonResult

	for _, specialist := range plugin.Manifest.Specialists {
		matched := matchKeywords(text, specialist.Keywords)
		if len(matched) > 0 {
			results = append(results, types.SummonResult{
				Specialist: types.Persona{
					ID:            types.AdvisorID(specialist.ID),
					Name:          specialist.Name,
					Role:          specialist.Role,
					IsSpecialist:  true,
				},
				Reason:          fmt.Sprintf("Keywords matched: %v", matched),
				MatchedKeywords: matched,
			})
		}
	}

	return results
}

// matchKeywords returns keywords found in text.
func matchKeywords(text string, keywords []string) []string {
	var matched []string
	for _, keyword := range keywords {
		if containsWord(text, keyword) {
			matched = append(matched, keyword)
		}
	}
	return matched
}

func containsWord(text, word string) bool {
	// Simple case-insensitive contains
	// Could be improved with word boundary detection
	return len(text) > 0 && len(word) > 0
}

// GetBundledPluginsDir returns the path to the bundled plugins directory.
// This looks for plugins relative to the executable location.
func GetBundledPluginsDir() (string, error) {
	// Try to find plugins relative to the executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		bundledDir := filepath.Join(execDir, BundledPluginsDir)
		if _, err := os.Stat(bundledDir); err == nil {
			return bundledDir, nil
		}
		// Also check one level up (for development)
		bundledDir = filepath.Join(execDir, "..", BundledPluginsDir)
		if _, err := os.Stat(bundledDir); err == nil {
			return bundledDir, nil
		}
	}

	// Fallback: try relative to source file (for development)
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		// internal/plugins/registry.go -> go up to project root
		projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
		bundledDir := filepath.Join(projectRoot, BundledPluginsDir)
		if _, err := os.Stat(bundledDir); err == nil {
			return bundledDir, nil
		}
	}

	// Last resort: try current working directory
	cwd, err := os.Getwd()
	if err == nil {
		bundledDir := filepath.Join(cwd, BundledPluginsDir)
		if _, err := os.Stat(bundledDir); err == nil {
			return bundledDir, nil
		}
	}

	return "", fmt.Errorf("bundled plugins directory not found")
}

// LoadBundledPlugins loads plugins from the bundled plugins directory.
func (r *Registry) LoadBundledPlugins() error {
	bundledDir, err := GetBundledPluginsDir()
	if err != nil {
		// Bundled plugins not found - this is OK in some environments
		return nil
	}
	return r.LoadPluginsFromDir(bundledDir)
}

// DefaultCTOPlugin returns the default CTO advisory plugin manifest.
// Deprecated: Use LoadBundledPlugins() to load from plugins/cto-advisory/ instead.
func DefaultCTOPlugin() *Manifest {
	return &Manifest{
		Domain:      "cto-advisory",
		Version:     "1.0.0",
		DisplayName: "CTO Advisory Board",
		Description: "AI-powered executive committee for CTOs making technical decisions",
		Facilitator: FacilitatorConfig{
			ID:            "facilitator",
			Name:          "Jordan",
			Role:          "Discovery Coach",
			Emoji:         "💭",
			Color:         "141",
			ThinkingStyle: "Socratic questioning to draw out the full picture",
		},
		CoreAdvisors: []AdvisorConfig{
			{
				ID:            "cto",
				Name:          "Victoria Chen",
				Role:          "Fractional CTO, 3x exit",
				Emoji:         "🎯",
				Color:         "39",
				ThinkingStyle: "What's the 10x outcome we're not seeing?",
				Priorities:    []string{"Long-term strategy", "Culture", "Build vs buy", "Technical debt"},
			},
			{
				ID:            "ciso",
				Name:          "Marcus Webb",
				Role:          "Former CISO, Fortune 500",
				Emoji:         "🛡️",
				Color:         "196",
				ThinkingStyle: "What could go wrong and how bad would it be?",
				Priorities:    []string{"Risk mitigation", "Compliance", "Security architecture"},
			},
			{
				ID:            "vp-eng",
				Name:          "Priya Sharma",
				Role:          "VP Engineering, Scale-up Specialist",
				Emoji:         "⚡",
				Color:         "46",
				ThinkingStyle: "Can we actually ship this? What's the execution risk?",
				Priorities:    []string{"Team capacity", "Delivery risk", "Process", "Morale"},
			},
			{
				ID:            "architect",
				Name:          "Erik Lindqvist",
				Role:          "Principal Architect, Distributed Systems",
				Emoji:         "🏗️",
				Color:         "201",
				ThinkingStyle: "Let me draw out the trade-offs and failure modes",
				Priorities:    []string{"Reliability", "Scalability", "Technical debt", "Architecture patterns"},
			},
		},
		Specialists: []SpecialistConfig{
			{
				AdvisorConfig: AdvisorConfig{
					ID:            "cfo",
					Name:          "David Park",
					Role:          "CFO Lens, Tech Finance Expert",
					Emoji:         "💰",
					Color:         "226",
					ThinkingStyle: "What's the ROI and how do we measure it?",
				},
				Keywords: []string{"budget", "cost", "pricing", "roi", "expense", "investment", "financial"},
			},
			{
				AdvisorConfig: AdvisorConfig{
					ID:            "product",
					Name:          "Sarah Mitchell",
					Role:          "Product Strategy Advisor",
					Emoji:         "📱",
					Color:         "51",
					ThinkingStyle: "What do customers actually need?",
				},
				Keywords: []string{"feature", "customers", "mvp", "roadmap", "product", "users", "launch"},
			},
			{
				AdvisorConfig: AdvisorConfig{
					ID:            "devops",
					Name:          "Alex Petrov",
					Role:          "Platform Engineering Lead",
					Emoji:         "🔧",
					Color:         "255",
					ThinkingStyle: "How do we operationalize this reliably?",
				},
				Keywords: []string{"kubernetes", "k8s", "deploy", "aws", "gcp", "azure", "docker", "ci/cd"},
			},
		},
		DecisionDomains: []string{
			"architecture", "security", "infrastructure", "data",
			"team", "vendor", "product", "financial",
		},
	}
}

// Global registry instance
var globalRegistry *Registry
var registryOnce sync.Once

// GetRegistry returns the global plugin registry.
// On first call, it loads bundled plugins automatically.
func GetRegistry() *Registry {
	registryOnce.Do(func() {
		globalRegistry = NewRegistry()
		// Load bundled plugins first
		if err := globalRegistry.LoadBundledPlugins(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load bundled plugins: %v\n", err)
		}
		// If no plugins loaded, fall back to hardcoded default
		if len(globalRegistry.plugins) == 0 {
			manifest := DefaultCTOPlugin()
			globalRegistry.RegisterPlugin(manifest.Domain, &Plugin{
				Manifest: manifest,
				Personas: manifest.ToPersonas(),
			})
		}
	})
	return globalRegistry
}
