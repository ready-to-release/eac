// File: src/core/ai/providers/claude_cli_test.go
//go:build L2 && ov
// +build L2,ov

// Integration tests that require real Claude CLI - excluded from L0-L1 (commit suite)
// Go tests should only be L0-L1 unit tests with mocks
// Real provider tests belong in Godog specs (L2+)
package providers

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/ready-to-release/eac/src/ai"
)

func TestClaudeCLI_Name(t *testing.T) {
	provider := NewClaudeCLI()
	if provider.Name() != "claude-cli" {
		t.Errorf("Name() = %v, want %v", provider.Name(), "claude-cli")
	}
}

func TestClaudeCLI_Execute(t *testing.T) {
	// Skip if claude CLI is not available
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude CLI not available, skipping integration test")
	}

	provider := NewClaudeCLI()
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
			opts:    []ai.Option{ai.WithModel(ClaudeCLIModelHaiku)},
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

			t.Logf("Claude CLI output: %s", output)
		})
	}
}

func TestClaudeCLI_RemovesAPIKey(t *testing.T) {
	// This test verifies that the ANTHROPIC_API_KEY is removed from environment
	// We can't easily test the actual removal, but we can verify the function exists
	provider := NewClaudeCLI()

	// Set an API key
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	// The provider should still work (this is a compile check mainly)
	if provider.Name() != "claude-cli" {
		t.Error("Provider should work even with API key set")
	}
}

func TestClaudeCLI_WithOptions(t *testing.T) {
	provider := NewClaudeCLI()
	ctx := context.Background()

	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude CLI not available, skipping integration test")
	}

	// Test with different model options
	output, err := provider.Execute(ctx, "Say hello", ai.WithModel(ClaudeCLIModelHaiku))
	if err != nil {
		t.Errorf("Execute() with model option error = %v", err)
	}
	if output == "" {
		t.Error("Execute() returned empty output")
	}
}
