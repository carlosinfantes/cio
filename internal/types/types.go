// Package types defines shared types for the CIO - Chief Intelligence Officer CLI.
package types

import (
	"fmt"
	"strings"
	"time"
)

// Mode represents the interaction mode with the advisory board.
type Mode string

const (
	ModePanel     Mode = "panel"
	ModeSocratic  Mode = "socratic"
	ModeAdvocate  Mode = "advocate"
	ModeFramework Mode = "framework"
)

// SessionMode represents the current interactive session mode.
type SessionMode string

const (
	SessionModeDiscovery SessionMode = "discovery"
	SessionModePanel     SessionMode = "panel"
)

// AdvisorID uniquely identifies an advisor.
type AdvisorID string

const (
	AdvisorCTO         AdvisorID = "cto"
	AdvisorCISO        AdvisorID = "ciso"
	AdvisorVPEng       AdvisorID = "vp-eng"
	AdvisorArchitect   AdvisorID = "architect"
	AdvisorCFO         AdvisorID = "cfo"
	AdvisorProduct     AdvisorID = "product"
	AdvisorDevOps      AdvisorID = "devops"
	AdvisorFacilitator AdvisorID = "facilitator"
)

// DRF Status represents the lifecycle status of a decision (DRF standard).
type DRFStatus string

const (
	DRFStatusDraft      DRFStatus = "draft"
	DRFStatusReview     DRFStatus = "review"
	DRFStatusApproved   DRFStatus = "approved"
	DRFStatusRejected   DRFStatus = "rejected"
	DRFStatusSuperseded DRFStatus = "superseded"
	DRFStatusArchived   DRFStatus = "archived"
)

// CognitivePhase represents phases in the decision process (DRF standard).
type CognitivePhase string

const (
	PhaseExploration CognitivePhase = "exploration"
	PhaseAnalysis    CognitivePhase = "analysis"
	PhaseSynthesis   CognitivePhase = "synthesis"
	PhaseDecision    CognitivePhase = "decision"
)

// InterventionType represents types of interventions (DRF standard).
type InterventionType string

const (
	InterventionQuestion      InterventionType = "question"
	InterventionChallenge     InterventionType = "challenge"
	InterventionConstraint    InterventionType = "constraint"
	InterventionInsight       InterventionType = "insight"
	InterventionExternalInput InterventionType = "external_input"
)

// CRFEntityType represents types of CRF entities.
type CRFEntityType string

const (
	CRFEntityOrganization CRFEntityType = "organization"
	CRFEntitySystem       CRFEntityType = "system"
	CRFEntityPolicy       CRFEntityType = "policy"
	CRFEntityFact         CRFEntityType = "fact"
	CRFEntityArchitecture CRFEntityType = "architecture"
	CRFEntityCapability   CRFEntityType = "capability"
)

// CRFRelationshipType represents types of relationships between CRF entities.
type CRFRelationshipType string

const (
	RelOwns          CRFRelationshipType = "owns"
	RelOwnedBy       CRFRelationshipType = "owned_by"
	RelDependsOn     CRFRelationshipType = "depends_on"
	RelDependencyOf  CRFRelationshipType = "dependency_of"
	RelConstrains    CRFRelationshipType = "constrains"
	RelConstrainedBy CRFRelationshipType = "constrained_by"
	RelInvalidates   CRFRelationshipType = "invalidates"
	RelInvalidatedBy CRFRelationshipType = "invalidated_by"
	RelPartOf        CRFRelationshipType = "part_of"
	RelContains      CRFRelationshipType = "contains"
	RelProduces      CRFRelationshipType = "produces"
	RelProducedBy    CRFRelationshipType = "produced_by"
	RelRelatedTo     CRFRelationshipType = "related_to"
)

