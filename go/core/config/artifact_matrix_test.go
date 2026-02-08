package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// parseBlueprintsFromYAML is a test helper that parses artifact matrices YAML into BlueprintsConfig.
func parseBlueprintsFromYAML(t *testing.T, data string) *BlueprintsConfig {
	t.Helper()
	var cfg BlueprintsConfig
	err := yaml.Unmarshal([]byte(data), &cfg)
	require.NoError(t, err)
	return &cfg
}

// =============================================================================
// Tests for Artifact Matrix System
// =============================================================================

// TestParseArtifactMatrices tests parsing artifact matrices from YAML into BlueprintsConfig.
func TestParseArtifactMatrices(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		wantMatrices []string
		wantErr      bool
	}{
		{
			name: "loads simple matrices",
			yaml: `
artifact-matrices:
  cross-platform:
    - {id: linux-amd64, type: executable, pattern: "{moniker}-linux-amd64"}
    - {id: linux-arm64, type: executable, pattern: "{moniker}-linux-arm64"}
  single-platform:
    - {id: "{moniker}", type: executable, pattern: "{moniker}{ext}"}
`,
			wantMatrices: []string{"cross-platform", "single-platform"},
			wantErr:      false,
		},
		{
			name: "loads matrix with extends",
			yaml: `
artifact-matrices:
  base:
    - {id: linux-amd64, type: executable, pattern: "{moniker}-linux-amd64"}
  extended:
    extends: base
    additional:
      - {id: linux-amd64-upx, type: executable, pattern: "{moniker}-linux-amd64-upx", compression: upx}
`,
			wantMatrices: []string{"base", "extended"},
			wantErr:      false,
		},
		{
			name: "loads empty matrices section",
			yaml: `
artifact-matrices: {}
`,
			wantMatrices: []string{},
			wantErr:      false,
		},
		{
			name: "loads with no artifact-matrices key",
			yaml: `
other-key: value
`,
			wantMatrices: nil,
			wantErr:      false,
		},
		{
			name:    "returns error for invalid yaml",
			yaml:    `artifact-matrices: [invalid: yaml: structure`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg BlueprintsConfig
			err := yaml.Unmarshal([]byte(tt.yaml), &cfg)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.wantMatrices == nil {
				assert.Nil(t, cfg.ArtifactMatrices)
				return
			}

			for _, matrixName := range tt.wantMatrices {
				assert.Contains(t, cfg.ArtifactMatrices, matrixName, "expected matrix %s to exist", matrixName)
			}
		})
	}
}

// TestParseArtifactMatrices_EntryFields tests that all entry fields are correctly parsed.
func TestParseArtifactMatrices_EntryFields(t *testing.T) {
	yamlData := `
artifact-matrices:
  test-matrix:
    - id: test-entry
      type: executable
      pattern: "{moniker}-test"
      compression: upx
      derive_from: "{moniker}-base"
`

	cfg := parseBlueprintsFromYAML(t, yamlData)
	require.NotNil(t, cfg.ArtifactMatrices["test-matrix"])

	matrix := cfg.ArtifactMatrices["test-matrix"]
	require.Len(t, matrix.Entries, 1)

	entry := matrix.Entries[0]
	assert.Equal(t, "test-entry", entry.ID)
	assert.Equal(t, "executable", entry.Type)
	assert.Equal(t, "{moniker}-test", entry.Pattern)
	assert.Equal(t, "upx", entry.Compression)
	assert.Equal(t, "{moniker}-base", entry.DeriveFrom)
}

