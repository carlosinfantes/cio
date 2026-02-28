// Package llm provides the OpenRouter API client for the CIO - Chief Intelligence Officer.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// OpenRouter API endpoint
	openRouterURL = "https://openrouter.ai/api/v1/chat/completions"
)

// Common errors
var (
	ErrNoAPIKey     = errors.New("no API key configured")
	ErrInvalidKey   = errors.New("invalid API key")
	ErrRateLimited  = errors.New("rate limited by API")
	ErrServerError  = errors.New("API server error")
	ErrNetworkError = errors.New("network error")
)

// Request represents a request to the LLM.
type Request struct {
	SystemPrompt string
	UserPrompt   string
	MaxTokens    int
}

// Response represents a response from the LLM.
type Response struct {
	Content      string
	InputTokens  int
	OutputTokens int
	Model        string
}

// Client wraps the OpenRouter API client.
type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewClient creates a new LLM client.
func NewClient(apiKey, model string) (*Client, error) {
	if apiKey == "" {
		return nil, ErrNoAPIKey
	}

	return &Client{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

// OpenRouter request/response types
type openRouterRequest struct {
	Model    string              `json:"model"`
	Messages []openRouterMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens,omitempty"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// Query sends a request to the LLM and returns the response.
func (c *Client) Query(ctx context.Context, req Request) (*Response, error) {
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	// Retry with exponential backoff
	var lastErr error
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 9 * time.Second}

	for attempt := 0; attempt <= len(delays); attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delays[attempt-1]):
			}
		}

		resp, err := c.doQuery(ctx, req, maxTokens)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if errors.Is(err, ErrInvalidKey) || errors.Is(err, ErrNoAPIKey) {
			return nil, err
		}

		// Don't retry if context is cancelled
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("after %d retries: %w", len(delays), lastErr)
}

func (c *Client) doQuery(ctx context.Context, req Request, maxTokens int) (*Response, error) {
	// Build messages
	messages := []openRouterMessage{}

	// Add system message if provided
	if req.SystemPrompt != "" {
		messages = append(messages, openRouterMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	// Add user message
	messages = append(messages, openRouterMessage{
		Role:    "user",
		Content: req.UserPrompt,
	})

	// Build request body
	reqBody := openRouterRequest{
		Model:     c.model,
		Messages:  messages,
		MaxTokens: maxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", openRouterURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/carlosinfantes/cio")
	httpReq.Header.Set("X-Title", "CIO - Chief Intelligence Officer")

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, classifyError(err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, classifyHTTPError(resp.StatusCode, string(body))
	}

	// Parse response
	var orResp openRouterResponse
	if err := json.Unmarshal(body, &orResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	// Check for API error in response
	if orResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", orResp.Error.Message)
	}

	// Extract content
	if len(orResp.Choices) == 0 {
		return nil, errors.New("no response from model")
	}

	return &Response{
		Content:      orResp.Choices[0].Message.Content,
		InputTokens:  orResp.Usage.PromptTokens,
		OutputTokens: orResp.Usage.CompletionTokens,
		Model:        orResp.Model,
	}, nil
}

// ValidateAPIKey checks if the API key is valid by making a minimal request.
func (c *Client) ValidateAPIKey(ctx context.Context) error {
	// Use a cheap/fast model for validation
	validationModel := "openai/gpt-3.5-turbo"

	messages := []openRouterMessage{
		{Role: "user", Content: "Hi"},
	}

	reqBody := openRouterRequest{
		Model:     validationModel,
		Messages:  messages,
		MaxTokens: 5,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", openRouterURL, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/carlosinfantes/cio")
	httpReq.Header.Set("X-Title", "CIO - Chief Intelligence Officer")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return classifyError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrInvalidKey
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return classifyHTTPError(resp.StatusCode, string(body))
	}

	return nil
}

func classifyHTTPError(statusCode int, body string) error {
	switch statusCode {
	case 401:
		return ErrInvalidKey
	case 429:
		return ErrRateLimited
	case 500, 502, 503:
		return ErrServerError
	default:
		return fmt.Errorf("HTTP %d: %s", statusCode, body)
	}
}

func classifyError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	if containsString(errStr, "connection") || containsString(errStr, "timeout") || containsString(errStr, "network") {
		return ErrNetworkError
	}

	return fmt.Errorf("request error: %w", err)
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findString(s, substr) >= 0
}

func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			tc := substr[j]
			// Case-insensitive comparison
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if tc >= 'A' && tc <= 'Z' {
				tc += 32
			}
			if sc != tc {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
