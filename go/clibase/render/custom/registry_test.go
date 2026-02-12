//go:build L1 && ov
// +build L1,ov

package custom

import (
	"sort"
	"strings"
	"testing"
)

// ---- Register tests ----

func TestRegister_NewRenderer(t *testing.T) {
	// Register a unique test renderer
	name := "test-register-new"
	called := false
	Register(name, func(yamlBytes []byte) (string, error) {
		called = true
		return "ok", nil
	}, []string{"test-cmd"})

	// Verify it can be retrieved
	renderer, err := Get(name, "test-cmd")
	if err != nil {
		t.Fatalf("Get(%q) returned error after Register: %v", name, err)
	}

	result, err := renderer(nil)
	if err != nil {
		t.Fatalf("renderer returned error: %v", err)
	}
	if result != "ok" {
		t.Errorf("renderer returned %q, want %q", result, "ok")
	}
	if !called {
		t.Error("registered renderer was not called")
	}
}

func TestRegister_DuplicatePanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected Register to panic on duplicate name, but it did not")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected panic with string message, got %T: %v", r, r)
		}
		if !strings.Contains(msg, "already registered") {
			t.Errorf("panic message should mention 'already registered', got: %s", msg)
		}
	}()

	name := "test-duplicate-panic"
	noop := func(yamlBytes []byte) (string, error) { return "", nil }
	Register(name, noop, []string{"*"})
	Register(name, noop, []string{"*"}) // should panic
}

func TestRegister_EmptyCommandsBecomesWildcard(t *testing.T) {
	name := "test-empty-commands"
	Register(name, func(yamlBytes []byte) (string, error) {
		return "wildcard", nil
	}, []string{})

	// Should be accessible from any command
	renderer, err := Get(name, "any-command")
	if err != nil {
		t.Fatalf("Get(%q, 'any-command') returned error: %v", name, err)
	}

	result, _ := renderer(nil)
	if result != "wildcard" {
		t.Errorf("got %q, want %q", result, "wildcard")
	}
}

// ---- Get tests ----

func TestGet_ExistingRenderer(t *testing.T) {
	// "summary" is registered via init() with wildcard commands
	renderer, err := Get("summary", "any-command")
	if err != nil {
		t.Fatalf("Get('summary') returned error: %v", err)
	}
	if renderer == nil {
		t.Fatal("Get('summary') returned nil renderer")
	}
}

func TestGet_NotFound(t *testing.T) {
	_, err := Get("nonexistent-renderer-xyz", "any-command")
	if err == nil {
		t.Error("expected error for nonexistent renderer, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestGet_UnsupportedCommand(t *testing.T) {
	// "count" is registered only for "get-modules"
	_, err := Get("count", "some-other-command")
	if err == nil {
		t.Error("expected error for unsupported command, got nil")
	}
	if !strings.Contains(err.Error(), "does not support") {
		t.Errorf("error should mention 'does not support', got: %v", err)
	}
}

func TestGet_SupportedCommand(t *testing.T) {
	// "count" is registered for "get-modules"
	renderer, err := Get("count", "get-modules")
	if err != nil {
		t.Fatalf("Get('count', 'get-modules') returned error: %v", err)
	}
	if renderer == nil {
		t.Fatal("Get('count', 'get-modules') returned nil renderer")
	}
}

// ---- List tests ----

func TestList_AllRenderers(t *testing.T) {
	// Passing empty string should return all registered renderers
	names := List("")
	if len(names) == 0 {
		t.Fatal("List('') returned no renderers, expected at least 'count' and 'summary'")
	}

	// Verify known renderers are present
	sort.Strings(names)
	foundCount := false
	foundSummary := false
	for _, n := range names {
		if n == "count" {
			foundCount = true
		}
		if n == "summary" {
			foundSummary = true
		}
	}

	if !foundCount {
		t.Errorf("expected 'count' in List(''), got: %v", names)
	}
	if !foundSummary {
		t.Errorf("expected 'summary' in List(''), got: %v", names)
	}
}

func TestList_FilteredByCommand(t *testing.T) {
	// "count" only supports "get-modules"
	// "summary" supports "*" (all commands)
	names := List("get-modules")

	foundCount := false
	foundSummary := false
	for _, n := range names {
		if n == "count" {
			foundCount = true
		}
		if n == "summary" {
			foundSummary = true
		}
	}

	if !foundCount {
		t.Errorf("expected 'count' in List('get-modules'), got: %v", names)
	}
	if !foundSummary {
		t.Errorf("expected 'summary' in List('get-modules') (wildcard), got: %v", names)
	}
}

func TestList_FilteredExcludesRestricted(t *testing.T) {
	// For a command that is NOT "get-modules", "count" should not appear
	names := List("get-files")

	for _, n := range names {
		if n == "count" {
			t.Errorf("'count' should not appear in List('get-files'), got: %v", names)
		}
	}
}

// ---- SupportsCommand tests ----

func TestSupportsCommand(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		cmdName  string
		want     bool
	}{
		{
			name:     "wildcard matches any command",
			commands: []string{"*"},
			cmdName:  "anything",
			want:     true,
		},
		{
			name:     "exact match",
			commands: []string{"get-modules"},
			cmdName:  "get-modules",
			want:     true,
		},
		{
			name:     "no match",
			commands: []string{"get-modules"},
			cmdName:  "get-files",
			want:     false,
		},
		{
			name:     "multiple commands match first",
			commands: []string{"get-modules", "get-files"},
			cmdName:  "get-modules",
			want:     true,
		},
		{
			name:     "multiple commands match second",
			commands: []string{"get-modules", "get-files"},
			cmdName:  "get-files",
			want:     true,
		},
		{
			name:     "multiple commands no match",
			commands: []string{"get-modules", "get-files"},
			cmdName:  "get-specs",
			want:     false,
		},
		{
			name:     "empty commands list does not match",
			commands: []string{},
			cmdName:  "anything",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &RendererRegistration{
				Commands: tt.commands,
			}
			got := reg.SupportsCommand(tt.cmdName)
			if got != tt.want {
				t.Errorf("SupportsCommand(%q) = %v, want %v (commands: %v)",
					tt.cmdName, got, tt.want, tt.commands)
			}
		})
	}
}
