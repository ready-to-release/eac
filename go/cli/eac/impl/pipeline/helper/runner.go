// Package pipelinerunner provides functionality to execute GitHub workflows
// respecting module dependencies
package pipelinerunner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/gitexec"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

var log = logging.C()

// PipelineRunner orchestrates execution of module pipelines.
type PipelineRunner struct {
	repoPath    string
	ghCLI       GitHubCLI
	wait        bool // Wait for pipeline completion
	timeout     int  // Timeout in seconds (0 = no timeout)
}

// New creates a new PipelineRunner.
func New(repoPath string) *PipelineRunner {
	return &PipelineRunner{
		repoPath: repoPath,
		ghCLI:    NewGitHubCLI(repoPath),
		wait:     true, // Default to waiting
		timeout:  0,    // Default to no timeout
	}
}

// SetWaitOptions configures wait behavior for the pipeline runner.
func (r *PipelineRunner) SetWaitOptions(wait bool, timeout int) {
	r.wait = wait
	r.timeout = timeout
}

// RunPipeline executes a single pipeline.
func (r *PipelineRunner) RunPipeline(moniker, ref string) error {
	// First check if the module exists in the registry
	registry, err := modules.LoadFromWorkspace(r.repoPath)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	if _, exists := registry.Get(moniker); !exists {
		return fmt.Errorf("module not found: %s", moniker)
	}

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

	// Wait for completion if wait flag is set
	if r.wait {
		log.Info("Waiting for completion...")

		var err error
		if r.timeout > 0 {
			err = r.ghCLI.WatchRunWithTimeout(runID, r.timeout)
		} else {
			err = r.ghCLI.WatchRun(runID)
		}

		if err != nil {
			return fmt.Errorf("pipeline failed for %s: %w", moniker, err)
		}

		log.Infof("✅ %s completed successfully", moniker)
	} else {
		log.Info("Pipeline started (not waiting for completion)")
	}

	return nil
}

// RunPipelines executes multiple pipelines respecting dependencies.
func (r *PipelineRunner) RunPipelines(monikers []string, ref string) error {
	if len(monikers) == 0 {
		log.Info("No modules specified")
		return nil
	}

	log.Infof("Calculating execution order for: %v", monikers)

	// Filter to only modules with workflow files
	filteredOrder := r.filterModulesWithWorkflows(monikers)

	if len(filteredOrder) == 0 {
		log.Info("No modules with workflows found")
		return nil
	}

	log.Info("")
	log.Info("Execution order:")
	for i, moniker := range filteredOrder {
		log.Infof("  %d. %s", i+1, moniker)
	}
	log.Info("")

	// Execute modules sequentially in dependency order
	return r.executeSequential(filteredOrder, ref)
}

// RunAllPipelines runs all modules in the repository.
func (r *PipelineRunner) RunAllPipelines(ref string) error {
	log.Info("Running all modules...")

	// Load all modules from registry
	registry, err := modules.LoadFromWorkspace(r.repoPath)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	var allMonikers []string
	for _, mod := range registry.All() {
		allMonikers = append(allMonikers, mod.Moniker)
	}

	// Filter to only modules with workflow files
	filteredOrder := r.filterModulesWithWorkflows(allMonikers)

	if len(filteredOrder) == 0 {
		log.Info("No modules with workflows found")
		return nil
	}

	log.Info("")
	log.Info("Execution order:")
	for i, moniker := range filteredOrder {
		log.Infof("  %d. %s", i+1, moniker)
	}
	log.Info("")

	// Execute modules sequentially in dependency order
	return r.executeSequential(filteredOrder, ref)
}

// RunAllChangedPipelines detects changed modules and runs their pipelines.
func (r *PipelineRunner) RunAllChangedPipelines(ref string) error {
	log.Info("Detecting changed modules...")

	// Determine what to compare against
	// If ref is provided and not the current branch, compare against that ref
	// Otherwise, detect uncommitted changes (diff HEAD)
	var output []byte
	var err error
	if ref != "" && ref != getCurrentBranch(r.repoPath) {
		// Compare against specific ref
		log.Infof("Comparing against ref: %s", ref)
		output, err = gitexec.Run(r.repoPath, "diff", "--name-only", ref)
	} else {
		// Detect uncommitted changes
		log.Info("Detecting uncommitted changes...")
		output, err = gitexec.Run(r.repoPath, "diff", "--name-only", "HEAD")
	}

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

// getCurrentBranch gets the current git branch name.
func getCurrentBranch(repoPath string) string {
	output, err := gitexec.Run(repoPath, "branch", "--show-current")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// executeSequential executes pipeline modules sequentially in dependency order.
// Each module waits for completion before the next starts (if wait is enabled).
func (r *PipelineRunner) executeSequential(order []string, ref string) error {
	for idx, moniker := range order {
		log.Info("================================================")
		log.Infof("Executing [%d/%d]: %s", idx+1, len(order), moniker)
		log.Info("================================================")
		log.Info("")

		workflowFile := moniker + ".yaml"
		log.Infof("Triggering workflow: %s", workflowFile)

		runID, err := r.ghCLI.TriggerWorkflow(workflowFile, ref)
		if err != nil {
			return fmt.Errorf("failed to trigger %s: %w", moniker, err)
		}

		log.Infof("  Started %s (run %s)", moniker, runID)

		// Wait for workflow to complete (if wait is enabled)
		if r.wait {
			log.Infof("Waiting for %s (run %s)...", moniker, runID)

			var err error
			if r.timeout > 0 {
				err = r.ghCLI.WatchRunWithTimeout(runID, r.timeout)
			} else {
				err = r.ghCLI.WatchRun(runID)
			}

			if err != nil {
				return fmt.Errorf("pipeline failed: %s: %w", moniker, err)
			}

			log.Infof("  ✅ %s completed", moniker)
		} else {
			log.Info("Pipeline started (not waiting for completion)")
		}

		log.Info("")
	}

	log.Info("================================================")
	log.Info("✅ All pipelines completed successfully!")
	log.Info("================================================")

	return nil
}

// filterModulesWithWorkflows filters the execution order to only include modules with workflow files.
func (r *PipelineRunner) filterModulesWithWorkflows(order []string) []string {
	workflowsDir := filepath.Join(r.repoPath, paths.GitHubDir, paths.WorkflowsDir)

	var filtered []string
	for _, moniker := range order {
		workflowFile := filepath.Join(workflowsDir, moniker+".yaml")
		if _, err := os.Stat(workflowFile); err == nil {
			filtered = append(filtered, moniker)
		}
	}

	return filtered
}
