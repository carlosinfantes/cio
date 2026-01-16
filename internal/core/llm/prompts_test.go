package llm

import (
	"strings"
	"testing"
	"time"

	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

// createTestCRFContext creates a CRFContext for testing
func createTestCRFContext() *types.CRFContext {
	now := time.Now()
	return &types.CRFContext{
		Organizations: []types.CRFDocument{
			{
				CRFVersion: "0.1.0",
				Entity: types.CRFEntity{
					ID:          "org-acme",
					Type:        types.CRFEntityOrganization,
					Name:        "Acme Corp",
					Description: "Series A SaaS company",
					Attributes: map[string]interface{}{
						"org_type":              "company",
						"industry":              "SaaS",
						"stage":                 "Series A",
						"business_model":        "B2B",
						"compliance_frameworks": []interface{}{"SOC2", "GDPR"},
					},
					Provenance: types.Provenance{
						Source:    "manual",
						CreatedAt: now,
					},
				},
			},
			{
				CRFVersion: "0.1.0",
				Entity: types.CRFEntity{
					ID:          "team-eng",
					Type:        types.CRFEntityOrganization,
					Name:        "Engineering Team",
					Description: "Engineering team at Acme Corp",
					Attributes: map[string]interface{}{
						"org_type":  "team",
						"headcount": 15,
						"skills":    []interface{}{"frontend", "backend", "devops"},
					},
					Provenance: types.Provenance{
						Source:    "manual",
						CreatedAt: now,
					},
				},
			},
		},
		Systems: []types.CRFDocument{
			{
				CRFVersion: "0.1.0",
				Entity: types.CRFEntity{
					ID:          "system-main",
					Type:        types.CRFEntitySystem,
					Name:        "Main Platform",
					Description: "Primary technology platform",
					Attributes: map[string]interface{}{
						"primary_language":   "TypeScript",
						"hosting":            "AWS",
						"technology_stack":   []interface{}{"ECS", "Lambda"},
					},
					Provenance: types.Provenance{
						Source:    "manual",
						CreatedAt: now,
					},
				},
			},
		},
		Facts: []types.CRFDocument{
			{
				CRFVersion: "0.1.0",
				Entity: types.CRFEntity{
					ID:          "fact-runway",
					Type:        types.CRFEntityFact,
					Name:        "Financial Runway",
					Description: "Current runway: 18 months",
					Attributes: map[string]interface{}{
						"fact_type": "constraint",
						"value":     18,
						"unit":      "months",
					},
					Provenance: types.Provenance{
						Source:    "manual",
						CreatedAt: now,
					},
				},
			},
		},
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	advisors := []types.Persona{
		{
			ID:            "cto",
			Name:          "Victoria Chen",
			Role:          "Fractional CTO",
			ThinkingStyle: "What's the 10x outcome?",
			Background:    "Former VP at Stripe",
			Priorities:    []string{"Strategy", "Hiring"},
		},
		{
			ID:            "ciso",
			Name:          "Marcus Webb",
			Role:          "CISO",
			ThinkingStyle: "What could go wrong?",
			Background:    "20 years security",
			Priorities:    []string{"Risk", "Compliance"},
		},
	}

	context := createTestCRFContext()

	tests := []struct {
		name     string
		advisors []types.Persona
		context  *types.CRFContext
		mode     types.Mode
		contains []string
	}{
		{
			name:     "panel mode with advisors and context",
			advisors: advisors,
			context:  context,
			mode:     types.ModePanel,
			contains: []string{
				"Victoria Chen",
				"Marcus Webb",
				"Acme Corp",
				"15",
				"MODE: PANEL DISCUSSION",
			},
		},
		{
			name:     "socratic mode",
			advisors: advisors,
			context:  context,
			mode:     types.ModeSocratic,
			contains: []string{
				"MODE: SOCRATIC",
				"clarifying questions",
			},
		},
		{
			name:     "advocate mode",
			advisors: advisors,
			context:  context,
			mode:     types.ModeAdvocate,
			contains: []string{
				"MODE: DEVIL'S ADVOCATE",
				"Challenge the premise",
			},
		},
		{
			name:     "framework mode",
			advisors: advisors,
			context:  context,
			mode:     types.ModeFramework,
			contains: []string{
				"MODE: DECISION FRAMEWORK",
				"evaluation matrix",
			},
		},
		{
			name:     "nil context",
			advisors: advisors,
			context:  nil,
			mode:     types.ModePanel,
			contains: []string{
				"Victoria Chen",
				"Marcus Webb",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSystemPrompt(tt.advisors, tt.context, tt.mode)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("prompt should contain %q", want)
				}
			}

			// Nil context should not have PROJECT CONTEXT section
			if tt.context == nil && strings.Contains(result, "PROJECT CONTEXT") {
				t.Error("nil context should not include PROJECT CONTEXT section")
			}
		})
	}
}

