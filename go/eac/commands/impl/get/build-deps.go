// Command: get build-deps
// Description: Get build dependencies for a module
// Args: module (required) - Module moniker
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
package get

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetBuildDeps)
}

// BuildDepsResult contains the build dependencies for a module
type BuildDepsResult struct {
	Module    string   `json:"module" yaml:"module"`
	Type      string   `json:"type" yaml:"type"`
	BuildDeps []string `json:"build_deps" yaml:"build_deps"`
}

func GetBuildDeps() int {
	args := os.Args[3:] // Skip program name, "get", and "build-deps"

	// Parse module moniker from args (skip flags)
	var moniker string
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--") {
			moniker = arg
			break
		}
	}

	if moniker == "" {
		log.Errorf("Usage: get build-deps <module>")
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Load module contracts to get module type
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		log.Errorf("failed to load module contracts: %v", err)
		return 1
	}

	// Find the module
	module, exists := moduleReport.Registry.Get(moniker)
	if !exists {
		log.Errorf("module not found: %s", moniker)
		return 1
	}

	// Get build deps from module types config
	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		log.Errorf("module types configuration not loaded")
		return 1
	}

	// Aggregate build deps from module and all its dependencies
	buildDeps := aggregateBuildDeps(moniker, moduleReport.Registry, cfg.ModuleTypes)

	// Use the shared get command helper for output formatting
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return BuildDepsResult{
			Module:    moniker,
			Type:      module.Type,
			BuildDeps: buildDeps,
		}, nil
	})
}

// aggregateBuildDeps collects build dependencies from a module and all its dependencies
func aggregateBuildDeps(moniker string, registry *modules.Registry, moduleTypes *config.ModuleTypesConfig) []string {
	seen := make(map[string]bool)
	depsSet := make(map[string]bool)

	var collect func(m string)
	collect = func(m string) {
		if seen[m] {
			return
		}
		seen[m] = true

		module, exists := registry.Get(m)
		if !exists {
			return
		}

		// Add this module's build deps
		deps := moduleTypes.GetBuildDeps(module.Type)
		for _, dep := range deps {
			depsSet[dep] = true
		}

		// Recurse into dependencies
		for _, depMoniker := range module.DependsOn {
			collect(depMoniker)
		}
	}

	collect(moniker)

	// Convert set to sorted slice for consistent output
	result := make([]string, 0, len(depsSet))
	for dep := range depsSet {
		result = append(result, dep)
	}

	// Sort for consistent output
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// GetBuildDepsPlain returns build deps as comma-separated string (for shell scripts)
func GetBuildDepsPlain(moniker string) (string, error) {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return "", fmt.Errorf("failed to find repository root: %v", err)
	}

	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load module contracts: %v", err)
	}

	_, exists := moduleReport.Registry.Get(moniker)
	if !exists {
		return "", fmt.Errorf("module not found: %s", moniker)
	}

	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		return "", fmt.Errorf("module types configuration not loaded")
	}

	// Aggregate build deps from module and all its dependencies
	buildDeps := aggregateBuildDeps(moniker, moduleReport.Registry, cfg.ModuleTypes)
	if len(buildDeps) == 0 {
		return "", nil
	}

	return strings.Join(buildDeps, ","), nil
}
