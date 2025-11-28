package apply

import (
	"fmt"

	"github.com/ready-to-release/eac/src/core/logging"
)

// Config holds configuration for the templates apply command
type Config struct {
	Source        string          // Local path or Git URL
	Destination   string          // Output directory
	ValuesFile    string          // JSON file with template values
	WorkspaceRoot string          // Repository root directory
	Debug         bool            // Debug mode flag
	Logger        *logging.Logger // Logger instance
	IsLocalSource bool            // True if source is local path
}

// NewConfig creates a new apply configuration
func NewConfig(workspaceRoot string, debug bool, logger *logging.Logger) *Config {
	return &Config{
		WorkspaceRoot: workspaceRoot,
		Debug:         debug,
		Logger:        logger,
		IsLocalSource: false,
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
	if c.Destination == "" {
		return fmt.Errorf("destination cannot be empty")
	}
	return nil
}
