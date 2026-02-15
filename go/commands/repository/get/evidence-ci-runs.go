package get

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/adapters/gh"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

type getEvidenceCIRunsCommand struct{}

var _ core.SimpleCommandPort = (*getEvidenceCIRunsCommand)(nil)

func (c *getEvidenceCIRunsCommand) Name() string { return "get evidence-ci-runs" }

func (c *getEvidenceCIRunsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-evidence-ci-runs",
		Short:         "Get CI run IDs for a module and its dependencies (for evidence building)",
		Long: "Returns a list of {module, workflow, run_id} for downloading test/scan artifacts.\n" +
			"Uses per-module change detection to find the appropriate CI run for each dependency.\n" +
			"Fails if any dependency with a ci-{module}.yaml workflow has no successful CI.\n" +
			"\n" +
			"Expected Output:\n" +
			"YAML/JSON list of CI runs containing:\n" +
			"  - module: the dependency module moniker\n" +
			"  - workflow: the ci-{module}.yaml workflow name\n" +
			"  - run_id: the GitHub Actions run ID to download artifacts from\n" +
			"\n" +
			"Example usage in CI:\n" +
			"  CI_RUNS=$(commands get evidence-ci-runs clie --format json)\n" +
			"  echo \"$CI_RUNS\" | jq -c '.[]' | while read entry; do\n" +
			"    module=$(echo \"$entry\" | jq -r '.module')\n" +
			"    run_id=$(echo \"$entry\" | jq -r '.run_id')\n" +
			"    gh run download \"$run_id\" --pattern \"test-results-${module}*\"\n" +
			"  done",
		Args: "module (required) - Module moniker to get evidence CI runs for",
		Flags: []core.FlagSpec{
			{Name: "format", Type: "string", Usage: "Output format (json outputs JSON array for shell parsing; otherwise uses standard get command formats)"},
		},
	}
}

func (c *getEvidenceCIRunsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetEvidenceCIRuns()
}

// EvidenceCIRun represents a CI run to download artifacts from.
type EvidenceCIRun struct {
	Module   string `json:"module" yaml:"module"`
	Workflow string `json:"workflow" yaml:"workflow"`
	RunID    int    `json:"run_id" yaml:"run_id"`
}

// EvidenceCIRunsResult contains the CI runs needed for evidence building.
type EvidenceCIRunsResult struct {
	Module  string          `json:"module" yaml:"module"`
	CIRuns  []EvidenceCIRun `json:"ci_runs" yaml:"ci_runs"`
	Skipped []string        `json:"skipped,omitempty" yaml:"skipped,omitempty"`
}

func GetEvidenceCIRuns() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "get", and "evidence-ci-runs"

	// Parse module moniker and flags from args
	var moniker string
	format := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--format" && i+1 < len(args) {
			format = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--format=") {
			format = strings.TrimPrefix(arg, "--format=")
		} else if !strings.HasPrefix(arg, "--") && moniker == "" {
			moniker = arg
		}
	}

	if moniker == "" {
		fmt.Fprintf(os.Stderr, "Usage: get evidence-ci-runs <module> [--format json]\n")
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Get the result
	result, err := getEvidenceCIRuns(moniker, workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Handle json format - output just the ci_runs array for shell parsing
	if format == "json" {
		jsonBytes, err := json.Marshal(result.CIRuns)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal JSON: %v\n", err)
			return 1
		}
		fmt.Println(string(jsonBytes))
		return 0
	}

	// Use the shared get command helper for YAML/JSON/TOML output
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return result, nil
	})
}

// getEvidenceCIRuns gets the CI runs needed for evidence building for a module.
func getEvidenceCIRuns(moniker, workspaceRoot string) (*EvidenceCIRunsResult, error) {
	// Load module registry
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load module registry: %w", err)
	}

	// Check module exists
	_, exists := moduleRegistry.Get(moniker)
	if !exists {
		return nil, fmt.Errorf("module not found: %s", moniker)
	}

	// Get transitive dependencies (including the module itself)
	depsToCheck := getTransitiveDependencies(moniker, moduleRegistry)

	// Create GitHub API client
	api := github.NewGHClient(gh.New(tool.GlobalToolSystem(), workspaceRoot), workspaceRoot)

	result := &EvidenceCIRunsResult{
		Module:  moniker,
		CIRuns:  []EvidenceCIRun{},
		Skipped: []string{},
	}

	// Check each module for CI workflow and get last successful run
	var missingCI []string
	for _, dep := range depsToCheck {
		workflowName := fmt.Sprintf("ci-%s.yaml", dep)
		workflowPath := filepath.Join(workspaceRoot, ".github", "workflows", workflowName)

		// Skip modules without CI workflows
		if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
			result.Skipped = append(result.Skipped, dep)
			continue
		}

		// Query for last successful CI run
		runs, err := api.ListRuns(workflowName, github.ListRunsOpts{
			Status: "success",
			Limit:  1,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to query CI runs for %s: %w", dep, err)
		}

		if len(runs) == 0 {
			missingCI = append(missingCI, dep)
			continue
		}

		result.CIRuns = append(result.CIRuns, EvidenceCIRun{
			Module:   dep,
			Workflow: workflowName,
			RunID:    runs[0].ID,
		})
	}

	// Fail if any required module has no successful CI
	if len(missingCI) > 0 {
		return nil, fmt.Errorf("no successful CI run found for module(s): %s", strings.Join(missingCI, ", "))
	}

	return result, nil
}

// getTransitiveDependencies returns a module and all its transitive dependencies.
func getTransitiveDependencies(moniker string, reg *modules.Registry) []string {
	seen := make(map[string]bool)
	result := []string{}

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

		// Recurse into dependencies
		for _, dep := range module.DependsOn {
			collect(dep)
		}
	}

	collect(moniker)
	return result
}
