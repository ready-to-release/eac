package release

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/tool"
)

type releaseExecuteLayersCommand struct{}

var _ core.SimpleCommandPort = (*releaseExecuteLayersCommand)(nil)

func (c *releaseExecuteLayersCommand) Name() string { return "release execute-layers" }

func (c *releaseExecuteLayersCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-execute-layers",
		Short:         "Execute releases layer by layer in dependency order",
		Long:          "Executes pending releases in dependency order, processing one layer at a time.\n\nThis command takes a JSON array of release layers (from check-pending-releases)\nand processes them in order:\n  1. For each module in the current layer:\n     - Dispatches the release workflow with version info\n  2. Waits for all workflows in the layer to complete\n  3. Moves to the next layer\n\nThe release workflow is responsible for creating the git tag on success\n(via gh release create). This ensures tags only exist for successful releases.\n\nThe layers JSON format is:\n  [[{module, version, tag, type}, ...], [...], ...]\n\nExpected Output:\n  - Progress messages for each release\n  - Exit code 0 if all releases succeed\n  - Exit code 1 if any release fails\n\nExample:\n  release execute-layers --layers '[[{\"module\":\"docs\",\"version\":\"2025.0116.1430\",\"tag\":\"docs/2025.0116.1430\",\"type\":\"calver\"}]]'\n  release execute-layers --layers-file layers.json",
		Flags: []core.FlagSpec{
			{Name: "layers", Type: "string", Usage: "JSON array of release layers"},
			{Name: "layers-file", Type: "string", Usage: "File containing JSON array of release layers"},
			{Name: "timeout", Type: "int", DefaultValue: "900", Usage: "Timeout per release in seconds (default: 900)"},
			{Name: "dry-run", Type: "bool", Usage: "Preview without dispatching workflows"},
		},
	}
}

func (c *releaseExecuteLayersCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleaseExecuteLayers()
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

// executeLayersFlags holds parsed CLI flags for the execute-layers command.
type executeLayersFlags struct {
	layersJSON string
	layersFile string
	timeout    int
	dryRun     bool
}

func ReleaseExecuteLayers() int {
	s, exitCode := newReleaseScaffoldNoFlags()
	if s == nil {
		return exitCode
	}
	workspaceRoot := s.WorkspaceRoot

	flags := parseExecuteLayersFlags()

	layersJSON, exitCode := loadLayersJSON(flags.layersJSON, flags.layersFile)
	if exitCode != 0 {
		return exitCode
	}

	layers, exitCode := parseReleaseLayers(layersJSON)
	if exitCode != 0 {
		return exitCode
	}

	if len(layers) == 0 {
		log.Info("No release layers to process")
		return 0
	}

	log.Infof("Processing %d release layer(s) in dependency order...", len(layers))
	log.Info("")

	failedModules := processAllLayers(workspaceRoot, layers, flags.timeout, flags.dryRun)

	return reportLayerResults(failedModules)
}

// parseExecuteLayersFlags extracts CLI flags from os.Args for the execute-layers command.
func parseExecuteLayersFlags() executeLayersFlags {
	f := executeLayersFlags{
		timeout: 900, // 15 minutes default per release
	}

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--layers" && i+1 < len(os.Args):
			f.layersJSON = os.Args[i+1]
			i++
		case arg == "--layers-file" && i+1 < len(os.Args):
			f.layersFile = os.Args[i+1]
			i++
		case arg == "--timeout" && i+1 < len(os.Args):
			if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
				f.timeout = v
			}
			i++
		case arg == "--dry-run":
			f.dryRun = true
		}
	}
	return f
}

// loadLayersJSON resolves the layers JSON string from either direct input or a file path.
// Returns the JSON string and an exit code (0 on success, 1 on failure).
func loadLayersJSON(layersJSON, layersFile string) (string, int) {
	if layersFile != "" {
		data, err := os.ReadFile(layersFile)
		if err != nil {
			log.Errorf("Error reading layers file: %v", err)
			return "", 1
		}
		layersJSON = string(data)
	}

	if layersJSON == "" {
		log.Error("Error: --layers or --layers-file is required")
		return "", 1
	}
	return layersJSON, 0
}

