package commitmessage

import (
	"testing"
)

func TestVerifyCommitMessageContract_ValidMessage(t *testing.T) {
	validMessage := `# src-commands: feat: add commit message verifier

This commit adds a contract verifier to validate commit messages
against the formal structure defined in the contract. Ensures
compliance with all formatting and structural requirements.

## src-commands

src-commands: feat: add commit message contract verifier

Implements validation logic to programmatically verify commit messages
against contract rules including heading format, line length, module
section structure, and semantic subject line formatting.

` + "```go" + `
+func VerifyCommitMessageContract(msg string) []ValidationError {
+	var errors []ValidationError
+	// Validation logic here
+	return errors
+}
` + "```" + `

` + "```yaml" + `
paths:
  - 'src/commands/**'
` + "```" + `

---

`

	errors := VerifyCommitMessageContract(validMessage)

	if len(errors) > 0 {
		t.Errorf("Expected valid message to have no errors, got %d errors:", len(errors))
		for _, err := range errors {
			t.Logf("  - %s", err.Error())
		}
	}
}

func TestVerifyCommitMessageContract_MissingTopHeading(t *testing.T) {
	invalidMessage := `This is not a heading

## Summary

Some summary text.
`

	errors := VerifyCommitMessageContract(invalidMessage)

	foundError := false
	for _, err := range errors {
		if err.Code == "MISSING_TOP_HEADING" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected MISSING_TOP_HEADING error")
	}
}

func TestVerifyCommitMessageContract_TitleTooLong(t *testing.T) {
	longTitle := `# cli: feat: this is a very long title that exceeds the maximum allowed length of 72 characters`

	errors := VerifyCommitMessageContract(longTitle)

	foundError := false
	for _, err := range errors {
		if err.Code == "TITLE_TOO_LONG" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected TITLE_TOO_LONG error")
	}
}

func TestVerifyCommitMessageContract_TitleTrailingPeriod(t *testing.T) {
	invalidMessage := `# cli: feat: add feature.

## Summary

Summary text here.
`

	errors := VerifyCommitMessageContract(invalidMessage)

	foundError := false
	for _, err := range errors {
		if err.Code == "TITLE_TRAILING_PERIOD" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected TITLE_TRAILING_PERIOD error")
	}
}

func TestVerifyCommitMessageContract_MissingTopLevelBody(t *testing.T) {
	invalidMessage := `# cli: feat: add feature

## cli

cli: feat: add some feature

Body text here.
`

	errors := VerifyCommitMessageContract(invalidMessage)

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
	invalidMessage := `# src-commands: feat: add feature

Summary text for the overall change.

## src-commands

This is not a valid subject line format

Body text here.

---

`

	errors := VerifyCommitMessageContract(invalidMessage)

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
	invalidMessage := `# src-commands: feat: add feature

Summary text for the overall change.

## src-commands

src-commands: feat: add feature

Body text.

` + "```go" + `
code here
// Missing closing fence

---

`

	errors := VerifyCommitMessageContract(invalidMessage)

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

func TestVerifyCommitMessageContract_ModuleHeaderWithColon(t *testing.T) {
	invalidMessage := `# Add feature

## Summary

Summary text.

## Files affected

| Status | File | Module |
| ------ | ---- | ------ |
| added  | file.go | mod |

---

## src-commands: feat: something

src-commands: feat: add feature

Body text.

---

`

	errors := VerifyCommitMessageContract(invalidMessage)

	foundError := false
	for _, err := range errors {
		if err.Code == "MODULE_HEADER_FORMAT" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected MODULE_HEADER_FORMAT error")
	}
}

func TestVerifyCommitMessageContract_LineTooLong(t *testing.T) {
	invalidMessage := `# Add feature

## Summary

This is a very long line that exceeds the maximum allowed length of 72 characters for body text in commit messages.

## Files affected

| Status | File | Module |
| ------ | ---- | ------ |
| added  | file.go | mod |

---

`

	errors := VerifyCommitMessageContract(invalidMessage)

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
