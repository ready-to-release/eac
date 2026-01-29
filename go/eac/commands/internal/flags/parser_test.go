package flags

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
)

func TestParser_AllSetsSubscribed(t *testing.T) {
	config := CommandFlagConfig{
		Command:   "build",
		Execution: true,
		Output:    true,
		Cache:     true,
		Module:    true,
		DryRun:    true,
	}
	env := &environment.Env{IsLocalConsole: true}
	p := NewParserWithEnv(config, env)

	args := []string{
		"--turbo",
		"--roof", "4",
		"--tui",
		"--debug",
		"--skip-cache",
		"--skip-deps",
		"--exclude", "eac-specs",
		"--dry-run",
		"module1", "module2",
	}

	result, err := p.Parse(args)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	// Check execution flags
	if result.Execution == nil {
		t.Fatal("Execution flags not set")
	}
	if !result.Execution.Turbo {
		t.Error("Turbo should be true")
	}
	if result.Execution.Roof != 4 {
		t.Errorf("Roof = %v, want 4", result.Execution.Roof)
	}

	// Check output flags
	if result.Output == nil {
		t.Fatal("Output flags not set")
	}
	if !result.Output.UseTUI {
		t.Error("UseTUI should be true")
	}
	if !result.Output.Debug {
		t.Error("Debug should be true")
	}

	// Check cache flags
	if result.Cache == nil {
		t.Fatal("Cache flags not set")
	}
	if !result.Cache.SkipCache {
		t.Error("SkipCache should be true")
	}
	if !result.Cache.SkipDeps {
		t.Error("SkipDeps should be true")
	}

	// Check module flags
	if result.Module == nil {
		t.Fatal("Module flags not set")
	}
	if result.Module.Exclude != "eac-specs" {
		t.Errorf("Exclude = %v, want eac-specs", result.Module.Exclude)
	}

	// Check dryrun flags
	if result.DryRun == nil {
		t.Fatal("DryRun flags not set")
	}
	if !result.DryRun.DryRun {
		t.Error("DryRun should be true")
	}

	// Check positional arguments
	if len(result.Positional) != 2 {
		t.Errorf("Positional = %v, want [module1, module2]", result.Positional)
	} else {
		if result.Positional[0] != "module1" || result.Positional[1] != "module2" {
			t.Errorf("Positional = %v, want [module1, module2]", result.Positional)
		}
	}
}

func TestParser_PartialSubscription(t *testing.T) {
	config := CommandFlagConfig{
		Command:   "lint",
		Execution: true,
		Output:    true,
		Cache:     false, // Not subscribed
		Module:    true,
		DryRun:    false, // Not subscribed
	}
	env := &environment.Env{IsLocalConsole: true}
	p := NewParserWithEnv(config, env)

	args := []string{
		"--turbo",
		"--skip-cache", // Should remain as unknown
		"--dry-run",    // Should remain as unknown
		"module1",
	}

	result, err := p.Parse(args)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	// Subscribed sets should be populated
	if result.Execution == nil {
		t.Fatal("Execution flags not set")
	}
	if !result.Execution.Turbo {
		t.Error("Turbo should be true")
	}

	// Unsubscribed sets should be nil
	if result.Cache != nil {
		t.Error("Cache should be nil when not subscribed")
	}
	if result.DryRun != nil {
		t.Error("DryRun should be nil when not subscribed")
	}

	// Unknown flags should be in Remaining
	if len(result.Remaining) != 2 {
		t.Errorf("Remaining = %v, want [--skip-cache, --dry-run]", result.Remaining)
	}

	// Positional should have module1
	if len(result.Positional) != 1 || result.Positional[0] != "module1" {
		t.Errorf("Positional = %v, want [module1]", result.Positional)
	}
}

func TestParser_NoSubscriptions(t *testing.T) {
	config := CommandFlagConfig{
		Command: "simple",
	}
	env := &environment.Env{IsLocalConsole: true}
	p := NewParserWithEnv(config, env)

	args := []string{"--unknown", "arg1", "arg2"}

	result, err := p.Parse(args)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if len(result.Remaining) != 1 || result.Remaining[0] != "--unknown" {
		t.Errorf("Remaining = %v, want [--unknown]", result.Remaining)
	}
	if len(result.Positional) != 2 {
		t.Errorf("Positional = %v, want [arg1, arg2]", result.Positional)
	}
}

