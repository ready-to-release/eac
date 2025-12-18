// Command: get changed-modules-ci
// Short: Get modules requiring rebuild since last successful CI run
// Flag.pr-base: type=string, usage=For PRs, the base SHA to compare against
// Flag.filter-workflows: type=bool, usage=Only include modules that have a ci-{module}.yaml workflow file
// Flag.format: type=string, usage=Output format (shell outputs shell variables; otherwise uses standard get command formats)
// Long:
// Long: Expected Output:
// Long: YAML list of modules needing rebuild based on per-module CI state.
// Long:
// Long: Per-module change detection:
// Long:   For each module with a ci-{module}.yaml workflow:
// Long:   1. Query the module's last successful CI run
// Long:   2. If CI passed at current HEAD SHA → skip (no rebuild needed)
// Long:   3. If CI passed at different SHA → check if module's files changed
// Long:   4. If no CI history → module needs CI (but NOT bootstrap for all)
// Long:
// Long: Output includes:
// Long:   - All modules requiring rebuild (directly changed + transitive dependents)
// Long:   - Directly changed modules (files modified since module's last CI success)
// Long:   - Invalidated modules (transitive dependents requiring rebuild)
// Long:   - Per-module CI status reasoning
// Long:
// Long: With --format shell, outputs shell variable assignments:
// Long:   MODULES="mod1 mod2 mod3"
// Long:   DIRECTLY_CHANGED="mod1 mod2"
// Long:   INVALIDATED="mod3"
// Long:   BASE_SHA="per-module"
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
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/github"
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
	// Per-module CI status (why each module needs/doesn't need CI)
	ModuleStatus map[string]ModuleCIStatus `json:"module_status,omitempty" yaml:"module_status,omitempty" toml:"module_status,omitempty"`
	// Modules that were skipped (have valid CI at HEAD)
	Skipped []string `json:"skipped,omitempty" yaml:"skipped,omitempty" toml:"skipped,omitempty"`
	// Workflow filtering (only present when --filter-workflows is used)
	FilteredOut []string `json:"filtered_out,omitempty" yaml:"filtered_out,omitempty" toml:"filtered_out,omitempty"`
}

// ModuleCIStatus tracks the CI status for a single module
type ModuleCIStatus struct {
	HasValidCI     bool   `json:"has_valid_ci" yaml:"has_valid_ci" toml:"has_valid_ci"`
	LastSuccessSHA string `json:"last_success_sha,omitempty" yaml:"last_success_sha,omitempty" toml:"last_success_sha,omitempty"`
	Reason         string `json:"reason" yaml:"reason" toml:"reason"`
	FilesChanged   int    `json:"files_changed,omitempty" yaml:"files_changed,omitempty" toml:"files_changed,omitempty"`
}

func GetChangedModulesCI() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse flags
	prBase := ""
	filterWorkflows := false
	format := ""

	for i, arg := range os.Args {
		switch arg {
		case "--pr-base":
			if i+1 < len(os.Args) {
				prBase = os.Args[i+1]
			}
		case "--filter-workflows":
			filterWorkflows = true
		case "--format":
			if i+1 < len(os.Args) {
				format = os.Args[i+1]
			}
		}
	}

	// Get current HEAD SHA using shared detection logic
	shaResult, err := DetectCurrentSHA(workspaceRoot, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: getting current SHA: %v\n", err)
		return 1
	}
	headSHA := shaResult.SHA

	// Build the result using per-module change detection
	result, err := buildPerModuleCIResult(workspaceRoot, headSHA, prBase, filterWorkflows)
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
		// Include MODULE_STATUS as JSON for trigger reasons
		if statusJSON, err := json.Marshal(result.ModuleStatus); err == nil {
			fmt.Printf("MODULE_STATUS='%s'\n", string(statusJSON))
		}
		return 0
	}

	// Use the shared get command helper for YAML/JSON/TOML output
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return result, nil
	})
}

