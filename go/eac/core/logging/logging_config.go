package logging

import (
	"os"
	"path/filepath"
	"strings"

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
	// Rolling log configuration (file only)
	MaxSizeMB   int `yaml:"max_size_mb,omitempty"`   // Max size in MB before rotation (default: 5)
	MaxBackups  int `yaml:"max_backups,omitempty"`   // Max number of old log files to keep (default: 3)
	MaxAgeDays  int `yaml:"max_age_days,omitempty"`  // Max days to retain old log files (default: 7)
	Compress    *bool `yaml:"compress,omitempty"`    // Compress rotated files (default: false)
}

// TargetConfig holds configuration for an extra log target (build/test specific logs)
type TargetConfig struct {
	// Path pattern with {unit} placeholder, e.g., "out/build/{unit}/build.log"
	Path string `yaml:"path"`
	// Levels to output: debug, info, warn, error
	Levels []string `yaml:"levels"`
	// Formatter type: raw, timestamped, json
	Formatter FormatterType `yaml:"formatter"`
}

// ResolveTargetPath replaces {unit} placeholder with actual unit name
func (t TargetConfig) ResolveTargetPath(repoRoot, unit string) string {
	path := strings.Replace(t.Path, "{unit}", unit, -1)
	return filepath.Join(repoRoot, path)
}

// LoggingConfig holds the complete logging configuration
type LoggingConfig struct {
	Console SinkConfig              `yaml:"console"`
	File    SinkConfig              `yaml:"file"`
	Targets map[string]TargetConfig `yaml:"targets"` // command -> target config
}

// GetTarget returns target config for a command, if configured
func (c LoggingConfig) GetTarget(command string) (TargetConfig, bool) {
	target, ok := c.Targets[command]
	return target, ok
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
			Levels:     []string{"debug", "info", "warn", "error"},
			Formatter:  FormatterTimestamped,
			Enabled:    &enabled,
			MaxSizeMB:  5,  // 5MB before rotation
			MaxBackups: 3,  // keep 3 old files
			MaxAgeDays: 7,  // delete files older than 7 days
		},
	}
}

// LoadLoggingConfig loads logging configuration with defaults from contracts.
// Merge order: contract defaults -> user config (.r2r/eac/logging.yml)
// Falls back to built-in defaults if contract defaults don't exist.
func LoadLoggingConfig(workspaceRoot string) LoggingConfig {
	// Load contract defaults first
	defaults := loadLoggingDefaults(workspaceRoot)

	// Load user config
	configPath := paths.EACLoggingConfigPath(workspaceRoot)
	data, err := os.ReadFile(configPath)
	if err != nil {
		// No user config, return defaults
		return defaults
	}

	var userCfg LoggingConfig
	if err := yaml.Unmarshal(data, &userCfg); err != nil {
		// Invalid user YAML, return defaults
		return defaults
	}

	// Merge user config on top of defaults
	return mergeLoggingConfig(defaults, userCfg)
}

// loadLoggingDefaults loads logging defaults from contracts/eac-core/0.1.0/defaults/logging.yml
func loadLoggingDefaults(workspaceRoot string) LoggingConfig {
	defaultsPath := paths.LoggingDefaultsPath(workspaceRoot)
	data, err := os.ReadFile(defaultsPath)
	if err != nil {
		// Contract defaults don't exist, use built-in defaults
		return DefaultLoggingConfig()
	}

	var cfg LoggingConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// Invalid defaults YAML, use built-in defaults
		return DefaultLoggingConfig()
	}

	// Apply built-in defaults for any missing fields
	return applyDefaults(cfg)
}

// mergeLoggingConfig merges user config on top of defaults
func mergeLoggingConfig(defaults, user LoggingConfig) LoggingConfig {
	result := defaults

	// Console: user overrides if specified
	if len(user.Console.Levels) > 0 {
		result.Console.Levels = user.Console.Levels
	}
	if user.Console.Formatter != "" {
		result.Console.Formatter = user.Console.Formatter
	}

	// File: user overrides if specified
	if len(user.File.Levels) > 0 {
		result.File.Levels = user.File.Levels
	}
	if user.File.Formatter != "" {
		result.File.Formatter = user.File.Formatter
	}
	if user.File.Enabled != nil {
		result.File.Enabled = user.File.Enabled
	}

	// Targets: user targets override defaults by command name
	if user.Targets != nil {
		if result.Targets == nil {
			result.Targets = make(map[string]TargetConfig)
		}
		for cmd, target := range user.Targets {
			result.Targets[cmd] = target
		}
	}

	return result
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
	// Rolling log defaults
	if cfg.File.MaxSizeMB == 0 {
		cfg.File.MaxSizeMB = defaults.File.MaxSizeMB
	}
	if cfg.File.MaxBackups == 0 {
		cfg.File.MaxBackups = defaults.File.MaxBackups
	}
	if cfg.File.MaxAgeDays == 0 {
		cfg.File.MaxAgeDays = defaults.File.MaxAgeDays
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
