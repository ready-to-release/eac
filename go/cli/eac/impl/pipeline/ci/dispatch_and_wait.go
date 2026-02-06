// Command: pipeline ci dispatch-and-wait
// Short: Wait for GitHub workflow runs to complete
// Long: Dispatch a workflow and wait for it to complete, or wait for an existing run.
// Long:
// Long: This command is useful for CI orchestration when you need to trigger
// Long: a workflow and wait for its completion before proceeding.
// Long:
// Long: Expected Output:
// Long:   - Workflow dispatch confirmation message
// Long:   - Live progress display with status updates
// Long:   - Exit code 0 on success, 1 on failure
// Long:
// Long: Example:
// Long:   pipeline ci dispatch-and-wait --workflow ci-r2r-cli.yaml --ref main
// Long:   pipeline ci dispatch-and-wait --run-id 12345678
// Long:   pipeline ci dispatch-and-wait --workflow ci-r2r-cli.yaml --ref main --timeout 600
// Flag.workflow: type=string, usage=Workflow file name to dispatch
// Flag.ref: type=string, usage=Git ref to run workflow on (default: current branch)
// Flag.run-id: type=string, usage=Existing run ID to wait for (skips dispatch)
// Flag.timeout: type=int, usage=Timeout in seconds (default: 300)
// Flag.inputs: type=string, usage=Workflow inputs as JSON object
package ci

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/ghexec"
	"github.com/ready-to-release/eac/go/clibase/gitexec"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(PipelineCIDispatchAndWait)
}

func PipelineCIDispatchAndWait() int {
	// Validate flags before parsing (args start at index 4 for "pipeline ci dispatch-and-wait")
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

	// Parse flags (args start at index 4 for "pipeline ci dispatch-and-wait")
	workflow := ""
	ref := ""
	runID := ""
	timeout := 300
	inputs := ""

	for i := 4; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--workflow" && i+1 < len(os.Args):
			workflow = os.Args[i+1]
			i++
		case arg == "--ref" && i+1 < len(os.Args):
			ref = os.Args[i+1]
			i++
		case arg == "--run-id" && i+1 < len(os.Args):
			runID = os.Args[i+1]
			i++
		case arg == "--timeout" && i+1 < len(os.Args):
			t, err := strconv.Atoi(os.Args[i+1])
			if err == nil {
				timeout = t
			}
			i++
		case arg == "--inputs" && i+1 < len(os.Args):
			inputs = os.Args[i+1]
			i++
		case arg == "--help" || arg == "-h":
			printDispatchAndWaitUsage()
			return 0
		}
	}

	// Validate arguments
	if runID == "" && workflow == "" {
		log.Error("Error: either --workflow or --run-id is required")
		printDispatchAndWaitUsage()
		return 1
	}

	// If run-id is provided, just wait for it
	if runID != "" {
		return waitForRun(runID, timeout, workspaceRoot)
	}

	// Otherwise, dispatch and wait
	return dispatchAndWait(workflow, ref, inputs, timeout, workspaceRoot)
}

func dispatchAndWait(workflow, ref, inputs string, timeout int, workspaceRoot string) int {
	// Get current ref if not specified
	if ref == "" {
		output, err := gitexec.Run(workspaceRoot, "rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			log.Errorf("Error getting current branch: %v", err)
			return 1
		}
		ref = strings.TrimSpace(string(output))
	}

	log.Infof("Dispatching workflow: %s on ref: %s", workflow, ref)

	// Build dispatch command
	args := []string{"workflow", "run", workflow, "--ref", ref}
	if inputs != "" {
		// Parse inputs JSON to pass as individual -f flags
		var inputMap map[string]string
		if err := json.Unmarshal([]byte(inputs), &inputMap); err != nil {
			log.Errorf("Error parsing inputs JSON: %v", err)
			return 1
		}
		for k, v := range inputMap {
			args = append(args, "-f", fmt.Sprintf("%s=%s", k, v))
		}
	}

	_, dispatchErr := ghexec.Run(workspaceRoot, args...)
	if dispatchErr != nil {
		log.Errorf("Error dispatching workflow: %v", dispatchErr)
		return 1
	}

	// Wait a moment for the run to appear
	time.Sleep(config.CIDispatchSettleTime())

	// Find the most recent run for this workflow
	runID, err := findLatestRunID(workflow, ref, workspaceRoot)
	if err != nil {
		log.Errorf("Error finding run ID: %v", err)
		return 1
	}

	log.Infof("Workflow dispatched. Run ID: %s", runID)

	return waitForRun(runID, timeout, workspaceRoot)
}

