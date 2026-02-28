// Package discovery handles discovery session storage and management.
package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/carlosinfantes/cio/internal/config"
	"github.com/carlosinfantes/cio/internal/types"
)

// SaveSession writes a discovery session to disk.
func SaveSession(session *types.DiscoverySession, name string) (string, error) {
	discoveryDir, err := config.GetDiscoveryDir()
	if err != nil {
		return "", err
	}

	if err := config.EnsureDir(discoveryDir); err != nil {
		return "", err
	}

	// Generate ID with optional name suffix
	if session.ID == "" {
		session.ID = generateSessionID(name)
	} else if name != "" {
		// Update ID with name if provided
		session.ID = generateSessionID(name)
	}

	session.UpdatedAt = time.Now()

	path := filepath.Join(discoveryDir, session.ID+".yaml")
	data, err := yaml.Marshal(session)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	return session.ID, nil
}

// LoadSession retrieves a discovery session by ID.
func LoadSession(id string) (*types.DiscoverySession, error) {
	discoveryDir, err := config.GetDiscoveryDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(discoveryDir, id+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var session types.DiscoverySession
	if err := yaml.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// ListSessions returns all saved discovery sessions.
func ListSessions() ([]types.DiscoverySession, error) {
	discoveryDir, err := config.GetDiscoveryDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(discoveryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []types.DiscoverySession{}, nil
		}
		return nil, err
	}

	var sessions []types.DiscoverySession
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".yaml")
		session, err := LoadSession(id)
		if err != nil || session == nil {
			continue
		}

		sessions = append(sessions, *session)
	}

	// Sort by updated date, newest first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// ListActiveSessions returns only active (not converted/abandoned) sessions.
func ListActiveSessions() ([]types.DiscoverySession, error) {
	all, err := ListSessions()
	if err != nil {
		return nil, err
	}

	var active []types.DiscoverySession
	for _, s := range all {
		if s.Status == types.DiscoveryStatusActive {
			active = append(active, s)
		}
	}

	return active, nil
}

// DeleteSession removes a discovery session.
func DeleteSession(id string) error {
	discoveryDir, err := config.GetDiscoveryDir()
	if err != nil {
		return err
	}

	path := filepath.Join(discoveryDir, id+".yaml")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// UpdateSessionStatus changes the status of a discovery session.
func UpdateSessionStatus(id string, status types.DiscoveryStatus) error {
	session, err := LoadSession(id)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("discovery session not found: %s", id)
	}

	session.Status = status
	session.UpdatedAt = time.Now()

	_, err = SaveSession(session, "")
	return err
}

// generateSessionID creates a unique ID for a discovery session.
func generateSessionID(name string) string {
	date := time.Now().Format("2006-01-02")
	timestamp := time.Now().Format("150405")

	if name != "" {
		slug := slugify(name)
		return fmt.Sprintf("disc-%s-%s", date, slug)
	}

	return fmt.Sprintf("disc-%s-%s", date, timestamp)
}

// slugify converts text to a URL-friendly slug.
func slugify(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove special characters, keep only alphanumeric and spaces
	reg := regexp.MustCompile(`[^a-z0-9\s]+`)
	text = reg.ReplaceAllString(text, "")

	// Split into words and take first 4
	words := strings.Fields(text)
	if len(words) > 4 {
		words = words[:4]
	}

	// Join with hyphens
	slug := strings.Join(words, "-")

	// Limit length
	if len(slug) > 40 {
		slug = slug[:40]
	}

	// Remove trailing hyphen
	slug = strings.TrimSuffix(slug, "-")

	return slug
}

// GetSessionSummary returns a brief summary of a session for listing.
func GetSessionSummary(session *types.DiscoverySession) string {
	if len(session.Messages) == 0 {
		return "(no messages)"
	}

	// Find first user message
	for _, msg := range session.Messages {
		if msg.Role == "user" {
			text := msg.Content
			if len(text) > 60 {
				text = text[:57] + "..."
			}
			return text
		}
	}

	return "(no user messages)"
}
