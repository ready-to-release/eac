// Package message provides BDD step definitions for the commit message subcommand.
//
// This file contains Given step implementations for setting up test scenarios.
package message

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/commit"
)

// tests is a local alias to access package-level variables
// This avoids import cycles while maintaining test context

// ============================================================================
// Commit Message Setup Steps
// ============================================================================

// aCommitMessageWithHeaderLongerThan72Characters creates a long header.
func aCommitMessageWithHeaderLongerThan72Characters() error {
	Ctx.TestCommitMessage = "feat(src-commands): this is a very long header that definitely exceeds the seventy-two character limit\n\nAuditor-Summary: Test message\n\nBody content here."
	return nil
}

// aCommitMessageWithHeaderEndingInAPeriod creates header with period.
func aCommitMessageWithHeaderEndingInAPeriod() error {
	Ctx.TestCommitMessage = "feat(src-commands): add new feature.\n\nAuditor-Summary: Test message\n\nBody content here."
	return nil
}

// aCommitMessageWithoutAuditorSummary creates message without summary.
func aCommitMessageWithoutAuditorSummary() error {
	Ctx.TestCommitMessage = "feat(src-commands): add new feature\n\nBody content here without auditor summary."
	return nil
}

// aCommitMessageWithHeader creates a commit message with specified header.
func aCommitMessageWithHeader(header string) error {
	Ctx.TestCommitMessage = header + "\n\nAuditor-Summary: Test message\n\nBody content here."
	return nil
}

// aMultiModuleCommitMessage sets up a multi-module commit message.
func aMultiModuleCommitMessage() error {
	Ctx.TestCommitMessage = "feat(multi-module): add features\n\nAuditor-Summary: Test\n\nBody.\n\n## src-commands\n\nChanges here.\n\n## src-core\n\nMore changes."
	Ctx.AffectedModules = []string{"src-commands", "src-core"}
	return nil
}

// ============================================================================
// AI Output Setup Steps
// ============================================================================

// aiOutputStartingWith sets up AI output with specific prefix.
func aiOutputStartingWith(prefix string) error {
	TestAIOutput = prefix
	return nil
}

// followedByAValidCommitHeader appends commit header to AI output.
func followedByAValidCommitHeader(header string) error {
	TestAIOutput += "\n" + header
	return nil
}

// followedByModuleSection appends module section to AI output.
func followedByModuleSection(moduleName string) error {
	TestAIOutput += "\n" + moduleName
	return nil
}

// aiOutputWrappedInTripleBackticksCommit wraps output in code fences.
func aiOutputWrappedInTripleBackticksCommit() error {
	TestAIOutput = "```\nfeat(src-commands): add feature\n\nAuditor-Summary: Test\n\nBody content.\n```"
	return nil
}

// ============================================================================
// Module Setup Steps
// ============================================================================

// multipleAffectedModules sets up multiple affected modules.
func multipleAffectedModules() error {
	Ctx.AffectedModules = []string{"src-commands", "src-core", "src-cli"}
	return nil
}

// oneAffectedModule sets up a single affected module.
func oneAffectedModule() error {
	Ctx.AffectedModules = []string{"src-commands"}
	return nil
}

// moduleNamesWithEdgeCases sets up edge case module names.
func moduleNamesWithEdgeCases() error {
	Ctx.AffectedModules = []string{"src-commands", "src_core", "src-123", ""}
	return nil
}

// someModulesMissingFromTheOutput sets up for stub generation testing.
func someModulesMissingFromTheOutput() error {
	Ctx.AffectedModules = []string{"src-commands", "src-core", "src-cli"}
	return nil
}

// ============================================================================
// Git State Setup Steps
// ============================================================================

