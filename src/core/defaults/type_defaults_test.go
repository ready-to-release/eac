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
			moniker:    "src-core",
			root:       "src/core",
			moduleType: "go-library",
			expected:   "specs/src-core/**",
		},
		{
			name:       "root substitution",
			pattern:    "{root}/tests",
			moniker:    "src-core",
			root:       "src/core",
			moduleType: "go-library",
			expected:   "src/core/tests",
		},
		{
			name:       "type substitution",
			pattern:    "types/{type}/config",
			moniker:    "src-core",
			root:       "src/core",
			moduleType: "go-library",
			expected:   "types/go-library/config",
		},
		{
			name:       "multiple variables",
			pattern:    "{root}/{moniker}/{type}.config",
			moniker:    "my-module",
			root:       "src/modules",
			moduleType: "custom-type",
			expected:   "src/modules/my-module/custom-type.config",
		},
		{
			name:       "no variables passthrough",
			pattern:    "**/*.go",
			moniker:    "src-core",
			root:       "src/core",
			moduleType: "go-library",
			expected:   "**/*.go",
		},
		{
			name:       "empty pattern",
			pattern:    "",
			moniker:    "src-core",
			root:       "src/core",
			moduleType: "go-library",
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
			result := SubstituteVariables(tt.pattern, tt.moniker, tt.root, tt.moduleType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSubstituteAll tests batch substitution
func TestSubstituteAll(t *testing.T) {
	t.Run("nil slice returns nil", func(t *testing.T) {
		result := SubstituteAll(nil, "moniker", "root", "type")
		assert.Nil(t, result)
	})

	t.Run("empty slice returns empty slice", func(t *testing.T) {
		result := SubstituteAll([]string{}, "moniker", "root", "type")
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("multiple patterns", func(t *testing.T) {
		patterns := []string{
			"specs/{moniker}/**",
			"{root}/tests",
			"**/*.go",
		}
		result := SubstituteAll(patterns, "src-core", "src/core", "go-library")

		assert.Len(t, result, 3)
		assert.Equal(t, "specs/src-core/**", result[0])
		assert.Equal(t, "src/core/tests", result[1])
		assert.Equal(t, "**/*.go", result[2])
	})

	t.Run("preserves order", func(t *testing.T) {
		patterns := []string{"a", "b", "c", "d"}
		result := SubstituteAll(patterns, "m", "r", "t")
		assert.Equal(t, patterns, result)
	})
}

// TestResolveDefaults_NoTypeDefaults tests generic fallback when no type defaults
func TestResolveDefaults_NoTypeDefaults(t *testing.T) {
	result := ResolveDefaults(
		nil, // no type defaults
		"test-module", "src/test", "unknown-type",
		nil, nil, nil, nil, // source, config, assets, tests
		"",  // changelog
		nil, // specs
		"", "", // test_impl, design
		nil, nil, // flags
	)

	// Generic defaults should be applied
	assert.Equal(t, "CHANGELOG.md", result.Changelog)
	assert.Equal(t, []string{"specs/test-module/**"}, result.Specs)
	assert.Equal(t, "src/test-module/tests", result.TestImpl)
	assert.Equal(t, "specs/test-module/.design", result.Design)
	assert.False(t, result.CatchAll)
	assert.False(t, result.OwnChildren)

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
			Specs:    []string{"specs/{moniker}/**"},
			TestImpl: "{root}/tests",
			Design:   "specs/{moniker}/.design",
		},
	}

	result := ResolveDefaults(
		typeDef,
		"my-lib", "src/lib", "go-library",
		nil, nil, nil, nil,
		"",
		nil,
		"", "",
		nil, nil,
	)

	// Type defaults with substitution
	assert.Equal(t, []string{"**/*.go"}, result.Source)
	assert.Equal(t, []string{"go.mod", "go.sum"}, result.Config)
	assert.Equal(t, []string{"README.md"}, result.Assets)
	assert.Equal(t, "HISTORY.md", result.Changelog)
	assert.Equal(t, []string{"specs/my-lib/**"}, result.Specs)
	assert.Equal(t, "src/lib/tests", result.TestImpl)
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
			Specs:    []string{"specs/{moniker}/**"},
			TestImpl: "{root}/tests",
		},
	}

	// Explicit values
	explicitSource := []string{"custom/*.go"}
	explicitConfig := []string{"custom.mod"}
	explicitSpecs := []string{"custom/specs/**"}

	result := ResolveDefaults(
		typeDef,
		"my-lib", "src/lib", "go-library",
		explicitSource, explicitConfig, nil, nil,
		"CUSTOM.md", // explicit changelog
		explicitSpecs,
		"custom/tests", "custom/design",
		nil, nil,
	)

	// Explicit values preserved
	assert.Equal(t, explicitSource, result.Source)
	assert.Equal(t, explicitConfig, result.Config)
	assert.Equal(t, "CUSTOM.md", result.Changelog)
	assert.Equal(t, explicitSpecs, result.Specs)
	assert.Equal(t, "custom/tests", result.TestImpl)
	assert.Equal(t, "custom/design", result.Design)
}

