package registry

import (
	"testing"
)

func TestBuildDeclarativeMetadata(t *testing.T) {
	t.Run("creates enable and disable flags", func(t *testing.T) {
		def := DeclarativeFlagDefLike{
			Behavior:    "cache",
			EnableFlag:  "--with-cache",
			DisableFlag: "--no-cache",
			DefaultOn:   true,
			Description: "Incremental caching",
		}

		flags := BuildDeclarativeMetadata(def)

		if len(flags) != 2 {
			t.Fatalf("Expected 2 flags, got %d", len(flags))
		}

		// Check enable flag
		enableFlag := flags[0]
		if enableFlag.Name != "with-cache" {
			t.Errorf("Expected enable flag name 'with-cache', got %q", enableFlag.Name)
		}
		if enableFlag.Behavior != "cache" {
			t.Errorf("Expected Behavior 'cache', got %q", enableFlag.Behavior)
		}
		if !enableFlag.IsEnableFlag {
			t.Error("Expected IsEnableFlag to be true")
		}
		if enableFlag.PairFlagName != "no-cache" {
			t.Errorf("Expected PairFlagName 'no-cache', got %q", enableFlag.PairFlagName)
		}
		if enableFlag.DefaultValue != "true" {
			t.Errorf("Expected DefaultValue 'true', got %q", enableFlag.DefaultValue)
		}
		if enableFlag.Usage != "Incremental caching" {
			t.Errorf("Expected Usage 'Incremental caching', got %q", enableFlag.Usage)
		}
		if enableFlag.Type != "bool" {
			t.Errorf("Expected Type 'bool', got %q", enableFlag.Type)
		}

		// Check disable flag
		disableFlag := flags[1]
		if disableFlag.Name != "no-cache" {
			t.Errorf("Expected disable flag name 'no-cache', got %q", disableFlag.Name)
		}
		if disableFlag.Behavior != "cache" {
			t.Errorf("Expected Behavior 'cache', got %q", disableFlag.Behavior)
		}
		if disableFlag.IsEnableFlag {
			t.Error("Expected IsEnableFlag to be false for disable flag")
		}
		if disableFlag.PairFlagName != "with-cache" {
			t.Errorf("Expected PairFlagName 'with-cache', got %q", disableFlag.PairFlagName)
		}
		if disableFlag.DefaultValue != "false" {
			t.Errorf("Expected DefaultValue 'false', got %q", disableFlag.DefaultValue)
		}
	})

	t.Run("handles environment-aware defaults", func(t *testing.T) {
		def := DeclarativeFlagDefLike{
			Behavior:    "tui",
			EnableFlag:  "--with-tui",
			DisableFlag: "--no-tui",
			DefaultOn:   true,
			EnvAware:    true,
			EnvDefaults: map[string]bool{
				"local": true,
				"CI":    false,
			},
			Description: "Terminal UI",
		}

		flags := BuildDeclarativeMetadata(def)

		enableFlag := flags[0]
		if !enableFlag.EnvAware {
			t.Error("Expected EnvAware to be true")
		}
		if len(enableFlag.EnvDefaults) != 2 {
			t.Errorf("Expected 2 env defaults, got %d", len(enableFlag.EnvDefaults))
		}
		if !enableFlag.EnvDefaults["local"] {
			t.Error("Expected EnvDefaults['local'] to be true")
		}
		if enableFlag.EnvDefaults["CI"] {
			t.Error("Expected EnvDefaults['CI'] to be false")
		}

		// Disable flag should not have env defaults
		disableFlag := flags[1]
		if disableFlag.EnvAware {
			t.Error("Expected EnvAware to be false for disable flag")
		}
	})

	t.Run("handles default off", func(t *testing.T) {
		def := DeclarativeFlagDefLike{
			Behavior:    "debug",
			EnableFlag:  "--with-debug",
			DisableFlag: "--no-debug",
			DefaultOn:   false,
			Description: "Debug mode",
		}

		flags := BuildDeclarativeMetadata(def)

		enableFlag := flags[0]
		if enableFlag.DefaultValue != "false" {
			t.Errorf("Expected DefaultValue 'false', got %q", enableFlag.DefaultValue)
		}
	})
}

func TestStripFlagPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"--with-cache", "with-cache"},
		{"--no-cache", "no-cache"},
		{"-v", "-v"}, // Single dash not stripped
		{"flag", "flag"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := stripFlagPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("stripFlagPrefix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBoolToString(t *testing.T) {
	if boolToString(true) != "true" {
		t.Errorf("boolToString(true) = %q, want 'true'", boolToString(true))
	}
	if boolToString(false) != "false" {
		t.Errorf("boolToString(false) = %q, want 'false'", boolToString(false))
	}
}
