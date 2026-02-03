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

	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/workunit"
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

// detectLocalChanges detects which modules need rebuilding based on build state.
// Uses workunit.StateManager.DetectModuleChanges for change detection.
func detectLocalChanges(workspaceRoot string, requestedModules []string) (*LocalChangedModulesResult, error) {
	// Load module registry
	reg, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil, err
	}

	// Build list of modules to check and their files
	var monikers []string
	moduleFiles := make(map[string][]string)

	if len(requestedModules) > 0 {
		// Only check requested modules
		for _, moniker := range requestedModules {
			contract, ok := reg.Get(moniker)
			if !ok {
				fmt.Fprintf(os.Stderr, "Warning: module not found: %s\n", moniker)
				continue
			}
			monikers = append(monikers, moniker)
			// Debug: print module patterns
			if os.Getenv("DEBUG_CACHE_CMD") != "" {
				patterns := contract.GetGlobPatterns()
				fmt.Fprintf(os.Stderr, "[DEBUG cmd] %s patterns=%v\n", moniker, patterns)
			}
			// Expand glob patterns to get source files
			files, err := hash.ExpandGlobPatterns(workspaceRoot, contract.GetGlobPatterns())
			if err == nil {
				moduleFiles[moniker] = files
			}
		}
	} else {
		// Check all modules
		for _, contract := range reg.All() {
			monikers = append(monikers, contract.Moniker)
			// Debug: print module patterns
			if os.Getenv("DEBUG_CACHE_CMD") != "" {
				patterns := contract.GetGlobPatterns()
				fmt.Fprintf(os.Stderr, "[DEBUG cmd] %s patterns=%v\n", contract.Moniker, patterns)
			}
			// Expand glob patterns to get source files
			files, err := hash.ExpandGlobPatterns(workspaceRoot, contract.GetGlobPatterns())
			if err == nil {
				moduleFiles[contract.Moniker] = files
			}
		}
	}

	// Debug: print discovered files
	if os.Getenv("DEBUG_CACHE_CMD") != "" {
		for moniker, files := range moduleFiles {
			fmt.Fprintf(os.Stderr, "[DEBUG cmd] %s files=%v\n", moniker, files)
		}
	}

	// Create hash provider for change detection
	hashProvider := func(module string) (string, error) {
		files, ok := moduleFiles[module]
		if !ok {
			return "", fmt.Errorf("no files for module %s", module)
		}
		return hash.Files(workspaceRoot, files)
	}

	// Use workunit StateManager for change detection
	stateMgr := workunit.NewStateManager(workspaceRoot)
	rule := workunit.DefaultRules[workunit.ContextBuild]
	changeResult, err := stateMgr.DetectModuleChanges(workunit.ContextBuild, monikers, rule, hashProvider, nil)
	if err != nil {
		return nil, err
	}

	// Debug: print results
	if os.Getenv("DEBUG_CACHE_CMD") != "" {
		fmt.Fprintf(os.Stderr, "[DEBUG cmd] changed=%v upToDate=%v reasons=%v\n",
			changeResult.ChangedModules, changeResult.UpToDateModules, changeResult.ChangeReasons)
	}

	return &LocalChangedModulesResult{
		Modules:       changeResult.ChangedModules,
		UpToDate:      changeResult.UpToDateModules,
		ChangeReasons: changeResult.ChangeReasons,
		IsFreshBuild:  changeResult.FreshRun,
		DetectionTime: changeResult.DetectionTime.Round(time.Millisecond).String(),
	}, nil
}
