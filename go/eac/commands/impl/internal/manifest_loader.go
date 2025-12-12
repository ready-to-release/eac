// Package internal provides build manifest loading and validation
package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// ManifestLoadResult represents the result of loading and validating a module's manifest
type ManifestLoadResult struct {
	Moniker          string
	Manifest         *ModuleManifest
	SchemaValid      bool
	ArtifactsValid   bool
	MissingArtifacts []string
	Error            string
}

// ManifestValidationSummary summarizes validation across all modules
type ManifestValidationSummary struct {
	Results       []ManifestLoadResult
	TotalModules  int
	ValidModules  int
	InvalidModules int
	MissingModules int
	AllValid      bool
}

// LoadAndValidateManifests loads and validates build manifests for a list of modules.
// It performs:
// 1. Schema validation against the build-manifest contract
// 2. Artifact existence validation (files actually exist on disk)
func LoadAndValidateManifests(workspaceRoot string, monikers []string, cfg *config.EACConfig) (*ManifestValidationSummary, error) {
	validator, err := GetManifestValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest validator: %w", err)
	}

	summary := &ManifestValidationSummary{
		Results:      make([]ManifestLoadResult, 0, len(monikers)),
		TotalModules: len(monikers),
	}

	for _, moniker := range monikers {
		result := loadAndValidateModule(workspaceRoot, moniker, cfg, validator)
		summary.Results = append(summary.Results, result)

		if result.Error != "" {
			summary.MissingModules++
		} else if result.SchemaValid && result.ArtifactsValid {
			summary.ValidModules++
		} else {
			summary.InvalidModules++
		}
	}

	summary.AllValid = summary.ValidModules == summary.TotalModules
	return summary, nil
}

// loadAndValidateModule loads and validates a single module's manifest
func loadAndValidateModule(workspaceRoot, moniker string, cfg *config.EACConfig, validator *ManifestValidator) ManifestLoadResult {
	result := ManifestLoadResult{
		Moniker: moniker,
	}

	// Check if the module declares build artifacts
	// Modules without artifacts don't produce manifests - they pass validation automatically
	if cfg != nil && cfg.Repository != nil {
		if module, ok := cfg.Repository.GetModule(moniker); ok {
			hasArtifacts := module.Build != nil && len(module.Build.Artifacts) > 0
			if !hasArtifacts {
				// Module doesn't declare artifacts - no manifest expected
				result.SchemaValid = true
				result.ArtifactsValid = true
				return result
			}
		}
	}

	// Get module build directory (relative path)
	var moduleBuildDirRel string
	if cfg != nil {
		moduleBuildDirRel = cfg.Repository.BuildOutputPath(moniker)
	} else {
		// Fallback: use default hardcoded path when config not provided
		// This maintains backward compatibility for tests and legacy callers
		moduleBuildDirRel = filepath.Join("out", "build", moniker)
	}

	// Make absolute path for file operations
	moduleBuildDir := filepath.Join(workspaceRoot, moduleBuildDirRel)

	// Load manifest
	manifest, err := LoadModuleManifest(moduleBuildDir)
	if err != nil {
		result.Error = fmt.Sprintf("manifest not found: %v", err)
		return result
	}
	result.Manifest = manifest

	// Validate against schema
	if err := validator.ValidateManifest(manifest); err != nil {
		result.SchemaValid = false
		result.Error = fmt.Sprintf("schema validation failed: %v", err)
		return result
	}
	result.SchemaValid = true

	// Validate artifacts exist on disk
	missingArtifacts := validateArtifactsExist(moduleBuildDir, manifest)
	result.MissingArtifacts = missingArtifacts
	result.ArtifactsValid = len(missingArtifacts) == 0

	if !result.ArtifactsValid {
		result.Error = fmt.Sprintf("missing artifacts: %v", missingArtifacts)
	}

	return result
}

// validateArtifactsExist checks that all artifacts in the manifest exist on disk
func validateArtifactsExist(buildDir string, manifest *ModuleManifest) []string {
	var missing []string

	for _, artifact := range manifest.Artifacts {
		// Path in manifest is relative to module build directory
		artifactPath := filepath.Join(buildDir, artifact.Path)

		// Normalize path separators
		artifactPath = filepath.FromSlash(artifactPath)

		if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
			missing = append(missing, artifact.ID)
		}
	}

	return missing
}

// GetModuleManifests loads manifests for multiple modules, returning a map
func GetModuleManifests(workspaceRoot string, monikers []string, cfg *config.EACConfig) (map[string]*ModuleManifest, error) {
	manifests := make(map[string]*ModuleManifest)

	for _, moniker := range monikers {
		moduleBuildDir := cfg.Repository.BuildOutputPathAbs(workspaceRoot, moniker)
		manifest, err := LoadModuleManifest(moduleBuildDir)
		if err != nil {
			// Module may not have been built yet - that's OK, just skip
			continue
		}
		manifests[moniker] = manifest
	}

	return manifests, nil
}

// IsManifestUpToDate checks if a module's manifest indicates tests are still valid.
// Returns true if:
// 1. Manifest exists and is schema-valid
// 2. All artifacts exist on disk
// 3. The git commit in manifest matches current commit OR verified_unchanged_at matches
func IsManifestUpToDate(workspaceRoot string, moniker string, cfg *config.EACConfig, currentGitCommit string) (bool, string) {
	moduleBuildDir := cfg.Repository.BuildOutputPathAbs(workspaceRoot, moniker)

	// Load manifest
	manifest, err := LoadModuleManifest(moduleBuildDir)
	if err != nil {
		return false, "manifest not found"
	}

	// Validate schema
	validator, err := GetManifestValidator()
	if err != nil {
		return false, "validator error"
	}
	if err := validator.ValidateManifest(manifest); err != nil {
		return false, "schema invalid"
	}

	// Check artifacts exist
	missing := validateArtifactsExist(moduleBuildDir, manifest)
	if len(missing) > 0 {
		return false, fmt.Sprintf("missing artifacts: %v", missing)
	}

	// Check if build is current
	if manifest.GitCommit == currentGitCommit {
		return true, "built at current commit"
	}

	// Check if verified unchanged at current commit
	if manifest.VerifiedUnchangedAt == currentGitCommit {
		return true, "verified unchanged at current commit"
	}

	return false, "build is stale (different git commit)"
}
