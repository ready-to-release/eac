package logging

import (
	"os"

	"github.com/ready-to-release/eac/go/eac/core/paths"
	"gopkg.in/yaml.v3"
)

// FormatterType defines the output format for log messages
type FormatterType string

const (
	// FormatterRaw outputs only the message (clean CLI output)
	FormatterRaw FormatterType = "raw"
	// FormatterTimestamped outputs "HH:MM:SS.mmm  LEVEL  module:message"
	FormatterTimestamped FormatterType = "timestamped"
	// FormatterJSON outputs structured JSON
	FormatterJSON FormatterType = "json"
)

// SinkConfig holds configuration for a single logging sink (console or file)
type SinkConfig struct {
	// Levels to output: debug, info, warn, error
	Levels []string `yaml:"levels"`
	// Formatter type: raw, timestamped, json
	Formatter FormatterType `yaml:"formatter"`
	// Enabled controls whether this sink is active (file only)
	Enabled *bool `yaml:"enabled,omitempty"`
}

// LoggingConfig holds the complete logging configuration
type LoggingConfig struct {
	Console SinkConfig `yaml:"console"`
	File    SinkConfig `yaml:"file"`
}

// DefaultLoggingConfig returns the default logging configuration
func DefaultLoggingConfig() LoggingConfig {
	enabled := true
	return LoggingConfig{
		Console: SinkConfig{
			Levels:    []string{"info", "warn", "error"},
			Formatter: FormatterRaw,
		},
		File: SinkConfig{
			Levels:    []string{"debug", "info", "warn", "error"},
			Formatter: FormatterJSON,
			Enabled:   &enabled,
		},
	}
}

// LoadLoggingConfig loads logging configuration from .r2r/eac/logging.yml
// Falls back to defaults if file doesn't exist or can't be parsed
func LoadLoggingConfig(workspaceRoot string) LoggingConfig {
	configPath := paths.EACLoggingConfigPath(workspaceRoot)

	data, err := os.ReadFile(configPath)
	if err != nil {
		// File doesn't exist or can't be read, use defaults
		return DefaultLoggingConfig()
	}

	var cfg LoggingConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// Invalid YAML, use defaults
		return DefaultLoggingConfig()
	}

	// Apply defaults for missing fields
	cfg = applyDefaults(cfg)

	return cfg
}

// applyDefaults fills in missing configuration with defaults
func applyDefaults(cfg LoggingConfig) LoggingConfig {
	defaults := DefaultLoggingConfig()

	// Console defaults
	if len(cfg.Console.Levels) == 0 {
		cfg.Console.Levels = defaults.Console.Levels
	}
	if cfg.Console.Formatter == "" {
		cfg.Console.Formatter = defaults.Console.Formatter
	}

	// File defaults
	if len(cfg.File.Levels) == 0 {
		cfg.File.Levels = defaults.File.Levels
	}
	if cfg.File.Formatter == "" {
		cfg.File.Formatter = defaults.File.Formatter
	}
	if cfg.File.Enabled == nil {
		cfg.File.Enabled = defaults.File.Enabled
	}

	return cfg
}

// HasLevel checks if a sink config includes a specific level
func (s *SinkConfig) HasLevel(level string) bool {
	for _, l := range s.Levels {
		if l == level {
			return true
		}
	}
	return false
}

// IsEnabled returns whether the sink is enabled (defaults to true)
func (s *SinkConfig) IsEnabled() bool {
	if s.Enabled == nil {
		return true
	}
	return *s.Enabled
}