func TestBuildSystemPromptAdvisorDetails(t *testing.T) {
	advisor := types.Persona{
		ID:            "cto",
		Name:          "Victoria Chen",
		Role:          "Fractional CTO",
		ThinkingStyle: "Strategic thinking",
		Background:    "Built 3 companies",
		Priorities:    []string{"Growth", "Technical debt"},
	}

	result := BuildSystemPrompt([]types.Persona{advisor}, nil, types.ModePanel)

	// Check all advisor details are included
	checks := []string{
		"Victoria Chen — Fractional CTO",
		"Strategic thinking",
		"Built 3 companies",
		"Growth, Technical debt",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("prompt should contain %q", check)
		}
	}
}

func TestBuildSystemPromptContextDetails(t *testing.T) {
	now := time.Now()
	context := &types.CRFContext{
		Organizations: []types.CRFDocument{
			{
				CRFVersion: "0.1.0",
				Entity: types.CRFEntity{
					ID:          "org-testco",
					Type:        types.CRFEntityOrganization,
					Name:        "TestCo",
					Description: "Seed FinTech company",
					Attributes: map[string]interface{}{
						"org_type":              "company",
						"industry":              "FinTech",
						"stage":                 "Seed",
						"business_model":        "B2C",
						"compliance_frameworks": []interface{}{"PCI-DSS"},
					},
					Provenance: types.Provenance{
						Source:    "manual",
						CreatedAt: now,
					},
				},
			},
			{
				CRFVersion: "0.1.0",
				Entity: types.CRFEntity{
					ID:          "team-eng",
					Type:        types.CRFEntityOrganization,
					Name:        "Engineering Team",
					Description: "Engineering team",
					Attributes: map[string]interface{}{
						"org_type":  "team",
						"headcount": 8,
					},
					Provenance: types.Provenance{
						Source:    "manual",
						CreatedAt: now,
					},
				},
			},
		},
		Systems: []types.CRFDocument{
			{
				CRFVersion: "0.1.0",
				Entity: types.CRFEntity{
					ID:          "system-main",
					Type:        types.CRFEntitySystem,
					Name:        "Main Platform",
					Description: "Primary platform",
					Attributes: map[string]interface{}{
						"primary_language": "Go",
						"hosting":          "GCP",
					},
					Provenance: types.Provenance{
						Source:    "manual",
						CreatedAt: now,
					},
				},
			},
		},
		Facts: []types.CRFDocument{
			{
				CRFVersion: "0.1.0",
				Entity: types.CRFEntity{
					ID:          "fact-runway",
					Type:        types.CRFEntityFact,
					Name:        "Runway",
					Description: "12 months runway",
					Attributes: map[string]interface{}{
						"fact_type": "constraint",
						"value":     12,
						"unit":      "months",
					},
					Provenance: types.Provenance{
						Source:    "manual",
						CreatedAt: now,
					},
				},
			},
		},
	}

	result := BuildSystemPrompt([]types.Persona{}, context, types.ModePanel)

	// Check for values that are actually included in the CRF context format
	// The format function outputs: company name, industry, compliance, team count, cloud hosting
	checks := []string{
		"TestCo",
		"FinTech",
		"PCI-DSS",
		"8",      // headcount
		"GCP",    // hosting
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("prompt should contain context detail %q", check)
		}
	}
}

func TestBuildUserPrompt(t *testing.T) {
	tests := []struct {
		name     string
		question string
		contains []string
	}{
		{
			name:     "simple question",
			question: "Should we use Kubernetes?",
			contains: []string{"Should we use Kubernetes?", "advisory board"},
		},
		{
			name:     "complex question",
			question: "We're considering a microservices architecture. What are the trade-offs?",
			contains: []string{"microservices", "trade-offs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildUserPrompt(tt.question)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("user prompt should contain %q", want)
				}
			}
		})
	}
}

func TestFormatModeInstructions(t *testing.T) {
	tests := []struct {
		mode     types.Mode
		contains string
	}{
		{types.ModePanel, "PANEL DISCUSSION"},
		{types.ModeSocratic, "SOCRATIC"},
		{types.ModeAdvocate, "DEVIL'S ADVOCATE"},
		{types.ModeFramework, "DECISION FRAMEWORK"},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			result := formatModeInstructions(tt.mode)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("mode %s should contain %q", tt.mode, tt.contains)
			}
		})
	}
}
