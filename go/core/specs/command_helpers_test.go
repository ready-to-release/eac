// Package specs contains godog step implementations for core features.
//
// This file provides in-process domain function implementations that replicate
// the behavior of CLI commands using only go/core imports. This avoids subprocess
// overhead (~200-500ms per call on Windows) for the 5 commands used by cache
// invalidation tests.
//
// These functions produce YAML/text output matching the CLI command output format
// so that the existing YAML assertion steps continue to work unchanged.
package specs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/hash"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/repository"
	"gopkg.in/yaml.v3"
)

// localChangedModulesResult matches the YAML output struct from
// go/commands/repository/get/changed-modules-local.go.
type localChangedModulesResult struct {
	Modules       []string          `yaml:"modules"`
	UpToDate      []string          `yaml:"up_to_date"`
	ChangeReasons map[string]string `yaml:"change_reasons,omitempty"`
	IsFreshBuild  bool              `yaml:"is_fresh_build"`
	DetectionTime string            `yaml:"detection_time"`
}

// ciChangedModulesResult matches the YAML output struct from
// go/commands/repository/get/changed-modules-ci.go.
type ciChangedModulesResult struct {
	Modules         []string                    `yaml:"modules"`
	DirectlyChanged []string                    `yaml:"directly_changed"`
	Invalidated     []string                    `yaml:"invalidated"`
	BaseSHA         string                      `yaml:"base_sha"`
	HeadSHA         string                      `yaml:"head_sha"`
	IsBootstrap     bool                        `yaml:"is_bootstrap"`
	ChangedFiles    []string                    `yaml:"changed_files"`
	ChangedFileCount int                        `yaml:"changed_file_count"`
	FilesByModule   map[string][]string         `yaml:"files_by_module"`
	ModuleStatus    map[string]moduleCIStatusResult `yaml:"module_status,omitempty"`
	Skipped         []string                    `yaml:"skipped,omitempty"`
	FilteredOut     []string                    `yaml:"filtered_out,omitempty"`
}

type moduleCIStatusResult struct {
	HasValidCI     bool   `yaml:"has_valid_ci"`
	LastSuccessSHA string `yaml:"last_success_sha,omitempty"`
	Reason         string `yaml:"reason"`
	FilesChanged   int    `yaml:"files_changed,omitempty"`
}

// mockedCIStatusJSON matches the JSON format written by runCommandWithMockedCI.
type mockedCIStatusJSON struct {
	LastSuccessSHA   string `json:"LastSuccessSHA"`
	HasFilesChanged  bool   `json:"HasFilesChanged"`
	HasValidCIAtHead bool   `json:"HasValidCIAtHead"`
	NoHistory        bool   `json:"NoHistory"`
}

// makeCoreInProcessDispatcher creates a CommandDispatcher that routes commands
// to in-process domain functions, falling back to subprocess for unknown commands.
func makeCoreInProcessDispatcher(ctx *eacgodog.TestContext) func(args []string) (string, int) {
	return func(args []string) (string, int) {
		if len(args) < 1 {
			return "", eacgodog.ExitCodeDispatchDeclined
		}

		workspaceRoot := ctx.IsolatedDir
		if workspaceRoot == "" {
			return "", eacgodog.ExitCodeDispatchDeclined
		}

		cmdName := strings.Join(args, " ")

		var dispatch func(string) (string, int)

		switch {
		case cmdName == "get changed-modules-local" ||
			strings.HasPrefix(cmdName, "get changed-modules-local "):
			dispatch = func(ws string) (string, int) { return detectLocalChangesInProcess(ws) }

		case cmdName == "get changed-modules-ci" ||
			strings.HasPrefix(cmdName, "get changed-modules-ci "):
			dispatch = func(ws string) (string, int) { return detectCIChangesInProcess(ws) }

		case strings.HasPrefix(cmdName, "build ") && strings.Contains(cmdName, "--dry-run"):
			dispatch = func(ws string) (string, int) { return buildDryRunInProcess(ws, args) }

		case strings.HasPrefix(cmdName, "lint ") && strings.Contains(cmdName, "--dry-run"):
			dispatch = func(ws string) (string, int) { return lintDryRunInProcess(ws, args) }

		case cmdName == "validate module-hierarchy":
			dispatch = func(ws string) (string, int) { return validateModuleHierarchyInProcess(ws) }

		default:
			return "", eacgodog.ExitCodeDispatchDeclined
		}

		// Apply mock overrides as environment variables (mirrors applyMockEnvironment
		// in the standard MakeInProcessDispatcher). This ensures commands like
		// "get changed-modules-ci" see CLIE_MOCK_CI_STATUS etc.
		restoreMockEnv := applyMockOverrides(ctx)
		defer restoreMockEnv()

		return dispatch(workspaceRoot)
	}
}

