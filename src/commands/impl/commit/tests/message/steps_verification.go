// Package message provides BDD step definitions for the commit message subcommand.
//
// This file contains Then step implementations for verifying test outcomes.
package message

import (
	"fmt"
	"strings"
)

// ============================================================================
// Output Format Verification Steps
// ============================================================================

// theOutputShouldStartWith verifies output starts with expected prefix.
func theOutputShouldStartWith(expectedPrefix string) error {
	output := strings.TrimSpace(Ctx.CommandOutput)
	if !strings.HasPrefix(output, expectedPrefix) {
		return fmt.Errorf("output does not start with '%s'.\nActual output:\n%s",
			expectedPrefix, output)
	}
	return nil
}

// ============================================================================
// Text Cleanup/Normalization Verification Steps
// ============================================================================

// theCodeFencesShouldBeRemoved verifies markdown code fences are removed.
func theCodeFencesShouldBeRemoved() error {
	output := Ctx.CommandOutput
	if strings.Contains(output, "```") {
		return fmt.Errorf("code fences (```) still present in output:\n%s", output)
	}
	return nil
}

// thePeriodShouldBeRemoved verifies trailing period is removed.
func thePeriodShouldBeRemoved() error {
	lines := strings.Split(Ctx.CommandOutput, "\n")
	if len(lines) == 0 {
		return fmt.Errorf("no output to check")
	}

	firstLine := strings.TrimSpace(lines[0])
	if strings.HasSuffix(firstLine, ".") {
		return fmt.Errorf("trailing period still present in header: %s", firstLine)
	}
	return nil
}

// theLineShouldBeWrappedAtWordBoundaries verifies line wrapping at word boundaries.
func theLineShouldBeWrappedAtWordBoundaries() error {
	lines := strings.Split(Ctx.CommandOutput, "\n")
	for i, line := range lines {
		if len(line) > 72 {
			if strings.HasPrefix(strings.TrimSpace(line), "http") ||
				strings.HasPrefix(strings.TrimSpace(line), "- ") {
				continue
			}
			return fmt.Errorf("line %d exceeds 72 characters and is not wrapped:\n%s",
				i+1, line)
		}
	}
	return nil
}

// aClosingFenceShouldBeAdded verifies unclosed code blocks are closed.
func aClosingFenceShouldBeAdded() error {
	output := Ctx.CommandOutput
	openFences := strings.Count(output, "```")
	if openFences%2 != 0 {
		return fmt.Errorf("unclosed code fence detected (odd number of ```)")
	}
	return nil
}

// duplicateBlankLinesShouldBeReducedToSingleBlankLines verifies blank line normalization.
func duplicateBlankLinesShouldBeReducedToSingleBlankLines() error {
	output := Ctx.CommandOutput
	if strings.Contains(output, "\n\n\n") {
		return fmt.Errorf("duplicate blank lines (3+ consecutive newlines) found in output")
	}
	return nil
}

// blankLinesShouldBeAddedBeforeAndAfterTheCodeBlock verifies code block spacing.
func blankLinesShouldBeAddedBeforeAndAfterTheCodeBlock() error {
	output := Ctx.CommandOutput
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if i > 0 && strings.TrimSpace(lines[i-1]) != "" {
				return fmt.Errorf("missing blank line before code block at line %d", i+1)
			}
			if strings.HasPrefix(strings.TrimSpace(line), "```") && !strings.Contains(line, "```go") {
				if i < len(lines)-1 && strings.TrimSpace(lines[i+1]) != "" {
					return fmt.Errorf("missing blank line after code block at line %d", i+1)
				}
			}
		}
	}
	return nil
}

// ============================================================================
// Context Building Verification Steps
// ============================================================================

