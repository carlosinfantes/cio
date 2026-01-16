package llm

import (
	"testing"

	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

func TestParseResponse(t *testing.T) {
	// Sample advisors for testing
	advisors := []types.Persona{
		{ID: "cto", Name: "Victoria Chen", Role: "Fractional CTO, 3x exit"},
		{ID: "ciso", Name: "Marcus Webb", Role: "Former CISO, Fortune 500"},
		{ID: "vp-eng", Name: "Priya Sharma", Role: "VP Engineering, Scale-up Specialist"},
		{ID: "architect", Name: "Erik Lindqvist", Role: "Principal Architect, Distributed Systems"},
	}

	tests := []struct {
		name           string
		content        string
		advisors       []types.Persona
		wantAdvisors   int
		wantSynthesis  bool
		wantAdvisorIDs []types.AdvisorID
	}{
		{
			name: "full response with all advisors and synthesis",
			content: `## Victoria Chen — Fractional CTO

This is the CTO's response about the technical decision.

## Marcus Webb — Former CISO

Security considerations are important here.

## Priya Sharma — VP Engineering

Team capacity is a concern.

## Erik Lindqvist — Principal Architect

Let me explain the architecture.

## Synthesis

The board recommends moving forward with caution.`,
			advisors:       advisors,
			wantAdvisors:   4,
			wantSynthesis:  true,
			wantAdvisorIDs: []types.AdvisorID{"cto", "ciso", "vp-eng", "architect"},
		},
		{
			name: "partial response with some advisors",
			content: `## Victoria Chen — CTO

Only CTO responded.

## Synthesis

Brief summary.`,
			advisors:       advisors,
			wantAdvisors:   1,
			wantSynthesis:  true,
			wantAdvisorIDs: []types.AdvisorID{"cto"},
		},
		{
			name: "first name matching",
			content: `## Victoria

Response using first name only.

## Marcus

Another response.`,
			advisors:       advisors,
			wantAdvisors:   2,
			wantSynthesis:  false,
			wantAdvisorIDs: []types.AdvisorID{"cto", "ciso"},
		},
		{
			name:           "unstructured response becomes synthesis",
			content:        "This is just plain text without any headers.",
			advisors:       advisors,
			wantAdvisors:   0,
			wantSynthesis:  true,
			wantAdvisorIDs: []types.AdvisorID{},
		},
		{
			name:           "empty content",
			content:        "",
			advisors:       advisors,
			wantAdvisors:   0,
			wantSynthesis:  false,
			wantAdvisorIDs: []types.AdvisorID{},
		},
		{
			name: "role-based matching",
			content: `## Fractional CTO

Matching by role keyword.

## Principal Architect

Another role-based match.`,
			advisors:       advisors,
			wantAdvisors:   2,
			wantSynthesis:  false,
			wantAdvisorIDs: []types.AdvisorID{"cto", "architect"},
		},
		{
			name: "synthesis variations",
			content: `## Victoria Chen

Response here.

## SYNTHESIS

Uppercase synthesis header.`,
			advisors:       advisors,
			wantAdvisors:   1,
			wantSynthesis:  true,
			wantAdvisorIDs: []types.AdvisorID{"cto"},
		},
		{
			name: "synthesis with prefix",
			content: `## Victoria Chen

Response here.

## Board Synthesis

Synthesis with prefix.`,
			advisors:       advisors,
			wantAdvisors:   1,
			wantSynthesis:  true,
			wantAdvisorIDs: []types.AdvisorID{"cto"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseResponse(tt.content, tt.advisors)

			// Check advisor count
			if len(result.Advisors) != tt.wantAdvisors {
				t.Errorf("got %d advisors, want %d", len(result.Advisors), tt.wantAdvisors)
			}

			// Check synthesis presence
			hasSynthesis := result.Synthesis != ""
			if hasSynthesis != tt.wantSynthesis {
				t.Errorf("synthesis present = %v, want %v", hasSynthesis, tt.wantSynthesis)
			}

			// Check advisor IDs
			if len(result.Advisors) == len(tt.wantAdvisorIDs) {
				for i, advisor := range result.Advisors {
					if advisor.AdvisorID != tt.wantAdvisorIDs[i] {
						t.Errorf("advisor[%d].ID = %s, want %s", i, advisor.AdvisorID, tt.wantAdvisorIDs[i])
					}
				}
			}
		})
	}
}

func TestMatchAdvisor(t *testing.T) {
	advisors := []types.Persona{
		{ID: "cto", Name: "Victoria Chen", Role: "Fractional CTO, 3x exit"},
		{ID: "ciso", Name: "Marcus Webb", Role: "Former CISO, Fortune 500"},
	}

	tests := []struct {
		name       string
		headerName string
		wantID     types.AdvisorID
		wantMatch  bool
	}{
		{"full name match", "Victoria Chen", "cto", true},
		{"full name with role", "Victoria Chen — CTO", "cto", true},
		{"first name only", "Victoria", "cto", true},
		{"case insensitive", "VICTORIA CHEN", "cto", true},
		{"role keyword match", "Fractional perspective", "cto", true},
		{"ciso by name", "Marcus Webb", "ciso", true},
		{"ciso by first name", "Marcus", "ciso", true},
		{"no match", "Unknown Person", "", false},
		{"partial name no match", "Vic", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, _, _ := matchAdvisor(tt.headerName, advisors)

			if tt.wantMatch {
				if id != tt.wantID {
					t.Errorf("matchAdvisor(%q) = %s, want %s", tt.headerName, id, tt.wantID)
				}
			} else {
				if id != "" {
					t.Errorf("matchAdvisor(%q) = %s, want no match", tt.headerName, id)
				}
			}
		})
	}
}

func TestParseResponseContent(t *testing.T) {
	advisors := []types.Persona{
		{ID: "cto", Name: "Victoria Chen", Role: "Fractional CTO"},
	}

	content := `## Victoria Chen — Fractional CTO

This is the actual content
with multiple lines
and some details.

## Synthesis

Final thoughts here.`

	result := ParseResponse(content, advisors)

	if len(result.Advisors) != 1 {
		t.Fatalf("expected 1 advisor, got %d", len(result.Advisors))
	}

	// Check content is properly extracted
	if result.Advisors[0].Response == "" {
		t.Error("advisor response should not be empty")
	}

	if result.Synthesis == "" {
		t.Error("synthesis should not be empty")
	}

	// Verify content doesn't include headers
	if result.Advisors[0].Response[:2] == "##" {
		t.Error("response should not start with header")
	}
}
