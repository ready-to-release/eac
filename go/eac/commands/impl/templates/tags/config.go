package tags

import (
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// Config holds configuration for the templates tags command
type Config struct {
	WorkspaceRoot  string
	TemplateSource string
	Validate       bool   // Whether to validate against template-tags.yml
	OutputFormat   string // "table", "json", "summary"
	Depth          string // "summary", "file", "category", "coverage", "consistency"
	Debug          bool
	Logger         *logging.Logger
}

// NewConfig creates a new config with defaults
func NewConfig(workspaceRoot string, debug bool, logger *logging.Logger) *Config {
	return &Config{
		WorkspaceRoot:  workspaceRoot,
		TemplateSource: "templates", // Default to local templates directory
		Validate:       true,        // Validate by default
		OutputFormat:   "table",
		Depth:          "summary",   // Default depth level
		Debug:          debug,
		Logger:         logger,
	}
}
