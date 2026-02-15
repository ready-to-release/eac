// Package internal provides shared utilities for work commands
package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
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
// For example, for "clie work create feature/test --debug",
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
	ts := tool.GlobalToolSystem()
	config.GitOps = NewDefaultGitOps(config.RepoRoot, ts)

	return config, nil
}

// RunCommand provides shared scaffolding for work subcommands.
// It validates flags, parses base config, verifies we're in a git repo,
// and invokes the command-specific logic. This eliminates repeated
// boilerplate across commit, create, merge, pull, and remove subcommands.
//
// The argOffset is the number of os.Args to skip (e.g., 2 for "work commit").
// The fn receives the parsed BaseConfig and remaining args, and returns an exit code.
func RunCommand(argOffset int, fn func(base *BaseConfig, args []string) int) int {
	startTime := time.Now()

	args := os.Args[argOffset:]

	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(args); err != nil {
		logging.C().Errorf("%v", err)
		return 1
	}

	// Parse base config
	base, err := ParseBaseConfig(args)
	if err != nil {
		logging.C().Errorf("Error: %v", err)
		return 1
	}
	defer func() { _ = base.Logger.Sync() }() //nolint:errcheck // best-effort sync

	// Verify git repository
	if err := EnsureInGitRepo(); err != nil {
		base.Logger.Error(fmt.Sprintf("Not in git repository: %v", err))
		return 1
	}

	exitCode := fn(base, args)

	if base.Debug {
		base.Logger.Debug(fmt.Sprintf("Command completed in %v", time.Since(startTime)))
	}

	return exitCode
}
