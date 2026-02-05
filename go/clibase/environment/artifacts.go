package environment

import (
	"github.com/ready-to-release/eac/go/core/environments"
)

// ArtifactsMode is an alias for environments.ArtifactsMode.
// Re-exported for backwards compatibility.
type ArtifactsMode = environments.ArtifactsMode

// Re-export constants for backwards compatibility.
const (
	ArtifactsModeAll     = environments.ArtifactsModeAll
	ArtifactsModeReduced = environments.ArtifactsModeReduced
)

// DefaultReducedPageLimit is the default page limit for PDF generation in reduced mode.
const DefaultReducedPageLimit = environments.DefaultReducedPageLimit

// ParseArtifactsMode parses a string into an ArtifactsMode.
// Re-exported for backwards compatibility.
var ParseArtifactsMode = environments.ParseArtifactsMode

// DefaultArtifactsMode returns the default artifacts mode for the environment.
// CI environments default to "all", everything else defaults to "reduced".
func (e *Env) DefaultArtifactsMode() ArtifactsMode {
	if e.IsCI {
		return ArtifactsModeAll
	}
	return ArtifactsModeReduced
}