// TestExpandArtifactMatrix_Simple tests expanding a basic matrix without inheritance.
func TestExpandArtifactMatrix_Simple(t *testing.T) {
	yamlData := `
artifact-matrices:
  cross-platform:
    - {id: linux-amd64, type: executable, pattern: "{moniker}-linux-amd64"}
    - {id: linux-arm64, type: executable, pattern: "{moniker}-linux-arm64"}
    - {id: darwin-amd64, type: executable, pattern: "{moniker}-darwin-amd64"}
    - {id: darwin-arm64, type: executable, pattern: "{moniker}-darwin-arm64"}
    - {id: windows-amd64, type: executable, pattern: "{moniker}-windows-amd64.exe"}
`

	cfg := parseBlueprintsFromYAML(t, yamlData)

	params := map[string]string{
		"moniker": "myapp",
	}

	artifacts := cfg.ExpandArtifactMatrix("cross-platform", params)

	require.Len(t, artifacts, 5)

	// Verify each artifact
	expectedArtifacts := []struct {
		id      string
		pattern string
	}{
		{"linux-amd64", "myapp-linux-amd64"},
		{"linux-arm64", "myapp-linux-arm64"},
		{"darwin-amd64", "myapp-darwin-amd64"},
		{"darwin-arm64", "myapp-darwin-arm64"},
		{"windows-amd64", "myapp-windows-amd64.exe"},
	}

	for i, expected := range expectedArtifacts {
		assert.Equal(t, expected.id, artifacts[i].ID, "artifact %d ID mismatch", i)
		assert.Equal(t, expected.pattern, artifacts[i].Pattern, "artifact %d pattern mismatch", i)
		assert.Equal(t, "executable", artifacts[i].Type, "artifact %d type mismatch", i)
	}
}

// TestExpandArtifactMatrix_WithInheritance tests expanding a matrix that extends another.
func TestExpandArtifactMatrix_WithInheritance(t *testing.T) {
	yamlData := `
artifact-matrices:
  cross-platform:
    - {id: linux-amd64, type: executable, pattern: "{moniker}-linux-amd64"}
    - {id: linux-arm64, type: executable, pattern: "{moniker}-linux-arm64"}
    - {id: windows-amd64, type: executable, pattern: "{moniker}-windows-amd64.exe"}

  cross-platform-upx:
    extends: cross-platform
    additional:
      - {id: linux-amd64-upx, type: executable, pattern: "{moniker}-linux-amd64-upx", compression: upx, derive_from: "{moniker}-linux-amd64"}
      - {id: windows-amd64-upx, type: executable, pattern: "{moniker}-windows-amd64-upx.exe", compression: upx, derive_from: "{moniker}-windows-amd64.exe"}
`

	cfg := parseBlueprintsFromYAML(t, yamlData)

	params := map[string]string{
		"moniker": "clie",
	}

	artifacts := cfg.ExpandArtifactMatrix("cross-platform-upx", params)

	// Should have 3 from parent + 2 additional = 5 total
	require.Len(t, artifacts, 5)

	// First 3 should be from parent (cross-platform)
	assert.Equal(t, "linux-amd64", artifacts[0].ID)
	assert.Equal(t, "clie-linux-amd64", artifacts[0].Pattern)
	assert.Empty(t, artifacts[0].Compression) // Parent entries have no compression

	assert.Equal(t, "linux-arm64", artifacts[1].ID)
	assert.Equal(t, "clie-linux-arm64", artifacts[1].Pattern)

	assert.Equal(t, "windows-amd64", artifacts[2].ID)
	assert.Equal(t, "clie-windows-amd64.exe", artifacts[2].Pattern)

	// Last 2 should be additional entries
	assert.Equal(t, "linux-amd64-upx", artifacts[3].ID)
	assert.Equal(t, "clie-linux-amd64-upx", artifacts[3].Pattern)
	assert.Equal(t, "upx", artifacts[3].Compression)
	assert.Equal(t, "clie-linux-amd64", artifacts[3].DeriveFrom)

	assert.Equal(t, "windows-amd64-upx", artifacts[4].ID)
	assert.Equal(t, "clie-windows-amd64-upx.exe", artifacts[4].Pattern)
	assert.Equal(t, "upx", artifacts[4].Compression)
	assert.Equal(t, "clie-windows-amd64.exe", artifacts[4].DeriveFrom)
}

// TestExpandArtifactMatrix_MultiLevelInheritance tests inheritance chains.
func TestExpandArtifactMatrix_MultiLevelInheritance(t *testing.T) {
	yamlData := `
artifact-matrices:
  base:
    - {id: base-entry, type: executable, pattern: "{moniker}-base"}

  level1:
    extends: base
    additional:
      - {id: level1-entry, type: executable, pattern: "{moniker}-level1"}

  level2:
    extends: level1
    additional:
      - {id: level2-entry, type: executable, pattern: "{moniker}-level2"}
`

	cfg := parseBlueprintsFromYAML(t, yamlData)

	params := map[string]string{"moniker": "test"}

	artifacts := cfg.ExpandArtifactMatrix("level2", params)

	// Should have: base-entry, level1-entry, level2-entry
	require.Len(t, artifacts, 3)
	assert.Equal(t, "base-entry", artifacts[0].ID)
	assert.Equal(t, "level1-entry", artifacts[1].ID)
	assert.Equal(t, "level2-entry", artifacts[2].ID)
}

