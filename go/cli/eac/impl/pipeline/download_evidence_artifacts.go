package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/get"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/ghexec"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/repository"
)

type pipelineDownloadEvidenceArtifactsCommand struct{}

var _ core.SimpleCommandPort = (*pipelineDownloadEvidenceArtifactsCommand)(nil)

func (c *pipelineDownloadEvidenceArtifactsCommand) Name() string {
	return "pipeline download-evidence-artifacts"
}

func (c *pipelineDownloadEvidenceArtifactsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "pipeline-download-evidence-artifacts",
		Short:         "Download test/scan artifacts for evidence building",
		Long:          "Downloads test results and scan artifacts from CI runs for a module and all its dependencies.\nUses the `get evidence-ci-runs` command to determine which CI runs to download from.\n\nFor each module in the dependency chain that has a ci-{module}.yaml workflow:\n  - Downloads test-results-{module}* artifacts to out/test/\n  - Downloads scan-results-{module} artifacts to out/scan/\n\nFails if any dependency with a CI workflow has no successful CI run.\n\nExample:\n  pipeline download-evidence-artifacts clie\n  pipeline download-evidence-artifacts eac-ext --output-dir custom/path",
		Args:          "module (required) - Module moniker to download evidence artifacts for",
		Flags: []core.FlagSpec{
			{Name: "output-dir", Type: "string", DefaultValue: "out", Usage: "Base output directory (test/ and scan/ subdirs will be created)"},
		},
	}
}

func (c *pipelineDownloadEvidenceArtifactsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return PipelineDownloadEvidenceArtifacts()
}

// DownloadResult tracks what was downloaded.
type DownloadResult struct {
	Module        string   `json:"module" yaml:"module"`
	RunID         int      `json:"run_id" yaml:"run_id"`
	TestArtifacts []string `json:"test_artifacts,omitempty" yaml:"test_artifacts,omitempty"`
	ScanArtifacts []string `json:"scan_artifacts,omitempty" yaml:"scan_artifacts,omitempty"`
	Error         string   `json:"error,omitempty" yaml:"error,omitempty"`
}

func PipelineDownloadEvidenceArtifacts() int {
	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "pipeline", and "download-evidence-artifacts"

	// Parse args
	var moniker string
	outputDir := "out"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--output-dir" && i+1 < len(args):
			outputDir = args[i+1]
			i++
		case strings.HasPrefix(arg, "--output-dir="):
			outputDir = strings.TrimPrefix(arg, "--output-dir=")
		case !strings.HasPrefix(arg, "--") && moniker == "":
			moniker = arg
		}
	}

	if moniker == "" {
		fmt.Fprintln(os.Stderr, "Usage: pipeline download-evidence-artifacts <module> [--output-dir DIR]")
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Create output directories
	testDir := filepath.Join(outputDir, "test")
	scanDir := filepath.Join(outputDir, "scan")

	if err := os.MkdirAll(testDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating test directory: %v\n", err)
		return 1
	}
	if err := os.MkdirAll(scanDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating scan directory: %v\n", err)
		return 1
	}

	// Get evidence CI runs
	fmt.Printf("=== Downloading evidence artifacts for: %s ===\n\n", moniker)

	ciRuns, skipped, err := getEvidenceCIRunsInternal(moniker, workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if len(skipped) > 0 {
		fmt.Printf("Skipped (no CI workflow): %s\n\n", strings.Join(skipped, ", "))
	}

	// Download artifacts from each CI run
	var results []DownloadResult
	hasErrors := false

	for _, run := range ciRuns {
		fmt.Printf("--- %s (run %d) ---\n", run.Module, run.RunID)

		result := DownloadResult{
			Module: run.Module,
			RunID:  run.RunID,
		}

		// Download test artifacts
		testArtifacts, err := downloadArtifacts(run.RunID, fmt.Sprintf("test-results-%s*", run.Module), testDir, workspaceRoot)
		if err != nil {
			fmt.Printf("  Warning: test artifacts: %v\n", err)
		} else if len(testArtifacts) > 0 {
			result.TestArtifacts = testArtifacts
			fmt.Printf("  Downloaded %d test artifact(s)\n", len(testArtifacts))
		} else {
			fmt.Printf("  No test artifacts found\n")
		}

		// Download scan artifacts
		scanArtifacts, err := downloadArtifacts(run.RunID, fmt.Sprintf("scan-results-%s", run.Module), scanDir, workspaceRoot)
		if err != nil {
			fmt.Printf("  Warning: scan artifacts: %v\n", err)
		} else if len(scanArtifacts) > 0 {
			result.ScanArtifacts = scanArtifacts
			fmt.Printf("  Downloaded %d scan artifact(s)\n", len(scanArtifacts))
		} else {
			fmt.Printf("  No scan artifacts found\n")
		}

		results = append(results, result)
		fmt.Println()
	}

	// Summary
	fmt.Println("=== Download Summary ===")
	totalTests := 0
	totalScans := 0
	for _, r := range results {
		totalTests += len(r.TestArtifacts)
		totalScans += len(r.ScanArtifacts)
	}
	fmt.Printf("Modules processed: %d\n", len(results))
	fmt.Printf("Test artifacts: %d\n", totalTests)
	fmt.Printf("Scan artifacts: %d\n", totalScans)

	if hasErrors {
		return 1
	}
	return 0
}

// getEvidenceCIRunsInternal gets CI runs for a module and its dependencies
// Returns (ciRuns, skippedModules, error).
func getEvidenceCIRunsInternal(moniker, workspaceRoot string) ([]get.EvidenceCIRun, []string, error) {
	// Load module registry
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load module registry: %w", err)
	}

	// Check module exists
	_, exists := moduleRegistry.Get(moniker)
	if !exists {
		return nil, nil, fmt.Errorf("module not found: %s", moniker)
	}

	// Get transitive dependencies
	deps := getTransitiveDeps(moniker, moduleRegistry)

	// Create GitHub API client
	api := github.NewGHClient(ghexec.New(workspaceRoot), workspaceRoot)

	var ciRuns []get.EvidenceCIRun
	var skipped []string
	var missingCI []string

	for _, dep := range deps {
		workflowName := fmt.Sprintf("ci-%s.yaml", dep)
		workflowPath := filepath.Join(workspaceRoot, ".github", "workflows", workflowName)

		// Skip modules without CI workflows
		if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
			skipped = append(skipped, dep)
			continue
		}

		// Query for last successful CI run
		runs, err := api.ListRuns(workflowName, github.ListRunsOpts{
			Status: "success",
			Limit:  1,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to query CI runs for %s: %w", dep, err)
		}

		if len(runs) == 0 {
			missingCI = append(missingCI, dep)
			continue
		}

		ciRuns = append(ciRuns, get.EvidenceCIRun{
			Module:   dep,
			Workflow: workflowName,
			RunID:    runs[0].ID,
		})
	}

	if len(missingCI) > 0 {
		return nil, nil, fmt.Errorf("no successful CI run found for module(s): %s", strings.Join(missingCI, ", "))
	}

	return ciRuns, skipped, nil
}

