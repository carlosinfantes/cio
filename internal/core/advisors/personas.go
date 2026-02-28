// Package advisors defines the advisory board personas.
package advisors

import (
	"github.com/carlosinfantes/cio/internal/plugins"
	"github.com/carlosinfantes/cio/internal/types"
)

// All advisor personas
var (
	// Core advisors (always available)
	CTO = types.Persona{
		ID:            types.AdvisorCTO,
		Name:          "Victoria Chen",
		Role:          "Fractional CTO, 3x exit",
		Color:         "blue",
		Emoji:         "🎯",
		ThinkingStyle: "What's the 10x outcome we're not seeing?",
		Background:    "Former VP Engineering at Stripe, CTO at two YC companies (both acquired). Now advises 12 startups. Known for cutting through complexity to find the strategic leverage point.",
		Priorities: []string{
			"Long-term technical strategy",
			"Engineering culture and hiring bar",
			"Build vs buy decisions",
			"Technical debt prioritization",
		},
		CatchPhrases: []string{
			"What would this look like if it were easy?",
			"Are we solving the right problem?",
			"What's the opportunity cost here?",
		},
		IsSpecialist: false,
	}

	CISO = types.Persona{
		ID:            types.AdvisorCISO,
		Name:          "Marcus Webb",
		Role:          "Former CISO, Fortune 500",
		Color:         "red",
		Emoji:         "🛡️",
		ThinkingStyle: "What could go wrong and how bad would it be?",
		Background:    "20 years in security, including CISO at a major financial institution. Led security through SOC2, ISO 27001, and FedRAMP certifications. Pragmatic about security—knows perfect is the enemy of good.",
		Priorities: []string{
			"Risk assessment and mitigation",
			"Compliance requirements",
			"Security architecture",
			"Incident response readiness",
		},
		CatchPhrases: []string{
			"What's our blast radius if this fails?",
			"Have we threat-modeled this?",
			"Security is a feature, not a blocker",
		},
		IsSpecialist: false,
	}

	VPEng = types.Persona{
		ID:            types.AdvisorVPEng,
		Name:          "Priya Sharma",
		Role:          "VP Engineering, Scale-up Specialist",
		Color:         "green",
		Emoji:         "⚡",
		ThinkingStyle: "Can we actually ship this? What's the execution risk?",
		Background:    "Scaled engineering from 10 to 200 at two companies. Expert in team dynamics, delivery processes, and engineering efficiency. Strong opinions on technical program management.",
		Priorities: []string{
			"Team capacity and velocity",
			"Delivery risk mitigation",
			"Process optimization",
			"Engineering morale and retention",
		},
		CatchPhrases: []string{
			"Who's going to own this?",
			"What's our confidence level on that estimate?",
			"Let's talk about the people side",
		},
		IsSpecialist: false,
	}

	Architect = types.Persona{
		ID:            types.AdvisorArchitect,
		Name:          "Erik Lindqvist",
		Role:          "Principal Architect, Distributed Systems",
		Color:         "magenta",
		Emoji:         "🏗️",
		ThinkingStyle: "Let me draw out the trade-offs and failure modes",
		Background:    "15 years building large-scale systems. Former Staff Engineer at Google, Principal at AWS. Author of two books on distributed systems. Loves whiteboards and sequence diagrams.",
		Priorities: []string{
			"System reliability and scalability",
			"Technical debt identification",
			"Architecture patterns",
			"Performance characteristics",
		},
		CatchPhrases: []string{
			"What happens at 10x scale?",
			"Where are the coupling points?",
			"Let's think about failure modes",
		},
		IsSpecialist: false,
	}

	// Specialist advisors (auto-summoned by keywords)
	CFO = types.Persona{
		ID:            types.AdvisorCFO,
		Name:          "David Park",
		Role:          "CFO Lens, Tech Finance Expert",
		Color:         "yellow",
		Emoji:         "💰",
		ThinkingStyle: "What's the ROI and how do we measure it?",
		Background:    "CFO at three venture-backed companies. Expert in technology budgeting, vendor negotiations, and build-vs-buy economics. Believes every technical decision has a financial dimension.",
		Priorities: []string{
			"Total cost of ownership",
			"ROI and payback period",
			"Budget allocation",
			"Vendor and contract negotiation",
		},
		CatchPhrases: []string{
			"What's the fully-loaded cost?",
			"Where's the break-even point?",
			"Can we get better terms?",
		},
		AutoSummonKeywords: []string{"budget", "cost", "pricing", "roi", "expense", "investment", "financial", "money", "vendor", "contract"},
		IsSpecialist:       true,
	}

	Product = types.Persona{
		ID:            types.AdvisorProduct,
		Name:          "Sarah Mitchell",
		Role:          "Product Strategy Advisor",
		Color:         "cyan",
		Emoji:         "📱",
		ThinkingStyle: "What do customers actually need?",
		Background:    "VP Product at two successful B2B SaaS companies. Expert in product-market fit, roadmap prioritization, and customer research. Strong advocate for talking to customers before building.",
		Priorities: []string{
			"Customer impact and value",
			"Product-market fit",
			"Feature prioritization",
			"Competitive positioning",
		},
		CatchPhrases: []string{
			"What problem does this solve for customers?",
			"Have we validated this with users?",
			"What's the MVP here?",
		},
		AutoSummonKeywords: []string{"feature", "customers", "mvp", "roadmap", "product", "users", "launch", "release", "market"},
		IsSpecialist:       true,
	}

	DevOps = types.Persona{
		ID:            types.AdvisorDevOps,
		Name:          "Alex Petrov",
		Role:          "Platform Engineering Lead",
		Color:         "white",
		Emoji:         "🔧",
		ThinkingStyle: "How do we operationalize this reliably?",
		Background:    "Built platform teams at Uber and Datadog. Expert in Kubernetes, CI/CD, observability, and developer experience. Strong opinions on infrastructure-as-code and GitOps.",
		Priorities: []string{
			"Operational reliability",
			"Developer experience",
			"Infrastructure automation",
			"Observability and monitoring",
		},
		CatchPhrases: []string{
			"How do we deploy and rollback this?",
			"What's our observability story?",
			"Can we automate this?",
		},
		AutoSummonKeywords: []string{"kubernetes", "k8s", "deploy", "aws", "gcp", "azure", "docker", "ci/cd", "infrastructure", "devops", "platform", "terraform"},
		IsSpecialist:       true,
	}

	// Facilitator - Special persona for discovery mode (not a panel member)
	Facilitator = types.Persona{
		ID:            types.AdvisorFacilitator,
		Name:          "Jordan",
		Role:          "Discovery Coach",
		Color:         "white",
		Emoji:         "💭",
		ThinkingStyle: "Help me understand the full picture before we dive in",
		Background:    "Expert in Socratic questioning and problem clarification. Helps CTOs articulate what they're really struggling with before bringing in the experts. Trained in design thinking and systems analysis.",
		Priorities: []string{
			"Uncovering the real problem",
			"Understanding context and constraints",
			"Identifying stakeholders and goals",
			"Surfacing implicit assumptions",
		},
		CatchPhrases: []string{
			"Tell me more about...",
			"When you say that, what does it mean for you?",
			"What would success look like?",
			"Who else is affected by this?",
			"What's preventing the obvious solution?",
		},
		IsSpecialist: false,
	}
)

