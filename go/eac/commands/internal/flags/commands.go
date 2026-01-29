package flags

import (
	"fmt"

	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
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
	SkipTUIDelay     bool
	Debug        bool
	ShowTimings  bool

	// Cache control
	SkipCache bool // Force full execution (skip incremental cache)
	SkipDeps  bool // Skip system dependency verification

	// Module control
	Exclude  string // Module exclusion pattern
	SkipDepm bool   // Skip module dependency processing

	// Dry run
	DryRun bool

	// Positional arguments (module monikers)
	Monikers []string

	// Remaining args (command-specific flags)
	Remaining []string
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
	}

	// Extract output flags
	if parsed.Output != nil {
		result.UseTUI = parsed.Output.UseTUI
		result.TUIHeight = parsed.Output.TUIHeight
		result.TUIASCIIMode = parsed.Output.TUIASCIIMode
		result.SkipTUIDelay = parsed.Output.SkipTUIDelay
		result.Debug = parsed.Output.Debug
		result.ShowTimings = parsed.Output.ShowTimings
	}

	// Extract cache flags
	if parsed.Cache != nil {
		result.SkipCache = parsed.Cache.SkipCache
		result.SkipDeps = parsed.Cache.SkipDeps
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
