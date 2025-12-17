// Command: pipeline wait
// Short: Wait for GitHub workflow runs to complete
// Long: Wait for GitHub workflow runs to complete with live progress display.
// Long:
// Long: This command monitors one or more GitHub Actions workflow runs and displays
// Long: their status in a compact, updating format. It exits with code 0 if all
// Long: workflows succeed, or code 1 if any fail.
// Long:
// Long: Expected Output:
// Long:   - Live progress display with status icons (✓, ✗, ◐, ○)
// Long:   - Exit code 0 if all workflows succeed
// Long:   - Exit code 1 if any workflow fails
// Long:
// Long: Example:
// Long:   pipeline wait 12345 12346 12347    # Wait for specific run IDs
// Long:   pipeline wait --timeout 600        # Wait up to 10 minutes (default: 30 min)
// Long:   pipeline wait --interval 5         # Poll every 5 seconds (default: 10)
// Flag.timeout: type=int, usage=Maximum wait time in seconds (default: 1800)
// Flag.interval: type=int, usage=Poll interval in seconds (default: 10)
package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

var pipelineWaitFlags = []flags.FlagDefinition{
	{Name: "--timeout", HasValue: true, ValueType: "int"},
	{Name: "--interval", HasValue: true, ValueType: "int"},
}

func init() {
	registry.Register(PipelineWait)
}

// WorkflowRun represents the status of a GitHub Actions workflow run
type WorkflowRun struct {
	DatabaseID int    `json:"databaseId"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	URL        string `json:"url"`
}

func PipelineWait() int {
	// Validate flags before parsing
	if err := flags.ValidateFlags(os.Args[3:], pipelineWaitFlags); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags and run IDs
	timeout := 1800 // 30 minutes default
	interval := 10  // 10 seconds default
	var runIDs []string

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--timeout" && i+1 < len(os.Args) {
			if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
				timeout = v
			}
			i++
		} else if arg == "--interval" && i+1 < len(os.Args) {
			if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
				interval = v
			}
			i++
		} else if !strings.HasPrefix(arg, "--") {
			runIDs = append(runIDs, arg)
		}
	}

	if len(runIDs) == 0 {
		log.Errorf("Error: no workflow run IDs provided")
		log.Errorf("Usage: pipeline wait <run-id> [run-id...]")
		return 1
	}

	// Track status of each run
	type runStatus struct {
		id         string
		name       string
		status     string
		conclusion string
		done       bool
	}

	runs := make([]runStatus, len(runIDs))
	for i, id := range runIDs {
		runs[i] = runStatus{id: id, status: "pending", name: fmt.Sprintf("run-%s", id)}
	}

	startTime := time.Now()
	lastLineCount := 0

	for {
		elapsed := time.Since(startTime)
		if elapsed.Seconds() > float64(timeout) {
			log.Info("")
			log.Infof("❌ Timeout after %v", elapsed.Round(time.Second))
			return 1
		}

		// Check status of all runs
		allDone := true
		anyFailed := false

		for i := range runs {
			if runs[i].done {
				continue
			}

			run, err := getWorkflowRunStatus(runs[i].id)
			if err != nil {
				runs[i].status = "error"
				continue
			}

			runs[i].name = run.Name
			runs[i].status = run.Status
			runs[i].conclusion = run.Conclusion

			if run.Status == "completed" {
				runs[i].done = true
				if run.Conclusion != "success" {
					anyFailed = true
				}
			} else {
				allDone = false
			}
		}

		// Clear previous output
		for i := 0; i < lastLineCount; i++ {
			log.Info("\033[A\033[K") // Move up and clear line
		}

		// Print compact status
		lineCount := 0
		elapsedStr := formatDuration(elapsed)

		log.Infof("⏱  %s", elapsedStr)
		lineCount++

		for _, r := range runs {
			var icon string
			switch {
			case r.done && r.conclusion == "success":
				icon = "✓"
			case r.done && r.conclusion != "success":
				icon = "✗"
			case r.status == "in_progress":
				icon = "◐"
			case r.status == "queued":
				icon = "○"
			default:
				icon = "?"
			}
			log.Infof("  %s %s", icon, r.name)
		}
		log.Info("")

		lastLineCount = lineCount

		if allDone {
			log.Info("")
			if anyFailed {
				log.Info("❌ One or more workflows failed")
				return 1
			}
			log.Info("✓ All workflows completed successfully")
			return 0
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// getWorkflowRunStatus fetches the status of a workflow run from GitHub
func getWorkflowRunStatus(runID string) (*WorkflowRun, error) {
	cmd := exec.Command("gh", "run", "view", runID, "--json", "databaseId,name,status,conclusion,url")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get run status: %w", err)
	}

	var run WorkflowRun
	if err := json.Unmarshal(output, &run); err != nil {
		return nil, fmt.Errorf("failed to parse run status: %w", err)
	}

	return &run, nil
}

// formatDuration formats a duration as MM:SS or HH:MM:SS
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
