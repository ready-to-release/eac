package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestToPersonalPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"repository.yml", "repository.personal.yml"},
		{"tools.yml", "tools.personal.yml"},
		{"/path/to/config.yml", "/path/to/config.personal.yml"},
		{"config.yaml", "config.personal.yaml"},
		{"/path/to/file.json", "/path/to/file.personal.json"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToPersonalPath(tt.input)
			if got != tt.want {
				t.Errorf("ToPersonalPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMergeYAML(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		override string
		wantKey  string
		wantVal  interface{}
	}{
		{
			name:     "simple override",
			base:     "key: base",
			override: "key: override",
			wantKey:  "key",
			wantVal:  "override",
		},
		{
			name:     "add new key",
			base:     "existing: value",
			override: "new: added",
			wantKey:  "new",
			wantVal:  "added",
		},
		{
			name:     "empty override returns base",
			base:     "key: value",
			override: "",
			wantKey:  "key",
			wantVal:  "value",
		},
		{
			name:     "empty base returns override",
			base:     "",
			override: "key: value",
			wantKey:  "key",
			wantVal:  "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged, err := MergeYAML([]byte(tt.base), []byte(tt.override))
			if err != nil {
				t.Fatalf("MergeYAML() error = %v", err)
			}

			// Parse merged result
			var result map[string]interface{}
			if err := parseYAML(merged, &result); err != nil {
				t.Fatalf("parsing merged result: %v", err)
			}

			got, ok := result[tt.wantKey]
			if !ok {
				t.Errorf("merged result missing key %q", tt.wantKey)
				return
			}

			if got != tt.wantVal {
				t.Errorf("merged[%q] = %v, want %v", tt.wantKey, got, tt.wantVal)
			}
		})
	}
}

func TestMergeYAML_DeepMerge(t *testing.T) {
	base := `
nested:
  key1: value1
  key2: value2
`
	override := `
nested:
  key2: overridden
  key3: added
`
	merged, err := MergeYAML([]byte(base), []byte(override))
	if err != nil {
		t.Fatalf("MergeYAML() error = %v", err)
	}

	var result map[string]interface{}
	if err := parseYAML(merged, &result); err != nil {
		t.Fatalf("parsing merged result: %v", err)
	}

	nested, ok := result["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("nested is not a map")
	}

	// Check key1 preserved from base
	if nested["key1"] != "value1" {
		t.Errorf("nested.key1 = %v, want %v", nested["key1"], "value1")
	}

	// Check key2 overridden
	if nested["key2"] != "overridden" {
		t.Errorf("nested.key2 = %v, want %v", nested["key2"], "overridden")
	}

	// Check key3 added
	if nested["key3"] != "added" {
		t.Errorf("nested.key3 = %v, want %v", nested["key3"], "added")
	}
}

func TestLoadWithPersonal(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create base config
	baseContent := `
name: base-name
version: "1.0"
settings:
  debug: false
  timeout: 30
`
	baseFile := filepath.Join(tmpDir, "config.yml")
	if err := os.WriteFile(baseFile, []byte(baseContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test without personal config
	t.Run("without personal", func(t *testing.T) {
		type Config struct {
			Name     string `yaml:"name"`
			Version  string `yaml:"version"`
			Settings struct {
				Debug   bool `yaml:"debug"`
				Timeout int  `yaml:"timeout"`
			} `yaml:"settings"`
		}

		cfg, err := LoadWithPersonal[Config](tmpDir, "config.yml")
		if err != nil {
			t.Fatalf("LoadWithPersonal() error = %v", err)
		}

		if cfg.Name != "base-name" {
			t.Errorf("Name = %q, want %q", cfg.Name, "base-name")
		}
		if cfg.Settings.Timeout != 30 {
			t.Errorf("Settings.Timeout = %d, want %d", cfg.Settings.Timeout, 30)
		}
	})

	// Create personal config
	personalContent := `
name: personal-name
settings:
  debug: true
`
	personalFile := ToPersonalPath(baseFile)
	if err := os.WriteFile(personalFile, []byte(personalContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with personal config
	t.Run("with personal", func(t *testing.T) {
		type Config struct {
			Name     string `yaml:"name"`
			Version  string `yaml:"version"`
			Settings struct {
				Debug   bool `yaml:"debug"`
				Timeout int  `yaml:"timeout"`
			} `yaml:"settings"`
		}

		cfg, err := LoadWithPersonal[Config](tmpDir, "config.yml")
		if err != nil {
			t.Fatalf("LoadWithPersonal() error = %v", err)
		}

		// Name should be overridden
		if cfg.Name != "personal-name" {
			t.Errorf("Name = %q, want %q", cfg.Name, "personal-name")
		}

		// Version should be preserved from base
		if cfg.Version != "1.0" {
			t.Errorf("Version = %q, want %q", cfg.Version, "1.0")
		}

		// Debug should be overridden
		if !cfg.Settings.Debug {
			t.Errorf("Settings.Debug = %v, want %v", cfg.Settings.Debug, true)
		}

		// Timeout should be preserved from base
		if cfg.Settings.Timeout != 30 {
			t.Errorf("Settings.Timeout = %d, want %d", cfg.Settings.Timeout, 30)
		}
	})
}

func TestHasPersonalConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create base config
	baseFile := filepath.Join(tmpDir, "config.yml")
	if err := os.WriteFile(baseFile, []byte("key: value"), 0644); err != nil {
		t.Fatal(err)
	}

	// No personal config yet
	if HasPersonalConfig(tmpDir, "config.yml") {
		t.Error("HasPersonalConfig() = true, want false (no personal file)")
	}

	// Create personal config
	personalFile := ToPersonalPath(baseFile)
	if err := os.WriteFile(personalFile, []byte("key: override"), 0644); err != nil {
		t.Fatal(err)
	}

	// Now should have personal config
	if !HasPersonalConfig(tmpDir, "config.yml") {
		t.Error("HasPersonalConfig() = false, want true (personal file exists)")
	}
}

func TestGetPersonalConfigInfo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create base config only
	baseFile := filepath.Join(tmpDir, "base.yml")
	if err := os.WriteFile(baseFile, []byte("key: value"), 0644); err != nil {
		t.Fatal(err)
	}

	info := GetPersonalConfigInfo(tmpDir, "base.yml")
	if !info.BaseExists {
		t.Error("BaseExists = false, want true")
	}
	if info.HasPersonal {
		t.Error("HasPersonal = true, want false")
	}

	// Create personal config
	if err := os.WriteFile(info.PersonalPath, []byte("key: override"), 0644); err != nil {
		t.Fatal(err)
	}

	info = GetPersonalConfigInfo(tmpDir, "base.yml")
	if !info.HasPersonal {
		t.Error("HasPersonal = false, want true")
	}
}

// parseYAML is a helper for tests
func parseYAML(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
