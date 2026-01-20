package test

import (
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

// FlattenModulesToTestComponentWork converts TestsByModulePath to component work layers.
// Returns two layers: parallel tests first, sequential tests second.
// Returns nil if no tests to execute.
func FlattenModulesToTestComponentWork(ctx *cmdframework.ExecutionContext) [][]orchestrator.ComponentWork {
	testCfg, ok := ctx.Config.Extra["testConfig"].(*TestFrameworkConfig)
	if !ok || testCfg == nil {
		return nil
	}

	testsByModulePath := testCfg.TestsByModulePath
	if len(testsByModulePath) == 0 {
		return nil
	}

	cfg := config.Global()
	if cfg == nil || cfg.ComponentTypes == nil {
		return nil
	}

	var parallelWork []orchestrator.ComponentWork
	var sequentialWork []orchestrator.ComponentWork

	for modulePath, tests := range testsByModulePath {
		if len(tests) == 0 {
			continue
		}

		// Check if any test is sequential
		hasSequential := false
		for i := range tests {
			if tests[i].IsSequential {
				hasSequential = true
				break
			}
		}

		// Determine weight from test type -> component type mapping
		weight := getTestComponentWeight(tests, cfg)

		work := orchestrator.ComponentWork{
			Module:        extractMonikerFromModulePath(modulePath),
			Component:     modulePath, // Use full path as component identifier
			ComponentType: getTestType(tests),
			Handler:       "test",
			Weight:        weight,
			BuildAfter:    nil, // Tests don't have intra-module deps
			Index:         0,   // Will be set per-layer below
		}

		if hasSequential {
			sequentialWork = append(sequentialWork, work)
		} else {
			parallelWork = append(parallelWork, work)
		}
	}

	// Set indices per-layer (Index must be relative to the layer, not global)
	for i := range parallelWork {
		parallelWork[i].Index = i
	}
	for i := range sequentialWork {
		sequentialWork[i].Index = i
	}

	// Build layers: parallel first, sequential second
	var layers [][]orchestrator.ComponentWork
	if len(parallelWork) > 0 {
		layers = append(layers, parallelWork)
	}
	if len(sequentialWork) > 0 {
		layers = append(layers, sequentialWork)
	}

	return layers
}

// getTestComponentWeight determines the weight for a set of tests based on their type.
func getTestComponentWeight(tests []testing.TestReference, cfg *config.EACConfig) int {
	if len(tests) == 0 {
		return 1
	}

	// Map test type to component type for weight lookup
	testType := tests[0].Type
	var compTypeName string

	switch testType {
	case "gotest", "godog":
		compTypeName = "go"
	case "mocha", "tscucumber":
		compTypeName = "typescript"
	default:
		compTypeName = "go" // Default
	}

	compType := cfg.ComponentTypes.Get(compTypeName)
	if compType != nil {
		return compType.GetTestWeight()
	}

	return 1
}

// getTestType returns the test type from the first test reference.
func getTestType(tests []testing.TestReference) string {
	if len(tests) == 0 {
		return "unknown"
	}
	return tests[0].Type
}

// extractMonikerFromModulePath extracts the module moniker from a module path.
// Input: "eac-core/config" or "eac-commands"
// Output: "eac-core" or "eac-commands"
func extractMonikerFromModulePath(modulePath string) string {
	if idx := strings.Index(modulePath, "/"); idx >= 0 {
		return modulePath[:idx]
	}
	return modulePath
}
