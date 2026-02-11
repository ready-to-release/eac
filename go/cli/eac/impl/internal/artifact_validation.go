// Package internal provides shared infrastructure for GET and SHOW commands
package internal

import (
	"fmt"
	"path/filepath"
	"sort"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
)

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
