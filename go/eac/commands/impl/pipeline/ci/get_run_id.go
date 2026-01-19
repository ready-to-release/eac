// Command: pipeline ci get-run-id
// Short: Get CI run ID for a workflow and commit SHA
// Long: Look up the run ID of a successful CI workflow run for a specific commit.
// Long:
// Long: This command is useful for release workflows that need to download
// Long: artifacts from a CI run at a specific commit SHA.
// Long:
// Long: Expected Output:
// Long:   - Run ID on stdout (for capturing in scripts)
// Long:   - Exit code 0 on success, 1 if no run found
// Long:
// Long: Example:
// Long:   pipeline ci get-run-id --workflow ci-docs.yaml --sha abc123
// Flag.workflow: type=string, usage=Workflow file name to look up
// Flag.sha: type=string, usage=Commit SHA to find run for
package ci

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(PipelineCIGetRunID)
}

func PipelineCIGetRunID() int {
	// Validate flags before parsing (args start at index 4 for "pipeline ci get-run-id")
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}
	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Parse flags (args start at index 4 for "pipeline ci get-run-id")
	workflow := ""
	sha := ""

	for i := 4; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--workflow" && i+1 < len(os.Args):
			workflow = os.Args[i+1]
			i++
		case arg == "--sha" && i+1 < len(os.Args):
			sha = os.Args[i+1]
			i++
		case arg == "--help" || arg == "-h":
			printGetRunIDUsage()
			return 0
		}
	}

	// Validate arguments
	if workflow == "" || sha == "" {
		log.Error("Error: both --workflow and --sha are required")
		printGetRunIDUsage()
		return 1
	}

	// Look up the successful CI run for this workflow and SHA
	runID, err := findRunIDByWorkflowAndSHA(workflow, sha, workspaceRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Output just the run ID (for script capture)
	// Use fmt.Print directly to avoid log prefix
	os.Stdout.WriteString(runID + "\n")
	return 0
}

func findRunIDByWorkflowAndSHA(workflow, sha, workspaceRoot string) (string, error) {
	// Get the successful run for this workflow at this commit SHA
	cmd := exec.Command("gh", "run", "list",
		"--workflow", workflow,
		"--commit", sha,
		"--status", "success",
		"--json", "databaseId",
		"-q", ".[0].databaseId",
	)
	cmd.Dir = workspaceRoot

	output, err := cmd.Output()
	if err != nil {
		// Try to get more info about the error
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			return "", newError("failed to query GitHub: %s", string(exitErr.Stderr))
		}
		return "", newError("failed to query GitHub: %v", err)
	}

	runID := strings.TrimSpace(string(output))
	if runID == "" {
		return "", newError("no successful run found for workflow %s at SHA %s", workflow, sha)
	}

	return runID, nil
}

func printGetRunIDUsage() {
	log.Info("Get CI run ID for a workflow and commit SHA")
	log.Info("")
	log.Info("Usage: pipeline ci get-run-id --workflow <name> --sha <sha>")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --workflow <name>    Workflow file name (e.g., ci-docs.yaml)")
	log.Info("  --sha <sha>          Commit SHA to find run for")
	log.Info("  -h, --help           Show this help message")
	log.Info("")
	log.Info("Examples:")
	log.Info("  pipeline ci get-run-id --workflow ci-docs.yaml --sha abc123def")
	log.Info("  RUN_ID=$(r2r eac pipeline ci get-run-id --workflow ci-books.yaml --sha $SHA)")
}

// newError is a helper to create formatted errors.
func newError(format string, args ...interface{}) error {
	return &simpleError{msg: sprintf(format, args...)}
}

type simpleError struct {
	msg string
}

func (e *simpleError) Error() string {
	return e.msg
}

func sprintf(format string, args ...interface{}) string {
	// Simple sprintf implementation to avoid importing fmt
	result := format
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			result = strings.Replace(result, "%s", v, 1)
			result = strings.Replace(result, "%v", v, 1)
		case error:
			result = strings.Replace(result, "%v", v.Error(), 1)
			result = strings.Replace(result, "%w", v.Error(), 1)
		default:
			result = strings.Replace(result, "%v", "?", 1)
		}
	}
	return result
}
