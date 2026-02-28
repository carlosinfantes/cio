// Package context handles loading and managing CRF context entities.
package context

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/carlosinfantes/cio/internal/config"
	"github.com/carlosinfantes/cio/internal/types"
)

// LoadCRFContext loads all CRF entities from the context directory.
func LoadCRFContext() (*types.CRFContext, error) {
	if !config.IsInitialized() {
		return nil, nil
	}

	contextDir, err := config.GetContextDir()
	if err != nil {
		return nil, err
	}

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
		doc, err := LoadCRFDocument(path)
		if err != nil {
			// Skip invalid files
			continue
		}
		if doc == nil {
			continue
		}

		// Categorize by entity type
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

	// Return nil if no entities were loaded
	if len(ctx.AllEntities()) == 0 {
		return nil, nil
	}

	return ctx, nil
}

// LoadCRFDocument loads a single CRF document from a file path.
func LoadCRFDocument(path string) (*types.CRFDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var doc types.CRFDocument
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	// Validate basic structure
	if doc.CRFVersion == "" || doc.Entity.ID == "" || doc.Entity.Name == "" {
		return nil, nil
	}

	return &doc, nil
}

// SaveCRFDocument saves a CRF document to the context directory with optional filename.
func SaveCRFDocument(doc *types.CRFDocument, filename ...string) error {
	contextDir, err := config.GetContextDir()
	if err != nil {
		return err
	}

	if err := config.EnsureDir(contextDir); err != nil {
		return err
	}

	// Use provided filename or generate one
	var fname string
	if len(filename) > 0 && filename[0] != "" {
		fname = filename[0]
	} else {
		fname = generateCRFFilename(doc.Entity.Type, doc.Entity.ID)
	}
	path := filepath.Join(contextDir, fname)

	data, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// generateCRFFilename creates a consistent filename for CRF entities.
func generateCRFFilename(entityType types.CRFEntityType, id string) string {
	prefix := string(entityType)
	// Use first 8 characters of UUID for readability
	shortID := id
	if len(id) > 8 {
		shortID = id[:8]
	}
	return prefix + "-" + shortID + ".yaml"
}

// GetCRFEntity retrieves a CRF entity by ID from loaded context.
func GetCRFEntity(ctx *types.CRFContext, id string) *types.CRFEntity {
	if ctx == nil {
		return nil
	}
	return ctx.GetEntityByID(id)
}

// CreateOrganizationEntity creates a new organization CRF entity.
func CreateOrganizationEntity(id, name, description string, attrs map[string]interface{}) *types.CRFDocument {
	now := time.Now()
	if attrs == nil {
		attrs = make(map[string]interface{})
	}
	return &types.CRFDocument{
		CRFVersion: "0.1.0",
		Entity: types.CRFEntity{
			ID:          id,
			Type:        types.CRFEntityOrganization,
			Name:        name,
			Description: description,
			Attributes:  attrs,
			Provenance: types.Provenance{
				Source:    "manual",
				CreatedAt: now,
				CreatedBy: "cio-init",
			},
			Tags: []string{"organization"},
		},
	}
}

// CreateTeamEntity creates a team organization entity.
func CreateTeamEntity(id, name, description string, headcount int, skills []string) *types.CRFDocument {
	now := time.Now()
	attrs := map[string]interface{}{
		"org_type":  "team",
		"headcount": headcount,
	}
	if len(skills) > 0 {
		attrs["skills"] = skills
	}
	doc := &types.CRFDocument{
		CRFVersion: "0.1.0",
		Entity: types.CRFEntity{
			ID:          id,
			Type:        types.CRFEntityOrganization,
			Name:        name,
			Description: description,
			Attributes:  attrs,
			Provenance: types.Provenance{
				Source:    "manual",
				CreatedAt: now,
				CreatedBy: "cio-init",
			},
			Tags: []string{"team"},
		},
	}

	return doc
}

// CreateSystemEntity creates a system CRF entity.
func CreateSystemEntity(id, name, description string, attrs map[string]interface{}) *types.CRFDocument {
	now := time.Now()
	if attrs == nil {
		attrs = make(map[string]interface{})
	}
	return &types.CRFDocument{
		CRFVersion: "0.1.0",
		Entity: types.CRFEntity{
			ID:          id,
			Type:        types.CRFEntitySystem,
			Name:        name,
			Description: description,
			Attributes:  attrs,
			Provenance: types.Provenance{
				Source:    "manual",
				CreatedAt: now,
				CreatedBy: "cio-init",
			},
			Tags: []string{"system", "infrastructure"},
		},
	}
}

// CreateCapabilityEntity creates a capability CRF entity.
func CreateCapabilityEntity(name, capType, proficiency, importance string, coverage int) *types.CRFDocument {
	now := time.Now()
	return &types.CRFDocument{
		CRFVersion: "0.1.0",
		Entity: types.CRFEntity{
			ID:          uuid.New().String(),
			Type:        types.CRFEntityCapability,
			Name:        name,
			Description: name + " capability",
			Attributes: map[string]interface{}{
				"capability_type":      capType,
				"proficiency":          proficiency,
				"coverage":             coverage,
				"strategic_importance": importance,
			},
			Provenance: types.Provenance{
				Source:    "manual",
				CreatedAt: now,
				CreatedBy: "cio-init",
			},
			Tags: []string{"capability"},
		},
	}
}

// CreateFactEntity creates a fact CRF entity.
func CreateFactEntity(id, name, description string, attrs map[string]interface{}) *types.CRFDocument {
	now := time.Now()
	if attrs == nil {
		attrs = make(map[string]interface{})
	}
	factType, _ := attrs["fact_type"].(string)
	if factType == "" {
		factType = "fact"
	}
	return &types.CRFDocument{
		CRFVersion: "0.1.0",
		Entity: types.CRFEntity{
			ID:          id,
			Type:        types.CRFEntityFact,
			Name:        name,
			Description: description,
			Attributes:  attrs,
			Provenance: types.Provenance{
				Source:    "manual",
				CreatedAt: now,
				CreatedBy: "cio-init",
			},
			Tags: []string{"fact", factType},
		},
	}
}

// CreatePolicyEntity creates a policy CRF entity.
func CreatePolicyEntity(name, policyType, enforcement, scope, rationale string) *types.CRFDocument {
	now := time.Now()
	return &types.CRFDocument{
		CRFVersion: "0.1.0",
		Entity: types.CRFEntity{
			ID:          uuid.New().String(),
			Type:        types.CRFEntityPolicy,
			Name:        name,
			Description: rationale,
			Attributes: map[string]interface{}{
				"policy_type": policyType,
				"enforcement": enforcement,
				"scope":       scope,
				"rationale":   rationale,
			},
			Provenance: types.Provenance{
				Source:    "manual",
				CreatedAt: now,
				CreatedBy: "cio-init",
			},
			Tags: []string{"policy", policyType},
		},
	}
}

// CheckContextStaleness checks if context is stale based on provenance dates.
func CheckContextStaleness(ctx *types.CRFContext, maxDays int) *types.StalenessWarning {
	if ctx == nil {
		return nil
	}

	var oldest time.Time
	var oldestFile string

	for _, entity := range ctx.AllEntities() {
		created := entity.Provenance.CreatedAt
		updated := entity.Provenance.UpdatedAt

		checkTime := created
		if !updated.IsZero() && updated.After(created) {
			checkTime = updated
		}

		if oldest.IsZero() || checkTime.Before(oldest) {
			oldest = checkTime
			oldestFile = entity.Name
		}
	}

	if oldest.IsZero() {
		return nil
	}

	daysSince := int(time.Since(oldest).Hours() / 24)
	if daysSince > maxDays {
		return &types.StalenessWarning{
			DaysSinceUpdate: daysSince,
			OldestFile:      oldestFile,
			LastUpdated:     oldest,
		}
	}

	return nil
}

// DeleteCRFEntity removes a CRF entity file by ID.
func DeleteCRFEntity(entityType types.CRFEntityType, id string) error {
	contextDir, err := config.GetContextDir()
	if err != nil {
		return err
	}

	filename := generateCRFFilename(entityType, id)
	path := filepath.Join(contextDir, filename)

	return os.Remove(path)
}

// ListCRFEntities returns all entities of a specific type.
func ListCRFEntities(ctx *types.CRFContext, entityType types.CRFEntityType) []types.CRFEntity {
	if ctx == nil {
		return nil
	}

	switch entityType {
	case types.CRFEntityOrganization:
		entities := make([]types.CRFEntity, len(ctx.Organizations))
		for i, doc := range ctx.Organizations {
			entities[i] = doc.Entity
		}
		return entities
	case types.CRFEntitySystem:
		entities := make([]types.CRFEntity, len(ctx.Systems))
		for i, doc := range ctx.Systems {
			entities[i] = doc.Entity
		}
		return entities
	case types.CRFEntityCapability:
		entities := make([]types.CRFEntity, len(ctx.Capabilities))
		for i, doc := range ctx.Capabilities {
			entities[i] = doc.Entity
		}
		return entities
	case types.CRFEntityFact:
		entities := make([]types.CRFEntity, len(ctx.Facts))
		for i, doc := range ctx.Facts {
			entities[i] = doc.Entity
		}
		return entities
	case types.CRFEntityPolicy:
		entities := make([]types.CRFEntity, len(ctx.Policies))
		for i, doc := range ctx.Policies {
			entities[i] = doc.Entity
		}
		return entities
	case types.CRFEntityArchitecture:
		entities := make([]types.CRFEntity, len(ctx.Architecture))
		for i, doc := range ctx.Architecture {
			entities[i] = doc.Entity
		}
		return entities
	default:
		return nil
	}
}