// getTransitiveDeps returns a module and all its transitive dependencies.
func getTransitiveDeps(moniker string, reg *modules.Registry) []string {
	seen := make(map[string]bool)
	var result []string

	var collect func(m string)
	collect = func(m string) {
		if seen[m] {
			return
		}
		seen[m] = true

		module, exists := reg.Get(m)
		if !exists {
			return
		}

		result = append(result, m)

		for _, dep := range module.DependsOn {
			collect(dep)
		}
	}

	collect(moniker)
	return result
}

// downloadArtifacts downloads artifacts matching a pattern from a workflow run
// Returns the list of downloaded artifact names.
func downloadArtifacts(runID int, pattern, destDir, workspaceRoot string) ([]string, error) {
	// Use gh run download with pattern matching
	output, exitCode, err := ghexec.RunCombined(context.Background(), workspaceRoot, "run", "download",
		fmt.Sprintf("%d", runID),
		"--pattern", pattern,
		"--dir", destDir,
	)
	if err != nil {
		return nil, fmt.Errorf("gh execution failed: %w", err)
	}
	if exitCode != 0 {
		// Check if it's just "no artifacts found" (exit code 1 with specific message)
		outputStr := string(output)
		if strings.Contains(outputStr, "no artifact") || strings.Contains(outputStr, "no matching") {
			return nil, nil // No artifacts is not an error
		}
		return nil, fmt.Errorf("%s", strings.TrimSpace(outputStr))
	}

	// Parse output to get downloaded artifact names
	// gh run download outputs lines like "Downloading artifact-name..."
	var downloaded []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Downloading ") {
			name := strings.TrimSuffix(strings.TrimPrefix(line, "Downloading "), "...")
			if name != "" {
				downloaded = append(downloaded, name)
			}
		}
	}

	// If we got no output but no error, assume success
	if len(downloaded) == 0 && err == nil {
		// Try to detect what was downloaded by listing the directory
		entries, readErr := os.ReadDir(destDir)
		if readErr == nil {
			for _, e := range entries {
				if e.IsDir() && strings.Contains(e.Name(), strings.TrimSuffix(strings.TrimSuffix(pattern, "*"), "-")) {
					downloaded = append(downloaded, e.Name())
				}
			}
		}
	}

	// Flatten directory structure: gh run download creates artifact-name/module/
	// but evidence loader expects just module/ directly under destDir
	if err := flattenArtifactDirs(destDir); err != nil {
		return downloaded, fmt.Errorf("failed to flatten directories: %w", err)
	}

	return downloaded, nil
}

// flattenArtifactDirs flattens the artifact directory structure.
// gh run download creates: destDir/artifact-name/module/files
// We need: destDir/module/files.
func flattenArtifactDirs(destDir string) error {
	entries, err := os.ReadDir(destDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		artifactDir := filepath.Join(destDir, entry.Name())

		// Check if this looks like an artifact directory (test-results-* or scan-results-*)
		if !strings.HasPrefix(entry.Name(), "test-results-") && !strings.HasPrefix(entry.Name(), "scan-results-") {
			continue
		}

		// Move contents up one level
		subEntries, err := os.ReadDir(artifactDir)
		if err != nil {
			continue
		}

		for _, subEntry := range subEntries {
			srcPath := filepath.Join(artifactDir, subEntry.Name())
			dstPath := filepath.Join(destDir, subEntry.Name())

			// If destination exists, merge directories
			if _, err := os.Stat(dstPath); err == nil {
				if subEntry.IsDir() {
					// Merge directory contents
					if err := mergeDirectories(srcPath, dstPath); err != nil {
						return fmt.Errorf("failed to merge %s: %w", subEntry.Name(), err)
					}
					os.RemoveAll(srcPath)
				}
				// Skip files that already exist
				continue
			}

			// Move to destination
			if err := os.Rename(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to move %s: %w", subEntry.Name(), err)
			}
		}

		// Remove empty artifact directory
		os.Remove(artifactDir)
	}

	return nil
}

// mergeDirectories recursively merges src into dst.
func mergeDirectories(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Create destination dir if needed
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return err
			}
			// Recurse
			if err := mergeDirectories(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file if it doesn't exist
			if _, err := os.Stat(dstPath); os.IsNotExist(err) {
				if err := os.Rename(srcPath, dstPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
