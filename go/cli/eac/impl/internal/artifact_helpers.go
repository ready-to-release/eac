// Package internal provides shared infrastructure for GET and SHOW commands
package internal

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
)

// ResolvedArtifact represents an artifact with metadata overrides applied and existence checked.
type ResolvedArtifact struct {
	// From artifact definition
	Type       string   `json:"type" yaml:"type"`
	ID         string   `json:"id" yaml:"id"`
	Pattern    string   `json:"pattern" yaml:"pattern"`
	Platforms  []string `json:"platforms,omitempty" yaml:"platforms,omitempty"`
	VerifyMode string   `json:"verify_mode,omitempty" yaml:"verify_mode,omitempty"`

	// Resolved information
	ResolvedName     string `json:"resolved_name" yaml:"resolved_name"`
	ResolvedPath     string `json:"resolved_path" yaml:"resolved_path"`
	MetadataOverride string `json:"metadata_override,omitempty" yaml:"metadata_override,omitempty"`
	Exists           bool   `json:"exists" yaml:"exists"`
	IsDirectory      bool   `json:"is_directory,omitempty" yaml:"is_directory,omitempty"`
	Error            string `json:"error,omitempty" yaml:"error,omitempty"`
}

// ArtifactResolutionSummary provides summary statistics for artifact resolution.
type ArtifactResolutionSummary struct {
	Total     int `json:"total" yaml:"total"`
	Exists    int `json:"exists" yaml:"exists"`
	Missing   int `json:"missing" yaml:"missing"`
	Overrides int `json:"overrides" yaml:"overrides"`
}

// ResolveArtifactsForModule resolves all artifacts for a module with metadata applied.
func ResolveArtifactsForModule(
	module *config.Module,
	buildDir string,
	targetOS, targetArch string,
) ([]ResolvedArtifact, *ArtifactResolutionSummary, error) {
	return ResolveArtifactsForModuleWithConfig(module, buildDir, targetOS, targetArch, nil)
}

// ResolveArtifactsForModuleWithConfig resolves all artifacts for a module with optional books config.
// Artifact resolution uses cfg.GetBuildArtifacts() which properly merges module-level and
// type-level artifacts (module-level takes priority).
func ResolveArtifactsForModuleWithConfig(
	module *config.Module,
	buildDir string,
	targetOS, targetArch string,
	cfg *config.EACConfig,
) ([]ResolvedArtifact, *ArtifactResolutionSummary, error) {
	if module == nil {
		return nil, nil, fmt.Errorf("module cannot be nil")
	}

	// Get merged artifacts from config (module-level takes priority over type-level)
	// Pass buildAll=true to get all artifacts including UPX variants for resolution
	var artifacts []config.Artifact
	if cfg != nil {
		artifacts = cfg.GetBuildArtifacts(module.Moniker, true)
	}

	if len(artifacts) == 0 {
		// No artifacts defined - not a buildable module
		return []ResolvedArtifact{}, &ArtifactResolutionSummary{}, nil
	}

	// Create artifact resolver
	resolver := config.NewArtifactResolverFull(
		module.Moniker,
		buildDir,
		targetOS,
		targetArch,
		module.Metadata,
	)

	// If no artifacts defined, module type doesn't produce artifacts.
	// The per-module manifest is the contract for validation (not markers).

	// Verify all artifacts
	results := resolver.VerifyArtifacts(artifacts)

	// Convert to ResolvedArtifact format
	resolved := make([]ResolvedArtifact, len(results))
	summary := &ArtifactResolutionSummary{}

	for i, result := range results {
		artifact := ResolvedArtifact{
			Type:         result.Artifact.Type,
			ID:           deriveArtifactID(result.Artifact, targetOS, targetArch),
			Pattern:      result.Artifact.Pattern,
			Platforms:    result.Artifact.Platforms,
			VerifyMode:   result.Artifact.Verify,
			ResolvedName: result.Pattern,
			ResolvedPath: result.Path,
			Exists:       result.Exists,
			IsDirectory:  result.IsDirectory,
		}

		if result.Error != nil {
			artifact.Error = result.Error.Error()
		}

		// Check if this artifact used a metadata override
		if module.Metadata != nil {
			metadataKey := fmt.Sprintf("%s-%s", result.Artifact.Type, artifact.ID)
			if _, ok := module.Metadata[metadataKey]; ok {
				artifact.MetadataOverride = metadataKey
				summary.Overrides++
			}
		}

		resolved[i] = artifact
		summary.Total++
		if artifact.Exists {
			summary.Exists++
		} else {
			summary.Missing++
		}
	}

	return resolved, summary, nil
}