// applyMockOverrides sets ctx.MockOverrides as environment variables and returns
// a cleanup function that restores previous values.
func applyMockOverrides(ctx *eacgodog.TestContext) func() {
	if len(ctx.MockOverrides) == 0 {
		return func() {}
	}

	origValues := make(map[string]string)
	origExists := make(map[string]bool)

	for key, value := range ctx.MockOverrides {
		if orig, exists := os.LookupEnv(key); exists {
			origValues[key] = orig
			origExists[key] = true
		}
		os.Setenv(key, value)
	}

	return func() {
		for key := range ctx.MockOverrides {
			if origExists[key] {
				os.Setenv(key, origValues[key])
			} else {
				os.Unsetenv(key)
			}
		}
	}
}

// detectLocalChangesInProcess replicates "get changed-modules-local" using go/core domain functions.
func detectLocalChangesInProcess(workspaceRoot string) (string, int) {
	startTime := time.Now()

	reg, err := modules.LoadFromWorkspaceNoValidation(workspaceRoot)
	if err != nil {
		return fmt.Sprintf("Error: %v\n", err), 1
	}

	contracts := reg.All()
	reader := coreoutput.NewReader(workspaceRoot)

	var changedModules, upToDateModules []string
	changeReasons := make(map[string]string)
	isFreshBuild := true

	for _, contract := range contracts {
		moniker := contract.Moniker

		manifests, err := reader.ListUoWs(core.ActionBuild, moniker)
		if err != nil || len(manifests) == 0 {
			changedModules = append(changedModules, moniker)
			changeReasons[moniker] = "no build manifests found"
			continue
		}

		isFreshBuild = false

		files, err := hash.ExpandGlobPatterns(workspaceRoot, contract.GetGlobPatterns())
		if err != nil || len(files) == 0 {
			changedModules = append(changedModules, moniker)
			changeReasons[moniker] = "no source files found"
			continue
		}

		currentHash, err := hash.Files(workspaceRoot, files)
		if err != nil {
			changedModules = append(changedModules, moniker)
			changeReasons[moniker] = fmt.Sprintf("hash error: %v", err)
			continue
		}

		needsRebuild := false
		var mismatchReason string
		for _, manifest := range manifests {
			if manifest.InputHash != currentHash {
				needsRebuild = true
				mismatchReason = fmt.Sprintf("input hash mismatch in %s:%s",
					manifest.Component, manifest.Tool)
				break
			}
		}

		if needsRebuild {
			changedModules = append(changedModules, moniker)
			changeReasons[moniker] = mismatchReason
		} else {
			upToDateModules = append(upToDateModules, moniker)
		}
	}

	result := &localChangedModulesResult{
		Modules:       changedModules,
		UpToDate:      upToDateModules,
		ChangeReasons: changeReasons,
		IsFreshBuild:  isFreshBuild,
		DetectionTime: time.Since(startTime).Round(time.Millisecond).String(),
	}

	out, err := yaml.Marshal(result)
	if err != nil {
		return fmt.Sprintf("Error: marshal failed: %v\n", err), 1
	}
	return string(out) + "\n", 0
}

