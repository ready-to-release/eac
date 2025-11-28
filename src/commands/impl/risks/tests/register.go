// Package tests provides test registration and setup for risks commands.
package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/commands/impl/risks/assessment"
	"github.com/ready-to-release/eac/src/commands/impl/risks/create"
	"github.com/ready-to-release/eac/src/core/git"
)

// InitializeRisksScenario registers all risks-related step definitions.
func InitializeRisksScenario(sc *godog.ScenarioContext) {
	// Initialize context before each scenario
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		if Ctx == nil {
			Ctx = NewTestContext()
		}
		Ctx.Reset()

		if SharedCtx != nil {
			SharedCtx.Reset()
		}

		// Set up mocks if in isolated mode
		if IsolatedTestProjectDir != "" {
			if err := SetupRisksMocks(); err != nil {
				return ctx, err
			}
		}

		return ctx, nil
	})

	// Register all step definitions
	registerAssessmentSteps(sc)
	registerCreateSteps(sc)
	registerListSteps(sc)

	// Cleanup after scenario
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		CleanupRisksMocks()
		return ctx, nil
	})
}

// SetupRisksMocks sets up git and AI mocks for isolated testing.
// Called automatically when @env:isolated-test-project tag is present.
func SetupRisksMocks() error {
	if IsolatedTestProjectDir == "" {
		return nil // Not in isolated mode
	}

	// Set up mock git repository
	TestMockRepo = git.NewMockRepository(IsolatedTestProjectDir).
		WithCurrentBranch("main").
		WithHeadSHA("abc1234567890")

	// Inject into assessment command
	assessment.SetGitRepo(TestMockRepo)

	// Load and set mock AI response for assessment command
	if OriginalRepoRoot != "" {
		mockResponsePath := filepath.Join(OriginalRepoRoot,
			"src/commands/impl/risks/tests/assets/mock-assessment-response.md")
		mockResponse, err := os.ReadFile(mockResponsePath)
		if err == nil {
			assessment.SetMockAIResponse(string(mockResponse))
			Ctx.MockAIResponse = string(mockResponse)
		}

		// Load mock response for create command (JSON)
		createResponsePath := filepath.Join(OriginalRepoRoot,
			"src/commands/impl/risks/tests/assets/mock-create-response.json")
		createResponse, err := os.ReadFile(createResponsePath)
		if err == nil {
			create.SetMockAIResponse(string(createResponse))
		}
	}

	return nil
}

// CleanupRisksMocks resets all mocks.
func CleanupRisksMocks() {
	assessment.ResetGitRepo()
	assessment.ResetMockAIResponse()
	create.ResetMockAIResponse()
	TestMockRepo = nil
}

// ============================================================================
// Helper Functions (used by step implementations)
// ============================================================================
// Note: These are NOT registered as steps - the main test runner handles step registration.
// These are internal helper functions used by step implementations.

func stdoutContains(text string) error {
	if !strings.Contains(Ctx.CommandOutput, text) {
		return fmt.Errorf("expected stdout to contain %q, got: %s", text, Ctx.CommandOutput)
	}
	return nil
}

func stderrContains(text string) error {
	// Check command output for error text
	if !strings.Contains(Ctx.CommandOutput, text) {
		return fmt.Errorf("expected stderr to contain %q", text)
	}
	return nil
}