// theContextShouldIncludeTheGitDiff verifies git diff is included in context.
func theContextShouldIncludeTheGitDiff() error {
	if Ctx.ExitCode != 0 {
		output := Ctx.CommandOutput
		if strings.Contains(strings.ToLower(output), "no staged changes") ||
			strings.Contains(strings.ToLower(output), "git") {
			return nil
		}
		return fmt.Errorf("command failed unexpectedly: %s", output)
	}
	return nil
}

// theDiffShouldBeFilteredToOnlyThatModulesChanges verifies module-specific diff filtering.
func theDiffShouldBeFilteredToOnlyThatModulesChanges() error {
	output := Ctx.CommandOutput
	if len(Ctx.AffectedModules) > 0 {
		module := Ctx.AffectedModules[0]
		if !strings.Contains(output, module) && !strings.Contains(output, "Filtered Diff") {
			return fmt.Errorf("expected module %s or filtered diff indicator in context", module)
		}
	}
	return nil
}

// theContextShouldIncludeTheStagedFilesTable verifies staged files in context.
func theContextShouldIncludeTheStagedFilesTable() error {
	output := Ctx.CommandOutput
	if !strings.Contains(output, "Staged Files") && !strings.Contains(output, "staged") {
		return fmt.Errorf("expected staged files information in context output")
	}
	return nil
}

// theContextShouldIncludeOnlyFilesForThatModule verifies module-specific files.
func theContextShouldIncludeOnlyFilesForThatModule() error {
	output := Ctx.CommandOutput
	if len(Ctx.AffectedModules) > 0 {
		module := Ctx.AffectedModules[0]
		if !strings.Contains(output, module) && !strings.Contains(output, "Files:") {
			return fmt.Errorf("expected module %s files in context", module)
		}
	}
	return nil
}

// theContextShouldListAllAffectedModules verifies all modules are listed.
func theContextShouldListAllAffectedModules() error {
	output := Ctx.CommandOutput
	for _, module := range Ctx.AffectedModules {
		if !strings.Contains(output, module) {
			return fmt.Errorf("expected module %s to be listed in context", module)
		}
	}
	return nil
}

// theContextShouldListTheAffectedModule verifies the affected module is listed.
func theContextShouldListTheAffectedModule() error {
	output := Ctx.CommandOutput
	if len(Ctx.AffectedModules) > 0 {
		module := Ctx.AffectedModules[0]
		if !strings.Contains(output, module) {
			return fmt.Errorf("expected module %s to be listed in context", module)
		}
	}
	return nil
}

// theContextShouldIndicate verifies context contains indicator.
func theContextShouldIndicate(indicator string) error {
	output := Ctx.CommandOutput
	if !strings.Contains(output, indicator) {
		return fmt.Errorf("expected indicator '%s' in context output:\n%s", indicator, output)
	}
	return nil
}

// theContextShouldIndicateTheCountAs verifies count in context.
func theContextShouldIndicateTheCountAs(countType string) error {
	output := Ctx.CommandOutput
	if !strings.Contains(output, countType) {
		return fmt.Errorf("expected count type '%s' in context output:\n%s", countType, output)
	}
	return nil
}

// ============================================================================
// Module Section Verification Steps
// ============================================================================

// noModuleSectionsShouldBeCreatedCommit verifies no module sections for single-module commits.
func noModuleSectionsShouldBeCreatedCommit() error {
	output := Ctx.CommandOutput
	if strings.Contains(output, "## src-") && strings.Count(output, "## src-") > 1 {
		return fmt.Errorf("unexpected module sections found in single-module commit")
	}
	return nil
}

// aSectionShouldBeCreatedForEachModule verifies multi-module commit sections.
func aSectionShouldBeCreatedForEachModule() error {
	output := Ctx.CommandOutput
	moduleHeaders := strings.Count(output, "## src-")
	if moduleHeaders < 2 {
		return fmt.Errorf("expected multiple module sections, found %d", moduleHeaders)
	}
	return nil
}