// detectCIChangesInProcess replicates "get changed-modules-ci" using go/core domain functions.
// Reads mock data from CLIE_MOCK_CI_STATUS, CLIE_MOCK_HEAD_SHA, CLIE_MOCK_CHANGED_FILES env vars.
func detectCIChangesInProcess(workspaceRoot string) (string, int) {
	headSHA := os.Getenv(environments.EnvCLIEMockHeadSHA)
	if headSHA == "" {
		return "Error: CLIE_MOCK_HEAD_SHA not set\n", 1
	}

	// Load mock CI status
	mockPath := os.Getenv(environments.EnvCLIEMockCIStatus)
	if mockPath == "" {
		return "Error: CLIE_MOCK_CI_STATUS not set\n", 1
	}

	data, err := os.ReadFile(mockPath)
	if err != nil {
		return fmt.Sprintf("Error: reading mock CI status: %v\n", err), 1
	}

	var mocks map[string]mockedCIStatusJSON
	if err := json.Unmarshal(data, &mocks); err != nil {
		return fmt.Sprintf("Error: parsing mock CI status: %v\n", err), 1
	}

	// Get changed files from env
	var changedFiles []string
	if files := os.Getenv(environments.EnvCLIEMockChangedFiles); files != "" {
		changedFiles = strings.Split(files, ",")
	}

	// Get all modules with CI workflows
	allModules, err := getAllModuleMonikersDomain(workspaceRoot)
	if err != nil {
		return fmt.Sprintf("Error: %v\n", err), 1
	}

	modulesWithCI, filteredOut := filterModulesWithWorkflowsDomain(allModules, workspaceRoot)

	// CI excluded file patterns
	ciExcludedPatterns := getCIExcludedFilesDomain(workspaceRoot)

	result := &ciChangedModulesResult{
		Modules:         []string{},
		DirectlyChanged: []string{},
		Invalidated:     []string{},
		BaseSHA:         "per-module",
		HeadSHA:         headSHA,
		IsBootstrap:     false,
		ChangedFiles:    []string{},
		ChangedFileCount: 0,
		FilesByModule:   map[string][]string{},
		ModuleStatus:    map[string]moduleCIStatusResult{},
		Skipped:         []string{},
		FilteredOut:     filteredOut,
	}

	allChangedFilesSet := make(map[string]bool)

	for _, module := range modulesWithCI {
		mock, hasMock := mocks[module]
		if !hasMock {
			// Module not in mock - treat as no CI history
			result.ModuleStatus[module] = moduleCIStatusResult{HasValidCI: false, Reason: "no_ci_history"}
			result.Modules = append(result.Modules, module)
			result.DirectlyChanged = append(result.DirectlyChanged, module)
			continue
		}

		status := checkModuleCIStatusFromMock(module, headSHA, mock, changedFiles, workspaceRoot, ciExcludedPatterns)
		result.ModuleStatus[module] = status

		if status.HasValidCI {
			result.Skipped = append(result.Skipped, module)
		} else {
			result.Modules = append(result.Modules, module)
			result.DirectlyChanged = append(result.DirectlyChanged, module)

			if status.FilesChanged > 0 && len(changedFiles) > 0 {
				moduleFiles := filterFilesForModuleDomain(changedFiles, module, workspaceRoot, ciExcludedPatterns)
				if len(moduleFiles) > 0 {
					result.FilesByModule[module] = moduleFiles
					for _, f := range moduleFiles {
						allChangedFilesSet[f] = true
					}
				}
			}
		}
	}

	// Transitive invalidation
	depGraph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err == nil {
		needsCISet := make(map[string]bool)
		for _, m := range result.DirectlyChanged {
			needsCISet[m] = true
		}

		var allInvalidated []string
		remainingSkipped := result.Skipped

		for {
			var newInvalidated []string
			var stillSkipped []string

			for _, skippedModule := range remainingSkipped {
				invalidated := false
				if deps, ok := depGraph.Dependencies[skippedModule]; ok {
					for _, dep := range deps {
						if needsCISet[dep] {
							newInvalidated = append(newInvalidated, skippedModule)
							result.ModuleStatus[skippedModule] = moduleCIStatusResult{
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
				break
			}

			for _, m := range newInvalidated {
				needsCISet[m] = true
				allInvalidated = append(allInvalidated, m)
			}
			remainingSkipped = stillSkipped
		}

		if len(allInvalidated) > 0 {
			result.Skipped = remainingSkipped
			result.Invalidated = allInvalidated
			result.Modules = append(result.Modules, allInvalidated...)
		}
	}

	for f := range allChangedFilesSet {
		result.ChangedFiles = append(result.ChangedFiles, f)
	}
	result.ChangedFileCount = len(result.ChangedFiles)

	// Filter modules with workflows (same as command)
	result.Modules, _ = filterModulesWithWorkflowsDomain(result.Modules, workspaceRoot)
	result.DirectlyChanged, _ = filterModulesWithWorkflowsDomain(result.DirectlyChanged, workspaceRoot)
	result.Invalidated, _ = filterModulesWithWorkflowsDomain(result.Invalidated, workspaceRoot)

	out, err := yaml.Marshal(result)
	if err != nil {
		return fmt.Sprintf("Error: marshal failed: %v\n", err), 1
	}
	return string(out) + "\n", 0
}

// buildDryRunInProcess replicates "build --dry-run <module>" output.
func buildDryRunInProcess(workspaceRoot string, args []string) (string, int) {
	module := extractModuleFromArgs(args)
	if module == "" {
		return "Error: no module specified\n", 1
	}

	reg, err := modules.LoadFromWorkspaceNoValidation(workspaceRoot)
	if err != nil {
		return fmt.Sprintf("Error: %v\n", err), 1
	}

	contract, ok := reg.Get(module)
	if !ok {
		return fmt.Sprintf("Error: module not found: %s\n", module), 1
	}

	// Check UoW manifests for cache status
	reader := coreoutput.NewReader(workspaceRoot)
	manifests, _ := reader.ListUoWs(core.ActionBuild, module)

	var sb strings.Builder
	if len(manifests) == 0 {
		// No prior build state
		sb.WriteString(fmt.Sprintf("Build: %s (dry-run)\n", module))
		sb.WriteString(fmt.Sprintf("Components: %v\n", contract.GetEnabledComponents()))
		sb.WriteString(fmt.Sprintf("\n%s would be built (changed)\n", module))
	} else {
		// Check if any manifests have mismatched hashes or previous failure
		files, _ := hash.ExpandGlobPatterns(workspaceRoot, contract.GetGlobPatterns())
		currentHash, _ := hash.Files(workspaceRoot, files)

		changed := false
		previousFailed := false
		for _, m := range manifests {
			if m.InputHash != currentHash {
				changed = true
				break
			}
			if m.ExitCode != 0 {
				previousFailed = true
			}
		}

		if changed || previousFailed {
			sb.WriteString(fmt.Sprintf("Build: %s (dry-run)\n", module))
			sb.WriteString(fmt.Sprintf("%s would be built (changed)\n", module))
		} else {
			sb.WriteString(fmt.Sprintf("Build: %s (dry-run)\n", module))
			sb.WriteString(fmt.Sprintf("%s up-to-date (cached)\n", module))
		}
	}

	return sb.String(), 0
}

// lintDryRunInProcess replicates "lint --dry-run <module>" output.
func lintDryRunInProcess(workspaceRoot string, args []string) (string, int) {
	module := extractModuleFromArgs(args)
	if module == "" {
		return "Error: no module specified\n", 1
	}

	reg, err := modules.LoadFromWorkspaceNoValidation(workspaceRoot)
	if err != nil {
		return fmt.Sprintf("Error: %v\n", err), 1
	}

	contract, ok := reg.Get(module)
	if !ok {
		return fmt.Sprintf("Error: module not found: %s\n", module), 1
	}

	// Check UoW manifests for lint cache
	reader := coreoutput.NewReader(workspaceRoot)
	manifests, _ := reader.ListUoWs(core.ActionLint, module)

	var sb strings.Builder
	if len(manifests) == 0 {
		sb.WriteString(fmt.Sprintf("Lint: %s (dry-run)\n", module))
		sb.WriteString(fmt.Sprintf("Components: %v\n", contract.GetEnabledComponents()))
		sb.WriteString(fmt.Sprintf("\n%s would be linted (changed)\n", module))
	} else {
		files, _ := hash.ExpandGlobPatterns(workspaceRoot, contract.GetGlobPatterns())
		currentHash, _ := hash.Files(workspaceRoot, files)

		changed := false
		previousFailed := false
		for _, m := range manifests {
			if m.InputHash != currentHash {
				changed = true
				break
			}
			if m.ExitCode != 0 {
				previousFailed = true
			}
		}

		if changed || previousFailed {
			sb.WriteString(fmt.Sprintf("Lint: %s (dry-run)\n", module))
			sb.WriteString(fmt.Sprintf("%s would be linted (changed)\n", module))
		} else {
			sb.WriteString(fmt.Sprintf("Lint: %s (dry-run)\n", module))
			sb.WriteString(fmt.Sprintf("%s up-to-date (cached)\n", module))
		}
	}

	return sb.String(), 0
}

// validateModuleHierarchyInProcess replicates "validate module-hierarchy" using go/core domain.
func validateModuleHierarchyInProcess(workspaceRoot string) (string, int) {
	reg, err := modules.LoadFromWorkspaceNoValidation(workspaceRoot)
	if err != nil {
		return fmt.Sprintf("Error: %v\n", err), 1
	}

	// Check for non-existent dependency references
	var issues []string
	allMonikers := make(map[string]bool)
	for _, m := range reg.All() {
		allMonikers[m.Moniker] = true
	}

	for _, m := range reg.All() {
		for _, dep := range m.GetDependencies() {
			if !allMonikers[dep] {
				issues = append(issues, fmt.Sprintf("Module '%s' depends on '%s', but '%s' does not exist", m.Moniker, dep, dep))
			}
		}
	}

	// Check for circular dependencies
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var detectCycle func(moniker string, path []string) bool
	detectCycle = func(moniker string, path []string) bool {
		visited[moniker] = true
		recStack[moniker] = true
		currentPath := append(path, moniker)

		contract, ok := reg.Get(moniker)
		if !ok {
			return false
		}

		for _, dep := range contract.GetDependencies() {
			if !visited[dep] {
				if detectCycle(dep, currentPath) {
					return true
				}
			} else if recStack[dep] {
				// Find cycle start
				cycleStart := 0
				for i, p := range currentPath {
					if p == dep {
						cycleStart = i
						break
					}
				}
				cycle := append(currentPath[cycleStart:], dep)
				issues = append(issues, fmt.Sprintf("Circular dependency: %s", strings.Join(cycle, " -> ")))
				return true
			}
		}

		recStack[moniker] = false
		return false
	}

	for _, m := range reg.All() {
		if !visited[m.Moniker] {
			detectCycle(m.Moniker, nil)
		}
	}

	if len(issues) == 0 {
		return "All module hierarchy checks passed!\n", 0
	}

	var sb strings.Builder
	sb.WriteString("Module Hierarchy Validation Report\n\n")
	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("  %s\n", issue))
	}
	return sb.String(), 1
}

// extractModuleFromArgs extracts the module name from command args, skipping flags.
func extractModuleFromArgs(args []string) string {
	// Skip the command verb (e.g., "build" or "lint")
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			continue
		}
		return arg
	}
	return ""
}