// Config holds the application configuration.
type Config struct {
	APIKey                  string      `yaml:"api_key"`
	Model                   string      `yaml:"model"`
	DefaultMode             Mode        `yaml:"default_mode"`
	DefaultAdvisors         []AdvisorID `yaml:"default_advisors"`
	AutoSummonSpecialists   bool        `yaml:"auto_summon_specialists"`
	ContextRefreshDays      int         `yaml:"context_refresh_days"`
	MaxAdvisors             int         `yaml:"max_advisors"`
	StartInDiscovery        bool        `yaml:"start_in_discovery"`
	DRFVersion              string      `yaml:"drf_version"`
	CRFVersion              string      `yaml:"crf_version"`
	EnableContextValidation bool        `yaml:"enable_context_validation"`
	SchemaPath              string      `yaml:"schema_path"`

	// Plugin system fields
	ActiveDomain     string   `yaml:"active_domain"`
	InstalledDomains []string `yaml:"installed_domains"`

	// Registry configuration
	RegistryURL string `yaml:"registry_url,omitempty"`
}

// DefaultRegistryURL is the default plugin registry location.
const DefaultRegistryURL = "https://raw.githubusercontent.com/carlosinfantes/cio-plugin-registry/main"

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Model:                   "anthropic/claude-3.5-sonnet",
		DefaultMode:             ModePanel,
		DefaultAdvisors:         []AdvisorID{AdvisorCTO, AdvisorCISO, AdvisorVPEng, AdvisorArchitect},
		AutoSummonSpecialists:   true,
		ContextRefreshDays:      30,
		MaxAdvisors:             5,
		StartInDiscovery:        true,
		DRFVersion:              "0.1.0",
		CRFVersion:              "0.1.0",
		EnableContextValidation: true,
		SchemaPath:              "",
		ActiveDomain:            "",
		InstalledDomains:        []string{},
		RegistryURL:             DefaultRegistryURL,
	}
}

// Persona represents an advisor's identity and behavior.
type Persona struct {
	ID                 AdvisorID
	Name               string
	Role               string
	Color              string
	Emoji              string
	ThinkingStyle      string
	Background         string
	Priorities         []string
	CatchPhrases       []string
	AutoSummonKeywords []string // For specialists only
	IsSpecialist       bool
}

// AdvisorResponse represents an individual advisor's contribution.
type AdvisorResponse struct {
	AdvisorID AdvisorID
	Name      string
	Role      string
	Response  string
}

// ParsedResponse represents the full parsed response from the LLM.
type ParsedResponse struct {
	Advisors  []AdvisorResponse
	Synthesis string
}

// SummonResult represents a specialist summoning with reason.
type SummonResult struct {
	Specialist      Persona
	Reason          string
	MatchedKeywords []string
}

// ============================================================================
// DRF (Decision Reasoning Format) Types - v0.1.0
// ============================================================================

// DRFDocument is a complete Decision Reasoning Format document.
type DRFDocument struct {
	DRFVersion         string              `yaml:"drf_version" json:"drf_version"`
	Decision           DRFDecision         `yaml:"decision" json:"decision"`
	Context            DRFContext          `yaml:"context" json:"context"`
	CognitiveState     CognitiveState      `yaml:"cognitive_state" json:"cognitive_state"`
	Reasoning          *Reasoning          `yaml:"reasoning,omitempty" json:"reasoning,omitempty"`
	Interventions      []Intervention      `yaml:"interventions,omitempty" json:"interventions,omitempty"`
	Assumptions        []Assumption        `yaml:"assumptions,omitempty" json:"assumptions,omitempty"`
	UnresolvedTensions []Tension           `yaml:"unresolved_tensions,omitempty" json:"unresolved_tensions,omitempty"`
	Synthesis          DRFSynthesis        `yaml:"synthesis" json:"synthesis"`
	ContextValidation  *ContextValidation  `yaml:"context_validation,omitempty" json:"context_validation,omitempty"`
	Meta               DRFMeta             `yaml:"meta" json:"meta"`
}

// DRFDecision contains the core identity and intent of a decision.
type DRFDecision struct {
	ID               string            `yaml:"id" json:"id"`
	Title            string            `yaml:"title" json:"title"`
	Domain           string            `yaml:"domain,omitempty" json:"domain,omitempty"`
	Intent           string            `yaml:"intent" json:"intent"`
	RelatedDecisions []RelatedDecision `yaml:"related_decisions,omitempty" json:"related_decisions,omitempty"`
}