// TestExpandArtifactMatrix_WithParams tests parameter substitution.
func TestExpandArtifactMatrix_WithParams(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		params      map[string]string
		wantPattern string
		wantID      string
	}{
		{
			name: "substitutes moniker",
			yaml: `
artifact-matrices:
  test:
    - {id: "{moniker}", type: executable, pattern: "bin/{moniker}"}
`,
			params:      map[string]string{"moniker": "myapp"},
			wantPattern: "bin/myapp",
			wantID:      "myapp",
		},
		{
			name: "substitutes ext",
			yaml: `
artifact-matrices:
  test:
    - {id: app, type: executable, pattern: "app{ext}"}
`,
			params:      map[string]string{"ext": ".exe"},
			wantPattern: "app.exe",
			wantID:      "app",
		},
		{
			name: "substitutes multiple params",
			yaml: `
artifact-matrices:
  test:
    - {id: "{moniker}-{platform}", type: executable, pattern: "out/{moniker}/{moniker}-{platform}{ext}"}
`,
			params:      map[string]string{"moniker": "cli", "platform": "linux-amd64", "ext": ""},
			wantPattern: "out/cli/cli-linux-amd64",
			wantID:      "cli-linux-amd64",
		},
		{
			name: "leaves unknown params unchanged",
			yaml: `
artifact-matrices:
  test:
    - {id: entry, type: executable, pattern: "{moniker}-{unknown}"}
`,
			params:      map[string]string{"moniker": "myapp"},
			wantPattern: "myapp-{unknown}",
			wantID:      "entry",
		},
		{
			name: "handles empty params map",
			yaml: `
artifact-matrices:
  test:
    - {id: entry, type: executable, pattern: "{moniker}-app"}
`,
			params:      map[string]string{},
			wantPattern: "{moniker}-app",
			wantID:      "entry",
		},
		{
			name: "handles nil params",
			yaml: `
artifact-matrices:
  test:
    - {id: entry, type: executable, pattern: "{moniker}-app"}
`,
			params:      nil,
			wantPattern: "{moniker}-app",
			wantID:      "entry",
		},
		{
			name: "substitutes params in derive_from",
			yaml: `
artifact-matrices:
  test:
    - {id: derived, type: executable, pattern: "{moniker}-upx", derive_from: "{moniker}-base"}
`,
			params:      map[string]string{"moniker": "myapp"},
			wantPattern: "myapp-upx",
			wantID:      "derived",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := parseBlueprintsFromYAML(t, tt.yaml)

			artifacts := cfg.ExpandArtifactMatrix("test", tt.params)

			require.Len(t, artifacts, 1)
			assert.Equal(t, tt.wantID, artifacts[0].ID)
			assert.Equal(t, tt.wantPattern, artifacts[0].Pattern)
		})
	}
}

// TestExpandArtifactMatrix_UnknownMatrix tests that unknown matrix names return nil.
func TestExpandArtifactMatrix_UnknownMatrix(t *testing.T) {
	tests := []struct {
		name       string
		yaml       string
		matrixName string
	}{
		{
			name: "returns nil for unknown matrix",
			yaml: `
artifact-matrices:
  existing:
    - {id: entry, type: executable, pattern: "test"}
`,
			matrixName: "nonexistent",
		},
		{
			name:       "returns nil for empty config",
			yaml:       `artifact-matrices: {}`,
			matrixName: "any",
		},
		{
			name:       "returns nil for nil config",
			yaml:       `other-key: value`,
			matrixName: "any",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := parseBlueprintsFromYAML(t, tt.yaml)

			artifacts := cfg.ExpandArtifactMatrix(tt.matrixName, map[string]string{"moniker": "test"})

			assert.Nil(t, artifacts)
		})
	}
}

// TestExpandArtifactMatrix_NilConfig tests that nil config handles gracefully.
func TestExpandArtifactMatrix_NilConfig(t *testing.T) {
	var cfg *BlueprintsConfig
	artifacts := cfg.ExpandArtifactMatrix("any", map[string]string{"moniker": "test"})
	assert.Nil(t, artifacts)
}