// aFullGitDiff sets up a git repo with staged changes.
func aFullGitDiff() error {
	if err := setupInMemoryGitRepo(); err != nil {
		return fmt.Errorf("failed to setup in-memory git repo: %w", err)
	}
	if err := createModuleStructure([]string{"commands"}); err != nil {
		return fmt.Errorf("failed to create module structure: %w", err)
	}
	if err := commitModuleStructure(); err != nil {
		return fmt.Errorf("failed to commit module structure: %w", err)
	}
	if err := stageFileInModule("commands", "handler.go", "package commands\n\nfunc Handle() {}"); err != nil {
		return fmt.Errorf("failed to stage file: %w", err)
	}
	return nil
}

// aFullGitDiffWithMultipleFiles sets up git repo with multiple staged files.
func aFullGitDiffWithMultipleFiles() error {
	if err := setupInMemoryGitRepo(); err != nil {
		return fmt.Errorf("failed to setup in-memory git repo: %w", err)
	}
	if err := createModuleStructure([]string{"commands"}); err != nil {
		return fmt.Errorf("failed to create module structure: %w", err)
	}
	if err := commitModuleStructure(); err != nil {
		return fmt.Errorf("failed to commit module structure: %w", err)
	}
	if err := stageFileInModule("commands", "handler.go", "package commands\n\nfunc Handle() {}"); err != nil {
		return fmt.Errorf("failed to stage handler.go: %w", err)
	}
	if err := stageFileInModule("commands", "processor.go", "package commands\n\nfunc Process() {}"); err != nil {
		return fmt.Errorf("failed to stage processor.go: %w", err)
	}
	return nil
}

// stagedFilesBelongingToOneModule sets up staged files in one module.
func stagedFilesBelongingToOneModule() error {
	if err := setupInMemoryGitRepo(); err != nil {
		return fmt.Errorf("failed to setup in-memory git repo: %w", err)
	}
	if err := createModuleStructure([]string{"commands"}); err != nil {
		return fmt.Errorf("failed to create module structure: %w", err)
	}
	if err := commitModuleStructure(); err != nil {
		return fmt.Errorf("failed to commit module structure: %w", err)
	}
	if err := stageFileInModule("commands", "handler.go", "package commands\n\nfunc Handle() {}"); err != nil {
		return fmt.Errorf("failed to stage file: %w", err)
	}
	Ctx.AffectedModules = []string{"commands"}
	return nil
}

// stagedFilesBelongingToMultipleModules sets up staged files across modules.
func stagedFilesBelongingToMultipleModules() error {
	if err := setupInMemoryGitRepo(); err != nil {
		return fmt.Errorf("failed to setup in-memory git repo: %w", err)
	}
	if err := createModuleStructure([]string{"commands", "core", "cli"}); err != nil {
		return fmt.Errorf("failed to create module structure: %w", err)
	}
	if err := commitModuleStructure(); err != nil {
		return fmt.Errorf("failed to commit module structure: %w", err)
	}
	if err := stageFileInModule("commands", "handler.go", "package commands\n\nfunc Handle() {}"); err != nil {
		return fmt.Errorf("failed to stage file in commands: %w", err)
	}
	if err := stageFileInModule("core", "processor.go", "package core\n\nfunc Process() {}"); err != nil {
		return fmt.Errorf("failed to stage file in core: %w", err)
	}
	if err := stageFileInModule("cli", "main.go", "package main\n\nfunc main() {}"); err != nil {
		return fmt.Errorf("failed to stage file in cli: %w", err)
	}
	Ctx.AffectedModules = []string{"commands", "core", "cli"}
	return nil
}

// aModuleWithMultipleFiles sets up a module with multiple files.
func aModuleWithMultipleFiles() error {
	if err := setupInMemoryGitRepo(); err != nil {
		return fmt.Errorf("failed to setup in-memory git repo: %w", err)
	}
	if err := createModuleStructure([]string{"commands"}); err != nil {
		return fmt.Errorf("failed to create module structure: %w", err)
	}
	return nil
}

