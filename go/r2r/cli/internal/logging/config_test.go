package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Console defaults
	if cfg.Console.Formatter != FormatterRaw {
		t.Errorf("Default console formatter should be 'raw', got %q", cfg.Console.Formatter)
	}
	if !cfg.Console.HasLevel("info") {
		t.Error("Default console should have 'info' level")
	}
	if !cfg.Console.HasLevel("warn") {
		t.Error("Default console should have 'warn' level")
	}
	if !cfg.Console.HasLevel("error") {
		t.Error("Default console should have 'error' level")
	}
	if cfg.Console.HasLevel("debug") {
		t.Error("Default console should not have 'debug' level")
	}

	// File defaults
	if cfg.File.Formatter != FormatterJSON {
		t.Errorf("Default file formatter should be 'json', got %q", cfg.File.Formatter)
	}
	if cfg.File.IsEnabled() {
		t.Error("Default file logging should be disabled")
	}
	if !cfg.File.HasLevel("debug") {
		t.Error("Default file should have 'debug' level")
	}
}

func TestSinkConfigHasLevel(t *testing.T) {
	sink := SinkConfig{
		Levels: []string{"info", "warn", "error"},
	}

	tests := []struct {
		level    string
		expected bool
	}{
		{"info", true},
		{"warn", true},
		{"error", true},
		{"debug", false},
		{"trace", false},
		{"", false},
	}

	for _, tt := range tests {
		got := sink.HasLevel(tt.level)
		if got != tt.expected {
			t.Errorf("HasLevel(%q) = %v, want %v", tt.level, got, tt.expected)
		}
	}
}

func TestSinkConfigIsEnabled(t *testing.T) {
	// nil Enabled pointer should default to true
	sink1 := SinkConfig{}
	if !sink1.IsEnabled() {
		t.Error("Sink with nil Enabled should default to true")
	}

	// Explicit true
	enabled := true
	sink2 := SinkConfig{Enabled: &enabled}
	if !sink2.IsEnabled() {
		t.Error("Sink with Enabled=true should be enabled")
	}

	// Explicit false
	disabled := false
	sink3 := SinkConfig{Enabled: &disabled}
	if sink3.IsEnabled() {
		t.Error("Sink with Enabled=false should be disabled")
	}
}

func TestApplyDefaults(t *testing.T) {
	// Empty config should get all defaults
	empty := LoggingConfig{}
	result := applyDefaults(empty)

	if len(result.Console.Levels) == 0 {
		t.Error("applyDefaults should fill in console levels")
	}
	if result.Console.Formatter == "" {
		t.Error("applyDefaults should fill in console formatter")
	}
	if len(result.File.Levels) == 0 {
		t.Error("applyDefaults should fill in file levels")
	}
	if result.File.Formatter == "" {
		t.Error("applyDefaults should fill in file formatter")
	}
	if result.File.Enabled == nil {
		t.Error("applyDefaults should fill in file enabled")
	}

	// Partial config should keep existing values
	partial := LoggingConfig{
		Console: SinkConfig{
			Levels:    []string{"debug", "error"},
			Formatter: FormatterJSON,
		},
	}
	result2 := applyDefaults(partial)

	if len(result2.Console.Levels) != 2 {
		t.Errorf("applyDefaults should preserve console levels, got %v", result2.Console.Levels)
	}
	if result2.Console.Formatter != FormatterJSON {
		t.Errorf("applyDefaults should preserve console formatter, got %q", result2.Console.Formatter)
	}
}

func TestLoadConfigWithNoFile(t *testing.T) {
	// Create a temp directory with no config files
	tmpDir, err := os.MkdirTemp("", "logging-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := LoadConfig(tmpDir)

	// Should return defaults
	defaults := DefaultConfig()
	if cfg.Console.Formatter != defaults.Console.Formatter {
		t.Errorf("LoadConfig with no file should return default formatter, got %q", cfg.Console.Formatter)
	}
}

func TestLoadConfigWithValidFile(t *testing.T) {
	// Create a temp directory with a valid config file
	tmpDir, err := os.MkdirTemp("", "logging-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .r2r directory
	r2rDir := filepath.Join(tmpDir, ".r2r")
	if err := os.MkdirAll(r2rDir, 0o755); err != nil {
		t.Fatalf("Failed to create .r2r dir: %v", err)
	}

	// Write config file
	configContent := `
console:
  levels:
    - debug
    - info
    - warn
    - error
  formatter: timestamped
file:
  enabled: true
  levels:
    - error
  formatter: json
`
	configPath := filepath.Join(r2rDir, "r2r-cli-logging.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg := LoadConfig(tmpDir)

	if cfg.Console.Formatter != FormatterTimestamped {
		t.Errorf("Expected console formatter 'timestamped', got %q", cfg.Console.Formatter)
	}
	if !cfg.Console.HasLevel("debug") {
		t.Error("Expected console to have 'debug' level from config")
	}
	if !cfg.File.IsEnabled() {
		t.Error("Expected file logging to be enabled from config")
	}
	if len(cfg.File.Levels) != 1 || cfg.File.Levels[0] != "error" {
		t.Errorf("Expected file levels to be [error], got %v", cfg.File.Levels)
	}
}

func TestLoadConfigWithInvalidYAML(t *testing.T) {
	// Create a temp directory with an invalid config file
	tmpDir, err := os.MkdirTemp("", "logging-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .r2r directory
	r2rDir := filepath.Join(tmpDir, ".r2r")
	if err := os.MkdirAll(r2rDir, 0o755); err != nil {
		t.Fatalf("Failed to create .r2r dir: %v", err)
	}

	// Write invalid config file
	configPath := filepath.Join(r2rDir, "r2r-cli-logging.yml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content: [[["), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg := LoadConfig(tmpDir)

	// Should fall back to defaults
	defaults := DefaultConfig()
	if cfg.Console.Formatter != defaults.Console.Formatter {
		t.Errorf("LoadConfig with invalid YAML should return defaults, got formatter %q", cfg.Console.Formatter)
	}
}

func TestFormatterTypeConstants(t *testing.T) {
	// Verify formatter constants have expected values
	if FormatterRaw != "raw" {
		t.Errorf("FormatterRaw should be 'raw', got %q", FormatterRaw)
	}
	if FormatterTimestamped != "timestamped" {
		t.Errorf("FormatterTimestamped should be 'timestamped', got %q", FormatterTimestamped)
	}
	if FormatterJSON != "json" {
		t.Errorf("FormatterJSON should be 'json', got %q", FormatterJSON)
	}
}
