// Package decisions handles decision storage and management using DRF format.
package decisions

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

// ListFilters defines filtering options for listing decisions.
type ListFilters struct {
	Status types.DRFStatus
	Tag    string
}

// GenerateUUID creates a new UUID v4 for decision identification.
func GenerateUUID() string {
	return uuid.New().String()
}

// CreateDRFDocument creates a new DRF document from advisory board output.
func CreateDRFDocument(
	question string,
	mode types.Mode,
	advisors []types.AdvisorID,
	parsedResponse types.ParsedResponse,
	crfContext *types.CRFContext,
) *types.DRFDocument {
	now := time.Now()

	// Map advisor responses to DRF interventions
	interventions := make([]types.Intervention, 0, len(parsedResponse.Advisors))
	for i, resp := range parsedResponse.Advisors {
		interventions = append(interventions, types.Intervention{
			ID:        fmt.Sprintf("int-%03d", i+1),
			Type:      types.InterventionInsight,
			Content:   resp.Response,
			Source:    fmt.Sprintf("%s (%s)", resp.Name, resp.Role),
			Timestamp: now,
		})
	}

	// Build context validation if CRF context is available
	var contextValidation *types.ContextValidation
	if crfContext != nil {
		contextRefs := buildContextRefs(crfContext)
		if len(contextRefs) > 0 {
			contextValidation = &types.ContextValidation{
				ValidatedAt: now,
				ContextRefs: contextRefs,
			}
		}
	}

	// Build environment from CRF context
	var environment *types.Environment
	if crfContext != nil {
		environment = buildEnvironmentFromCRF(crfContext)
	}

	// Infer domain from question
	domain := inferDomain(question)

	return &types.DRFDocument{
		DRFVersion: "0.1.0",
		Decision: types.DRFDecision{
			ID:     GenerateUUID(),
			Title:  truncateTitle(question, 200),
			Domain: domain,
			Intent: question,
		},
		Context: types.DRFContext{
			Constraints: extractConstraintsFromCRF(crfContext),
			Objectives:  []types.Objective{{Description: "Make informed decision on: " + question, Priority: "must_have"}},
			Environment: environment,
		},
		CognitiveState: types.CognitiveState{
			Phase:      mapModeToPhase(mode),
			Confidence: 50, // Default confidence, can be adjusted
			PhaseNotes: fmt.Sprintf("Advisory board discussion using %s mode", mode),
		},
		Reasoning: &types.Reasoning{
			PatternsApplied: mapModeToPatterns(mode),
			Notes:           fmt.Sprintf("Multi-perspective analysis by %d advisors", len(advisors)),
		},
		Interventions: interventions,
		Synthesis: types.DRFSynthesis{
			Decision:  parsedResponse.Synthesis,
			Rationale: "Advisory board consensus from expert perspectives",
		},
		ContextValidation: contextValidation,
		Meta: types.DRFMeta{
			CreatedAt: now,
			Status:    types.DRFStatusDraft,
			Source:    "cto-advisory-board",
			Tags:      []string{string(mode)},
			Actors:    buildActors(advisors),
		},
	}
}

// mapModeToPhase maps advisory board mode to DRF cognitive phase.
func mapModeToPhase(mode types.Mode) types.CognitivePhase {
	switch mode {
	case types.ModeFramework:
		return types.PhaseAnalysis
	case types.ModeSocratic:
		return types.PhaseExploration
	case types.ModeAdvocate:
		return types.PhaseAnalysis
	default: // panel
		return types.PhaseSynthesis
	}
}

// mapModeToPatterns maps advisory board mode to DRF reasoning patterns.
func mapModeToPatterns(mode types.Mode) []string {
	switch mode {
	case types.ModePanel:
		return []string{"consensus", "deliberative"}
	case types.ModeSocratic:
		return []string{"deliberative", "systematic"}
	case types.ModeAdvocate:
		return []string{"contrafactual", "risk_based"}
	case types.ModeFramework:
		return []string{"comparative", "cost_benefit", "systematic"}
	default:
		return []string{"consensus"}
	}
}