// parseReleaseLayers unmarshals a JSON string into release layer slices.
// Returns the parsed layers and an exit code (0 on success, 1 on failure).
func parseReleaseLayers(layersJSON string) ([][]LayerModule, int) {
	var layers [][]LayerModule
	if err := json.Unmarshal([]byte(layersJSON), &layers); err != nil {
		log.Errorf("Error parsing layers JSON: %v", err)
		return nil, 1
	}
	return layers, 0
}

// processAllLayers iterates through each layer in dependency order, dispatching
// workflows and awaiting their results. Returns the list of failed module names.
func processAllLayers(workspaceRoot string, layers [][]LayerModule, timeout int, dryRun bool) []string {
	var failedModules []string

	for layerIdx, layer := range layers {
		if len(layer) == 0 {
			continue
		}

		log.Info("")
		log.Info("==========================================")
		log.Infof("Layer %d/%d (%d module(s))", layerIdx+1, len(layers), len(layer))
		log.Info("==========================================")

		layerRuns, failed := dispatchLayerModules(workspaceRoot, layer, dryRun)
		failedModules = append(failedModules, failed...)

		if dryRun {
			continue
		}

		failed = awaitLayerRuns(layerRuns, layerIdx, timeout)
		failedModules = append(failedModules, failed...)
	}

	return failedModules
}

// dispatchLayerModules triggers release workflows for all modules in a single layer.
// Returns the list of dispatched runs and any module names that failed to dispatch.
func dispatchLayerModules(workspaceRoot string, layer []LayerModule, dryRun bool) ([]LayerRun, []string) {
	var layerRuns []LayerRun
	var failedModules []string

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

	return layerRuns, failedModules
}

// awaitLayerRuns waits for all dispatched workflow runs in a layer to complete.
// Returns the list of module names that failed or encountered errors.
func awaitLayerRuns(layerRuns []LayerRun, layerIdx, timeout int) []string {
	if len(layerRuns) == 0 {
		return nil
	}

	var failedModules []string

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

	return failedModules
}

// reportLayerResults logs the final outcome and returns the appropriate exit code.
func reportLayerResults(failedModules []string) int {
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
	// Semver releases (clie, eac-ext) require version input
	// Calver releases (books, docs, clie-eac-bundle) auto-generate versions
	args := []string{"workflow", "run", workflow}
	if versionType == "semver" {
		args = append(args, "-f", fmt.Sprintf("version=%s", version))
	}
	output, exitCode, err := tool.GlobalToolSystem().RunToolCombined(context.Background(), "gh", workspaceRoot, args...)
	if err != nil {
		return "", fmt.Errorf("dispatch failed: %w", err)
	}
	if exitCode != 0 {
		return "", fmt.Errorf("dispatch failed: %s (exit %d)", strings.TrimSpace(string(output)), exitCode)
	}

	// Wait a moment for GitHub to register the run
	time.Sleep(3 * time.Second)

	// Find the run ID - look for the most recent queued/in_progress run
	for i := 0; i < 10; i++ {
		output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", workspaceRoot, "run", "list", "-w", workflow, "-L", "1",
			"--json", "databaseId,status",
			"-q", `.[0] | select(.status == "queued" or .status == "in_progress" or .status == "pending") | .databaseId`)
		if err == nil && len(strings.TrimSpace(string(output))) > 0 {
			return strings.TrimSpace(string(output)), nil
		}

		// Also check for recently completed (might have started fast)
		output, err = tool.GlobalToolSystem().RunTool(context.Background(), "gh", workspaceRoot, "run", "list", "-w", workflow, "-L", "1",
			"--json", "databaseId,status,createdAt",
			"-q", `.[0].databaseId`)
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
		output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", ".", "run", "view", runID, "--json", "status,conclusion")
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