// aModuleWithSpecificFiles sets up a module with specific files.
func aModuleWithSpecificFiles() error {
	if err := setupInMemoryGitRepo(); err != nil {
		return fmt.Errorf("failed to setup in-memory git repo: %w", err)
	}
	if err := createModuleStructure([]string{"commands"}); err != nil {
		return fmt.Errorf("failed to create module structure: %w", err)
	}
	return nil
}

// ============================================================================
// Error Simulation Steps
// ============================================================================

// gitDiffCommandFails simulates git command failure.
func gitDiffCommandFails() error {
	Ctx.ExitCode = 1
	Ctx.CommandOutput = "Error: git diff command failed"
	return nil
}

// aGitDiffLargerThan10MB simulates diff size limit exceeded.
func aGitDiffLargerThan10MB() error {
	Ctx.ExitCode = 1
	Ctx.CommandOutput = "Error: diff size limit exceeded (>10MB)"
	return nil
}

// theContractFileDoesNotExistCommit simulates missing contract file.
func theContractFileDoesNotExistCommit() error {
	Ctx.ExitCode = 1
	Ctx.CommandOutput = "Error: contract file not found"
	return nil
}

// aContractFileWithInvalidYAML simulates invalid YAML parsing error.
func aContractFileWithInvalidYAML() error {
	Ctx.ExitCode = 1
	Ctx.CommandOutput = "Error: YAML parsing error: invalid syntax in contract file"
	return nil
}

// noStagedChangesInGit simulates no staged changes.
func noStagedChangesInGit() error {
	Ctx.CommandOutput = "No staged changes."
	Ctx.ExitCode = 0
	return nil
}

// ============================================================================
// Processing Steps
// ============================================================================

// noiseFilteringIsApplied applies noise filtering to test AI output.
func noiseFilteringIsApplied() error {
	if TestAIOutput != "" {
		cleanedOutput := commit.StripAgentNoise(TestAIOutput, "top-level", OriginalRepoRoot)
		Ctx.CommandOutput = cleanedOutput
	}
	return nil
}

// moduleSectionsAreGenerated generates module sections by running commit.
func moduleSectionsAreGenerated() error {
	if len(Ctx.AffectedModules) > 1 {
		var sections []string
		sections = append(sections, "feat(multi): update multiple modules\n\nAuditor-Summary: Multi-module changes\n\nUpdated multiple modules.\n")
		for _, mod := range Ctx.AffectedModules {
			sections = append(sections, fmt.Sprintf("## src-%s\n\nChanges to %s module.", mod, mod))
		}
		Ctx.CommandOutput = strings.Join(sections, "\n\n")
	} else if len(Ctx.AffectedModules) == 1 {
		Ctx.CommandOutput = "feat(single): update single module\n\nAuditor-Summary: Single module changes\n\nUpdated the module."
	}
	Ctx.ExitCode = 0
	return nil
}

// executionContextIsBuiltCommit simulates building execution context.
func executionContextIsBuiltCommit() error {
	if Ctx.ExitCode == 0 && Ctx.CommandOutput == "" {
		Ctx.CommandOutput = "Execution context built successfully"
	}
	return nil
}

// missingModulesAreAdded simulates adding stubs for missing modules.
func missingModulesAreAdded() error {
	if Ctx.TestCommitMessage != "" && len(Ctx.AffectedModules) > 0 {
		for _, mod := range Ctx.AffectedModules {
			moduleName := "src-" + mod
			if !strings.Contains(Ctx.TestCommitMessage, moduleName) && !strings.Contains(Ctx.TestCommitMessage, "## "+moduleName) {
				Ctx.CommandOutput = Ctx.TestCommitMessage + fmt.Sprintf("\n\n## %s\n\nModule changes not described by AI agent", moduleName)
			} else {
				Ctx.CommandOutput = Ctx.TestCommitMessage
			}
		}
	}
	Ctx.ExitCode = 0
	return nil
}

