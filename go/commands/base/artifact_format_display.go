package base

import (
	"fmt"

	"github.com/ready-to-release/eac/go/core/config"
)

// FormatArtifactStatus returns a human-readable status string for an artifact.
func FormatArtifactStatus(artifact config.ResolvedArtifact) string {
	if artifact.Exists {
		if artifact.MetadataOverride != "" {
			return fmt.Sprintf("✓ (Override: %s)", artifact.MetadataOverride)
		}
		return "✓"
	}

	if artifact.Error != "" {
		return fmt.Sprintf("✗ (%s)", artifact.Error)
	}

	return "✗"
}
