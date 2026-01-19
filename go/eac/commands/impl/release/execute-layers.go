// Command: release execute-layers
// Short: Execute releases layer by layer in dependency order
// Long: Executes pending releases in dependency order, processing one layer at a time.
// Long:
// Long: This command takes a JSON array of release layers (from check-pending-releases)
// Long: and processes them in order:
// Long:   1. For each module in the current layer:
// Long:      - Dispatches the release workflow with version info
// Long:   2. Waits for all workflows in the layer to complete
// Long:   3. Moves to the next layer
// Long:
// Long: The release workflow is responsible for creating the git tag on success
// Long: (via gh release create). This ensures tags only exist for successful releases.
// Long:
// Long: The layers JSON format is:
// Long:   [[{module, version, tag, type}, ...], [...], ...]
// Long:
// Long: Expected Output:
// Long:   - Progress messages for each release
// Long:   - Exit code 0 if all releases succeed
// Long:   - Exit code 1 if any release fails
// Long:
// Long: Example:
// Long:   release execute-layers --layers '[[{"module":"docs","version":"2025.0116.1430","tag":"docs/2025.0116.1430","type":"calver"}]]'
// Long:   release execute-layers --layers-file layers.json
// Flag.layers: type=string, usage=JSON array of release layers
// Flag.layers-file: type=string, usage=File containing JSON array of release layers
// Flag.timeout: type=int, default=900, usage=Timeout per release in seconds (default: 900)
// Flag.dry-run: type=bool, usage=Preview without dispatching workflows
package release

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ReleaseExecuteLayers)
}

// LayerModule represents a module in a release layer.
type LayerModule struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	Tag     string `json:"tag"`
	Type    string `json:"type"`
}

// LayerRun tracks a dispatched workflow run.
type LayerRun struct {
	Module string
	RunID  string
}

func ReleaseExecuteLayers() int {
	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Parse flags
	layersJSON := ""
	layersFile := ""
	timeout := 900 // 15 minutes default per release
	dryRun := false

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--layers" && i+1 < len(os.Args):
			layersJSON = os.Args[i+1]
			i++
		case arg == "--layers-file" && i+1 < len(os.Args):
			layersFile = os.Args[i+1]
			i++
		case arg == "--timeout" && i+1 < len(os.Args):
			if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
				timeout = v
			}
			i++
		case arg == "--dry-run":
			dryRun = true
		}
	}

	// Load layers JSON
	if layersFile != "" {
		data, err := os.ReadFile(layersFile)
		if err != nil {
			log.Errorf("Error reading layers file: %v", err)
			return 1
		}
		layersJSON = string(data)
	}

	if layersJSON == "" {
		log.Error("Error: --layers or --layers-file is required")
		return 1
	}

	// Parse layers
	var layers [][]LayerModule
	if err := json.Unmarshal([]byte(layersJSON), &layers); err != nil {
		log.Errorf("Error parsing layers JSON: %v", err)
		return 1
	}

	if len(layers) == 0 {
		log.Info("No release layers to process")
		return 0
	}

	log.Infof("Processing %d release layer(s) in dependency order...", len(layers))
	log.Info("")

	failedModules := []string{}

	// Process each layer sequentially
	for layerIdx, layer := range layers {
		if len(layer) == 0 {
			continue
		}

		log.Info("")
		log.Info("==========================================")
		log.Infof("Layer %d/%d (%d module(s))", layerIdx+1, len(layers), len(layer))
		log.Info("==========================================")

		// Collect run IDs for this layer
		var layerRuns []LayerRun

		// Trigger all modules in this layer
		for _, mod := range layer {
			log.Info("")
			log.Infof("  [%s] (%s) Releasing version: %s", mod.Module, mod.Type, mod.Version)

			workflow := fmt.Sprintf("release-%s.yaml", mod.Module)
			workflowPath := filepath.Join(workspaceRoot, ".github", "workflows", workflow)

			if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
				log.Warnf("  [%s] No release workflow found: %s", mod.Module, workflow)
				continue
			}

			if dryRun {
				log.Infof("  [%s] (dry-run) Would dispatch %s with version=%s", mod.Module, workflow, mod.Version)
				continue
			}

			// Dispatch workflow (version only for semver modules)
			log.Infof("  [%s] Dispatching %s", mod.Module, workflow)
			runID, err := dispatchAndGetRunID(workspaceRoot, workflow, mod.Version, mod.Type)
			if err != nil {
				log.Errorf("  [%s] Error dispatching workflow: %v", mod.Module, err)
				failedModules = append(failedModules, mod.Module)
				continue
			}

			if runID != "" {
				log.Infof("  [%s] Run ID: %s", mod.Module, runID)
				layerRuns = append(layerRuns, LayerRun{Module: mod.Module, RunID: runID})
			}
		}

		if dryRun {
			continue
		}

		// Await all runs in this layer
		if len(layerRuns) > 0 {
			log.Info("")
			log.Infof("  Awaiting layer %d completion...", layerIdx+1)

			for _, run := range layerRuns {
				success, err := awaitWorkflowRun(run.RunID, timeout)
				if err != nil {
					log.Errorf("  [%s] Error awaiting: %v", run.Module, err)
					failedModules = append(failedModules, fmt.Sprintf("%s(error)", run.Module))
				} else if !success {
					log.Errorf("  [%s] ❌ Failed", run.Module)
					failedModules = append(failedModules, run.Module)
				} else {
					log.Infof("  [%s] ✅ Completed successfully", run.Module)
				}
			}
		}
	}

	log.Info("")
	if len(failedModules) > 0 {
		log.Warnf("Some releases failed: %s", strings.Join(failedModules, ", "))
		return 1
	}
	log.Info("✅ All releases completed successfully")
	return 0
}

