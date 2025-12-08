// Package internal provides shared infrastructure for GET and SHOW commands
package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

// ResolvedArtifact represents an artifact with metadata overrides applied and existence checked
type ResolvedArtifact struct {
	// From artifact definition
	Type      string   `json:"type" yaml:"type"`
	ID        string   `json:"id" yaml:"id"`
	Pattern   string   `json:"pattern" yaml:"pattern"`
	Platforms []string `json:"platforms,omitempty" yaml:"platforms,omitempty"`
	VerifyMode string  `json:"verify_mode,omitempty" yaml:"verify_mode,omitempty"`

	// Resolved information
	ResolvedName     string `json:"resolved_name" yaml:"resolved_name"`
	ResolvedPath     string `json:"resolved_path" yaml:"resolved_path"`
	MetadataOverride string `json:"metadata_override,omitempty" yaml:"metadata_override,omitempty"`
	Exists           bool   `json:"exists" yaml:"exists"`
	IsDirectory      bool   `json:"is_directory,omitempty" yaml:"is_directory,omitempty"`
	Error            string `json:"error,omitempty" yaml:"error,omitempty"`
}

// ArtifactResolutionSummary provides summary statistics for artifact resolution
type ArtifactResolutionSummary struct {
	Total     int `json:"total" yaml:"total"`
	Exists    int `json:"exists" yaml:"exists"`
	Missing   int `json:"missing" yaml:"missing"`
	Overrides int `json:"overrides" yaml:"overrides"`
}

// ResolveArtifactsForModule resolves all artifacts for a module with metadata applied
func ResolveArtifactsForModule(
	module *config.Module,
	moduleType *config.ModuleTypeDef,
	buildDir string,
	targetOS, targetArch string,
) ([]ResolvedArtifact, *ArtifactResolutionSummary, error) {
	return ResolveArtifactsForModuleWithConfig(module, moduleType, buildDir, targetOS, targetArch, nil)
}