// RelatedDecision represents a reference to another decision.
type RelatedDecision struct {
	ID           string `yaml:"id" json:"id"`
	Relationship string `yaml:"relationship" json:"relationship"` // supersedes, superseded_by, depends_on, dependency_of, related_to, conflicts_with
	Description  string `yaml:"description,omitempty" json:"description,omitempty"`
}

// DRFContext contains environmental and situational context.
type DRFContext struct {
	Constraints []Constraint `yaml:"constraints" json:"constraints"`
	Objectives  []Objective  `yaml:"objectives" json:"objectives"`
	Environment *Environment `yaml:"environment,omitempty" json:"environment,omitempty"`
}

// Constraint represents a hard constraint that must be satisfied.
type Constraint struct {
	Description string `yaml:"description" json:"description"`
	Source      string `yaml:"source,omitempty" json:"source,omitempty"`
	Negotiable  bool   `yaml:"negotiable,omitempty" json:"negotiable,omitempty"`
}

// Objective represents a goal or success criterion.
type Objective struct {
	Description string `yaml:"description" json:"description"`
	Priority    string `yaml:"priority,omitempty" json:"priority,omitempty"` // must_have, should_have, nice_to_have
	Measurable  bool   `yaml:"measurable,omitempty" json:"measurable,omitempty"`
}

// Environment describes relevant environmental factors.
type Environment struct {
	Technical      string `yaml:"technical,omitempty" json:"technical,omitempty"`
	Organizational string `yaml:"organizational,omitempty" json:"organizational,omitempty"`
	Temporal       string `yaml:"temporal,omitempty" json:"temporal,omitempty"`
}

// CognitiveState represents the current phase and confidence.
type CognitiveState struct {
	Phase      CognitivePhase `yaml:"phase" json:"phase"`
	Confidence int            `yaml:"confidence" json:"confidence"` // 0-100
	PhaseNotes string         `yaml:"phase_notes,omitempty" json:"phase_notes,omitempty"`
}

// Reasoning captures explicit reasoning patterns applied.
type Reasoning struct {
	PatternsApplied []string `yaml:"patterns_applied,omitempty" json:"patterns_applied,omitempty"`
	Notes           string   `yaml:"notes,omitempty" json:"notes,omitempty"`
}

// Intervention represents a key question, challenge, or input.
type Intervention struct {
	ID        string           `yaml:"id" json:"id"`
	Type      InterventionType `yaml:"type" json:"type"`
	Content   string           `yaml:"content" json:"content"`
	Source    string           `yaml:"source,omitempty" json:"source,omitempty"`
	Timestamp time.Time        `yaml:"timestamp,omitempty" json:"timestamp,omitempty"`
	Impact    string           `yaml:"impact,omitempty" json:"impact,omitempty"`
}

// Assumption represents an explicit or implicit premise.
type Assumption struct {
	Description string `yaml:"description" json:"description"`
	Validated   bool   `yaml:"validated" json:"validated"`
	Confidence  int    `yaml:"confidence,omitempty" json:"confidence,omitempty"` // 0-100
	Source      string `yaml:"source,omitempty" json:"source,omitempty"`
}

// Tension represents an unresolved trade-off or risk.
type Tension struct {
	Description string `yaml:"description" json:"description"`
	Impact      string `yaml:"impact" json:"impact"` // low, medium, high, critical
	Mitigation  string `yaml:"mitigation,omitempty" json:"mitigation,omitempty"`
	AcceptedBy  string `yaml:"accepted_by,omitempty" json:"accepted_by,omitempty"`
}

// DRFSynthesis contains the consolidated decision outcome.
type DRFSynthesis struct {
	Decision     string        `yaml:"decision" json:"decision"`
	Rationale    string        `yaml:"rationale" json:"rationale"`
	FollowUps    []FollowUp    `yaml:"follow_ups,omitempty" json:"follow_ups,omitempty"`
	Alternatives []Alternative `yaml:"alternatives,omitempty" json:"alternatives,omitempty"`
}

// FollowUp represents a next step or action required.
type FollowUp struct {
	Action  string `yaml:"action" json:"action"`
	Owner   string `yaml:"owner,omitempty" json:"owner,omitempty"`
	DueDate string `yaml:"due_date,omitempty" json:"due_date,omitempty"`
}

