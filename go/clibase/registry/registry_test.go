package registry

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseParentCommandMetadata tests parsing of parent command metadata from comments.
func TestParseParentCommandMetadata(t *testing.T) {
	// Create a temp file with parent command metadata
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_command.go")

	content := `// Command: get
// Short: Retrieve repository data in structured format
// IsParent: true
// Group.Configuration: config
// Group.Repository Structure: modules, dependencies, execution-order
// Group.Files and Changes: files, changed-modules
// Example: r2r get modules
// Example: r2r get dependencies --format=yaml
package test
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	metadata := extractCommandMetadata(testFile)

	// Test basic fields
	if metadata.CommandName != "get" {
		t.Errorf("Expected CommandName 'get', got %q", metadata.CommandName)
	}
	if metadata.Short != "Retrieve repository data in structured format" {
		t.Errorf("Expected Short description, got %q", metadata.Short)
	}

	// Test IsParent
	if !metadata.IsParent {
		t.Error("Expected IsParent to be true")
	}

	// Test Groups
	expectedGroups := map[string][]string{
		"Configuration":        {"config"},
		"Repository Structure": {"modules", "dependencies", "execution-order"},
		"Files and Changes":    {"files", "changed-modules"},
	}
	if len(metadata.Groups) != len(expectedGroups) {
		t.Errorf("Expected %d groups, got %d", len(expectedGroups), len(metadata.Groups))
	}
	for name, expected := range expectedGroups {
		got, ok := metadata.Groups[name]
		if !ok {
			t.Errorf("Missing group %q", name)
			continue
		}
		if len(got) != len(expected) {
			t.Errorf("Group %q: expected %d subcommands, got %d", name, len(expected), len(got))
			continue
		}
		for i, sub := range expected {
			if got[i] != sub {
				t.Errorf("Group %q[%d]: expected %q, got %q", name, i, sub, got[i])
			}
		}
	}

	// Test Examples
	expectedExamples := []string{
		"r2r get modules",
		"r2r get dependencies --format=yaml",
	}
	if len(metadata.Examples) != len(expectedExamples) {
		t.Errorf("Expected %d examples, got %d", len(expectedExamples), len(metadata.Examples))
	}
	for i, expected := range expectedExamples {
		if i < len(metadata.Examples) && metadata.Examples[i] != expected {
			t.Errorf("Example[%d]: expected %q, got %q", i, expected, metadata.Examples[i])
		}
	}
}

// TestParseLeafCommand tests that leaf commands (non-parent) work correctly.
func TestParseLeafCommand(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "leaf_command.go")

	content := `// Command: get modules
// Short: Get all module contracts
// Flag.format: type=string, default=json, usage=Output format (json|yaml|table)
// Flag.debug: type=bool, shorthand=d, usage=Enable debug output
package test
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	metadata := extractCommandMetadata(testFile)

	if metadata.CommandName != "get modules" {
		t.Errorf("Expected CommandName 'get modules', got %q", metadata.CommandName)
	}
	if metadata.IsParent {
		t.Error("Expected IsParent to be false for leaf command")
	}
	if len(metadata.Groups) != 0 {
		t.Errorf("Expected no groups for leaf command, got %d", len(metadata.Groups))
	}
}

// TestSubcommandGroup tests the SubcommandGroup type.
func TestSubcommandGroup(t *testing.T) {
	group := SubcommandGroup{
		Name:        "Configuration",
		Subcommands: []string{"config", "settings"},
	}

	if group.Name != "Configuration" {
		t.Errorf("Expected Name 'Configuration', got %q", group.Name)
	}
	if len(group.Subcommands) != 2 {
		t.Errorf("Expected 2 subcommands, got %d", len(group.Subcommands))
	}
}

// TestCommandRegistrationParentFields tests that parent fields are stored in registration.
func TestCommandRegistrationParentFields(t *testing.T) {
	reg := &CommandRegistration{
		ActualCommand: "get",
		CanonicalName: "get",
		Short:         "Retrieve repository data",
		IsParent:      true,
		Subcommands: []SubcommandGroup{
			{Name: "Configuration", Subcommands: []string{"config"}},
		},
		Examples: []string{"r2r get modules"},
	}

	if !reg.IsParent {
		t.Error("Expected IsParent to be true")
	}
	if len(reg.Subcommands) != 1 {
		t.Errorf("Expected 1 subcommand group, got %d", len(reg.Subcommands))
	}
	if len(reg.Examples) != 1 {
		t.Errorf("Expected 1 example, got %d", len(reg.Examples))
	}
}

