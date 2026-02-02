package mkdocs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/books"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

// SiteRenderHandler handles HTML site generation from base-site output.
// This is the second stage of the MkDocs build pipeline:
// 1. Load base-site manifest to get staged content location
// 2. Run MkDocs in Docker to generate HTML
// 3. Write manifest for cache validation
type SiteRenderHandler struct {
	manifestStore *ManifestStore
}

// NewSiteRenderHandler creates a new site render handler.
func NewSiteRenderHandler(workspaceRoot string) *SiteRenderHandler {
	return &SiteRenderHandler{
		manifestStore: NewManifestStore(workspaceRoot),
	}
}

// Name returns the handler identifier.
func (h *SiteRenderHandler) Name() string { return "site-render-tool" }

// Requirements returns system dependencies (Docker required).
func (h *SiteRenderHandler) Requirements() []string { return []string{"docker"} }

// resolveBaseSiteComponent determines the base-site component this site-render depends on.
// Resolution order:
// 1. Component's depends_on field (explicit dependency)
// 2. Naming convention: "{name}-base" for component "{name}"
// 3. Default to "base-site"
func resolveBaseSiteComponent(module interfaces.ModuleContractPort, componentName string) string {
	concrete := adapters.UnwrapModule(module)
	if concrete != nil {
		if comp, ok := concrete.Components[componentName]; ok && comp != nil {
			// Check depends_on for base-site type dependency
			for _, dep := range comp.DependsOn {
				// Return first dependency (assumed to be base-site)
				// This works because site-render/pdf-render components
				// should only depend on their base-site component
				return dep
			}
		}
	}

	// Naming convention: "site" -> "site-base"
	if componentName != "" && componentName != "site" {
		return componentName + "-base"
	}

	// Default fallback
	return "base-site"
}

// ValidateModule checks if a module has valid base-site dependency.
func (h *SiteRenderHandler) ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error {
	// Resolve the base-site component this site-render depends on
	baseSiteComp := resolveBaseSiteComponent(module, component)

	// Check if base-site manifest exists
	manifest, err := h.manifestStore.Load(module.GetMoniker(), baseSiteComp)
	if err != nil {
		return fmt.Errorf("failed to load base-site manifest for %s: %w", baseSiteComp, err)
	}
	if manifest == nil {
		return fmt.Errorf("base-site '%s' not built for module %s (component: %s)", baseSiteComp, module.GetMoniker(), component)
	}
	return nil
}

// ListArtifacts returns artifact paths that would be produced.
func (h *SiteRenderHandler) ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string {
	return []string{"site/", ".manifest.json"}
}

// IsContainer returns true as site rendering requires Docker.
func (h *SiteRenderHandler) IsContainer() bool { return true }

// IsHostInstalled returns false as site rendering uses Docker.
func (h *SiteRenderHandler) IsHostInstalled() bool { return false }