// deriveArtifactID derives an ID for an artifact based on its type and context.
func deriveArtifactID(artifact config.Artifact, targetOS, targetArch string) string {
	// If artifact has explicit ID, use it
	if artifact.ID != "" {
		return artifact.ID
	}

	// For executables, use {os}-{arch} format with compression suffix if applicable
	if artifact.Type == config.ArtifactTypeExecutable {
		baseID := fmt.Sprintf("%s-%s", targetOS, targetArch)

		// Append compression suffix for compressed variants
		switch artifact.Compression {
		case config.CompressionUPX:
			return baseID + "-upx"
		case config.CompressionStrip:
			return baseID + "-strip"
		}

		return baseID
	}

	// For other types, extract from pattern
	// Remove common variables to get a reasonable ID
	id := artifact.Pattern
	id = filepath.Base(id)

	// Remove extension for files
	if artifact.Type == config.ArtifactTypeFile {
		ext := filepath.Ext(id)
		if ext != "" {
			id = id[:len(id)-len(ext)]
		}
	}

	return id
}

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

// ModuleValidationResult represents validation results for a single module.
type ModuleValidationResult struct {
	Moniker           string                     `json:"moniker" yaml:"moniker"`
	Type              string                     `json:"type" yaml:"type"`
	IsDependency      bool                       `json:"is_dependency" yaml:"is_dependency"`
	HasBuildArtifacts bool                       `json:"has_build_artifacts" yaml:"has_build_artifacts"`
	Artifacts         []ResolvedArtifact         `json:"artifacts" yaml:"artifacts"`
	Summary           *ArtifactResolutionSummary `json:"summary" yaml:"summary"`
	Error             string                     `json:"error,omitempty" yaml:"error,omitempty"`
}

// ValidationResults represents validation results for a module and all dependencies.
type ValidationResults struct {
	TargetModule string                   `json:"target_module" yaml:"target_module"`
	Modules      []ModuleValidationResult `json:"modules" yaml:"modules"`
	TotalModules int                      `json:"total_modules" yaml:"total_modules"`
	PassedCount  int                      `json:"passed_count" yaml:"passed_count"`
	FailedCount  int                      `json:"failed_count" yaml:"failed_count"`
	Passed       bool                     `json:"passed" yaml:"passed"`
}

// ValidateArtifactsTargetOnly validates artifacts for a single module without checking dependencies.
// Use this in release workflows where CI already validated and dependencies are not downloaded.
func ValidateArtifactsTargetOnly(
	targetModule string,
	cfg *config.EACConfig,
	registry *modules.Registry,
	targetOS, targetArch string,
	workspaceRoot string,
) (*ValidationResults, error) {
	// Get requested artifacts from UoW manifests via coreoutput.Reader
	// Note: UoW manifests track actual artifacts produced, not "requested" artifacts
	// The validation logic handles empty requestedArtifacts as "validate all"
	var requestedArtifacts []string
	reader := coreoutput.NewReader(workspaceRoot)
	if moduleView, err := reader.GetModule(core.ActionBuild, targetModule); err == nil && moduleView != nil {
		requestedArtifacts = extractRequestedArtifactsFromModuleView(moduleView)
	}

	// Validate only the target module
	results := &ValidationResults{
		TargetModule: targetModule,
		Modules:      []ModuleValidationResult{},
	}

	modResult := validateSingleModule(targetModule, targetModule, cfg, registry, targetOS, targetArch, workspaceRoot, requestedArtifacts)
	results.Modules = append(results.Modules, modResult)

	if modResult.Error != "" || (modResult.Summary != nil && modResult.Summary.Missing > 0) {
		results.FailedCount++
	} else {
		results.PassedCount++
	}

	results.TotalModules = 1
	results.Passed = results.FailedCount == 0

	return results, nil
}