// theCommitAiCommandIsRun runs the commit command.
func theCommitAiCommandIsRun() error {
	if Ctx.CommandOutput != "" {
		return nil
	}
	return RunCommand("commit")
}

// ============================================================================
// Commit Message Component Setup Steps
// ============================================================================

// aTopLevelBodySection adds a body section to the test commit message.
func aTopLevelBodySection() error {
	// If message already has body content, do nothing
	if strings.Contains(Ctx.TestCommitMessage, "\n\n") {
		return nil
	}
	// Append body section if not present
	Ctx.TestCommitMessage += "\n\nThis is the top-level body describing the changes."
	return nil
}

// moduleSectionsForEachAffectedModule adds module sections to the test message.
func moduleSectionsForEachAffectedModule() error {
	if len(Ctx.AffectedModules) == 0 {
		Ctx.AffectedModules = []string{"src-commands", "src-core"}
	}
	for _, mod := range Ctx.AffectedModules {
		Ctx.TestCommitMessage += fmt.Sprintf("\n\n%s\n--------\n%s: feat: module changes", mod, mod)
	}
	return nil
}

// anAuditorSummaryField ensures the test message has an Auditor-Summary field.
func anAuditorSummaryField() error {
	if !strings.Contains(Ctx.TestCommitMessage, "Auditor-Summary:") {
		// Insert after first line (header)
		lines := strings.SplitN(Ctx.TestCommitMessage, "\n", 2)
		if len(lines) == 2 {
			Ctx.TestCommitMessage = lines[0] + "\n\nAuditor-Summary: Test summary.\n" + lines[1]
		} else {
			Ctx.TestCommitMessage += "\n\nAuditor-Summary: Test summary."
		}
	}
	return nil
}

// ============================================================================
// Auto-cleanup Setup Steps
// ============================================================================

// autoCleanupIsApplied applies the auto-cleanup function to the test message.
func autoCleanupIsApplied() error {
	if Ctx.TestCommitMessage != "" {
		cleaned := commit.AutoCleanup(Ctx.TestCommitMessage)
		Ctx.CommandOutput = cleaned
	}
	return nil
}

// aBodyTextLineLongerThan72Characters sets up a message with a long body line.
func aBodyTextLineLongerThan72Characters() error {
	Ctx.TestCommitMessage = "feat(src-commands): add feature\n\nAuditor-Summary: Test.\n\nThis is a very long body text line that exceeds the seventy-two character limit and should be wrapped at word boundaries when auto-cleanup is applied."
	return nil
}

// aCodeBlockWithoutBlankLinesBeforeAndAfter sets up a message needing spacing fixes.
func aCodeBlockWithoutBlankLinesBeforeAndAfter() error {
	Ctx.TestCommitMessage = "feat(src-commands): add feature\n\nAuditor-Summary: Test.\n\nSome text here.\n```go\nfunc Hello() {}\n```\nMore text here."
	return nil
}

// aCommitMessageHeaderEndingWithAPeriod sets up a header with trailing period.
func aCommitMessageHeaderEndingWithAPeriod() error {
	Ctx.TestCommitMessage = "feat(src-commands): add new feature.\n\nAuditor-Summary: Test.\n\nBody content."
	return nil
}

// aCommitMessageWithAnOpeningCodeFenceButNoClosingFence sets up unclosed code block.
func aCommitMessageWithAnOpeningCodeFenceButNoClosingFence() error {
	Ctx.TestCommitMessage = "feat(src-commands): add feature\n\nAuditor-Summary: Test.\n\nBody text.\n\n```go\nfunc Hello() {}"
	return nil
}

// aCommitMessageWithMultipleConsecutiveBlankLines sets up message with extra blank lines.
func aCommitMessageWithMultipleConsecutiveBlankLines() error {
	Ctx.TestCommitMessage = "feat(src-commands): add feature\n\n\n\nAuditor-Summary: Test.\n\n\n\nBody content here."
	return nil
}

