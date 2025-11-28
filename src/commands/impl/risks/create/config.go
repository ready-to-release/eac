package create

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/core/repository"
)

// Config holds the configuration for the create command
type Config struct {
	AssessmentPath string // file or folder
	Force          bool   // -f flag
	AllowOrphans   bool   // --allow-orphans flag (requires --force)
	OutputDir      string // -o flag
	PromptPath     string // -p flag
	Debug          bool   // -D flag
	WorkspaceRoot  string
}

// parseConfig parses command-line arguments into a Config
func parseConfig() (*Config, error) {
	config := &Config{
		OutputDir: "specs/risk-controls/", // default
	}

	args := os.Args[3:] // Skip "r2r", "risks", "create"

	positional := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-f", "--force":
			config.Force = true
		case "--allow-orphans":
			config.AllowOrphans = true
		case "-o", "--output":
			if i+1 < len(args) {
				config.OutputDir = args[i+1]
				i++
			}
		case "-p", "--prompt":
			if i+1 < len(args) {
				config.PromptPath = args[i+1]
				i++
			}
		case "-D", "--debug":
			config.Debug = true
		case "--help", "-h":
			return nil, fmt.Errorf("help requested")
		default:
			if !strings.HasPrefix(arg, "-") {
				positional = append(positional, arg)
			}
		}
	}

	// Validate flag combinations
	if config.AllowOrphans && !config.Force {
		return nil, fmt.Errorf("--allow-orphans requires --force flag")
	}

	if len(positional) == 0 {
		return nil, fmt.Errorf("usage: risks create <assessment-file-or-folder> [flags]")
	}

	config.AssessmentPath = positional[0]

	// Security: prevent path traversal
	if strings.Contains(config.AssessmentPath, "..") || strings.Contains(config.OutputDir, "..") {
		return nil, fmt.Errorf("security error: path traversal detected")
	}

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}
	config.WorkspaceRoot = workspaceRoot

	return config, nil
}