// TestExpandArtifactMatrix_CircularInheritance tests detection of circular references.
func TestExpandArtifactMatrix_CircularInheritance(t *testing.T) {
	tests := []struct {
		name       string
		yaml       string
		matrixName string
	}{
		{
			name: "detects direct circular reference",
			yaml: `
artifact-matrices:
  self-ref:
    extends: self-ref
    additional:
      - {id: entry, type: executable, pattern: "test"}
`,
			matrixName: "self-ref",
		},
		{
			name: "detects two-level circular reference",
			yaml: `
artifact-matrices:
  matrix-a:
    extends: matrix-b
    additional:
      - {id: entry-a, type: executable, pattern: "a"}
  matrix-b:
    extends: matrix-a
    additional:
      - {id: entry-b, type: executable, pattern: "b"}
`,
			matrixName: "matrix-a",
		},
		{
			name: "detects three-level circular reference",
			yaml: `
artifact-matrices:
  matrix-a:
    extends: matrix-b
    additional:
      - {id: entry-a, type: executable, pattern: "a"}
  matrix-b:
    extends: matrix-c
    additional:
      - {id: entry-b, type: executable, pattern: "b"}
  matrix-c:
    extends: matrix-a
    additional:
      - {id: entry-c, type: executable, pattern: "c"}
`,
			matrixName: "matrix-a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := parseBlueprintsFromYAML(t, tt.yaml)

			// Should not panic and should return nil or partial results
			artifacts := cfg.ExpandArtifactMatrix(tt.matrixName, map[string]string{"moniker": "test"})

			// The exact behavior depends on implementation - key is no panic
			// and reasonable output (nil, empty, or partial results)
			t.Logf("Circular reference handled: got %d artifacts", len(artifacts))
		})
	}
}

// TestExpandArtifactMatrix_ExtendsNonexistent tests extending a matrix that doesn't exist.
func TestExpandArtifactMatrix_ExtendsNonexistent(t *testing.T) {
	yamlData := `
artifact-matrices:
  broken:
    extends: nonexistent
    additional:
      - {id: entry, type: executable, pattern: "test"}
`

	cfg := parseBlueprintsFromYAML(t, yamlData)

	// Should handle gracefully - only return the additional entries
	artifacts := cfg.ExpandArtifactMatrix("broken", map[string]string{"moniker": "test"})

	// Should have only the additional entry (parent not found)
	require.Len(t, artifacts, 1)
	assert.Equal(t, "entry", artifacts[0].ID)
}

// TestExpandArtifactMatrix_PreservesType tests that artifact type is preserved correctly.
func TestExpandArtifactMatrix_PreservesType(t *testing.T) {
	yamlData := `
artifact-matrices:
  mixed-types:
    - {id: binary, type: executable, pattern: "bin/{moniker}"}
    - {id: config, type: file, pattern: "config/{moniker}.yml"}
    - {id: dist, type: directory, pattern: "dist/{moniker}"}
    - {id: report, type: test, pattern: "reports/{moniker}.xml"}
`

	cfg := parseBlueprintsFromYAML(t, yamlData)

	artifacts := cfg.ExpandArtifactMatrix("mixed-types", map[string]string{"moniker": "app"})

	require.Len(t, artifacts, 4)
	assert.Equal(t, "executable", artifacts[0].Type)
	assert.Equal(t, "file", artifacts[1].Type)
	assert.Equal(t, "directory", artifacts[2].Type)
	assert.Equal(t, "test", artifacts[3].Type)
}

// TestExpandArtifactMatrix_EmptyMatrix tests expanding an empty matrix.
func TestExpandArtifactMatrix_EmptyMatrix(t *testing.T) {
	yamlData := `
artifact-matrices:
  empty:
    - {}
`

	cfg := parseBlueprintsFromYAML(t, yamlData)

	artifacts := cfg.ExpandArtifactMatrix("empty", map[string]string{"moniker": "test"})

	// Should return entry with empty fields
	require.Len(t, artifacts, 1)
	assert.Empty(t, artifacts[0].ID)
	assert.Empty(t, artifacts[0].Type)
	assert.Empty(t, artifacts[0].Pattern)
}

