// Package gemini provides an AI provider adapter for the Google Gemini API.
package gemini

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/generative-ai-go/genai"
	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/aiproviders"
	"google.golang.org/api/option"
)

// DefaultModel is the default model for the Gemini provider.
const DefaultModel = "gemini-1.5-pro"

func init() {
	aiproviders.Register("gemini", func(config *ai.ProviderConfig) (ai.Provider, error) {
		return New(config.AI.APIKey, config.AI.Model)
	})
}

// Provider implements ai.Provider using the Google Gemini API.
// The genai.Client is cached after first creation to reduce connection overhead.
type Provider struct {
	apiKey string
	model  string

	mu     sync.Mutex
	client *genai.Client
}

// New creates a Gemini provider.
// Returns error if API key or model is empty (fail fast).
func New(apiKey, model string) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for gemini provider")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required for gemini provider")
	}

	return &Provider{
		apiKey: apiKey,
		model:  model,
	}, nil
}

// Name returns "gemini" for provider identification.
func (p *Provider) Name() string {
	return "gemini"
}

// getOrCreateClient returns the cached genai.Client, creating it on first call.
func (p *Provider) getOrCreateClient(ctx context.Context) (*genai.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		return p.client, nil
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(p.apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}
	p.client = client
	return client, nil
}

// Close releases the cached client resources. Safe to call multiple times.
func (p *Provider) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		p.client.Close()
		p.client = nil
	}
}

// Execute runs a prompt through Gemini API.
func (p *Provider) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	options := ai.ApplyOptions(opts...)
	model := p.model
	if options.Model != "" {
		model = options.Model
	}

	client, err := p.getOrCreateClient(ctx)
	if err != nil {
		return "", err
	}

	genModel := client.GenerativeModel(model)
	genModel.SetTemperature(float32(options.Temperature))
	genModel.SetMaxOutputTokens(int32(options.MaxTokens))

	resp, err := genModel.GenerateContent(ctx, genai.Text(input))
	if err != nil {
		return "", fmt.Errorf("gemini API call failed: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("gemini returned no candidates")
	}

	if resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned empty content")
	}

	part := resp.Candidates[0].Content.Parts[0]
	if textPart, ok := part.(genai.Text); ok {
		return string(textPart), nil
	}

	return "", fmt.Errorf("gemini returned non-text content")
}

// Compile-time interface check.
var _ ai.Provider = (*Provider)(nil)
