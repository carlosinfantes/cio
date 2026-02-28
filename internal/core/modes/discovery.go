// Package modes implements the different interaction modes for the advisory board.
package modes

import (
	"context"

	"github.com/carlosinfantes/cio/internal/core/advisors"
	"github.com/carlosinfantes/cio/internal/core/llm"
	"github.com/carlosinfantes/cio/internal/types"
)

// DiscoveryResponse represents the facilitator's response in discovery mode.
type DiscoveryResponse struct {
	Content string
}

// Discovery executes a single turn in discovery mode with the facilitator.
func Discovery(ctx context.Context, client *llm.Client, session *types.DiscoverySession, userMessage string, projectCtx *types.CRFContext) (DiscoveryResponse, error) {
	facilitator := advisors.GetFacilitator()

	// Build system prompt
	systemPrompt := llm.BuildFacilitatorSystemPrompt(facilitator, projectCtx)

	// Determine if this is the first message
	isFirstMessage := len(session.Messages) == 0

	// Build user prompt with conversation history
	conversationHistory := ""
	if !isFirstMessage {
		conversationHistory = session.GetConversationText()
	}

	userPrompt := llm.BuildFacilitatorUserPrompt(conversationHistory, userMessage, isFirstMessage)

	// Query the LLM
	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    1024, // Shorter responses for discovery
	})
	if err != nil {
		return DiscoveryResponse{}, err
	}

	// Add messages to session
	if !isFirstMessage {
		session.AddMessage("user", userMessage)
	}
	session.AddMessage("facilitator", resp.Content)

	return DiscoveryResponse{
		Content: resp.Content,
	}, nil
}

// DiscoveryGreeting generates the initial facilitator greeting.
func DiscoveryGreeting(ctx context.Context, client *llm.Client, projectCtx *types.CRFContext) (string, error) {
	facilitator := advisors.GetFacilitator()

	systemPrompt := llm.BuildFacilitatorSystemPrompt(facilitator, projectCtx)
	userPrompt := llm.BuildFacilitatorUserPrompt("", "", true)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    256, // Very short for greeting
	})
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// PanelWithBrief executes a panel discussion using a brief from discovery.
func PanelWithBrief(ctx context.Context, client *llm.Client, brief *types.Brief, advisorPersonas []types.Persona, projectCtx *types.CRFContext, mode types.Mode) (types.ParsedResponse, error) {
	systemPrompt := llm.BuildSystemPrompt(advisorPersonas, projectCtx, mode)
	userPrompt := llm.BuildPanelPromptWithBrief(brief)

	resp, err := client.Query(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
	})
	if err != nil {
		return types.ParsedResponse{}, err
	}

	return llm.ParseResponse(resp.Content, advisorPersonas), nil
}
