// File: go/cli/eac/impl/create/risk-assess/assess_config.go
// Purpose: Configuration parsing and validation for risk assessment command
//
// This file contains parseAssessConfig which handles CLI flag parsing, module
// discovery/validation, and workspace root resolution.

package riskassess

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/repository"
)

// loadModuleRegistry loads the module registry from the workspace.
func loadModuleRegistry(workspaceRoot string) (*modules.Registry, error) {
	return modules.LoadFromWorkspace(workspaceRoot)
}

// parseAssessConfig parses command line configuration.
func parseAssessConfig() (*AssessConfig, error) {
	args := os.Args[3:] // Skip program name, "create", and "risk-assess"

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		return nil, err
	}

	config := &AssessConfig{
		MaxEvidenceAge: 24 * time.Hour,
	}

	// Parse debug flag using shared package
	config.Debug = flags.ParseDebugFlag(args)
	config.Sequential = flags.HasFlag(args, "--sequential", "")

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find workspace root: %w", err)
	}

	config.WorkspaceRoot = workspaceRoot

	// Collect positional arguments (module names) before flags
	seen := make(map[string]bool)
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Stop at first flag
		if strings.HasPrefix(arg, "-") {
			break
		}

		// Check for duplicates
		if seen[arg] {
			return nil, fmt.Errorf("duplicate module specified: %s", arg)
		}
		seen[arg] = true

		// Collect module name
		config.Modules = append(config.Modules, arg)
	}

	// Parse flags starting from where modules ended
	if err := parseAssessFlags(config, args[len(config.Modules):]); err != nil {
		return nil, err
	}

	// Resolve and validate modules
	if err := resolveModules(config); err != nil {
		return nil, err
	}

	// Validate required flags
	if config.ProfilePath == "" {
		return nil, fmt.Errorf("--profile flag is required")
	}

	return config, nil
}

// parseAssessFlags parses flag arguments (everything after positional module names).
func parseAssessFlags(config *AssessConfig, args []string) error {
	i := 0
	for i < len(args) {
		arg := args[i]

		switch {
		case arg == "--profile" || arg == "-p":
			if i+1 >= len(args) {
				return fmt.Errorf("--profile requires a value")
			}
			config.ProfilePath = args[i+1]
			// Make path absolute if relative
			if !filepath.IsAbs(config.ProfilePath) {
				config.ProfilePath = filepath.Join(config.WorkspaceRoot, config.ProfilePath)
			}
			i += 2

		case arg == "--max-evidence-age":
			if i+1 >= len(args) {
				return fmt.Errorf("--max-evidence-age requires a value")
			}
			duration, err := time.ParseDuration(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid duration: %s", args[i+1])
			}

			// Validate duration is positive and reasonable
			if duration <= 0 {
				return fmt.Errorf("--max-evidence-age must be positive")
			}

			const maxDuration = 30 * 24 * time.Hour // 30 days
			if duration > maxDuration {
				return fmt.Errorf("--max-evidence-age too large (max: 30d)")
			}

			config.MaxEvidenceAge = duration
			i += 2

		case arg == "--debug" || arg == "-d" || arg == "--sequential":
			// Already handled by shared flags package
			i++

		case strings.HasPrefix(arg, "-"):
			return fmt.Errorf("unknown flag: %s", arg)

		default:
			return fmt.Errorf("unexpected positional argument after flags: %s", arg)
		}
	}

	return nil
}

// resolveModules discovers or validates module names against the workspace registry.
func resolveModules(config *AssessConfig) error {
	// Load registry to validate modules or discover all
	registry, err := loadModuleRegistry(config.WorkspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to load modules: %w", err)
	}

	// If no modules specified, discover all modules
	if len(config.Modules) == 0 {
		allModules := registry.All()
		for _, mod := range allModules {
			config.Modules = append(config.Modules, mod.Moniker)
		}

		if len(config.Modules) == 0 {
			return fmt.Errorf("no modules found in workspace")
		}
	} else {
		// Validate that all specified modules exist
		allModules := make(map[string]bool)
		for _, mod := range registry.All() {
			allModules[mod.Moniker] = true
		}

		var invalidModules []string
		for _, moduleName := range config.Modules {
			if !allModules[moduleName] {
				invalidModules = append(invalidModules, moduleName)
			}
		}

		if len(invalidModules) > 0 {
			availableModules := make([]string, 0, len(allModules))
			for mod := range allModules {
				availableModules = append(availableModules, mod)
			}
			return fmt.Errorf(`unknown module(s): %s

Available modules:
  %s

Try:
  - Check module name spelling
  - List all modules: show modules
  - View module contracts: cat .eac/repository.yml`,
				strings.Join(invalidModules, ", "),
				strings.Join(availableModules, ", "))
		}
	}

	return nil
}
