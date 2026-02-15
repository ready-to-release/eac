// Package aiutil provides shared utilities for AI-powered commit message generation commands
// (create commit-message, create squash-message).
package aiutil

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/adapters/ai"
	"github.com/ready-to-release/eac/go/adapters/ai/providers"
	coreai "github.com/ready-to-release/eac/go/core/ai"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
)

// GenerationParams configures a single AI generation call.
type GenerationParams struct {
	WorkspaceRoot  string
	Prompt         string // Full prompt (template + user input)
	TypeName       string // e.g. coreai.TypeCommitMessage, coreai.TypeSquashMessage
	SchemaFilename string // e.g. "commit-message.schema.json"
	Model          string // Optional model override from agent frontmatter
	Debug          bool
	TestExecutor   *ai.Executor // Optional: use this executor instead of creating one (for testing)
}

// ExecuteGeneration performs the shared AI executor pipeline:
// create executor → register providers → load AI config → build schema validator →
// build retry config → generate with retry.
func ExecuteGeneration(params GenerationParams) (*coreai.RetryResult, error) {
	var executor *ai.Executor
	if params.TestExecutor != nil {
		executor = params.TestExecutor
	} else {
		executor = ai.NewExecutor(params.WorkspaceRoot)
		providers.RegisterBuiltIn(executor)
	}

	var executorAdapter *ai.ExecutorAdapter
	if params.Model != "" {
		executorAdapter = ai.NewExecutorAdapterWithModel(executor, params.Model)
	} else {
		executorAdapter = ai.NewExecutorAdapter(executor)
	}

	aiConfig, err := coreai.LoadAIConfig(params.WorkspaceRoot)
	if err != nil {
		logging.C().Warnf("Could not load AI config, using default retry strategy: %v", err)
		aiConfig = nil
	}

	schemaPath := filepath.Join(
		paths.ContractsVersionPath(params.WorkspaceRoot, paths.EACCoreModule, paths.DefaultsVersion),
		paths.SchemasDir,
		params.SchemaFilename,
	)
	validator, err := domain.NewJSONSchemaValidator(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON schema validator: %w", err)
	}

	opts := []coreai.RetryConfigOption{
		coreai.WithLogger(logging.C().Zap()),
	}
	if params.Debug {
		opts = append(opts, coreai.WithDebug(true))
	}

	retryConfig, err := coreai.BuildRetryConfig(
		params.TypeName,
		coreai.FormatJSON,
		executorAdapter,
		validator,
		params.WorkspaceRoot,
		aiConfig,
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build retry config: %w", err)
	}

	result, err := coreai.GenerateWithRetry(context.Background(), retryConfig, params.Prompt)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	return result, nil
}

// LogDebugArtifact logs debug content with labeled sections to the log file.
// Used by AI generation commands to record intermediate outputs for troubleshooting.
func LogDebugArtifact(log *logging.ComponentLogger, label, content string) {
	log.Debug(fmt.Sprintf("=== %s START ===", label))
	log.Debug(content)
	log.Debug(fmt.Sprintf("=== %s END ===", label))
}

// ExtractFirstSentence extracts the first sentence from text.
// Returns the first sentence terminated by a period, question mark, or exclamation mark.
// If no sentence delimiter is found, returns the first line with a trailing period.
// The fallback parameter is used when text is empty.
func ExtractFirstSentence(text, fallback string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		if fallback != "" {
			return ensurePeriod(fallback)
		}
		return "Summary of changes."
	}

	// Find first sentence ending
	for _, delim := range []string{". ", ".\n", "? ", "?\n", "! ", "!\n"} {
		if idx := strings.Index(text, delim); idx != -1 {
			return strings.TrimSpace(text[:idx+1])
		}
	}

	// If text ends with punctuation, return first line
	if strings.HasSuffix(text, ".") || strings.HasSuffix(text, "?") || strings.HasSuffix(text, "!") {
		lines := strings.Split(text, "\n")
		return strings.TrimSpace(lines[0])
	}

	// Fallback: first line with period
	lines := strings.Split(text, "\n")
	firstLine := strings.TrimSpace(lines[0])
	if firstLine == "" {
		return ensurePeriod(fallback)
	}
	return ensurePeriod(firstLine)
}

func ensurePeriod(s string) string {
	if s == "" {
		return "."
	}
	if !strings.HasSuffix(s, ".") {
		return s + "."
	}
	return s
}
