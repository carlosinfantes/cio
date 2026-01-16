// Package modes implements the different interaction modes for the advisory board.
package modes

import (
	"context"

	"github.com/carlosinfantes/cto-advisory-board/internal/core/llm"
	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

// Panel executes a panel discussion with the advisory board.
func Panel(ctx context.Context, client *llm.Client, question string, advisors []types.Persona, projectCtx *types.CRFContext) (types.ParsedResponse, error) {
	// Build prompts
	systemPrompt := llm.BuildSystemPrompt(advisors, projectCtx, types.ModePanel)
	userPrompt := llm.BuildUserPrompt(question)

	// Query the LLM
	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
	})
	if err != nil {
		return types.ParsedResponse{}, err
	}

	// Parse the response
	parsed := llm.ParseResponse(resp.Content, advisors)

	return parsed, nil
}

// Socratic executes a Socratic dialogue mode.
func Socratic(ctx context.Context, client *llm.Client, question string, advisors []types.Persona, projectCtx *types.CRFContext) (types.ParsedResponse, error) {
	systemPrompt := llm.BuildSystemPrompt(advisors, projectCtx, types.ModeSocratic)
	userPrompt := llm.BuildUserPrompt(question)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
	})
	if err != nil {
		return types.ParsedResponse{}, err
	}

	return llm.ParseResponse(resp.Content, advisors), nil
}

// Advocate executes devil's advocate mode.
func Advocate(ctx context.Context, client *llm.Client, question string, advisors []types.Persona, projectCtx *types.CRFContext) (types.ParsedResponse, error) {
	systemPrompt := llm.BuildSystemPrompt(advisors, projectCtx, types.ModeAdvocate)
	userPrompt := llm.BuildUserPrompt(question)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
	})
	if err != nil {
		return types.ParsedResponse{}, err
	}

	return llm.ParseResponse(resp.Content, advisors), nil
}

// Framework executes decision framework mode.
func Framework(ctx context.Context, client *llm.Client, question string, advisors []types.Persona, projectCtx *types.CRFContext) (types.ParsedResponse, error) {
	systemPrompt := llm.BuildSystemPrompt(advisors, projectCtx, types.ModeFramework)
	userPrompt := llm.BuildUserPrompt(question)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
	})
	if err != nil {
		return types.ParsedResponse{}, err
	}

	return llm.ParseResponse(resp.Content, advisors), nil
}

// PanelWithContext executes a panel discussion with session context.
func PanelWithContext(ctx context.Context, client *llm.Client, question string, advisors []types.Persona, projectCtx *types.CRFContext, sessionContext string) (types.ParsedResponse, error) {
	systemPrompt := llm.BuildSystemPrompt(advisors, projectCtx, types.ModePanel)
	userPrompt := llm.BuildUserPromptWithContext(question, sessionContext)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
	})
	if err != nil {
		return types.ParsedResponse{}, err
	}

	return llm.ParseResponse(resp.Content, advisors), nil
}

// SocraticWithContext executes Socratic mode with session context.
func SocraticWithContext(ctx context.Context, client *llm.Client, question string, advisors []types.Persona, projectCtx *types.CRFContext, sessionContext string) (types.ParsedResponse, error) {
	systemPrompt := llm.BuildSystemPrompt(advisors, projectCtx, types.ModeSocratic)
	userPrompt := llm.BuildUserPromptWithContext(question, sessionContext)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
	})
	if err != nil {
		return types.ParsedResponse{}, err
	}

	return llm.ParseResponse(resp.Content, advisors), nil
}

// AdvocateWithContext executes devil's advocate mode with session context.
func AdvocateWithContext(ctx context.Context, client *llm.Client, question string, advisors []types.Persona, projectCtx *types.CRFContext, sessionContext string) (types.ParsedResponse, error) {
	systemPrompt := llm.BuildSystemPrompt(advisors, projectCtx, types.ModeAdvocate)
	userPrompt := llm.BuildUserPromptWithContext(question, sessionContext)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
	})
	if err != nil {
		return types.ParsedResponse{}, err
	}

	return llm.ParseResponse(resp.Content, advisors), nil
}

// FrameworkWithContext executes decision framework mode with session context.
func FrameworkWithContext(ctx context.Context, client *llm.Client, question string, advisors []types.Persona, projectCtx *types.CRFContext, sessionContext string) (types.ParsedResponse, error) {
	systemPrompt := llm.BuildSystemPrompt(advisors, projectCtx, types.ModeFramework)
	userPrompt := llm.BuildUserPromptWithContext(question, sessionContext)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
	})
	if err != nil {
		return types.ParsedResponse{}, err
	}

	return llm.ParseResponse(resp.Content, advisors), nil
}

// SocraticQuestions generates clarifying questions for the Socratic flow.
func SocraticQuestions(ctx context.Context, client *llm.Client, question string, projectCtx *types.CRFContext) ([]string, error) {
	systemPrompt := llm.BuildSocraticQuestionsPrompt(projectCtx)
	userPrompt := llm.BuildUserPrompt(question)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    1024,
	})
	if err != nil {
		return nil, err
	}

	return llm.ParseSocraticQuestions(resp.Content), nil
}

// SocraticPanelWithAnswers executes panel discussion with enriched context from Socratic Q&A.
func SocraticPanelWithAnswers(ctx context.Context, client *llm.Client, state *types.SocraticState, advisors []types.Persona, projectCtx *types.CRFContext) (types.ParsedResponse, error) {
	systemPrompt := llm.BuildSystemPrompt(advisors, projectCtx, types.ModePanel)
	enrichedContext := state.GetEnrichedContext()
	userPrompt := llm.BuildUserPromptWithContext(state.Question, enrichedContext)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
	})
	if err != nil {
		return types.ParsedResponse{}, err
	}

	return llm.ParseResponse(resp.Content, advisors), nil
}

// FrameworkCriteria generates evaluation criteria for a decision framework.
func FrameworkCriteria(ctx context.Context, client *llm.Client, question string, options []string, projectCtx *types.CRFContext) ([]types.FrameworkCriterion, error) {
	systemPrompt := llm.BuildFrameworkCriteriaPrompt(question, options, projectCtx)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   "Generate evaluation criteria for this decision.",
		MaxTokens:    1024,
	})
	if err != nil {
		return nil, err
	}

	return llm.ParseFrameworkCriteria(resp.Content), nil
}

// FrameworkScoring scores options against criteria.
func FrameworkScoring(ctx context.Context, client *llm.Client, state *types.FrameworkState, projectCtx *types.CRFContext) error {
	systemPrompt := llm.BuildFrameworkScoringPrompt(state, projectCtx)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   "Score each option against each criterion, then provide your recommendation.",
		MaxTokens:    2048,
	})
	if err != nil {
		return err
	}

	llm.ParseFrameworkScores(resp.Content, state)
	return nil
}

// SummarizeLongQuestion summarizes a question that exceeds the threshold.
func SummarizeLongQuestion(ctx context.Context, client *llm.Client, question string) (string, error) {
	systemPrompt := llm.BuildSummarizeQuestionPrompt()
	userPrompt := llm.BuildSummarizeQuestionUserPrompt(question)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    512,
	})
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// IsLongQuestion checks if a question exceeds the summarization threshold.
func IsLongQuestion(question string) bool {
	return len(question) > llm.LongQuestionThreshold
}
