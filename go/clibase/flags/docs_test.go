package flags

import (
	"strings"
	"testing"
)

func TestFlagDocGenerator_GenerateFlagMatrix(t *testing.T) {
	gen := NewFlagDocGenerator()
	doc := gen.GenerateFlagMatrix()

	// Should have main header
	if !strings.Contains(doc, "# Shared Flags") {
		t.Error("Document should contain main header")
	}

	// Should have all section headers
	sections := []string{
		"## Execution Control",
		"## Output Control",
		"## Cache Control",
		"## Module Selection",
		"## Dry Run",
		"## Command Availability Matrix",
		"## Command-Specific Flags",
	}
	for _, section := range sections {
		if !strings.Contains(doc, section) {
			t.Errorf("Document should contain section: %s", section)
		}
	}

	// Should have flag tables
	flags := []string{
		"`--turbo`",
		"`--roof`",
		"`--with-tui`",
		"`--debug`",
		"`--skip-cache`",
		"`--no-deps`",
		"`--exclude`",
		"`--skip-depm`",
		"`--dry-run`",
	}
	for _, flag := range flags {
		if !strings.Contains(doc, flag) {
			t.Errorf("Document should contain flag: %s", flag)
		}
	}

	// Should have command-specific flags
	commandFlags := []string{
		"`--tags`",
		"`--list-only`",
		"`--with-tidy`",
	}
	for _, flag := range commandFlags {
		if !strings.Contains(doc, flag) {
			t.Errorf("Document should contain command-specific flag: %s", flag)
		}
	}

	// Should have skip-depm note
	if !strings.Contains(doc, "not implemented") {
		t.Error("Document should contain skip-depm note")
	}
}

func TestFlagDocGenerator_GetCommandDoc(t *testing.T) {
	gen := NewFlagDocGenerator()

	tests := []struct {
		name      string
		wantFound bool
	}{
		{"build", true},
		{"test", true},
		{"lint", true},
		{"scan", true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := gen.GetCommandDoc(tt.name)
			if tt.wantFound && doc == nil {
				t.Errorf("GetCommandDoc(%s) returned nil, want document", tt.name)
			}
			if !tt.wantFound && doc != nil {
				t.Errorf("GetCommandDoc(%s) returned document, want nil", tt.name)
			}
		})
	}
}

func TestFlagDocGenerator_CommandDocProperties(t *testing.T) {
	gen := NewFlagDocGenerator()

	// Build and test should support skip-depm
	build := gen.GetCommandDoc("build")
	if !build.SkipDepmSupported {
		t.Error("build should support --skip-depm")
	}

	test := gen.GetCommandDoc("test")
	if !test.SkipDepmSupported {
		t.Error("test should support --skip-depm")
	}

	// Lint and scan should NOT support skip-depm
	lint := gen.GetCommandDoc("lint")
	if lint.SkipDepmSupported {
		t.Error("lint should not support --skip-depm")
	}

	scan := gen.GetCommandDoc("scan")
	if scan.SkipDepmSupported {
		t.Error("scan should not support --skip-depm")
	}
}

func TestFlagDocGenerator_GenerateCommandUsage(t *testing.T) {
	gen := NewFlagDocGenerator()

	usage := gen.GenerateCommandUsage("build")

	// Should have usage line
	if !strings.Contains(usage, "Usage: eac build") {
		t.Error("Usage should contain command name")
	}

	// Should have shared flags
	if !strings.Contains(usage, "--turbo") {
		t.Error("Usage should contain shared flags")
	}

	// Should have build-specific flags
	if !strings.Contains(usage, "--with-tidy") {
		t.Error("Usage should contain build-specific flags")
	}
}

func TestFlagDocGenerator_GenerateCommandUsage_Unknown(t *testing.T) {
	gen := NewFlagDocGenerator()

	usage := gen.GenerateCommandUsage("unknown")

	if usage != "" {
		t.Errorf("GenerateCommandUsage(unknown) = %q, want empty string", usage)
	}
}

func TestFlagDocGenerator_AllCommandsSubscribeToAllSets(t *testing.T) {
	gen := NewFlagDocGenerator()

	// All 4 commands should subscribe to all 5 flag sets
	commands := []string{"build", "test", "lint", "scan"}

	for _, name := range commands {
		doc := gen.GetCommandDoc(name)
		if doc == nil {
			t.Fatalf("Command %s not found", name)
		}

		if !doc.Execution {
			t.Errorf("%s should subscribe to Execution", name)
		}
		if !doc.Output {
			t.Errorf("%s should subscribe to Output", name)
		}
		if !doc.Cache {
			t.Errorf("%s should subscribe to Cache", name)
		}
		if !doc.Module {
			t.Errorf("%s should subscribe to Module", name)
		}
		if !doc.DryRun {
			t.Errorf("%s should subscribe to DryRun", name)
		}
	}
}
