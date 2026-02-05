package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectMainPackages tests detection of main.go files in various locations.
func TestDetectMainPackages(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, root string) // creates test files
		moniker string
		want    []mainLocation
	}{
		{
			name: "detects main.go at root",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "main.go"), "package main")
			},
			moniker: "r2r-cli",
			want: []mainLocation{
				{name: "r2r-cli", path: "main.go"},
			},
		},
		{
			name: "detects main.go in cmd directory",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "cmd", "main.go"), "package main")
			},
			moniker: "eac",
			want: []mainLocation{
				{name: "eac", path: "cmd/main.go"},
			},
		},
		{
			name: "detects multiple binaries in cmd/*/main.go",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "cmd", "server", "main.go"), "package main")
				writeFile(t, filepath.Join(root, "cmd", "client", "main.go"), "package main")
				writeFile(t, filepath.Join(root, "cmd", "admin", "main.go"), "package main")
			},
			moniker: "myapp",
			want: []mainLocation{
				{name: "myapp-admin", path: "cmd/admin/main.go"},
				{name: "myapp-client", path: "cmd/client/main.go"},
				{name: "myapp-server", path: "cmd/server/main.go"},
			},
		},
		{
			name: "prefers root main.go over cmd/main.go",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "main.go"), "package main")
				writeFile(t, filepath.Join(root, "cmd", "main.go"), "package main")
			},
			moniker: "app",
			want: []mainLocation{
				{name: "app", path: "main.go"},
			},
		},
		{
			name: "prefers cmd/main.go over cmd/*/main.go",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "cmd", "main.go"), "package main")
				writeFile(t, filepath.Join(root, "cmd", "server", "main.go"), "package main")
			},
			moniker: "app",
			want: []mainLocation{
				{name: "app", path: "cmd/main.go"},
			},
		},
		{
			name: "returns empty when no main.go found",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "lib.go"), "package mylib")
			},
			moniker: "mylib",
			want:    nil,
		},
		{
			name: "ignores non-directory entries in cmd/",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "cmd", "README.md"), "# Commands")
				writeFile(t, filepath.Join(root, "cmd", "server", "main.go"), "package main")
			},
			moniker: "app",
			want: []mainLocation{
				{name: "app-server", path: "cmd/server/main.go"},
			},
		},
		{
			name: "ignores directories without main.go in cmd/*/",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "cmd", "server", "main.go"), "package main")
				writeFile(t, filepath.Join(root, "cmd", "internal", "helper.go"), "package internal")
			},
			moniker: "app",
			want: []mainLocation{
				{name: "app-server", path: "cmd/server/main.go"},
			},
		},
		{
			name: "handles empty directory",
			setup: func(t *testing.T, root string) {
				// No files created
			},
			moniker: "empty",
			want:    nil,
		},
		{
			name: "handles missing cmd directory",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "lib.go"), "package lib")
			},
			moniker: "lib",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			got := detectMainPackages(root, tt.moniker)

			// Sort results for deterministic comparison
			sortMainLocations(got)
			sortMainLocations(tt.want)

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestInferGoArtifacts_SingleMain tests artifact inference for a single main.go at root.
func TestInferGoArtifacts_SingleMain(t *testing.T) {
	root := t.TempDir()
	goRoot := filepath.Join(root, "go", "myapp")
	require.NoError(t, os.MkdirAll(goRoot, 0755))
	writeFile(t, filepath.Join(goRoot, "main.go"), "package main")

	compTypes := &ComponentTypesConfig{
		ComponentTypes: map[string]*ComponentType{
			"go": {
				Builders: []string{"go"},
				Defaults: &ComponentTypeDefaults{
					ArtifactPattern: "{moniker}{ext}",
				},
			},
		},
	}

	got := InferGoArtifacts("myapp", "go/myapp", root, compTypes)

	require.Len(t, got, 1)
	assert.Equal(t, "myapp", got[0].ID)
	assert.Equal(t, "executable", got[0].Type)
	assert.Equal(t, "{moniker}{ext}", got[0].Pattern)
}

