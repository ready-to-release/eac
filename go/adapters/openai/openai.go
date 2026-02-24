// Package openai provides an AI provider adapter for the OpenAI API.
package openai

import (
	"context"
	"fmt"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/aiproviders"
	goopenai "github.com/sashabaranov/go-openai"
)

// DefaultModel is the default model for the OpenAI provider.
const DefaultModel = "gpt-4-turbo"

func init() {
	aiproviders.Register("openai", func(config *ai.ProviderConfig) (ai.Provider, error) {
		return New(config.AI.APIKey, config.AI.Model)
	})
}

// Provider implements ai.Provider using the OpenAI API.
type Provider struct {
	client *goopenai.Client
	model  string
}

// New creates an OpenAI provider.
// Returns error if API key or model is empty (fail fast).
func New(apiKey, model string) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for openai provider")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required for openai provider")
	}

	client := goopenai.NewClient(apiKey)

	return &Provider{
		client: client,
		model:  model,
	}, nil
}

// Name returns "openai" for provider identification.
func (p *Provider) Name() string {
	return "openai"
}

// Execute runs a prompt through OpenAI API.
func (p *Provider) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	options := ai.ApplyOptions(opts...)
	model := p.model
	if options.Model != "" {
		model = options.Model
	}

	req := goopenai.ChatCompletionRequest{
		Model: model,
		Messages: []goopenai.ChatCompletionMessage{
			{
				Role:    goopenai.ChatMessageRoleUser,
				Content: input,
			},
		},
		Temperature: float32(options.Temperature),
		MaxTokens:   options.MaxTokens,
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("openai API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}

	return resp.Choices[0].Message.Content, nil
}

// Compile-time interface check.
var _ ai.Provider = (*Provider)(nil)
