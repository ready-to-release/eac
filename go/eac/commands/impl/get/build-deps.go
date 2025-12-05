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

	get "github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetBuildDeps)
}

// BuildDepsResult contains the build dependencies for a module
type BuildDepsResult struct {
	Module   string   `json:"module" yaml:"module"`
	Type     string   `json:"type" yaml:"type"`
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

	buildDeps := cfg.ModuleTypes.GetBuildDeps(module.Type)
	if buildDeps == nil {
		buildDeps = []string{}
	}

	// Use the shared get command helper for output formatting
	return get.ExecuteGetCommand(func() (interface{}, error) {
		return BuildDepsResult{
			Module:    moniker,
			Type:      module.Type,
			BuildDeps: buildDeps,
		}, nil
	})
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

	module, exists := moduleReport.Registry.Get(moniker)
	if !exists {
		return "", fmt.Errorf("module not found: %s", moniker)
	}

	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		return "", fmt.Errorf("module types configuration not loaded")
	}

	buildDeps := cfg.ModuleTypes.GetBuildDeps(module.Type)
	if buildDeps == nil {
		return "", nil
	}

	return strings.Join(buildDeps, ","), nil
}
