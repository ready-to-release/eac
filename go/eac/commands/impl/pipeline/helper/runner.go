// Package pipelinerunner provides functionality to execute GitHub workflows
// respecting module dependencies
package pipelinerunner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

var log = logging.C()

// PipelineRunner orchestrates execution of module pipelines.
type PipelineRunner struct {
	repoPath string
	ghCLI    GitHubCLI
}

// New creates a new PipelineRunner.
func New(repoPath string) *PipelineRunner {
	return &PipelineRunner{
		repoPath: repoPath,
		ghCLI:    NewGitHubCLI(repoPath),
	}
}

// RunPipeline executes a single pipeline.
func (r *PipelineRunner) RunPipeline(moniker, ref string) error {
	workflowFile := moniker + ".yaml"

	// Check if workflow file exists
	workflowPath := paths.WorkflowPath(r.repoPath, workflowFile)
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		return fmt.Errorf("workflow file not found: %s\nHint: Create %s/%s/%s", workflowPath, paths.GitHubDir, paths.WorkflowsDir, workflowFile)
	}

	log.Infof("Triggering workflow: %s", workflowFile)

	runID, err := r.ghCLI.TriggerWorkflow(workflowFile, ref)
	if err != nil {
		return err
	}

	log.Infof("Started run %s for %s", runID, moniker)
	log.Info("Waiting for completion...")

	if err := r.ghCLI.WatchRun(runID); err != nil {
		return fmt.Errorf("pipeline failed for %s: %w", moniker, err)
	}

	log.Infof("✅ %s completed successfully", moniker)
	return nil
}

// RunPipelines executes multiple pipelines respecting dependencies.
func (r *PipelineRunner) RunPipelines(monikers []string, ref string) error {
	if len(monikers) == 0 {
		log.Info("No modules specified")
		return nil
	}

	log.Infof("Calculating execution order for: %v", monikers)

	// Calculate execution order
	plan, err := repository.CalculateExecutionOrder(monikers, r.repoPath)
	if err != nil {
		return fmt.Errorf("failed to calculate execution order: %w", err)
	}

	// Filter to only modules with workflow files
	filteredPlan, err := r.filterModulesWithWorkflows(plan)
	if err != nil {
		return err
	}

	if len(filteredPlan.ExecutionOrder) == 0 {
		log.Info("No modules with workflows found")
		return nil
	}

	log.Info("")
	log.Info("Execution plan:")
	for i, layer := range filteredPlan.Layers {
		log.Infof("  Layer %d: %v", i, layer)
	}
	log.Info("")

	// Execute layers sequentially
	return r.executeLayers(filteredPlan, ref)
}

// RunAllPipelines runs all modules in the repository.
func (r *PipelineRunner) RunAllPipelines(ref string) error {
	log.Info("Running all modules in dependency order...")

	// Pass nil to calculate order for all modules
	plan, err := repository.CalculateExecutionOrder(nil, r.repoPath)
	if err != nil {
		return fmt.Errorf("failed to calculate execution order: %w", err)
	}

	// Filter to only modules with workflow files
	filteredPlan, err := r.filterModulesWithWorkflows(plan)
	if err != nil {
		return err
	}

	if len(filteredPlan.ExecutionOrder) == 0 {
		log.Info("No modules with workflows found")
		return nil
	}

	log.Info("")
	log.Info("Execution plan:")
	for i, layer := range filteredPlan.Layers {
		log.Infof("  Layer %d: %v", i, layer)
	}
	log.Info("")

	// Execute layers sequentially
	return r.executeLayers(filteredPlan, ref)
}

// RunAllChangedPipelines detects changed modules and runs their pipelines.
func (r *PipelineRunner) RunAllChangedPipelines(ref string) error {
	log.Info("Detecting changed modules...")

	// Get changed files using git diff
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	cmd.Dir = r.repoPath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get changed files: %w", err)
	}

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(changedFiles) == 1 && changedFiles[0] == "" {
		log.Info("No files changed")
		return nil
	}

	log.Info("Changed files:")
	for _, f := range changedFiles {
		log.Infof("  %s", f)
	}
	log.Info("")

	// Map to modules
	modules, err := repository.GetChangedModules(changedFiles, r.repoPath)
	if err != nil {
		return fmt.Errorf("failed to get changed modules: %w", err)
	}

	if len(modules) == 0 {
		log.Info("No modules changed")
		return nil
	}

	log.Infof("Changed modules: %v", modules)
	log.Info("")

	// Run pipelines for changed modules
	return r.RunPipelines(modules, ref)
}

// executeLayers executes pipeline layers sequentially, with parallel execution within each layer.
func (r *PipelineRunner) executeLayers(plan *repository.ExecutionPlan, ref string) error {
	for layerIdx, layer := range plan.Layers {
		log.Info("================================================")
		log.Infof("Executing Layer %d: %v", layerIdx, layer)
		log.Info("================================================")
		log.Info("")

		// Start all workflows in this layer (parallel)
		runIDs := make(map[string]string) // moniker -> runID
		for _, moniker := range layer {
			workflowFile := moniker + ".yaml"
			log.Infof("Triggering workflow: %s", workflowFile)

			runID, err := r.ghCLI.TriggerWorkflow(workflowFile, ref)
			if err != nil {
				return fmt.Errorf("failed to trigger %s: %w", moniker, err)
			}

			runIDs[moniker] = runID
			log.Infof("  Started %s (run %s)", moniker, runID)
		}

		log.Info("")

		// Wait for all workflows in this layer to complete
		for _, moniker := range layer {
			runID := runIDs[moniker]
			log.Infof("Waiting for %s (run %s)...", moniker, runID)

			if err := r.ghCLI.WatchRun(runID); err != nil {
				return fmt.Errorf("pipeline failed: %s: %w", moniker, err)
			}

			log.Infof("  ✅ %s completed", moniker)
		}

		log.Info("")
		log.Infof("✅ Layer %d completed successfully", layerIdx)
		log.Info("")
	}

	log.Info("================================================")
	log.Info("✅ All pipelines completed successfully!")
	log.Info("================================================")

	return nil
}

// filterModulesWithWorkflows filters the execution plan to only include modules with workflow files.
func (r *PipelineRunner) filterModulesWithWorkflows(plan *repository.ExecutionPlan) (*repository.ExecutionPlan, error) {
	workflowsDir := filepath.Join(r.repoPath, paths.GitHubDir, paths.WorkflowsDir)

	filtered := &repository.ExecutionPlan{
		Layers:         [][]string{},
		ExecutionOrder: []string{},
		LayerCount:     0,
	}

	for _, layer := range plan.Layers {
		filteredLayer := []string{}

		for _, moniker := range layer {
			workflowFile := filepath.Join(workflowsDir, moniker+".yaml")
			if _, err := os.Stat(workflowFile); err == nil {
				filteredLayer = append(filteredLayer, moniker)
				filtered.ExecutionOrder = append(filtered.ExecutionOrder, moniker)
			}
		}

		if len(filteredLayer) > 0 {
			filtered.Layers = append(filtered.Layers, filteredLayer)
		}
	}

	filtered.LayerCount = len(filtered.Layers)

	return filtered, nil
}
