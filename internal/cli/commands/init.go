// Package commands implements the init wizard for the CIO - Chief Intelligence Officer.
package commands

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/carlosinfantes/cio/internal/cli/output"
	"github.com/carlosinfantes/cio/internal/config"
	ctxLoader "github.com/carlosinfantes/cio/internal/core/context"
	"github.com/carlosinfantes/cio/internal/core/llm"
	"github.com/carlosinfantes/cio/internal/types"
)

func init() {
	rootCmd.AddCommand(newInitCmd())
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new CIO - Chief Intelligence Officer project",
		Long:  "Run the setup wizard to configure your project context and API key.",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("51")).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("51")).
		Padding(0, 2).
		Margin(1, 0)

	fmt.Println(headerStyle.Render("CTO ADVISORY BOARD — PROJECT SETUP"))

	// Check if already initialized
	if config.IsInitialized() {
		var reinit bool
		prompt := &survey.Confirm{
			Message: "Project already initialized. Reinitialize?",
			Default: false,
		}
		if err := survey.AskOne(prompt, &reinit); err != nil {
			return err
		}
		if !reinit {
			output.PrintInfo("Setup cancelled.")
			return nil
		}
	}

	// Step 1: API Key
	apiKey, err := stepAPIKey()
	if err != nil {
		return err
	}

	// Step 2: Domain Selection (NEW)
	domain, err := stepDomainSelection()
	if err != nil {
		return err
	}

	// Step 3: Model Preference (NEW)
	model, err := stepModelPreference()
	if err != nil {
		return err
	}

	// Step 4: Company Basics
	company, err := stepCompanyBasics()
	if err != nil {
		return err
	}

	// Step 5: Team Structure
	team, err := stepTeamStructure()
	if err != nil {
		return err
	}

	// Step 6: Tech Stack
	techStack, err := stepTechStack()
	if err != nil {
		return err
	}

	// Step 7: Constraints
	constraints, err := stepConstraints(company.Stage)
	if err != nil {
		return err
	}

	// Step 8: Challenges
	challenges, err := stepChallenges()
	if err != nil {
		return err
	}

	// Save everything
	output.PrintInfo("Saving configuration...")

	if err := config.EnsureAdvisoryDir(); err != nil {
		output.PrintError(fmt.Sprintf("Creating directories: %v", err))
		return err
	}

	// Save config
	cfg := types.DefaultConfig()
	cfg.APIKey = apiKey
	cfg.ActiveDomain = domain
	cfg.Model = model
	if err := config.Save(cfg); err != nil {
		output.PrintError(fmt.Sprintf("Saving config: %v", err))
		return err
	}

	// Save CRF context files
	now := time.Now()

	// Organization entity (company)
	if err := saveOrganizationCRF(company, constraints, challenges, now); err != nil {
		output.PrintError(fmt.Sprintf("Saving organization context: %v", err))
		return err
	}

	// Team entity
	if err := saveTeamCRF(team, company.Name, now); err != nil {
		output.PrintError(fmt.Sprintf("Saving team context: %v", err))
		return err
	}

	// System entity (tech stack)
	if err := saveSystemCRF(techStack, company.Name, now); err != nil {
		output.PrintError(fmt.Sprintf("Saving system context: %v", err))
		return err
	}

	// Constraint facts
	if err := saveConstraintsCRF(constraints, company.Name, now); err != nil {
		output.PrintError(fmt.Sprintf("Saving constraints context: %v", err))
		return err
	}

	// Success message
	fmt.Println()
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("46")).
		Padding(0, 1).
		Margin(1, 0)

	successMsg := fmt.Sprintf(`✓ Created:
  .cio/config.yaml
  .cio/context/organization.yaml
  .cio/context/teams.yaml
  .cio/context/systems.yaml
  .cio/context/facts.yaml
  .cio/plugins/installed/
  .cio/plugins/custom/

Domain: %s
Model:  %s

Next steps:
  1. Install domain: cio plugin install %s
  2. Ask a question: cio "your question"`, domain, model, domain)

	fmt.Println(successStyle.Render(successMsg))

	return nil
}

// ----------------------------------------------------------------------------
// Wizard Steps
// ----------------------------------------------------------------------------

