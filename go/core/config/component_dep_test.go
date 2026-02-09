package config

import (
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseComponentDep_TwoParts(t *testing.T) {
	p, err := ParseComponentDep("contracts-core:go")
	require.NoError(t, err)
	assert.Equal(t, "contracts-core", p.Module)
	assert.Equal(t, "go", p.ComponentName)
	assert.Empty(t, p.ComponentType)
	assert.Empty(t, p.Tool)
	assert.Equal(t, "contracts-core:go", p.Raw)
}

func TestParseComponentDep_ThreeParts(t *testing.T) {
	p, err := ParseComponentDep("moduleB:go:go")
	require.NoError(t, err)
	assert.Equal(t, "moduleB", p.Module)
	assert.Equal(t, "go", p.ComponentName)
	assert.Equal(t, "go", p.ComponentType)
	assert.Empty(t, p.Tool)
}

func TestParseComponentDep_FourParts(t *testing.T) {
	p, err := ParseComponentDep("moduleB:go:go:gotest")
	require.NoError(t, err)
	assert.Equal(t, "moduleB", p.Module)
	assert.Equal(t, "go", p.ComponentName)
	assert.Equal(t, "go", p.ComponentType)
	assert.Equal(t, "gotest", p.Tool)
}

func TestParseComponentDep_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"one part", "moduleB"},
		{"five parts", "a:b:c:d:e"},
		{"empty module", ":go"},
		{"empty component", "moduleB:"},
		{"empty middle", "moduleB::go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseComponentDep(tt.input)
			assert.Error(t, err, "should error for %q", tt.input)
		})
	}
}

func TestParsedComponentDep_MatchesUnitID_TwoPart(t *testing.T) {
	dep := ParsedComponentDep{Module: "B", ComponentName: "go"}

	// Matches any B:go UoW regardless of type/tool
	assert.True(t, dep.MatchesUnitID(core.UnitID{Module: "B", ComponentName: "go", ComponentType: "go", Tool: "go-build"}))
	assert.True(t, dep.MatchesUnitID(core.UnitID{Module: "B", ComponentName: "go", ComponentType: "go", Tool: "gotest"}))

	// Wrong module or component
	assert.False(t, dep.MatchesUnitID(core.UnitID{Module: "C", ComponentName: "go", ComponentType: "go", Tool: "go-build"}))
	assert.False(t, dep.MatchesUnitID(core.UnitID{Module: "B", ComponentName: "docs", ComponentType: "docs", Tool: "mkdocs"}))
}

func TestParsedComponentDep_MatchesUnitID_ThreePart(t *testing.T) {
	dep := ParsedComponentDep{Module: "B", ComponentName: "go", ComponentType: "go"}

	assert.True(t, dep.MatchesUnitID(core.UnitID{Module: "B", ComponentName: "go", ComponentType: "go", Tool: "go-build"}))
	assert.True(t, dep.MatchesUnitID(core.UnitID{Module: "B", ComponentName: "go", ComponentType: "go", Tool: "gotest"}))

	// Wrong component type
	assert.False(t, dep.MatchesUnitID(core.UnitID{Module: "B", ComponentName: "go", ComponentType: "gherkin", Tool: "godog"}))
}

func TestParsedComponentDep_MatchesUnitID_FourPart(t *testing.T) {
	dep := ParsedComponentDep{Module: "B", ComponentName: "go", ComponentType: "go", Tool: "gotest"}

	assert.True(t, dep.MatchesUnitID(core.UnitID{Module: "B", ComponentName: "go", ComponentType: "go", Tool: "gotest"}))

	// Wrong tool
	assert.False(t, dep.MatchesUnitID(core.UnitID{Module: "B", ComponentName: "go", ComponentType: "go", Tool: "go-build"}))
}
