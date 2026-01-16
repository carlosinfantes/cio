package advisors

import (
	"testing"

	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

func TestCoreAdvisors(t *testing.T) {
	advisors := CoreAdvisors()

	if len(advisors) != 4 {
		t.Errorf("expected 4 core advisors, got %d", len(advisors))
	}

	// Verify core advisor IDs
	expectedIDs := []types.AdvisorID{types.AdvisorCTO, types.AdvisorCISO, types.AdvisorVPEng, types.AdvisorArchitect}
	for i, advisor := range advisors {
		if advisor.ID != expectedIDs[i] {
			t.Errorf("advisor[%d].ID = %s, want %s", i, advisor.ID, expectedIDs[i])
		}
		if advisor.IsSpecialist {
			t.Errorf("core advisor %s should not be specialist", advisor.ID)
		}
	}
}

func TestSpecialists(t *testing.T) {
	specialists := Specialists()

	if len(specialists) != 3 {
		t.Errorf("expected 3 specialists, got %d", len(specialists))
	}

	// Verify specialist IDs
	expectedIDs := []types.AdvisorID{types.AdvisorCFO, types.AdvisorProduct, types.AdvisorDevOps}
	for i, specialist := range specialists {
		if specialist.ID != expectedIDs[i] {
			t.Errorf("specialist[%d].ID = %s, want %s", i, specialist.ID, expectedIDs[i])
		}
		if !specialist.IsSpecialist {
			t.Errorf("specialist %s should be marked as specialist", specialist.ID)
		}
		if len(specialist.AutoSummonKeywords) == 0 {
			t.Errorf("specialist %s should have auto-summon keywords", specialist.ID)
		}
	}
}

func TestAllAdvisors(t *testing.T) {
	all := AllAdvisors()

	if len(all) != 7 {
		t.Errorf("expected 7 total advisors, got %d", len(all))
	}

	// First 4 should be core, last 3 specialists
	for i := 0; i < 4; i++ {
		if all[i].IsSpecialist {
			t.Errorf("advisor[%d] should be core, not specialist", i)
		}
	}
	for i := 4; i < 7; i++ {
		if !all[i].IsSpecialist {
			t.Errorf("advisor[%d] should be specialist", i)
		}
	}
}

func TestGetByID(t *testing.T) {
	tests := []struct {
		id        types.AdvisorID
		wantFound bool
		wantName  string
	}{
		{types.AdvisorCTO, true, "Victoria Chen"},
		{types.AdvisorCISO, true, "Marcus Webb"},
		{types.AdvisorVPEng, true, "Priya Sharma"},
		{types.AdvisorArchitect, true, "Erik Lindqvist"},
		{types.AdvisorCFO, true, "David Park"},
		{types.AdvisorProduct, true, "Sarah Mitchell"},
		{types.AdvisorDevOps, true, "Alex Petrov"},
		{"unknown", false, ""},
		{"", false, ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			advisor, found := GetByID(tt.id)

			if found != tt.wantFound {
				t.Errorf("GetByID(%s) found = %v, want %v", tt.id, found, tt.wantFound)
			}

			if tt.wantFound && advisor.Name != tt.wantName {
				t.Errorf("GetByID(%s).Name = %s, want %s", tt.id, advisor.Name, tt.wantName)
			}
		})
	}
}

