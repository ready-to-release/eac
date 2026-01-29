package tui

import (
	"os"
	"testing"
)

func TestShouldUseTUI_HelpRequest(t *testing.T) {
	// --help should never use TUI
	if ShouldUseTUI("build", false, true) {
		t.Error("ShouldUseTUI should return false for help requests")
	}
}

func TestShouldUseTUI_CIEnvironment(t *testing.T) {
	// Simulate CI environment
	oldCI := os.Getenv("CI")
	os.Setenv("CI", "true")
	defer func() {
		if oldCI == "" {
			os.Unsetenv("CI")
		} else {
			os.Setenv("CI", oldCI)
		}
	}()

	if ShouldUseTUI("build", false, false) {
		t.Error("ShouldUseTUI should return false in CI environment")
	}
}

func TestResolveBinding_Exact(t *testing.T) {
	// Set up test registry
	r := NewRegistry()
	r.Register("build", mockFactory("build"))
	r.Register("*", mockFactory("default"))

	// Temporarily replace global registry for this test
	oldRegistry := globalRegistry
	globalRegistry = r
	defer func() { globalRegistry = oldRegistry }()

	// build has exact binding
	binding := resolveBinding("build")
	if binding == nil {
		t.Fatal("Expected binding for 'build', got nil")
	}
	if binding.IsDefault {
		t.Error("'build' should not be default binding")
	}
	if binding.Pattern != "build" {
		t.Errorf("Expected pattern 'build', got %q", binding.Pattern)
	}
}

func TestResolveBinding_Default(t *testing.T) {
	// Set up test registry
	r := NewRegistry()
	r.Register("build", mockFactory("build"))
	r.Register("*", mockFactory("default"))

	// Temporarily replace global registry for this test
	oldRegistry := globalRegistry
	globalRegistry = r
	defer func() { globalRegistry = oldRegistry }()

	// unknown command should get default
	binding := resolveBinding("unknown")
	if binding == nil {
		t.Fatal("Expected default binding, got nil")
	}
	if !binding.IsDefault {
		t.Error("'unknown' should use default binding")
	}
	if binding.Pattern != "*" {
		t.Errorf("Expected pattern '*', got %q", binding.Pattern)
	}
}