// Build executes HTML site generation from base-site output.
func (h *SiteRenderHandler) Build(
	module interfaces.ModuleContractPort,
	workspaceRoot, outputDir string,
	logWriter io.Writer,
	opts BuildOptions,
) BuildResult {
	startTime := time.Now()

	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		logln(logWriter, "Error: invalid module type")
		return BuildResult{ExitCode: 1}
	}

	// Resolve base-site component this site-render depends on
	baseSiteComponent := resolveBaseSiteComponent(module, opts.Component)

	logln(logWriter, "\n=== Rendering site: %s/%s (base: %s) ===", concrete.Moniker, opts.Component, baseSiteComponent)

	// Load base-site manifest
	baseManifest, err := h.manifestStore.Load(concrete.Moniker, baseSiteComponent)
	if err != nil {
		logln(logWriter, "❌ Failed to load base-site manifest: %v", err)
		return BuildResult{ExitCode: 1}
	}
	if baseManifest == nil {
		logln(logWriter, "❌ Base-site not built - run preprocessing first")
		return BuildResult{ExitCode: 1}
	}

	// Compute input hash for cache check
	containerHash := getContainerImageHash(workspaceRoot, "site-render-tool")
	inputHash := computeSiteRenderInputHash(baseManifest.OutputHash, containerHash)

	// Check cache - skip if base-site unchanged
	if !opts.Force && !opts.Reproducible {
		manifest, err := h.manifestStore.Load(concrete.Moniker, "site")
		if err == nil && manifest != nil && manifest.InputHash == inputHash {
			siteDir := filepath.Join(outputDir, "site")
			if _, statErr := os.Stat(filepath.Join(siteDir, "index.html")); statErr == nil {
				logln(logWriter, "⏭️  Site rendering skipped (cached, hash: %s...)", inputHash[:8])
				return BuildResult{
					ExitCode:  ExitCodeSkipped,
					Cached:    true,
					Manifest:  manifest,
					Artifacts: manifest.Artifacts,
					Duration:  time.Since(startTime),
				}
			}
		}
	}

	// Resolve staging directory from base-site manifest
	// Use the book name from manifest metadata (staging is keyed by book, not component)
	bookName := baseManifest.Metadata["book"]
	if bookName == "" {
		bookName = baseSiteComponent // fallback
	}
	stagingDir := paths.BookStagingCachePath(workspaceRoot, concrete.Moniker, bookName)
	if _, err := os.Stat(stagingDir); os.IsNotExist(err) {
		logln(logWriter, "❌ Staging directory not found: %s", stagingDir)
		return BuildResult{ExitCode: 1}
	}

	// Generate mkdocs.yml config
	// Use container mount path /staging (where staging dir is mounted in Docker)
	configPath := filepath.Join(outputDir, "mkdocs.yml")
	configOpts := books.ConfigOptions{
		SiteName:     "Documentation",
		DocsDir:      "/staging", // Container mount path for staging directory
		OutputFormat: "site",
	}
	if err := books.WriteMkDocsConfig(workspaceRoot, configPath, configOpts); err != nil {
		logln(logWriter, "❌ Failed to generate mkdocs.yml: %v", err)
		return BuildResult{ExitCode: 1}
	}

	// Copy mkdocs macros script
	macrosSource := filepath.Join(workspaceRoot, "containers", "site-render-tool", "mkdocs_macros.py")
	macrosTarget := filepath.Join(outputDir, "main.py")
	if macrosData, err := os.ReadFile(macrosSource); err == nil {
		_ = os.WriteFile(macrosTarget, macrosData, 0o644)
	}

	// Create output directory
	siteDir := filepath.Join(outputDir, "site")
	if err := os.MkdirAll(siteDir, 0o755); err != nil {
		logln(logWriter, "❌ Failed to create output directory: %v", err)
		return BuildResult{ExitCode: 1}
	}

	// Run MkDocs via tool bridge
	bridge := tool.GlobalHandlerToolBridge()

	// Get weight for resource scaling (default 1)
	weight := opts.Weight
	if weight <= 0 {
		weight = 1
	}

	tc := &tool.ToolContext{
		WorkspaceRoot: workspaceRoot,
		StagingDir:    stagingDir,
		OutputDir:     outputDir,
		ConfigPath:    configPath,
		LogWriter:     logWriter,
		Weight:        weight,
	}

	exitCode, err := bridge.ExecuteTool(context.Background(), "site-render-tool", tc)
	if err != nil {
		logln(logWriter, "❌ Tool execution failed: %v", err)
		return BuildResult{ExitCode: 1}
	}
	if exitCode != 0 {
		logln(logWriter, "❌ MkDocs build failed with exit code %d", exitCode)
		return BuildResult{ExitCode: exitCode}
	}

	// Compute output hash
	outputHash, err := computeStagingHash(siteDir)
	if err != nil {
		logln(logWriter, "   ⚠️  Failed to compute output hash: %v", err)
		outputHash = ""
	}

	// Write manifest
	manifest := &ComponentManifest{
		Version:    ManifestVersion,
		Type:       ComponentTypeSiteRender,
		Module:     concrete.Moniker,
		Component:  "site",
		InputHash:  inputHash,
		OutputHash: outputHash,
		Timestamp:  time.Now().UTC(),
		Duration:   time.Since(startTime),
		Dependencies: map[string]string{
			"base-site": baseManifest.OutputHash,
		},
		Artifacts: []string{"site/"},
	}

	if err := h.manifestStore.Save(manifest); err != nil {
		logln(logWriter, "   ⚠️  Failed to save manifest: %v", err)
	}

	logln(logWriter, "✅ Site rendering complete")
	logln(logWriter, "   HTML Output: %s", siteDir)

	return BuildResult{
		ExitCode:  0,
		Cached:    false,
		Manifest:  manifest,
		Artifacts: manifest.Artifacts,
		Duration:  time.Since(startTime),
	}
}

// computeSiteRenderInputHash computes the cache key for site rendering.
func computeSiteRenderInputHash(baseSiteOutputHash, containerHash string) string {
	// Simple concatenation for now - could use sha256 for more security
	return fmt.Sprintf("%s:%s", baseSiteOutputHash[:min(16, len(baseSiteOutputHash))], containerHash[:min(8, len(containerHash))])
}

// getContainerImageHash returns a hash representing the container image state.
// In production, this would check the image digest. For now, use a simplified version.
func getContainerImageHash(workspaceRoot, containerName string) string {
	// Hash the Dockerfile content as a proxy for image version
	dockerfilePath := filepath.Join(workspaceRoot, "containers", containerName, "Dockerfile")
	data, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return "unknown"
	}
	// Return first 8 chars of content hash
	hash, _ := computeFileHash(data)
	if len(hash) > 8 {
		return hash[:8]
	}
	return hash
}

// computeFileHash computes SHA256 hash of data.
func computeFileHash(data []byte) (string, error) {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}

