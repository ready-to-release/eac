// Command: pipeline check-recent-run
// Short: Check if a recent successful workflow run exists
//   --workflow <name>: Workflow file name (required)
//   --sha <sha>: HEAD SHA to check (default: current HEAD)
//   --since <duration>: Time window to check (default: 2h)
//   --format shell: Output as shell variables
// Long:
// Long: Checks if a successful workflow run exists for the given SHA within
// Long: the specified time window. Used to skip redundant CI runs.
// Long:
// Long: Output:
// Long:   Default: "true" or "false"
// Long:   --format shell: HAS_RECENT="true/false"
// Long:
// Long: Example:
// Long:   pipeline check-recent-run --workflow ci-trigger.yaml --since 2h
package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/github"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(PipelineCheckRecentRun)
}

func PipelineCheckRecentRun() int {
	// Parse flags
	workflow := ""
	sha := ""
	since := 2 * time.Hour
	format := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--workflow" && i+1 < len(os.Args):
			workflow = os.Args[i+1]
			i++
		case arg == "--sha" && i+1 < len(os.Args):
			sha = os.Args[i+1]
			i++
		case arg == "--since" && i+1 < len(os.Args):
			d, err := time.ParseDuration(os.Args[i+1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid duration: %v\n", err)
				return 1
			}
			since = d
			i++
		case arg == "--format" && i+1 < len(os.Args):
			format = os.Args[i+1]
			i++
		}
	}

	if workflow == "" {
		fmt.Fprintf(os.Stderr, "Error: --workflow is required\n")
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Default to HEAD
	if sha == "" {
		cmd := exec.Command("git", "rev-parse", "HEAD")
		cmd.Dir = workspaceRoot
		output, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get HEAD SHA: %v\n", err)
			return 1
		}
		sha = strings.TrimSpace(string(output))
	}

	// Use GitHub API
	api := github.Global()
	if api == nil {
		api = github.NewGHClient(workspaceRoot)
	}

	hasRecent, err := api.HasRecentSuccess(workflow, sha, since)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if format == "shell" {
		fmt.Printf("HAS_RECENT=\"%t\"\n", hasRecent)
	} else {
		fmt.Println(hasRecent)
	}

	return 0
}
