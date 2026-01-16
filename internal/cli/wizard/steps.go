package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/carlosinfantes/cto-advisory-board/internal/config"
	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

// stepAPIKey prompts for the OpenRouter API key.
func (w *Wizard) stepAPIKey() error {
	fmt.Println("─── Step 1: API Configuration ───")
	fmt.Println()
	fmt.Println("The advisory board uses OpenRouter to access AI models.")
	fmt.Println("Get your API key at: https://openrouter.ai/keys")
	fmt.Println()

	// Check for existing environment variable
	if envKey := os.Getenv("OPENROUTER_API_KEY"); envKey != "" {
		fmt.Printf("Found OPENROUTER_API_KEY in environment: %s\n", maskAPIKey(envKey))
		use, _ := w.promptWithDefault("Use this key?", "yes")
		if strings.ToLower(use) == "yes" || strings.ToLower(use) == "y" {
			w.config.APIKey = envKey
			fmt.Println()
			return nil
		}
	}

	apiKey, err := w.promptRequired("Enter your OpenRouter API key")
	if err != nil {
		return err
	}
	w.config.APIKey = apiKey
	fmt.Println()
	return nil
}

// stepDomainSelection prompts for domain/plugin selection.
func (w *Wizard) stepDomainSelection() error {
	fmt.Println("─── Step 2: Domain Selection ───")
	fmt.Println()
	fmt.Println("Advisory boards are domain-specific. Choose your domain:")
	fmt.Println()

	// For now, show available domains (will integrate with registry later)
	domains := []string{
		"cto-advisory - AI-powered executive committee for CTOs",
		"legal-advisory - Legal counsel for business decisions",
		"medical-advisory - Healthcare practice advisory",
		"custom - Create a custom domain",
	}

	idx, err := w.promptSelect("", domains, 0)
	if err != nil {
		return err
	}

	switch idx {
	case 0:
		w.config.ActiveDomain = "cto-advisory"
	case 1:
		w.config.ActiveDomain = "legal-advisory"
	case 2:
		w.config.ActiveDomain = "medical-advisory"
	case 3:
		domain, err := w.promptRequired("Enter custom domain name")
		if err != nil {
			return err
		}
		w.config.ActiveDomain = domain
	}

	fmt.Println()
	fmt.Printf("Note: Run 'cto plugin install %s' after setup to download the domain package.\n", w.config.ActiveDomain)
	fmt.Println()
	return nil
}

// stepModelPreference prompts for LLM model selection.
func (w *Wizard) stepModelPreference() error {
	fmt.Println("─── Step 3: Model Preference ───")
	fmt.Println()
	fmt.Println("Choose your preferred AI model:")
	fmt.Println()

	models := []string{
		"anthropic/claude-3.5-sonnet - Balanced (recommended)",
		"anthropic/claude-3-opus - Most capable, higher cost",
		"openai/gpt-4o - OpenAI alternative",
		"openai/gpt-4o-mini - Fast and economical",
		"custom - Enter a custom model name",
	}

	idx, err := w.promptSelect("", models, 0)
	if err != nil {
		return err
	}

	switch idx {
	case 0:
		w.config.Model = "anthropic/claude-3.5-sonnet"
	case 1:
		w.config.Model = "anthropic/claude-3-opus"
	case 2:
		w.config.Model = "openai/gpt-4o"
	case 3:
		w.config.Model = "openai/gpt-4o-mini"
	case 4:
		model, err := w.promptRequired("Enter model name (e.g., anthropic/claude-3-haiku)")
		if err != nil {
			return err
		}
		w.config.Model = model
	}

	fmt.Println()
	return nil
}

// stepCognitiveStyle prompts for interaction style preference.
func (w *Wizard) stepCognitiveStyle() error {
	fmt.Println("─── Step 4: Interaction Style ───")
	fmt.Println()
	fmt.Println("How would you like to interact with the advisory board?")
	fmt.Println()

	styles := []string{
		"Discovery First - Start with facilitator to clarify your challenge",
		"Panel Direct - Go straight to the advisory panel",
		"Socratic - Deep exploration through clarifying questions",
	}

	idx, err := w.promptSelect("", styles, 0)
	if err != nil {
		return err
	}

	switch idx {
	case 0:
		w.config.StartInDiscovery = true
		w.config.DefaultMode = types.ModePanel
	case 1:
		w.config.StartInDiscovery = false
		w.config.DefaultMode = types.ModePanel
	case 2:
		w.config.StartInDiscovery = false
		w.config.DefaultMode = types.ModeSocratic
	}

	fmt.Println()
	return nil
}

// stepOrganizationContext prompts for organization details.
func (w *Wizard) stepOrganizationContext() error {
	fmt.Println("─── Step 5: Organization Context ───")
	fmt.Println()
	fmt.Println("Tell us about your organization (helps advisors give relevant advice):")
	fmt.Println()

	// Organization name
	orgName, err := w.promptRequired("Organization name")
	if err != nil {
		return err
	}

	// Industry
	industry, err := w.promptWithDefault("Industry", "technology")
	if err != nil {
		return err
	}

	// Company stage
	stages := []string{
		"pre-seed",
		"seed",
		"series-a",
		"series-b+",
		"growth",
		"public",
	}
	fmt.Println("\nCompany stage:")
	stageIdx, err := w.promptSelect("", stages, 1)
	if err != nil {
		return err
	}
	stage := stages[stageIdx]

	// Team size
	teamSize, err := w.promptWithDefault("Team size (engineers)", "10")
	if err != nil {
		return err
	}

	// Primary language
	language, err := w.promptWithDefault("Primary programming language", "typescript")
	if err != nil {
		return err
	}

	// Cloud provider
	clouds := []string{
		"aws",
		"gcp",
		"azure",
		"multi-cloud",
		"on-premise",
		"none",
	}
	fmt.Println("\nCloud provider:")
	cloudIdx, err := w.promptSelect("", clouds, 0)
	if err != nil {
		return err
	}
	cloud := clouds[cloudIdx]

	// Save organization context
	if err := w.saveOrganizationContext(orgName, industry, stage, teamSize, language, cloud); err != nil {
		return fmt.Errorf("saving organization context: %w", err)
	}

	fmt.Println()
	return nil
}

// saveOrganizationContext creates the CRF organization file.
func (w *Wizard) saveOrganizationContext(name, industry, stage, teamSize, language, cloud string) error {
	contextDir, err := config.GetContextDir()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(contextDir, 0755); err != nil {
		return err
	}

	// Create organization CRF document
	org := types.CRFDocument{
		CRFVersion: w.config.CRFVersion,
		Entity: types.CRFEntity{
			ID:          fmt.Sprintf("org-%d", time.Now().Unix()),
			Type:        types.CRFEntityOrganization,
			Name:        name,
			Description: fmt.Sprintf("%s in the %s industry", name, industry),
			Attributes: map[string]interface{}{
				"org_type":       "company",
				"industry":       industry,
				"stage":          stage,
				"size":           teamSize,
				"primary_lang":   language,
				"cloud_provider": cloud,
			},
			Provenance: types.Provenance{
				Source:    "wizard",
				CreatedAt: time.Now(),
			},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(org)
	if err != nil {
		return err
	}

	// Write file
	filePath := filepath.Join(contextDir, "organization.yaml")
	return os.WriteFile(filePath, data, 0644)
}
