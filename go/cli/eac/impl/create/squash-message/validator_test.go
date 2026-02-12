package squashmessage

import (
	"strings"
	"testing"
)

func TestValidate_ValidMessage(t *testing.T) {
	v := NewSquashMessageValidator()
	msg := "feat(auth): add OAuth2 support\n\nDetailed body text."

	errors := v.Validate(msg, nil)
	if len(errors) != 0 {
		t.Errorf("expected no errors for valid message, got %d: %v", len(errors), errors)
	}
}

func TestValidate_EmptyMessage(t *testing.T) {
	v := NewSquashMessageValidator()

	errors := v.Validate("", nil)
	if len(errors) == 0 {
		t.Error("expected error for empty message")
	}
}

func TestValidate_BlankFirstLine(t *testing.T) {
	v := NewSquashMessageValidator()

	errors := v.Validate("   \nsome body", nil)
	if len(errors) == 0 {
		t.Error("expected error for blank first line")
	}
}

func TestValidate_MissingColon(t *testing.T) {
	v := NewSquashMessageValidator()
	msg := "feat(auth) add OAuth2 support"

	errors := v.Validate(msg, nil)
	found := false
	for _, e := range errors {
		if strings.Contains(e.Message, "format") {
			found = true
		}
	}
	if !found {
		t.Error("expected format error for missing colon in header")
	}
}

func TestValidate_HeaderTooLong(t *testing.T) {
	v := NewSquashMessageValidator()
	// Create a header longer than 72 chars
	msg := "feat(auth): " + strings.Repeat("x", 70)

	errors := v.Validate(msg, nil)
	found := false
	for _, e := range errors {
		if strings.Contains(e.Message, "72") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected length error for long header (%d chars), got %v", len(strings.SplitN(msg, "\n", 2)[0]), errors)
	}
}

func TestValidate_HeaderExactly72(t *testing.T) {
	v := NewSquashMessageValidator()
	// Build exactly 72-char header: "feat(x): " is 9 chars, need 63 more
	msg := "feat(x): " + strings.Repeat("a", 63)
	if len(msg) != 72 {
		t.Fatalf("test setup error: header is %d chars, expected 72", len(msg))
	}

	errors := v.Validate(msg, nil)
	for _, e := range errors {
		if strings.Contains(e.Message, "72") {
			t.Error("header at exactly 72 chars should not trigger length error")
		}
	}
}

func TestVerifyImplementation(t *testing.T) {
	v := NewSquashMessageValidator()
	errors := v.VerifyImplementation()
	if errors != nil {
		t.Errorf("expected nil from VerifyImplementation, got %v", errors)
	}
}