// TestResolveDefaults_EmptySlicePreserved tests that explicit empty slice is preserved
func TestResolveDefaults_EmptySlicePreserved(t *testing.T) {
	typeDef := &TypeDefaults{
		Repo: &RepoDefaults{
			Specs: []string{"specs/{moniker}/**"},
		},
	}

	// Explicit empty slice (not nil)
	emptySpecs := []string{}

	result := ResolveDefaults(
		typeDef,
		"my-lib", "src/lib", "go-library",
		nil, nil, nil, nil,
		"",
		emptySpecs, // explicit empty
		"", "",
		nil, nil,
	)

	// Empty slice should be preserved, not replaced with type default
	assert.NotNil(t, result.Specs)
	assert.Empty(t, result.Specs)
}

// TestResolveDefaults_FlagHandling tests boolean flag resolution
func TestResolveDefaults_FlagHandling(t *testing.T) {
	t.Run("nil flags use generic defaults", func(t *testing.T) {
		result := ResolveDefaults(
			nil,
			"test", "src", "type",
			nil, nil, nil, nil,
			"", nil, "", "",
			nil, nil, // nil flags
		)

		assert.False(t, result.CatchAll)
		assert.False(t, result.OwnChildren)
	})

	t.Run("type defaults for flags", func(t *testing.T) {
		catchAll := true
		ownChildren := true
		typeDef := &TypeDefaults{
			Flags: &FlagsDefaults{
				CatchAll:         &catchAll,
				OwnChildrenFiles: &ownChildren,
			},
		}

		result := ResolveDefaults(
			typeDef,
			"test", "src", "type",
			nil, nil, nil, nil,
			"", nil, "", "",
			nil, nil,
		)

		assert.True(t, result.CatchAll)
		assert.True(t, result.OwnChildren)
	})

	t.Run("explicit flags override type defaults", func(t *testing.T) {
		catchAll := true
		typeDef := &TypeDefaults{
			Flags: &FlagsDefaults{
				CatchAll: &catchAll,
			},
		}

		explicitCatchAll := false
		result := ResolveDefaults(
			typeDef,
			"test", "src", "type",
			nil, nil, nil, nil,
			"", nil, "", "",
			&explicitCatchAll, nil,
		)

		assert.False(t, result.CatchAll)
	})
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
		nil, nil, nil, nil,
		"",
		nil,
		"", "",
		nil, nil,
	)

	// Type default for source
	assert.Equal(t, []string{"**/*.custom"}, result.Source)

	// Generic defaults for others
	assert.Equal(t, "CHANGELOG.md", result.Changelog)
	assert.Equal(t, []string{"specs/my-module/**"}, result.Specs)
	assert.Equal(t, "src/my-module/tests", result.TestImpl)
	assert.Equal(t, "specs/my-module/.design", result.Design)
}

// TestResolveDefaults_TypeDefaultsWithNilFields tests graceful handling of nil fields
func TestResolveDefaults_TypeDefaultsWithNilFields(t *testing.T) {
	// Type with nil Files and Repo
	typeDef := &TypeDefaults{
		Files: nil,
		Repo:  nil,
		Flags: nil,
	}

	result := ResolveDefaults(
		typeDef,
		"my-module", "src/mod", "custom-type",
		nil, nil, nil, nil,
		"",
		nil,
		"", "",
		nil, nil,
	)

	// Should use generic defaults
	assert.Equal(t, "CHANGELOG.md", result.Changelog)
	assert.Equal(t, []string{"specs/my-module/**"}, result.Specs)
}
