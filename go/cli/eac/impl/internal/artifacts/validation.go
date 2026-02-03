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

// ValidateBuildArtifacts validates that build artifacts exist and are up-to-date for the given modules.
// It performs:
// 1. UoW manifest existence check (manifests exist in out/build/{module}/{component}_{tool}/)
// 2. Artifact existence validation (files actually exist on disk)
// 3. Staleness check (source files unchanged since build)
// Returns ArtifactValidationInfo with details about missing/stale artifacts.
func ValidateBuildArtifacts(
	moduleList []string,
	cfg *config.EACConfig,
	workspaceRoot string,
	moduleRegistry *modules.Registry,
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
				result := reader.ValidateUoW(workunit.ContextBuild, module, uow.Component, uow.Tool)
				if !result.Valid {
					if len(result.MissingArtifacts) > 0 {
						for _, missing := range result.MissingArtifacts {
							moduleErrors = append(moduleErrors, uow.Component+"_"+uow.Tool+": "+missing)
						}
					}
					if len(result.CorruptArtifacts) > 0 {
						for _, corrupt := range result.CorruptArtifacts {
							moduleErrors = append(moduleErrors, uow.Component+"_"+uow.Tool+": hash mismatch for "+corrupt)
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

	// Check for staleness using workunit StateManager change detection
	var staleModules []string
	staleReasons := make(map[string]string)

	if moduleRegistry != nil {
		// Build list of modules with contracts
		var monikers []string
		moduleFiles := make(map[string][]string)
		for _, moniker := range moduleList {
			if contract, ok := moduleRegistry.Get(moniker); ok {
				monikers = append(monikers, moniker)
				// Expand glob patterns to get source files
				files, err := hash.ExpandGlobPatterns(workspaceRoot, contract.GetGlobPatterns())
				if err == nil {
					moduleFiles[moniker] = files
				}
			}
		}

		if len(monikers) > 0 {
			// Create hash provider
			hashProvider := func(module string) (string, error) {
				files, ok := moduleFiles[module]
				if !ok {
					return "", nil
				}
				return hash.Files(workspaceRoot, files)
			}

			stateMgr := workunit.NewStateManager(workspaceRoot)
			rule := workunit.DefaultRules[workunit.ContextBuild]
			changeResult, err := stateMgr.DetectModuleChanges(workunit.ContextBuild, monikers, rule, hashProvider, nil)
			if err != nil {
				log.Debugf("Failed to detect changes for staleness check: %v", err)
			} else if !changeResult.FreshRun {
				// Report changed modules as stale (they need rebuild)
				for _, moniker := range changeResult.ChangedModules {
					// Only report as stale if artifacts are present (otherwise it's already in missingFrom)
					if !utils.Contains(missingFrom, moniker) {
						staleModules = append(staleModules, moniker)
						if reason, ok := changeResult.ChangeReasons[moniker]; ok {
							staleReasons[moniker] = reason
						} else {
							staleReasons[moniker] = "source files changed since build"
						}
					}
				}
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