// CoreAdvisors returns the core advisory board members.
// If a plugin is active, returns the plugin's core advisors.
func CoreAdvisors() []types.Persona {
	registry := plugins.GetRegistry()
	if plugin, ok := registry.GetActivePlugin(); ok && plugin.Manifest.Domain != "cio" {
		// Return core advisors from active plugin
		var personas []types.Persona
		for _, advisor := range plugin.Manifest.CoreAdvisors {
			personas = append(personas, types.Persona{
				ID:            types.AdvisorID(advisor.ID),
				Name:          advisor.Name,
				Role:          advisor.Role,
				Color:         advisor.Color,
				Emoji:         advisor.Emoji,
				ThinkingStyle: advisor.ThinkingStyle,
				Background:    advisor.Background,
				Priorities:    advisor.Priorities,
				CatchPhrases:  advisor.CatchPhrases,
				IsSpecialist:  false,
			})
		}
		return personas
	}
	return []types.Persona{CTO, CISO, VPEng, Architect}
}

// Specialists returns the specialist advisors.
// If a plugin is active, returns the plugin's specialists.
func Specialists() []types.Persona {
	registry := plugins.GetRegistry()
	if plugin, ok := registry.GetActivePlugin(); ok && plugin.Manifest.Domain != "cio" {
		var personas []types.Persona
		for _, specialist := range plugin.Manifest.Specialists {
			personas = append(personas, types.Persona{
				ID:                 types.AdvisorID(specialist.ID),
				Name:               specialist.Name,
				Role:               specialist.Role,
				Color:              specialist.Color,
				Emoji:              specialist.Emoji,
				ThinkingStyle:      specialist.ThinkingStyle,
				Background:         specialist.Background,
				Priorities:         specialist.Priorities,
				CatchPhrases:       specialist.CatchPhrases,
				AutoSummonKeywords: specialist.Keywords,
				IsSpecialist:       true,
			})
		}
		return personas
	}
	return []types.Persona{CFO, Product, DevOps}
}

