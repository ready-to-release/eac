package gherkin

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/validation"
)

// newTestValidator creates a Validator with minimal configuration suitable for unit tests.
// Uses nil contract (defaults to built-in naming pattern) and empty tags config.
func newTestValidator() *Validator {
	return NewValidator(nil, &config.TestingTagsConfig{})
}

// buildSpec constructs a minimal Gherkin spec by prepending commentPrefix before a
// structurally valid Feature/Rule/Scenario body. Used to isolate Phase 0 detection.
func buildSpec(commentPrefix string) string {
	return commentPrefix + `@deps:go @ov
Feature: eac-cli_intent-test

  Rule: Test rule

    @L2 @ov
    Scenario: Test scenario
      Given a precondition
      When an action occurs
      Then the result is expected
`
}

// hasErrorCode reports whether the given error code appears in the list.
func hasErrorCode(errors []validation.ValidationError, code string) bool {
	for _, e := range errors {
		if e.GetCode() == code {
			return true
		}
	}
	return false
}

// TestPhase0_BothCommentsPresent verifies that no ACD comment errors are emitted
// when both # Intent: and # Architecture: appear before Feature:.
func TestPhase0_BothCommentsPresent(t *testing.T) {
	v := newTestValidator()
	content := buildSpec("# Intent: Test intent for unit test\n# Architecture: Affects validator only\n\n")
	errors := v.Validate(content, nil)

	if hasErrorCode(errors, "MISSING_INTENT_COMMENT") {
		t.Error("unexpected MISSING_INTENT_COMMENT when # Intent: is present")
	}
	if hasErrorCode(errors, "MISSING_ARCHITECTURE_COMMENT") {
		t.Error("unexpected MISSING_ARCHITECTURE_COMMENT when # Architecture: is present")
	}
}

// TestPhase0_MissingIntentComment verifies MISSING_INTENT_COMMENT is emitted
// when # Architecture: is present but # Intent: is absent.
func TestPhase0_MissingIntentComment(t *testing.T) {
	v := newTestValidator()
	content := buildSpec("# Architecture: Affects validator only\n\n")
	errors := v.Validate(content, nil)

	if !hasErrorCode(errors, "MISSING_INTENT_COMMENT") {
		t.Error("expected MISSING_INTENT_COMMENT when # Intent: is absent")
	}
	if hasErrorCode(errors, "MISSING_ARCHITECTURE_COMMENT") {
		t.Error("unexpected MISSING_ARCHITECTURE_COMMENT when # Architecture: is present")
	}
}

// TestPhase0_MissingArchitectureComment verifies MISSING_ARCHITECTURE_COMMENT is emitted
// when # Intent: is present but # Architecture: is absent.
func TestPhase0_MissingArchitectureComment(t *testing.T) {
	v := newTestValidator()
	content := buildSpec("# Intent: Test intent for unit test\n\n")
	errors := v.Validate(content, nil)

	if hasErrorCode(errors, "MISSING_INTENT_COMMENT") {
		t.Error("unexpected MISSING_INTENT_COMMENT when # Intent: is present")
	}
	if !hasErrorCode(errors, "MISSING_ARCHITECTURE_COMMENT") {
		t.Error("expected MISSING_ARCHITECTURE_COMMENT when # Architecture: is absent")
	}
}

// TestPhase0_BothCommentsMissing verifies both ACD error codes are emitted
// when neither comment is present.
func TestPhase0_BothCommentsMissing(t *testing.T) {
	v := newTestValidator()
	content := buildSpec("")
	errors := v.Validate(content, nil)

	if !hasErrorCode(errors, "MISSING_INTENT_COMMENT") {
		t.Error("expected MISSING_INTENT_COMMENT when # Intent: is absent")
	}
	if !hasErrorCode(errors, "MISSING_ARCHITECTURE_COMMENT") {
		t.Error("expected MISSING_ARCHITECTURE_COMMENT when # Architecture: is absent")
	}
}

// TestPhase0_IntentWithWhitespaceOnly verifies that "# Intent:   " (whitespace only)
// is treated as missing, not as valid content.
func TestPhase0_IntentWithWhitespaceOnly(t *testing.T) {
	v := newTestValidator()
	content := buildSpec("# Intent:   \n# Architecture: Affects validator only\n\n")
	errors := v.Validate(content, nil)

	if !hasErrorCode(errors, "MISSING_INTENT_COMMENT") {
		t.Error("expected MISSING_INTENT_COMMENT when # Intent: has only whitespace content")
	}
}

// TestPhase0_ArchitectureWithWhitespaceOnly verifies that "# Architecture:   " (whitespace only)
// is treated as missing, not as valid content.
func TestPhase0_ArchitectureWithWhitespaceOnly(t *testing.T) {
	v := newTestValidator()
	content := buildSpec("# Intent: Test intent for unit test\n# Architecture:   \n\n")
	errors := v.Validate(content, nil)

	if !hasErrorCode(errors, "MISSING_ARCHITECTURE_COMMENT") {
		t.Error("expected MISSING_ARCHITECTURE_COMMENT when # Architecture: has only whitespace content")
	}
}

// TestPhase0_IntentAfterFeature verifies that # Intent: appearing after Feature:
// is ignored by the !state.seenFeature guard and triggers MISSING_INTENT_COMMENT.
func TestPhase0_IntentAfterFeature(t *testing.T) {
	v := newTestValidator()
	content := `@deps:go @ov
Feature: eac-cli_intent-test

  # Intent: This appears too late — after the Feature: declaration
  # Architecture: This also appears too late

  Rule: Test rule

    @L2 @ov
    Scenario: Test scenario
      Given a precondition
`
	errors := v.Validate(content, nil)

	if !hasErrorCode(errors, "MISSING_INTENT_COMMENT") {
		t.Error("expected MISSING_INTENT_COMMENT when # Intent: appears after Feature:")
	}
	if !hasErrorCode(errors, "MISSING_ARCHITECTURE_COMMENT") {
		t.Error("expected MISSING_ARCHITECTURE_COMMENT when # Architecture: appears after Feature:")
	}
}

// TestPhase0_DuplicateIntent_FirstValidSecondEmpty verifies that a valid first # Intent:
// is not overwritten by a subsequent empty # Intent: (the "last-wins" bug fixed by I5).
func TestPhase0_DuplicateIntent_FirstValidSecondEmpty(t *testing.T) {
	v := newTestValidator()
	// First occurrence has content (valid). Second is empty (would invalidate if not guarded).
	content := buildSpec("# Intent: Valid intent here\n# Intent:   \n# Architecture: Affects validator only\n\n")
	errors := v.Validate(content, nil)

	if hasErrorCode(errors, "MISSING_INTENT_COMMENT") {
		t.Error("expected no MISSING_INTENT_COMMENT: first valid # Intent: must not be overwritten by a subsequent empty occurrence")
	}
}

// TestPhase0_DuplicateArchitecture_FirstValidSecondEmpty verifies the same guard for
// # Architecture: duplicate lines.
func TestPhase0_DuplicateArchitecture_FirstValidSecondEmpty(t *testing.T) {
	v := newTestValidator()
	content := buildSpec("# Intent: Test intent for unit test\n# Architecture: Affects validator only\n# Architecture:   \n\n")
	errors := v.Validate(content, nil)

	if hasErrorCode(errors, "MISSING_ARCHITECTURE_COMMENT") {
		t.Error("expected no MISSING_ARCHITECTURE_COMMENT: first valid # Architecture: must not be overwritten by a subsequent empty occurrence")
	}
}
