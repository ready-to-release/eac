// Package pipelinerunner provides functionality to execute GitHub workflows
// respecting module dependencies
package pipelinerunner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/internal/repository"
)

// PipelineRunner orchestrates execution of module pipelines
type PipelineRunner struct {
	repoPath string
	version  string
	ghCLI    GitHubCLI
}

// New creates a new PipelineRunner
func New(repoPath string, version string) *PipelineRunner {
	return &PipelineRunner{
		repoPath: repoPath,
		version:  version,
		ghCLI:    NewGitHubCLI(repoPath),
	}
}

// RunPipeline executes a single pipeline
func (r *PipelineRunner) RunPipeline(moniker string, ref string) error {
	workflowFile := moniker + ".yaml"

	// Check if workflow file exists
	workflowPath := filepath.Join(r.repoPath, ".github", "workflows", workflowFile)
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		return fmt.Errorf("workflow file not found: %s\nHint: Create .github/workflows/%s", workflowPath, workflowFile)
	}

	fmt.Printf("Triggering workflow: %s\n", workflowFile)

	runID, err := r.ghCLI.TriggerWorkflow(workflowFile, ref)
	if err != nil {
		return err
	}

	fmt.Printf("Started run %s for %s\n", runID, moniker)
	fmt.Printf("Waiting for completion...\n")

	if err := r.ghCLI.WatchRun(runID); err != nil {
		return fmt.Errorf("pipeline failed for %s: %w", moniker, err)
	}

	fmt.Printf("✅ %s completed successfully\n", moniker)
	return nil
}

// RunPipelines executes multiple pipelines respecting dependencies
func (r *PipelineRunner) RunPipelines(monikers []string, ref string) error {
	if len(monikers) == 0 {
		fmt.Println("No modules specified")
		return nil
	}

	fmt.Printf("Calculating execution order for: %v\n", monikers)

	// Calculate execution order
	plan, err := repository.CalculateExecutionOrder(monikers, r.repoPath, r.version)
	if err != nil {
		return fmt.Errorf("failed to calculate execution order: %w", err)
	}

	// Filter to only modules with workflow files
	filteredPlan, err := r.filterModulesWithWorkflows(plan)
	if err != nil {
		return err
	}

	if len(filteredPlan.ExecutionOrder) == 0 {
		fmt.Println("No modules with workflows found")
		return nil
	}

	fmt.Printf("\nExecution plan:\n")
	for i, layer := range filteredPlan.Layers {
		fmt.Printf("  Layer %d: %v\n", i, layer)
	}
	fmt.Println()

	// Execute layers sequentially
	return r.executeLayers(filteredPlan, ref)
}

// RunAllPipelines runs all modules in the repository
func (r *PipelineRunner) RunAllPipelines(ref string) error {
	fmt.Println("Running all modules in dependency order...")

	// Pass nil to calculate order for all modules
	plan, err := repository.CalculateExecutionOrder(nil, r.repoPath, r.version)
	if err != nil {
		return fmt.Errorf("failed to calculate execution order: %w", err)
	}

	// Filter to only modules with workflow files
	filteredPlan, err := r.filterModulesWithWorkflows(plan)
	if err != nil {
		return err
	}

	if len(filteredPlan.ExecutionOrder) == 0 {
		fmt.Println("No modules with workflows found")
		return nil
	}

	fmt.Printf("\nExecution plan:\n")
	for i, layer := range filteredPlan.Layers {
		fmt.Printf("  Layer %d: %v\n", i, layer)
	}
	fmt.Println()

	// Execute layers sequentially
	return r.executeLayers(filteredPlan, ref)
}

// RunAllChangedPipelines detects changed modules and runs their pipelines
func (r *PipelineRunner) RunAllChangedPipelines(ref string) error {
	fmt.Println("Detecting changed modules...")

	// Get changed files using git diff
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	cmd.Dir = r.repoPath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get changed files: %w", err)
	}

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(changedFiles) == 1 && changedFiles[0] == "" {
		fmt.Println("No files changed")
		return nil
	}

	fmt.Printf("Changed files:\n")
	for _, f := range changedFiles {
		fmt.Printf("  %s\n", f)
	}
	fmt.Println()

	// Map to modules
	modules, err := repository.GetChangedModules(changedFiles, r.repoPath, r.version)
	if err != nil {
		return fmt.Errorf("failed to get changed modules: %w", err)
	}

	if len(modules) == 0 {
		fmt.Println("No modules changed")
		return nil
	}

	fmt.Printf("Changed modules: %v\n\n", modules)

	// Run pipelines for changed modules
	return r.RunPipelines(modules, ref)
}

// executeLayers executes pipeline layers sequentially, with parallel execution within each layer
func (r *PipelineRunner) executeLayers(plan *repository.ExecutionPlan, ref string) error {
	for layerIdx, layer := range plan.Layers {
		fmt.Printf("================================================\n")
		fmt.Printf("Executing Layer %d: %v\n", layerIdx, layer)
		fmt.Printf("================================================\n\n")

		// Start all workflows in this layer (parallel)
		runIDs := make(map[string]string) // moniker -> runID
		for _, moniker := range layer {
			workflowFile := moniker + ".yaml"
			fmt.Printf("Triggering workflow: %s\n", workflowFile)

			runID, err := r.ghCLI.TriggerWorkflow(workflowFile, ref)
			if err != nil {
				return fmt.Errorf("failed to trigger %s: %w", moniker, err)
			}

			runIDs[moniker] = runID
			fmt.Printf("  Started %s (run %s)\n", moniker, runID)
		}

		fmt.Println()

		// Wait for all workflows in this layer to complete
		for _, moniker := range layer {
			runID := runIDs[moniker]
			fmt.Printf("Waiting for %s (run %s)...\n", moniker, runID)

			if err := r.ghCLI.WatchRun(runID); err != nil {
				return fmt.Errorf("pipeline failed: %s: %w", moniker, err)
			}

			fmt.Printf("  ✅ %s completed\n", moniker)
		}

		fmt.Printf("\n✅ Layer %d completed successfully\n\n", layerIdx)
	}

	fmt.Println("================================================")
	fmt.Println("✅ All pipelines completed successfully!")
	fmt.Println("================================================")

	return nil
}

// filterModulesWithWorkflows filters the execution plan to only include modules with workflow files
func (r *PipelineRunner) filterModulesWithWorkflows(plan *repository.ExecutionPlan) (*repository.ExecutionPlan, error) {
	workflowsDir := filepath.Join(r.repoPath, ".github", "workflows")

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