// Alternative represents a ranked alternative outcome.
type Alternative struct {
	Decision                    string `yaml:"decision" json:"decision"`
	RationaleAgainst            string `yaml:"rationale_against" json:"rationale_against"`
	ConditionsForReconsideration string `yaml:"conditions_for_reconsideration,omitempty" json:"conditions_for_reconsideration,omitempty"`
}

// ContextValidation links the decision to CRF entities.
type ContextValidation struct {
	ValidatedAt    time.Time       `yaml:"validated_at,omitempty" json:"validated_at,omitempty"`
	ContextRefs    []ContextRef    `yaml:"context_refs,omitempty" json:"context_refs,omitempty"`
	ContextOutputs []ContextOutput `yaml:"context_outputs,omitempty" json:"context_outputs,omitempty"`
}

// ContextRef references a CRF entity for validation.
type ContextRef struct {
	ContextID        string `yaml:"context_id" json:"context_id"`
	ContextType      string `yaml:"context_type" json:"context_type"`
	ContextName      string `yaml:"context_name,omitempty" json:"context_name,omitempty"`
	ValidationStatus string `yaml:"validation_status" json:"validation_status"` // satisfied, violated, acknowledged, not_applicable
	AdvisoryNotes    string `yaml:"advisory_notes,omitempty" json:"advisory_notes,omitempty"`
}

// ContextOutput describes CRF entities this decision affects.
type ContextOutput struct {
	Action     string                 `yaml:"action" json:"action"` // creates, updates, invalidates
	EntityType string                 `yaml:"entity_type" json:"entity_type"`
	EntityID   string                 `yaml:"entity_id,omitempty" json:"entity_id,omitempty"`
	EntityData map[string]interface{} `yaml:"entity_data,omitempty" json:"entity_data,omitempty"`
	Reason     string                 `yaml:"reason,omitempty" json:"reason,omitempty"`
}

