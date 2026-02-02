package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultLoggingConfig(t *testing.T) {
	cfg := DefaultLoggingConfig()

	// Console defaults
	if cfg.Console.Formatter != FormatterRaw {
		t.Errorf("expected console formatter %q, got %q", FormatterRaw, cfg.Console.Formatter)
	}
	if !cfg.Console.HasLevel("info") {
		t.Error("expected console to have info level")
	}
	if !cfg.Console.HasLevel("warn") {
		t.Error("expected console to have warn level")
	}
	if !cfg.Console.HasLevel("error") {
		t.Error("expected console to have error level")
	}
	if cfg.Console.HasLevel("debug") {
		t.Error("expected console NOT to have debug level by default")
	}

	// File defaults
	if cfg.File.Formatter != FormatterTimestamped {
		t.Errorf("expected file formatter %q, got %q", FormatterTimestamped, cfg.File.Formatter)
	}
	if !cfg.File.HasLevel("debug") {
		t.Error("expected file to have debug level")
	}
	if !cfg.File.HasLevel("info") {
		t.Error("expected file to have info level")
	}
	if !cfg.File.IsEnabled() {
		t.Error("expected file to be enabled by default")
	}
}

func TestLoadLoggingConfig_FileNotExists(t *testing.T) {
	// Load from non-existent directory should return defaults
	cfg := LoadLoggingConfig("/nonexistent/path")

	if cfg.Console.Formatter != FormatterRaw {
		t.Errorf("expected default console formatter, got %q", cfg.Console.Formatter)
	}
	if cfg.File.Formatter != FormatterTimestamped {
		t.Errorf("expected default file formatter, got %q", cfg.File.Formatter)
	}
}

func TestLoadLoggingConfig_ValidYAML(t *testing.T) {
	// Create temp directory with config file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".eac")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	configContent := `
console:
  levels:
    - debug
    - info
  formatter: timestamped

file:
  levels:
    - error
  formatter: json
  enabled: false
`
	configPath := filepath.Join(configDir, "logging.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := LoadLoggingConfig(tmpDir)

	// Console config
	if cfg.Console.Formatter != FormatterTimestamped {
		t.Errorf("expected console formatter %q, got %q", FormatterTimestamped, cfg.Console.Formatter)
	}
	if !cfg.Console.HasLevel("debug") {
		t.Error("expected console to have debug level")
	}
	if !cfg.Console.HasLevel("info") {
		t.Error("expected console to have info level")
	}
	if cfg.Console.HasLevel("error") {
		t.Error("expected console NOT to have error level")
	}

	// File config
	if cfg.File.Formatter != FormatterJSON {
		t.Errorf("expected file formatter %q, got %q", FormatterJSON, cfg.File.Formatter)
	}
	if !cfg.File.HasLevel("error") {
		t.Error("expected file to have error level")
	}
	if cfg.File.HasLevel("debug") {
		t.Error("expected file NOT to have debug level")
	}
	if cfg.File.IsEnabled() {
		t.Error("expected file to be disabled")
	}
}

func TestLoadLoggingConfig_InvalidYAML(t *testing.T) {
	// Create temp directory with invalid config file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".eac")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	configContent := `not valid yaml: {{{`
	configPath := filepath.Join(configDir, "logging.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should return defaults on invalid YAML
	cfg := LoadLoggingConfig(tmpDir)

	if cfg.Console.Formatter != FormatterRaw {
		t.Errorf("expected default console formatter on invalid YAML, got %q", cfg.Console.Formatter)
	}
}

func TestLoadLoggingConfig_PartialConfig(t *testing.T) {
	// Create temp directory with partial config file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".eac")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Only console formatter specified, rest should use defaults
	configContent := `
console:
  formatter: json
`
	configPath := filepath.Join(configDir, "logging.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := LoadLoggingConfig(tmpDir)

	// Console formatter should be from config
	if cfg.Console.Formatter != FormatterJSON {
		t.Errorf("expected console formatter %q, got %q", FormatterJSON, cfg.Console.Formatter)
	}
	// Console levels should use defaults
	if !cfg.Console.HasLevel("info") {
		t.Error("expected console to have default info level")
	}
	// File should use all defaults
	if cfg.File.Formatter != FormatterTimestamped {
		t.Errorf("expected file formatter %q, got %q", FormatterTimestamped, cfg.File.Formatter)
	}
	if !cfg.File.IsEnabled() {
		t.Error("expected file to be enabled by default")
	}
}

func TestSinkConfig_HasLevel(t *testing.T) {
	cfg := SinkConfig{
		Levels: []string{"info", "warn"},
	}

	if !cfg.HasLevel("info") {
		t.Error("expected HasLevel(info) to return true")
	}
	if !cfg.HasLevel("warn") {
		t.Error("expected HasLevel(warn) to return true")
	}
	if cfg.HasLevel("debug") {
		t.Error("expected HasLevel(debug) to return false")
	}
	if cfg.HasLevel("error") {
		t.Error("expected HasLevel(error) to return false")
	}
}

func TestSinkConfig_IsEnabled(t *testing.T) {
	// Nil enabled should default to true
	cfg := SinkConfig{}
	if !cfg.IsEnabled() {
		t.Error("expected IsEnabled to return true when Enabled is nil")
	}

	// Explicit true
	enabled := true
	cfg = SinkConfig{Enabled: &enabled}
	if !cfg.IsEnabled() {
		t.Error("expected IsEnabled to return true")
	}

	// Explicit false
	disabled := false
	cfg = SinkConfig{Enabled: &disabled}
	if cfg.IsEnabled() {
		t.Error("expected IsEnabled to return false")
	}
}
