// Package remote implements the plugin registry client.
package remote

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultRegistryURL is the default plugin registry URL.
const DefaultRegistryURL = "https://raw.githubusercontent.com/carlosinfantes/cio-plugin-registry/main"

// Client is the registry HTTP client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	cache      *RegistryIndex
	cacheTime  time.Time
	cacheTTL   time.Duration
}

// NewClient creates a new registry client.
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultRegistryURL
	}
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cacheTTL: 5 * time.Minute,
	}
}

// FetchIndex downloads and parses the registry index.
func (c *Client) FetchIndex() (*RegistryIndex, error) {
	// Check cache
	if c.cache != nil && time.Since(c.cacheTime) < c.cacheTTL {
		return c.cache, nil
	}

	url := c.BaseURL + "/index.json"
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var index RegistryIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("parsing index: %w", err)
	}

	// Update cache
	c.cache = &index
	c.cacheTime = time.Now()

	return &index, nil
}

// GetPlugin fetches information about a specific plugin.
func (c *Client) GetPlugin(domain string) (*PluginEntry, error) {
	index, err := c.FetchIndex()
	if err != nil {
		return nil, err
	}

	for _, plugin := range index.Plugins {
		if plugin.Domain == domain {
			return &plugin, nil
		}
	}

	return nil, fmt.Errorf("plugin not found: %s", domain)
}

// Search searches plugins by query string.
func (c *Client) Search(query string) ([]PluginEntry, error) {
	index, err := c.FetchIndex()
	if err != nil {
		return nil, err
	}

	if query == "" {
		return index.Plugins, nil
	}

	var results []PluginEntry
	for _, plugin := range index.Plugins {
		if plugin.Matches(query) {
			results = append(results, plugin)
		}
	}

	return results, nil
}

// ListCategories returns all available categories.
func (c *Client) ListCategories() ([]Category, error) {
	index, err := c.FetchIndex()
	if err != nil {
		return nil, err
	}

	return index.Categories, nil
}

// GetPluginsByCategory returns plugins in a specific category.
func (c *Client) GetPluginsByCategory(categoryID string) ([]PluginEntry, error) {
	index, err := c.FetchIndex()
	if err != nil {
		return nil, err
	}

	var results []PluginEntry
	for _, plugin := range index.Plugins {
		if plugin.Category == categoryID {
			results = append(results, plugin)
		}
	}

	return results, nil
}
