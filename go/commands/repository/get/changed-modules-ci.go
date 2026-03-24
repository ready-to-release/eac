package get

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/adapters/gh"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

type getChangedModulesCICommand struct{}

var _ core.SimpleCommandPort = (*getChangedModulesCICommand)(nil)

func (c *getChangedModulesCICommand) Name() string { return "get changed-modules-ci" }

func (c *getChangedModulesCICommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-changed-modules-ci",
		Short:         "Get modules requiring rebuild since last successful CI run",
		Long: "",
		Notes: "Expected Output:\nYAML list of modules needing rebuild based on per-module CI state.\n\nPer-module change detection:\n  For each module with a ci-{module}.yaml workflow:\n  1. Query the module's last successful CI run\n  2. If CI passed at current HEAD SHA -> skip (no rebuild needed)\n  3. If CI passed at different SHA -> check if module's files changed\n  4. If no CI history -> module needs CI (but NOT bootstrap for all)\n\nOutput includes:\n  - All modules requiring rebuild (directly changed + transitive dependents)\n  - Directly changed modules (files modified since module's last CI success)\n  - Invalidated modules (transitive dependents requiring rebuild)\n  - Per-module CI status reasoning\n\nWith --format shell, outputs shell variable assignments:\n  MODULES=\"mod1 mod2 mod3\"\n  DIRECTLY_CHANGED=\"mod1 mod2\"\n  INVALIDATED=\"mod3\"\n  BASE_SHA=\"per-module\"\n  IS_BOOTSTRAP=\"false\"\n  CHANGED_FILE_COUNT=\"5\"",
		Flags: []core.FlagSpec{
			{Name: "pr-base", Type: "string", Usage: "For PRs, the base SHA to compare against"},
			{Name: "filter-workflows", Type: "bool", Usage: "Only include modules that have a ci-{module}.yaml workflow file"},
			{Name: "format", Type: "string", Usage: "Output format (shell outputs shell variables; otherwise uses standard get command formats)"},
		},
	}
}

func (c *getChangedModulesCICommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetChangedModulesCI()
}

// CIExcludedFilePatterns are documentation files that don't trigger CI rebuilds.
// These patterns are matched against file basenames within each module's component roots.
var CIExcludedFilePatterns = []string{
	"CHANGELOG.md",
	"README.md",
	"CONTRIBUTING.md",
	"LICENSE",
}

// CIChangedModulesResult represents the output of the get changed-modules-ci command.
type CIChangedModulesResult struct {
	Modules         []string `json:"modules" yaml:"modules" toml:"modules"`
	DirectlyChanged []string `json:"directly_changed" yaml:"directly_changed" toml:"directly_changed"`
	Invalidated     []string `json:"invalidated" yaml:"invalidated" toml:"invalidated"`
	BaseSHA         string   `json:"base_sha" yaml:"base_sha" toml:"base_sha"`
	HeadSHA         string   `json:"head_sha" yaml:"head_sha" toml:"head_sha"`
	IsBootstrap     bool     `json:"is_bootstrap" yaml:"is_bootstrap" toml:"is_bootstrap"`
	// Additional context for CI reasoning
	ChangedFiles     []string            `json:"changed_files" yaml:"changed_files" toml:"changed_files"`
	ChangedFileCount int                 `json:"changed_file_count" yaml:"changed_file_count" toml:"changed_file_count"`
	FilesByModule    map[string][]string `json:"files_by_module" yaml:"files_by_module" toml:"files_by_module"`
	// Per-module CI status (why each module needs/doesn't need CI)
	ModuleStatus map[string]ModuleCIStatus `json:"module_status,omitempty" yaml:"module_status,omitempty" toml:"module_status,omitempty"`
	// Modules that were skipped (have valid CI at HEAD)
	Skipped []string `json:"skipped,omitempty" yaml:"skipped,omitempty" toml:"skipped,omitempty"`
	// Workflow filtering (only present when --filter-workflows is used)
	FilteredOut []string `json:"filtered_out,omitempty" yaml:"filtered_out,omitempty" toml:"filtered_out,omitempty"`
}

