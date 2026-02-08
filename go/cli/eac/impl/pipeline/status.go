// Command: pipeline status
// Short: Show CI pipeline status for commits
// Long: Show CI pipeline status for commits.
// Long:
// Long: This command displays the status of GitHub Actions workflows for a specific
// Long: commit or branch. By default, it shows the status for the current branch HEAD.
// Long:
// Long: Use --ref to check status for a specific branch or tag.
// Long: Use --commit to check status for a specific commit SHA.
// Long:
// Long: Expected Output:
// Long:   - Commit SHA being checked
// Long:   - List of workflow names with their status (success, failure, in progress)
// Long:   - Exit code 0 for success (even if workflows are failing)
// Long:   - Exit code 1 for errors (missing GitHub CLI, invalid ref, etc.)
// Long:
// Long: Example:
// Long:   pipeline status                    # Check current branch HEAD
// Long:   pipeline status --ref develop      # Check develop branch
// Long:   pipeline status --commit abc123    # Check specific commit
// Flag.ref: type=string, usage=Git ref (branch/tag) to check status for
// Flag.commit: type=string, usage=Commit SHA to check status for
package pipeline

import (
	"os"
	"strings"

	pipelinerunner "github.com/ready-to-release/eac/go/cli/eac/impl/pipeline/helper"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/repository"
)

func PipelineStatus() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags
	var ref string
	var commit string

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--ref" {
			if i+1 < len(os.Args) {
				ref = os.Args[i+1]
				i++
			}
		} else if arg == "--commit" {
			if i+1 < len(os.Args) {
				commit = os.Args[i+1]
				i++
			}
		}
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Create GitHub CLI wrapper
	ghCLI := pipelinerunner.NewGitHubCLI(workspaceRoot)

	// Get commit SHA
	var commitSHA string
	if commit != "" {
		// Use provided commit SHA
		commitSHA = commit
	} else {
		// Get SHA for ref (or HEAD if not specified)
		if ref == "" {
			ref = "HEAD"
		}
		commitSHA, err = ghCLI.GetCommitSHA(ref)
		if err != nil {
			// Check if this is an "invalid ref" error
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "invalid ref") {
				log.Errorf("Error: %v", err)
				return 1
			}
			log.Errorf("Error: failed to get commit SHA for ref %s: %v", ref, err)
			return 1
		}
	}

	// List workflow runs for the commit
	runs, err := ghCLI.ListWorkflowRuns(commitSHA)
	if err != nil {
		// Check if this is a "gh CLI not available" error
		if strings.Contains(err.Error(), "gh") || strings.Contains(err.Error(), "GitHub CLI") {
			log.Errorf("Error: GitHub CLI (gh) is required but not available: %v", err)
			return 1
		}
		log.Errorf("Error: failed to list workflow runs: %v", err)
		return 1
	}

	// Display results
	log.Infof("Pipeline status for commit: %s", commitSHA)
	log.Infof("")

	if len(runs) == 0 {
		log.Infof("No workflow runs found for this commit")
		return 0
	}

	log.Infof("Workflows:")
	for _, run := range runs {
		status := formatWorkflowStatus(run)
		log.Infof("  %s: %s", run.WorkflowName, status)
	}

	return 0
}

// formatWorkflowStatus formats the workflow status for display.
func formatWorkflowStatus(run pipelinerunner.WorkflowRunSummary) string {
	if run.Status == "completed" {
		switch run.Conclusion {
		case "success":
			return "✓ success"
		case "failure":
			return "✗ failure"
		case "cancelled":
			return "⊘ cancelled"
		case "skipped":
			return "- skipped"
		default:
			return run.Conclusion
		}
	}
	return run.Status
}
