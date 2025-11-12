package repository

import (
	"strings"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// EnrichFilesWithModules takes a list of files and determines which module(s) own each file.
// Returns a list of files with their module ownership information.
//
// A file can belong to multiple modules if their glob patterns overlap.
//
// Parameters:
//   - files: List of FileInfo from GetRepositoryFiles
//   - workspaceRoot: Root directory of the workspace
//   - version: Module contract version (e.g., "0.1.0")
//
// Returns:
//   - List of RepositoryFileWithModule with normalized paths and module ownership
//   - Error if module contracts cannot be loaded
//
// Example:
//
//	files, _ := repository.GetRepositoryFiles(true, false, "")
//	enriched, _ := repository.EnrichFilesWithModules(files, "/workspace", "0.1.0")
//	for _, f := range enriched {
//	    fmt.Printf("%s -> %v\n", f.Name, f.Modules)
//	}
func EnrichFilesWithModules(files []FileInfo, workspaceRoot string, version string) ([]RepositoryFileWithModule, error) {
	// Load module contracts
	registry, err := modules.LoadFromWorkspace(workspaceRoot, version)
	if err != nil {
		return nil, NewRepositoryError("enrich", workspaceRoot, err, "failed to load module contracts")
	}

	// Create result list
	result := make([]RepositoryFileWithModule, 0, len(files))

	// For each file, determine which module(s) own it
	for _, file := range files {
		// Normalize path to forward slashes
		normalizedPath := strings.ReplaceAll(file.Path, "\\", "/")

		// Use registry's FindModulesForFile which handles:
		// - exclude_children_owned_source filtering
		// - catch-all module fallback
		matchingModules := registry.FindModulesForFile(normalizedPath)

		// Filter to only the closest modules in parent chain
		closestModules := filterClosestModules(matchingModules, registry)

		// Extract monikers from filtered modules
		owningModules := make([]string, 0, len(closestModules))
		for _, module := range closestModules {
			owningModules = append(owningModules, module.Moniker)
		}

		// Add to result (even if no modules match - will have empty Modules slice)
		result = append(result, RepositoryFileWithModule{
			Name:    normalizedPath,
			Modules: owningModules,
		})
	}

	return result, nil
}

// filterClosestModules filters a list of modules to only include those that are
// closest in the parent chain hierarchy. When multiple modules match a file,
// this ensures we only keep the most specific (deepest) modules.
//
// Special handling for "repository" module:
//   - The "repository" module is a catch-all that should only own files
//     that no other specific module claims
//   - If any non-repository modules match, repository is excluded
//   - Repository is only kept if it's the ONLY match
//
// Special handling for exclude_children_owned_source:
//   - If a parent module has exclude_children_owned_source: true (default)
//     and any of its direct children match the file, the parent is excluded
//   - This allows children to take ownership of files in their parent's source space
//
// Algorithm:
//  1. Check if "repository" is present with other modules → exclude repository
//  2. Check exclude_children_owned_source → exclude parents when children match
//  3. Calculate depth for each remaining module (distance from root)
//  4. Find the maximum depth
//  5. Keep only modules at maximum depth
//
// Examples:
//   - [claude-agents (depth 2), repository (depth 1)] → [claude-agents]
//   - [repository] → [repository] (only match)
//   - [claude (parent), claude-agents (child)] → [claude-agents] (if claude has exclude_children_owned_source: true)
//   - [docs-guide (depth 2), docs-reference (depth 2), docs (depth 1)] → [docs-guide, docs-reference]
//   - [readme (depth 1), repository (depth 1)] → [readme] (repository excluded)
//
// Parameters:
//   - matchingModules: List of modules that match a file
//   - registry: Module registry for parent chain resolution
//
// Returns:
//   - Filtered list containing only the closest (most specific) modules
func filterClosestModules(matchingModules []*modules.ModuleContract, registry *modules.Registry) []*modules.ModuleContract {
	// No filtering needed for 0 or 1 modules
	if len(matchingModules) <= 1 {
		return matchingModules
	}

	// Special handling for "repository" module
	// Repository should only own files that no other module claims
	hasRepository := false
	hasOtherModules := false

	for _, module := range matchingModules {
		if module.Moniker == "repository" {
			hasRepository = true
		} else {
			hasOtherModules = true
		}
	}

	// If repository is present with other modules, exclude it
	if hasRepository && hasOtherModules {
		filteredModules := make([]*modules.ModuleContract, 0, len(matchingModules)-1)
		for _, module := range matchingModules {
			if module.Moniker != "repository" {
				filteredModules = append(filteredModules, module)
			}
		}
		matchingModules = filteredModules
	}

	// If only repository remains or no modules, return as is
	if len(matchingModules) <= 1 {
		return matchingModules
	}

	// Filter modules based on exclude_children_owned_source
	// If a parent has this set to true and any of its children match, exclude the parent
	modulesToExclude := make(map[string]bool)

	for _, parent := range matchingModules {
		// Check if this module should exclude children
		if parent.Source.ExcludeChildrenOwnedSource != nil && *parent.Source.ExcludeChildrenOwnedSource {
			// Check if any child modules are in the matching list
			for _, candidate := range matchingModules {
				if candidate.Moniker == parent.Moniker {
					continue
				}
				// Check if candidate is a child of parent
				if candidate.Parent == parent.Moniker {
					// Child found - mark parent for exclusion
					modulesToExclude[parent.Moniker] = true
					break
				}
			}
		}
	}

	// Apply exclusions
	if len(modulesToExclude) > 0 {
		filteredModules := make([]*modules.ModuleContract, 0, len(matchingModules))
		for _, module := range matchingModules {
			if !modulesToExclude[module.Moniker] {
				filteredModules = append(filteredModules, module)
			}
		}
		matchingModules = filteredModules
	}

	// If only one module remains, return as is
	if len(matchingModules) <= 1 {
		return matchingModules
	}

	// Calculate depth for each module
	depths := make(map[string]int)
	maxDepth := 0

	for _, module := range matchingModules {
		depth, err := modules.GetDepth(module, registry)
		if err != nil {
			// On error, default to depth 1 (root level)
			// This ensures we don't lose modules due to validation errors
			depth = 1
		}

		depths[module.Moniker] = depth
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	// Filter to only modules at maximum depth
	result := make([]*modules.ModuleContract, 0, len(matchingModules))
	for _, module := range matchingModules {
		if depths[module.Moniker] == maxDepth {
			result = append(result, module)
		}
	}

	return result
}

// GetRepositoryFilesWithModules is a convenience function that combines
// GetRepositoryFiles and EnrichFilesWithModules in one call.
//
// Parameters:
//   - trackedOnly: if true, only return files tracked by Git
//   - includeIgnored: if true, include files ignored by .gitignore
//   - stagedOnly: if true, only return files currently staged in Git index
//   - rootPath: repository root (if empty, will be detected automatically)
//   - version: module contract version (e.g., "0.1.0")
//
// Returns:
//   - List of files with module ownership information
//   - Error if repository operations or module loading fails
//
// Example:
//
//	files, err := repository.GetRepositoryFilesWithModules(true, false, false, "", "0.1.0")
//	for _, f := range files {
//	    if len(f.Modules) > 1 {
//	        fmt.Printf("Multi-ownership: %s -> %v\n", f.Name, f.Modules)
//	    }
//	}
func GetRepositoryFilesWithModules(trackedOnly bool, includeIgnored bool, stagedOnly bool, rootPath string, version string) ([]RepositoryFileWithModule, error) {
	// Get repository root if not provided
	if rootPath == "" {
		var err error
		rootPath, err = GetRepositoryRoot("")
		if err != nil {
			return nil, err
		}
	}

	// Get all repository files (exclude Git internal files by default)
	files, err := GetRepositoryFiles(trackedOnly, includeIgnored, false, stagedOnly, rootPath)
	if err != nil {
		return nil, err
	}

	// Enrich with module ownership
	return EnrichFilesWithModules(files, rootPath, version)
}

// GetFilesByModule groups files by their owning module(s).
// Returns a map of module moniker -> list of file paths.
//
// Files with multiple owners will appear in multiple module lists.
// Files with no owners will not appear in the result.
//
// Example:
//
//	files, _ := repository.GetRepositoryFilesWithModules(true, false, "", "0.1.0")
//	byModule := repository.GetFilesByModule(files)
//	for module, paths := range byModule {
//	    fmt.Printf("%s: %d files\n", module, len(paths))
//	}
func GetFilesByModule(files []RepositoryFileWithModule) map[string][]string {
	result := make(map[string][]string)

	for _, file := range files {
		for _, module := range file.Modules {
			result[module] = append(result[module], file.Name)
		}
	}

	return result
}

// GetMultiOwnershipFiles returns files that belong to multiple modules.
// Useful for detecting overlapping module boundaries.
//
// Example:
//
//	files, _ := repository.GetRepositoryFilesWithModules(true, false, "", "0.1.0")
//	multiOwned := repository.GetMultiOwnershipFiles(files)
//	fmt.Printf("Found %d files with multiple owners\n", len(multiOwned))
//	for _, f := range multiOwned {
//	    fmt.Printf("  %s: %v\n", f.Name, f.Modules)
//	}
func GetMultiOwnershipFiles(files []RepositoryFileWithModule) []RepositoryFileWithModule {
	result := []RepositoryFileWithModule{}

	for _, file := range files {
		if len(file.Modules) > 1 {
			result = append(result, file)
		}
	}

	return result
}

// GetOrphanFiles returns files that don't belong to any module.
// Useful for finding files that aren't covered by module contracts.
//
// Example:
//
//	files, _ := repository.GetRepositoryFilesWithModules(true, false, "", "0.1.0")
//	orphans := repository.GetOrphanFiles(files)
//	fmt.Printf("Found %d orphan files\n", len(orphans))
//	for _, f := range orphans {
//	    fmt.Printf("  %s\n", f.Name)
//	}
func GetOrphanFiles(files []RepositoryFileWithModule) []RepositoryFileWithModule {
	result := []RepositoryFileWithModule{}

	for _, file := range files {
		if len(file.Modules) == 0 {
			result = append(result, file)
		}
	}

	return result
}
