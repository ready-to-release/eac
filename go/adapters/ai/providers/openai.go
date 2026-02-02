// File: go/eac/adapters/ai/providers/openai.go
package providers

import (
	"context"
	"fmt"

	"github.com/ready-to-release/eac/go/adapters/ai"
	openai "github.com/sashabaranov/go-openai"
)

// DefaultOpenAIModel is the default model for OpenAI provider
// Change this constant when upgrading to a newer model version.
const DefaultOpenAIModel = "gpt-4-turbo"

// OpenAI provider uses OpenAI API with API key authentication.
type OpenAI struct {
	client *openai.Client
	model  string
}

// NewOpenAI creates an OpenAI provider
// Returns error if API key or model is empty (fail fast).
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

// Name returns the provider name.
func (p *OpenAI) Name() string {
	return "openai"
}

// Execute runs a prompt through OpenAI API.
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
