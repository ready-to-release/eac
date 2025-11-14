// File: src/core/ai/providers/openai_test.go
package providers

import (
	"context"
	"os"
	"testing"

	"github.com/ready-to-release/eac/src/core/ai"
)

func TestOpenAI_Name(t *testing.T) {
	provider, _ := NewOpenAI("test-key", "gpt-4")
	if provider.Name() != "openai" {
		t.Errorf("Name() = %v, want openai", provider.Name())
	}
}

func TestOpenAI_Execute(t *testing.T) {
	// Skip if no API key (integration test)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenAI integration test: OPENAI_API_KEY not set")
	}

	provider, err := NewOpenAI(apiKey, "gpt-4")
	if err != nil {
		t.Fatalf("NewOpenAI() error = %v", err)
	}

	ctx := context.Background()
	response, err := provider.Execute(ctx, "Say 'test' and nothing else")

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if response == "" {
		t.Errorf("Execute() returned empty response")
	}

	t.Logf("OpenAI response: %s", response)
}

func TestOpenAI_ValidationError(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		model     string
		wantErr   bool
		errString string
	}{
		{
			name:      "empty API key returns error",
			apiKey:    "",
			model:     "gpt-4",
			wantErr:   true,
			errString: "API key is required",
		},
		{
			name:      "empty model returns error",
			apiKey:    "test-key",
			model:     "",
			wantErr:   true,
			errString: "model is required",
		},
		{
			name:    "valid parameters",
			apiKey:  "test-key",
			model:   "gpt-4",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOpenAI(tt.apiKey, tt.model)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewOpenAI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errString != "" && !contains(err.Error(), tt.errString) {
					t.Errorf("NewOpenAI() error = %v, want error containing %v", err, tt.errString)
				}
			}

			if !tt.wantErr && provider == nil {
				t.Errorf("NewOpenAI() returned nil provider for valid input")
			}
		})
	}
}

func TestOpenAI_WithOptions(t *testing.T) {
	// Skip if no API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenAI integration test: OPENAI_API_KEY not set")
	}

	provider, err := NewOpenAI(apiKey, "gpt-4")
	if err != nil {
		t.Fatalf("NewOpenAI() error = %v", err)
	}

	ctx := context.Background()

	// Test with custom options
	response, err := provider.Execute(ctx, "Say 'options test'",
		ai.WithModel("gpt-3.5-turbo"),
		ai.WithTemperature(0.5),
		ai.WithMaxTokens(100),
	)

	if err != nil {
		t.Errorf("Execute() with options error = %v, want nil", err)
	}

	if response == "" {
		t.Errorf("Execute() returned empty response")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}
