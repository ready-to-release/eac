//go:build L1 && ov
// +build L1,ov

package commitmessage

import (
	"testing"
)

func TestVerifyCommitMessageContract_ValidMessage(t *testing.T) {
	validMessage := `feat(eac-commands): add commit message verifier

Auditor-Summary: Implemented contract verifier to validate commit
messages against formal structure requirements.

This commit adds a contract verifier to validate commit messages
against the formal structure defined in the contract. Ensures
compliance with all formatting and structural requirements.

eac-commands
------------
eac-commands: feat: add commit message contract verifier

Implements validation logic to programmatically verify commit messages
against contract rules including heading format, line length, module
section structure, and semantic subject line formatting.

` + "```go" + `
+func VerifyCommitMessageContract(msg string, nil) []ValidationError {
+	var errors []ValidationError
+	// Validation logic here
+	return errors
+}
` + "```" + `

` + "```yaml" + `
paths:
  - 'go/eac/commands/**'
` + "```" + `

`

	errors := VerifyCommitMessageContract(validMessage, []string{"eac-commands"})

	if len(errors) > 0 {
		t.Errorf("Expected valid message to have no errors, got %d errors:", len(errors))
		for _, err := range errors {
			t.Logf("  - %s", err.Error())
		}
	}
}

func TestVerifyCommitMessageContract_MissingTopHeading(t *testing.T) {
	invalidMessage := `This is not a heading

Auditor-Summary: Some summary.

Some summary text.
`

	errors := VerifyCommitMessageContract(invalidMessage, nil)

	foundError := false
	for _, err := range errors {
		if err.Code == "INVALID_HEADER_FORMAT" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected INVALID_HEADER_FORMAT error")
	}
}

func TestVerifyCommitMessageContract_TitleTooLong(t *testing.T) {
	longTitle := `feat(cli): this is a very long title that exceeds the maximum allowed length`

	errors := VerifyCommitMessageContract(longTitle, nil)

	foundError := false
	for _, err := range errors {
		if err.Code == "HEADER_TOO_LONG" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected HEADER_TOO_LONG error")
	}
}

func TestVerifyCommitMessageContract_TitleTrailingPeriod(t *testing.T) {
	invalidMessage := `feat(cli): add feature.

Auditor-Summary: Added new feature.

Summary text here.
`

	errors := VerifyCommitMessageContract(invalidMessage, nil)

	foundError := false
	for _, err := range errors {
		if err.Code == "HEADER_TRAILING_PERIOD" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected HEADER_TRAILING_PERIOD error")
	}
}

func TestVerifyCommitMessageContract_MissingTopLevelBody(t *testing.T) {
	invalidMessage := `feat(cli): add feature

cli
---

cli: feat: add some feature

Body text here.
`

	errors := VerifyCommitMessageContract(invalidMessage, nil)

	foundError := false
	for _, err := range errors {
		if err.Code == "MISSING_TOP_LEVEL_BODY" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected MISSING_TOP_LEVEL_BODY error")
	}
}

func TestVerifyCommitMessageContract_InvalidSubjectFormat(t *testing.T) {
	invalidMessage := `feat(eac-commands): add feature

Auditor-Summary: Summary text for the overall change.

Summary text for the overall change.

eac-commands
------------

This is not a valid subject line format

Body text here.

`

	errors := VerifyCommitMessageContract(invalidMessage, []string{"eac-commands"})

	foundError := false
	for _, err := range errors {
		if err.Code == "INVALID_SUBJECT_FORMAT" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected INVALID_SUBJECT_FORMAT error")
	}
}

func TestVerifyCommitMessageContract_UnclosedCodeBlock(t *testing.T) {
	invalidMessage := `feat(eac-commands): add feature

Auditor-Summary: Summary text for the overall change.

Summary text for the overall change.

eac-commands
------------

eac-commands: feat: add feature

Body text.

` + "```go" + `
code here
// Missing closing fence

`

	errors := VerifyCommitMessageContract(invalidMessage, nil)

	foundError := false
	for _, err := range errors {
		if err.Code == "UNCLOSED_CODE_BLOCK" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected UNCLOSED_CODE_BLOCK error")
	}
}

func TestVerifyCommitMessageContract_LineTooLong(t *testing.T) {
	invalidMessage := `feat(module): add feature

Auditor-Summary: Adding new feature to module.

This is a very long line that exceeds the maximum allowed length of 72 characters for body text in commit messages.

Changes: 1 file, +10 insertions, -0 deletions

`

	errors := VerifyCommitMessageContract(invalidMessage, nil)

	foundError := false
	for _, err := range errors {
		if err.Code == "LINE_TOO_LONG" && err.Severity == "warning" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected LINE_TOO_LONG warning")
	}
}

func TestVerifyCommitMessageContract_SectionSeparatorsNotOrphaned(t *testing.T) {
	// This test verifies that --- separators between module sections
	// are not flagged as orphaned dashes lines
	validMessage := `feat(multi-module): enhance commit generation with validation

Auditor-Summary: Integrated contract-based validation into commit
message generation with anti-corruption filtering.

This commit enhances the AI-powered commit message generation with a
comprehensive validation framework.

claude-agents
---------
claude-agents: docs: add MCP command usage guidelines during boot

Added documentation clarifying that data-heavy MCP commands should not
be called during initialization.

---

claude
---------
claude: docs: update MCP command usage guidance

Updated agent.md to clarify that get-files should only be called
on-demand since it loads significant tokens.

---

docs
---------
docs: docs: add filtering options to get files command

Expanded documentation for the get files command with comprehensive
filtering options.
`

	errors := VerifyCommitMessageContract(validMessage, []string{"claude-agents", "claude", "docs"})

	// Check that we don't have any ORPHANED_DASHES_LINE errors
	for _, err := range errors {
		if err.Code == "ORPHANED_DASHES_LINE" {
			t.Errorf("Section separator (---) incorrectly flagged as orphaned dashes line at line %d: %s", err.Line, err.Message)
		}
	}
}

func TestVerifyCommitMessageContract_OrphanedDashesLineDetected(t *testing.T) {
	// This test verifies that actual orphaned dashes lines (not ---)
	// are still detected correctly
	invalidMessage := `feat(eac-commands): add feature

Auditor-Summary: Summary text.

Some body text here.

---------

More text here.
`

	errors := VerifyCommitMessageContract(invalidMessage, []string{"eac-commands"})

	foundError := false
	for _, err := range errors {
		if err.Code == "ORPHANED_DASHES_LINE" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected ORPHANED_DASHES_LINE error for dashes without module name")
	}
}
