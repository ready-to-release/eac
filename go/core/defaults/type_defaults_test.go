//go:build L0 && ov

package defaults

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSubstituteVariables tests variable substitution in patterns
func TestSubstituteVariables(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		moniker    string
		root       string
		moduleType string
		expected   string
	}{
		{
			name:       "moniker substitution",
			pattern:    "specs/{moniker}/**",
			moniker:    "core",
			root:       "go/eac/core",
			moduleType: "go",
			expected:   "specs/core/**",
		},
		{
			name:       "root substitution",
			pattern:    "{root}/tests",
			moniker:    "core",
			root:       "go/eac/core",
			moduleType: "go",
			expected:   "go/eac/core/tests",
		},
		{
			name:       "type substitution",
			pattern:    "types/{type}/config",
			moniker:    "core",
			root:       "go/eac/core",
			moduleType: "go",
			expected:   "types/go/config",
		},
		{
			name:       "multiple variables",
			pattern:    "{root}/{moniker}/{type}.config",
			moniker:    "my-module",
			root:       "go/eac/modules",
			moduleType: "custom-type",
			expected:   "go/eac/modules/my-module/custom-type.config",
		},
		{
			name:       "no variables passthrough",
			pattern:    "**/*.go",
			moniker:    "core",
			root:       "go/eac/core",
			moduleType: "go",
			expected:   "**/*.go",
		},
		{
			name:       "empty pattern",
			pattern:    "",
			moniker:    "core",
			root:       "go/eac/core",
			moduleType: "go",
			expected:   "",
		},
		{
			name:       "repeated variable",
			pattern:    "{moniker}/{moniker}",
			moniker:    "test",
			root:       "src",
			moduleType: "lib",
			expected:   "test/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubstituteVariables(tt.pattern, tt.moniker, tt.root, tt.moduleType, nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSubstituteVariables_WithPathVars tests path variable substitution
func TestSubstituteVariables_WithPathVars(t *testing.T) {
	pathVars := map[string]string{
		"specs_root": "specs",
	}

	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{
			name:     "specs_root substitution",
			pattern:  "{specs_root}/{moniker}/**",
			expected: "specs/test-mod/**",
		},
		{
			name:     "combined with standard vars",
			pattern:  "{specs_root}/{moniker}/{type}",
			expected: "specs/test-mod/go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubstituteVariables(tt.pattern, "test-mod", "src/test", "go", pathVars)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSubstituteAll tests batch substitution
func TestSubstituteAll(t *testing.T) {
	t.Run("nil slice returns nil", func(t *testing.T) {
		result := SubstituteAll(nil, "moniker", "root", "type", nil)
		assert.Nil(t, result)
	})

	t.Run("empty slice returns empty slice", func(t *testing.T) {
		result := SubstituteAll([]string{}, "moniker", "root", "type", nil)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("multiple patterns", func(t *testing.T) {
		patterns := []string{
			"specs/{moniker}/**",
			"{root}/tests",
			"**/*.go",
		}
		result := SubstituteAll(patterns, "core", "go/eac/core", "go", nil)

		assert.Len(t, result, 3)
		assert.Equal(t, "specs/core/**", result[0])
		assert.Equal(t, "go/eac/core/tests", result[1])
		assert.Equal(t, "**/*.go", result[2])
	})

	t.Run("preserves order", func(t *testing.T) {
		patterns := []string{"a", "b", "c", "d"}
		result := SubstituteAll(patterns, "m", "r", "t", nil)
		assert.Equal(t, patterns, result)
	})
}

// TestResolveDefaults_NoTypeDefaults tests generic fallback when no type defaults
func TestResolveDefaults_NoTypeDefaults(t *testing.T) {
	result := ResolveDefaults(
		nil, // no type defaults
		"test-module", "src/test", "unknown-type",
		nil,                // pathVars
		nil, nil, nil, nil, // source, config, assets, tests
		"",     // changelog
		"", "", // workflowCI, workflowRelease
		nil,    // specs
		"", "", // test_impl, design
	)

	// Generic defaults should be applied
	assert.Equal(t, "CHANGELOG.md", result.Changelog)
	assert.Equal(t, []string{"specs/test-module/**"}, result.Specs)
	assert.Empty(t, result.TestImpl) // No fallback - must come from config
	assert.Equal(t, "specs/test-module/.design", result.Design)

	// No generic defaults for source/config/assets/tests
	assert.Nil(t, result.Source)
	assert.Nil(t, result.Config)
	assert.Nil(t, result.Assets)
	assert.Nil(t, result.Tests)
}

// TestResolveDefaults_WithTypeDefaults tests type-specific defaults
func TestResolveDefaults_WithTypeDefaults(t *testing.T) {
	typeDef := &TypeDefaults{
		Files: &FilesDefaults{
			Source:    []string{"**/*.go"},
			Config:    []string{"go.mod", "go.sum"},
			Assets:    []string{"README.md"},
			Changelog: "HISTORY.md",
		},
		Repo: &RepoDefaults{
			Specs:  []string{"{specs_root}/{moniker}/**"},
			Design: "{specs_root}/{moniker}/.design",
		},
	}

	pathVars := map[string]string{
		"specs_root": "specs",
	}

	result := ResolveDefaults(
		typeDef,
		"my-lib", "src/lib", "go",
		pathVars,
		nil, nil, nil, nil,
		"",
		"", "",
		nil,
		"", "",
	)

	// Type defaults with substitution
	assert.Equal(t, []string{"**/*.go"}, result.Source)
	assert.Equal(t, []string{"go.mod", "go.sum"}, result.Config)
	assert.Equal(t, []string{"README.md"}, result.Assets)
	assert.Equal(t, "HISTORY.md", result.Changelog)
	assert.Equal(t, []string{"specs/my-lib/**"}, result.Specs)
	assert.Empty(t, result.TestImpl) // TestImpl requires explicit component config
	assert.Equal(t, "specs/my-lib/.design", result.Design)
}

// TestResolveDefaults_ExplicitOverridesType tests that explicit values take precedence
func TestResolveDefaults_ExplicitOverridesType(t *testing.T) {
	typeDef := &TypeDefaults{
		Files: &FilesDefaults{
			Source:    []string{"**/*.go"},
			Config:    []string{"go.mod"},
			Changelog: "HISTORY.md",
		},
		Repo: &RepoDefaults{
			Specs: []string{"{specs_root}/{moniker}/**"},
		},
	}

	pathVars := map[string]string{
		"specs_root": "specs",
	}

	// Explicit values
	explicitSource := []string{"custom/*.go"}
	explicitConfig := []string{"custom.mod"}
	explicitSpecs := []string{"custom/specs/**"}

	result := ResolveDefaults(
		typeDef,
		"my-lib", "src/lib", "go",
		pathVars,
		explicitSource, explicitConfig, nil, nil,
		"CUSTOM.md", // explicit changelog
		"", "",
		explicitSpecs,
		"custom/tests", "custom/design",
	)

	// Explicit values preserved
	assert.Equal(t, explicitSource, result.Source)
	assert.Equal(t, explicitConfig, result.Config)
	assert.Equal(t, "CUSTOM.md", result.Changelog)
	assert.Equal(t, explicitSpecs, result.Specs)
	assert.Equal(t, "custom/tests", result.TestImpl)
	assert.Equal(t, "custom/design", result.Design)
}

// TestResolveDefaults_EmptySliceUsesDefault tests that empty slice is treated as unset
// and gets filled with the default value (consistent with nil behavior)
func TestResolveDefaults_EmptySliceUsesDefault(t *testing.T) {
	typeDef := &TypeDefaults{
		Repo: &RepoDefaults{
			Specs: []string{"{specs_root}/{moniker}/**"},
		},
	}

	// Empty slice (not nil) should be treated as unset
	emptySpecs := []string{}

	result := ResolveDefaults(
		typeDef,
		"my-lib", "src/lib", "go",
		map[string]string{"specs_root": "specs"}, // pathVars
		nil, nil, nil, nil,
		"",
		"", "",
		emptySpecs, // explicit empty - treated as unset
		"", "",
	)

	// Empty slice should use type default (with substitution applied)
	assert.NotNil(t, result.Specs)
	assert.Equal(t, []string{"specs/my-lib/**"}, result.Specs)
}

// TestResolveDefaults_PartialTypeDefaults tests when type has only some defaults
func TestResolveDefaults_PartialTypeDefaults(t *testing.T) {
	// Type only defines source, nothing else
	typeDef := &TypeDefaults{
		Files: &FilesDefaults{
			Source: []string{"**/*.custom"},
		},
	}

	result := ResolveDefaults(
		typeDef,
		"my-module", "src/mod", "custom-type",
		nil, // pathVars
		nil, nil, nil, nil,
		"",
		"", "",
		nil,
		"", "",
	)

	// Type default for source
	assert.Equal(t, []string{"**/*.custom"}, result.Source)

	// Generic defaults for others
	assert.Equal(t, "CHANGELOG.md", result.Changelog)
	assert.Equal(t, []string{"specs/my-module/**"}, result.Specs)
	assert.Empty(t, result.TestImpl) // No fallback - must come from config
	assert.Equal(t, "specs/my-module/.design", result.Design)
}

// TestResolveDefaults_TypeDefaultsWithNilFields tests graceful handling of nil fields
func TestResolveDefaults_TypeDefaultsWithNilFields(t *testing.T) {
	// Type with nil Files and Repo
	typeDef := &TypeDefaults{
		Files: nil,
		Repo:  nil,
	}

	result := ResolveDefaults(
		typeDef,
		"my-module", "src/mod", "custom-type",
		nil, // pathVars
		nil, nil, nil, nil,
		"",
		"", "",
		nil,
		"", "",
	)

	// Should use generic defaults
	assert.Equal(t, "CHANGELOG.md", result.Changelog)
	assert.Equal(t, []string{"specs/my-module/**"}, result.Specs)
}
