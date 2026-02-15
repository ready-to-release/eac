// Package internal provides shared infrastructure for SHOW commands
package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/config"
)

// FormatArtifactTable creates a formatted table of artifacts with status.
func FormatArtifactTable(
	artifacts []config.ResolvedArtifact,
	summary *config.ArtifactResolutionSummary,
	showAllPlatforms bool,
) string {
	if len(artifacts) == 0 {
		return "No artifacts defined for this module type"
	}

	// Build table
	tb := render.NewTableBuilder().
		WithHeaders("Type", "ID", "Pattern", "Resolved", "Path", "Exists", "Override")

	for i := range artifacts {
		art := &artifacts[i]
		exists := "✗"
		if art.Exists {
			exists = "✓"
		}

		override := "-"
		if art.MetadataOverride != "" {
			override = art.MetadataOverride
		}

		// Shorten path for readability
		shortPath := art.ResolvedPath
		if len(shortPath) > 40 {
			shortPath = "..." + shortPath[len(shortPath)-37:]
		}

		tb.AddRow(
			art.Type,
			art.ID,
			art.Pattern,
			art.ResolvedName,
			shortPath,
			exists,
			override,
		)
	}

	var output strings.Builder
	output.WriteString(tb.Build())
	output.WriteString("\n\n")

	// Add summary
	if summary != nil {
		output.WriteString("Summary:\n")
		output.WriteString(fmt.Sprintf("  Total artifacts: %d\n", summary.Total))
		output.WriteString(fmt.Sprintf("  Found: %d/%d (%.0f%%)\n",
			summary.Exists, summary.Total,
			float64(summary.Exists)/float64(summary.Total)*100))
		output.WriteString(fmt.Sprintf("  Missing: %d/%d\n", summary.Missing, summary.Total))
		if summary.Overrides > 0 {
			output.WriteString(fmt.Sprintf("  Metadata overrides: %d/%d (%.0f%%)\n",
				summary.Overrides, summary.Total,
				float64(summary.Overrides)/float64(summary.Total)*100))
		}
	}

	return output.String()
}

// FormatMetadataOverrides creates a formatted display of metadata overrides.
func FormatMetadataOverrides(
	metadata map[string]string,
	artifacts []config.ResolvedArtifact,
) string {
	if len(metadata) == 0 {
		return "No metadata overrides defined"
	}

	// Filter to only artifact-related metadata
	artifactMeta := make(map[string]string)
	for key, value := range metadata {
		// Check if this is an artifact override (matches pattern: {type}-{variant})
		if isArtifactMetadata(key) {
			artifactMeta[key] = value
		}
	}

	if len(artifactMeta) == 0 {
		return "No artifact metadata overrides defined"
	}

	var output strings.Builder
	output.WriteString("Metadata Overrides:\n")

	// Build table
	tb := render.NewTableBuilder().
		WithHeaders("Key", "Value", "Status")

	for key, value := range artifactMeta {
		// Check if this override is actually used
		status := "not used"
		for i := range artifacts {
			art := &artifacts[i]
			if art.MetadataOverride == key {
				status = "✓ applied"
				break
			}
		}

		tb.AddRow(key, value, status)
	}

	output.WriteString(tb.Build())
	return output.String()
}

// isArtifactMetadata checks if a metadata key is an artifact override.
func isArtifactMetadata(key string) bool {
	// Artifact metadata keys follow pattern: {type}-{variant}
	// Types: executable, file, directory, image, marker
	artifactTypes := []string{"executable", "file", "directory", "image", "marker"}

	for _, t := range artifactTypes {
		if strings.HasPrefix(key, t+"-") {
			return true
		}
	}

	return false
}

// FormatArtifactDetails formats detailed information about a single artifact.
func FormatArtifactDetails(artifact config.ResolvedArtifact) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Type: %s\n", artifact.Type))
	output.WriteString(fmt.Sprintf("Pattern: %s\n", artifact.Pattern))
	output.WriteString(fmt.Sprintf("Resolved: %s\n", artifact.ResolvedName))
	output.WriteString(fmt.Sprintf("Path: %s\n", artifact.ResolvedPath))

	if len(artifact.Platforms) > 0 {
		output.WriteString(fmt.Sprintf("Platforms: %s\n", strings.Join(artifact.Platforms, ", ")))
	}

	if artifact.VerifyMode != "" {
		output.WriteString(fmt.Sprintf("Verify Mode: %s\n", artifact.VerifyMode))
	}

	if artifact.MetadataOverride != "" {
		output.WriteString(fmt.Sprintf("Metadata Override: %s\n", artifact.MetadataOverride))
	}

	// Check file info if exists
	if artifact.Exists {
		info, err := os.Stat(artifact.ResolvedPath)
		if err == nil {
			output.WriteString(fmt.Sprintf("Size: %s\n", config.FormatArtifactSize(info.Size())))
			if artifact.IsDirectory {
				output.WriteString("Type: Directory\n")
			}
		}
	} else {
		output.WriteString("Status: Not found\n")
	}

	if artifact.Error != "" {
		output.WriteString(fmt.Sprintf("Error: %s\n", artifact.Error))
	}

	return output.String()
}

// FormatArtifactSummaryHeader creates a formatted header for artifact display.
func FormatArtifactSummaryHeader(
	moduleName, moduleType string,
	buildDir string,
	platform string,
	metadataCount int,
) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("# Artifacts: %s (%s)\n\n", moduleName, moduleType))
	output.WriteString(fmt.Sprintf("Build Directory: %s\n", buildDir))
	if platform != "" {
		output.WriteString(fmt.Sprintf("Platform: %s\n", platform))
	}
	if metadataCount > 0 {
		output.WriteString(fmt.Sprintf("Metadata Overrides: %d\n", metadataCount))
	}
	output.WriteString("\n")

	return output.String()
}
