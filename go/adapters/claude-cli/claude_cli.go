// Package claudecli provides an AI provider adapter for the Claude CLI tool.
package claudecli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/aiproviders"
)

// Claude CLI model names (using full model IDs for consistency across providers).
const (
	ModelHaiku  = "claude-3-haiku-20240307"
	ModelSonnet = "claude-3-5-sonnet-20240620"
	ModelOpus   = "claude-3-opus-20240229"
)

// DefaultModel is the default model for the Claude CLI provider.
const DefaultModel = ModelHaiku

func init() {
	aiproviders.Register("claude-cli", func(config *ai.ProviderConfig) (ai.Provider, error) {
		return New(), nil
	})
}

// Provider implements ai.Provider using the Claude CLI tool.
type Provider struct {
	defaultModel string
}

// New creates a Claude CLI provider.
// Uses Claude Pro subscription for authentication (removes API key from environment).
func New() *Provider {
	return &Provider{
		defaultModel: DefaultModel,
	}
}

// Name returns "claude-cli" for provider identification.
func (p *Provider) Name() string {
	return "claude-cli"
}

// Execute sends input to Claude via CLI tool.
//
// CRITICAL: Removes ANTHROPIC_API_KEY from environment to force subscription auth.
// This allows using Claude Pro credits instead of API credits.
func (p *Provider) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	options := ai.ApplyOptions(opts...)

	model := p.defaultModel
	if options.Model != "" {
		model = options.Model
	}

	// Convert full model ID to CLI short name if needed
	model = mapModelToCLIName(model)

	args := []string{
		"--print",
	}

	if model != "" {
		args = append(args, "--model", model)
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Stdin = strings.NewReader(input)

	// CRITICAL: Remove ANTHROPIC_API_KEY to force subscription auth
	cmd.Env = removeAPIKeyFromEnv(os.Environ())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrText := stderr.String()
		stdoutText := stdout.String()

		return "", fmt.Errorf("claude CLI execution failed: %w\nStderr: %s\nStdout: %s",
			err, stderrText, stdoutText)
	}

	output := strings.TrimSpace(stdout.String())
	return output, nil
}

// mapModelToCLIName converts full model IDs to CLI short names.
func mapModelToCLIName(model string) string {
	switch model {
	case "claude-3-haiku-20240307":
		return "haiku"
	case "claude-3-5-sonnet-20240620", "claude-3-5-sonnet-20241022":
		return "sonnet"
	case "claude-3-opus-20240229":
		return "opus"
	default:
		return model
	}
}

// removeAPIKeyFromEnv removes ANTHROPIC_API_KEY from environment variables.
func removeAPIKeyFromEnv(environ []string) []string {
	var filtered []string
	for _, env := range environ {
		if !strings.HasPrefix(env, "ANTHROPIC_API_KEY=") {
			filtered = append(filtered, env)
		}
	}
	return filtered
}

// Compile-time interface check.
var _ ai.Provider = (*Provider)(nil)
