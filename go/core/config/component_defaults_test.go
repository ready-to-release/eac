package config

import (
	"testing"
)

func TestDeriveBookName(t *testing.T) {
	tests := []struct {
		name     string
		compName string
		want     string
	}{
		{"tutorials-base derives tutorials", "tutorials-base", "tutorials"},
		{"howto-base derives howto", "howto-base", "howto"},
		{"reference-eac-base derives reference-eac", "reference-eac-base", "reference-eac"},
		{"base-site requires explicit config", "base-site", ""},
		{"tutorials is book name", "tutorials", "tutorials"},
		{"site is book name", "site", "site"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveBookName(tt.compName)
			if got != tt.want {
				t.Errorf("deriveBookName(%q) = %q, want %q", tt.compName, got, tt.want)
			}
		})
	}
}

func TestApplyComponentDefaults(t *testing.T) {
	// Create a module with docs-pdf component
	m := &Module{
		Moniker: "books",
		Components: ModuleComponents{
			"tutorials-pdf": &ComponentEntry{Type: ComponentTypeDocsPdf},
		},
	}

	// Create component type defaults
	docsPdfDefaults := &ComponentTypeDefaults{
		BookFromName: true,
		Theme:        "dark",
	}

	// Apply defaults for docs-pdf
	docsPdfCT := &ComponentType{Defaults: docsPdfDefaults}
	entry := m.Components["tutorials-pdf"]
	m.applyComponentDefaults("tutorials-pdf", entry, docsPdfCT)

	// Verify book was derived (strips -pdf suffix)
	if entry.Config == nil || entry.Config["book"] != "tutorials-pdf" {
		t.Errorf("Expected book='tutorials-pdf', got %v", entry.Config)
	}

	// Verify theme was set
	if entry.Config["theme"] != "dark" {
		t.Errorf("Expected theme='dark', got %v", entry.Config)
	}
}

func TestComponentEntryConfigMethods(t *testing.T) {
	entry := &ComponentEntry{
		Config: map[string]string{
			"book":  "tutorials",
			"theme": "light",
		},
		DependsOn: []string{"tutorials-base"},
	}

	if got := entry.GetBook(); got != "tutorials" {
		t.Errorf("GetBook() = %q, want 'tutorials'", got)
	}

	if got := entry.GetTheme(); got != "light" {
		t.Errorf("GetTheme() = %q, want 'light'", got)
	}

	if got := entry.GetConfig("missing"); got != "" {
		t.Errorf("GetConfig('missing') = %q, want ''", got)
	}

	deps := entry.GetDependsOn()
	if len(deps) != 1 || deps[0] != "tutorials-base" {
		t.Errorf("GetDependsOn() = %v, want ['tutorials-base']", deps)
	}

	// Test nil entry
	var nilEntry *ComponentEntry
	if got := nilEntry.GetBook(); got != "" {
		t.Errorf("nil.GetBook() = %q, want ''", got)
	}
	if got := nilEntry.GetDependsOn(); got != nil {
		t.Errorf("nil.GetDependsOn() = %v, want nil", got)
	}
}


func TestComponentTypeDefaultsArtifactPattern(t *testing.T) {
	tests := []struct {
		name     string
		defaults *ComponentTypeDefaults
		moniker  string
		ext      string
		want     string
	}{
		{
			name: "pattern with moniker and ext placeholders",
			defaults: &ComponentTypeDefaults{
				ArtifactPattern: "{moniker}{ext}",
			},
			moniker: "clie",
			ext:     ".exe",
			want:    "clie.exe",
		},
		{
			name: "pattern without placeholders",
			defaults: &ComponentTypeDefaults{
				ArtifactPattern: "binary",
			},
			moniker: "any",
			ext:     ".exe",
			want:    "binary",
		},
		{
			name:     "nil defaults returns empty",
			defaults: nil,
			moniker:  "clie",
			ext:      ".exe",
			want:     "",
		},
		{
			name: "empty pattern returns empty",
			defaults: &ComponentTypeDefaults{
				ArtifactPattern: "",
			},
			moniker: "clie",
			ext:     ".exe",
			want:    "",
		},
		{
			name: "moniker only pattern",
			defaults: &ComponentTypeDefaults{
				ArtifactPattern: "{moniker}-cli",
			},
			moniker: "eac",
			ext:     "",
			want:    "eac-cli",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.defaults.GetArtifactName(tt.moniker, tt.ext)
			if got != tt.want {
				t.Errorf("GetArtifactName(%q, %q) = %q, want %q", tt.moniker, tt.ext, got, tt.want)
			}
		})
	}
}

