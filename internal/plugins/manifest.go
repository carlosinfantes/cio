// Package plugins implements the domain plugin system for the advisory board.
// Plugins allow extending the advisory board to different domains like legal,
// architecture (buildings), healthcare, etc.
package plugins

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/carlosinfantes/cio/internal/types"
)

// Manifest defines the structure of a domain plugin.
type Manifest struct {
	// Domain identifier (e.g., "cio", "legal-advisory")
	Domain string `yaml:"domain" json:"domain"`

	// Version following semver
	Version string `yaml:"version" json:"version"`

	// Human-readable name
	DisplayName string `yaml:"display_name" json:"display_name"`

	// Description of the domain
	Description string `yaml:"description" json:"description"`

	// Facilitator configuration
	Facilitator FacilitatorConfig `yaml:"facilitator" json:"facilitator"`

	// Core advisors (always available)
	CoreAdvisors []AdvisorConfig `yaml:"core_advisors" json:"core_advisors"`

	// Specialist advisors (auto-summoned by keywords)
	Specialists []SpecialistConfig `yaml:"specialists,omitempty" json:"specialists,omitempty"`

	// Specialist trigger keywords
	SpecialistTriggers map[string]TriggerConfig `yaml:"specialist_triggers,omitempty" json:"specialist_triggers,omitempty"`

	// Context entity types for this domain
	ContextEntities []EntityTypeConfig `yaml:"context_entities,omitempty" json:"context_entities,omitempty"`

	// Decision domains (categories for DRF)
	DecisionDomains []string `yaml:"decision_domains,omitempty" json:"decision_domains,omitempty"`

	// Custom prompts directory (relative to plugin root)
	PromptsDir string `yaml:"prompts_dir,omitempty" json:"prompts_dir,omitempty"`

	// Plugin metadata
	Author        string `yaml:"author,omitempty" json:"author,omitempty"`
	Homepage      string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	License       string `yaml:"license,omitempty" json:"license,omitempty"`
	MinCLIVersion string `yaml:"min_cli_version,omitempty" json:"min_cli_version,omitempty"`

	// Plugin settings (domain-specific defaults)
	Settings PluginSettings `yaml:"settings,omitempty" json:"settings,omitempty"`
}

// PluginSettings contains domain-specific default settings.
type PluginSettings struct {
	// DefaultMode for this domain
	DefaultMode string `yaml:"default_mode,omitempty" json:"default_mode,omitempty"`

	// DefaultAdvisors to include in panel
	DefaultAdvisors []string `yaml:"default_advisors,omitempty" json:"default_advisors,omitempty"`

	// StartInDiscovery whether to start with facilitator
	StartInDiscovery bool `yaml:"start_in_discovery,omitempty" json:"start_in_discovery,omitempty"`

	// MaxAdvisors in a session
	MaxAdvisors int `yaml:"max_advisors,omitempty" json:"max_advisors,omitempty"`

	// Custom settings map for domain-specific config
	Custom map[string]interface{} `yaml:"custom,omitempty" json:"custom,omitempty"`
}

// FacilitatorConfig defines the domain's facilitator (Jordan equivalent).
type FacilitatorConfig struct {
	ID            string `yaml:"id" json:"id"`
	Name          string `yaml:"name" json:"name"`
	Role          string `yaml:"role" json:"role"`
	Emoji         string `yaml:"emoji,omitempty" json:"emoji,omitempty"`
	Color         string `yaml:"color,omitempty" json:"color,omitempty"`
	ThinkingStyle string `yaml:"thinking_style,omitempty" json:"thinking_style,omitempty"`
	GreetingPrompt string `yaml:"greeting_prompt,omitempty" json:"greeting_prompt,omitempty"`
}

// AdvisorConfig defines a core advisor for the domain.
type AdvisorConfig struct {
	ID             string   `yaml:"id" json:"id"`
	Name           string   `yaml:"name" json:"name"`
	Role           string   `yaml:"role" json:"role"`
	Emoji          string   `yaml:"emoji,omitempty" json:"emoji,omitempty"`
	Color          string   `yaml:"color,omitempty" json:"color,omitempty"`
	ThinkingStyle  string   `yaml:"thinking_style" json:"thinking_style"`
	Background     string   `yaml:"background,omitempty" json:"background,omitempty"`
	Priorities     []string `yaml:"priorities,omitempty" json:"priorities,omitempty"`
	CatchPhrases   []string `yaml:"catch_phrases,omitempty" json:"catch_phrases,omitempty"`
}

