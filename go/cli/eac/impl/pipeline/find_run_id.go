// Command: pipeline find-run-id
// Short: Find workflow run ID by SHA
//
//	--workflow <name>: Workflow file name (required)
//	--sha <sha>: HEAD SHA to find (required)
//	--status <status>: Filter by status (optional, e.g., success)
//
// Long:
// Long: Finds a workflow run with the given HEAD SHA. Returns the run ID
// Long: which can be used with gh run download.
// Long:
// Long: Output: Run ID (empty if not found)
// Long:
// Long: Example:
// Long:   pipeline find-run-id --workflow ci-r2r-cli.yaml --sha abc123
// Long:   pipeline find-run-id --workflow ci-docs.yaml --sha abc123 --status success
package pipeline

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/repository"

	"github.com/ready-to-release/eac/go/clibase/ghexec"
)

func init() {
	registry.Register(PipelineFindRunID)
}

func PipelineFindRunID() int {
	// Parse flags
	workflow := ""
	sha := ""
	status := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--workflow" && i+1 < len(os.Args):
			workflow = os.Args[i+1]
			i++
		case arg == "--sha" && i+1 < len(os.Args):
			sha = os.Args[i+1]
			i++
		case arg == "--status" && i+1 < len(os.Args):
			status = os.Args[i+1]
			i++
		}
	}

	if workflow == "" {
		fmt.Fprintf(os.Stderr, "Error: --workflow is required\n")
		return 1
	}
	if sha == "" {
		fmt.Fprintf(os.Stderr, "Error: --sha is required\n")
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Use GitHub API
	api := github.Global()
	if api == nil {
		api = github.NewGHClient(ghexec.New(workspaceRoot), workspaceRoot)
	}

	// List runs and filter
	runs, err := api.ListRuns(workflow, github.ListRunsOpts{
		Status: status,
		Limit:  10,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Find matching SHA
	for _, run := range runs {
		if run.HeadSHA == sha {
			fmt.Println(run.ID)
			return 0
		}
	}

	// Not found - empty output, success exit (caller checks empty)
	return 0
}
