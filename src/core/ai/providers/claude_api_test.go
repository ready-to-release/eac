// File: src/core/ai/providers/claude_api_test.go
package providers

import (
	"context"
	"os"
	"testing"

	"github.com/ready-to-release/eac/src/core/ai"
)

func TestClaudeAPI_Name(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = "test-key" // Use dummy key for name test
	}

	provider, err := NewClaudeAPI(apiKey, "claude-3-haiku-20240307")
	if err != nil {
		t.Fatalf("NewClaudeAPI() error = %v", err)
	}

	if provider.Name() != "claude-api" {
		t.Errorf("Name() = %v, want %v", provider.Name(), "claude-api")
	}
}

func TestClaudeAPI_Execute(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	provider, err := NewClaudeAPI(apiKey, "claude-3-haiku-20240307")
	if err != nil {
		t.Fatalf("NewClaudeAPI() error = %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		opts    []ai.Option
		wantErr bool
	}{
		{
			name:    "simple prompt execution",
			input:   "Say 'test' and nothing else.",
			opts:    []ai.Option{},
			wantErr: false,
		},
		{
			name:    "execution with model option",
			input:   "Say 'hello'",
			opts:    []ai.Option{ai.WithModel("claude-3-haiku-20240307")},
			wantErr: false,
		},
		{
			name:    "execution with temperature option",
			input:   "Say 'world'",
			opts:    []ai.Option{ai.WithTemperature(0.5)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := provider.Execute(ctx, tt.input, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && output == "" {
				t.Error("Execute() returned empty output")
			}

			t.Logf("Claude API output: %s", output)
		})
	}
}

func TestClaudeAPI_ValidationError(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		model   string
		wantErr bool
	}{
		{
			name:    "empty API key returns error",
			apiKey:  "",
			model:   "claude-3-haiku-20240307",
			wantErr: true,
		},
		{
			name:    "empty model returns error",
			apiKey:  "test-key",
			model:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClaudeAPI(tt.apiKey, tt.model)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewClaudeAPI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
