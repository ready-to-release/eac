// Package paths provides centralized path constants and utilities for the EAC repository.
package paths

import "path/filepath"

// ============================================================================
// Container/Build Helpers
// ============================================================================

// ContainersPath returns the path to a container directory.
func ContainersPath(repoRoot, container string) string {
	return filepath.Join(repoRoot, "containers", container)
}

// ContainerDockerfilePath returns the path to a container's Dockerfile.
func ContainerDockerfilePath(repoRoot, container string) string {
	return filepath.Join(repoRoot, "containers", container, "Dockerfile")
}

// MkDocsConfigPath returns the path to mkdocs.yml in an output directory.
func MkDocsConfigPath(outputDir string) string {
	return filepath.Join(outputDir, "mkdocs.yml")
}

// MkDocsSiteTemplatePath returns the path to the mkdocs-render-oci template.
func MkDocsSiteTemplatePath(repoRoot string) string {
	return filepath.Join(repoRoot, "containers", "mkdocs-render-oci", "mkdocs.yml")
}

// MkDocsPdfTemplatePath returns the path to the pdf-oci template.
func MkDocsPdfTemplatePath(repoRoot string) string {
	return filepath.Join(repoRoot, "containers", "pdf-oci", "mkdocs.yml")
}

// SiteOutputPath returns the path to the site output directory.
func SiteOutputPath(outputDir string) string {
	return filepath.Join(outputDir, "site")
}

// ServeOutputPath returns the path to the serve output directory (for live docs serving).
func ServeOutputPath(repoRoot string) string {
	return outSubPath(repoRoot, "serve")
}

// ============================================================================
// Staging/Asset Helpers (for book building)
// ============================================================================

// StagingAssetsPath returns the path to assets directory within staging.
func StagingAssetsPath(stagingDir string) string {
	return filepath.Join(stagingDir, "assets")
}

// DocsSourcePath returns the path to docs directory within a source root.
func DocsSourcePath(sourceRoot string) string {
	return filepath.Join(sourceRoot, DocsDir)
}