func TestComponentTypeResources(t *testing.T) {
	t.Run("GetWeight returns cpus value", func(t *testing.T) {
		ct := &ComponentType{
			Resources: &ComponentTypeResources{
				CPUs:   4,
				Memory: "8g",
			},
		}
		if got := ct.GetWeight(); got != 4 {
			t.Errorf("GetWeight() = %d, want 4", got)
		}
	})

	t.Run("GetWeight returns 1 when no resources", func(t *testing.T) {
		ct := &ComponentType{}
		if got := ct.GetWeight(); got != 1 {
			t.Errorf("GetWeight() = %d, want 1", got)
		}
	})

	t.Run("GetWeight returns 1 when cpus is 0", func(t *testing.T) {
		ct := &ComponentType{
			Resources: &ComponentTypeResources{
				CPUs: 0,
			},
		}
		if got := ct.GetWeight(); got != 1 {
			t.Errorf("GetWeight() = %d, want 1", got)
		}
	})

	t.Run("GetMemory returns memory value", func(t *testing.T) {
		ct := &ComponentType{
			Resources: &ComponentTypeResources{
				Memory: "8g",
			},
		}
		if got := ct.GetMemory(); got != "8g" {
			t.Errorf("GetMemory() = %q, want '8g'", got)
		}
	})

	t.Run("GetMemory returns empty when no resources", func(t *testing.T) {
		ct := &ComponentType{}
		if got := ct.GetMemory(); got != "" {
			t.Errorf("GetMemory() = %q, want ''", got)
		}
	})
}

func TestDeriveChangelogPath(t *testing.T) {
	tests := []struct {
		name string
		m    *Module
		want string
	}{
		{
			name: "explicit changelog takes precedence",
			m: &Module{
				Moniker: "clie",
				Versioning: &ModuleVersioning{
					ReleaseType: "published",
					Changelog:   "custom/path/CHANGELOG.md",
				},
			},
			want: "custom/path/CHANGELOG.md",
		},
		{
			name: "published release type derives from moniker",
			m: &Module{
				Moniker: "clie",
				Versioning: &ModuleVersioning{
					ReleaseType: "published",
				},
			},
			want: "release/clie/CHANGELOG.md",
		},
		{
			name: "bundle release type derives from moniker",
			m: &Module{
				Moniker: "eac-ext",
				Versioning: &ModuleVersioning{
					ReleaseType: "bundle",
				},
			},
			want: "release/eac-ext/CHANGELOG.md",
		},
		{
			name: "internal with go component derives from go_root",
			m: &Module{
				Moniker: "core",
				Versioning: &ModuleVersioning{
					ReleaseType: "internal",
				},
				Components: ModuleComponents{
					"go": &ComponentEntry{Root: "go/eac/core"},
				},
			},
			want: "go/eac/core/CHANGELOG.md",
		},
		{
			name: "internal without go component returns empty",
			m: &Module{
				Moniker: "docs",
				Versioning: &ModuleVersioning{
					ReleaseType: "internal",
				},
				Components: ModuleComponents{
					"assets": &ComponentEntry{Root: "docs"},
				},
			},
			want: "",
		},
		{
			name: "no versioning returns empty",
			m: &Module{
				Moniker: "supporting",
			},
			want: "",
		},
		{
			name: "release_type none returns empty",
			m: &Module{
				Moniker: "testdata",
				Versioning: &ModuleVersioning{
					ReleaseType: "none",
				},
			},
			want: "",
		},
		{
			name: "empty release type returns empty",
			m: &Module{
				Moniker: "implicit",
				Versioning: &ModuleVersioning{
					Scheme: "Implicit",
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveChangelogPath(tt.m)
			if got != tt.want {
				t.Errorf("deriveChangelogPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
