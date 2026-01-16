// Package remote defines the registry index types.
package remote

import (
	"strings"
	"time"
)

// RegistryIndex represents the plugin registry index file.
type RegistryIndex struct {
	RegistryVersion string        `json:"registry_version" yaml:"registry_version"`
	LastUpdated     time.Time     `json:"last_updated" yaml:"last_updated"`
	Plugins         []PluginEntry `json:"plugins" yaml:"plugins"`
	Categories      []Category    `json:"categories" yaml:"categories"`
}

// PluginEntry represents a plugin in the registry index.
type PluginEntry struct {
	// Identifiers
	Domain      string `json:"domain" yaml:"domain"`
	DisplayName string `json:"display_name" yaml:"display_name"`
	Version     string `json:"version" yaml:"version"`

	// Metadata
	Description string   `json:"description" yaml:"description"`
	Author      string   `json:"author" yaml:"author"`
	License     string   `json:"license,omitempty" yaml:"license,omitempty"`
	Homepage    string   `json:"homepage,omitempty" yaml:"homepage,omitempty"`
	Repository  string   `json:"repository,omitempty" yaml:"repository,omitempty"`
	Keywords    []string `json:"keywords,omitempty" yaml:"keywords,omitempty"`

	// Download info
	DownloadURL string `json:"download_url" yaml:"download_url"`
	Checksum    string `json:"checksum,omitempty" yaml:"checksum,omitempty"`
	Size        int64  `json:"size,omitempty" yaml:"size,omitempty"`

	// Requirements
	MinCLIVersion string `json:"min_cli_version,omitempty" yaml:"min_cli_version,omitempty"`

	// Classification
	Category string `json:"category" yaml:"category"`
	Featured bool   `json:"featured,omitempty" yaml:"featured,omitempty"`

	// Stats
	Downloads int `json:"downloads,omitempty" yaml:"downloads,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

// Category represents a plugin category.
type Category struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// Matches checks if a plugin matches a search query.
func (p *PluginEntry) Matches(query string) bool {
	query = strings.ToLower(query)

	// Check domain
	if strings.Contains(strings.ToLower(p.Domain), query) {
		return true
	}

	// Check display name
	if strings.Contains(strings.ToLower(p.DisplayName), query) {
		return true
	}

	// Check description
	if strings.Contains(strings.ToLower(p.Description), query) {
		return true
	}

	// Check keywords
	for _, keyword := range p.Keywords {
		if strings.Contains(strings.ToLower(keyword), query) {
			return true
		}
	}

	// Check category
	if strings.Contains(strings.ToLower(p.Category), query) {
		return true
	}

	return false
}

// GetDefaultIndex returns a default index with placeholder plugins.
// Used when the registry is not available.
func GetDefaultIndex() *RegistryIndex {
	return &RegistryIndex{
		RegistryVersion: "1.0.0",
		LastUpdated:     time.Now(),
		Plugins: []PluginEntry{
			{
				Domain:        "cto-advisory",
				DisplayName:   "CTO Advisory Board",
				Version:       "1.0.0",
				Description:   "AI-powered executive committee for CTOs making technical decisions",
				Author:        "CTO Advisory Board Team",
				License:       "MIT",
				Keywords:      []string{"technology", "architecture", "engineering", "startup"},
				Category:      "technology",
				Featured:      true,
				MinCLIVersion: "1.0.0",
			},
			{
				Domain:        "legal-advisory",
				DisplayName:   "Legal Advisory Board",
				Version:       "1.0.0",
				Description:   "AI-powered legal counsel for business decisions",
				Author:        "CTO Advisory Board Team",
				License:       "MIT",
				Keywords:      []string{"legal", "compliance", "contracts", "corporate"},
				Category:      "legal",
				Featured:      true,
				MinCLIVersion: "1.0.0",
			},
			{
				Domain:        "medical-advisory",
				DisplayName:   "Medical Practice Advisory",
				Version:       "1.0.0",
				Description:   "AI-powered advisory for healthcare practice decisions",
				Author:        "CTO Advisory Board Team",
				License:       "MIT",
				Keywords:      []string{"healthcare", "medical", "clinic", "practice"},
				Category:      "healthcare",
				MinCLIVersion: "1.0.0",
			},
		},
		Categories: []Category{
			{ID: "technology", Name: "Technology & Engineering", Description: "Technical decision-making"},
			{ID: "legal", Name: "Legal & Compliance", Description: "Legal counsel and compliance"},
			{ID: "healthcare", Name: "Healthcare", Description: "Medical and healthcare advisory"},
			{ID: "business", Name: "Business & Operations", Description: "General business operations"},
		},
	}
}
