// Package artifacts provides build artifact validation utilities
package artifacts

import (
	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/utils"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
	"github.com/ready-to-release/eac/go/eac/core/hash"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/workunit"
)

var log = logging.C()

// ValidateBuildArtifacts validates that build artifacts exist and are up-to-date for the given modules.
// It performs:
// 1. Manifest schema validation against the build-manifest contract
// 2. Artifact existence validation (files actually exist on disk)
// 3. Staleness check (source files unchanged since build)
// Returns ArtifactValidationInfo with details about missing/stale artifacts.
func ValidateBuildArtifacts(
	moduleList []string,
	cfg *config.EACConfig,
	workspaceRoot string,
	moduleRegistry *modules.Registry,
) *initsummary.ArtifactValidationInfo {
	// Use the manifest loader to validate manifests against schema and check artifacts
	summary, err := implinternal.LoadAndValidateManifests(workspaceRoot, moduleList, cfg)
	if err != nil {
		log.Debugf("Manifest validation error: %v", err)
		// If we can't validate, report all modules as missing
		return &initsummary.ArtifactValidationInfo{
			Validated:      true,
			ModulesChecked: moduleList,
			AllPresent:     false,
			MissingFrom:    moduleList,
		}
	}

	var missingFrom []string
	missingDetails := make(map[string][]string)

	for _, result := range summary.Results {
		if result.Error != "" {
			log.Debugf("Module %s: %s", result.Moniker, result.Error)
			missingFrom = append(missingFrom, result.Moniker)
			missingDetails[result.Moniker] = []string{result.Error}
		} else if !result.SchemaValid {
			log.Debugf("Module %s: manifest schema invalid", result.Moniker)
			missingFrom = append(missingFrom, result.Moniker)
			missingDetails[result.Moniker] = []string{"manifest schema invalid"}
		} else if !result.ArtifactsValid {
			log.Debugf("Module %s: missing artifacts %v", result.Moniker, result.MissingArtifacts)
			missingFrom = append(missingFrom, result.Moniker)
			missingDetails[result.Moniker] = result.MissingArtifacts
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
			changeResult, err := stateMgr.DetectModuleChanges(workunit.ContextBuild, monikers, rule, hashProvider)
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
