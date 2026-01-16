// Package plugins implements plugin loading for full plugin packages.
package plugins

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/carlosinfantes/cto-advisory-board/internal/config"
	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

// PluginPackage represents a fully loaded plugin with all its components.
type PluginPackage struct {
	// Path to the plugin directory
	Path string

	// Manifest contains the main plugin configuration
	Manifest *Manifest

	// Personas loaded from the personas/ directory
	Personas map[string]*PersonaFile

	// CognitiveProcesses loaded from the cognitive/ directory
	CognitiveProcesses map[string]*CognitiveProcess

	// Prompts loaded from the prompts/ directory
	Prompts map[string]string

	// Settings loaded from settings.yaml
	Settings *PluginSettings
}

// PersonaFile represents a detailed persona configuration file.
type PersonaFile struct {
	ID            string   `yaml:"id"`
	Name          string   `yaml:"name"`
	Role          string   `yaml:"role"`
	Emoji         string   `yaml:"emoji,omitempty"`
	Color         string   `yaml:"color,omitempty"`
	ThinkingStyle string   `yaml:"thinking_style"`
	Background    string   `yaml:"background,omitempty"`
	Priorities    []string `yaml:"priorities,omitempty"`
	CatchPhrases  []string `yaml:"catch_phrases,omitempty"`
	Keywords      []string `yaml:"keywords,omitempty"` // For specialists
	SystemPrompt  string   `yaml:"system_prompt,omitempty"`
	Traits        []string `yaml:"traits,omitempty"`
}

// CognitiveProcess represents a cognitive process definition.
type CognitiveProcess struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Phases      []CognitivePhase  `yaml:"phases,omitempty"`
	Prompts     map[string]string `yaml:"prompts,omitempty"`
}

// CognitivePhase represents a phase in a cognitive process.
type CognitivePhase struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Prompt      string `yaml:"prompt,omitempty"`
}

// LoadPluginPackage loads a complete plugin package from a directory.
func LoadPluginPackage(pluginDir string) (*PluginPackage, error) {
	pkg := &PluginPackage{
		Path:               pluginDir,
		Personas:           make(map[string]*PersonaFile),
		CognitiveProcesses: make(map[string]*CognitiveProcess),
		Prompts:            make(map[string]string),
	}

	// Load manifest (required)
	manifestPath := filepath.Join(pluginDir, "manifest.yaml")
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}
	pkg.Manifest = manifest

	// Load personas (optional)
	personasDir := filepath.Join(pluginDir, "personas")
	if _, err := os.Stat(personasDir); err == nil {
		if err := pkg.loadPersonas(personasDir); err != nil {
			return nil, fmt.Errorf("loading personas: %w", err)
		}
	}

	// Load cognitive processes (optional)
	cognitiveDir := filepath.Join(pluginDir, "cognitive")
	if _, err := os.Stat(cognitiveDir); err == nil {
		if err := pkg.loadCognitiveProcesses(cognitiveDir); err != nil {
			return nil, fmt.Errorf("loading cognitive processes: %w", err)
		}
	}

	// Load prompts (optional)
	promptsDir := filepath.Join(pluginDir, "prompts")
	if _, err := os.Stat(promptsDir); err == nil {
		if err := pkg.loadPrompts(promptsDir); err != nil {
			return nil, fmt.Errorf("loading prompts: %w", err)
		}
	}

	// Load settings (optional)
	settingsPath := filepath.Join(pluginDir, "settings.yaml")
	if _, err := os.Stat(settingsPath); err == nil {
		settings, err := loadSettings(settingsPath)
		if err != nil {
			return nil, fmt.Errorf("loading settings: %w", err)
		}
		pkg.Settings = settings
	}

	return pkg, nil
}

// loadPersonas loads persona files from the personas directory.
func (pkg *PluginPackage) loadPersonas(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var persona PersonaFile
		if err := yaml.Unmarshal(data, &persona); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		// Use filename without extension as key if ID not set
		key := persona.ID
		if key == "" {
			key = filepath.Base(path)
			key = key[:len(key)-len(filepath.Ext(key))]
		}

		pkg.Personas[key] = &persona
		return nil
	})
}

