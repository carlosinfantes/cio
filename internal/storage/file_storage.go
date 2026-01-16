// Package storage provides file-based storage implementation.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/carlosinfantes/cto-advisory-board/internal/config"
	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

// FileStorage implements Storage interface using file-based persistence.
type FileStorage struct {
	baseDir string
}

// NewFileStorage creates a new file-based storage instance.
func NewFileStorage() (*FileStorage, error) {
	baseDir, err := config.GetAdvisoryDir()
	if err != nil {
		return nil, err
	}
	return &FileStorage{baseDir: baseDir}, nil
}

// NewFileStorageWithDir creates a file storage with a specific base directory.
func NewFileStorageWithDir(baseDir string) *FileStorage {
	return &FileStorage{baseDir: baseDir}
}

// GetStorageType returns the storage type identifier.
func (fs *FileStorage) GetStorageType() StorageType {
	return StorageTypeFile
}

// IsInitialized checks if the storage has been initialized.
func (fs *FileStorage) IsInitialized() bool {
	configPath := filepath.Join(fs.baseDir, "config.yaml")
	_, err := os.Stat(configPath)
	return err == nil
}

// Initialize creates the necessary directory structure.
func (fs *FileStorage) Initialize() error {
	dirs := []string{
		fs.baseDir,
		filepath.Join(fs.baseDir, "context"),
		filepath.Join(fs.baseDir, "decisions"),
		filepath.Join(fs.baseDir, "discovery"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// ============================================================================
// Context (CRF) Operations
// ============================================================================

// LoadContext loads all CRF entities from the context directory.
func (fs *FileStorage) LoadContext() (*types.CRFContext, error) {
	contextDir := filepath.Join(fs.baseDir, "context")

	entries, err := os.ReadDir(contextDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	ctx := &types.CRFContext{
		Organizations: []types.CRFDocument{},
		Systems:       []types.CRFDocument{},
		Capabilities:  []types.CRFDocument{},
		Facts:         []types.CRFDocument{},
		Policies:      []types.CRFDocument{},
		Architecture:  []types.CRFDocument{},
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join(contextDir, entry.Name())
		doc, err := fs.loadCRFDocument(path)
		if err != nil || doc == nil {
			continue
		}

		switch doc.Entity.Type {
		case types.CRFEntityOrganization:
			ctx.Organizations = append(ctx.Organizations, *doc)
		case types.CRFEntitySystem:
			ctx.Systems = append(ctx.Systems, *doc)
		case types.CRFEntityCapability:
			ctx.Capabilities = append(ctx.Capabilities, *doc)
		case types.CRFEntityFact:
			ctx.Facts = append(ctx.Facts, *doc)
		case types.CRFEntityPolicy:
			ctx.Policies = append(ctx.Policies, *doc)
		case types.CRFEntityArchitecture:
			ctx.Architecture = append(ctx.Architecture, *doc)
		}
	}

	if len(ctx.AllEntities()) == 0 {
		return nil, nil
	}

	return ctx, nil
}

func (fs *FileStorage) loadCRFDocument(path string) (*types.CRFDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var doc types.CRFDocument
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	if doc.CRFVersion == "" || doc.Entity.ID == "" || doc.Entity.Name == "" {
		return nil, nil
	}

	return &doc, nil
}

// SaveEntity saves a CRF entity to the context directory.
func (fs *FileStorage) SaveEntity(doc *types.CRFDocument) error {
	contextDir := filepath.Join(fs.baseDir, "context")
	if err := os.MkdirAll(contextDir, 0755); err != nil {
		return err
	}

	filename := fs.generateCRFFilename(doc.Entity.Type, doc.Entity.ID)
	path := filepath.Join(contextDir, filename)

	data, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (fs *FileStorage) generateCRFFilename(entityType types.CRFEntityType, id string) string {
	prefix := string(entityType)
	shortID := id
	if len(id) > 8 {
		shortID = id[:8]
	}
	return prefix + "-" + shortID + ".yaml"
}

// DeleteEntity removes a CRF entity file by ID.
func (fs *FileStorage) DeleteEntity(entityType types.CRFEntityType, id string) error {
	contextDir := filepath.Join(fs.baseDir, "context")
	filename := fs.generateCRFFilename(entityType, id)
	path := filepath.Join(contextDir, filename)
	return os.Remove(path)
}

// GetEntity retrieves a CRF entity by ID.
func (fs *FileStorage) GetEntity(id string) (*types.CRFEntity, error) {
	ctx, err := fs.LoadContext()
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		return nil, nil
	}
	return ctx.GetEntityByID(id), nil
}

// ListEntities returns all entities of a specific type.
func (fs *FileStorage) ListEntities(entityType types.CRFEntityType) ([]types.CRFEntity, error) {
	ctx, err := fs.LoadContext()
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		return nil, nil
	}

	switch entityType {
	case types.CRFEntityOrganization:
		entities := make([]types.CRFEntity, len(ctx.Organizations))
		for i, doc := range ctx.Organizations {
			entities[i] = doc.Entity
		}
		return entities, nil
	case types.CRFEntitySystem:
		entities := make([]types.CRFEntity, len(ctx.Systems))
		for i, doc := range ctx.Systems {
			entities[i] = doc.Entity
		}
		return entities, nil
	case types.CRFEntityCapability:
		entities := make([]types.CRFEntity, len(ctx.Capabilities))
		for i, doc := range ctx.Capabilities {
			entities[i] = doc.Entity
		}
		return entities, nil
	case types.CRFEntityFact:
		entities := make([]types.CRFEntity, len(ctx.Facts))
		for i, doc := range ctx.Facts {
			entities[i] = doc.Entity
		}
		return entities, nil
	case types.CRFEntityPolicy:
		entities := make([]types.CRFEntity, len(ctx.Policies))
		for i, doc := range ctx.Policies {
			entities[i] = doc.Entity
		}
		return entities, nil
	case types.CRFEntityArchitecture:
		entities := make([]types.CRFEntity, len(ctx.Architecture))
		for i, doc := range ctx.Architecture {
			entities[i] = doc.Entity
		}
		return entities, nil
	default:
		return nil, nil
	}
}

// ============================================================================
// Decision (DRF) Operations
// ============================================================================

// SaveDecision writes a DRF document to disk.
func (fs *FileStorage) SaveDecision(doc *types.DRFDocument) error {
	decisionsDir := filepath.Join(fs.baseDir, "decisions")
	if err := os.MkdirAll(decisionsDir, 0755); err != nil {
		return err
	}

	path := filepath.Join(decisionsDir, doc.Decision.ID+".yaml")
	data, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetDecision retrieves a DRF document by ID.
func (fs *FileStorage) GetDecision(id string) (*types.DRFDocument, error) {
	decisionsDir := filepath.Join(fs.baseDir, "decisions")
	path := filepath.Join(decisionsDir, id+".yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var doc types.DRFDocument
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

// ListDecisions returns all DRF documents, optionally filtered.
func (fs *FileStorage) ListDecisions(filter *DecisionFilter) ([]types.DRFDocument, error) {
	decisionsDir := filepath.Join(fs.baseDir, "decisions")

	entries, err := os.ReadDir(decisionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []types.DRFDocument{}, nil
		}
		return nil, err
	}

	var docs []types.DRFDocument
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".yaml")
		doc, err := fs.GetDecision(id)
		if err != nil || doc == nil {
			continue
		}

		// Apply filters
		if filter != nil {
			if filter.Status != "" && doc.Meta.Status != filter.Status {
				continue
			}
			if filter.Tag != "" && !containsTag(doc.Meta.Tags, filter.Tag) {
				continue
			}
			if filter.Domain != "" && doc.Decision.Domain != filter.Domain {
				continue
			}
		}

		docs = append(docs, *doc)
	}

	// Sort by created date, newest first
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Meta.CreatedAt.After(docs[j].Meta.CreatedAt)
	})

	// Apply limit and offset
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(docs) {
			docs = docs[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(docs) {
			docs = docs[:filter.Limit]
		}
	}

	return docs, nil
}

// UpdateDecisionStatus changes the status of a DRF document.
func (fs *FileStorage) UpdateDecisionStatus(id string, status types.DRFStatus) error {
	doc, err := fs.GetDecision(id)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("decision not found: %s", id)
	}

	doc.Meta.Status = status
	doc.Meta.UpdatedAt = time.Now()

	return fs.SaveDecision(doc)
}

// SearchDecisions searches decisions by title or intent.
func (fs *FileStorage) SearchDecisions(query string) ([]types.DRFDocument, error) {
	docs, err := fs.ListDecisions(nil)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []types.DRFDocument
	for _, doc := range docs {
		if strings.Contains(strings.ToLower(doc.Decision.Title), query) ||
			strings.Contains(strings.ToLower(doc.Decision.Intent), query) {
			results = append(results, doc)
		}
	}

	return results, nil
}

func containsTag(tags []string, tag string) bool {
	tag = strings.ToLower(tag)
	for _, t := range tags {
		if strings.ToLower(t) == tag {
			return true
		}
	}
	return false
}

// ============================================================================
// Discovery Session Operations
// ============================================================================

// SaveDiscoverySession saves a discovery session to disk.
func (fs *FileStorage) SaveDiscoverySession(session *types.DiscoverySession, name string) (string, error) {
	discoveryDir := filepath.Join(fs.baseDir, "discovery")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		return "", err
	}

	// Generate ID if needed
	id := session.ID
	if id == "" {
		if name != "" {
			id = name + "-" + time.Now().Format("20060102-150405")
		} else {
			id = uuid.New().String()[:8]
		}
		session.ID = id
	}

	path := filepath.Join(discoveryDir, id+".yaml")
	data, err := yaml.Marshal(session)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	return id, nil
}

