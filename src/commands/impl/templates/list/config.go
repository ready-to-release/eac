package list

import (
	"fmt"

	"github.com/ready-to-release/eac/src/core/logging"
)

// Config holds configuration for the templates list command
type Config struct {
	TemplateSource string          // Git URL or local path
	WorkspaceRoot  string          // Repository root directory
	Debug          bool            // Debug mode flag
	Logger         *logging.Logger // Logger instance
}

// NewConfig creates a new list configuration with defaults
func NewConfig(workspaceRoot string, debug bool, logger *logging.Logger) *Config {
	return &Config{
		TemplateSource: defaultTemplateListRepo,
		WorkspaceRoot:  workspaceRoot,
		Debug:          debug,
		Logger:         logger,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.WorkspaceRoot == "" {
		return fmt.Errorf("workspace root cannot be empty")
	}
	if c.Logger == nil {
		return fmt.Errorf("logger cannot be nil")
	}
	if c.TemplateSource == "" {
		return fmt.Errorf("template source cannot be empty")
	}
	return nil
}

// Default template repository URL
const defaultTemplateListRepo = "https://github.com/ready-to-release/eac"
