// File: src/core/ai/providers/gemini_test.go
package providers

import (
	"context"
	"os"
	"testing"

	"github.com/ready-to-release/eac/src/core/ai"
)

func TestGemini_Name(t *testing.T) {
	provider, _ := NewGemini("test-key", "gemini-pro")
	if provider.Name() != "gemini" {
		t.Errorf("Name() = %v, want gemini", provider.Name())
	}
}

func TestGemini_Execute(t *testing.T) {
	// Skip if no API key (integration test)
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping Gemini integration test: GOOGLE_API_KEY not set")
	}

	provider, err := NewGemini(apiKey, "gemini-pro")
	if err != nil {
		t.Fatalf("NewGemini() error = %v", err)
	}

	ctx := context.Background()
	response, err := provider.Execute(ctx, "Say 'test' and nothing else")

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if response == "" {
		t.Errorf("Execute() returned empty response")
	}

	t.Logf("Gemini response: %s", response)
}

func TestGemini_ValidationError(t *testing.T) {
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
			model:     "gemini-pro",
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
			model:   "gemini-pro",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewGemini(tt.apiKey, tt.model)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewGemini() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errString != "" && !containsSubstr(err.Error(), tt.errString) {
					t.Errorf("NewGemini() error = %v, want error containing %v", err, tt.errString)
				}
			}

			if !tt.wantErr && provider == nil {
				t.Errorf("NewGemini() returned nil provider for valid input")
			}
		})
	}
}

func TestGemini_WithOptions(t *testing.T) {
	// Skip if no API key
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping Gemini integration test: GOOGLE_API_KEY not set")
	}

	provider, err := NewGemini(apiKey, "gemini-pro")
	if err != nil {
		t.Fatalf("NewGemini() error = %v", err)
	}

	ctx := context.Background()

	// Test with custom options
	response, err := provider.Execute(ctx, "Say 'options test'",
		ai.WithModel("gemini-1.5-pro"),
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
func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || containsSubstr(s[1:], substr)))
}