func TestGetByIDs(t *testing.T) {
	tests := []struct {
		name    string
		ids     []types.AdvisorID
		wantLen int
	}{
		{"single ID", []types.AdvisorID{types.AdvisorCTO}, 1},
		{"multiple IDs", []types.AdvisorID{types.AdvisorCTO, types.AdvisorCISO}, 2},
		{"all core", []types.AdvisorID{types.AdvisorCTO, types.AdvisorCISO, types.AdvisorVPEng, types.AdvisorArchitect}, 4},
		{"mixed core and specialist", []types.AdvisorID{types.AdvisorCTO, types.AdvisorCFO}, 2},
		{"with unknown ID", []types.AdvisorID{types.AdvisorCTO, "unknown"}, 1},
		{"empty list", []types.AdvisorID{}, 0},
		{"all unknown", []types.AdvisorID{"unknown1", "unknown2"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetByIDs(tt.ids)

			if len(result) != tt.wantLen {
				t.Errorf("GetByIDs() returned %d advisors, want %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestSummonSpecialists(t *testing.T) {
	tests := []struct {
		name        string
		question    string
		wantIDs     []types.AdvisorID
		wantMinimum int
	}{
		{
			name:        "CFO keywords - budget",
			question:    "What's the budget impact?",
			wantIDs:     []types.AdvisorID{types.AdvisorCFO},
			wantMinimum: 1,
		},
		{
			name:        "CFO keywords - cost",
			question:    "How much will this cost?",
			wantIDs:     []types.AdvisorID{types.AdvisorCFO},
			wantMinimum: 1,
		},
		{
			name:        "CFO keywords - ROI",
			question:    "What's the ROI on this investment?",
			wantIDs:     []types.AdvisorID{types.AdvisorCFO},
			wantMinimum: 1,
		},
		{
			name:        "Product keywords - feature",
			question:    "Should we add this feature?",
			wantIDs:     []types.AdvisorID{types.AdvisorProduct},
			wantMinimum: 1,
		},
		{
			name:        "Product keywords - MVP",
			question:    "What's the MVP for this?",
			wantIDs:     []types.AdvisorID{types.AdvisorProduct},
			wantMinimum: 1,
		},
		{
			name:        "Product keywords - customers",
			question:    "What do customers want?",
			wantIDs:     []types.AdvisorID{types.AdvisorProduct},
			wantMinimum: 1,
		},
		{
			name:        "DevOps keywords - kubernetes",
			question:    "Should we adopt Kubernetes?",
			wantIDs:     []types.AdvisorID{types.AdvisorDevOps},
			wantMinimum: 1,
		},
		{
			name:        "DevOps keywords - docker",
			question:    "How should we containerize with Docker?",
			wantIDs:     []types.AdvisorID{types.AdvisorDevOps},
			wantMinimum: 1,
		},
		{
			name:        "DevOps keywords - AWS",
			question:    "What AWS services should we use?",
			wantIDs:     []types.AdvisorID{types.AdvisorDevOps},
			wantMinimum: 1,
		},
		{
			name:        "multiple specialists",
			question:    "What's the cost of deploying to kubernetes?",
			wantIDs:     []types.AdvisorID{types.AdvisorCFO, types.AdvisorDevOps},
			wantMinimum: 2,
		},
		{
			name:        "case insensitive - uppercase",
			question:    "KUBERNETES deployment BUDGET",
			wantIDs:     []types.AdvisorID{types.AdvisorCFO, types.AdvisorDevOps},
			wantMinimum: 2,
		},
		{
			name:        "no keywords match",
			question:    "Should we hire more engineers?",
			wantIDs:     []types.AdvisorID{},
			wantMinimum: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SummonSpecialists(tt.question)

			if len(result) < tt.wantMinimum {
				t.Errorf("SummonSpecialists(%q) returned %d specialists, want at least %d",
					tt.question, len(result), tt.wantMinimum)
			}

			// Verify expected IDs are present
			for _, wantID := range tt.wantIDs {
				found := false
				for _, specialist := range result {
					if specialist.Specialist.ID == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("SummonSpecialists(%q) should include %s", tt.question, wantID)
				}
			}
		})
	}
}

func TestSummonSpecialistsNoDuplicates(t *testing.T) {
	// Question with multiple keywords for same specialist
	question := "What's the budget and cost and ROI and investment?"

	result := SummonSpecialists(question)

	// Count CFO occurrences
	cfoCount := 0
	for _, specialist := range result {
		if specialist.Specialist.ID == types.AdvisorCFO {
			cfoCount++
		}
	}

	if cfoCount > 1 {
		t.Errorf("SummonSpecialists should not return duplicates, got %d CFO entries", cfoCount)
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"HELLO", "hello"},
		{"Hello World", "hello world"},
		{"already lowercase", "already lowercase"},
		{"MiXeD CaSe", "mixed case"},
		{"", ""},
		{"123ABC", "123abc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toLower(tt.input)
			if got != tt.want {
				t.Errorf("toLower(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", "xyz", false},
		{"hello", "hello", true},
		{"hello", "", true},
		{"", "", true},
		{"", "x", false},
		{"kubernetes", "k8s", false},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestAdvisorPersonaCompleteness(t *testing.T) {
	// Verify all advisors have required fields
	for _, advisor := range AllAdvisors() {
		t.Run(string(advisor.ID), func(t *testing.T) {
			if advisor.ID == "" {
				t.Error("advisor ID should not be empty")
			}
			if advisor.Name == "" {
				t.Error("advisor Name should not be empty")
			}
			if advisor.Role == "" {
				t.Error("advisor Role should not be empty")
			}
			if advisor.ThinkingStyle == "" {
				t.Error("advisor ThinkingStyle should not be empty")
			}
			if advisor.Background == "" {
				t.Error("advisor Background should not be empty")
			}
			if len(advisor.Priorities) == 0 {
				t.Error("advisor should have at least one priority")
			}
		})
	}
}
