// Command: get changed-modules-local
// Short: Get modules requiring rebuild based on local build state
//
//	--as-yaml: Output as YAML (default)
//	--as-json: Output as JSON
//	--as-toml: Output as TOML
//
// Long:
// Long: Expected Output:
// Long: YAML list of modules needing rebuild based on local state, including:
// Long:   - Modules requiring rebuild (source files changed since last build)
// Long:   - Up-to-date modules (no changes detected)
// Long:   - Change reasons for each module
// Long:   - Fresh build flag (true if no build state exists)
// Long:   - Detection timestamp
package get

import (
	"fmt"
	"os"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/hash"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(GetChangedModulesLocal)
}

// changedModulesLocalFlags defines valid flags for the get changed-modules-local command

// LocalChangedModulesResult represents the output of the get changed-modules-local command.
type LocalChangedModulesResult struct {
	Modules       []string          `json:"modules" yaml:"modules" toml:"modules"`
	UpToDate      []string          `json:"up_to_date" yaml:"up_to_date" toml:"up_to_date"`
	ChangeReasons map[string]string `json:"change_reasons,omitempty" yaml:"change_reasons,omitempty" toml:"change_reasons,omitempty"`
	IsFreshBuild  bool              `json:"is_fresh_build" yaml:"is_fresh_build" toml:"is_fresh_build"`
	DetectionTime string            `json:"detection_time" yaml:"detection_time" toml:"detection_time"`
}

func GetChangedModulesLocal() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse optional module filter from args
	var requestedModules []string
	args := os.Args[3:] // Skip program name, "get", "changed-modules-local"
	for _, arg := range args {
		// Skip flags
		if len(arg) > 0 && arg[0] == '-' {
			continue
		}
		requestedModules = append(requestedModules, arg)
	}

	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return detectLocalChanges(workspaceRoot, requestedModules)
	})
}

// detectLocalChanges detects which modules need rebuilding based on UoW manifests.
// Uses DiskOutputReader to check UoW-level cache state aggregated to module level.
func detectLocalChanges(workspaceRoot string, requestedModules []string) (*LocalChangedModulesResult, error) {
	startTime := time.Now()

	// Load module registry
	reg, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil, err
	}

	// Collect contracts to check
	var contracts []*modules.ModuleContract
	if len(requestedModules) > 0 {
		for _, moniker := range requestedModules {
			contract, ok := reg.Get(moniker)
			if !ok {
				fmt.Fprintf(os.Stderr, "Warning: module not found: %s\n", moniker)
				continue
			}
			contracts = append(contracts, contract)
		}
	} else {
		contracts = reg.All()
	}

	// Build list of modules to check and their files
	var monikers []string
	moduleFiles := make(map[string][]string)

	for _, contract := range contracts {
		moniker := contract.Moniker
		monikers = append(monikers, moniker)

		// Debug: print module patterns
		if os.Getenv(environments.EnvDebugCacheCmd) != "" {
			patterns := contract.GetGlobPatterns()
			fmt.Fprintf(os.Stderr, "[DEBUG cmd] %s patterns=%v\n", moniker, patterns)
		}

		// Expand glob patterns to get source files
		files, err := hash.ExpandGlobPatterns(workspaceRoot, contract.GetGlobPatterns())
		if err == nil {
			moduleFiles[moniker] = files
		}
	}

	// Debug: print discovered files
	if os.Getenv(environments.EnvDebugCacheCmd) != "" {
		for moniker, files := range moduleFiles {
			fmt.Fprintf(os.Stderr, "[DEBUG cmd] %s files=%v\n", moniker, files)
		}
	}

	// Use DiskOutputReader for UoW-based change detection
	reader := coreoutput.NewReader(workspaceRoot)

	var changedModules []string
	var upToDateModules []string
	changeReasons := make(map[string]string)
	isFreshBuild := true

	for _, moniker := range monikers {
		// Get all UoW manifests for this module
		manifests, err := reader.ListUoWs(core.ActionBuild, moniker)
		if err != nil || len(manifests) == 0 {
			// No manifests = module needs build
			changedModules = append(changedModules, moniker)
			changeReasons[moniker] = "no build manifests found"
			if os.Getenv(environments.EnvDebugCacheCmd) != "" {
				fmt.Fprintf(os.Stderr, "[DEBUG cmd] %s: no manifests\n", moniker)
			}
			continue
		}

		// At least one module has manifests, so it's not a fresh build
		isFreshBuild = false

		// Compute current input hash for the module
		files, ok := moduleFiles[moniker]
		if !ok || len(files) == 0 {
			changedModules = append(changedModules, moniker)
			changeReasons[moniker] = "no source files found"
			continue
		}

		currentHash, err := hash.Files(workspaceRoot, files)
		if err != nil {
			changedModules = append(changedModules, moniker)
			changeReasons[moniker] = fmt.Sprintf("hash error: %v", err)
			continue
		}

		// Check if all manifests have matching input hash
		// A module needs rebuild if ANY UoW has a mismatched hash
		needsRebuild := false
		var mismatchReason string
		for _, manifest := range manifests {
			if manifest.InputHash != currentHash {
				needsRebuild = true
				mismatchReason = fmt.Sprintf("input hash mismatch in %s:%s", manifest.Component, manifest.Tool)
				break
			}
		}

		if needsRebuild {
			changedModules = append(changedModules, moniker)
			changeReasons[moniker] = mismatchReason
			if os.Getenv(environments.EnvDebugCacheCmd) != "" {
				fmt.Fprintf(os.Stderr, "[DEBUG cmd] %s: %s\n", moniker, mismatchReason)
			}
		} else {
			upToDateModules = append(upToDateModules, moniker)
			if os.Getenv(environments.EnvDebugCacheCmd) != "" {
				fmt.Fprintf(os.Stderr, "[DEBUG cmd] %s: up to date\n", moniker)
			}
		}
	}

	// Debug: print results
	if os.Getenv(environments.EnvDebugCacheCmd) != "" {
		fmt.Fprintf(os.Stderr, "[DEBUG cmd] changed=%v upToDate=%v reasons=%v\n",
			changedModules, upToDateModules, changeReasons)
	}

	return &LocalChangedModulesResult{
		Modules:       changedModules,
		UpToDate:      upToDateModules,
		ChangeReasons: changeReasons,
		IsFreshBuild:  isFreshBuild,
		DetectionTime: time.Since(startTime).Round(time.Millisecond).String(),
	}, nil
}
