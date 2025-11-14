// File: src/core/ai/providers/claude_cli.go
package providers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/src/core/ai"
)

// ClaudeCLI provider uses Claude CLI tool with subscription authentication
//
// Intent: Invoke Claude via CLI tool using subscription credits (no API costs).
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Single Execute() method with clear flow
//   - Explicit command building with args
//   - Environment variable removal is visible and explicit
//   - Error messages include both stdout and stderr for debugging
//
// Easy to change:
//   - Model mapping is isolated in helper function
//   - Command args are built in clear steps
//   - Can add more CLI flags without changing structure
//   - removeAPIKeyFromEnv is a pure function
//
// Hard to break:
//   - Tests use exec.LookPath to skip if claude CLI unavailable
//   - Errors wrapped with context
//   - Both stdout and stderr captured for debugging
//   - API key removal is tested separately
type ClaudeCLI struct {
	defaultModel string
}

// NewClaudeCLI creates a Claude CLI provider
// Uses Claude Pro subscription for authentication (removes API key from environment)
func NewClaudeCLI() *ClaudeCLI {
	return &ClaudeCLI{
		defaultModel: "sonnet", // Default to sonnet for quality
	}
}

// Name returns "claude-cli" for provider identification
func (p *ClaudeCLI) Name() string {
	return "claude-cli"
}

// Execute sends input to Claude via CLI tool
//
// CRITICAL: Removes ANTHROPIC_API_KEY from environment to force subscription auth.
// This allows using Claude Pro credits instead of API credits.
func (p *ClaudeCLI) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	// Apply options with defaults
	options := ai.ApplyOptions(opts...)

	// Determine model to use
	model := p.defaultModel
	if options.Model != "" {
		model = options.Model
	}

	// Build command arguments
	args := []string{
		"--print",
	}

	// Add model if specified
	if model != "" {
		args = append(args, "--model", model)
	}

	// Create command
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Stdin = strings.NewReader(input)

	// CRITICAL: Remove ANTHROPIC_API_KEY to force subscription auth
	// See docs/reference/modules/src-commands/claude-constraints.md for rationale
	cmd.Env = removeAPIKeyFromEnv(os.Environ())

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	if err := cmd.Run(); err != nil {
		stderrText := stderr.String()
		stdoutText := stdout.String()

		return "", fmt.Errorf("claude CLI execution failed: %w\nStderr: %s\nStdout: %s",
			err, stderrText, stdoutText)
	}

	// Return output (trimmed)
	output := strings.TrimSpace(stdout.String())
	return output, nil
}

// removeAPIKeyFromEnv removes ANTHROPIC_API_KEY from environment variables
// This forces Claude CLI to use subscription auth instead of API key
//
// This is a pure function - it doesn't modify the input, just returns a filtered copy
func removeAPIKeyFromEnv(environ []string) []string {
	var filtered []string
	for _, env := range environ {
		if !strings.HasPrefix(env, "ANTHROPIC_API_KEY=") {
			filtered = append(filtered, env)
		}
	}
	return filtered
}