func stepAPIKey() (string, error) {
	stepHeader("Step 1 of 8", "API Configuration")

	// Check for existing key
	existingCfg, _ := config.Load()
	if existingCfg.APIKey != "" {
		var useExisting bool
		prompt := &survey.Confirm{
			Message: "Existing API key found. Keep it?",
			Default: true,
		}
		if err := survey.AskOne(prompt, &useExisting); err != nil {
			return "", err
		}
		if useExisting {
			output.PrintSuccess("Using existing API key")
			return existingCfg.APIKey, nil
		}
	}

	// Get new key
	var apiKey string
	prompt := &survey.Password{
		Message: "Enter your OpenRouter API key:",
	}
	if err := survey.AskOne(prompt, &apiKey, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	// Validate API key (uses GPT-3.5 for cheap validation via OpenRouter)
	output.PrintInfo("Validating API key...")

	client, err := llm.NewClient(apiKey, "openai/gpt-3.5-turbo")
	if err != nil {
		output.PrintError("Invalid API key format")
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.ValidateAPIKey(ctx); err != nil {
		if err == llm.ErrInvalidKey {
			output.PrintError("Invalid API key. Please check and try again.")
			fmt.Println("Get your key at: https://openrouter.ai/keys")
			os.Exit(1)
		}
		// Other errors - warn but continue
		output.PrintInfo(fmt.Sprintf("Could not verify key (%v). Proceeding anyway.", err))
	} else {
		output.PrintSuccess("API key validated")
	}

	return apiKey, nil
}

func stepDomainSelection() (string, error) {
	stepHeader("Step 2 of 8", "Domain Selection")

	fmt.Println("Advisory boards are domain-specific. Choose your domain:")
	fmt.Println("(You can install more domains later with 'cto plugin install')")
	fmt.Println()

	var domain string
	prompt := &survey.Select{
		Message: "Select domain:",
		Options: []string{
			"cio - Executive committee for CTOs",
			"legal-advisory - Legal counsel for business decisions",
			"medical-advisory - Healthcare practice advisory",
			"custom - Create or install a custom domain",
		},
		Default: "cio - Executive committee for CTOs",
	}
	if err := survey.AskOne(prompt, &domain); err != nil {
		return "", err
	}

	// Extract domain name from selection
	switch {
	case strings.HasPrefix(domain, "cio"):
		return "cio", nil
	case strings.HasPrefix(domain, "legal-advisory"):
		return "legal-advisory", nil
	case strings.HasPrefix(domain, "medical-advisory"):
		return "medical-advisory", nil
	default:
		// Custom domain
		var customDomain string
		customPrompt := &survey.Input{
			Message: "Enter custom domain name:",
		}
		if err := survey.AskOne(customPrompt, &customDomain, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}
		return customDomain, nil
	}
}

func stepModelPreference() (string, error) {
	stepHeader("Step 3 of 8", "Model Preference")

	fmt.Println("Choose your preferred AI model:")
	fmt.Println()

	var model string
	prompt := &survey.Select{
		Message: "Select model:",
		Options: []string{
			"anthropic/claude-3.5-sonnet - Balanced (recommended)",
			"anthropic/claude-3-opus - Most capable, higher cost",
			"openai/gpt-4o - OpenAI alternative",
			"openai/gpt-4o-mini - Fast and economical",
			"custom - Enter a custom model name",
		},
		Default: "anthropic/claude-3.5-sonnet - Balanced (recommended)",
	}
	if err := survey.AskOne(prompt, &model); err != nil {
		return "", err
	}

	// Extract model name from selection
	switch {
	case strings.HasPrefix(model, "anthropic/claude-3.5-sonnet"):
		return "anthropic/claude-3.5-sonnet", nil
	case strings.HasPrefix(model, "anthropic/claude-3-opus"):
		return "anthropic/claude-3-opus", nil
	case strings.HasPrefix(model, "openai/gpt-4o-mini"):
		return "openai/gpt-4o-mini", nil
	case strings.HasPrefix(model, "openai/gpt-4o"):
		return "openai/gpt-4o", nil
	default:
		// Custom model
		var customModel string
		customPrompt := &survey.Input{
			Message: "Enter model name (e.g., anthropic/claude-3-haiku):",
		}
		if err := survey.AskOne(customPrompt, &customModel, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}
		return customModel, nil
	}
}

type companyInfo struct {
	Name          string
	Industry      string
	Stage         string
	Founded       int
	BusinessModel string
}

func stepCompanyBasics() (*companyInfo, error) {
	stepHeader("Step 4 of 8", "Company Basics")

	var answers struct {
		Name          string
		Industry      string
		Stage         string
		Founded       string
		BusinessModel string
	}

	questions := []*survey.Question{
		{
			Name:     "Name",
			Prompt:   &survey.Input{Message: "Company name:"},
			Validate: survey.Required,
		},
		{
			Name:   "Industry",
			Prompt: &survey.Input{Message: "Industry:", Default: "technology"},
		},
		{
			Name: "Stage",
			Prompt: &survey.Select{
				Message: "Company stage:",
				Options: []string{
					"pre-seed",
					"seed",
					"series-a",
					"series-b",
					"series-c",
					"growth",
					"public",
					"bootstrapped",
				},
				Default: "seed",
			},
		},
		{
			Name:   "Founded",
			Prompt: &survey.Input{Message: "Year founded:", Default: fmt.Sprintf("%d", time.Now().Year())},
		},
		{
			Name:   "BusinessModel",
			Prompt: &survey.Input{Message: "Business model (brief):", Default: "B2B SaaS"},
		},
	}

	if err := survey.Ask(questions, &answers); err != nil {
		return nil, err
	}

	founded, _ := strconv.Atoi(answers.Founded)
	if founded == 0 {
		founded = time.Now().Year()
	}

	return &companyInfo{
		Name:          answers.Name,
		Industry:      answers.Industry,
		Stage:         answers.Stage,
		Founded:       founded,
		BusinessModel: answers.BusinessModel,
	}, nil
}

type teamInfo struct {
	Total         int
	Breakdown     map[string]int
	UnfilledRoles []string
}

func stepTeamStructure() (*teamInfo, error) {
	stepHeader("Step 5 of 8", "Team Structure")

	var total string
	prompt := &survey.Input{
		Message: "Total engineers:",
		Default: "10",
	}
	if err := survey.AskOne(prompt, &total); err != nil {
		return nil, err
	}

	totalNum, _ := strconv.Atoi(total)
	if totalNum <= 0 {
		totalNum = 10
	}

	// Optional breakdown
	var wantBreakdown bool
	confirmPrompt := &survey.Confirm{
		Message: "Add team breakdown? (optional)",
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &wantBreakdown); err != nil {
		return nil, err
	}

	breakdown := make(map[string]int)
	if wantBreakdown {
		var answers struct {
			Backend  string
			Frontend string
			Platform string
			Mobile   string
		}

		questions := []*survey.Question{
			{Name: "Backend", Prompt: &survey.Input{Message: "  Backend engineers:", Default: "0"}},
			{Name: "Frontend", Prompt: &survey.Input{Message: "  Frontend engineers:", Default: "0"}},
			{Name: "Platform", Prompt: &survey.Input{Message: "  Platform/DevOps:", Default: "0"}},
			{Name: "Mobile", Prompt: &survey.Input{Message: "  Mobile:", Default: "0"}},
		}

		if err := survey.Ask(questions, &answers); err != nil {
			return nil, err
		}

		breakdown["backend"], _ = strconv.Atoi(answers.Backend)
		breakdown["frontend"], _ = strconv.Atoi(answers.Frontend)
		breakdown["platform"], _ = strconv.Atoi(answers.Platform)
		breakdown["mobile"], _ = strconv.Atoi(answers.Mobile)
	}

	// Unfilled roles
	var rolesStr string
	rolesPrompt := &survey.Input{
		Message: "Key unfilled roles (comma-separated):",
		Default: "",
	}
	if err := survey.AskOne(rolesPrompt, &rolesStr); err != nil {
		return nil, err
	}

	var unfilledRoles []string
	if rolesStr != "" {
		for _, role := range strings.Split(rolesStr, ",") {
			role = strings.TrimSpace(role)
			if role != "" {
				unfilledRoles = append(unfilledRoles, role)
			}
		}
	}

	return &teamInfo{
		Total:         totalNum,
		Breakdown:     breakdown,
		UnfilledRoles: unfilledRoles,
	}, nil
}

type techStackInfo struct {
	Primary        string
	Languages      []string
	Cloud          string
	Infrastructure []string
}

func stepTechStack() (*techStackInfo, error) {
	stepHeader("Step 6 of 8", "Tech Stack")

	var answers struct {
		Primary    string
		OtherLangs string
		Cloud      string
		Infra      string
	}

	questions := []*survey.Question{
		{
			Name:   "Primary",
			Prompt: &survey.Input{Message: "Primary language:", Default: "typescript"},
		},
		{
			Name:   "OtherLangs",
			Prompt: &survey.Input{Message: "Other languages (comma-separated):", Default: ""},
		},
		{
			Name: "Cloud",
			Prompt: &survey.Select{
				Message: "Cloud provider:",
				Options: []string{"aws", "gcp", "azure", "other", "none"},
				Default: "aws",
			},
		},
		{
			Name:   "Infra",
			Prompt: &survey.Input{Message: "Key infrastructure (comma-separated):", Default: ""},
		},
	}

	if err := survey.Ask(questions, &answers); err != nil {
		return nil, err
	}

	languages := []string{answers.Primary}
	if answers.OtherLangs != "" {
		for _, lang := range strings.Split(answers.OtherLangs, ",") {
			lang = strings.TrimSpace(lang)
			if lang != "" {
				languages = append(languages, lang)
			}
		}
	}

	var infra []string
	if answers.Infra != "" {
		for _, i := range strings.Split(answers.Infra, ",") {
			i = strings.TrimSpace(i)
			if i != "" {
				infra = append(infra, i)
			}
		}
	}

	return &techStackInfo{
		Primary:        answers.Primary,
		Languages:      languages,
		Cloud:          answers.Cloud,
		Infrastructure: infra,
	}, nil
}

type constraintsInfo struct {
	Compliance []string
	Runway     int
	Deadline   string
}

func stepConstraints(stage string) (*constraintsInfo, error) {
	stepHeader("Step 7 of 8", "Constraints")

	var complianceStr string
	compliancePrompt := &survey.Input{
		Message: "Compliance requirements (comma-separated):",
		Default: "",
	}
	if err := survey.AskOne(compliancePrompt, &complianceStr); err != nil {
		return nil, err
	}

	var compliance []string
	if complianceStr != "" {
		for _, c := range strings.Split(complianceStr, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				compliance = append(compliance, c)
			}
		}
	}

	var runway int
	// Only ask for runway if not public/bootstrapped
	if stage != "public" && stage != "bootstrapped" {
		var runwayStr string
		runwayPrompt := &survey.Input{
			Message: "Runway (months):",
			Default: "18",
		}
		if err := survey.AskOne(runwayPrompt, &runwayStr); err != nil {
			return nil, err
		}
		runway, _ = strconv.Atoi(runwayStr)
	}

	var deadline string
	deadlinePrompt := &survey.Input{
		Message: "Critical deadlines (optional):",
		Default: "",
	}
	if err := survey.AskOne(deadlinePrompt, &deadline); err != nil {
		return nil, err
	}

	return &constraintsInfo{
		Compliance: compliance,
		Runway:     runway,
		Deadline:   deadline,
	}, nil
}

func stepChallenges() (string, error) {
	stepHeader("Step 8 of 8", "Current Challenges (Optional)")

	var challenges string
	prompt := &survey.Input{
		Message: "What's your biggest technical challenge right now?",
		Default: "",
	}
	if err := survey.AskOne(prompt, &challenges); err != nil {
		return "", err
	}

	return challenges, nil
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

func stepHeader(step, title string) {
	stepStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	fmt.Println()
	fmt.Println(stepStyle.Render(fmt.Sprintf("%s: %s", step, title)))
	fmt.Println(dividerStyle.Render(strings.Repeat("─", 40)))
}

// ----------------------------------------------------------------------------
// CRF Entity Creation
// ----------------------------------------------------------------------------

func saveOrganizationCRF(company *companyInfo, constraints *constraintsInfo, challenges string, now time.Time) error {
	// Map stage to size
	sizeMap := map[string]string{
		"pre-seed":     "startup",
		"seed":         "startup",
		"series-a":     "small",
		"series-b":     "medium",
		"series-c":     "medium",
		"growth":       "large",
		"public":       "enterprise",
		"bootstrapped": "small",
	}
	size := sizeMap[company.Stage]
	if size == "" {
		size = "startup"
	}

	// Build attributes
	attrs := map[string]interface{}{
		"org_type":       "company",
		"size":           size,
		"industry":       company.Industry,
		"stage":          company.Stage,
		"founded":        company.Founded,
		"business_model": company.BusinessModel,
	}
	if len(constraints.Compliance) > 0 {
		attrs["compliance_frameworks"] = constraints.Compliance
	}
	if challenges != "" {
		attrs["current_challenge"] = challenges
	}

	doc := ctxLoader.CreateOrganizationEntity(
		fmt.Sprintf("org-%s", strings.ToLower(strings.ReplaceAll(company.Name, " ", "-"))),
		company.Name,
		fmt.Sprintf("%s company in the %s industry", company.Stage, company.Industry),
		attrs,
	)

	return ctxLoader.SaveCRFDocument(doc, "organization.yaml")
}

func saveTeamCRF(team *teamInfo, companyName string, now time.Time) error {
	// Build skills list from breakdown
	var skills []string
	for role, count := range team.Breakdown {
		if count > 0 {
			skills = append(skills, role)
		}
	}

	attrs := map[string]interface{}{
		"org_type":  "team",
		"headcount": team.Total,
	}
	if len(skills) > 0 {
		attrs["skills"] = skills
	}
	if len(team.UnfilledRoles) > 0 {
		attrs["unfilled_roles"] = team.UnfilledRoles
	}

	// Add structure breakdown
	if len(team.Breakdown) > 0 {
		attrs["structure"] = team.Breakdown
	}

	doc := ctxLoader.CreateTeamEntity(
		"team-engineering",
		"Engineering Team",
		fmt.Sprintf("Engineering team at %s", companyName),
		team.Total,
		skills,
	)

	// Merge additional attributes
	for k, v := range attrs {
		doc.Entity.Attributes[k] = v
	}

	return ctxLoader.SaveCRFDocument(doc, "teams.yaml")
}

func saveSystemCRF(tech *techStackInfo, companyName string, now time.Time) error {
	attrs := map[string]interface{}{
		"system_type":      "platform",
		"status":           "production",
		"hosting":          tech.Cloud,
		"primary_language": tech.Primary,
		"languages":        tech.Languages,
	}
	if len(tech.Infrastructure) > 0 {
		attrs["technology_stack"] = tech.Infrastructure
	}

	doc := ctxLoader.CreateSystemEntity(
		"system-main",
		"Main Platform",
		fmt.Sprintf("Primary technology platform at %s", companyName),
		attrs,
	)

	return ctxLoader.SaveCRFDocument(doc, "systems.yaml")
}

func saveConstraintsCRF(constraints *constraintsInfo, companyName string, now time.Time) error {
	contextDir, err := config.GetContextDir()
	if err != nil {
		return err
	}

	var docs []types.CRFDocument

	// Runway fact
	if constraints.Runway > 0 {
		doc := ctxLoader.CreateFactEntity(
			"fact-runway",
			"Financial Runway",
			fmt.Sprintf("Current runway: %d months", constraints.Runway),
			map[string]interface{}{
				"fact_type": "constraint",
				"value":     constraints.Runway,
				"unit":      "months",
			},
		)
		docs = append(docs, *doc)
	}

	// Deadline fact
	if constraints.Deadline != "" {
		doc := ctxLoader.CreateFactEntity(
			"fact-deadline",
			"Critical Deadline",
			constraints.Deadline,
			map[string]interface{}{
				"fact_type": "timeline",
				"value":     constraints.Deadline,
			},
		)
		docs = append(docs, *doc)
	}

	// Compliance facts
	for _, comp := range constraints.Compliance {
		doc := ctxLoader.CreateFactEntity(
			fmt.Sprintf("fact-compliance-%s", strings.ToLower(strings.ReplaceAll(comp, " ", "-"))),
			fmt.Sprintf("%s Compliance", comp),
			fmt.Sprintf("Compliance requirement: %s", comp),
			map[string]interface{}{
				"fact_type": "constraint",
				"value":     comp,
			},
		)
		docs = append(docs, *doc)
	}

	// If no constraints, create an empty facts file placeholder
	if len(docs) == 0 {
		return nil
	}

	// Save all constraint documents to the facts.yaml file
	for i := range docs {
		filename := "facts.yaml"
		if i > 0 {
			filename = fmt.Sprintf("facts_%d.yaml", i)
		}
		if err := ctxLoader.SaveCRFDocument(&docs[i], filename); err != nil {
			return err
		}
	}

	_ = contextDir // Used in error path
	return nil
}
