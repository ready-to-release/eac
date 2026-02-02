// Package internal provides shared utilities for work commands
package internal

import (
	"fmt"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
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
		Debug: flags.ParseDebugFlag(args),
	}

	// Get repository root
	var err error
	config.RepoRoot, err = repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Initialize logger using ConfigureLogging
	// Logs to unified out/commands.log
	if err := logging.ConfigureLoggingSimple(config.RepoRoot, "work", nil, config.Debug); err != nil {
		return nil, fmt.Errorf("failed to configure logging: %w", err)
	}
	config.Logger, err = logging.NewDefault("work", config.RepoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Initialize git operations
	config.GitOps = GetGitOps(config.RepoRoot)

	return config, nil
}
