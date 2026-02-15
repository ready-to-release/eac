//go:build L0

package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArea(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Area
		wantErr bool
	}{
		{name: "mermaid", input: "mermaid", want: AreaMermaid},
		{name: "drawio", input: "drawio", want: AreaDrawio},
		{name: "command-refs", input: "command-refs", want: AreaCommandRefs},
		{name: "all", input: "all", want: AreaAll},
		{name: "invalid", input: "invalid", wantErr: true},
		{name: "empty", input: "", wantErr: true},
		{name: "uppercase", input: "MERMAID", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArea(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid area")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidAreas(t *testing.T) {
	areas := ValidAreas()
	assert.Len(t, areas, 4)
	assert.Contains(t, areas, AreaMermaid)
	assert.Contains(t, areas, AreaDrawio)
	assert.Contains(t, areas, AreaCommandRefs)
	assert.Contains(t, areas, AreaAll)
}

func TestValidAreaStrings(t *testing.T) {
	strs := ValidAreaStrings()
	assert.Len(t, strs, 4)
	assert.Contains(t, strs, "mermaid")
	assert.Contains(t, strs, "drawio")
	assert.Contains(t, strs, "command-refs")
	assert.Contains(t, strs, "all")
}

func TestNewAreaSet_Empty(t *testing.T) {
	// Empty list means all areas
	set := NewAreaSet([]Area{})
	assert.True(t, set.IsAll())
	assert.True(t, set.Includes(AreaMermaid))
	assert.True(t, set.Includes(AreaDrawio))
	assert.True(t, set.Includes(AreaCommandRefs))
	assert.Nil(t, set.Selected())
}

func TestNewAreaSet_All(t *testing.T) {
	// "all" in the list means all areas
	set := NewAreaSet([]Area{AreaAll})
	assert.True(t, set.IsAll())
	assert.True(t, set.Includes(AreaMermaid))
	assert.True(t, set.Includes(AreaDrawio))
	assert.True(t, set.Includes(AreaCommandRefs))
}

func TestNewAreaSet_Single(t *testing.T) {
	set := NewAreaSet([]Area{AreaMermaid})
	assert.False(t, set.IsAll())
	assert.True(t, set.Includes(AreaMermaid))
	assert.False(t, set.Includes(AreaDrawio))
	assert.False(t, set.Includes(AreaCommandRefs))

	selected := set.Selected()
	assert.Len(t, selected, 1)
	assert.Contains(t, selected, AreaMermaid)
}

func TestNewAreaSet_Multiple(t *testing.T) {
	set := NewAreaSet([]Area{AreaMermaid, AreaDrawio})
	assert.False(t, set.IsAll())
	assert.True(t, set.Includes(AreaMermaid))
	assert.True(t, set.Includes(AreaDrawio))
	assert.False(t, set.Includes(AreaCommandRefs))

	selected := set.Selected()
	assert.Len(t, selected, 2)
	assert.Contains(t, selected, AreaMermaid)
	assert.Contains(t, selected, AreaDrawio)
}

func TestNewAreaSet_AllTrumpsOthers(t *testing.T) {
	// If "all" is in the list with others, "all" takes precedence
	set := NewAreaSet([]Area{AreaMermaid, AreaAll, AreaDrawio})
	assert.True(t, set.IsAll())
	assert.True(t, set.Includes(AreaMermaid))
	assert.True(t, set.Includes(AreaDrawio))
	assert.True(t, set.Includes(AreaCommandRefs))
}

func TestAreaSet_Includes_UnknownArea(t *testing.T) {
	// Unknown areas should not be included in a specific set
	set := NewAreaSet([]Area{AreaMermaid})
	assert.False(t, set.Includes(Area("unknown")))

	// But should be included in "all" set
	allSet := NewAreaSet([]Area{})
	assert.True(t, allSet.Includes(Area("unknown")))
}
