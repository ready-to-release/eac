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

func TestDeriveBaseSiteName(t *testing.T) {
	tests := []struct {
		name       string
		compName   string
		components ModuleComponents
		want       string
	}{
		{
			name:     "tutorials finds tutorials-base",
			compName: "tutorials",
			components: ModuleComponents{
				"tutorials-base": &ComponentEntry{Type: "base-site"},
				"tutorials":      &ComponentEntry{Type: "pdf-render"},
			},
			want: "tutorials-base",
		},
		{
			name:     "site finds base-site as fallback",
			compName: "site",
			components: ModuleComponents{
				"base-site": &ComponentEntry{Type: "base-site"},
				"site":      &ComponentEntry{Type: "site-render"},
			},
			want: "base-site",
		},
		{
			name:     "no matching base component",
			compName: "orphan",
			components: ModuleComponents{
				"orphan": &ComponentEntry{Type: "pdf-render"},
			},
			want: "",
		},
		{
			name:     "prefers specific over fallback",
			compName: "tutorials",
			components: ModuleComponents{
				"tutorials-base": &ComponentEntry{Type: "base-site"},
				"base-site":      &ComponentEntry{Type: "base-site"},
				"tutorials":      &ComponentEntry{Type: "pdf-render"},
			},
			want: "tutorials-base",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveBaseSiteName(tt.compName, tt.components)
			if got != tt.want {
				t.Errorf("deriveBaseSiteName(%q, ...) = %q, want %q", tt.compName, got, tt.want)
			}
		})
	}
}

