package config

import (
	"fmt"
	"strings"
)

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

// IsBookModule checks if a module type is a book module (uses mkdocs handler).
func IsBookModule(moduleType string) bool {
	// Check by type name - container modules with mkdocs handler
	return moduleType == "container" || strings.Contains(moduleType, "mkdocs")
}
