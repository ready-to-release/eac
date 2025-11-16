// File: src/core/ai/providers/gemini.go
// Intent: Call Google Gemini API using official SDK
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

	"github.com/google/generative-ai-go/genai"
	"github.com/ready-to-release/eac/src/core/ai"
	"google.golang.org/api/option"
)

// Gemini provider uses Google Gemini API with API key authentication
type Gemini struct {
	apiKey string
	model  string
}

// NewGemini creates a Gemini provider
// Returns error if API key or model is empty (fail fast)
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

// Name returns the provider name
func (p *Gemini) Name() string {
	return "gemini"
}

// Execute runs a prompt through Gemini API
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

	// Create client
	client, err := genai.NewClient(ctx, option.WithAPIKey(p.apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create gemini client: %w", err)
	}
	defer client.Close()

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