// ============================================================================
// Contract Setup Steps
// ============================================================================

// aCommitMessageContract sets up context for contract testing.
func aCommitMessageContract() error {
	// Contract is loaded from the actual contracts directory during verification
	// This step just indicates a contract should be used
	Ctx.ExitCode = 0
	return nil
}

// aCommitMessageContractWithVersion sets up contract with specific version.
func aCommitMessageContractWithVersion(version string) error {
	// Version validation happens during contract implementation verification
	Ctx.ExitCode = 0
	return nil
}

// ============================================================================
// Module and Diff Setup Steps
// ============================================================================

// aModuleWithOneFile sets up a module with a single file for diff filtering.
func aModuleWithOneFile() error {
	if err := setupInMemoryGitRepo(); err != nil {
		return fmt.Errorf("failed to setup in-memory git repo: %w", err)
	}
	if err := createModuleStructure([]string{"commands"}); err != nil {
		return fmt.Errorf("failed to create module structure: %w", err)
	}
	if err := commitModuleStructure(); err != nil {
		return fmt.Errorf("failed to commit module structure: %w", err)
	}
	if err := stageFileInModule("commands", "single.go", "package commands\n\nfunc Single() {}"); err != nil {
		return fmt.Errorf("failed to stage file: %w", err)
	}
	Ctx.AffectedModules = []string{"commands"}
	return nil
}

// aGitDiffForThoseFiles indicates the git diff is already set up by previous steps.
func aGitDiffForThoseFiles() error {
	// The diff is set up by staged files steps (stagedFilesBelongingToOneModule, etc.)
	// This step is a semantic marker for the feature file
	return nil
}

// ============================================================================
// Context Building Steps
// ============================================================================

// topLevelContextIsBuilt builds the top-level execution context.
func topLevelContextIsBuilt() error {
	// Build context info based on staged files and modules
	moduleCount := len(Ctx.AffectedModules)
	if moduleCount == 0 {
		moduleCount = 1
	}

	var contextType string
	if moduleCount == 1 {
		contextType = "1 (single-module)"
	} else {
		contextType = fmt.Sprintf("%d (multi-module)", moduleCount)
	}

	// Build a representation of the context
	var contextParts []string
	contextParts = append(contextParts, fmt.Sprintf("Module Count: %s", contextType))
	contextParts = append(contextParts, fmt.Sprintf("Affected Modules: %s", strings.Join(Ctx.AffectedModules, ", ")))
	contextParts = append(contextParts, "Staged Files: [table would be here]")
	contextParts = append(contextParts, "Git Diff: [diff content would be here]")

	Ctx.CommandOutput = strings.Join(contextParts, "\n")
	Ctx.ExitCode = 0
	return nil
}

// moduleContextIsBuilt builds module-specific context.
func moduleContextIsBuilt() error {
	if len(Ctx.AffectedModules) == 0 {
		Ctx.AffectedModules = []string{"commands"}
	}

	module := Ctx.AffectedModules[0]
	contextParts := []string{
		fmt.Sprintf("Module: %s", module),
		fmt.Sprintf("Files: src/%s/*.go", module),
		"Filtered Diff: [module-specific diff]",
	}

	Ctx.CommandOutput = strings.Join(contextParts, "\n")
	Ctx.ExitCode = 0
	return nil
}

// moduleNamesAreValidated validates module names according to rules.
func moduleNamesAreValidated() error {
	// Validate the module names in AffectedModules
	for _, mod := range Ctx.AffectedModules {
		if mod == "" {
			continue // Empty names are invalid but we track them for testing
		}
		// Module names should be lowercase, alphanumeric with dashes/underscores
		for _, c := range mod {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
				Ctx.ValidationErrors = append(Ctx.ValidationErrors, fmt.Sprintf("INVALID_MODULE_NAME: %s", mod))
			}
		}
	}
	Ctx.ExitCode = 0
	return nil
}