// TestIsParentFalseByDefault tests that IsParent defaults to false.
func TestIsParentFalseByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "simple_command.go")

	content := `// Command: build
// Short: Build the project
package test
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	metadata := extractCommandMetadata(testFile)

	if metadata.IsParent {
		t.Error("Expected IsParent to be false by default")
	}
}

// TestGroupOrderPreservation tests that group order is preserved.
func TestGroupOrderPreservation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "ordered_groups.go")

	content := `// Command: get
// IsParent: true
// Group.First: a, b
// Group.Second: c, d
// Group.Third: e, f
package test
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	metadata := extractCommandMetadata(testFile)

	expectedOrder := []string{"First", "Second", "Third"}
	if len(metadata.GroupOrder) != len(expectedOrder) {
		t.Errorf("Expected %d group order entries, got %d", len(expectedOrder), len(metadata.GroupOrder))
	}
	for i, expected := range expectedOrder {
		if i < len(metadata.GroupOrder) && metadata.GroupOrder[i] != expected {
			t.Errorf("GroupOrder[%d]: expected %q, got %q", i, expected, metadata.GroupOrder[i])
		}
	}
}

// TestFlagMetadataDeclarativeFields tests the new declarative flag fields.
func TestFlagMetadataDeclarativeFields(t *testing.T) {
	t.Run("IsBehaviorFlag returns true when Behavior is set", func(t *testing.T) {
		flag := FlagMetadata{
			Name:     "with-cache",
			Behavior: "cache",
		}
		if !flag.IsBehaviorFlag() {
			t.Error("Expected IsBehaviorFlag() to return true")
		}
	})

	t.Run("IsBehaviorFlag returns false when Behavior is empty", func(t *testing.T) {
		flag := FlagMetadata{
			Name: "format",
		}
		if flag.IsBehaviorFlag() {
			t.Error("Expected IsBehaviorFlag() to return false")
		}
	})

	t.Run("declarative fields are correctly stored", func(t *testing.T) {
		envDefaults := map[string]bool{
			"local": true,
			"CI":    false,
		}
		flag := FlagMetadata{
			Name:         "with-tui",
			Type:         "bool",
			DefaultValue: "true",
			Usage:        "Enable terminal UI",
			Behavior:     "tui",
			IsEnableFlag: true,
			PairFlagName: "no-tui",
			EnvAware:     true,
			EnvDefaults:  envDefaults,
		}

		if flag.Behavior != "tui" {
			t.Errorf("Expected Behavior 'tui', got %q", flag.Behavior)
		}
		if !flag.IsEnableFlag {
			t.Error("Expected IsEnableFlag to be true")
		}
		if flag.PairFlagName != "no-tui" {
			t.Errorf("Expected PairFlagName 'no-tui', got %q", flag.PairFlagName)
		}
		if !flag.EnvAware {
			t.Error("Expected EnvAware to be true")
		}
		if len(flag.EnvDefaults) != 2 {
			t.Errorf("Expected 2 env defaults, got %d", len(flag.EnvDefaults))
		}
		if !flag.EnvDefaults["local"] {
			t.Error("Expected EnvDefaults['local'] to be true")
		}
		if flag.EnvDefaults["CI"] {
			t.Error("Expected EnvDefaults['CI'] to be false")
		}
	})
}

// TestFlagMetadataBackwardCompatibility tests that existing code works with new fields.
func TestFlagMetadataBackwardCompatibility(t *testing.T) {
	// Existing code should continue to work without setting new fields
	flag := FlagMetadata{
		Name:         "format",
		Shorthand:    "f",
		Type:         "string",
		DefaultValue: "json",
		Usage:        "Output format",
		Required:     false,
	}

	// New fields should have zero values
	if flag.Behavior != "" {
		t.Errorf("Expected empty Behavior, got %q", flag.Behavior)
	}
	if flag.IsEnableFlag {
		t.Error("Expected IsEnableFlag to be false by default")
	}
	if flag.PairFlagName != "" {
		t.Errorf("Expected empty PairFlagName, got %q", flag.PairFlagName)
	}
	if flag.EnvAware {
		t.Error("Expected EnvAware to be false by default")
	}
	if flag.EnvDefaults != nil {
		t.Error("Expected nil EnvDefaults by default")
	}

	// Helper methods should work correctly with zero values
	if flag.IsBehaviorFlag() {
		t.Error("Expected IsBehaviorFlag() to return false for regular flag")
	}
}
