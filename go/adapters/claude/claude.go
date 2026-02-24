// Package claude provides an AI provider adapter for the Anthropic Claude API.
package claude

import (
	"context"
	"fmt"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/aiproviders"
)

// DefaultModel is the default model for the Claude API provider.
const DefaultModel = "claude-3-haiku-20240307"

func init() {
	aiproviders.Register("claude-api", func(config *ai.ProviderConfig) (ai.Provider, error) {
		return New(config.AI.APIKey, config.AI.Model)
	})
}

// Provider implements ai.Provider using the Anthropic API.
type Provider struct {
	client anthropic.Client
	model  string
}

// New creates a Claude API provider.
// Returns error if API key or model is empty (fail fast).
func New(apiKey, model string) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for claude-api provider")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required for claude-api provider")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Provider{
		client: client,
		model:  model,
	}, nil
}

// Name returns "claude-api" for provider identification.
func (p *Provider) Name() string {
	return "claude-api"
}

// Execute sends input to Claude API and returns the response.
func (p *Provider) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	options := ai.ApplyOptions(opts...)
	model := p.model
	if options.Model != "" {
		model = options.Model
	}

	message, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model: anthropic.Model(model),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(input)),
		},
		MaxTokens:   int64(options.MaxTokens),
		Temperature: anthropic.Float(options.Temperature),
	})
	if err != nil {
		return "", fmt.Errorf("claude API call failed: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("claude returned no content")
	}

	var result string
	for i := range message.Content {
		block := &message.Content[i]
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

// Compile-time interface check.
var _ ai.Provider = (*Provider)(nil)