// ValidateArtifactsWithDependencies validates artifacts for a module and all transitive dependencies.
func ValidateArtifactsWithDependencies(
	targetModule string,
	cfg *config.EACConfig,
	registry *modules.Registry,
	targetOS, targetArch string,
	workspaceRoot string,
) (*ValidationResults, error) {
	// Get all transitive dependencies
	allModules := make(map[string]bool)
	allModules[targetModule] = true

	if err := addDependenciesRecursive(targetModule, registry, allModules); err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Load module views from UoW manifests to get artifacts and platform info
	reader := coreoutput.NewReader(workspaceRoot)
	moduleViews := make(map[string]*coreoutput.ModuleView)
	for moniker := range allModules {
		if moduleView, err := reader.GetModule(core.ActionBuild, moniker); err == nil && moduleView != nil {
			moduleViews[moniker] = moduleView
		}
	}

	// Validate each module
	results := &ValidationResults{
		TargetModule: targetModule,
		Modules:      []ModuleValidationResult{},
	}

	for moniker := range allModules {
		// Get requested artifacts from module view
		var requestedArtifacts []string
		if moduleView, ok := moduleViews[moniker]; ok {
			requestedArtifacts = extractRequestedArtifactsFromModuleView(moduleView)
		}

		// For dependencies, use the platform from manifest (where they were built)
		// This supports cross-platform CI (e.g., Linux build with --all, Windows test)
		resolveOS, resolveArch := targetOS, targetArch
		if moniker != targetModule {
			if moduleView, ok := moduleViews[moniker]; ok {
				if platforms := extractPlatformsFromModuleView(moduleView); len(platforms) > 0 {
					resolveOS = platforms[0].OS
					resolveArch = platforms[0].Arch
				}
			}
		}

		modResult := validateSingleModule(moniker, targetModule, cfg, registry, resolveOS, resolveArch, workspaceRoot, requestedArtifacts)
		results.Modules = append(results.Modules, modResult)

		if modResult.Error != "" || modResult.Summary.Missing > 0 {
			results.FailedCount++
		} else {
			results.PassedCount++
		}
	}

	// Sort results: target module first, then dependencies alphabetically
	sort.Slice(results.Modules, func(i, j int) bool {
		if results.Modules[i].Moniker == targetModule {
			return true
		}
		if results.Modules[j].Moniker == targetModule {
			return false
		}
		return results.Modules[i].Moniker < results.Modules[j].Moniker
	})

	results.TotalModules = len(results.Modules)
	results.Passed = results.FailedCount == 0

	return results, nil
}

