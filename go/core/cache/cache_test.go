package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Level Tests
// =============================================================================

func TestAllLevels_ReturnsConcreteValues(t *testing.T) {
	levels := AllLevels()

	assert.Len(t, levels, 2)
	assert.Contains(t, levels, LevelLocal)
	assert.Contains(t, levels, LevelRemote)
	assert.NotContains(t, levels, LevelAll, "AllLevels should not include wildcard")
}

// =============================================================================
// Type Tests
// =============================================================================

func TestAllTypes_ReturnsConcreteValues(t *testing.T) {
	types := AllTypes()

	assert.Len(t, types, 6)
	assert.Contains(t, types, TypeRegistry)
	assert.Contains(t, types, TypeState)
	assert.Contains(t, types, TypeAsset)
	assert.Contains(t, types, TypeLayer)
	assert.Contains(t, types, TypeWork)
	assert.Contains(t, types, TypeCI)
	assert.NotContains(t, types, TypeAll, "AllTypes should not include wildcard")
}

// =============================================================================
// Spec.String() Tests
// =============================================================================

func TestSpec_String(t *testing.T) {
	tests := []struct {
		name     string
		spec     Spec
		expected string
	}{
		{
			name:     "all:all returns 'all'",
			spec:     Spec{Level: LevelAll, Type: TypeAll},
			expected: "all",
		},
		{
			name:     "local:all returns 'local'",
			spec:     Spec{Level: LevelLocal, Type: TypeAll},
			expected: "local",
		},
		{
			name:     "remote:all returns 'remote'",
			spec:     Spec{Level: LevelRemote, Type: TypeAll},
			expected: "remote",
		},
		{
			name:     "all:state returns 'state'",
			spec:     Spec{Level: LevelAll, Type: TypeState},
			expected: "state",
		},
		{
			name:     "all:asset returns 'asset'",
			spec:     Spec{Level: LevelAll, Type: TypeAsset},
			expected: "asset",
		},
		{
			name:     "local:state returns 'local:state'",
			spec:     Spec{Level: LevelLocal, Type: TypeState},
			expected: "local:state",
		},
		{
			name:     "remote:layer returns 'remote:layer'",
			spec:     Spec{Level: LevelRemote, Type: TypeLayer},
			expected: "remote:layer",
		},
		{
			name:     "local:registry returns 'local:registry'",
			spec:     Spec{Level: LevelLocal, Type: TypeRegistry},
			expected: "local:registry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.spec.String())
		})
	}
}

// =============================================================================
// Spec.Matches() Tests
// =============================================================================

func TestSpec_Matches(t *testing.T) {
	tests := []struct {
		name     string
		spec     Spec
		level    Level
		typ      Type
		expected bool
	}{
		// Wildcard matches
		{
			name:     "all:all matches local:state",
			spec:     Spec{Level: LevelAll, Type: TypeAll},
			level:    LevelLocal,
			typ:      TypeState,
			expected: true,
		},
		{
			name:     "all:all matches remote:layer",
			spec:     Spec{Level: LevelAll, Type: TypeAll},
			level:    LevelRemote,
			typ:      TypeLayer,
			expected: true,
		},
		{
			name:     "local:all matches local:state",
			spec:     Spec{Level: LevelLocal, Type: TypeAll},
			level:    LevelLocal,
			typ:      TypeState,
			expected: true,
		},
		{
			name:     "local:all matches local:asset",
			spec:     Spec{Level: LevelLocal, Type: TypeAll},
			level:    LevelLocal,
			typ:      TypeAsset,
			expected: true,
		},
		{
			name:     "local:all does NOT match remote:state",
			spec:     Spec{Level: LevelLocal, Type: TypeAll},
			level:    LevelRemote,
			typ:      TypeState,
			expected: false,
		},
		{
			name:     "all:state matches local:state",
			spec:     Spec{Level: LevelAll, Type: TypeState},
			level:    LevelLocal,
			typ:      TypeState,
			expected: true,
		},
		{
			name:     "all:state matches remote:state",
			spec:     Spec{Level: LevelAll, Type: TypeState},
			level:    LevelRemote,
			typ:      TypeState,
			expected: true,
		},
		{
			name:     "all:state does NOT match local:asset",
			spec:     Spec{Level: LevelAll, Type: TypeState},
			level:    LevelLocal,
			typ:      TypeAsset,
			expected: false,
		},
		// Exact matches
		{
			name:     "local:state matches local:state",
			spec:     Spec{Level: LevelLocal, Type: TypeState},
			level:    LevelLocal,
			typ:      TypeState,
			expected: true,
		},
		{
			name:     "local:state does NOT match remote:state",
			spec:     Spec{Level: LevelLocal, Type: TypeState},
			level:    LevelRemote,
			typ:      TypeState,
			expected: false,
		},
		{
			name:     "local:state does NOT match local:asset",
			spec:     Spec{Level: LevelLocal, Type: TypeState},
			level:    LevelLocal,
			typ:      TypeAsset,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.spec.Matches(tt.level, tt.typ))
		})
	}
}

// =============================================================================
// Spec.Validate() Tests
// =============================================================================

