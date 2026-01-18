// Command: get ci-dispatch-layers
// Short: Get modules in CI dispatch layers based on artifact dependencies
// Long: Computes CI dispatch layers based on which modules need artifacts from other modules' CI.
// Long:
// Long: CI artifact dependencies are declared in module contracts via `depends_on_ci`.
// Long: Modules that don't have CI deps are in layer 0.
// Long: Modules that download from layer 0 modules are in layer 1, etc.
// Long:
// Long: Expected Output:
// Long:   JSON/YAML with layers array, each containing modules to dispatch.
// Long:   Layers should be dispatched in order, waiting for each to complete.
// Long:
// Long: Example:
// Long:   get ci-dispatch-layers --modules "eac-core r2r-cli ext-eac"
// Long:   get ci-dispatch-layers --modules "eac-core r2r-cli ext-eac" --format shell
// Flag.modules: type=string, usage=Space-separated list of modules to dispatch
// Flag.format: type=string, usage=Output format (yaml, json, shell)
package get

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

func init() {
	registry.Register(GetCIDispatchLayers)
}

// CIDispatchLayersResult represents the output of the get ci-dispatch-layers command.
type CIDispatchLayersResult struct {
	// Layers of modules to dispatch, in order
	Layers [][]string `json:"layers" yaml:"layers" toml:"layers"`
	// Total number of modules
	TotalModules int `json:"total_modules" yaml:"total_modules" toml:"total_modules"`
	// Per-module layer assignment
	ModuleLayers map[string]int `json:"module_layers" yaml:"module_layers" toml:"module_layers"`
}

func GetCIDispatchLayers() int {
	// Parse flags
	modulesStr := ""
	format := ""

	for i, arg := range os.Args {
		switch arg {
		case "--modules":
			if i+1 < len(os.Args) {
				modulesStr = os.Args[i+1]
			}
		case "--format":
			if i+1 < len(os.Args) {
				format = os.Args[i+1]
			}
		case "--help", "-h":
			printCIDispatchLayersUsage()
			return 0
		}
	}

	if modulesStr == "" {
		fmt.Fprintf(os.Stderr, "Error: --modules is required\n")
		printCIDispatchLayersUsage()
		return 1
	}

	// Parse modules list
	modules := parseModuleList(modulesStr)
	if len(modules) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no modules specified\n")
		return 1
	}

	// Load config to get CI deps from module contracts
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return 1
	}

	// Build CI artifact deps map from module contracts
	ciArtifactDeps := buildCIArtifactDeps(cfg.Repository)

	// Compute layers
	result := computeCIDispatchLayers(modules, ciArtifactDeps)

	// Handle shell format output
	if format == "shell" {
		// Output layer info to stderr for logging
		fmt.Fprintf(os.Stderr, "CI Dispatch Layers:\n")
		for i, layer := range result.Layers {
			fmt.Fprintf(os.Stderr, "  Layer %d: %s\n", i, strings.Join(layer, " "))
		}
		fmt.Fprintf(os.Stderr, "\n")

		// Output variables to stdout for eval
		fmt.Printf("LAYER_COUNT=%d\n", len(result.Layers))
		for i, layer := range result.Layers {
			fmt.Printf("LAYER_%d=\"%s\"\n", i, strings.Join(layer, " "))
		}
		// Also output as JSON for complex processing
		if layersJSON, err := json.Marshal(result.Layers); err == nil {
			fmt.Printf("LAYERS_JSON='%s'\n", string(layersJSON))
		}
		return 0
	}

	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return result, nil
	})
}

// buildCIArtifactDeps builds a map of module -> CI artifact dependencies from config.
func buildCIArtifactDeps(cfg *config.RepositoryConfig) map[string][]string {
	deps := make(map[string][]string)
	for i := range cfg.Modules {
		module := &cfg.Modules[i]
		if len(module.CIDeps) > 0 {
			deps[module.Moniker] = module.CIDeps
		}
	}
	return deps
}

// computeCIDispatchLayers computes which layer each module belongs to.
func computeCIDispatchLayers(modules []string, ciArtifactDeps map[string][]string) *CIDispatchLayersResult {
	result := &CIDispatchLayersResult{
		Layers:       [][]string{},
		TotalModules: len(modules),
		ModuleLayers: make(map[string]int),
	}

	// Create a set of modules to dispatch
	moduleSet := make(map[string]bool)
	for _, m := range modules {
		moduleSet[m] = true
	}

	// Compute layer for each module
	// A module is in layer N if it depends on a module in layer N-1
	// Layer 0 = modules with no CI deps (or deps not in dispatch set)
	layer0 := []string{}
	layer1 := []string{}

	for _, module := range modules {
		deps, hasDeps := ciArtifactDeps[module]
		if !hasDeps || len(deps) == 0 {
			// No CI artifact deps -> layer 0
			layer0 = append(layer0, module)
			result.ModuleLayers[module] = 0
			continue
		}

		// Check if any deps are in the dispatch set
		hasDispatchedDep := false
		for _, dep := range deps {
			if moduleSet[dep] {
				hasDispatchedDep = true
				break
			}
		}

		if hasDispatchedDep {
			// Dep is being dispatched -> layer 1
			layer1 = append(layer1, module)
			result.ModuleLayers[module] = 1
		} else {
			// Dep is not being dispatched (already has CI) -> layer 0
			layer0 = append(layer0, module)
			result.ModuleLayers[module] = 0
		}
	}

	// Sort layers for deterministic output
	sort.Strings(layer0)
	sort.Strings(layer1)

	// Build layers array
	if len(layer0) > 0 {
		result.Layers = append(result.Layers, layer0)
	}
	if len(layer1) > 0 {
		result.Layers = append(result.Layers, layer1)
	}

	return result
}

func printCIDispatchLayersUsage() {
	fmt.Println("Get modules in CI dispatch layers based on artifact dependencies")
	fmt.Println("")
	fmt.Println("Usage: r2r get ci-dispatch-layers --modules <modules> [flags]")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  --modules <modules>   Space-separated list of modules to dispatch (required)")
	fmt.Println("  --format shell        Output as shell variable assignments for eval")
	fmt.Println("  --as-yaml             Output as YAML (default)")
	fmt.Println("  --as-json             Output as JSON")
	fmt.Println("")
	fmt.Println("CI Artifact Dependencies:")
	fmt.Println("  Declared in module contracts via 'depends_on_ci' field.")
	fmt.Println("  These dependencies require CI artifact downloads before the module's CI can run.")
	fmt.Println("")
	fmt.Println("Layer Logic:")
	fmt.Println("  Layer 0: Modules with no CI artifact deps (or deps not in dispatch set)")
	fmt.Println("  Layer 1: Modules that depend on layer 0 modules")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Get layers for a set of modules")
	fmt.Println("  r2r get ci-dispatch-layers --modules \"eac-core r2r-cli ext-eac\"")
	fmt.Println("")
	fmt.Println("  # Output for shell scripting")
	fmt.Println("  eval $(r2r get ci-dispatch-layers --modules \"r2r-cli ext-eac\" --format shell)")
	fmt.Println("  for i in $(seq 0 $((LAYER_COUNT-1))); do")
	fmt.Println("    layer_var=\"LAYER_$i\"")
	fmt.Println("    echo \"Layer $i: ${!layer_var}\"")
	fmt.Println("  done")
}
