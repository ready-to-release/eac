package templates

import (
	"fmt"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

// BaseConfig holds configuration common to all templates commands.
type BaseConfig struct {
	WorkspaceRoot string
	Debug         bool
	Logger        *logging.Logger
}

// parseBaseFlags parses common flags (--debug/-d) from args and returns BaseConfig
// Returns the config and remaining unparsed args.
func parseBaseFlags(args []string) (*BaseConfig, []string, error) {
	// Parse debug flag using shared package
	debug := flags.ParseDebugFlag(args)
	var remaining []string

	// Parse flags
	for _, arg := range args {
		switch arg {
		case "--debug", "-d":
			// Already handled by shared flags package
		default:
			// Keep non-flag args for further processing
			remaining = append(remaining, arg)
		}
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Initialize logger
	var logger *logging.Logger
	if debug {
		logger, err = logging.NewWithDebug("templates", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("templates", workspaceRoot)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	config := &BaseConfig{
		WorkspaceRoot: workspaceRoot,
		Debug:         debug,
		Logger:        logger,
	}

	return config, remaining, nil
}

// initLogger creates a logger for templates commands
// This is a convenience function for commands that need to initialize logger independently.
func initLogger(debug bool, workspaceRoot string) (*logging.Logger, error) {
	var logger *logging.Logger
	var err error

	if debug {
		logger, err = logging.NewWithDebug("templates", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("templates", workspaceRoot)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return logger, nil
}

// hasDebugFlag checks if --debug or -d flag is present in args.
func hasDebugFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--debug" || arg == "-d" {
			return true
		}
	}
	return false
}

// getWorkspaceRoot returns the repository root directory
// This is a convenience wrapper around repository.GetRepositoryRoot.
func getWorkspaceRoot() (string, error) {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return "", fmt.Errorf("failed to find repository root: %w", err)
	}
	return workspaceRoot, nil
}
