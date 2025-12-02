package assessment

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// Config holds the configuration for the assessment command
type Config struct {
	Scope         string // staged, changed, all
	Destination   string // output file path
	PromptPath    string // custom prompt
	Debug         bool   // -D flag
	WorkspaceRoot string
}

// parseConfig parses command-line arguments into a Config
func parseConfig() (*Config, error) {
	config := &Config{
		Scope: "staged", // default
	}

	args := os.Args[3:] // Skip "r2r", "risks", "assessment"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-s", "--scope":
			if i+1 < len(args) {
				config.Scope = args[i+1]
				i++
			}
		case "-d", "--destination":
			if i+1 < len(args) {
				config.Destination = args[i+1]
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
		}
	}

	// Validate scope
	validScopes := map[string]bool{
		"staged":  true,
		"changed": true,
		"all":     true,
	}
	if !validScopes[config.Scope] {
		return nil, fmt.Errorf("invalid scope: %s (must be: staged, changed, or all)", config.Scope)
	}

	// Default destination with timestamp
	if config.Destination == "" {
		timestamp := time.Now().Format("2006-01-02-15-04-05")
		config.Destination = fmt.Sprintf(".docs/reference/risk-assessment-%s.md", timestamp)
	}

	// Security: prevent path traversal
	if strings.Contains(config.Destination, "..") {
		return nil, fmt.Errorf("security error: path traversal detected in destination")
	}

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}
	config.WorkspaceRoot = workspaceRoot

	return config, nil
}

// ensureOutputDir ensures the output directory exists
func ensureOutputDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	return nil
}