// inferDomain attempts to categorize the decision based on question content.
func inferDomain(question string) string {
	q := strings.ToLower(question)

	domainKeywords := map[string][]string{
		"architecture":   {"architecture", "design", "pattern", "microservice", "monolith", "api"},
		"security":       {"security", "auth", "encryption", "compliance", "vulnerability", "soc2", "gdpr"},
		"infrastructure": {"kubernetes", "k8s", "aws", "gcp", "azure", "docker", "deploy", "ci/cd", "infrastructure"},
		"data":           {"database", "postgres", "mysql", "redis", "cache", "data", "storage"},
		"team":           {"hire", "team", "engineer", "headcount", "org"},
		"vendor":         {"vendor", "buy", "tool", "license", "contract", "saas"},
		"product":        {"feature", "user", "customer", "product", "roadmap", "mvp"},
		"financial":      {"budget", "cost", "roi", "expense", "pricing"},
	}

	for domain, keywords := range domainKeywords {
		for _, keyword := range keywords {
			if strings.Contains(q, keyword) {
				return domain
			}
		}
	}

	return "general"
}

// truncateTitle ensures the title doesn't exceed max length.
func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen-3] + "..."
}

// buildActors creates actor entries from advisor IDs.
func buildActors(advisors []types.AdvisorID) []types.Actor {
	actors := make([]types.Actor, 0, len(advisors))
	for _, id := range advisors {
		actors = append(actors, types.Actor{
			Name: string(id),
			Role: "contributor",
		})
	}
	return actors
}

// buildContextRefs creates context references from CRF entities.
func buildContextRefs(ctx *types.CRFContext) []types.ContextRef {
	if ctx == nil {
		return nil
	}

	var refs []types.ContextRef

	// Add organization reference
	if org := ctx.GetOrganization(); org != nil {
		refs = append(refs, types.ContextRef{
			ContextID:        org.ID,
			ContextType:      string(org.Type),
			ContextName:      org.Name,
			ValidationStatus: "satisfied",
		})
	}

	// Add policy references
	for _, doc := range ctx.Policies {
		refs = append(refs, types.ContextRef{
			ContextID:        doc.Entity.ID,
			ContextType:      string(doc.Entity.Type),
			ContextName:      doc.Entity.Name,
			ValidationStatus: "acknowledged",
			AdvisoryNotes:    "Policy reviewed during advisory discussion",
		})
	}

	// Add fact/constraint references
	for _, doc := range ctx.Facts {
		refs = append(refs, types.ContextRef{
			ContextID:        doc.Entity.ID,
			ContextType:      string(doc.Entity.Type),
			ContextName:      doc.Entity.Name,
			ValidationStatus: "satisfied",
		})
	}

	return refs
}

// extractConstraintsFromCRF extracts constraints from CRF facts and policies.
func extractConstraintsFromCRF(ctx *types.CRFContext) []types.Constraint {
	if ctx == nil {
		return []types.Constraint{}
	}

	var constraints []types.Constraint

	// Extract from facts (constraints type)
	for _, doc := range ctx.Facts {
		if factType, ok := doc.Entity.Attributes["fact_type"].(string); ok && factType == "constraint" {
			constraints = append(constraints, types.Constraint{
				Description: doc.Entity.Name,
				Source:      "organizational-context",
				Negotiable:  false,
			})
		}
	}

	// Extract from policies
	for _, doc := range ctx.Policies {
		enforcement, _ := doc.Entity.Attributes["enforcement"].(string)
		constraints = append(constraints, types.Constraint{
			Description: doc.Entity.Name,
			Source:      "policy",
			Negotiable:  enforcement != "mandatory",
		})
	}

	return constraints
}

// buildEnvironmentFromCRF creates environment info from CRF context.
func buildEnvironmentFromCRF(ctx *types.CRFContext) *types.Environment {
	if ctx == nil {
		return nil
	}

	env := &types.Environment{}

	// Technical environment from systems
	var techParts []string
	for _, doc := range ctx.Systems {
		if stack, ok := doc.Entity.Attributes["technology_stack"].([]interface{}); ok {
			for _, tech := range stack {
				if t, ok := tech.(string); ok {
					techParts = append(techParts, t)
				}
			}
		}
		if hosting, ok := doc.Entity.Attributes["hosting"].(string); ok {
			techParts = append(techParts, hosting)
		}
	}
	if len(techParts) > 0 {
		env.Technical = strings.Join(techParts, ", ")
	}

	// Organizational from company
	if org := ctx.GetOrganization(); org != nil {
		var orgParts []string
		if industry, ok := org.Attributes["industry"].(string); ok {
			orgParts = append(orgParts, industry)
		}
		if size, ok := org.Attributes["size"].(string); ok {
			orgParts = append(orgParts, size)
		}
		if len(orgParts) > 0 {
			env.Organizational = strings.Join(orgParts, ", ")
		}
	}

	return env
}

