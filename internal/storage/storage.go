// Package storage provides an abstraction layer for data persistence.
// This enables switching between file-based, SQLite, and remote API storage.
package storage

import (
	"github.com/carlosinfantes/cio/internal/types"
)

// Storage defines the interface for all persistence operations.
// Implementations can be file-based (current), SQLite (future), or API-based (remote).
type Storage interface {
	// Context operations (CRF)
	LoadContext() (*types.CRFContext, error)
	SaveEntity(entity *types.CRFDocument) error
	DeleteEntity(entityType types.CRFEntityType, id string) error
	GetEntity(id string) (*types.CRFEntity, error)
	ListEntities(entityType types.CRFEntityType) ([]types.CRFEntity, error)

	// Decision operations (DRF)
	SaveDecision(doc *types.DRFDocument) error
	GetDecision(id string) (*types.DRFDocument, error)
	ListDecisions(filter *DecisionFilter) ([]types.DRFDocument, error)
	UpdateDecisionStatus(id string, status types.DRFStatus) error
	SearchDecisions(query string) ([]types.DRFDocument, error)

	// Discovery session operations
	SaveDiscoverySession(session *types.DiscoverySession, name string) (string, error)
	LoadDiscoverySession(id string) (*types.DiscoverySession, error)
	ListDiscoverySessions() ([]DiscoverySessionSummary, error)
	DeleteDiscoverySession(id string) error

	// Configuration operations
	LoadConfig() (*types.Config, error)
	SaveConfig(cfg *types.Config) error

	// Health and metadata
	IsInitialized() bool
	Initialize() error
	GetStorageType() StorageType
}

// StorageType identifies the storage backend.
type StorageType string

const (
	StorageTypeFile   StorageType = "file"
	StorageTypeSQLite StorageType = "sqlite"
	StorageTypeAPI    StorageType = "api"
)

// DecisionFilter defines filtering options for listing decisions.
type DecisionFilter struct {
	Status   types.DRFStatus
	Tag      string
	Domain   string
	Since    string // ISO date string
	Limit    int
	Offset   int
}

// DiscoverySessionSummary provides a lightweight view of a discovery session.
type DiscoverySessionSummary struct {
	ID           string
	Name         string
	MessageCount int
	Status       types.DiscoveryStatus
	CreatedAt    string
	UpdatedAt    string
}

// EntityFilter defines filtering options for listing entities.
type EntityFilter struct {
	Type types.CRFEntityType
	Tags []string
}