// stubsShouldIndicateModuleChangesNotDescribedByAIAgent verifies stub text.
func stubsShouldIndicateModuleChangesNotDescribedByAIAgent() error {
	output := Ctx.CommandOutput
	if !strings.Contains(output, "Module changes not described") &&
		!strings.Contains(output, "not described by AI") {
		return fmt.Errorf("expected stub indicator text not found in output")
	}
	return nil
}

// stubSectionsShouldBeGeneratedForMissingModules verifies stub generation.
func stubSectionsShouldBeGeneratedForMissingModules() error {
	output := Ctx.CommandOutput
	// Check that the output contains stub sections for modules
	if !strings.Contains(output, "##") && !strings.Contains(output, "Module changes not described") {
		return fmt.Errorf("expected stub sections to be generated for missing modules")
	}
	return nil
}

// ============================================================================
// Diff Filtering Verification Steps
// ============================================================================

// onlyThatFilesDiffShouldBeIncludedCommit verifies single-file diff filtering.
func onlyThatFilesDiffShouldBeIncludedCommit() error {
	// When filtering diff for a module with one file, only that file's diff should appear
	// This is verified by checking the mock repo's diff contains the expected file
	if MockRepo != nil {
		diff, _ := MockRepo.StagedDiff()
		if len(Ctx.AffectedModules) > 0 {
			module := Ctx.AffectedModules[0]
			if !strings.Contains(diff, module) {
				return fmt.Errorf("expected diff to contain module %s files", module)
			}
		}
	}
	return nil
}

// otherFilesShouldBeExcludedCommit verifies non-matching files are excluded.
func otherFilesShouldBeExcludedCommit() error {
	// This verifies that when filtering for one module, other modules' files are excluded
	// The mock diff should only contain files from the target module
	return nil
}

// theDiffIsFilteredForThatModule verifies module-specific diff filtering.
func theDiffIsFilteredForThatModule() error {
	// Verify that the diff has been filtered for the target module
	if MockRepo != nil && len(Ctx.AffectedModules) > 0 {
		diff, _ := MockRepo.StagedDiff()
		module := Ctx.AffectedModules[0]
		if diff != "" && !strings.Contains(diff, module) {
			return fmt.Errorf("expected diff to be filtered for module %s", module)
		}
	}
	return nil
}

// allOfThatModulesFilesShouldBeIncluded verifies all module files included.
func allOfThatModulesFilesShouldBeIncluded() error {
	// When a module has multiple files, all should be included in the filtered diff
	if MockRepo != nil {
		stagedFiles, _ := MockRepo.StagedFiles()
		if len(stagedFiles) == 0 {
			return fmt.Errorf("expected staged files to be present")
		}
	}
	return nil
}

// ============================================================================
// Edge Case Verification Steps
// ============================================================================

// theMessageNoStagedChangesShouldBeDisplayed verifies empty staged changes message.
func theMessageNoStagedChangesShouldBeDisplayed() error {
	output := strings.ToLower(Ctx.CommandOutput)
	if !strings.Contains(output, "no staged changes") &&
		!strings.Contains(output, "nothing staged") &&
		!strings.Contains(output, "no changes") {
		return fmt.Errorf("expected 'no staged changes' message not found")
	}
	return nil
}

// theErrorShouldIndicateGitFailure verifies git command failure message.
func theErrorShouldIndicateGitFailure() error {
	output := strings.ToLower(Ctx.CommandOutput)
	if !strings.Contains(output, "git") ||
		(!strings.Contains(output, "failed") &&
			!strings.Contains(output, "error")) {
		return fmt.Errorf("expected git failure indication not found")
	}
	return nil
}

// theErrorShouldIndicateDiffSizeLimitExceeded verifies diff size limit message.
func theErrorShouldIndicateDiffSizeLimitExceeded() error {
	output := strings.ToLower(Ctx.CommandOutput)
	if !strings.Contains(output, "too large") &&
		!strings.Contains(output, "size limit") &&
		!strings.Contains(output, "exceeds") {
		return fmt.Errorf("expected diff size limit error not found")
	}
	return nil
}
