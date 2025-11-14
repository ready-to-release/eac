// File: src/core/ai/providers/openai.go
// Intent: Call OpenAI API using official SDK for GPT models
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear validation at construction
//   - Single Execute() method with options
//   - API client initialization separated
//   - Error messages with context
//
// Easy to change:
//   - SDK version can be upgraded
//   - Model and options are configurable
//   - No global state
//   - Can switch models at runtime via options
//
// Hard to break:
//   - API key validation at construction
//   - Model validation at construction
//   - Context allows cancellation
//   - Errors wrapped with context

package providers

import (
	"context"
	"fmt"

	"github.com/ready-to-release/eac/src/core/ai"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAI provider uses OpenAI API with API key authentication
type OpenAI struct {
	client *openai.Client
	model  string
}

// NewOpenAI creates an OpenAI provider
// Returns error if API key or model is empty (fail fast)
func NewOpenAI(apiKey, model string) (*OpenAI, error) {
	// Validate required fields
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for openai provider")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required for openai provider")
	}

	// Create OpenAI client
	client := openai.NewClient(apiKey)

	return &OpenAI{
		client: client,
		model:  model,
	}, nil
}

// Name returns the provider name
func (p *OpenAI) Name() string {
	return "openai"
}

// Execute runs a prompt through OpenAI API
func (p *OpenAI) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	// Apply options
	options := &ai.ExecuteOptions{
		Model:       p.model,
		Temperature: 0.3,
		MaxTokens:   4000,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Build request
	req := openai.ChatCompletionRequest{
		Model: options.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: input,
			},
		},
		Temperature: float32(options.Temperature),
		MaxTokens:   options.MaxTokens,
	}

	// Call API
	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("openai API call failed: %w", err)
	}

	// Extract response
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}

	return resp.Choices[0].Message.Content, nil
}