func TestApplyComponentDefaults(t *testing.T) {
	// Create a module with base-site and pdf-render components
	m := &Module{
		Moniker: "books",
		Components: ModuleComponents{
			"tutorials-base": &ComponentEntry{Type: "base-site"},
			"tutorials":      &ComponentEntry{Type: "pdf-render"},
		},
	}

	// Create component type defaults
	baseSiteDefaults := &ComponentTypeDefaults{
		BookFromName: true,
	}
	pdfRenderDefaults := &ComponentTypeDefaults{
		AutoBaseSite: true,
		Theme:        "dark",
	}

	// Apply defaults for base-site
	baseSiteCT := &ComponentType{Defaults: baseSiteDefaults}
	entry := m.Components["tutorials-base"]
	m.applyComponentDefaults("tutorials-base", entry, baseSiteCT)

	// Verify book was derived
	if entry.Config == nil || entry.Config["book"] != "tutorials" {
		t.Errorf("Expected book='tutorials', got %v", entry.Config)
	}

	// Apply defaults for pdf-render
	pdfRenderCT := &ComponentType{Defaults: pdfRenderDefaults}
	entry2 := m.Components["tutorials"]
	m.applyComponentDefaults("tutorials", entry2, pdfRenderCT)

	// Verify depends_on was auto-added
	if len(entry2.DependsOn) != 1 || entry2.DependsOn[0] != "tutorials-base" {
		t.Errorf("Expected depends_on=['tutorials-base'], got %v", entry2.DependsOn)
	}

	// Verify theme was set
	if entry2.Config == nil || entry2.Config["theme"] != "dark" {
		t.Errorf("Expected theme='dark', got %v", entry2.Config)
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

func TestInjectPrerequisiteComponents(t *testing.T) {
	t.Run("injects base-site for pdf-render when missing", func(t *testing.T) {
		m := &Module{
			Moniker: "books",
			Components: ModuleComponents{
				"tutorials": &ComponentEntry{Type: "pdf-render"},
			},
		}

		// Component types config with inject_prerequisite
		compTypes := &ComponentTypesConfig{
			ComponentTypes: map[string]*ComponentType{
				"pdf-render": {
					Defaults: &ComponentTypeDefaults{
						InjectPrerequisite: "base-site",
					},
				},
				"base-site": {
					Defaults: &ComponentTypeDefaults{
						BookFromName: true,
					},
				},
			},
		}

		// Inject prerequisites
		m.injectPrerequisiteComponents(compTypes)

		// Verify tutorials-base was injected
		if !m.Components.HasComponent("tutorials-base") {
			t.Fatal("Expected tutorials-base to be injected")
		}

		// Verify injected component has correct type
		injected := m.Components["tutorials-base"]
		if injected.Type != "base-site" {
			t.Errorf("Expected type='base-site', got %q", injected.Type)
		}

		// Verify injected component has book config
		if injected.Config == nil || injected.Config["book"] != "tutorials" {
			t.Errorf("Expected config.book='tutorials', got %v", injected.Config)
		}

		// Verify render component has depends_on set
		render := m.Components["tutorials"]
		if len(render.DependsOn) != 1 || render.DependsOn[0] != "tutorials-base" {
			t.Errorf("Expected depends_on=['tutorials-base'], got %v", render.DependsOn)
		}
	})

	t.Run("skips injection when prerequisite already exists", func(t *testing.T) {
		m := &Module{
			Moniker: "books",
			Components: ModuleComponents{
				"tutorials-base": &ComponentEntry{
					Type:   "base-site",
					Config: map[string]string{"book": "custom-book"},
				},
				"tutorials": &ComponentEntry{Type: "pdf-render"},
			},
		}

		compTypes := &ComponentTypesConfig{
			ComponentTypes: map[string]*ComponentType{
				"pdf-render": {
					Defaults: &ComponentTypeDefaults{
						InjectPrerequisite: "base-site",
					},
				},
			},
		}

		m.injectPrerequisiteComponents(compTypes)

		// Verify existing component was not modified
		existing := m.Components["tutorials-base"]
		if existing.Config["book"] != "custom-book" {
			t.Errorf("Existing component should not be modified, got book=%q", existing.Config["book"])
		}
	})

	t.Run("handles nil component types", func(t *testing.T) {
		m := &Module{
			Moniker: "books",
			Components: ModuleComponents{
				"tutorials": &ComponentEntry{Type: "pdf-render"},
			},
		}

		// Should not panic with nil compTypes
		m.injectPrerequisiteComponents(nil)

		// Should not inject anything
		if m.Components.HasComponent("tutorials-base") {
			t.Error("Should not inject when compTypes is nil")
		}
	})

	t.Run("multiple render components get individual prerequisites", func(t *testing.T) {
		m := &Module{
			Moniker: "books",
			Components: ModuleComponents{
				"tutorials": &ComponentEntry{Type: "pdf-render"},
				"howto":     &ComponentEntry{Type: "pdf-render"},
				"reference": &ComponentEntry{Type: "site-render"},
			},
		}

		compTypes := &ComponentTypesConfig{
			ComponentTypes: map[string]*ComponentType{
				"pdf-render": {
					Defaults: &ComponentTypeDefaults{
						InjectPrerequisite: "base-site",
					},
				},
				"site-render": {
					Defaults: &ComponentTypeDefaults{
						InjectPrerequisite: "base-site",
					},
				},
			},
		}

		m.injectPrerequisiteComponents(compTypes)

		// Verify all three base components were injected
		expected := []string{"tutorials-base", "howto-base", "reference-base"}
		for _, name := range expected {
			if !m.Components.HasComponent(name) {
				t.Errorf("Expected %s to be injected", name)
			}
		}
	})
}

func TestDerivePrerequisiteName(t *testing.T) {
	tests := []struct {
		compName   string
		prereqType string
		want       string
	}{
		{"tutorials", "base-site", "tutorials-base"},
		{"howto", "base-site", "howto-base"},
		{"reference", "base-site", "reference-base"},
		{"foo", "other-type", "foo-other-type"},
	}

	for _, tt := range tests {
		t.Run(tt.compName+"_"+tt.prereqType, func(t *testing.T) {
			got := derivePrerequisiteName(tt.compName, tt.prereqType)
			if got != tt.want {
				t.Errorf("derivePrerequisiteName(%q, %q) = %q, want %q",
					tt.compName, tt.prereqType, got, tt.want)
			}
		})
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
			moniker: "r2r",
			ext:     ".exe",
			want:    "r2r.exe",
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
			moniker:  "r2r",
			ext:      ".exe",
			want:     "",
		},
		{
			name: "empty pattern returns empty",
			defaults: &ComponentTypeDefaults{
				ArtifactPattern: "",
			},
			moniker: "r2r",
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
				Moniker: "r2r-cli",
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
				Moniker: "r2r-cli",
				Versioning: &ModuleVersioning{
					ReleaseType: "published",
				},
			},
			want: "release/r2r-cli/CHANGELOG.md",
		},
		{
			name: "bundle release type derives from moniker",
			m: &Module{
				Moniker: "ext-eac",
				Versioning: &ModuleVersioning{
					ReleaseType: "bundle",
				},
			},
			want: "release/ext-eac/CHANGELOG.md",
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
					"markdown": &ComponentEntry{Root: "docs"},
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
