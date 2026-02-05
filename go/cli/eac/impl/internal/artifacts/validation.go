// Package artifacts provides build artifact validation utilities
package artifacts

import (
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/clibase/utils"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/logging"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/workunit"
)

var log = logging.C()

// filterToExpected returns only manifests matching the expected UoWs list.
// If expected is nil or empty, returns all manifests unchanged.
func filterToExpected(manifests []*coreoutput.UoWManifest, expected []workunit.UnitID) []*coreoutput.UoWManifest {
	if len(expected) == 0 {
		return manifests
	}

	// Build set of expected component-tool keys (dash separator matches UnitID.DirName())
	expectedSet := make(map[string]bool)
	for _, id := range expected {
		key := id.Component + "-" + id.Tool
		expectedSet[key] = true
	}

	var filtered []*coreoutput.UoWManifest
	for _, m := range manifests {
		key := m.Component + "-" + m.Tool
		if expectedSet[key] {
			filtered = append(filtered, m)
		} else {
			log.Debugf("Filtering out orphaned manifest: %s/%s-%s", m.Module, m.Component, m.Tool)
		}
	}
	return filtered
}

// ValidateBuildArtifacts validates that build artifacts exist and are up-to-date for the given modules.
// It performs:
// 1. UoW manifest existence check (manifests exist in out/build/{module}/{component}-{tool}/)
// 2. Artifact existence validation (files actually exist on disk)
// 3. Staleness check (source files unchanged since build)
// Returns ArtifactValidationInfo with details about missing/stale artifacts.
//
// Note: This function checks ALL manifests on disk. If you want to filter to only
// currently-expected UoWs (to ignore orphaned manifests from removed components),
// use ValidateBuildArtifactsWithExpected instead.
func ValidateBuildArtifacts(
	moduleList []string,
	cfg *config.EACConfig,
	workspaceRoot string,
	moduleRegistry *modules.Registry,
) *initsummary.ArtifactValidationInfo {
	return ValidateBuildArtifactsWithExpected(moduleList, cfg, workspaceRoot, moduleRegistry, nil)
}

// ValidateBuildArtifactsWithExpected validates build artifacts, optionally filtering to expected UoWs.
// If expectedUoWs is nil, all manifests on disk are checked (legacy behavior).
// If expectedUoWs is provided, only manifests matching those UoWs are checked for staleness,
// preventing false positives from orphaned manifests left by removed component types.
func ValidateBuildArtifactsWithExpected(
	moduleList []string,
	cfg *config.EACConfig,
	workspaceRoot string,
	moduleRegistry *modules.Registry,
	expectedUoWs map[string][]workunit.UnitID,
) *initsummary.ArtifactValidationInfo {
	reader := coreoutput.NewReader(workspaceRoot)

	var missingFrom []string
	missingDetails := make(map[string][]string)

	// Validate each module using UoW manifests
	for _, module := range moduleList {
		// Check if any UoW manifests exist for this module
		if !reader.HasManifests(workunit.ContextBuild, module) {
			log.Debugf("Module %s: no UoW manifests found", module)
			missingFrom = append(missingFrom, module)
			missingDetails[module] = []string{"no build manifests found"}
			continue
		}

		// Get module view to validate all UoWs
		moduleView, err := reader.GetModule(workunit.ContextBuild, module)
		if err != nil {
			log.Debugf("Module %s: failed to read UoW manifests: %v", module, err)
			missingFrom = append(missingFrom, module)
			missingDetails[module] = []string{err.Error()}
			continue
		}

		// Validate artifacts for each component/UoW
		var moduleErrors []string
		for _, comp := range moduleView.Components {
			for _, uow := range comp.UoWs {
				id := workunit.UnitID{
					Context:   workunit.ContextBuild,
					Module:    module,
					Component: uow.Component,
					Tool:      uow.Tool,
					Extra:     uow.Extra,
				}
				result := reader.ValidateUoW(id)
				if !result.Valid {
					uowName := uow.DirName()
					if len(result.MissingArtifacts) > 0 {
						for _, missing := range result.MissingArtifacts {
							moduleErrors = append(moduleErrors, uowName+": "+missing)
						}
					}
					if len(result.CorruptArtifacts) > 0 {
						for _, corrupt := range result.CorruptArtifacts {
							moduleErrors = append(moduleErrors, uowName+": hash mismatch for "+corrupt)
						}
					}
					if result.Error != nil && len(moduleErrors) == 0 {
						moduleErrors = append(moduleErrors, result.Error.Error())
					}
				}
			}
		}

		if len(moduleErrors) > 0 {
			log.Debugf("Module %s: validation errors %v", module, moduleErrors)
			missingFrom = append(missingFrom, module)
			missingDetails[module] = moduleErrors
		}
	}

	// Check for staleness using UoW manifest input hashes
	var staleModules []string
	staleReasons := make(map[string]string)

	if moduleRegistry != nil {
		// Build map of current source hashes per module
		moduleFiles := make(map[string][]string)
		for _, moniker := range moduleList {
			if contract, ok := moduleRegistry.Get(moniker); ok {
				files, err := hash.ExpandGlobPatterns(workspaceRoot, contract.GetGlobPatterns())
				if err == nil {
					moduleFiles[moniker] = files
				}
			}
		}

		// Check each module's UoW manifests for staleness
		for _, moniker := range moduleList {
			// Skip modules already in missingFrom (no artifacts)
			if utils.Contains(missingFrom, moniker) {
				continue
			}

			// Get all UoW manifests for this module
			manifests, err := reader.ListUoWs(workunit.ContextBuild, moniker)
			if err != nil || len(manifests) == 0 {
				// No manifests = skip (already checked for missing)
				continue
			}

			// Filter to expected UoWs if provided (ignores orphaned manifests)
			if expectedUoWs != nil {
				manifests = filterToExpected(manifests, expectedUoWs[moniker])
				if len(manifests) == 0 {
					// No expected manifests found for this module
					continue
				}
			}

			// Compute current input hash
			files, ok := moduleFiles[moniker]
			if !ok || len(files) == 0 {
				continue
			}

			currentHash, err := hash.Files(workspaceRoot, files)
			if err != nil {
				continue
			}

			// Check staleness: a module is stale only if NO manifest matches the current hash.
			// This is resilient to hash divergence where parallel component builds captured
			// different filesystem snapshots (e.g., due to go mod tidy modifying go.sum).
			// If at least one manifest matches, the sources haven't changed since that build.
			hasNonEmpty := false
			anyMatch := false
			for _, manifest := range manifests {
				// Skip manifests with empty InputHash (legacy artifacts from pre-UoW-cache)
				if manifest.InputHash == "" {
					log.Debugf("Skipping manifest with empty InputHash: %s/%s_%s",
						moniker, manifest.Component, manifest.Tool)
					continue
				}
				hasNonEmpty = true
				if manifest.InputHash == currentHash {
					anyMatch = true
					break
				}
			}
			if hasNonEmpty && !anyMatch {
				staleModules = append(staleModules, moniker)
				staleReasons[moniker] = "source files changed since build"
			}
		}
	}

	return &initsummary.ArtifactValidationInfo{
		Validated:              true,
		ModulesChecked:         moduleList,
		AllPresent:             len(missingFrom) == 0,
		MissingFrom:            missingFrom,
		MissingArtifactDetails: missingDetails,
		AllCurrent:             len(staleModules) == 0,
		StaleModules:           staleModules,
		StaleReasons:           staleReasons,
	}
}