// ResolveArtifactsForModuleWithConfig resolves all artifacts for a module with optional books config
func ResolveArtifactsForModuleWithConfig(
	module *config.Module,
	moduleType *config.ModuleTypeDef,
	buildDir string,
	targetOS, targetArch string,
	cfg *config.EACConfig,
) ([]ResolvedArtifact, *ArtifactResolutionSummary, error) {
	if module == nil {
		return nil, nil, fmt.Errorf("module cannot be nil")
	}
	if moduleType == nil || moduleType.Build == nil {
		// No artifacts defined for this type
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

	// Check if this is a book module with PDF wildcards that need expansion
	artifacts := moduleType.Build.Artifacts
	if cfg != nil && isBookModule(module.Type) {
		// For resolution/validation, expand all books (not just default)
		// Filtering by default/all is done later based on requested artifacts
		artifacts = expandBookArtifacts(module, moduleType.Build.Artifacts, cfg, true)
	}

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

// deriveArtifactID derives an ID for an artifact based on its type and context
func deriveArtifactID(artifact config.Artifact, targetOS, targetArch string) string {
	// If artifact has explicit ID, use it
	if artifact.ID != "" {
		return artifact.ID
	}

	// For executables, use {os}-{arch} format with compression suffix if applicable
	if artifact.Type == config.ArtifactTypeExecutable {
		baseID := fmt.Sprintf("%s-%s", targetOS, targetArch)

		// Append compression suffix for compressed variants
		if artifact.Compression == config.CompressionUPX {
			return baseID + "-upx"
		} else if artifact.Compression == config.CompressionStrip {
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

// FormatArtifactStatus returns a human-readable status string for an artifact
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

// FormatArtifactSize returns a human-readable size string for an artifact
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

// isBookModule checks if a module type is a book module
func isBookModule(moduleType string) bool {
	return strings.Contains(moduleType, "mkdocs")
}

// expandBookArtifacts expands wildcard PDF patterns to specific book PDFs
func expandBookArtifacts(module *config.Module, artifacts []config.Artifact, cfg *config.EACConfig, buildAll bool) []config.Artifact {
	var expanded []config.Artifact

	for _, artifact := range artifacts {
		// Check if this is a PDF wildcard that needs expansion
		if artifact.Type == config.ArtifactTypeFile && artifact.Pattern == "*.pdf" {
			// Get books for this module
			books := cfg.GetBooksByModule(module.Moniker)

			// Expand to specific book PDFs
			for _, book := range books {
				// Skip non-default books unless --all flag is used
				if !buildAll && !book.IsDefault() {
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

// ModuleValidationResult represents validation results for a single module
type ModuleValidationResult struct {
	Moniker          string                     `json:"moniker" yaml:"moniker"`
	Type             string                     `json:"type" yaml:"type"`
	IsDependency     bool                       `json:"is_dependency" yaml:"is_dependency"`
	HasBuildArtifacts bool                      `json:"has_build_artifacts" yaml:"has_build_artifacts"`
	Artifacts        []ResolvedArtifact         `json:"artifacts" yaml:"artifacts"`
	Summary          *ArtifactResolutionSummary `json:"summary" yaml:"summary"`
	Error            string                     `json:"error,omitempty" yaml:"error,omitempty"`
}

// ValidationResults represents validation results for a module and all dependencies
type ValidationResults struct {
	TargetModule string                   `json:"target_module" yaml:"target_module"`
	Modules      []ModuleValidationResult `json:"modules" yaml:"modules"`
	TotalModules int                      `json:"total_modules" yaml:"total_modules"`
	PassedCount  int                      `json:"passed_count" yaml:"passed_count"`
	FailedCount  int                      `json:"failed_count" yaml:"failed_count"`
	Passed       bool                     `json:"passed" yaml:"passed"`
}

// ValidateArtifactsWithDependencies validates artifacts for a module and all transitive dependencies
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

	// Load build manifest to get requested artifacts for each module
	buildOutputDir := filepath.Join(workspaceRoot, cfg.Repository.Paths.Out.Build)
	manifest, err := LoadManifest(buildOutputDir)
	requestedArtifactsMap := make(map[string][]string)
	if err == nil && manifest != nil {
		// Extract requested artifacts for all modules
		for moniker, moduleBuild := range manifest.Modules {
			requestedArtifactsMap[moniker] = moduleBuild.RequestedArtifacts
		}
	}
	// If manifest doesn't exist or can't be loaded, requestedArtifactsMap will be empty
	// and validation will check all artifacts (fallback behavior)

	// Validate each module
	results := &ValidationResults{
		TargetModule: targetModule,
		Modules:      []ModuleValidationResult{},
	}

	for moniker := range allModules {
		requestedArtifacts := requestedArtifactsMap[moniker]
		modResult := validateSingleModule(moniker, targetModule, cfg, registry, targetOS, targetArch, workspaceRoot, requestedArtifacts)
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

// validateSingleModule validates artifacts for a single module
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
		result.Error = fmt.Sprintf("module contract not found")
		return result
	}
	result.Type = moduleContract.Type

	// Get module from config
	module, ok := cfg.Modules.GetModule(moniker)
	if !ok {
		result.Error = fmt.Sprintf("module not found in config")
		return result
	}

	// Get module type
	moduleType := cfg.ModuleTypes.Get(module.Type)
	if moduleType == nil {
		result.Error = fmt.Sprintf("module type '%s' not found", module.Type)
		return result
	}

	// Check if module has build artifacts
	if moduleType.Build == nil || len(moduleType.Build.Artifacts) == 0 {
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
		module, moduleType, buildDir, targetOS, targetArch, cfg,
	)
	if err != nil {
		result.Error = err.Error()
		return result
	}

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
				if artifact.Exists {
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
		result.Artifacts = artifacts
		result.Summary = summary
	}

	return result
}

// addDependenciesRecursive recursively adds all dependencies of a module
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
// based on the module type and whether --all mode is requested
func DetermineRequestedArtifacts(
	module *config.Module,
	moduleType *config.ModuleTypeDef,
	buildAll bool,
	cfg *config.EACConfig,
) []string {
	if module == nil || moduleType == nil || moduleType.Build == nil {
		return []string{}
	}

	artifacts := moduleType.Build.Artifacts
	if cfg != nil && isBookModule(module.Type) {
		artifacts = expandBookArtifacts(module, moduleType.Build.Artifacts, cfg, buildAll)
	}

	// Check if running in CI (for compression artifacts)
	isCI := os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"

	var requestedIDs []string
	currentOS := runtime.GOOS
	currentArch := runtime.GOARCH

	for _, artifact := range artifacts {
		// Skip UPX-compressed artifacts in non-CI environments (tools not available locally)
		if artifact.Compression == config.CompressionUPX && !isCI {
			continue
		}
		if artifact.Type == config.ArtifactTypeExecutable && buildAll && len(artifact.Platforms) > 0 {
			// For executables in --all mode, build for all specified platforms
			for _, os := range artifact.Platforms {
				// Determine architectures from the pattern
				var archs []string
				pattern := artifact.Pattern

				// Parse pattern to determine supported architectures
				if strings.Contains(pattern, "-amd64") || strings.Contains(pattern, "{os}-amd64") {
					// amd64-only pattern (includes UPX variants: {moniker}-{os}-amd64-upx{ext})
					archs = []string{"amd64"}
				} else if strings.Contains(pattern, "-arm64") || strings.Contains(pattern, "{os}-arm64") {
					// arm64-only pattern
					archs = []string{"arm64"}
				} else if os == "windows" {
					// Generic pattern on Windows - only amd64 supported
					archs = []string{"amd64"}
				} else {
					// Generic pattern on Linux/Darwin - both architectures
					archs = []string{"amd64", "arm64"}
				}

				for _, arch := range archs {
					artifactID := deriveArtifactID(artifact, os, arch)
					requestedIDs = append(requestedIDs, artifactID)
				}
			}
		} else {
			// For non-executables or default mode, use current platform
			artifactID := deriveArtifactID(artifact, currentOS, currentArch)
			requestedIDs = append(requestedIDs, artifactID)
		}
	}

	return requestedIDs
}