// buildPerModuleCIResult implements per-module change detection.
// Instead of using a single repo-wide base SHA, it checks each module's CI workflow independently.
func buildPerModuleCIResult(workspaceRoot, headSHA, prBase string, filterWorkflows bool) (*CIChangedModulesResult, error) {
	// Get all modules with CI workflows
	allModules, err := getAllModuleMonikers(workspaceRoot)
	if err != nil {
		return nil, err
	}

	// Filter to only modules with CI workflows
	modulesWithCI, filteredOut := filterModulesWithWorkflows(allModules, workspaceRoot)

	// Create GitHub API client
	api := github.NewGHClient(workspaceRoot)

	// Track results
	result := &CIChangedModulesResult{
		Modules:          []string{},
		DirectlyChanged:  []string{},
		Invalidated:      []string{},
		BaseSHA:          "per-module", // Indicate per-module mode
		HeadSHA:          headSHA,
		IsBootstrap:      false,
		ChangedFiles:     []string{},
		ChangedFileCount: 0,
		FilesByModule:    map[string][]string{},
		ModuleStatus:     map[string]ModuleCIStatus{},
		Skipped:          []string{},
		FilteredOut:      filteredOut,
	}

	// CI-excluded files (changelogs, release workflows, etc.)
	ciExcludedFiles := getCIExcludedFiles(workspaceRoot)

	// Track all changed files across modules for aggregate reporting
	allChangedFilesSet := make(map[string]bool)

	// Check each module's CI status
	for _, module := range modulesWithCI {
		status := checkModuleCIStatusPerModule(module, headSHA, prBase, workspaceRoot, api, ciExcludedFiles)
		result.ModuleStatus[module] = status

		if status.HasValidCI {
			// Module has valid CI at HEAD - skip it
			result.Skipped = append(result.Skipped, module)
		} else {
			// Module needs CI
			result.Modules = append(result.Modules, module)
			result.DirectlyChanged = append(result.DirectlyChanged, module)

			// Track changed files for this module
			if status.FilesChanged > 0 {
				// Get the actual changed files for this module for reporting
				if baseSHA := status.LastSuccessSHA; baseSHA != "" {
					files, _ := getChangedFilesBetweenSHAs(baseSHA, headSHA, workspaceRoot)
					moduleFiles := filterFilesForModule(files, module, workspaceRoot, ciExcludedFiles)
					result.FilesByModule[module] = moduleFiles
					for _, f := range files {
						allChangedFilesSet[f] = true
					}
				}
			}
		}
	}

	// Calculate transitive invalidation
	// Modules that have valid CI but depend on modules that need CI
	depGraph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err == nil {
		needsCISet := make(map[string]bool)
		for _, m := range result.DirectlyChanged {
			needsCISet[m] = true
		}

		// Check each skipped module to see if its dependencies need CI
		newInvalidated := []string{}
		for _, skippedModule := range result.Skipped {
			if deps, ok := depGraph.Dependencies[skippedModule]; ok {
				for _, dep := range deps {
					if needsCISet[dep] {
						// This skipped module depends on a module that needs CI
						newInvalidated = append(newInvalidated, skippedModule)
						result.ModuleStatus[skippedModule] = ModuleCIStatus{
							HasValidCI: false,
							Reason:     fmt.Sprintf("dependency %s needs CI", dep),
						}
						break
					}
				}
			}
		}

		// Move invalidated modules from Skipped to Invalidated
		if len(newInvalidated) > 0 {
			// Remove from skipped
			newSkipped := []string{}
			invalidatedSet := make(map[string]bool)
			for _, m := range newInvalidated {
				invalidatedSet[m] = true
			}
			for _, m := range result.Skipped {
				if !invalidatedSet[m] {
					newSkipped = append(newSkipped, m)
				}
			}
			result.Skipped = newSkipped
			result.Invalidated = newInvalidated
			result.Modules = append(result.Modules, newInvalidated...)
		}
	}

	// Aggregate changed files for summary
	for f := range allChangedFilesSet {
		result.ChangedFiles = append(result.ChangedFiles, f)
	}
	result.ChangedFileCount = len(result.ChangedFiles)

	// Apply workflow filter if not already applied
	if !filterWorkflows {
		// Already filtered above, but re-filter the final lists to be safe
		result.Modules, _ = filterModulesWithWorkflows(result.Modules, workspaceRoot)
		result.DirectlyChanged, _ = filterModulesWithWorkflows(result.DirectlyChanged, workspaceRoot)
		result.Invalidated, _ = filterModulesWithWorkflows(result.Invalidated, workspaceRoot)
	}

	return result, nil
}