// validateSingleModule validates artifacts for a single module.
func validateSingleModule(
	moniker string,
	targetModule string,
	cfg *config.EACConfig,
	registry *modules.Registry,
	targetOS, targetArch string,
	workspaceRoot string,
	requestedArtifacts []string,
) ModuleValidationResult {
	result := ModuleValidationResult{
		Moniker:      moniker,
		IsDependency: moniker != targetModule,
	}

	// Get module contract
	moduleContract, exists := registry.Get(moniker)
	if !exists {
		result.Error = "module contract not found"
		return result
	}
	result.Type = moduleContract.GetComponentTypesDisplay()

	// Get module from config
	module, ok := cfg.Repository.GetModule(moniker)
	if !ok {
		result.Error = "module not found in config"
		return result
	}

	// For container-type DEPENDENCIES, skip local artifact validation.
	// Container deps are pulled from registry at runtime, not downloaded as build artifacts.
	// We only download the build manifest (to know the image tag), not the full build output.
	if result.IsDependency && isContainerModule(module) {
		result.HasBuildArtifacts = false
		result.Summary = &ArtifactResolutionSummary{}
		return result
	}

	// Check if module has build artifacts using merged config
	// cfg.GetBuildArtifacts handles module-level vs type-level priority and book artifacts
	allArtifacts := cfg.GetBuildArtifacts(moniker, true)
	if len(allArtifacts) == 0 {
		result.HasBuildArtifacts = false
		result.Summary = &ArtifactResolutionSummary{}
		return result
	}

	result.HasBuildArtifacts = true

	// Build directory (make absolute path from relative path)
	buildDirRel := cfg.Repository.BuildOutputPath(moniker)
	buildDir := filepath.Join(workspaceRoot, buildDirRel)

	// Resolve artifacts
	artifacts, summary, err := ResolveArtifactsForModuleWithConfig(
		module, buildDir, targetOS, targetArch, cfg,
	)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Check if this module has docker_build with push=true
	// If so, image artifacts are pushed to registry and may not exist locally
	dockerConfig := module.GetDockerBuildConfig()
	imagesPushedToRegistry := dockerConfig != nil && dockerConfig.ShouldPush()

	// Filter artifacts to only requested ones if requestedArtifacts is specified
	if len(requestedArtifacts) > 0 {
		var filteredArtifacts []ResolvedArtifact
		filteredSummary := &ArtifactResolutionSummary{}

		for _, artifact := range artifacts {
			// Check if this artifact ID is in the requested list
			isRequested := false
			for _, reqID := range requestedArtifacts {
				if artifact.ID == reqID {
					isRequested = true
					break
				}
			}

			if isRequested {
				filteredArtifacts = append(filteredArtifacts, artifact)
				filteredSummary.Total++

				// For image artifacts with push=true, trust that buildx push succeeded
				// (buildx would have failed if push failed). The image may not exist locally.
				if artifact.Type == "image" && imagesPushedToRegistry {
					// Treat pushed images as existing
					artifact.Exists = true
					filteredSummary.Exists++
				} else if artifact.Exists {
					filteredSummary.Exists++
				} else {
					filteredSummary.Missing++
				}
				if artifact.MetadataOverride != "" {
					filteredSummary.Overrides++
				}
			}
		}

		result.Artifacts = filteredArtifacts
		result.Summary = filteredSummary
	} else {
		// No filtering - validate all artifacts (fallback behavior)
		// Still need to handle pushed images
		if imagesPushedToRegistry {
			for i := range artifacts {
				if artifacts[i].Type == "image" {
					artifacts[i].Exists = true
					summary.Missing--
					summary.Exists++
				}
			}
		}
		result.Artifacts = artifacts
		result.Summary = summary
	}

	return result
}

// addDependenciesRecursive recursively adds all dependencies of a module.
func addDependenciesRecursive(moniker string, registry *modules.Registry, result map[string]bool) error {
	module, exists := registry.Get(moniker)
	if !exists {
		return fmt.Errorf("module '%s' not found in registry", moniker)
	}

	var missingDeps []string

	for _, dep := range module.DependsOn {
		// Check if dependency exists in registry
		if _, exists := registry.Get(dep); !exists {
			missingDeps = append(missingDeps, fmt.Sprintf("%s->%s", moniker, dep))
			continue
		}

		if !result[dep] {
			result[dep] = true
			if err := addDependenciesRecursive(dep, registry, result); err != nil {
				return err
			}
		}
	}

	if len(missingDeps) > 0 {
		return fmt.Errorf("missing dependency contracts: %v", missingDeps)
	}

	return nil
}

// DetermineRequestedArtifacts returns the list of artifact IDs that should be built
// based on the module and whether --all mode is requested.
//
// This delegates to config.EACConfig.GetBuildArtifactIDs which encapsulates all
// artifact merging and filtering logic (module vs type-level, UPX filtering, etc.)
func DetermineRequestedArtifacts(
	module *config.Module,
	buildAll bool,
	cfg *config.EACConfig,
) []string {
	if module == nil || cfg == nil {
		return []string{}
	}

	// Delegate to core config - single source of truth for artifact resolution
	return cfg.GetBuildArtifactIDs(module.Moniker, buildAll)
}