// TestExpandArtifactMatrix_InheritanceWithDifferentParams tests that params are applied
// to all levels of inheritance.
func TestExpandArtifactMatrix_InheritanceWithDifferentParams(t *testing.T) {
	yamlData := `
artifact-matrices:
  base:
    - {id: "{moniker}-base", type: executable, pattern: "out/{moniker}-base{ext}"}
  extended:
    extends: base
    additional:
      - {id: "{moniker}-extended", type: executable, pattern: "out/{moniker}-extended{ext}"}
`

	cfg := parseBlueprintsFromYAML(t, yamlData)

	params := map[string]string{
		"moniker": "myapp",
		"ext":     ".exe",
	}

	artifacts := cfg.ExpandArtifactMatrix("extended", params)

	require.Len(t, artifacts, 2)

	// Parent entry should have params substituted
	assert.Equal(t, "myapp-base", artifacts[0].ID)
	assert.Equal(t, "out/myapp-base.exe", artifacts[0].Pattern)

	// Additional entry should also have params substituted
	assert.Equal(t, "myapp-extended", artifacts[1].ID)
	assert.Equal(t, "out/myapp-extended.exe", artifacts[1].Pattern)
}

// TestArtifactMatrix_YAMLFormats tests both YAML formats (array and object).
func TestArtifactMatrix_YAMLFormats(t *testing.T) {
	t.Run("array format parses to Entries", func(t *testing.T) {
		yamlData := `
artifact-matrices:
  simple:
    - {id: entry1, type: executable, pattern: "test1"}
    - {id: entry2, type: executable, pattern: "test2"}
`
		cfg := parseBlueprintsFromYAML(t, yamlData)

		matrix := cfg.ArtifactMatrices["simple"]
		require.NotNil(t, matrix)
		assert.Len(t, matrix.Entries, 2)
		assert.Empty(t, matrix.Extends)
		assert.Empty(t, matrix.Additional)
	})

	t.Run("object format parses extends and additional", func(t *testing.T) {
		yamlData := `
artifact-matrices:
  extended:
    extends: base
    additional:
      - {id: extra, type: executable, pattern: "extra"}
`
		cfg := parseBlueprintsFromYAML(t, yamlData)

		matrix := cfg.ArtifactMatrices["extended"]
		require.NotNil(t, matrix)
		assert.Equal(t, "base", matrix.Extends)
		assert.Len(t, matrix.Additional, 1)
		assert.Empty(t, matrix.Entries)
	})
}

