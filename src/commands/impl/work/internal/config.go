// Package internal provides shared utilities for work commands
package internal

import (
	"fmt"

	"github.com/ready-to-release/eac/src/core/logging"
	"github.com/ready-to-release/eac/src/core/repository"
)

// BaseConfig holds common configuration for all work commands.
// This includes debug mode, repository root, logger, and git operations.
type BaseConfig struct {
	Debug    bool
	RepoRoot string
	Logger   *logging.Logger
	GitOps   WorkGitOperations
}

// ParseBaseConfig parses common configuration from command-line arguments.
// It extracts the --debug/-d flag, gets the repository root,
// initializes the logger, and sets up git operations.
//
// The args parameter should be the command arguments AFTER the subcommand.
// For example, for "r2r work create feature/test --debug",
// args should be ["feature/test", "--debug"].
func ParseBaseConfig(args []string) (*BaseConfig, error) {
	config := &BaseConfig{
		Debug: false,
	}

	// Parse debug flag
	for _, arg := range args {
		if arg == "--debug" || arg == "-d" {
			config.Debug = true
			break
		}
	}

	// Get repository root
	var err error
	config.RepoRoot, err = repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Initialize logger
	config.Logger, err = InitLogger(config.Debug, config.RepoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize git operations
	config.GitOps = GetGitOps(config.RepoRoot)

	return config, nil
}

// HasFlag checks if a specific flag is present in the arguments.
func HasFlag(args []string, flag string, shorthand string) bool {
	for _, arg := range args {
		if arg == flag || arg == shorthand {
			return true
		}
	}
	return false
}

// GetFlagValue retrieves the value for a flag in --flag=value format.
// Returns empty string if not found.
func GetFlagValue(args []string, prefix string) string {
	for _, arg := range args {
		if len(arg) > len(prefix) && arg[:len(prefix)] == prefix && arg[len(prefix)] == '=' {
			return arg[len(prefix)+1:]
		}
	}
	return ""
}

// GetPositionalArgs returns non-flag arguments.
// Arguments starting with -- or - are considered flags.
func GetPositionalArgs(args []string) []string {
	var positional []string
	for _, arg := range args {
		if len(arg) == 0 {
			continue
		}
		// Skip flags
		if arg[0] == '-' {
			continue
		}
		positional = append(positional, arg)
	}
	return positional
}
