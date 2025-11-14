// File: src/core/ai/provider_test.go
package ai

import (
	"context"
	"testing"
)

// TestProviderInterface verifies that all providers implement the Provider interface
// This is a compile-time check that will fail if providers don't implement the interface
func TestProviderInterface(t *testing.T) {
	var _ Provider = &MockProvider{}
	// Additional providers will be added here as they are implemented:
	// var _ Provider = &ClaudeAPI{}
	// var _ Provider = &ClaudeCLI{}
	// var _ Provider = &OpenAI{}
}

// TestProviderExecuteContract tests the basic contract that all providers must follow
func TestProviderExecuteContract(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		input    string
		opts     []Option
		wantErr  bool
	}{
		{
			name:     "mock provider executes successfully",
			provider: NewMockProvider("mock-response"),
			input:    "test input",
			opts:     []Option{},
			wantErr:  false,
		},
		{
			name:     "mock provider with model option",
			provider: NewMockProvider("response"),
			input:    "input",
			opts:     []Option{WithModel("test-model")},
			wantErr:  false,
		},
		{
			name:     "mock provider with multiple options",
			provider: NewMockProvider("response"),
			input:    "input",
			opts: []Option{
				WithModel("test-model"),
				WithTemperature(0.7),
				WithMaxTokens(1000),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			output, err := tt.provider.Execute(ctx, tt.input, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && output == "" {
				t.Error("Execute() returned empty output when expecting non-empty")
			}
		})
	}
}

// TestProviderName verifies that providers return their name correctly
func TestProviderName(t *testing.T) {
	provider := NewMockProvider("test")
	if provider.Name() != "mock" {
		t.Errorf("Name() = %v, want %v", provider.Name(), "mock")
	}
}