func TestParser_TUIValidationError(t *testing.T) {
	config := CommandFlagConfig{
		Command: "build",
		Output:  true,
	}
	// CI environment rejects explicit --tui
	env := &environment.Env{IsCI: true}
	p := NewParserWithEnv(config, env)

	args := []string{"--tui"}

	_, err := p.Parse(args)
	if err == nil {
		t.Error("Parse() expected error for --tui in CI environment")
	}
}

func TestParser_FlagParsingError(t *testing.T) {
	config := CommandFlagConfig{
		Command:   "build",
		Execution: true,
	}
	env := &environment.Env{IsLocalConsole: true}
	p := NewParserWithEnv(config, env)

	args := []string{"--roof", "invalid"}

	_, err := p.Parse(args)
	if err == nil {
		t.Error("Parse() expected error for invalid --roof value")
	}
}

func TestParser_AppliesDefaults(t *testing.T) {
	config := CommandFlagConfig{
		Command: "build",
		Output:  true,
	}
	// Local console defaults to TUI
	env := &environment.Env{IsLocalConsole: true}
	p := NewParserWithEnv(config, env)

	result, err := p.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if !result.Output.UseTUI {
		t.Error("UseTUI should default to true in local console")
	}
	if result.Output.TUIHeight != tui.DefaultHeight {
		t.Errorf("TUIHeight = %v, want %v", result.Output.TUIHeight, tui.DefaultHeight)
	}
}

func TestParser_AllFlags(t *testing.T) {
	config := CommandFlagConfig{
		Command:   "build",
		Execution: true,
		Output:    true,
		Cache:     true,
		Module:    true,
		DryRun:    true,
	}
	p := NewParserWithEnv(config, &environment.Env{})

	flags := p.AllFlags()

	// Count expected flags:
	// Execution: turbo, roof (2)
	// Output: tui, no-tui, tui-height, ascii, skip-tui-delay, debug, timings (7)
	// Cache: skip-cache, skip-deps (2)
	// Module: exclude, skip-depm (2)
	// DryRun: dry-run (1)
	// Total: 14
	expectedCount := 14
	if len(flags) != expectedCount {
		t.Errorf("AllFlags() returned %d flags, want %d", len(flags), expectedCount)
	}

	// Verify some key flags are present
	flagNames := make(map[string]bool)
	for _, f := range flags {
		flagNames[f.Name] = true
	}

	expected := []string{"turbo", "roof", "tui", "debug", "skip-cache", "exclude", "dry-run"}
	for _, name := range expected {
		if !flagNames[name] {
			t.Errorf("AllFlags() missing %s flag", name)
		}
	}
}

func TestParser_GenerateUsage(t *testing.T) {
	config := CommandFlagConfig{
		Command:   "build",
		Execution: true,
		Output:    true,
	}
	p := NewParserWithEnv(config, &environment.Env{})

	usage := p.GenerateUsage()

	// Should contain section headers (using descriptions)
	if !contains(usage, "parallel execution") {
		t.Errorf("Usage should contain execution section, got: %s", usage)
	}
	if !contains(usage, "terminal user interface") {
		t.Errorf("Usage should contain output section, got: %s", usage)
	}

	// Should contain flag names
	if !contains(usage, "--turbo") {
		t.Error("Usage should contain --turbo")
	}
	if !contains(usage, "--tui") {
		t.Error("Usage should contain --tui")
	}
	if !contains(usage, "--debug") {
		t.Error("Usage should contain --debug")
	}

	// Should show short form
	if !contains(usage, "-d") {
		t.Error("Usage should contain -d shorthand for debug")
	}
}

func TestParser_MixedArgsAndFlags(t *testing.T) {
	config := CommandFlagConfig{
		Command:   "test",
		Execution: true,
		Cache:     true,
	}
	env := &environment.Env{IsLocalConsole: true}
	p := NewParserWithEnv(config, env)

	// Args mixed with flags
	args := []string{
		"module1",
		"--turbo",
		"module2",
		"--skip-cache",
		"module3",
	}

	result, err := p.Parse(args)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if !result.Execution.Turbo {
		t.Error("Turbo should be true")
	}
	if !result.Cache.SkipCache {
		t.Error("SkipCache should be true")
	}

	// All positional arguments should be captured
	if len(result.Positional) != 3 {
		t.Errorf("Positional = %v, want 3 modules", result.Positional)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
