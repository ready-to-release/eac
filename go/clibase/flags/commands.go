package flags

import (
	"fmt"

	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/core/cache"
)

// Predefined flag configurations for each command.
// These define which flag sets each command subscribes to.

// BuildConfig returns the flag configuration for the build command.
func BuildConfig() CommandFlagConfig {
	return CommandFlagConfig{
		Command:   "build",
		Execution: true,
		Output:    true,
		Cache:     true,
		Module:    true,
		DryRun:    true,
	}
}

// TestConfig returns the flag configuration for the test command.
func TestConfig() CommandFlagConfig {
	return CommandFlagConfig{
		Command:   "test",
		Execution: true,
		Output:    true,
		Cache:     true,
		Module:    true,
		DryRun:    true,
	}
}

// LintConfig returns the flag configuration for the lint command.
func LintConfig() CommandFlagConfig {
	return CommandFlagConfig{
		Command:   "lint",
		Execution: true,
		Output:    true,
		Cache:     true,
		Module:    true,
		DryRun:    true,
	}
}

// ScanConfig returns the flag configuration for the scan command.
func ScanConfig() CommandFlagConfig {
	return CommandFlagConfig{
		Command:   "scan",
		Execution: true,
		Output:    true,
		Cache:     true,
		Module:    true,
		DryRun:    true,
	}
}

// SharedFlags holds all parsed shared flags from flag sets.
// This is the output that commands consume after parsing.
type SharedFlags struct {
	// Execution control
	Turbo          bool
	MaxConcurrency int // 0 = auto-detect

	// Output control
	UseTUI       bool
	TUIHeight    int
	TUIASCIIMode bool
	TUI3Demo     bool
	SkipTUIDelay bool
	Debug        bool
	ShowTimings  bool

	// Cache control
	CacheConfig *cache.Config // Fine-grained cache control via --skip-cache=<spec>
	SkipDeps    bool          // Skip system dependency verification

	// Module control
	Exclude  string // Module exclusion pattern
	SkipDepm bool   // Skip module dependency processing

	// Dry run
	DryRun bool

	// Positional arguments (module monikers)
	Monikers []string

	// Remaining args (command-specific flags)
	Remaining []string

	// Declarative state tracking
	// These fields indicate whether the user explicitly set the flags
	CacheExplicit    bool // True if --with-cache or --no-cache was used
	DepsExplicit     bool // True if --with-deps or --no-deps was used
	ParallelExplicit bool // True if --parallel or --sequential was used
	TUIExplicit      bool // True if --with-tui or --no-tui was used
}

// ParseSharedFlags parses command-line arguments using the specified command configuration.
// Returns SharedFlags populated from the parsed flag sets, plus any remaining arguments.
func ParseSharedFlags(config CommandFlagConfig, args []string) (*SharedFlags, error) {
	return ParseSharedFlagsWithEnv(config, args, environment.Detect())
}

// ParseSharedFlagsWithEnv parses with a specific environment (for testing).
func ParseSharedFlagsWithEnv(config CommandFlagConfig, args []string, env *environment.Env) (*SharedFlags, error) {
	parser := NewParserWithEnv(config, env)

	parsed, err := parser.Parse(args)
	if err != nil {
		return nil, err
	}

	result := &SharedFlags{
		Monikers:  parsed.Positional,
		Remaining: parsed.Remaining,
	}

	// Extract execution flags
	if parsed.Execution != nil {
		result.Turbo = parsed.Execution.Turbo
		result.MaxConcurrency = parsed.Execution.Roof
		result.ParallelExplicit = parsed.Execution.ParallelExplicit
	}

	// Extract output flags
	if parsed.Output != nil {
		result.UseTUI = parsed.Output.UseTUI
		result.TUIHeight = parsed.Output.TUIHeight
		result.TUIASCIIMode = parsed.Output.TUIASCIIMode
		result.TUI3Demo = parsed.Output.TUI3Demo
		result.SkipTUIDelay = parsed.Output.SkipTUIDelay
		result.Debug = parsed.Output.Debug
		result.ShowTimings = parsed.Output.ShowTimings
		result.TUIExplicit = parsed.Output.TUIExplicitlySet
	}

	// Extract cache flags
	if parsed.Cache != nil {
		result.CacheConfig = parsed.Cache.CacheConfig
		result.SkipDeps = parsed.Cache.SkipDeps
		result.CacheExplicit = parsed.Cache.CacheExplicit
		result.DepsExplicit = parsed.Cache.DepsExplicit
	}

	// Extract module flags
	if parsed.Module != nil {
		result.Exclude = parsed.Module.Exclude
		result.SkipDepm = parsed.Module.SkipDepm

		// Validate --skip-depm for lint/scan
		if result.SkipDepm && (config.Command == "lint" || config.Command == "scan") {
			return nil, fmt.Errorf("--skip-depm is not implemented for %s command", config.Command)
		}
	}

	// Extract dry-run flag
	if parsed.DryRun != nil {
		result.DryRun = parsed.DryRun.DryRun
	}

	return result, nil
}