func TestSpec_Validate(t *testing.T) {
	tests := []struct {
		name      string
		spec      Spec
		expectErr bool
		errMsg    string
	}{
		{
			name:      "local:work is valid",
			spec:      Spec{Level: LevelLocal, Type: TypeWork},
			expectErr: false,
		},
		{
			name:      "remote:work is invalid",
			spec:      Spec{Level: LevelRemote, Type: TypeWork},
			expectErr: true,
			errMsg:    "work cache type is local-only",
		},
		{
			name:      "all:work allows wildcard level",
			spec:      Spec{Level: LevelAll, Type: TypeWork},
			expectErr: false,
		},
		{
			name:      "local:state is valid",
			spec:      Spec{Level: LevelLocal, Type: TypeState},
			expectErr: false,
		},
		{
			name:      "remote:layer is valid",
			spec:      Spec{Level: LevelRemote, Type: TypeLayer},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// ParseSpec() Tests
// =============================================================================

func TestParseSpec_ValidSpecs(t *testing.T) {
	tests := []struct {
		input    string
		expected Spec
	}{
		// Wildcards
		{"all", Spec{Level: LevelAll, Type: TypeAll}},
		{"ALL", Spec{Level: LevelAll, Type: TypeAll}},     // Case insensitive
		{"  all  ", Spec{Level: LevelAll, Type: TypeAll}}, // Trimmed

		// Level-only (implies all types)
		{"local", Spec{Level: LevelLocal, Type: TypeAll}},
		{"remote", Spec{Level: LevelRemote, Type: TypeAll}},
		{"LOCAL", Spec{Level: LevelLocal, Type: TypeAll}},

		// Type-only (implies all levels)
		{"state", Spec{Level: LevelAll, Type: TypeState}},
		{"asset", Spec{Level: LevelAll, Type: TypeAsset}},
		{"registry", Spec{Level: LevelAll, Type: TypeRegistry}},
		{"layer", Spec{Level: LevelAll, Type: TypeLayer}},
		{"work", Spec{Level: LevelAll, Type: TypeWork}},
		{"STATE", Spec{Level: LevelAll, Type: TypeState}},

		// Type-only CI
		{"ci", Spec{Level: LevelAll, Type: TypeCI}},

		// Explicit level:type
		{"local:state", Spec{Level: LevelLocal, Type: TypeState}},
		{"local:asset", Spec{Level: LevelLocal, Type: TypeAsset}},
		{"local:registry", Spec{Level: LevelLocal, Type: TypeRegistry}},
		{"local:layer", Spec{Level: LevelLocal, Type: TypeLayer}},
		{"local:work", Spec{Level: LevelLocal, Type: TypeWork}},
		{"remote:state", Spec{Level: LevelRemote, Type: TypeState}},
		{"remote:layer", Spec{Level: LevelRemote, Type: TypeLayer}},
		{"remote:registry", Spec{Level: LevelRemote, Type: TypeRegistry}},
		{"remote:ci", Spec{Level: LevelRemote, Type: TypeCI}},
		{"LOCAL:STATE", Spec{Level: LevelLocal, Type: TypeState}},
		{"  local:state  ", Spec{Level: LevelLocal, Type: TypeState}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			spec, err := ParseSpec(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, spec)
		})
	}
}

func TestParseSpec_InvalidSpecs(t *testing.T) {
	tests := []struct {
		input  string
		errMsg string
	}{
		{"invalid", "unknown cache spec"},
		{"foo:bar", "unknown cache level"},
		{"local:invalid", "unknown cache type"},
		{"invalid:state", "unknown cache level"},
		{"a:b:c", "invalid cache spec format"},
		{"remote:work", "work cache type is local-only"}, // Validation error
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseSpec(tt.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// =============================================================================
// ParseSpecs() Tests
// =============================================================================

func TestParseSpecs_ValidSpecs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Spec
	}{
		{
			name:     "single spec",
			input:    "local:state",
			expected: []Spec{{Level: LevelLocal, Type: TypeState}},
		},
		{
			name:  "multiple specs comma-separated",
			input: "local:state,local:asset",
			expected: []Spec{
				{Level: LevelLocal, Type: TypeState},
				{Level: LevelLocal, Type: TypeAsset},
			},
		},
		{
			name:  "multiple specs with spaces",
			input: "local:state, local:asset , remote:layer",
			expected: []Spec{
				{Level: LevelLocal, Type: TypeState},
				{Level: LevelLocal, Type: TypeAsset},
				{Level: LevelRemote, Type: TypeLayer},
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "only commas and spaces",
			input:    " , , ",
			expected: nil,
		},
		{
			name:  "mixed shorthand and explicit",
			input: "state,local:registry",
			expected: []Spec{
				{Level: LevelAll, Type: TypeState},
				{Level: LevelLocal, Type: TypeRegistry},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specs, err := ParseSpecs(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, specs)
		})
	}
}

func TestParseSpecs_InvalidSpecs(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		errMsg string
	}{
		{
			name:   "one invalid in list",
			input:  "local:state,invalid,local:asset",
			errMsg: "unknown cache spec",
		},
		{
			name:   "validation error in list",
			input:  "local:state,remote:work",
			errMsg: "work cache type is local-only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSpecs(tt.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// =============================================================================
// isLevel() and isType() Tests (via ParseSpec behavior)
// =============================================================================

func TestParseSpec_AllKnownLevels(t *testing.T) {
	validLevels := []string{"local", "remote", "all"}
	for _, level := range validLevels {
		t.Run(level, func(t *testing.T) {
			spec, err := ParseSpec(level)
			require.NoError(t, err)
			assert.Equal(t, TypeAll, spec.Type)
		})
	}
}

func TestParseSpec_AllKnownTypes(t *testing.T) {
	validTypes := []string{"state", "asset", "registry", "layer", "work", "ci"}
	for _, typ := range validTypes {
		t.Run(typ, func(t *testing.T) {
			spec, err := ParseSpec(typ)
			require.NoError(t, err)
			assert.Equal(t, LevelAll, spec.Level)
		})
	}
}
