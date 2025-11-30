// Command: get changed modules ci
// Description: Get modules requiring rebuild since last successful CI run
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
//   --pr-base <sha>: For PRs, the base SHA to compare against
//   --workflow <name>: Workflow name to find last success (default: "Change Trigger")
//   --branch <name>: Branch to check for last success (default: main)
// HasSideEffects: false
package get

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	get "github.com/ready-to-release/eac/src/commands/impl/get/internal"
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(GetChangedModulesCI)
}

// CIChangedModulesResult represents the output of the get changed modules ci command
type CIChangedModulesResult struct {
	Modules         []string `json:"modules" yaml:"modules" toml:"modules"`
	DirectlyChanged []string `json:"directly_changed" yaml:"directly_changed" toml:"directly_changed"`
	Invalidated     []string `json:"invalidated" yaml:"invalidated" toml:"invalidated"`
	BaseSHA         string   `json:"base_sha" yaml:"base_sha" toml:"base_sha"`
	HeadSHA         string   `json:"head_sha" yaml:"head_sha" toml:"head_sha"`
	IsBootstrap     bool     `json:"is_bootstrap" yaml:"is_bootstrap" toml:"is_bootstrap"`
	// Additional context for CI reasoning
	ChangedFiles      []string          `json:"changed_files" yaml:"changed_files" toml:"changed_files"`
	ChangedFileCount  int               `json:"changed_file_count" yaml:"changed_file_count" toml:"changed_file_count"`
	FilesByModule     map[string][]string `json:"files_by_module" yaml:"files_by_module" toml:"files_by_module"`
}

func GetChangedModulesCI() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse flags
	prBase := ""
	workflow := "Change Trigger"
	branch := "main"

	for i, arg := range os.Args {
		switch arg {
		case "--pr-base":
			if i+1 < len(os.Args) {
				prBase = os.Args[i+1]
			}
		case "--workflow":
			if i+1 < len(os.Args) {
				workflow = os.Args[i+1]
			}
		case "--branch":
			if i+1 < len(os.Args) {
				branch = os.Args[i+1]
			}
		}
	}

	// Determine base SHA
	baseSHA, isBootstrap, err := determineBaseSHA(prBase, workflow, branch, workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error determining base SHA: %v\n", err)
		return 1
	}

	// Get current HEAD SHA
	headSHA, err := getCurrentSHA(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current SHA: %v\n", err)
		return 1
	}

	// Use the shared get command helper
	return get.ExecuteGetCommand(func() (interface{}, error) {
		// If bootstrap (no previous success), return all modules
		if isBootstrap {
			allModules, err := getAllModuleMonikers(workspaceRoot)
			if err != nil {
				return nil, err
			}
			return CIChangedModulesResult{
				Modules:          allModules,
				DirectlyChanged:  allModules,
				Invalidated:      []string{},
				BaseSHA:          "",
				HeadSHA:          headSHA,
				IsBootstrap:      true,
				ChangedFiles:     []string{},
				ChangedFileCount: 0,
				FilesByModule:    map[string][]string{},
			}, nil
		}

		// Get changed files between base and head
		changedFiles, err := getChangedFilesBetweenSHAs(baseSHA, headSHA, workspaceRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to get changed files: %w", err)
		}

		if len(changedFiles) == 0 {
			return CIChangedModulesResult{
				Modules:          []string{},
				DirectlyChanged:  []string{},
				Invalidated:      []string{},
				BaseSHA:          baseSHA,
				HeadSHA:          headSHA,
				IsBootstrap:      false,
				ChangedFiles:     []string{},
				ChangedFileCount: 0,
				FilesByModule:    map[string][]string{},
			}, nil
		}

		// Get directly changed modules
		directlyChanged, err := repository.GetChangedModules(changedFiles, workspaceRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to get changed modules: %w", err)
		}

		// Get all modules requiring rebuild (includes transitive dependents)
		allRequiringRebuild, err := repository.GetModulesRequiringRebuild(changedFiles, workspaceRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to get modules requiring rebuild: %w", err)
		}

		// Calculate invalidated (all requiring rebuild minus directly changed)
		directlyChangedSet := make(map[string]bool)
		for _, m := range directlyChanged {
			directlyChangedSet[m] = true
		}

		invalidated := []string{}
		for _, m := range allRequiringRebuild {
			if !directlyChangedSet[m] {
				invalidated = append(invalidated, m)
			}
		}

		// Build files-by-module map for detailed reasoning
		filesByModule, err := getFilesByModule(changedFiles, workspaceRoot)
		if err != nil {
			// Non-fatal: just use empty map if we can't build it
			filesByModule = map[string][]string{}
		}

		return CIChangedModulesResult{
			Modules:          allRequiringRebuild,
			DirectlyChanged:  directlyChanged,
			Invalidated:      invalidated,
			BaseSHA:          baseSHA,
			HeadSHA:          headSHA,
			IsBootstrap:      false,
			ChangedFiles:     changedFiles,
			ChangedFileCount: len(changedFiles),
			FilesByModule:    filesByModule,
		}, nil
	})
}

// determineBaseSHA determines the base SHA for comparison
// Returns (baseSHA, isBootstrap, error)
func determineBaseSHA(prBase, workflow, branch, workspaceRoot string) (string, bool, error) {
	// If PR base is provided, use it directly
	if prBase != "" {
		return prBase, false, nil
	}

	// Otherwise, query gh CLI for last successful workflow run
	baseSHA, err := getLastSuccessfulCISHA(workflow, branch, workspaceRoot)
	if err != nil {
		// If no previous successful run, this is a bootstrap
		return "", true, nil
	}

	if baseSHA == "" {
		return "", true, nil
	}

	return baseSHA, false, nil
}

// getLastSuccessfulCISHA queries gh CLI for the last successful workflow run SHA
func getLastSuccessfulCISHA(workflow, branch, workspaceRoot string) (string, error) {
	// gh run list -b <branch> -s success -w "<workflow>" -L 1 --json headSha -q '.[0].headSha'
	cmd := exec.Command("gh", "run", "list",
		"-b", branch,
		"-s", "success",
		"-w", workflow,
		"-L", "1",
		"--json", "headSha",
		"-q", ".[0].headSha",
	)
	cmd.Dir = workspaceRoot

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("gh command failed: %w", err)
	}

	sha := strings.TrimSpace(string(output))
	return sha, nil
}

// getCurrentSHA gets the current HEAD SHA
func getCurrentSHA(workspaceRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = workspaceRoot

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getChangedFilesBetweenSHAs gets the list of files changed between two SHAs
func getChangedFilesBetweenSHAs(baseSHA, headSHA, workspaceRoot string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", baseSHA+".."+headSHA)
	cmd.Dir = workspaceRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}

// getAllModuleMonikers returns all module monikers (for bootstrap case)
func getAllModuleMonikers(workspaceRoot string) ([]string, error) {
	graph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err != nil {
		return nil, err
	}
	return graph.Modules, nil
}

// getFilesByModule maps changed files to their owning modules
func getFilesByModule(changedFiles []string, workspaceRoot string) (map[string][]string, error) {
	registry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)

	for _, filePath := range changedFiles {
		if filePath == "" {
			continue
		}
		matchingModules := registry.FindModulesForFile(filePath)
		if len(matchingModules) == 0 {
			// File doesn't belong to any module - track as "unowned"
			result["(unowned)"] = append(result["(unowned)"], filePath)
		} else {
			for _, module := range matchingModules {
				result[module.Moniker] = append(result[module.Moniker], filePath)
			}
		}
	}

	return result, nil
}
