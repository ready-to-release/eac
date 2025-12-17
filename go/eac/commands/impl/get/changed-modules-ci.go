// Command: get changed-modules-ci
// Short: Get modules requiring rebuild since last successful CI run
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
//   --format shell: Output as shell variable assignments for eval
//   --pr-base <sha>: For PRs, the base SHA to compare against
//   --workflow <name>: Workflow name to find last success (default: "CI Trigger")
//   --branch <name>: Branch to check for last success (default: main)
//   --filter-workflows: Only include modules that have a ci-{module}.yaml workflow file
// Long:
// Long: Expected Output:
// Long: YAML list of modules needing rebuild based on CI state, including:
// Long:   - All modules requiring rebuild (directly changed + transitive dependents)
// Long:   - Directly changed modules (files modified since base SHA)
// Long:   - Invalidated modules (transitive dependents requiring rebuild)
// Long:   - Base and head SHAs, bootstrap flag, changed files list
// Long:   - Files-by-module mapping for detailed change reasoning
// Long:
// Long: With --format shell, outputs shell variable assignments:
// Long:   MODULES="mod1 mod2 mod3"
// Long:   DIRECTLY_CHANGED="mod1 mod2"
// Long:   INVALIDATED="mod3"
// Long:   BASE_SHA="abc123"
// Long:   IS_BOOTSTRAP="false"
// Long:   CHANGED_FILE_COUNT="5"
package get

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetChangedModulesCI)
}

// CIChangedModulesResult represents the output of the get changed-modules-ci command
type CIChangedModulesResult struct {
	Modules         []string `json:"modules" yaml:"modules" toml:"modules"`
	DirectlyChanged []string `json:"directly_changed" yaml:"directly_changed" toml:"directly_changed"`
	Invalidated     []string `json:"invalidated" yaml:"invalidated" toml:"invalidated"`
	BaseSHA         string   `json:"base_sha" yaml:"base_sha" toml:"base_sha"`
	HeadSHA         string   `json:"head_sha" yaml:"head_sha" toml:"head_sha"`
	IsBootstrap     bool     `json:"is_bootstrap" yaml:"is_bootstrap" toml:"is_bootstrap"`
	// Additional context for CI reasoning
	ChangedFiles      []string            `json:"changed_files" yaml:"changed_files" toml:"changed_files"`
	ChangedFileCount  int                 `json:"changed_file_count" yaml:"changed_file_count" toml:"changed_file_count"`
	FilesByModule     map[string][]string `json:"files_by_module" yaml:"files_by_module" toml:"files_by_module"`
	// Workflow filtering (only present when --filter-workflows is used)
	FilteredOut []string `json:"filtered_out,omitempty" yaml:"filtered_out,omitempty" toml:"filtered_out,omitempty"`
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
	workflow := "CI Trigger"
	branch := "main"
	filterWorkflows := false
	format := ""

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
		case "--filter-workflows":
			filterWorkflows = true
		case "--format":
			if i+1 < len(os.Args) {
				format = os.Args[i+1]
			}
		}
	}

	// Determine base SHA
	baseSHA, isBootstrap, err := determineBaseSHA(prBase, workflow, branch, workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: determining base SHA: %v\n", err)
		return 1
	}

	// Get current HEAD SHA using shared detection logic
	shaResult, err := DetectCurrentSHA(workspaceRoot, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: getting current SHA: %v\n", err)
		return 1
	}
	headSHA := shaResult.SHA

	// Build the result
	result, err := buildCIChangedModulesResult(workspaceRoot, baseSHA, headSHA, isBootstrap, filterWorkflows)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Handle shell format output
	if format == "shell" {
		fmt.Printf("MODULES=\"%s\"\n", strings.Join(result.Modules, " "))
		fmt.Printf("DIRECTLY_CHANGED=\"%s\"\n", strings.Join(result.DirectlyChanged, " "))
		fmt.Printf("INVALIDATED=\"%s\"\n", strings.Join(result.Invalidated, " "))
		fmt.Printf("BASE_SHA=\"%s\"\n", result.BaseSHA)
		fmt.Printf("IS_BOOTSTRAP=\"%t\"\n", result.IsBootstrap)
		fmt.Printf("CHANGED_FILE_COUNT=\"%d\"\n", result.ChangedFileCount)
		// Include FILES_BY_MODULE as JSON for detailed reporting
		if filesJSON, err := json.Marshal(result.FilesByModule); err == nil {
			fmt.Printf("FILES_BY_MODULE='%s'\n", string(filesJSON))
		}
		return 0
	}

	// Use the shared get command helper for YAML/JSON/TOML output
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return result, nil
	})
}