// ModuleCIStatus tracks the CI status for a single module.
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

	// Get current HEAD SHA (check for mock first)
	headSHA := getMockedHeadSHA()
	if headSHA == "" {
		shaResult, err := DetectCurrentSHA(workspaceRoot, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: getting current SHA: %v\n", err)
			return 1
		}
		headSHA = shaResult.SHA
	}

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

	// Create GitHub API client and CI run querier (may not be used if mocking)
	api := github.NewGHClient(gh.New(tool.GlobalToolSystem(), workspaceRoot), workspaceRoot)
	querier := NewGHCIRunQuerier(api)

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

	// Check for mock support
	mockedStatus := loadMockedCIStatus()
	mockedChangedFiles := getMockedChangedFiles()

	// Check each module's CI status
	for _, module := range modulesWithCI {
		var status ModuleCIStatus

		// Use mocked status if available
		if mockedStatus != nil {
			if mock, ok := mockedStatus[module]; ok {
				status = checkModuleCIStatusMocked(module, headSHA, mock, mockedChangedFiles, workspaceRoot, ciExcludedFiles)
			} else {
				// Module not in mock - treat as no CI history
				status = ModuleCIStatus{HasValidCI: false, Reason: "no_ci_history"}
			}
		} else {
			// Use real GitHub API via querier
			status = checkModuleCIStatusPerModule(module, headSHA, prBase, workspaceRoot, querier, ciExcludedFiles)
		}
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
				// Get changed files for this module for reporting
				var files []string
				if mockedChangedFiles != nil {
					files = mockedChangedFiles
				} else if mockedStatus == nil {
					// Only call git diff when not in mocked mode (tool system may not be initialized)
					if baseSHA := status.LastSuccessSHA; baseSHA != "" {
						files, _ = getChangedFilesBetweenSHAs(baseSHA, headSHA, workspaceRoot)
					}
				}
				if len(files) > 0 {
					moduleFiles := filterFilesForModule(files, module, workspaceRoot, ciExcludedFiles)
					result.FilesByModule[module] = moduleFiles
					// Only count files directly owned by this module
					for _, f := range moduleFiles {
						allChangedFilesSet[f] = true
					}
				}
			}
		}
	}

	// Calculate transitive invalidation (iteratively until no new modules invalidated)
	// Modules that have valid CI but depend on modules that need CI
	depGraph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err == nil {
		needsCISet := make(map[string]bool)
		for _, m := range result.DirectlyChanged {
			needsCISet[m] = true
		}

		// Iteratively propagate invalidation until no new modules are found
		allInvalidated := []string{}
		remainingSkipped := result.Skipped

		for {
			newInvalidated := []string{}
			stillSkipped := []string{}

			for _, skippedModule := range remainingSkipped {
				invalidated := false
				if deps, ok := depGraph.Dependencies[skippedModule]; ok {
					for _, dep := range deps {
						if needsCISet[dep] {
							// This skipped module depends on a module that needs CI
							newInvalidated = append(newInvalidated, skippedModule)
							result.ModuleStatus[skippedModule] = ModuleCIStatus{
								HasValidCI: false,
								Reason:     fmt.Sprintf("dependency %s needs CI", dep),
							}
							invalidated = true
							break
						}
					}
				}
				if !invalidated {
					stillSkipped = append(stillSkipped, skippedModule)
				}
			}

			if len(newInvalidated) == 0 {
				break // No more modules to invalidate
			}

			// Add newly invalidated to needsCISet for next iteration
			for _, m := range newInvalidated {
				needsCISet[m] = true
				allInvalidated = append(allInvalidated, m)
			}
			remainingSkipped = stillSkipped
		}

		// Update results
		if len(allInvalidated) > 0 {
			result.Skipped = remainingSkipped
			result.Invalidated = allInvalidated
			result.Modules = append(result.Modules, allInvalidated...)
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
func checkModuleCIStatusPerModule(module, headSHA, prBase, workspaceRoot string, querier cache.CIRunQuerier, ciExcludedFiles map[string]bool) ModuleCIStatus {
	workflowName := fmt.Sprintf("ci-%s.yaml", module)

	// Query module's CI workflow for last successful run via the querier port
	lastSuccessSHA, err := querier.LastSuccessfulRunSHA(workflowName)
	if err != nil {
		return ModuleCIStatus{
			HasValidCI: false,
			Reason:     fmt.Sprintf("query_failed: %v", err),
		}
	}

	// No successful CI runs for this module
	if lastSuccessSHA == "" {
		return ModuleCIStatus{
			HasValidCI: false,
			Reason:     "no_ci_history",
		}
	}

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

// filterFilesForModule returns files that are directly owned by a given module.
// This only returns files that the module directly owns, NOT files that affect it via dependencies.
func filterFilesForModule(files []string, module, workspaceRoot string, ciExcludedFiles map[string]bool) []string {
	// First filter out CI-excluded files
	filteredFiles := filterOutCIExcludedFiles(files, ciExcludedFiles)

	if len(filteredFiles) == 0 {
		return []string{}
	}

	// Get files directly owned by this module (not transitive)
	directFiles, err := getFilesByModule(filteredFiles, workspaceRoot)
	if err != nil {
		return []string{}
	}

	if moduleFiles, ok := directFiles[module]; ok {
		return moduleFiles
	}

	return []string{}
}

// buildCIChangedModulesResult builds the result structure (legacy, kept for compatibility).
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
// Returns (baseSHA, isBootstrap, error).
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

// getLastSuccessfulCISHA queries gh CLI for the last successful workflow run SHA.
func getLastSuccessfulCISHA(workflow, branch, workspaceRoot string) (string, error) {
	// gh run list -b <branch> -s success -w "<workflow>" -L 1 --json headSha -q '.[0].headSha'
	output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", workspaceRoot, "run", "list",
		"-b", branch,
		"-s", "success",
		"-w", workflow,
		"-L", "1",
		"--json", "headSha",
		"-q", ".[0].headSha",
	)
	if err != nil {
		return "", fmt.Errorf("gh command failed: %w", err)
	}

	sha := strings.TrimSpace(string(output))
	return sha, nil
}

// getCurrentSHA gets the current HEAD SHA.
func getCurrentSHA(workspaceRoot string) (string, error) {
	ts := tool.GlobalToolSystem()
	output, err := ts.RunTool(context.Background(), "git", workspaceRoot, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getChangedFilesBetweenSHAs gets the list of files changed between two SHAs.
func getChangedFilesBetweenSHAs(baseSHA, headSHA, workspaceRoot string) ([]string, error) {
	ts := tool.GlobalToolSystem()
	output, err := ts.RunTool(context.Background(), "git", workspaceRoot, "diff", "--name-only", baseSHA+".."+headSHA)
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}

// getAllModuleMonikers returns all module monikers (for bootstrap case).
func getAllModuleMonikers(workspaceRoot string) ([]string, error) {
	graph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err != nil {
		return nil, err
	}
	return graph.Modules, nil
}

// getFilesByModule maps changed files to their owning modules.
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
// Returns (filtered modules, modules that were filtered out).
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

// getCIExcludedFiles returns a set of file paths/patterns that should not trigger CI.
// These are files that are owned by modules but don't affect module functionality:
// - CHANGELOG.md: release documentation only
// - README.md: documentation only
// - CONTRIBUTING.md: contribution guidelines
// - LICENSE: legal text
func getCIExcludedFiles(workspaceRoot string) map[string]bool {
	result := make(map[string]bool)

	// Use the package-level excluded file patterns
	commonPatterns := CIExcludedFilePatterns

	registry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return result // Return empty on error - non-fatal
	}

	// Add patterns for each module's component roots
	for _, module := range registry.All() {
		for _, root := range module.GetComponentRoots() {
			for _, pattern := range commonPatterns {
				// Add both with and without trailing separator for flexibility
				path := filepath.Join(root, pattern)
				path = filepath.ToSlash(path) // Normalize to forward slashes
				result[path] = true
			}
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
			if matched, matchErr := filepath.Match(pattern, filePath); matchErr == nil && matched {
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

// ============================================================================
// Mock support for testing
// ============================================================================

// mockedCIStatus represents mocked CI status for a module (matches step definitions).
type mockedCIStatus struct {
	LastSuccessSHA   string `json:"LastSuccessSHA"`
	HasFilesChanged  bool   `json:"HasFilesChanged"`
	HasValidCIAtHead bool   `json:"HasValidCIAtHead"`
	NoHistory        bool   `json:"NoHistory"`
}

// loadMockedCIStatus loads mocked CI status from CLIE_MOCK_CI_STATUS environment variable.
// Returns nil if no mock is configured.
func loadMockedCIStatus() map[string]mockedCIStatus {
	mockPath := os.Getenv(environments.EnvCLIEMockCIStatus)
	if mockPath == "" {
		return nil
	}

	data, err := os.ReadFile(mockPath)
	if err != nil {
		return nil
	}

	var mocks map[string]mockedCIStatus
	if err := json.Unmarshal(data, &mocks); err != nil {
		return nil
	}
	return mocks
}

// getMockedHeadSHA returns the mocked HEAD SHA if CLIE_MOCK_HEAD_SHA is set.
func getMockedHeadSHA() string {
	return os.Getenv(environments.EnvCLIEMockHeadSHA)
}

// getMockedChangedFiles returns the mocked changed files if CLIE_MOCK_CHANGED_FILES is set.
func getMockedChangedFiles() []string {
	files := os.Getenv(environments.EnvCLIEMockChangedFiles)
	if files == "" {
		return nil
	}
	return strings.Split(files, ",")
}

// checkModuleCIStatusMocked checks CI status using mocked data.
func checkModuleCIStatusMocked(module, headSHA string, mock mockedCIStatus, changedFiles []string, workspaceRoot string, ciExcludedFiles map[string]bool) ModuleCIStatus {
	// No CI history
	if mock.NoHistory {
		return ModuleCIStatus{
			HasValidCI: false,
			Reason:     "no_ci_history",
		}
	}

	// Valid CI at HEAD
	if mock.HasValidCIAtHead {
		return ModuleCIStatus{
			HasValidCI:     true,
			LastSuccessSHA: headSHA,
			Reason:         "valid_ci_at_head",
		}
	}

	// CI at different SHA - check if module files changed
	if !mock.HasFilesChanged {
		// No files changed that affect this module
		return ModuleCIStatus{
			HasValidCI:     true,
			LastSuccessSHA: mock.LastSuccessSHA,
			Reason:         "no_affecting_changes",
		}
	}

	// HasFilesChanged is true - if no specific files provided, trust the mock
	if len(changedFiles) == 0 {
		// Mock says files changed, but no specific files listed - trust the mock
		return ModuleCIStatus{
			HasValidCI:     false,
			LastSuccessSHA: mock.LastSuccessSHA,
			Reason:         fmt.Sprintf("files_changed_since_%s", mock.LastSuccessSHA[:min(7, len(mock.LastSuccessSHA))]),
			FilesChanged:   1, // Indicate at least 1 file changed
		}
	}

	// Filter changed files to those affecting this module
	moduleFiles := filterFilesForModule(changedFiles, module, workspaceRoot, ciExcludedFiles)
	if len(moduleFiles) == 0 {
		// All changed files were CI-excluded
		return ModuleCIStatus{
			HasValidCI:     true,
			LastSuccessSHA: mock.LastSuccessSHA,
			Reason:         "only_ci_excluded_changes",
		}
	}

	// Files affecting this module changed - needs CI
	return ModuleCIStatus{
		HasValidCI:     false,
		LastSuccessSHA: mock.LastSuccessSHA,
		Reason:         fmt.Sprintf("files_changed_since_%s", mock.LastSuccessSHA[:min(7, len(mock.LastSuccessSHA))]),
		FilesChanged:   len(moduleFiles),
	}
}
