// Package internal provides shared infrastructure for GET and SHOW commands
package internal

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
)

// isContainerModule checks if a module produces container images that are pushed to registry.
// For such modules, we don't download build artifacts - we pull from registry instead.
func isContainerModule(module *config.Module) bool {
	// Check if module has docker_build with push enabled
	dockerConfig := module.GetDockerBuildConfig()
	return dockerConfig != nil && dockerConfig.ShouldPush()
}

// expandBookArtifacts expands wildcard PDF patterns to specific book PDFs
// The first book in the module's books list is the default; others require --all flag.
func expandBookArtifacts(module *config.Module, artifacts []config.Artifact, cfg *config.EACConfig, buildAll bool) []config.Artifact {
	var expanded []config.Artifact

	for _, artifact := range artifacts {
		// Check if this is a PDF wildcard that needs expansion
		if artifact.Type == config.ArtifactTypeFile && artifact.Pattern == "*.pdf" {
			// Get books for this module
			books := cfg.GetBooksByModule(module.Moniker)

			// Expand to specific book PDFs
			for i, book := range books {
				// Skip non-default books (not first) unless --all flag is used
				isDefault := i == 0
				if !buildAll && !isDefault {
					continue
				}

				// Get the output mode (theme)
				output := book.GetOutput()

				// Parse theme from output (e.g., "pdf-dark" -> "dark")
				theme := "dark" // default
				if strings.HasPrefix(output, "pdf-") {
					theme = strings.TrimPrefix(output, "pdf-")
					// Handle "pdf-all" by creating artifacts for both themes
					if theme == "all" {
						for _, t := range []string{"dark", "light"} {
							pdfName := fmt.Sprintf("%s-%s.pdf", book.Name, t)
							expanded = append(expanded, config.Artifact{
								Type:    artifact.Type,
								Pattern: pdfName,
								ID:      fmt.Sprintf("%s-%s", book.Name, t),
								Verify:  artifact.Verify,
							})
						}
						continue
					}
				}

				// Create artifact for this book's PDF
				pdfName := fmt.Sprintf("%s-%s.pdf", book.Name, theme)
				expanded = append(expanded, config.Artifact{
					Type:    artifact.Type,
					Pattern: pdfName,
					ID:      book.Name,
					Verify:  artifact.Verify,
				})
			}
		} else {
			// Keep non-PDF artifacts as-is
			expanded = append(expanded, artifact)
		}
	}

	return expanded
}
