// Package pipeline contains godog step implementations for eac-cli.
//
// This file contains pipeline status command step definitions.
package pipeline

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/core/environments"
	eacgodog "github.com/ready-to-release/eac/go/godog"
)

// registerStatusSteps registers step definitions for pipeline status command features.
func registerStatusSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Setup steps
	sc.Step(`^GitHub Actions workflows exist for the repository$`, func() error {
		return githubActionsWorkflowsExist(ctx)
	})
	sc.Step(`^all workflows are passing on main$`, func() error {
		return allWorkflowsPassing(ctx)
	})
	sc.Step(`^a workflow is failing on main$`, func() error {
		return workflowIsFailing(ctx)
	})
	sc.Step(`^branch "([^"]*)" exists with workflows$`, func(branchName string) error {
		return branchExistsWithWorkflows(ctx, branchName)
	})
	sc.Step(`^a commit "([^"]*)" has workflow runs$`, func(commitSHA string) error {
		return commitHasWorkflowRuns(ctx, commitSHA)
	})
	sc.Step(`^no workflows have run for the commit$`, func() error {
		return noWorkflowsForCommit(ctx)
	})

	// Verification steps
	sc.Step(`^I see workflow names and their status$`, func() error {
		return iSeeWorkflowNamesAndStatus(ctx)
	})
	sc.Step(`^I see the commit SHA being checked$`, func() error {
		return iSeeCommitSHA(ctx)
	})
	sc.Step(`^I see status indicators showing success$`, func() error {
		return iSeeSuccessIndicators(ctx)
	})
	sc.Step(`^I see status indicators showing failure$`, func() error {
		return iSeeFailureIndicators(ctx)
	})
	sc.Step(`^I see the failing workflow name$`, func() error {
		return iSeeFailingWorkflowName(ctx)
	})
	sc.Step(`^I see status for the develop branch HEAD$`, func() error {
		return iSeeStatusForBranch(ctx, "develop")
	})
	sc.Step(`^I see status for commit "([^"]*)"$`, func(commitSHA string) error {
		return iSeeStatusForCommit(ctx, commitSHA)
	})
	sc.Step(`^I see "([^"]*)" or similar message$`, func(message string) error {
		return iSeeSimilarMessage(ctx, message)
	})
}

// ============================================================================
// Setup Steps
// ============================================================================

func githubActionsWorkflowsExist(ctx *eacgodog.TestContext) error {
	// Create a test module with workflows
	// The createTestModule helper already creates workflow files
	return createTestModule(ctx, "test-module", nil)
}

func allWorkflowsPassing(ctx *eacgodog.TestContext) error {
	// Create a test module - by default all workflows pass in mock
	// No need to set R2R_MOCK_FAILING_WORKFLOW
	return createTestModule(ctx, "test-module", nil)
}

func workflowIsFailing(ctx *eacgodog.TestContext) error {
	// Create a test module and set mock to return failure
	if err := createTestModule(ctx, "test-module", nil); err != nil {
		return err
	}

	// Set environment variable to simulate failing workflow
	ctx.SetMockOverride("R2R_MOCK_FAILING_WORKFLOW", "test-module")
	return nil
}

func branchExistsWithWorkflows(ctx *eacgodog.TestContext, branchName string) error {
	// Create a test module first
	if err := createTestModule(ctx, "test-module", nil); err != nil {
		return err
	}

	// Create and check out the branch
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = ctx.IsolatedDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %w\nOutput: %s", branchName, err, string(output))
	}

	// Commit the module in this branch
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("Add test-module to %s", branchName))
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func commitHasWorkflowRuns(ctx *eacgodog.TestContext, commitSHA string) error {
	// Create a test module
	if err := createTestModule(ctx, "test-module", nil); err != nil {
		return err
	}

	// Commit it
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-m", "Test commit")
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// The commit SHA "abc123" is handled by the mock - it will return workflows for any SHA
	return nil
}
func noWorkflowsForCommit(ctx *eacgodog.TestContext) error {
	// Set environment to return no workflows
	ctx.SetMockOverride(environments.EnvR2RMockNoWorkflows, "true")
	return nil
}

// ============================================================================
// Verification Steps
// ============================================================================

func iSeeWorkflowNamesAndStatus(ctx *eacgodog.TestContext) error {
	output := ctx.CommandOutput
	if !strings.Contains(output, "Workflows:") && !strings.Contains(output, "workflow") {
		return fmt.Errorf("expected workflow information in output, got: %s", output)
	}
	return nil
}

func iSeeCommitSHA(ctx *eacgodog.TestContext) error {
	output := ctx.CommandOutput
	if !strings.Contains(output, "commit") && !strings.Contains(output, "SHA") {
		return fmt.Errorf("expected commit SHA in output, got: %s", output)
	}
	return nil
}

func iSeeSuccessIndicators(ctx *eacgodog.TestContext) error {
	output := ctx.CommandOutput
	if !strings.Contains(output, "success") {
		return fmt.Errorf("expected success indicators in output, got: %s", output)
	}
	return nil
}

func iSeeFailureIndicators(ctx *eacgodog.TestContext) error {
	output := ctx.CommandOutput
	if !strings.Contains(output, "failure") {
		return fmt.Errorf("expected failure indicators in output, got: %s", output)
	}
	return nil
}

func iSeeFailingWorkflowName(ctx *eacgodog.TestContext) error {
	output := ctx.CommandOutput
	// The failing workflow name should be in the output
	if !strings.Contains(output, "test-module") {
		return fmt.Errorf("expected failing workflow name in output, got: %s", output)
	}
	return nil
}

func iSeeStatusForBranch(ctx *eacgodog.TestContext, branchName string) error {
	output := ctx.CommandOutput
	// Should see commit SHA and workflow status
	if !strings.Contains(output, "commit") {
		return fmt.Errorf("expected commit information for branch %s in output, got: %s", branchName, output)
	}
	return nil
}

func iSeeStatusForCommit(ctx *eacgodog.TestContext, commitSHA string) error {
	output := ctx.CommandOutput
	// Should see the commit SHA in the output
	if !strings.Contains(output, commitSHA) && !strings.Contains(output, "commit") {
		return fmt.Errorf("expected commit %s in output, got: %s", commitSHA, output)
	}
	return nil
}

func iSeeSimilarMessage(ctx *eacgodog.TestContext, message string) error {
	output := ctx.CommandOutput

	// Normalize the message for flexible matching
	normalizedMessage := strings.ToLower(message)
	normalizedOutput := strings.ToLower(output)

	// Split on "or" to check multiple possible messages
	possibleMessages := strings.Split(normalizedMessage, " or ")

	for _, msg := range possibleMessages {
		msg = strings.TrimSpace(msg)
		// Check for key words in the message
		words := strings.Fields(msg)
		matchCount := 0
		for _, word := range words {
			if strings.Contains(normalizedOutput, word) {
				matchCount++
			}
		}
		// If most words match, consider it a match
		if matchCount >= len(words)/2 && matchCount > 0 {
			return nil
		}
	}

	return fmt.Errorf("expected message similar to %q in output, got: %s", message, output)
}
