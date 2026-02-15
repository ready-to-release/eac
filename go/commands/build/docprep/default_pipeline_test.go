package docprep

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPipeline_PhaseCount(t *testing.T) {
	pipeline := DefaultPipeline()
	phases := pipeline.Phases()

	require.Len(t, phases, 14, "DefaultPipeline should have 14 phases")
}

func TestDefaultPipeline_PhaseOrder(t *testing.T) {
	pipeline := DefaultPipeline()
	phases := pipeline.Phases()

	expected := []string{
		"scan-asset-refs",
		"copy-sources",
		"build-file-index",
		"convert-attr-images",
		"link-translation",
		"execute-commands",
		"navigation-structure",
		"inline-content",
		"command-help",
		"navigation-macros",
		"diagram-processing",
		"image-constraints",
		"fix-broken-links",
		"cleanup-assets",
	}

	require.Len(t, phases, len(expected))

	for i, phase := range phases {
		assert.Equal(t, expected[i], phase.Name(), "Phase %d name mismatch", i)
	}
}

func TestDefaultPipeline_AllPhasesNonNil(t *testing.T) {
	pipeline := DefaultPipeline()
	for _, phase := range pipeline.Phases() {
		assert.NotNil(t, phase, "All phases should be non-nil")
		assert.NotEmpty(t, phase.Name(), "All phases should have names")
	}
}