// SpecialistConfig defines a specialist advisor.
type SpecialistConfig struct {
	AdvisorConfig `yaml:",inline"`
	Keywords      []string `yaml:"keywords,omitempty" json:"keywords,omitempty"`
}

// TriggerConfig defines auto-summon triggers for specialists.
type TriggerConfig struct {
	Keywords []string `yaml:"keywords" json:"keywords"`
}

// EntityTypeConfig defines a context entity type for the domain.
type EntityTypeConfig struct {
	Type        string            `yaml:"type" json:"type"`
	DisplayName string            `yaml:"display_name" json:"display_name"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Attributes  []AttributeConfig `yaml:"attributes,omitempty" json:"attributes,omitempty"`
}

// AttributeConfig defines an attribute for a context entity.
type AttributeConfig struct {
	Name        string   `yaml:"name" json:"name"`
	Type        string   `yaml:"type" json:"type"` // string, number, boolean, array, enum
	Required    bool     `yaml:"required,omitempty" json:"required,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	EnumValues  []string `yaml:"enum_values,omitempty" json:"enum_values,omitempty"`
}

// LoadManifest loads a plugin manifest from a file.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("validating manifest: %w", err)
	}

	return &manifest, nil
}

// Validate checks the manifest for required fields and consistency.
func (m *Manifest) Validate() error {
	if m.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if m.Version == "" {
		return fmt.Errorf("version is required")
	}
	if m.DisplayName == "" {
		return fmt.Errorf("display_name is required")
	}
	if m.Facilitator.Name == "" {
		return fmt.Errorf("facilitator.name is required")
	}
	if len(m.CoreAdvisors) == 0 {
		return fmt.Errorf("at least one core_advisor is required")
	}

	// Validate advisors
	for i, advisor := range m.CoreAdvisors {
		if advisor.ID == "" {
			return fmt.Errorf("core_advisors[%d].id is required", i)
		}
		if advisor.Name == "" {
			return fmt.Errorf("core_advisors[%d].name is required", i)
		}
		if advisor.Role == "" {
			return fmt.Errorf("core_advisors[%d].role is required", i)
		}
	}

	return nil
}

// ToPersonas converts the plugin's advisors to Persona types.
func (m *Manifest) ToPersonas() []types.Persona {
	personas := make([]types.Persona, 0, len(m.CoreAdvisors)+len(m.Specialists))

	for _, advisor := range m.CoreAdvisors {
		personas = append(personas, types.Persona{
			ID:            types.AdvisorID(advisor.ID),
			Name:          advisor.Name,
			Role:          advisor.Role,
			Color:         advisor.Color,
			Emoji:         advisor.Emoji,
			ThinkingStyle: advisor.ThinkingStyle,
			Background:    advisor.Background,
			Priorities:    advisor.Priorities,
			CatchPhrases:  advisor.CatchPhrases,
			IsSpecialist:  false,
		})
	}

	for _, specialist := range m.Specialists {
		personas = append(personas, types.Persona{
			ID:                 types.AdvisorID(specialist.ID),
			Name:               specialist.Name,
			Role:               specialist.Role,
			Color:              specialist.Color,
			Emoji:              specialist.Emoji,
			ThinkingStyle:      specialist.ThinkingStyle,
			Background:         specialist.Background,
			Priorities:         specialist.Priorities,
			CatchPhrases:       specialist.CatchPhrases,
			AutoSummonKeywords: specialist.Keywords,
			IsSpecialist:       true,
		})
	}

	return personas
}

// GetFacilitatorPersona returns the facilitator as a Persona.
func (m *Manifest) GetFacilitatorPersona() types.Persona {
	return types.Persona{
		ID:            types.AdvisorID(m.Facilitator.ID),
		Name:          m.Facilitator.Name,
		Role:          m.Facilitator.Role,
		Color:         m.Facilitator.Color,
		Emoji:         m.Facilitator.Emoji,
		ThinkingStyle: m.Facilitator.ThinkingStyle,
	}
}

// SaveManifest writes a manifest to a file.
func SaveManifest(manifest *Manifest, path string) error {
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	return nil
}
