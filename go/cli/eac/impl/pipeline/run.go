// Command: pipeline run
// Short: Execute module pipelines respecting dependencies
// Long: Execute module pipelines respecting dependencies.
// Long:
// Long: This command runs the full pipeline (build, test, validate) for modules in
// Long: dependency order. If no modules are specified, all modules are processed.
// Long:
// Long: Use --changed-only to run pipelines only for modules with uncommitted changes,
// Long: which is useful for incremental CI/CD workflows.
// Long:
// Long: Use --ref to specify a git reference (branch, tag, commit) to compare against
// Long: when determining which modules have changed.
// Long:
// Long: Expected Output:
// Long:   - Per-module pipeline execution results (build, test, validate stages)
// Long:   - Exit code 0 if all pipelines pass
// Long:   - Exit code 1 if any pipeline fails
// Long:
// Long: Example:
// Long:   pipeline run                    # Run all modules
// Long:   pipeline run --changed-only     # Run only changed modules
// Long:   pipeline run core clie   # Run specific modules
// Flag.changed-only: type=bool, usage=Only run pipelines for changed modules
// Flag.ref: type=string, usage=Git ref to compare against (default: current branch)
// Flag.wait: type=bool, usage=Wait for pipeline completion before returning
// Flag.timeout: type=int, usage=Timeout in seconds when waiting (default: no timeout)
package pipeline

import (
	"fmt"
	"os"
	"strings"

	pipelinerunner "github.com/ready-to-release/eac/go/cli/eac/impl/pipeline/helper"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/gitexec"
	"github.com/ready-to-release/eac/go/core/repository"
)

func PipelineRun() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}
	// Parse flags
	changedOnly := false
	wait := false
	timeout := 0
	ref := getCurrentBranch()
	var monikers []string

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--changed-only" {
			changedOnly = true
		} else if arg == "--wait" {
			wait = true
		} else if arg == "--timeout" {
			if i+1 < len(os.Args) {
				// Parse timeout value
				fmt.Sscanf(os.Args[i+1], "%d", &timeout)
				i++
			}
		} else if arg == "--ref" {
			if i+1 < len(os.Args) {
				ref = os.Args[i+1]
				i++
			}
		} else if !strings.HasPrefix(arg, "--") {
			monikers = append(monikers, arg)
		}
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	runner := pipelinerunner.New(workspaceRoot)

	// Configure wait options
	runner.SetWaitOptions(wait, timeout)

	var pipelineErr error
	if changedOnly {
		// --changed-only flag → run only changed modules
		pipelineErr = runner.RunAllChangedPipelines(ref)
	} else if len(monikers) == 0 {
		// No monikers specified → run ALL modules
		pipelineErr = runner.RunAllPipelines(ref)
	} else if len(monikers) == 1 {
		// Single moniker → run single pipeline
		pipelineErr = runner.RunPipeline(monikers[0], ref)
	} else {
		// Multiple monikers → run with dependency ordering
		pipelineErr = runner.RunPipelines(monikers, ref)
	}

	if pipelineErr != nil {
		log.Errorf("Error: %v", pipelineErr)
		return 1
	}

	return 0
}

// getCurrentBranch gets the current git branch name.
func getCurrentBranch() string {
	output, err := gitexec.Run(".", "branch", "--show-current")
	if err != nil {
		return "main"
	}
	return strings.TrimSpace(string(output))
}
