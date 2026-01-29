package flags

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
)

func TestBuildConfig(t *testing.T) {
	cfg := BuildConfig()

	if cfg.Command != "build" {
		t.Errorf("Command = %v, want build", cfg.Command)
	}
	if !cfg.Execution || !cfg.Output || !cfg.Cache || !cfg.Module || !cfg.DryRun {
		t.Error("Build should subscribe to all flag sets")
	}
}

func TestTestConfig(t *testing.T) {
	cfg := TestConfig()

	if cfg.Command != "test" {
		t.Errorf("Command = %v, want test", cfg.Command)
	}
	if !cfg.Execution || !cfg.Output || !cfg.Cache || !cfg.Module || !cfg.DryRun {
		t.Error("Test should subscribe to all flag sets")
	}
}

func TestLintConfig(t *testing.T) {
	cfg := LintConfig()

	if cfg.Command != "lint" {
		t.Errorf("Command = %v, want lint", cfg.Command)
	}
	if !cfg.Execution || !cfg.Output || !cfg.Cache || !cfg.Module || !cfg.DryRun {
		t.Error("Lint should subscribe to all flag sets")
	}
}

func TestScanConfig(t *testing.T) {
	cfg := ScanConfig()

	if cfg.Command != "scan" {
		t.Errorf("Command = %v, want scan", cfg.Command)
	}
	if !cfg.Execution || !cfg.Output || !cfg.Cache || !cfg.Module || !cfg.DryRun {
		t.Error("Scan should subscribe to all flag sets")
	}
}

func TestParseSharedFlags_AllFlags(t *testing.T) {
	env := &environment.Env{IsLocalConsole: true}
	args := []string{
		"--turbo",
		"--roof", "4",
		"--tui",
		"--debug",
		"--timings",
		"--skip-cache",
		"--skip-deps",
		"--exclude", "test-*",
		"--skip-depm",
		"--dry-run",
		"module1", "module2",
	}

	flags, err := ParseSharedFlagsWithEnv(BuildConfig(), args, env)
	if err != nil {
		t.Fatalf("ParseSharedFlags() unexpected error: %v", err)
	}

	// Execution flags
	if !flags.Turbo {
		t.Error("Turbo should be true")
	}
	if flags.MaxConcurrency != 4 {
		t.Errorf("MaxConcurrency = %v, want 4", flags.MaxConcurrency)
	}

	// Output flags
	if !flags.UseTUI {
		t.Error("UseTUI should be true")
	}
	if !flags.Debug {
		t.Error("Debug should be true")
	}
	if !flags.ShowTimings {
		t.Error("ShowTimings should be true")
	}

	// Cache flags
	if !flags.SkipCache {
		t.Error("SkipCache should be true")
	}
	if !flags.SkipDeps {
		t.Error("SkipDeps should be true")
	}

	// Module flags
	if flags.Exclude != "test-*" {
		t.Errorf("Exclude = %v, want test-*", flags.Exclude)
	}
	if !flags.SkipDepm {
		t.Error("SkipDepm should be true")
	}

	// DryRun
	if !flags.DryRun {
		t.Error("DryRun should be true")
	}

	// Monikers
	if len(flags.Monikers) != 2 {
		t.Errorf("Monikers = %v, want [module1, module2]", flags.Monikers)
	}
}

func TestParseSharedFlags_WithRemaining(t *testing.T) {
	env := &environment.Env{IsLocalConsole: true}
	args := []string{
		"--turbo",
		"--tidy-first",  // Build-specific (unknown flag)
		"--version=1.0", // Build-specific (unknown flag)
		"module1",
	}

	flags, err := ParseSharedFlagsWithEnv(BuildConfig(), args, env)
	if err != nil {
		t.Fatalf("ParseSharedFlags() unexpected error: %v", err)
	}

	// Shared flags should be parsed
	if !flags.Turbo {
		t.Error("Turbo should be true")
	}

	// Unknown flags should be in Remaining
	if len(flags.Remaining) != 2 {
		t.Errorf("Remaining = %v, want 2 items (--tidy-first, --version=1.0)", flags.Remaining)
	}

	// Module should be in Monikers
	if len(flags.Monikers) != 1 || flags.Monikers[0] != "module1" {
		t.Errorf("Monikers = %v, want [module1]", flags.Monikers)
	}
}

func TestParseSharedFlags_SkipDepm_Lint_Error(t *testing.T) {
	env := &environment.Env{IsLocalConsole: true}
	args := []string{"--skip-depm"}

	_, err := ParseSharedFlagsWithEnv(LintConfig(), args, env)
	if err == nil {
		t.Error("ParseSharedFlags() expected error for --skip-depm in lint")
	}
}

func TestParseSharedFlags_SkipDepm_Scan_Error(t *testing.T) {
	env := &environment.Env{IsLocalConsole: true}
	args := []string{"--skip-depm"}

	_, err := ParseSharedFlagsWithEnv(ScanConfig(), args, env)
	if err == nil {
		t.Error("ParseSharedFlags() expected error for --skip-depm in scan")
	}
}

func TestParseSharedFlags_SkipDepm_Build_OK(t *testing.T) {
	env := &environment.Env{IsLocalConsole: true}
	args := []string{"--skip-depm"}

	flags, err := ParseSharedFlagsWithEnv(BuildConfig(), args, env)
	if err != nil {
		t.Fatalf("ParseSharedFlags() unexpected error: %v", err)
	}
	if !flags.SkipDepm {
		t.Error("SkipDepm should be true")
	}
}

func TestParseSharedFlags_SkipDepm_Test_OK(t *testing.T) {
	env := &environment.Env{IsLocalConsole: true}
	args := []string{"--skip-depm"}

	flags, err := ParseSharedFlagsWithEnv(TestConfig(), args, env)
	if err != nil {
		t.Fatalf("ParseSharedFlags() unexpected error: %v", err)
	}
	if !flags.SkipDepm {
		t.Error("SkipDepm should be true")
	}
}

func TestParseSharedFlags_Defaults(t *testing.T) {
	env := &environment.Env{IsLocalConsole: true}
	args := []string{}

	flags, err := ParseSharedFlagsWithEnv(BuildConfig(), args, env)
	if err != nil {
		t.Fatalf("ParseSharedFlags() unexpected error: %v", err)
	}

	// Should have defaults
	if !flags.UseTUI {
		t.Error("UseTUI should default to true in local console")
	}
	if flags.Turbo {
		t.Error("Turbo should default to false")
	}
	if flags.SkipCache {
		t.Error("SkipCache should default to false")
	}
	if flags.DryRun {
		t.Error("DryRun should default to false")
	}
}