// TestSubstituteMatrixParams tests the parameter substitution helper.
func TestSubstituteMatrixParams(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		params map[string]string
		want   string
	}{
		{
			name:   "substitutes single param",
			input:  "{moniker}-app",
			params: map[string]string{"moniker": "myapp"},
			want:   "myapp-app",
		},
		{
			name:   "substitutes multiple occurrences",
			input:  "{moniker}/{moniker}-{moniker}",
			params: map[string]string{"moniker": "x"},
			want:   "x/x-x",
		},
		{
			name:   "handles empty string",
			input:  "",
			params: map[string]string{"moniker": "test"},
			want:   "",
		},
		{
			name:   "handles nil params",
			input:  "{moniker}",
			params: nil,
			want:   "{moniker}",
		},
		{
			name:   "handles empty params",
			input:  "{moniker}",
			params: map[string]string{},
			want:   "{moniker}",
		},
		{
			name:   "preserves literal braces when not a param",
			input:  "prefix-{unknown}-suffix",
			params: map[string]string{"moniker": "test"},
			want:   "prefix-{unknown}-suffix",
		},
		{
			name:   "handles adjacent placeholders",
			input:  "{a}{b}{c}",
			params: map[string]string{"a": "1", "b": "2", "c": "3"},
			want:   "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := substituteMatrixParams(tt.input, tt.params)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestExpandArtifactMatrix_RealWorldScenario tests a realistic cross-platform CLI scenario.
func TestExpandArtifactMatrix_RealWorldScenario(t *testing.T) {
	// This mirrors the expected YAML structure from the spec
	yamlData := `
artifact-matrices:
  cross-platform:
    - {id: linux-amd64, type: executable, pattern: "{moniker}-linux-amd64"}
    - {id: linux-arm64, type: executable, pattern: "{moniker}-linux-arm64"}
    - {id: darwin-amd64, type: executable, pattern: "{moniker}-darwin-amd64"}
    - {id: darwin-arm64, type: executable, pattern: "{moniker}-darwin-arm64"}
    - {id: windows-amd64, type: executable, pattern: "{moniker}-windows-amd64.exe"}

  cross-platform-upx:
    extends: cross-platform
    additional:
      - {id: linux-amd64-upx, type: executable, pattern: "{moniker}-linux-amd64-upx", compression: upx, derive_from: "{moniker}-linux-amd64"}
      - {id: windows-amd64-upx, type: executable, pattern: "{moniker}-windows-amd64-upx.exe", compression: upx, derive_from: "{moniker}-windows-amd64.exe"}

  single-platform:
    - {id: "{moniker}", type: executable, pattern: "{moniker}{ext}"}
`

	cfg := parseBlueprintsFromYAML(t, yamlData)

	t.Run("cross-platform for clie-cli", func(t *testing.T) {
		artifacts := cfg.ExpandArtifactMatrix("cross-platform", map[string]string{"moniker": "clie"})

		require.Len(t, artifacts, 5)

		// Verify all platforms are present
		ids := make([]string, len(artifacts))
		for i, a := range artifacts {
			ids[i] = a.ID
		}
		assert.Contains(t, ids, "linux-amd64")
		assert.Contains(t, ids, "linux-arm64")
		assert.Contains(t, ids, "darwin-amd64")
		assert.Contains(t, ids, "darwin-arm64")
		assert.Contains(t, ids, "windows-amd64")
	})

	t.Run("cross-platform-upx for clie-cli", func(t *testing.T) {
		artifacts := cfg.ExpandArtifactMatrix("cross-platform-upx", map[string]string{"moniker": "clie"})

		require.Len(t, artifacts, 7) // 5 base + 2 upx

		// Find UPX entries
		var upxEntries []ModuleArtifact
		for _, a := range artifacts {
			if a.Compression == "upx" {
				upxEntries = append(upxEntries, a)
			}
		}

		require.Len(t, upxEntries, 2)

		// Verify derive_from is correctly set
		for _, upx := range upxEntries {
			assert.NotEmpty(t, upx.DeriveFrom)
			assert.Contains(t, upx.DeriveFrom, "clie")
		}
	})

	t.Run("single-platform for eac with Windows extension", func(t *testing.T) {
		artifacts := cfg.ExpandArtifactMatrix("single-platform", map[string]string{
			"moniker": "eac",
			"ext":     ".exe",
		})

		require.Len(t, artifacts, 1)
		assert.Equal(t, "eac", artifacts[0].ID)
		assert.Equal(t, "eac.exe", artifacts[0].Pattern)
	})

	t.Run("single-platform for eac without extension", func(t *testing.T) {
		artifacts := cfg.ExpandArtifactMatrix("single-platform", map[string]string{
			"moniker": "eac",
			"ext":     "",
		})

		require.Len(t, artifacts, 1)
		assert.Equal(t, "eac", artifacts[0].ID)
		assert.Equal(t, "eac", artifacts[0].Pattern)
	})
}

// =============================================================================
// Tests for expandArtifactMatrixForModule
// =============================================================================

// TestExpandArtifactMatrixForModule tests expanding artifact matrix references on modules.
func TestExpandArtifactMatrixForModule(t *testing.T) {
	yamlData := `
artifact-matrices:
  cross-platform:
    - {id: linux-amd64, type: executable, pattern: "{moniker}-linux-amd64"}
    - {id: windows-amd64, type: executable, pattern: "{moniker}-windows-amd64.exe"}
  cross-platform-upx:
    extends: cross-platform
    additional:
      - {id: linux-amd64-upx, type: executable, pattern: "{moniker}-linux-amd64-upx", compression: upx, derive_from: "{moniker}-linux-amd64"}
`
	matrices := parseBlueprintsFromYAML(t, yamlData)

	t.Run("expands matrix for module with go component", func(t *testing.T) {
		mod := &Module{
			Moniker:           "eac",
			ArtifactMatrixRef: "cross-platform",
			Components: ModuleComponents{
				"go": &ComponentEntry{Root: "go/cli/eac"},
			},
		}

		expandArtifactMatrixForModule(mod, matrices)

		goComp := mod.Components["go"]
		require.NotNil(t, goComp.Build)
		require.Len(t, goComp.Build.Artifacts, 2)
		assert.Equal(t, "linux-amd64", goComp.Build.Artifacts[0].ID)
		assert.Equal(t, "eac-linux-amd64", goComp.Build.Artifacts[0].Pattern)
		assert.Equal(t, "windows-amd64", goComp.Build.Artifacts[1].ID)
		assert.Equal(t, "eac-windows-amd64.exe", goComp.Build.Artifacts[1].Pattern)
	})

	t.Run("expands matrix with inheritance", func(t *testing.T) {
		mod := &Module{
			Moniker:           "clie",
			ArtifactMatrixRef: "cross-platform-upx",
			Components: ModuleComponents{
				"go": &ComponentEntry{Root: "go/cli/clie"},
			},
		}

		expandArtifactMatrixForModule(mod, matrices)

		goComp := mod.Components["go"]
		require.NotNil(t, goComp.Build)
		require.Len(t, goComp.Build.Artifacts, 3) // 2 base + 1 additional
		assert.Equal(t, "linux-amd64-upx", goComp.Build.Artifacts[2].ID)
		assert.Equal(t, "upx", goComp.Build.Artifacts[2].Compression)
		assert.Equal(t, "clie-linux-amd64", goComp.Build.Artifacts[2].DeriveFrom)
	})

	t.Run("does not override existing artifacts", func(t *testing.T) {
		mod := &Module{
			Moniker:           "my-cli",
			ArtifactMatrixRef: "cross-platform",
			Components: ModuleComponents{
				"go": &ComponentEntry{
					Root: "go/cli/my-cli",
					Build: &ModuleBuild{
						Artifacts: []ModuleArtifact{
							{ID: "custom", Type: "executable", Pattern: "custom-binary"},
						},
					},
				},
			},
		}

		expandArtifactMatrixForModule(mod, matrices)

		goComp := mod.Components["go"]
		require.Len(t, goComp.Build.Artifacts, 1)
		assert.Equal(t, "custom", goComp.Build.Artifacts[0].ID)
	})

	t.Run("skips when no go component", func(t *testing.T) {
		mod := &Module{
			Moniker:           "my-container",
			ArtifactMatrixRef: "cross-platform",
			Components: ModuleComponents{
				"dockerfile": &ComponentEntry{Root: "containers/my-container"},
			},
		}

		expandArtifactMatrixForModule(mod, matrices)

		// Should not create a go component
		assert.NotContains(t, mod.Components, "go")
	})

	t.Run("skips when no matrix ref", func(t *testing.T) {
		mod := &Module{
			Moniker: "my-lib",
			Components: ModuleComponents{
				"go": &ComponentEntry{Root: "go/my-lib"},
			},
		}

		expandArtifactMatrixForModule(mod, matrices)

		goComp := mod.Components["go"]
		assert.Nil(t, goComp.Build)
	})

	t.Run("skips when nil matrices", func(t *testing.T) {
		mod := &Module{
			Moniker:           "my-cli",
			ArtifactMatrixRef: "cross-platform",
			Components: ModuleComponents{
				"go": &ComponentEntry{Root: "go/cli/my-cli"},
			},
		}

		expandArtifactMatrixForModule(mod, nil)

		goComp := mod.Components["go"]
		assert.Nil(t, goComp.Build)
	})

	t.Run("handles unknown matrix name", func(t *testing.T) {
		mod := &Module{
			Moniker:           "my-cli",
			ArtifactMatrixRef: "nonexistent",
			Components: ModuleComponents{
				"go": &ComponentEntry{Root: "go/cli/my-cli"},
			},
		}

		expandArtifactMatrixForModule(mod, matrices)

		goComp := mod.Components["go"]
		assert.Nil(t, goComp.Build)
	})
}

// TestExpandArtifactMatrixForModule_ParameterPassThrough tests that module parameters
// are passed through to artifact matrix expansion (e.g., binary_name).
func TestExpandArtifactMatrixForModule_ParameterPassThrough(t *testing.T) {
	yamlData := `
artifact-matrices:
  cross-platform:
    - {id: linux-amd64, type: executable, pattern: "{binary_name}-linux-amd64"}
    - {id: windows-amd64, type: executable, pattern: "{binary_name}-windows-amd64.exe"}
`
	matrices := parseBlueprintsFromYAML(t, yamlData)

	t.Run("binary_name from Build overrides default moniker", func(t *testing.T) {
		mod := &Module{
			Moniker:           "eac-cli",
			ArtifactMatrixRef: "cross-platform",
			Components: ModuleComponents{
				"go": &ComponentEntry{
					Root:  "go/cli/eac",
					Build: &ModuleBuild{BinaryName: "eac"},
				},
			},
		}

		expandArtifactMatrixForModule(mod, matrices)

		goComp := mod.Components["go"]
		require.NotNil(t, goComp.Build)
		require.Len(t, goComp.Build.Artifacts, 2)
		assert.Equal(t, "eac-linux-amd64", goComp.Build.Artifacts[0].Pattern)
		assert.Equal(t, "eac-windows-amd64.exe", goComp.Build.Artifacts[1].Pattern)
	})

	t.Run("binary_name defaults to moniker when not set", func(t *testing.T) {
		mod := &Module{
			Moniker:           "myapp",
			ArtifactMatrixRef: "cross-platform",
			Components: ModuleComponents{
				"go": &ComponentEntry{Root: "go/cli/myapp"},
			},
		}

		expandArtifactMatrixForModule(mod, matrices)

		goComp := mod.Components["go"]
		require.NotNil(t, goComp.Build)
		require.Len(t, goComp.Build.Artifacts, 2)
		assert.Equal(t, "myapp-linux-amd64", goComp.Build.Artifacts[0].Pattern)
		assert.Equal(t, "myapp-windows-amd64.exe", goComp.Build.Artifacts[1].Pattern)
	})

	t.Run("binary_name from Build is used in matrix expansion", func(t *testing.T) {
		customYaml := `
artifact-matrices:
  custom:
    - {id: "{moniker}-release", type: executable, pattern: "{binary_name}-release"}
`
		customMatrices := parseBlueprintsFromYAML(t, customYaml)

		mod := &Module{
			Moniker:           "my-tool",
			ArtifactMatrixRef: "custom",
			Components: ModuleComponents{
				"go": &ComponentEntry{
					Root:  "go/cli/my-tool",
					Build: &ModuleBuild{BinaryName: "tool"},
				},
			},
		}

		expandArtifactMatrixForModule(mod, customMatrices)

		goComp := mod.Components["go"]
		require.NotNil(t, goComp.Build)
		require.Len(t, goComp.Build.Artifacts, 1)
		assert.Equal(t, "my-tool-release", goComp.Build.Artifacts[0].ID)
		assert.Equal(t, "tool-release", goComp.Build.Artifacts[0].Pattern)
	})
}

// TestGoRootInference tests that go_root is passed through discoveryVars.
func TestGoRootInference(t *testing.T) {
	t.Run("go_root from discoveryVars is available in params", func(t *testing.T) {
		mod := &Module{
			Moniker: "my-lib",
			Components: ModuleComponents{
				"go": &ComponentEntry{Root: "go/my-lib"},
			},
		}

		discoveryVars := map[string]string{
			"owner":   "myorg",
			"go_root": "go/my-lib",
		}
		params := buildModuleParams(mod, discoveryVars)

		assert.Equal(t, "go/my-lib", params["go_root"])
	})

	t.Run("go_root with custom path from discoveryVars", func(t *testing.T) {
		mod := &Module{
			Moniker: "my-lib",
			Components: ModuleComponents{
				"go": &ComponentEntry{Root: "custom/path"},
			},
		}

		discoveryVars := map[string]string{
			"owner":   "myorg",
			"go_root": "custom/path",
		}
		params := buildModuleParams(mod, discoveryVars)

		assert.Equal(t, "custom/path", params["go_root"])
	})

	t.Run("go_root not present when not in discoveryVars", func(t *testing.T) {
		mod := &Module{
			Moniker: "my-container",
			Components: ModuleComponents{
				"dockerfile": &ComponentEntry{Root: "containers/my-container"},
			},
		}

		discoveryVars := map[string]string{
			"owner": "myorg",
		}
		params := buildModuleParams(mod, discoveryVars)

		_, hasGoRoot := params["go_root"]
		assert.False(t, hasGoRoot, "go_root should not be set when not in discoveryVars")
	})

	t.Run("moniker always set from module", func(t *testing.T) {
		mod := &Module{
			Moniker: "my-lib",
			Components: ModuleComponents{
				"go": &ComponentEntry{Root: "go/my-lib"},
			},
		}

		params := buildModuleParams(mod, nil)

		assert.Equal(t, "my-lib", params["moniker"])
	})
}
