package cells

import (
	"fmt"
	"strings"
)

// ArtifactsCell displays artifacts for the selected unit.
type ArtifactsCell struct {
	artifacts []ArtifactInfo
}

// NewArtifactsCell creates a new artifacts cell.
func NewArtifactsCell() *ArtifactsCell {
	return &ArtifactsCell{}
}

// SetArtifacts sets the artifact information.
func (c *ArtifactsCell) SetArtifacts(artifacts []ArtifactInfo) {
	c.artifacts = artifacts
}

// Render returns the artifacts display.
func (c *ArtifactsCell) Render(width, height int) string {
	if len(c.artifacts) == 0 {
		return "artifacts: none"
	}

	var parts []string
	for _, a := range c.artifacts {
		parts = append(parts, fmt.Sprintf("%s (%s)", a.Name, formatSize(a.Size)))
	}

	return "artifacts: " + strings.Join(parts, ", ")
}

// ZoneID returns empty string (not an interactive zone).
func (c *ArtifactsCell) ZoneID() string {
	return ""
}

// formatSize formats a byte size in human-readable form.
func formatSize(bytes int64) string {
	const (
		KB = 1000
		MB = 1000 * KB
		GB = 1000 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