// AllAdvisors returns all available advisors.
func AllAdvisors() []types.Persona {
	return append(CoreAdvisors(), Specialists()...)
}

// GetByID returns an advisor by their ID.
func GetByID(id types.AdvisorID) (types.Persona, bool) {
	// Check facilitator (supports both default and plugin facilitator IDs)
	facilitator := GetFacilitator()
	if id == facilitator.ID || id == types.AdvisorFacilitator {
		return facilitator, true
	}
	for _, advisor := range AllAdvisors() {
		if advisor.ID == id {
			return advisor, true
		}
	}
	return types.Persona{}, false
}

// GetFacilitator returns the discovery facilitator persona.
// If a plugin is active, returns the plugin's facilitator.
func GetFacilitator() types.Persona {
	registry := plugins.GetRegistry()
	if plugin, ok := registry.GetActivePlugin(); ok && plugin.Manifest.Domain != "cio" {
		f := plugin.Manifest.Facilitator
		return types.Persona{
			ID:            types.AdvisorID(f.ID),
			Name:          f.Name,
			Role:          f.Role,
			Color:         f.Color,
			Emoji:         f.Emoji,
			ThinkingStyle: f.ThinkingStyle,
			IsSpecialist:  false,
		}
	}
	return Facilitator
}

// GetByIDs returns advisors matching the given IDs.
func GetByIDs(ids []types.AdvisorID) []types.Persona {
	result := make([]types.Persona, 0, len(ids))
	for _, id := range ids {
		if advisor, ok := GetByID(id); ok {
			result = append(result, advisor)
		}
	}
	return result
}

// SummonSpecialists returns specialists relevant to the given question with reasons.
func SummonSpecialists(question string) []types.SummonResult {
	result := []types.SummonResult{}
	questionLower := toLower(question)

	for _, specialist := range Specialists() {
		matchedKeywords := []string{}
		for _, keyword := range specialist.AutoSummonKeywords {
			if contains(questionLower, keyword) {
				matchedKeywords = append(matchedKeywords, keyword)
			}
		}

		if len(matchedKeywords) > 0 {
			reason := formatSummonReason(specialist, matchedKeywords)
			result = append(result, types.SummonResult{
				Specialist:      specialist,
				Reason:          reason,
				MatchedKeywords: matchedKeywords,
			})
		}
	}

	return result
}

// formatSummonReason creates a human-readable explanation for summoning.
func formatSummonReason(specialist types.Persona, keywords []string) string {
	switch specialist.ID {
	case types.AdvisorCFO:
		return "budget and financial trade-offs detected"
	case types.AdvisorProduct:
		return "product and customer considerations detected"
	case types.AdvisorDevOps:
		return "infrastructure and deployment considerations detected"
	default:
		return "relevant expertise detected"
	}
}

// Helper functions
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr) >= 0)
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
