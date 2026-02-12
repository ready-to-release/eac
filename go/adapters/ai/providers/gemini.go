// File: go/eac/adapters/ai/providers/gemini.go
package providers

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"github.com/ready-to-release/eac/go/adapters/ai"
	"google.golang.org/api/option"
)

// DefaultGeminiModel is the default model for Gemini provider
// Change this constant when upgrading to a newer model version.
const DefaultGeminiModel = "gemini-1.5-pro"

// Gemini provider uses Google Gemini API with API key authentication.
// The genai.Client is cached after first creation to reduce connection overhead
// for repeated invocations.
type Gemini struct {
	apiKey string
	model  string

	mu     sync.Mutex
	client *genai.Client
}

// NewGemini creates a Gemini provider
// Returns error if API key or model is empty (fail fast).
func NewGemini(apiKey, model string) (*Gemini, error) {
	// Validate required fields
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for gemini provider")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required for gemini provider")
	}

	return &Gemini{
		apiKey: apiKey,
		model:  model,
	}, nil
}

// Name returns the provider name.
func (p *Gemini) Name() string {
	return "gemini"
}

// getOrCreateClient returns the cached genai.Client, creating it on first call.
// Thread-safe via mutex.
func (p *Gemini) getOrCreateClient(ctx context.Context) (*genai.Client, error) {
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
func (p *Gemini) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		p.client.Close()
		p.client = nil
	}
}

// Execute runs a prompt through Gemini API.
func (p *Gemini) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	// Apply options
	options := &ai.ExecuteOptions{
		Model:       p.model,
		Temperature: 0.3,
		MaxTokens:   4000,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Get or create cached client
	client, err := p.getOrCreateClient(ctx)
	if err != nil {
		return "", err
	}

	// Get model
	model := client.GenerativeModel(options.Model)

	// Configure generation parameters
	model.SetTemperature(float32(options.Temperature))
	model.SetMaxOutputTokens(int32(options.MaxTokens))

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(input))
	if err != nil {
		return "", fmt.Errorf("gemini API call failed: %w", err)
	}

	// Extract response
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("gemini returned no candidates")
	}

	if resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned empty content")
	}

	// Extract text from first part
	part := resp.Candidates[0].Content.Parts[0]
	if textPart, ok := part.(genai.Text); ok {
		return string(textPart), nil
	}

	return "", fmt.Errorf("gemini returned non-text content")
}