// checkModuleCIStatusPerModule checks the CI status for a single module.
// Returns whether the module has valid CI and why.
func checkModuleCIStatusPerModule(module, headSHA, prBase, workspaceRoot string, api github.API, ciExcludedFiles map[string]bool) ModuleCIStatus {
	workflowName := fmt.Sprintf("ci-%s.yaml", module)

	// Query module's CI workflow for last successful run
	runs, err := api.ListRuns(workflowName, github.ListRunsOpts{
		Status: "success",
		Limit:  1,
	})
	if err != nil {
		return ModuleCIStatus{
			HasValidCI: false,
			Reason:     fmt.Sprintf("query_failed: %v", err),
		}
	}

	// No successful CI runs for this module
	if len(runs) == 0 {
		return ModuleCIStatus{
			HasValidCI: false,
			Reason:     "no_ci_history",
		}
	}

	lastSuccessSHA := runs[0].HeadSHA

	// If CI passed at current HEAD, module has valid CI
	if lastSuccessSHA == headSHA {
		return ModuleCIStatus{
			HasValidCI:     true,
			LastSuccessSHA: lastSuccessSHA,
			Reason:         "valid_ci_at_head",
		}
	}

	// CI passed at different SHA - check if module's files changed
	baseSHA := lastSuccessSHA
	if prBase != "" {
		// For PRs, use PR base instead of last CI success
		baseSHA = prBase
	}

	// Get files changed since last CI success
	changedFiles, err := getChangedFilesBetweenSHAs(baseSHA, headSHA, workspaceRoot)
	if err != nil {
		return ModuleCIStatus{
			HasValidCI:     false,
			LastSuccessSHA: lastSuccessSHA,
			Reason:         fmt.Sprintf("diff_failed: %v", err),
		}
	}

	// Filter to files that affect this module (directly or via dependencies)
	moduleFiles := filterFilesForModule(changedFiles, module, workspaceRoot, ciExcludedFiles)

	if len(moduleFiles) == 0 {
		// No files affecting this module changed - CI is still valid
		return ModuleCIStatus{
			HasValidCI:     true,
			LastSuccessSHA: lastSuccessSHA,
			Reason:         "no_affecting_changes",
		}
	}

	// Files affecting this module changed - needs new CI
	return ModuleCIStatus{
		HasValidCI:     false,
		LastSuccessSHA: lastSuccessSHA,
		Reason:         fmt.Sprintf("files_changed_since_%s", lastSuccessSHA[:min(7, len(lastSuccessSHA))]),
		FilesChanged:   len(moduleFiles),
	}
}

// filterFilesForModule returns files that affect a given module (directly or via dependencies).
func filterFilesForModule(files []string, module, workspaceRoot string, ciExcludedFiles map[string]bool) []string {
	// First filter out CI-excluded files
	filteredFiles := filterOutCIExcludedFiles(files, ciExcludedFiles)

	if len(filteredFiles) == 0 {
		return []string{}
	}

	// Get modules affected by these files (including transitive dependents)
	affectedModules, err := repository.GetModulesRequiringRebuild(filteredFiles, workspaceRoot)
	if err != nil {
		// On error, assume all files affect this module
		return filteredFiles
	}

	// Check if our module is in the affected set
	for _, affected := range affectedModules {
		if affected == module {
			// Module is affected - return the files that directly belong to it
			directFiles, _ := getFilesByModule(filteredFiles, workspaceRoot)
			if moduleFiles, ok := directFiles[module]; ok {
				return moduleFiles
			}
			// If we can't determine direct files, return all (conservative)
			return filteredFiles
		}
	}

	// Module not affected by these files
	return []string{}
}

// buildCIChangedModulesResult builds the result structure (legacy, kept for compatibility)
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