// dispatchAndGetRunID dispatches a workflow and returns the run ID.
func dispatchAndGetRunID(workspaceRoot, workflow, version, versionType string) (string, error) {
	// Dispatch the workflow
	// Semver releases (r2r-cli, ext-eac) require version input
	// Calver releases (books, docs, r2r-eac-bundle) auto-generate versions
	args := []string{"workflow", "run", workflow}
	if versionType == "semver" {
		args = append(args, "-f", fmt.Sprintf("version=%s", version))
	}
	cmd := exec.Command("gh", args...)
	cmd.Dir = workspaceRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("dispatch failed: %s: %w", strings.TrimSpace(string(output)), err)
	}

	// Wait a moment for GitHub to register the run
	time.Sleep(3 * time.Second)

	// Find the run ID - look for the most recent queued/in_progress run
	for i := 0; i < 10; i++ {
		cmd = exec.Command("gh", "run", "list", "-w", workflow, "-L", "1",
			"--json", "databaseId,status",
			"-q", `.[0] | select(.status == "queued" or .status == "in_progress" or .status == "pending") | .databaseId`)
		cmd.Dir = workspaceRoot

		output, err := cmd.Output()
		if err == nil && len(strings.TrimSpace(string(output))) > 0 {
			return strings.TrimSpace(string(output)), nil
		}

		// Also check for recently completed (might have started fast)
		cmd = exec.Command("gh", "run", "list", "-w", workflow, "-L", "1",
			"--json", "databaseId,status,createdAt",
			"-q", `.[0].databaseId`)
		cmd.Dir = workspaceRoot

		output, err = cmd.Output()
		if err == nil && len(strings.TrimSpace(string(output))) > 0 {
			return strings.TrimSpace(string(output)), nil
		}

		time.Sleep(2 * time.Second)
	}

	return "", fmt.Errorf("could not find workflow run after dispatch")
}

// awaitWorkflowRun waits for a workflow run to complete.
func awaitWorkflowRun(runID string, timeout int) (success bool, err error) {
	startTime := time.Now()

	for {
		elapsed := time.Since(startTime)
		if elapsed.Seconds() >= float64(timeout) {
			return false, fmt.Errorf("timeout after %v", elapsed.Round(time.Second))
		}

		// Get status
		cmd := exec.Command("gh", "run", "view", runID, "--json", "status,conclusion")
		output, err := cmd.Output()
		if err != nil {
			return false, err
		}

		var result struct {
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		}
		if err := json.Unmarshal(output, &result); err != nil {
			return false, err
		}

		if result.Status == "completed" {
			return result.Conclusion == "success", nil
		}

		time.Sleep(30 * time.Second)
	}
}
