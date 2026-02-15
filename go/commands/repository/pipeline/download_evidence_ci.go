package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/adapters/gh"
	"github.com/ready-to-release/eac/go/commands/repository/get"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/tool"
)

// getEvidenceCIRunsInternal gets CI runs for a module and its dependencies
// Returns (ciRuns, skippedModules, error).
func getEvidenceCIRunsInternal(moniker, workspaceRoot string) ([]get.EvidenceCIRun, []string, error) {
	// Load module registry
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load module registry: %w", err)
	}

	// Check module exists
	_, exists := moduleRegistry.Get(moniker)
	if !exists {
		return nil, nil, fmt.Errorf("module not found: %s", moniker)
	}

	// Get transitive dependencies
	deps := getTransitiveDeps(moniker, moduleRegistry)

	// Create GitHub API client
	api := github.NewGHClient(gh.New(tool.GlobalToolSystem(), workspaceRoot), workspaceRoot)

	var ciRuns []get.EvidenceCIRun
	var skipped []string
	var missingCI []string

	for _, dep := range deps {
		workflowName := fmt.Sprintf("ci-%s.yaml", dep)
		workflowPath := filepath.Join(workspaceRoot, ".github", "workflows", workflowName)

		// Skip modules without CI workflows
		if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
			skipped = append(skipped, dep)
			continue
		}

		// Query for last successful CI run
		runs, err := api.ListRuns(workflowName, github.ListRunsOpts{
			Status: "success",
			Limit:  1,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to query CI runs for %s: %w", dep, err)
		}

		if len(runs) == 0 {
			missingCI = append(missingCI, dep)
			continue
		}

		ciRuns = append(ciRuns, get.EvidenceCIRun{
			Module:   dep,
			Workflow: workflowName,
			RunID:    runs[0].ID,
		})
	}

	if len(missingCI) > 0 {
		return nil, nil, fmt.Errorf("no successful CI run found for module(s): %s", strings.Join(missingCI, ", "))
	}

	return ciRuns, skipped, nil
}

// getTransitiveDeps returns a module and all its transitive dependencies.
func getTransitiveDeps(moniker string, reg *modules.Registry) []string {
	seen := make(map[string]bool)
	var result []string

	var collect func(m string)
	collect = func(m string) {
		if seen[m] {
			return
		}
		seen[m] = true

		module, exists := reg.Get(m)
		if !exists {
			return
		}

		result = append(result, m)

		for _, dep := range module.DependsOn {
			collect(dep)
		}
	}

	collect(moniker)
	return result
}
