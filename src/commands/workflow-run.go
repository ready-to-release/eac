// Command: workflow run
// Description: Execute module workflows respecting dependencies
// Usage:
//   workflow run                    # Run ALL modules
//   workflow run --changed-only     # Run only changed modules
//   workflow run <moniker>          # Run specific module
//   workflow run <m1> <m2> ...      # Run multiple modules
// Flags:
//   --changed-only: Only run workflows for changed modules
//   --ref <ref>: Git ref to use (default: current branch)
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/src/commands/workflowrunner"
)

func init() {
	Register("workflow run", WorkflowRun)
}

func WorkflowRun() int {
	// Parse flags
	changedOnly := false
	ref := getCurrentBranch()
	var monikers []string

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--changed-only" {
			changedOnly = true
		} else if arg == "--ref" {
			if i+1 < len(os.Args) {
				ref = os.Args[i+1]
				i++
			}
		} else if !strings.HasPrefix(arg, "--") {
			monikers = append(monikers, arg)
		}
	}

	runner := workflowrunner.New("../..", "0.1.0")

	var err error
	if changedOnly {
		// --changed-only flag → run only changed modules
		err = runner.RunAllChangedWorkflows(ref)
	} else if len(monikers) == 0 {
		// No monikers specified → run ALL modules
		err = runner.RunAllWorkflows(ref)
	} else if len(monikers) == 1 {
		// Single moniker → run single workflow
		err = runner.RunWorkflow(monikers[0], ref)
	} else {
		// Multiple monikers → run with dependency ordering
		err = runner.RunWorkflows(monikers, ref)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}

// getCurrentBranch gets the current git branch name
func getCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "main"
	}
	return strings.TrimSpace(string(output))
}
