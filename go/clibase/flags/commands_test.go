package flags

import (
	"testing"

	"github.com/ready-to-release/eac/go/clibase/environment"
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
		"--with-tui",
		"--debug",
		"--timings",
		"--skip-cache",
		"--no-deps",
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
	if !flags.CacheConfig.ShouldSkipState() {
		t.Error("ShouldSkipState() should be true")
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
		"--with-tidy",   // Build-specific (unknown flag)
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
		t.Errorf("Remaining = %v, want 2 items (--with-tidy, --version=1.0)", flags.Remaining)
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
	if flags.CacheConfig.ShouldSkipState() {
		t.Error("ShouldSkipState() should default to false")
	}
	if flags.DryRun {
		t.Error("DryRun should default to false")
	}
}

func TestParseSharedFlags_DeclarativeState(t *testing.T) {
	tests := []struct {
		name                 string
		args                 []string
		wantCacheExplicit    bool
		wantDepsExplicit     bool
		wantParallelExplicit bool
		wantTUIExplicit      bool
	}{
		{
			name:              "no explicit flags",
			args:              []string{"module1"},
			wantCacheExplicit: false,
			wantDepsExplicit:  false,
			wantParallelExplicit: false,
			wantTUIExplicit:      false,
		},
		{
			name:              "with-cache sets CacheExplicit",
			args:              []string{"--with-cache", "module1"},
			wantCacheExplicit: true,
			wantDepsExplicit:  false,
			wantParallelExplicit: false,
			wantTUIExplicit:      false,
		},
		{
			name:              "no-cache sets CacheExplicit",
			args:              []string{"--no-cache", "module1"},
			wantCacheExplicit: true,
			wantDepsExplicit:  false,
			wantParallelExplicit: false,
			wantTUIExplicit:      false,
		},
		{
			name:              "skip-cache sets CacheExplicit (backward compat)",
			args:              []string{"--skip-cache", "module1"},
			wantCacheExplicit: true,
			wantDepsExplicit:  false,
			wantParallelExplicit: false,
			wantTUIExplicit:      false,
		},
		{
			name:              "with-deps sets DepsExplicit",
			args:              []string{"--with-deps", "module1"},
			wantCacheExplicit: false,
			wantDepsExplicit:  true,
			wantParallelExplicit: false,
			wantTUIExplicit:      false,
		},
		{
			name:              "parallel sets ParallelExplicit",
			args:              []string{"--parallel", "module1"},
			wantCacheExplicit: false,
			wantDepsExplicit:  false,
			wantParallelExplicit: true,
			wantTUIExplicit:      false,
		},
		{
			name:              "sequential sets ParallelExplicit",
			args:              []string{"--sequential", "module1"},
			wantCacheExplicit: false,
			wantDepsExplicit:  false,
			wantParallelExplicit: true,
			wantTUIExplicit:      false,
		},
		{
			name:              "with-tui sets TUIExplicit",
			args:              []string{"--with-tui", "module1"},
			wantCacheExplicit: false,
			wantDepsExplicit:  false,
			wantParallelExplicit: false,
			wantTUIExplicit:      true,
		},
		{
			name:              "no-tui sets TUIExplicit",
			args:              []string{"--no-tui", "module1"},
			wantCacheExplicit: false,
			wantDepsExplicit:  false,
			wantParallelExplicit: false,
			wantTUIExplicit:      true,
		},
		{
			name:              "all declarative flags",
			args:              []string{"--with-cache", "--with-deps", "--parallel", "--no-tui", "module1"},
			wantCacheExplicit: true,
			wantDepsExplicit:  true,
			wantParallelExplicit: true,
			wantTUIExplicit:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &environment.Env{IsLocalConsole: true}
			flags, err := ParseSharedFlagsWithEnv(BuildConfig(), tt.args, env)
			if err != nil {
				t.Fatalf("ParseSharedFlagsWithEnv() error: %v", err)
			}

			if flags.CacheExplicit != tt.wantCacheExplicit {
				t.Errorf("CacheExplicit = %v, want %v", flags.CacheExplicit, tt.wantCacheExplicit)
			}
			if flags.DepsExplicit != tt.wantDepsExplicit {
				t.Errorf("DepsExplicit = %v, want %v", flags.DepsExplicit, tt.wantDepsExplicit)
			}
			if flags.ParallelExplicit != tt.wantParallelExplicit {
				t.Errorf("ParallelExplicit = %v, want %v", flags.ParallelExplicit, tt.wantParallelExplicit)
			}
			if flags.TUIExplicit != tt.wantTUIExplicit {
				t.Errorf("TUIExplicit = %v, want %v", flags.TUIExplicit, tt.wantTUIExplicit)
			}
		})
	}
}