// TestInferGoArtifacts_CmdMain tests artifact inference for main.go in cmd/ directory.
func TestInferGoArtifacts_CmdMain(t *testing.T) {
	root := t.TempDir()
	goRoot := filepath.Join(root, "go", "cli", "eac")
	require.NoError(t, os.MkdirAll(filepath.Join(goRoot, "cmd"), 0755))
	writeFile(t, filepath.Join(goRoot, "cmd", "main.go"), "package main")

	compTypes := &ComponentTypesConfig{
		ComponentTypes: map[string]*ComponentType{
			"go": {
				Builders: []string{"go"},
				Defaults: &ComponentTypeDefaults{
					ArtifactPattern: "{moniker}{ext}",
				},
			},
		},
	}

	got := InferGoArtifacts("eac", "go/cli/eac", root, compTypes)

	require.Len(t, got, 1)
	assert.Equal(t, "eac", got[0].ID)
	assert.Equal(t, "executable", got[0].Type)
	assert.Equal(t, "{moniker}{ext}", got[0].Pattern)
}

// TestInferGoArtifacts_MultiMain tests artifact inference for multiple binaries in cmd/*/main.go.
func TestInferGoArtifacts_MultiMain(t *testing.T) {
	root := t.TempDir()
	goRoot := filepath.Join(root, "go", "multibin")
	cmdDir := filepath.Join(goRoot, "cmd")
	require.NoError(t, os.MkdirAll(filepath.Join(cmdDir, "server"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(cmdDir, "client"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(cmdDir, "admin"), 0755))

	writeFile(t, filepath.Join(cmdDir, "server", "main.go"), "package main")
	writeFile(t, filepath.Join(cmdDir, "client", "main.go"), "package main")
	writeFile(t, filepath.Join(cmdDir, "admin", "main.go"), "package main")

	compTypes := &ComponentTypesConfig{
		ComponentTypes: map[string]*ComponentType{
			"go": {
				Builders: []string{"go"},
				Defaults: &ComponentTypeDefaults{
					ArtifactPattern: "{moniker}{ext}",
				},
			},
		},
	}

	got := InferGoArtifacts("multibin", "go/multibin", root, compTypes)

	require.Len(t, got, 3)

	// Sort by ID for deterministic comparison
	sortArtifactsByID(got)

	assert.Equal(t, "multibin-admin", got[0].ID)
	assert.Equal(t, "executable", got[0].Type)
	assert.Equal(t, "{moniker}{ext}", got[0].Pattern)

	assert.Equal(t, "multibin-client", got[1].ID)
	assert.Equal(t, "executable", got[1].Type)

	assert.Equal(t, "multibin-server", got[2].ID)
	assert.Equal(t, "executable", got[2].Type)
}

// TestInferGoArtifacts_SkipsExplicitArtifacts tests that inference is skipped when artifacts already exist.
func TestInferGoArtifacts_SkipsExplicitArtifacts(t *testing.T) {
	root := t.TempDir()
	goRoot := filepath.Join(root, "go", "r2r", "cli")
	require.NoError(t, os.MkdirAll(goRoot, 0755))
	writeFile(t, filepath.Join(goRoot, "main.go"), "package main")

	compTypes := &ComponentTypesConfig{
		ComponentTypes: map[string]*ComponentType{
			"go": {
				Builders: []string{"go"},
				Defaults: &ComponentTypeDefaults{
					ArtifactPattern: "{moniker}{ext}",
				},
			},
		},
	}

	// Create a module with explicit artifacts already defined
	module := &Module{
		Moniker: "r2r-cli",
		Components: ModuleComponents{
			"go": &ComponentEntry{
				Root: "go/cli/r2r",
				Build: &ModuleBuild{
					Artifacts: []ModuleArtifact{
						{ID: "r2r", Type: "executable", Pattern: "custom-{moniker}"},
					},
				},
			},
		},
	}

	// When explicit artifacts exist, InferGoArtifacts should return nil
	got := InferGoArtifactsForModule(module, root, compTypes)
	assert.Nil(t, got, "Should return nil when explicit artifacts exist")
}

// TestInferGoArtifacts_UsesPatternFromComponentTypes tests that artifact_pattern from defaults is used.
func TestInferGoArtifacts_UsesPatternFromComponentTypes(t *testing.T) {
	tests := []struct {
		name            string
		artifactPattern string
		wantPattern     string
	}{
		{
			name:            "uses default pattern from component types",
			artifactPattern: "{moniker}{ext}",
			wantPattern:     "{moniker}{ext}",
		},
		{
			name:            "uses custom pattern with prefix",
			artifactPattern: "bin-{moniker}{ext}",
			wantPattern:     "bin-{moniker}{ext}",
		},
		{
			name:            "uses pattern with suffix",
			artifactPattern: "{moniker}-cli{ext}",
			wantPattern:     "{moniker}-cli{ext}",
		},
		{
			name:            "uses complex pattern",
			artifactPattern: "out/{moniker}/{moniker}{ext}",
			wantPattern:     "out/{moniker}/{moniker}{ext}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			goRoot := filepath.Join(root, "go", "app")
			require.NoError(t, os.MkdirAll(goRoot, 0755))
			writeFile(t, filepath.Join(goRoot, "main.go"), "package main")

			compTypes := &ComponentTypesConfig{
				ComponentTypes: map[string]*ComponentType{
					"go": {
						Builders: []string{"go"},
						Defaults: &ComponentTypeDefaults{
							ArtifactPattern: tt.artifactPattern,
						},
					},
				},
			}

			got := InferGoArtifacts("app", "go/app", root, compTypes)

			require.Len(t, got, 1)
			assert.Equal(t, tt.wantPattern, got[0].Pattern)
		})
	}
}

// TestInferGoArtifacts_NilComponentTypes tests behavior when ComponentTypesConfig is nil.
func TestInferGoArtifacts_NilComponentTypes(t *testing.T) {
	root := t.TempDir()
	goRoot := filepath.Join(root, "go", "app")
	require.NoError(t, os.MkdirAll(goRoot, 0755))
	writeFile(t, filepath.Join(goRoot, "main.go"), "package main")

	// Should use fallback pattern when compTypes is nil
	got := InferGoArtifacts("app", "go/app", root, nil)

	require.Len(t, got, 1)
	assert.Equal(t, "app", got[0].ID)
	assert.Equal(t, "executable", got[0].Type)
	// When nil, should use a default pattern
	assert.Equal(t, "{moniker}{ext}", got[0].Pattern)
}

// TestInferGoArtifacts_MissingGoComponentType tests behavior when "go" component type is not defined.
func TestInferGoArtifacts_MissingGoComponentType(t *testing.T) {
	root := t.TempDir()
	goRoot := filepath.Join(root, "go", "app")
	require.NoError(t, os.MkdirAll(goRoot, 0755))
	writeFile(t, filepath.Join(goRoot, "main.go"), "package main")

	compTypes := &ComponentTypesConfig{
		ComponentTypes: map[string]*ComponentType{
			// No "go" type defined
			"typescript": {
				Builders: []string{"npm"},
			},
		},
	}

	// Should use fallback pattern when "go" type not found
	got := InferGoArtifacts("app", "go/app", root, compTypes)

	require.Len(t, got, 1)
	assert.Equal(t, "{moniker}{ext}", got[0].Pattern)
}

// TestInferGoArtifacts_NoDefaults tests behavior when "go" type has no defaults.
func TestInferGoArtifacts_NoDefaults(t *testing.T) {
	root := t.TempDir()
	goRoot := filepath.Join(root, "go", "app")
	require.NoError(t, os.MkdirAll(goRoot, 0755))
	writeFile(t, filepath.Join(goRoot, "main.go"), "package main")

	compTypes := &ComponentTypesConfig{
		ComponentTypes: map[string]*ComponentType{
			"go": {
				Builders: []string{"go"},
				Defaults: nil, // No defaults
			},
		},
	}

	got := InferGoArtifacts("app", "go/app", root, compTypes)

	require.Len(t, got, 1)
	assert.Equal(t, "{moniker}{ext}", got[0].Pattern)
}

// TestInferGoArtifacts_EmptyArtifactPattern tests behavior when artifact_pattern is empty.
func TestInferGoArtifacts_EmptyArtifactPattern(t *testing.T) {
	root := t.TempDir()
	goRoot := filepath.Join(root, "go", "app")
	require.NoError(t, os.MkdirAll(goRoot, 0755))
	writeFile(t, filepath.Join(goRoot, "main.go"), "package main")

	compTypes := &ComponentTypesConfig{
		ComponentTypes: map[string]*ComponentType{
			"go": {
				Builders: []string{"go"},
				Defaults: &ComponentTypeDefaults{
					ArtifactPattern: "", // Empty pattern
				},
			},
		},
	}

	got := InferGoArtifacts("app", "go/app", root, compTypes)

	require.Len(t, got, 1)
	// Should use fallback pattern when artifact_pattern is empty
	assert.Equal(t, "{moniker}{ext}", got[0].Pattern)
}

// TestInferGoArtifacts_LibraryModule tests that no artifacts are inferred for library modules.
func TestInferGoArtifacts_LibraryModule(t *testing.T) {
	root := t.TempDir()
	goRoot := filepath.Join(root, "go", "eac", "core")
	require.NoError(t, os.MkdirAll(goRoot, 0755))
	// Library with no main.go
	writeFile(t, filepath.Join(goRoot, "config.go"), "package config")
	writeFile(t, filepath.Join(goRoot, "config_test.go"), "package config")

	compTypes := &ComponentTypesConfig{
		ComponentTypes: map[string]*ComponentType{
			"go": {
				Builders: []string{"go"},
				Defaults: &ComponentTypeDefaults{
					ArtifactPattern: "{moniker}{ext}",
				},
			},
		},
	}

	got := InferGoArtifacts("core", "go/eac/core", root, compTypes)

	assert.Nil(t, got, "Library modules should not have inferred artifacts")
}

// TestInferGoArtifacts_RelativePaths tests that goRoot is correctly resolved against repoRoot.
func TestInferGoArtifacts_RelativePaths(t *testing.T) {
	root := t.TempDir()

	// Create nested structure
	goRoot := filepath.Join(root, "go", "nested", "deep", "app")
	require.NoError(t, os.MkdirAll(goRoot, 0755))
	writeFile(t, filepath.Join(goRoot, "main.go"), "package main")

	compTypes := &ComponentTypesConfig{
		ComponentTypes: map[string]*ComponentType{
			"go": {
				Builders: []string{"go"},
				Defaults: &ComponentTypeDefaults{
					ArtifactPattern: "{moniker}{ext}",
				},
			},
		},
	}

	got := InferGoArtifacts("app", "go/nested/deep/app", root, compTypes)

	require.Len(t, got, 1)
	assert.Equal(t, "app", got[0].ID)
}

// TestDetectMainPackages_SortsResults tests that cmd/*/main.go results are sorted alphabetically.
func TestDetectMainPackages_SortsResults(t *testing.T) {
	root := t.TempDir()
	cmdDir := filepath.Join(root, "cmd")

	// Create in non-alphabetical order
	dirs := []string{"zebra", "alpha", "mike", "bravo"}
	for _, d := range dirs {
		dir := filepath.Join(cmdDir, d)
		require.NoError(t, os.MkdirAll(dir, 0755))
		writeFile(t, filepath.Join(dir, "main.go"), "package main")
	}

	got := detectMainPackages(root, "app")

	// Results should be sorted alphabetically by full name (moniker-component)
	require.Len(t, got, 4)
	assert.Equal(t, "app-alpha", got[0].name)
	assert.Equal(t, "app-bravo", got[1].name)
	assert.Equal(t, "app-mike", got[2].name)
	assert.Equal(t, "app-zebra", got[3].name)
}

// Helper functions

// writeFile creates a file with the given content, creating parent directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}

// sortMainLocations sorts a slice of mainLocation by name for deterministic comparison.
func sortMainLocations(locs []mainLocation) {
	if len(locs) <= 1 {
		return
	}
	// Simple bubble sort for small slices
	for i := 0; i < len(locs)-1; i++ {
		for j := 0; j < len(locs)-i-1; j++ {
			if locs[j].name > locs[j+1].name {
				locs[j], locs[j+1] = locs[j+1], locs[j]
			}
		}
	}
}

// sortArtifactsByID sorts a slice of ModuleArtifact by ID for deterministic comparison.
func sortArtifactsByID(artifacts []ModuleArtifact) {
	if len(artifacts) <= 1 {
		return
	}
	// Simple bubble sort for small slices
	for i := 0; i < len(artifacts)-1; i++ {
		for j := 0; j < len(artifacts)-i-1; j++ {
			if artifacts[j].ID > artifacts[j+1].ID {
				artifacts[j], artifacts[j+1] = artifacts[j+1], artifacts[j]
			}
		}
	}
}

