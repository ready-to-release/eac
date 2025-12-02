package list

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// Config holds the configuration for the list command
type Config struct {
	Filter        string // all, orphaned, linked, missing-links
	JSON          bool   // -j flag
	Debug         bool   // -D flag
	WorkspaceRoot string
}

// parseConfig parses command-line arguments into a Config
func parseConfig() (*Config, error) {
	config := &Config{
		Filter: "all", // default
	}

	args := os.Args[3:] // Skip "r2r", "risks", "list"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-f", "--filter":
			if i+1 < len(args) {
				config.Filter = args[i+1]
				i++
			}
		case "-j", "--json":
			config.JSON = true
		case "-D", "--debug":
			config.Debug = true
		case "--help", "-h":
			return nil, fmt.Errorf("help requested")
		}
	}

	// Validate filter
	validFilters := map[string]bool{
		"all":           true,
		"orphaned":      true,
		"linked":        true,
		"missing-links": true,
	}
	if !validFilters[config.Filter] {
		return nil, fmt.Errorf("invalid filter: %s (must be: all, orphaned, linked, or missing-links)", config.Filter)
	}

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}
	config.WorkspaceRoot = workspaceRoot

	return config, nil
}