// loadCognitiveProcesses loads cognitive process files.
func (pkg *PluginPackage) loadCognitiveProcesses(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var process CognitiveProcess
		if err := yaml.Unmarshal(data, &process); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		// Use filename without extension as key
		key := filepath.Base(path)
		key = key[:len(key)-len(filepath.Ext(key))]

		pkg.CognitiveProcesses[key] = &process
		return nil
	})
}

// loadPrompts loads prompt template files.
func (pkg *PluginPackage) loadPrompts(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Use filename without extension as key
		key := filepath.Base(path)
		ext := filepath.Ext(key)
		if ext != "" {
			key = key[:len(key)-len(ext)]
		}

		pkg.Prompts[key] = string(data)
		return nil
	})
}

// loadSettings loads the settings.yaml file.
func loadSettings(path string) (*PluginSettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var settings PluginSettings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// GetAllPersonas returns all personas from both manifest and persona files.
func (pkg *PluginPackage) GetAllPersonas() []types.Persona {
	// Start with manifest personas
	personas := pkg.Manifest.ToPersonas()

	// Merge/override with persona files
	for _, pf := range pkg.Personas {
		found := false
		for i, p := range personas {
			if string(p.ID) == pf.ID {
				// Override with file data
				personas[i] = personaFileToPersona(pf)
				found = true
				break
			}
		}
		if !found {
			// Add new persona from file
			personas = append(personas, personaFileToPersona(pf))
		}
	}

	return personas
}

// personaFileToPersona converts a PersonaFile to a types.Persona.
func personaFileToPersona(pf *PersonaFile) types.Persona {
	return types.Persona{
		ID:                 types.AdvisorID(pf.ID),
		Name:               pf.Name,
		Role:               pf.Role,
		Color:              pf.Color,
		Emoji:              pf.Emoji,
		ThinkingStyle:      pf.ThinkingStyle,
		Background:         pf.Background,
		Priorities:         pf.Priorities,
		CatchPhrases:       pf.CatchPhrases,
		AutoSummonKeywords: pf.Keywords,
		IsSpecialist:       len(pf.Keywords) > 0,
	}
}

// GetPrompt returns a prompt by name, checking plugin first then defaults.
func (pkg *PluginPackage) GetPrompt(name string) string {
	if prompt, ok := pkg.Prompts[name]; ok {
		return prompt
	}
	return ""
}

// LoadInstalledPlugins loads all installed plugin packages.
func LoadInstalledPlugins() ([]*PluginPackage, error) {
	var packages []*PluginPackage

	// Load from installed directory
	installedDir, err := config.GetInstalledPluginsDir()
	if err == nil {
		installed, _ := loadPluginsFromDir(installedDir)
		packages = append(packages, installed...)
	}

	// Load from custom directory (higher priority)
	customDir, err := config.GetCustomPluginsDir()
	if err == nil {
		custom, _ := loadPluginsFromDir(customDir)
		packages = append(packages, custom...)
	}

	return packages, nil
}

// loadPluginsFromDir loads all plugin packages from a directory.
func loadPluginsFromDir(dir string) ([]*PluginPackage, error) {
	var packages []*PluginPackage

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return packages, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginDir := filepath.Join(dir, entry.Name())
		pkg, err := LoadPluginPackage(pluginDir)
		if err != nil {
			continue // Skip invalid plugins
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}

// FindPlugin finds a plugin by domain name.
func FindPlugin(domain string) (*PluginPackage, error) {
	packages, err := LoadInstalledPlugins()
	if err != nil {
		return nil, err
	}

	for _, pkg := range packages {
		if pkg.Manifest.Domain == domain {
			return pkg, nil
		}
	}

	return nil, fmt.Errorf("plugin not found: %s", domain)
}