// extractRequestedArtifactsFromModuleView extracts artifact IDs from a ModuleView.
// Note: UoW manifests track actual artifacts produced, not "requested" artifacts.
// This returns the IDs of all artifacts that were actually produced.
// Callers should treat empty result as "validate all artifacts".
func extractRequestedArtifactsFromModuleView(view *coreoutput.ModuleView) []string {
	if view == nil {
		return nil
	}

	var ids []string
	for _, comp := range view.Components {
		for _, uow := range comp.UoWs {
			for _, art := range uow.Artifacts {
				if art.ID != "" {
					ids = append(ids, art.ID)
				}
			}
		}
	}
	return ids
}

// PlatformInfo describes a platform that was built.
type PlatformInfo struct {
	OS   string `json:"os"`   // Operating system (windows, linux, darwin)
	Arch string `json:"arch"` // Architecture (amd64, arm64)
}

// ArtifactInfo describes a single built artifact.
type ArtifactInfo struct {
	Type     string   `json:"type"`               // Artifact type (executable, file, directory, image)
	ID       string   `json:"id"`                 // Artifact identifier
	Name     string   `json:"name"`               // Resolved artifact name or image reference
	Path     string   `json:"path"`               // Relative path from build root, or image reference for type=image
	Platform string   `json:"platform,omitempty"` // Platform (e.g., "windows-amd64", "linux/amd64") if applicable
	Size     int64    `json:"size,omitempty"`     // File size in bytes (for file-based artifacts)
	SHA256   string   `json:"sha256,omitempty"`   // SHA-256 hash of artifact content (for file-based artifacts)
	Digest   string   `json:"digest,omitempty"`   // Image digest (for type=image, e.g., "sha256:abc123...")
	Tags     []string `json:"tags,omitempty"`     // Image tags (for type=image)
	Registry string   `json:"registry,omitempty"` // Container registry (for type=image, e.g., "ghcr.io")
}

// extractPlatformsFromModuleView extracts platform info from artifact paths in a ModuleView.
func extractPlatformsFromModuleView(view *coreoutput.ModuleView) []PlatformInfo {
	if view == nil {
		return nil
	}

	seen := make(map[string]bool)
	var platforms []PlatformInfo

	for _, comp := range view.Components {
		for _, uow := range comp.UoWs {
			for _, art := range uow.Artifacts {
				platform := inferPlatformFromPath(art.Path)
				if platform == "" || seen[platform] {
					continue
				}
				seen[platform] = true

				os, arch := parsePlatformString(platform)
				if os != "" && arch != "" {
					platforms = append(platforms, PlatformInfo{OS: os, Arch: arch})
				}
			}
		}
	}

	return platforms
}

// inferPlatformFromPath tries to extract platform info from artifact path.
// Common patterns: "eac-linux-amd64", "app-darwin-arm64.exe", "linux/amd64"
func inferPlatformFromPath(path string) string {
	patterns := []struct {
		pattern  string
		platform string
	}{
		{"linux-amd64", "linux-amd64"},
		{"linux-arm64", "linux-arm64"},
		{"darwin-amd64", "darwin-amd64"},
		{"darwin-arm64", "darwin-arm64"},
		{"windows-amd64", "windows-amd64"},
		{"windows-arm64", "windows-arm64"},
		{"linux/amd64", "linux-amd64"},
		{"linux/arm64", "linux-arm64"},
	}

	lowerPath := strings.ToLower(path)
	for _, p := range patterns {
		if strings.Contains(lowerPath, p.pattern) {
			return p.platform
		}
	}

	return ""
}

// parsePlatformString parses "os-arch" or "os/arch" format into separate OS and Arch.
func parsePlatformString(platform string) (os, arch string) {
	for i := len(platform) - 1; i >= 0; i-- {
		if platform[i] == '-' || platform[i] == '/' {
			return platform[:i], platform[i+1:]
		}
	}
	return "", ""
}