// Domain helper functions (replicate logic from go/commands/repository/get/).

func getAllModuleMonikersDomain(workspaceRoot string) ([]string, error) {
	graph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err != nil {
		return nil, err
	}
	return graph.Modules, nil
}

func filterModulesWithWorkflowsDomain(monikers []string, workspaceRoot string) ([]string, []string) {
	var filtered, filteredOut []string
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

// ciExcludedFilePatterns are documentation files that don't trigger CI rebuilds.
var ciExcludedFilePatterns = []string{
	"CHANGELOG.md",
	"README.md",
	"CONTRIBUTING.md",
	"LICENSE",
}

func getCIExcludedFilesDomain(workspaceRoot string) map[string]bool {
	result := make(map[string]bool)
	reg, err := modules.LoadFromWorkspaceNoValidation(workspaceRoot)
	if err != nil {
		return result
	}

	for _, module := range reg.All() {
		for _, root := range module.GetComponentRoots() {
			for _, pattern := range ciExcludedFilePatterns {
				path := filepath.Join(root, pattern)
				path = filepath.ToSlash(path)
				result[path] = true
			}
		}
	}
	return result
}

func checkModuleCIStatusFromMock(module, headSHA string, mock mockedCIStatusJSON, changedFiles []string, workspaceRoot string, ciExcludedFiles map[string]bool) moduleCIStatusResult {
	if mock.NoHistory {
		return moduleCIStatusResult{HasValidCI: false, Reason: "no_ci_history"}
	}

	if mock.HasValidCIAtHead {
		return moduleCIStatusResult{
			HasValidCI:     true,
			LastSuccessSHA: headSHA,
			Reason:         "valid_ci_at_head",
		}
	}

	if !mock.HasFilesChanged {
		return moduleCIStatusResult{
			HasValidCI:     true,
			LastSuccessSHA: mock.LastSuccessSHA,
			Reason:         "no_affecting_changes",
		}
	}

	// HasFilesChanged is true
	if len(changedFiles) == 0 {
		shortSHA := mock.LastSuccessSHA
		if len(shortSHA) > 7 {
			shortSHA = shortSHA[:7]
		}
		return moduleCIStatusResult{
			HasValidCI:     false,
			LastSuccessSHA: mock.LastSuccessSHA,
			Reason:         fmt.Sprintf("files_changed_since_%s", shortSHA),
			FilesChanged:   1,
		}
	}

	moduleFiles := filterFilesForModuleDomain(changedFiles, module, workspaceRoot, ciExcludedFiles)
	if len(moduleFiles) == 0 {
		return moduleCIStatusResult{
			HasValidCI:     true,
			LastSuccessSHA: mock.LastSuccessSHA,
			Reason:         "only_ci_excluded_changes",
		}
	}

	shortSHA := mock.LastSuccessSHA
	if len(shortSHA) > 7 {
		shortSHA = shortSHA[:7]
	}
	return moduleCIStatusResult{
		HasValidCI:     false,
		LastSuccessSHA: mock.LastSuccessSHA,
		Reason:         fmt.Sprintf("files_changed_since_%s", shortSHA),
		FilesChanged:   len(moduleFiles),
	}
}

func filterFilesForModuleDomain(files []string, module, workspaceRoot string, ciExcludedFiles map[string]bool) []string {
	// Filter out CI-excluded files
	var filtered []string
	for _, f := range files {
		excluded := false
		for pattern := range ciExcludedFiles {
			if f == pattern {
				excluded = true
				break
			}
			if strings.ContainsAny(pattern, "*?") {
				if matched, _ := filepath.Match(pattern, f); matched {
					excluded = true
					break
				}
			}
		}
		if !excluded {
			filtered = append(filtered, f)
		}
	}

	if len(filtered) == 0 {
		return []string{}
	}

	// Map files to modules
	reg, err := modules.LoadFromWorkspaceNoValidation(workspaceRoot)
	if err != nil {
		return []string{}
	}

	var moduleFiles []string
	for _, f := range filtered {
		if f == "" {
			continue
		}
		matchingModules := reg.FindModulesForFile(f)
		for _, m := range matchingModules {
			if m.Moniker == module {
				moduleFiles = append(moduleFiles, f)
				break
			}
		}
	}
	return moduleFiles
}
