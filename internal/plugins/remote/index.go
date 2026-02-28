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
	Emoji       string   `json:"emoji,omitempty" yaml:"emoji,omitempty"`
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
	Stars     int `json:"stars,omitempty" yaml:"stars,omitempty"`

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

// GetDefaultIndex returns a default index with all official plugins.
// Used as fallback when the registry is not reachable.
func GetDefaultIndex() *RegistryIndex {
	return &RegistryIndex{
		RegistryVersion: "1.0.0",
		LastUpdated:     time.Now(),
		Plugins: []PluginEntry{
			{Domain: "career-advisory", Emoji: "\U0001F9ED", DisplayName: "Career & Growth Advisory Board", Version: "1.0.0", Description: "AI-powered career advisory for professionals", Author: "CIO Team", Category: "personal", Stars: 128, Downloads: 3204},
			{Domain: "cio", Emoji: "\U0001F4AD", DisplayName: "CIO - Chief Intelligence Officer", Version: "1.0.0", Description: "AI-powered executive committee for intelligent decision-making", Author: "CIO Team", Category: "technology", Stars: 247, Downloads: 3782},
			{Domain: "creative-advisory", Emoji: "\u2728", DisplayName: "Creative & Projects Advisory Board", Version: "1.0.0", Description: "AI-powered advisory for creators, writers, and indie builders", Author: "CIO Team", Category: "personal", Stars: 312, Downloads: 7298},
			{Domain: "data-ai-advisory", Emoji: "\U0001F9E0", DisplayName: "Data & AI Advisory Board", Version: "1.0.0", Description: "AI-powered advisory for data strategy and AI governance", Author: "CIO Team", Category: "technology", Stars: 389, Downloads: 7448},
			{Domain: "financial-advisory", Emoji: "\U0001F4CA", DisplayName: "Financial Advisory Board", Version: "1.0.0", Description: "AI-powered CFO-level counsel for financial strategy", Author: "CIO Team", Category: "business", Featured: true, Stars: 456, Downloads: 9146},
			{Domain: "legal-advisory", Emoji: "\u2696\uFE0F", DisplayName: "Legal Advisory Board", Version: "1.0.0", Description: "AI-powered legal counsel for business decisions", Author: "CIO Team", Category: "legal", Stars: 67, Downloads: 953},
			{Domain: "marketing-advisory", Emoji: "\U0001F4E3", DisplayName: "Marketing & Brand Advisory Board", Version: "1.0.0", Description: "AI-powered CMO-level counsel for marketing strategy", Author: "CIO Team", Category: "business", Featured: true, Stars: 534, Downloads: 13959},
			{Domain: "people-advisory", Emoji: "\U0001F917", DisplayName: "People & Culture Advisory Board", Version: "1.0.0", Description: "AI-powered CHRO-level counsel for people strategy", Author: "CIO Team", Category: "business", Stars: 142, Downloads: 3492},
			{Domain: "personal-finance", Emoji: "\U0001F9ED", DisplayName: "Personal Finance Advisory Board", Version: "1.0.0", Description: "AI-powered personal finance advisory for individuals", Author: "CIO Team", Category: "personal", Stars: 89, Downloads: 1175},
			{Domain: "product-advisory", Emoji: "\U0001F4A1", DisplayName: "Product & Growth Advisory Board", Version: "1.0.0", Description: "AI-powered CPO-level counsel for product strategy", Author: "CIO Team", Category: "business", Stars: 198, Downloads: 2982},
			{Domain: "security-advisory", Emoji: "\U0001F50D", DisplayName: "Security Advisory Board", Version: "1.0.0", Description: "AI-powered CISO-level counsel for cybersecurity strategy", Author: "CIO Team", Category: "technology", Stars: 276, Downloads: 4706},
			{Domain: "startup-advisory", Emoji: "\U0001F680", DisplayName: "Startup Advisory Board", Version: "1.0.0", Description: "AI-powered advisory for founders and early-stage companies", Author: "CIO Team", Category: "business", Featured: true, Stars: 421, Downloads: 8158},
			{Domain: "wellness-advisory", Emoji: "\U0001F331", DisplayName: "Health & Wellness Advisory Board", Version: "1.0.0", Description: "AI-powered wellness advisory for health and lifestyle", Author: "CIO Team", Category: "personal", Stars: 73, Downloads: 1110},
		},
		Categories: []Category{
			{ID: "technology", Name: "Technology & Engineering", Description: "Technical decision-making and engineering advisory"},
			{ID: "legal", Name: "Legal & Compliance", Description: "Legal counsel, contracts, and compliance advisory"},
			{ID: "healthcare", Name: "Healthcare", Description: "Medical and healthcare advisory"},
			{ID: "business", Name: "Business & Operations", Description: "General business operations and specialty domains"},
			{ID: "personal", Name: "Personal & Life", Description: "Personal advisory for life decisions, wellness, finances, and growth"},
		},
	}
}