// DRFMeta contains metadata about the decision document.
type DRFMeta struct {
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at,omitempty" json:"updated_at,omitempty"`
	Status    DRFStatus `yaml:"status" json:"status"`
	Actors    []Actor   `yaml:"actors,omitempty" json:"actors,omitempty"`
	Source    string    `yaml:"source,omitempty" json:"source,omitempty"`
	Tags      []string  `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// Actor represents a person or system involved in the decision.
type Actor struct {
	Name  string `yaml:"name" json:"name"`
	Role  string `yaml:"role" json:"role"` // author, reviewer, approver, contributor, stakeholder
	Email string `yaml:"email,omitempty" json:"email,omitempty"`
}

// ============================================================================
// CRF (Context Reasoning Format) Types - v0.1.0
// ============================================================================

// CRFDocument is a complete Context Reasoning Format document.
type CRFDocument struct {
	CRFVersion string    `yaml:"crf_version" json:"crf_version"`
	Entity     CRFEntity `yaml:"entity" json:"entity"`
}

// CRFEntity represents a node in the context knowledge graph.
type CRFEntity struct {
	ID            string                 `yaml:"id" json:"id"`
	Type          CRFEntityType          `yaml:"type" json:"type"`
	Name          string                 `yaml:"name" json:"name"`
	Description   string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Attributes    map[string]interface{} `yaml:"attributes,omitempty" json:"attributes,omitempty"`
	Validity      *Validity              `yaml:"validity,omitempty" json:"validity,omitempty"`
	Relationships []Relationship         `yaml:"relationships,omitempty" json:"relationships,omitempty"`
	Supersedes    *Supersedes            `yaml:"supersedes,omitempty" json:"supersedes,omitempty"`
	Provenance    Provenance             `yaml:"provenance" json:"provenance"`
	Tags          []string               `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// Validity specifies temporal bounds for context validity.
type Validity struct {
	ValidFrom  time.Time `yaml:"valid_from,omitempty" json:"valid_from,omitempty"`
	ValidUntil time.Time `yaml:"valid_until,omitempty" json:"valid_until,omitempty"`
}

// Relationship represents an edge to another entity.
type Relationship struct {
	TargetID    string              `yaml:"target_id" json:"target_id"`
	Type        CRFRelationshipType `yaml:"type" json:"type"`
	Description string              `yaml:"description,omitempty" json:"description,omitempty"`
}

// Supersedes references the entity this one replaces.
type Supersedes struct {
	EntityID     string    `yaml:"entity_id" json:"entity_id"`
	Reason       string    `yaml:"reason,omitempty" json:"reason,omitempty"`
	SupersededAt time.Time `yaml:"superseded_at,omitempty" json:"superseded_at,omitempty"`
}

// Provenance tracks origin and authorship.
type Provenance struct {
	Source    string    `yaml:"source" json:"source"`
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`
	CreatedBy string    `yaml:"created_by,omitempty" json:"created_by,omitempty"`
	UpdatedAt time.Time `yaml:"updated_at,omitempty" json:"updated_at,omitempty"`
	UpdatedBy string    `yaml:"updated_by,omitempty" json:"updated_by,omitempty"`
}

// ============================================================================
// CRF Entity Attribute Types (typed helpers for common attribute patterns)
// ============================================================================

// OrganizationAttributes holds attributes for organization entities.
type OrganizationAttributes struct {
	OrgType              string   `yaml:"org_type,omitempty" json:"org_type,omitempty"` // company, division, department, team, squad, working_group
	Size                 string   `yaml:"size,omitempty" json:"size,omitempty"`         // startup, small, medium, large, enterprise
	Headcount            int      `yaml:"headcount,omitempty" json:"headcount,omitempty"`
	Location             string   `yaml:"location,omitempty" json:"location,omitempty"`
	Industry             string   `yaml:"industry,omitempty" json:"industry,omitempty"`
	ComplianceFrameworks []string `yaml:"compliance_frameworks,omitempty" json:"compliance_frameworks,omitempty"`
}

// SystemAttributes holds attributes for system entities.
type SystemAttributes struct {
	SystemType         string   `yaml:"system_type,omitempty" json:"system_type,omitempty"` // application, service, platform, infrastructure, database, integration, tool
	Status             string   `yaml:"status,omitempty" json:"status,omitempty"`           // planned, development, staging, production, deprecated, decommissioned
	Criticality        string   `yaml:"criticality,omitempty" json:"criticality,omitempty"` // low, medium, high, critical
	TechnologyStack    []string `yaml:"technology_stack,omitempty" json:"technology_stack,omitempty"`
	Hosting            string   `yaml:"hosting,omitempty" json:"hosting,omitempty"`
	DataClassification string   `yaml:"data_classification,omitempty" json:"data_classification,omitempty"` // public, internal, confidential, restricted
}

// PolicyAttributes holds attributes for policy entities.
type PolicyAttributes struct {
	PolicyType        string `yaml:"policy_type,omitempty" json:"policy_type,omitempty"` // governance, security, compliance, architectural, operational, financial
	Enforcement       string `yaml:"enforcement,omitempty" json:"enforcement,omitempty"` // mandatory, recommended, advisory
	Scope             string `yaml:"scope,omitempty" json:"scope,omitempty"`
	Rationale         string `yaml:"rationale,omitempty" json:"rationale,omitempty"`
	ExceptionsProcess string `yaml:"exceptions_process,omitempty" json:"exceptions_process,omitempty"`
	Owner             string `yaml:"owner,omitempty" json:"owner,omitempty"`
	ReviewCycle       string `yaml:"review_cycle,omitempty" json:"review_cycle,omitempty"`
}

// FactAttributes holds attributes for fact entities.
type FactAttributes struct {
	FactType        string      `yaml:"fact_type,omitempty" json:"fact_type,omitempty"` // contract, budget, timeline, constraint, metric, event, status
	Value           interface{} `yaml:"value,omitempty" json:"value,omitempty"`
	Unit            string      `yaml:"unit,omitempty" json:"unit,omitempty"`
	Confidence      int         `yaml:"confidence,omitempty" json:"confidence,omitempty"` // 0-100
	SourceReference string      `yaml:"source_reference,omitempty" json:"source_reference,omitempty"`
	Verified        bool        `yaml:"verified,omitempty" json:"verified,omitempty"`
	VerifiedAt      time.Time   `yaml:"verified_at,omitempty" json:"verified_at,omitempty"`
}

// ArchitectureAttributes holds attributes for architecture entities.
type ArchitectureAttributes struct {
	ArchitectureType string   `yaml:"architecture_type,omitempty" json:"architecture_type,omitempty"` // pattern, principle, standard, guideline, reference, decision
	Domain           string   `yaml:"domain,omitempty" json:"domain,omitempty"`
	Maturity         string   `yaml:"maturity,omitempty" json:"maturity,omitempty"`       // emerging, established, mature, declining, deprecated
	AdoptionStatus   string   `yaml:"adoption_status,omitempty" json:"adoption_status,omitempty"` // proposed, pilot, adopted, standard, retiring
	Alternatives     []string `yaml:"alternatives,omitempty" json:"alternatives,omitempty"`
}

// CapabilityAttributes holds attributes for capability entities.
type CapabilityAttributes struct {
	CapabilityType       string `yaml:"capability_type,omitempty" json:"capability_type,omitempty"` // skill, tool, process, practice, certification
	Proficiency          string `yaml:"proficiency,omitempty" json:"proficiency,omitempty"`         // none, beginner, intermediate, advanced, expert
	Coverage             int    `yaml:"coverage,omitempty" json:"coverage,omitempty"`               // 0-100 percentage
	TrainingAvailable    bool   `yaml:"training_available,omitempty" json:"training_available,omitempty"`
	StrategicImportance  string `yaml:"strategic_importance,omitempty" json:"strategic_importance,omitempty"` // low, medium, high, critical
}

// ============================================================================
// CRF Context Collection (for loading and managing CRF entities)
// ============================================================================

// CRFContext holds all loaded CRF entities grouped by type.
type CRFContext struct {
	Organizations []CRFDocument
	Systems       []CRFDocument
	Capabilities  []CRFDocument
	Facts         []CRFDocument
	Policies      []CRFDocument
	Architecture  []CRFDocument
}

// GetOrganization returns the root organization entity (company).
func (c *CRFContext) GetOrganization() *CRFEntity {
	for _, doc := range c.Organizations {
		if attrs, ok := doc.Entity.Attributes["org_type"].(string); ok && attrs == "company" {
			return &doc.Entity
		}
	}
	if len(c.Organizations) > 0 {
		return &c.Organizations[0].Entity
	}
	return nil
}

// GetTeams returns all team organization entities.
func (c *CRFContext) GetTeams() []CRFEntity {
	var teams []CRFEntity
	for _, doc := range c.Organizations {
		if attrs, ok := doc.Entity.Attributes["org_type"].(string); ok && attrs == "team" {
			teams = append(teams, doc.Entity)
		}
	}
	return teams
}

// GetEntityByID finds an entity by its ID across all types.
func (c *CRFContext) GetEntityByID(id string) *CRFEntity {
	for _, doc := range c.Organizations {
		if doc.Entity.ID == id {
			return &doc.Entity
		}
	}
	for _, doc := range c.Systems {
		if doc.Entity.ID == id {
			return &doc.Entity
		}
	}
	for _, doc := range c.Capabilities {
		if doc.Entity.ID == id {
			return &doc.Entity
		}
	}
	for _, doc := range c.Facts {
		if doc.Entity.ID == id {
			return &doc.Entity
		}
	}
	for _, doc := range c.Policies {
		if doc.Entity.ID == id {
			return &doc.Entity
		}
	}
	for _, doc := range c.Architecture {
		if doc.Entity.ID == id {
			return &doc.Entity
		}
	}
	return nil
}

// AllEntities returns all entities as a flat list.
func (c *CRFContext) AllEntities() []CRFEntity {
	var all []CRFEntity
	for _, doc := range c.Organizations {
		all = append(all, doc.Entity)
	}
	for _, doc := range c.Systems {
		all = append(all, doc.Entity)
	}
	for _, doc := range c.Capabilities {
		all = append(all, doc.Entity)
	}
	for _, doc := range c.Facts {
		all = append(all, doc.Entity)
	}
	for _, doc := range c.Policies {
		all = append(all, doc.Entity)
	}
	for _, doc := range c.Architecture {
		all = append(all, doc.Entity)
	}
	return all
}

// ============================================================================
// Advisory Board Flow Types (preserved from original)
// ============================================================================

// SocraticState tracks the two-turn Socratic flow.
type SocraticState struct {
	Question  string         // Original question
	Questions []string       // Clarifying questions from LLM
	Answers   map[int]string // User answers indexed by question number
	Phase     string         // "questions" | "answering" | "complete"
}

// NewSocraticState creates a new Socratic state for a question.
func NewSocraticState(question string) *SocraticState {
	return &SocraticState{
		Question: question,
		Answers:  make(map[int]string),
		Phase:    "questions",
	}
}

// AllAnswered returns true if all questions have been answered.
func (s *SocraticState) AllAnswered() bool {
	return len(s.Answers) >= len(s.Questions)
}

// GetEnrichedContext returns the question + Q&A for panel prompt.
func (s *SocraticState) GetEnrichedContext() string {
	var sb strings.Builder
	sb.WriteString("Original question: " + s.Question + "\n\n")
	sb.WriteString("Clarifying information:\n")
	for i, q := range s.Questions {
		if answer, ok := s.Answers[i]; ok {
			sb.WriteString(fmt.Sprintf("Q: %s\nA: %s\n\n", q, answer))
		}
	}
	return sb.String()
}

// FrameworkCriterion represents an evaluation criterion for decision framework.
type FrameworkCriterion struct {
	Name        string  // e.g., "Cost", "Scalability"
	Description string  // Brief explanation
	Weight      float64 // 1-5 importance weight
}

// FrameworkScore represents a score for one option on one criterion.
type FrameworkScore struct {
	Option    string
	Criterion string
	Score     int    // 1-5
	Rationale string // Brief explanation
}

// FrameworkState tracks the multi-turn Framework decision flow.
type FrameworkState struct {
	Question       string               // Original question (e.g., "AWS vs GCP")
	Options        []string             // Parsed options (e.g., ["AWS", "GCP"])
	Criteria       []FrameworkCriterion // Evaluation criteria
	Scores         []FrameworkScore     // Scores for each option/criterion
	Recommendation string               // Final recommendation
	Confidence     string               // low/medium/high
	Phase          string               // "criteria" | "confirming" | "scoring" | "complete"
}

// NewFrameworkState creates a new Framework state.
func NewFrameworkState(question string, options []string) *FrameworkState {
	return &FrameworkState{
		Question: question,
		Options:  options,
		Criteria: []FrameworkCriterion{},
		Scores:   []FrameworkScore{},
		Phase:    "criteria",
	}
}

// GetScoreMatrix returns scores organized by option and criterion.
func (f *FrameworkState) GetScoreMatrix() map[string]map[string]int {
	matrix := make(map[string]map[string]int)
	for _, opt := range f.Options {
		matrix[opt] = make(map[string]int)
	}
	for _, score := range f.Scores {
		if _, ok := matrix[score.Option]; ok {
			matrix[score.Option][score.Criterion] = score.Score
		}
	}
	return matrix
}

// GetWeightedScores calculates weighted totals for each option.
func (f *FrameworkState) GetWeightedScores() map[string]float64 {
	totals := make(map[string]float64)
	matrix := f.GetScoreMatrix()

	for _, opt := range f.Options {
		var total float64
		for _, crit := range f.Criteria {
			if score, ok := matrix[opt][crit.Name]; ok {
				total += float64(score) * crit.Weight
			}
		}
		totals[opt] = total
	}
	return totals
}

// ============================================================================
// Utility Types
// ============================================================================

// StalenessWarning indicates context files need updating.
type StalenessWarning struct {
	DaysSinceUpdate int
	OldestFile      string
	LastUpdated     time.Time
}

// ContextConflict represents a contradiction between context and decisions.
type ContextConflict struct {
	Field         string // entity id or name
	ContextValue  string
	DecisionValue string
	DecisionID    string
	Severity      string // "warning", "info"
}

// UpdateSuggestion suggests a context entity update based on question content.
type UpdateSuggestion struct {
	EntityID string
	Field    string
	OldValue string
	NewValue string
	Reason   string // What triggered the suggestion
}
