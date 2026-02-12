// Package internal provides shared infrastructure for GET and SHOW commands
package internal

import (
	"fmt"
	"strings"
)

// FormatArtifactStatus returns a human-readable status string for an artifact.
func FormatArtifactStatus(artifact ResolvedArtifact) string {
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

// FormatArtifactSize returns a human-readable size string for an artifact.
func FormatArtifactSize(sizeBytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case sizeBytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(sizeBytes)/float64(GB))
	case sizeBytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(sizeBytes)/float64(MB))
	case sizeBytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(sizeBytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", sizeBytes)
	}
}

// isBookModule checks if a module type is a book module (uses mkdocs handler).
func isBookModule(moduleType string) bool {
	// Check by type name - container modules with mkdocs handler
	return moduleType == "container" || strings.Contains(moduleType, "mkdocs")
}