func findLatestRunID(workflow, ref, workspaceRoot string) (string, error) {
	// Get the most recent run for this workflow on this ref
	output, err := ghexec.Run(workspaceRoot, "run", "list",
		"--workflow", workflow,
		"--branch", ref,
		"--limit", "1",
		"--json", "databaseId",
		"-q", ".[0].databaseId",
	)
	if err != nil {
		return "", fmt.Errorf("failed to find run: %w", err)
	}

	runID := strings.TrimSpace(string(output))
	if runID == "" {
		return "", fmt.Errorf("no runs found for workflow %s on ref %s", workflow, ref)
	}

	return runID, nil
}

func waitForRun(runID string, timeout int, workspaceRoot string) int {
	log.Infof("Waiting for run %s to complete (timeout: %ds)...", runID, timeout)

	startTime := time.Now()
	pollInterval := config.CIPollInterval()

	for {
		// Check if timeout exceeded
		if time.Since(startTime) > time.Duration(timeout)*time.Second {
			log.Errorf("Timeout exceeded waiting for run %s", runID)
			return 1
		}

		// Get run status
		status, conclusion, err := getRunStatus(runID, workspaceRoot)
		if err != nil {
			log.Errorf("Error checking run status: %v", err)
			return 1
		}

		if status == "completed" {
			log.Infof("Run %s completed with conclusion: %s", runID, conclusion)
			if conclusion == "success" {
				return 0
			}
			return 1
		}

		log.Infof("Run status: %s (elapsed: %s)", status, time.Since(startTime).Round(time.Second))
		time.Sleep(pollInterval)
	}
}

func getRunStatus(runID, workspaceRoot string) (status, conclusion string, err error) {
	output, err := ghexec.Run(workspaceRoot, "run", "view", runID,
		"--json", "status,conclusion",
		"-q", ".status + \"|\" + .conclusion",
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to get run status: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) >= 1 {
		status = parts[0]
	}
	if len(parts) >= 2 {
		conclusion = parts[1]
	}

	return status, conclusion, nil
}

func printDispatchAndWaitUsage() {
	log.Info("Dispatch a workflow and wait for completion, or wait for an existing run")
	log.Info("")
	log.Info("Usage: pipeline ci dispatch-and-wait [flags]")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --workflow <name>    Workflow file name to dispatch (e.g., ci-r2r-cli.yaml)")
	log.Info("  --ref <ref>          Git ref to run workflow on (default: current branch)")
	log.Info("  --run-id <id>        Existing run ID to wait for (skips dispatch)")
	log.Info("  --timeout <seconds>  Timeout in seconds (default: 300)")
	log.Info("  --inputs <json>      Workflow inputs as JSON object")
	log.Info("  -h, --help           Show this help message")
	log.Info("")
	log.Info("Examples:")
	log.Info("  pipeline ci dispatch-and-wait --workflow ci-r2r-cli.yaml --ref main")
	log.Info("  pipeline ci dispatch-and-wait --run-id 12345678 --timeout 600")
	log.Info("  pipeline ci dispatch-and-wait --workflow ci-r2r-cli.yaml --inputs '{\"version\":\"1.0.0\"}'")
}
