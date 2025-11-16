// File: src/core/ai/providers/claude_api.go
package providers

import (
	"context"
	"fmt"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/ready-to-release/eac/src/core/ai"
)

// ClaudeAPI provider uses Anthropic API with API key authentication
//
// Intent: Call Claude API directly using API key (costs apply, more control).
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear separation: validation, API call, response extraction
//   - API key and model stored in struct
//   - Execute() method has single responsibility
//   - Error messages include context
//
// Easy to change:
//   - API client initialization is separate
//   - Can switch to different SDK easily
//   - Model and options are configurable
//   - No global state
//
// Hard to break:
//   - API key validation happens at construction
//   - Model validation happens at construction
//   - Context allows cancellation
//   - Errors wrapped with context
type ClaudeAPI struct {
	client anthropic.Client
	model  string
}

// NewClaudeAPI creates a Claude API provider
// Returns error if API key or model is empty (fail fast)
func NewClaudeAPI(apiKey, model string) (*ClaudeAPI, error) {
	// Validate required fields
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for claude-api provider")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required for claude-api provider")
	}

	// Create Anthropic client
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &ClaudeAPI{
		client: client,
		model:  model,
	}, nil
}

// Name returns "claude-api" for provider identification
func (p *ClaudeAPI) Name() string {
	return "claude-api"
}

// Execute sends input to Claude API and returns the response
func (p *ClaudeAPI) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	// Apply options with defaults
	options := &ai.ExecuteOptions{
		Model:       p.model,
		Temperature: 0.3,
		MaxTokens:   4000,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Create message request
	message, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model: anthropic.Model(options.Model),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(input)),
		},
		MaxTokens:   int64(options.MaxTokens),
		Temperature: anthropic.Float(options.Temperature),
	})

	if err != nil {
		return "", fmt.Errorf("claude API call failed: %w", err)
	}

	// Extract text from response
	if len(message.Content) == 0 {
		return "", fmt.Errorf("claude returned no content")
	}

	// Extract text from content blocks using AsText()
	var result string
	for _, block := range message.Content {
		textBlock := block.AsText()
		if textBlock.Text != "" {
			result += textBlock.Text
		}
	}

	if result == "" {
		return "", fmt.Errorf("claude returned non-text content")
	}

	return result, nil
}
