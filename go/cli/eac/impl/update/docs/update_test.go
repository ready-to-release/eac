//go:build L0 && ov

package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags_Defaults(t *testing.T) {
	opts, err := parseFlags([]string{})
	assert.NoError(t, err)
	assert.False(t, opts.DryRun)
	assert.False(t, opts.Force)
	assert.False(t, opts.Verbose)
	assert.False(t, opts.PruneCache)
	// Default areas includes all
	assert.True(t, opts.Areas.Includes(AreaMermaid))
	assert.True(t, opts.Areas.Includes(AreaDrawio))
	assert.True(t, opts.Areas.Includes(AreaCommandRefs))
}

func TestParseFlags_DryRun(t *testing.T) {
	opts, err := parseFlags([]string{"--dry-run"})
	assert.NoError(t, err)
	assert.True(t, opts.DryRun)
}

func TestParseFlags_Force(t *testing.T) {
	opts, err := parseFlags([]string{"--force"})
	assert.NoError(t, err)
	assert.True(t, opts.Force)
}

func TestParseFlags_Verbose(t *testing.T) {
	opts, err := parseFlags([]string{"-v"})
	assert.NoError(t, err)
	assert.True(t, opts.Verbose)
}

func TestParseFlags_VerboseLong(t *testing.T) {
	opts, err := parseFlags([]string{"--verbose"})
	assert.NoError(t, err)
	assert.True(t, opts.Verbose)
}

func TestParseFlags_PruneCache(t *testing.T) {
	opts, err := parseFlags([]string{"--prune-cache"})
	assert.NoError(t, err)
	assert.True(t, opts.PruneCache)
}

func TestParseFlags_AreaMermaid(t *testing.T) {
	opts, err := parseFlags([]string{"--area", "mermaid"})
	assert.NoError(t, err)
	assert.True(t, opts.Areas.Includes(AreaMermaid))
	assert.False(t, opts.Areas.Includes(AreaDrawio))
}

func TestParseFlags_AreaEquals(t *testing.T) {
	opts, err := parseFlags([]string{"--area=drawio"})
	assert.NoError(t, err)
	assert.True(t, opts.Areas.Includes(AreaDrawio))
	assert.False(t, opts.Areas.Includes(AreaMermaid))
}

func TestParseFlags_InvalidArea(t *testing.T) {
	_, err := parseFlags([]string{"--area", "invalid"})
	assert.Error(t, err)
}

func TestParseFlags_AreaMissingValue(t *testing.T) {
	_, err := parseFlags([]string{"--area"})
	assert.Error(t, err)
}

func TestParseFlags_MultipleAreas(t *testing.T) {
	opts, err := parseFlags([]string{"--area", "mermaid", "--area", "drawio"})
	assert.NoError(t, err)
	assert.True(t, opts.Areas.Includes(AreaMermaid))
	assert.True(t, opts.Areas.Includes(AreaDrawio))
	assert.False(t, opts.Areas.Includes(AreaCommandRefs))
}

func TestParseFlags_Combined(t *testing.T) {
	opts, err := parseFlags([]string{"--dry-run", "--force", "-v", "--area", "mermaid"})
	assert.NoError(t, err)
	assert.True(t, opts.DryRun)
	assert.True(t, opts.Force)
	assert.True(t, opts.Verbose)
	assert.True(t, opts.Areas.Includes(AreaMermaid))
}
