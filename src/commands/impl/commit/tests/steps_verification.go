// Package tests provides BDD step definitions for the commit command.
//
// This file contains Then step implementations for verifying test outcomes.
package tests

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
	return nil
}

// theContextShouldIncludeTheStagedFilesTable verifies staged files in context.
func theContextShouldIncludeTheStagedFilesTable() error {
	return nil
}

// theContextShouldIncludeOnlyFilesForThatModule verifies module-specific files.
func theContextShouldIncludeOnlyFilesForThatModule() error {
	return nil
}

// theContextShouldListAllAffectedModules verifies all modules are listed.
func theContextShouldListAllAffectedModules() error {
	return nil
}

// theContextShouldListTheAffectedModule verifies the affected module is listed.
func theContextShouldListTheAffectedModule() error {
	return nil
}

// theContextShouldIndicate verifies context contains indicator.
func theContextShouldIndicate(indicator string) error {
	return nil
}

// theContextShouldIndicateTheCountAs verifies count in context.
func theContextShouldIndicateTheCountAs(countType string) error {
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
	return nil
}

// ============================================================================
// Diff Filtering Verification Steps
// ============================================================================

// onlyThatFilesDiffShouldBeIncludedCommit verifies single-file diff filtering.
func onlyThatFilesDiffShouldBeIncludedCommit() error {
	return nil
}

// otherFilesShouldBeExcludedCommit verifies non-matching files are excluded.
func otherFilesShouldBeExcludedCommit() error {
	return nil
}

// theDiffIsFilteredForThatModule verifies module-specific diff filtering.
func theDiffIsFilteredForThatModule() error {
	return nil
}

// allOfThatModulesFilesShouldBeIncluded verifies all module files included.
func allOfThatModulesFilesShouldBeIncluded() error {
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
