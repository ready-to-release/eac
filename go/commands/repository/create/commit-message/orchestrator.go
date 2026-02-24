package commitmessage

import (
	"fmt"

	commitmessageinternal "github.com/ready-to-release/eac/go/commands/repository/create/commit-message/internal"
	"github.com/ready-to-release/eac/go/commands/repository/create/aiutil"
	"github.com/ready-to-release/eac/go/core/logging"
)

func CreateCommitMessage() int {
	return createCommitMessageWithDeps(defaultDeps())
}

func createCommitMessageWithDeps(deps *Deps) int {
	// Parse configuration early to get debug mode, auto-commit flag, and workspace root
	debug, autoCommit, workspaceRoot, err := parseConfig()
	if err != nil {
		log.Errorf("ERROR: %v", err)
		return 1
	}

	// Configure logging system (component loggers + file logging)
	if err := logging.ConfigureLoggingSimple(workspaceRoot, "commands", nil, debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	log.Debugf("Logging configured (debug=%v)", debug)

	// Retry loop for regenerating commit message if validation fails
	// Limited to prevent infinite loops
	const maxRetries = 5
	attempt := 0

	for attempt < maxRetries {
		attempt++

		// Show warning after multiple retries
		if attempt > 3 {
			log.Warnf("retry attempt: attempt=%d, max=%d", attempt, maxRetries)
		}

		result, shouldRetry, generatedMessage := commitAIAttemptWithMessage(deps, workspaceRoot, debug)
		if !shouldRetry {
			// If successful and auto-commit is enabled, perform the commit
			if result == 0 && autoCommit && generatedMessage != "" {
				return performAutoCommit(deps, workspaceRoot, generatedMessage)
			}
			return result
		}

		// Check if max retries reached
		if attempt >= maxRetries {
			log.Error("Maximum retry attempts reached")
			log.Info("The AI is having difficulty generating a valid commit message.")
			log.Info("Please try one of the following:")
			log.Info("  - Simplify your staged changes")
			log.Info("  - Split changes across multiple commits")
			log.Info("  - Write commit message manually with: git commit")
			return 1
		}

		log.Infof("retrying commit message generation: attempt=%d, max=%d", attempt+1, maxRetries)
	}

	return 1
}

// commitAIAttempt performs a single attempt at generating and committing
// Returns (exit code, should retry).
func commitAIAttempt(deps *Deps, workspaceRoot string, debug bool) (int, bool) {
	exitCode, shouldRetry, _ := commitAIAttemptWithMessage(deps, workspaceRoot, debug)
	return exitCode, shouldRetry
}

// commitAIAttemptWithMessage performs a single attempt at generating commit message
// Returns (exit code, should retry, generated message).
func commitAIAttemptWithMessage(deps *Deps, workspaceRoot string, debug bool) (int, bool, string) {
	// Phase 1: Verify Contract Implementation
	if err := verifyContractImplementation(workspaceRoot); err != nil {
		return 1, false, ""
	}

	// Phase 2: Build Execution Context
	cfg, stagedFilesTable, diffStats, err := buildExecutionContext(deps, workspaceRoot, debug)
	if err != nil {
		log.Errorf("ERROR: Build context failed: %v", err)
		return 1, false, ""
	}
	if cfg == nil {
		log.Info("No staged changes.")
		return 1, false, ""
	}

	// Phase 3: Generate Top-Level Summary
	topLevel, err := generateTopLevelSummary(deps, cfg, stagedFilesTable, diffStats)
	if err != nil {
		log.Errorf("ERROR: Top-level generation failed: %v", err)
		return 1, false, ""
	}

	// Phase 4: Generate Module Sections
	moduleSections, err := generateModuleSections(deps, cfg)
	if err != nil {
		log.Errorf("ERROR: Module section generation failed: %v", err)
		return 1, false, ""
	}

	// Phase 5: Assemble Final Message
	finalMessage := assembleFinalMessage(cfg, topLevel, moduleSections)

	// Phase 6: Validate and Output (message only - no git commit)
	exitCode, shouldRetry := validateAndOutput(cfg, finalMessage)
	return exitCode, shouldRetry, finalMessage
}

// Phase 4: Generate Top-Level Summary.
func generateTopLevelSummary(deps *Deps, cfg *executionConfig, stagedFilesTable, diffStats string) (string, error) {
	topLevelContext := buildTopLevelContext(stagedFilesTable, cfg.gitDiff, diffStats, cfg.affectedModules)

	logDebugArtifact("TOP-LEVEL-CONTEXT", topLevelContext)

	var topLevelOutput string
	var providerName string
	err := aiutil.WithProgress("🤖 Generating top-level commit summary...", func() error {
		result, genErr := generateWithPromptResult(deps, "top-level", topLevelContext, cfg.workspaceRoot, cfg.affectedModules, cfg.debug, nil)
		if result != nil {
			topLevelOutput = result.Output
			providerName = result.ProviderName
		}
		return genErr
	})
	if err != nil {
		return "", fmt.Errorf("running commit-message-top-level agent: %w", err)
	}

	// Log provider info when debug is enabled
	if providerName != "" {
		log.Debugf("AI provider used: %s", providerName)
	}

	// Strip out any module sections the AI may have added after "Changes:" line
	// Module sections will be generated separately and appended later
	topLevelOutput = stripModuleSectionsFromTopLevel(topLevelOutput)

	logDebugArtifact("TOP-LEVEL-OUTPUT", topLevelOutput)

	return topLevelOutput, nil
}

// Phase 5: Generate Module Sections (multi-module only).
func generateModuleSections(deps *Deps, cfg *executionConfig) ([]string, error) {
	// Use parallel implementation for performance (60-70% speedup for multi-module commits)
	// Sequential: N modules × 5s = 15s for 3 modules
	// Parallel:   max(5s) = 5s for 3 modules
	return generateModuleSectionsParallel(deps, cfg, nil)
}

// Phase 6: Assemble Final Message.
func assembleFinalMessage(cfg *executionConfig, topLevel string, moduleSections []string) string {
	// Combine sections
	combinedMessage := combineCommitSections(topLevel, moduleSections)
	logDebugArtifact("COMBINED-MESSAGE", combinedMessage)

	// Auto-cleanup
	cleanedOutput := commitmessageinternal.AutoCleanup(combinedMessage)
	logDebugArtifact("AFTER-CLEANUP", cleanedOutput)

	// NOTE: addMissingModules fallback is no longer needed with parallel generation
	// The new generateModuleSectionsParallel guarantees all affected modules get sections
	// Keeping the fallback causes duplicate sections, so it's disabled
	// cleanedOutput = addMissingModules(cleanedOutput, cfg.affectedModules, cfg.stagedFiles, cfg.gitDiff)
	// logDebugArtifact(logger, "AFTER-MISSING-MODULES", cleanedOutput)

	return cleanedOutput
}