// LoadDiscoverySession loads a discovery session by ID.
func (fs *FileStorage) LoadDiscoverySession(id string) (*types.DiscoverySession, error) {
	discoveryDir := filepath.Join(fs.baseDir, "discovery")
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

// ListDiscoverySessions returns summaries of all saved discovery sessions.
func (fs *FileStorage) ListDiscoverySessions() ([]DiscoverySessionSummary, error) {
	discoveryDir := filepath.Join(fs.baseDir, "discovery")

	entries, err := os.ReadDir(discoveryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []DiscoverySessionSummary{}, nil
		}
		return nil, err
	}

	var summaries []DiscoverySessionSummary
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".yaml")
		session, err := fs.LoadDiscoverySession(id)
		if err != nil || session == nil {
			continue
		}

		summaries = append(summaries, DiscoverySessionSummary{
			ID:           session.ID,
			MessageCount: len(session.Messages),
			Status:       session.Status,
			CreatedAt:    session.CreatedAt.Format(time.RFC3339),
		})
	}

	return summaries, nil
}

// DeleteDiscoverySession removes a discovery session file.
func (fs *FileStorage) DeleteDiscoverySession(id string) error {
	discoveryDir := filepath.Join(fs.baseDir, "discovery")
	path := filepath.Join(discoveryDir, id+".yaml")
	return os.Remove(path)
}

// ============================================================================
// Configuration Operations
// ============================================================================

// LoadConfig loads the configuration from disk.
func (fs *FileStorage) LoadConfig() (*types.Config, error) {
	path := filepath.Join(fs.baseDir, "config.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cfg types.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveConfig saves the configuration to disk.
func (fs *FileStorage) SaveConfig(cfg *types.Config) error {
	if err := os.MkdirAll(fs.baseDir, 0755); err != nil {
		return err
	}

	path := filepath.Join(fs.baseDir, "config.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ============================================================================
// Singleton and Factory
// ============================================================================

var defaultStorage Storage

// GetDefaultStorage returns the default storage instance.
func GetDefaultStorage() (Storage, error) {
	if defaultStorage == nil {
		fs, err := NewFileStorage()
		if err != nil {
			return nil, err
		}
		defaultStorage = fs
	}
	return defaultStorage, nil
}

// SetDefaultStorage sets the default storage instance (useful for testing).
func SetDefaultStorage(s Storage) {
	defaultStorage = s
}