// buildCIChangedModulesResult builds the result structure
func buildCIChangedModulesResult(workspaceRoot, baseSHA, headSHA string, isBootstrap, filterWorkflows bool) (*CIChangedModulesResult, error) {
	// If bootstrap (no previous success), return all modules
	if isBootstrap {
		allModules, err := getAllModuleMonikers(workspaceRoot)
		if err != nil {
			return nil, err
		}

		// Apply workflow filter if requested
		var filteredOut []string
		if filterWorkflows {
			allModules, filteredOut = filterModulesWithWorkflows(allModules, workspaceRoot)
		}

		return &CIChangedModulesResult{
			Modules:          allModules,
			DirectlyChanged:  allModules,
			Invalidated:      []string{},
			BaseSHA:          "",
			HeadSHA:          headSHA,
			IsBootstrap:      true,
			ChangedFiles:     []string{},
			ChangedFileCount: 0,
			FilesByModule:    map[string][]string{},
			FilteredOut:      filteredOut,
		}, nil
	}

	// Get changed files between base and head
	changedFiles, err := getChangedFilesBetweenSHAs(baseSHA, headSHA, workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	// Filter out files that are owned by modules but shouldn't trigger CI:
	// - Release workflows (files.workflows.release): only affect release process
	// - Changelogs (files.changelog): only affect release documentation
	// - README.md files: documentation only
	ciExcludedFiles := getCIExcludedFiles(workspaceRoot)
	changedFiles = filterOutCIExcludedFiles(changedFiles, ciExcludedFiles)

	if len(changedFiles) == 0 {
		return &CIChangedModulesResult{
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

	// Apply workflow filter if requested
	var filteredOut []string
	if filterWorkflows {
		allRequiringRebuild, filteredOut = filterModulesWithWorkflows(allRequiringRebuild, workspaceRoot)
		directlyChanged, _ = filterModulesWithWorkflows(directlyChanged, workspaceRoot)
		invalidated, _ = filterModulesWithWorkflows(invalidated, workspaceRoot)
	}

	return &CIChangedModulesResult{
		Modules:          allRequiringRebuild,
		DirectlyChanged:  directlyChanged,
		Invalidated:      invalidated,
		BaseSHA:          baseSHA,
		HeadSHA:          headSHA,
		IsBootstrap:      false,
		ChangedFiles:     changedFiles,
		ChangedFileCount: len(changedFiles),
		FilesByModule:    filesByModule,
		FilteredOut:      filteredOut,
	}, nil
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

// filterModulesWithWorkflows filters modules to only those that have a ci-{module}.yaml workflow file
// Returns (filtered modules, modules that were filtered out)
func filterModulesWithWorkflows(monikers []string, workspaceRoot string) ([]string, []string) {
	filtered := []string{}
	filteredOut := []string{}

	for _, moniker := range monikers {
		workflowPath := filepath.Join(workspaceRoot, ".github", "workflows", fmt.Sprintf("ci-%s.yaml", moniker))
		if _, err := os.Stat(workflowPath); err == nil {
			filtered = append(filtered, moniker)
		} else {
			filteredOut = append(filteredOut, moniker)
		}
	}

	return filtered, filteredOut
}

// getCIExcludedFiles returns a set of file paths that should not trigger CI.
// Uses files.ignore patterns from module contracts - files that are owned but
// changes to them don't affect module functionality.
func getCIExcludedFiles(workspaceRoot string) map[string]bool {
	result := make(map[string]bool)

	registry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return result // Return empty on error - non-fatal
	}

	for _, module := range registry.All() {
		// Use files.ignore patterns from module contract
		for _, pattern := range module.Files.Ignore {
			resolvedPattern := resolveIgnorePattern(pattern, module.Files.Root)
			result[resolvedPattern] = true
		}
	}

	return result
}

// resolveIgnorePattern resolves an ignore pattern to an absolute repo path.
// Patterns starting with .github/, release/, or containing .. are resolved relative to module root.
// Other patterns are joined with the module root.
func resolveIgnorePattern(pattern, moduleRoot string) string {
	// Normalize pattern to forward slashes
	pattern = strings.ReplaceAll(pattern, "\\", "/")

	// If pattern contains .., resolve it relative to module root using filepath.Clean
	if strings.Contains(pattern, "..") && moduleRoot != "" && moduleRoot != "/" {
		resolved := filepath.Join(moduleRoot, pattern)
		resolved = filepath.Clean(resolved)
		return strings.ReplaceAll(resolved, "\\", "/")
	}

	// If pattern starts with known repo-root prefixes, use as-is
	if strings.HasPrefix(pattern, ".github/") || strings.HasPrefix(pattern, "release/") ||
		strings.HasPrefix(pattern, "contracts/") || strings.HasPrefix(pattern, "go/") {
		return pattern
	}

	// Otherwise, resolve relative to module root
	if moduleRoot != "" && moduleRoot != "/" {
		resolved := filepath.Join(moduleRoot, pattern)
		return strings.ReplaceAll(resolved, "\\", "/")
	}

	return pattern
}

// filterOutCIExcludedFiles removes files that shouldn't trigger CI from the changed files list.
// Supports both exact matches and glob patterns from files.ignore.
func filterOutCIExcludedFiles(files []string, excludedPatterns map[string]bool) []string {
	result := make([]string, 0, len(files))
	for _, f := range files {
		if isFileExcluded(f, excludedPatterns) {
			continue
		}
		result = append(result, f)
	}
	return result
}

// isFileExcluded checks if a file matches any of the excluded patterns.
// Supports exact matches and glob patterns (*, **, ?).
func isFileExcluded(filePath string, patterns map[string]bool) bool {
	for pattern := range patterns {
		// Try exact match first
		if filePath == pattern {
			return true
		}
		// Try glob match for patterns containing wildcards
		if strings.ContainsAny(pattern, "*?") {
			if matched, _ := filepath.Match(pattern, filePath); matched {
				return true
			}
			// Handle ** patterns by checking if file is under the pattern's directory
			if strings.Contains(pattern, "**") {
				// Convert ** pattern to prefix match
				prefix := strings.Split(pattern, "**")[0]
				if strings.HasPrefix(filePath, prefix) {
					return true
				}
			}
		}
	}
	return false
}