// SaveDRFDocument writes a DRF document to disk.
func SaveDRFDocument(doc *types.DRFDocument) error {
	decisionsDir, err := config.GetDecisionsDir()
	if err != nil {
		return err
	}

	if err := config.EnsureDir(decisionsDir); err != nil {
		return err
	}

	path := filepath.Join(decisionsDir, doc.Decision.ID+".yaml")
	data, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetDRFDocument retrieves a DRF document by ID.
func GetDRFDocument(id string) (*types.DRFDocument, error) {
	decisionsDir, err := config.GetDecisionsDir()
	if err != nil {
		return nil, err
	}

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

// ListDRFDocuments returns all DRF documents, optionally filtered.
func ListDRFDocuments(filters *ListFilters) ([]types.DRFDocument, error) {
	decisionsDir, err := config.GetDecisionsDir()
	if err != nil {
		return nil, err
	}

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
		doc, err := GetDRFDocument(id)
		if err != nil || doc == nil {
			continue
		}

		// Apply filters
		if filters != nil {
			if filters.Status != "" && doc.Meta.Status != filters.Status {
				continue
			}
			if filters.Tag != "" && !containsTag(doc.Meta.Tags, filters.Tag) {
				continue
			}
		}

		docs = append(docs, *doc)
	}

	// Sort by created date, newest first
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Meta.CreatedAt.After(docs[j].Meta.CreatedAt)
	})

	return docs, nil
}

// UpdateStatus changes the status of a DRF document.
func UpdateStatus(id string, status types.DRFStatus) error {
	doc, err := GetDRFDocument(id)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("decision not found: %s", id)
	}

	doc.Meta.Status = status
	doc.Meta.UpdatedAt = time.Now()

	return SaveDRFDocument(doc)
}

// AddTag adds a tag to a DRF document if not already present.
func AddTag(id, tag string) error {
	doc, err := GetDRFDocument(id)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("decision not found: %s", id)
	}

	tag = strings.ToLower(strings.TrimSpace(tag))
	if containsTag(doc.Meta.Tags, tag) {
		return nil
	}

	doc.Meta.Tags = append(doc.Meta.Tags, tag)
	doc.Meta.UpdatedAt = time.Now()

	return SaveDRFDocument(doc)
}

// RemoveTag removes a tag from a DRF document.
func RemoveTag(id, tag string) error {
	doc, err := GetDRFDocument(id)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("decision not found: %s", id)
	}

	tag = strings.ToLower(strings.TrimSpace(tag))
	var newTags []string
	for _, t := range doc.Meta.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}

	doc.Meta.Tags = newTags
	doc.Meta.UpdatedAt = time.Now()

	return SaveDRFDocument(doc)
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

// UpdateSynthesis updates the synthesis (decision outcome) of a DRF document.
func UpdateSynthesis(id string, decision, rationale string) error {
	doc, err := GetDRFDocument(id)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("decision not found: %s", id)
	}

	doc.Synthesis.Decision = decision
	doc.Synthesis.Rationale = rationale
	doc.Meta.UpdatedAt = time.Now()

	return SaveDRFDocument(doc)
}

// AddFollowUp adds a follow-up action to a DRF document.
func AddFollowUp(id string, followUp types.FollowUp) error {
	doc, err := GetDRFDocument(id)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("decision not found: %s", id)
	}

	doc.Synthesis.FollowUps = append(doc.Synthesis.FollowUps, followUp)
	doc.Meta.UpdatedAt = time.Now()

	return SaveDRFDocument(doc)
}

// SetConfidence updates the confidence level of a DRF document.
func SetConfidence(id string, confidence int) error {
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 100 {
		confidence = 100
	}

	doc, err := GetDRFDocument(id)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("decision not found: %s", id)
	}

	doc.CognitiveState.Confidence = confidence
	doc.Meta.UpdatedAt = time.Now()

	return SaveDRFDocument(doc)
}

// ApproveDecision transitions a DRF document to approved status.
func ApproveDecision(id string) error {
	doc, err := GetDRFDocument(id)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("decision not found: %s", id)
	}

	// Can only approve from draft or review status
	if doc.Meta.Status != types.DRFStatusDraft && doc.Meta.Status != types.DRFStatusReview {
		return fmt.Errorf("cannot approve decision in %s status", doc.Meta.Status)
	}

	doc.Meta.Status = types.DRFStatusApproved
	doc.CognitiveState.Phase = types.PhaseDecision
	doc.Meta.UpdatedAt = time.Now()

	return SaveDRFDocument(doc)
}

// SearchDecisions searches decisions by title or intent.
func SearchDecisions(query string) ([]types.DRFDocument, error) {
	docs, err := ListDRFDocuments(nil)
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
