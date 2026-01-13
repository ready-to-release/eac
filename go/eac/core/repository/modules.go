package repository

import (
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// logModules is the package-level logger for module operations
var logModules = logging.C("modules")

// EnrichFilesWithModules takes a list of files and determines which module(s) own each file.
// Returns a list of files with their module ownership information.
//
// A file can belong to multiple modules if their glob patterns overlap.
//
// Parameters:
//   - files: List of FileInfo from GetRepositoryFiles
//   - workspaceRoot: Root directory of the workspace
//     (no version - repository config is unversioned)
//
// Returns:
//   - List of RepositoryFileWithModule with normalized paths and module ownership
//   - Error if module contracts cannot be loaded
//
// Example:
//
//	repo, _ := git.Open("/workspace")
//	files, _ := repository.GetRepositoryFiles(repo, true, false, false, false)
//	enriched, _ := repository.EnrichFilesWithModules(files, "/workspace")
//	for _, f := range enriched {
//	    fmt.Printf("%s -> %v\n", f.Name, f.Modules)
//	}
func EnrichFilesWithModules(files []FileInfo, workspaceRoot string) ([]RepositoryFileWithModule, error) {
	logModules.Debug("EnrichFilesWithModules: start")
	// Load module contracts
	logModules.Debug("EnrichFilesWithModules: calling modules.LoadFromWorkspace")
	registry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil, NewRepositoryError("enrich", workspaceRoot, err, "failed to load module contracts")
	}
	logModules.Debug("EnrichFilesWithModules: modules.LoadFromWorkspace complete")

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

		// Filter out repository-root modules when other modules match
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

// filterClosestModules filters a list of modules to only include the most specific ones.
// When multiple modules match a file, this excludes repository-root modules if other
// modules also match.
//
// Special handling for "repository-root" type modules:
//   - Repository-root modules should only own files that no other module claims
//   - If any non-repository-root modules match, repository-root is excluded
//   - Repository-root is only kept if it's the ONLY match
//
// Examples:
//   - [claude, repository] → [claude] (repository-root excluded)
//   - [repository] → [repository] (only match)
//   - [docs, eac-core] → [docs, eac-core] (both kept)
//
// Parameters:
//   - matchingModules: List of modules that match a file
//   - registry: Module registry (unused, kept for API compatibility)
//
// Returns:
//   - Filtered list containing only the most specific modules
func filterClosestModules(matchingModules []*modules.ModuleContract, registry *modules.Registry) []*modules.ModuleContract {
	// No filtering needed for 0 or 1 modules
	if len(matchingModules) <= 1 {
		return matchingModules
	}

	// Special handling for "repository-root" type modules
	// Repository-root modules should only own files that no other module claims
	hasRepositoryRoot := false
	hasOtherModules := false

	for _, module := range matchingModules {
		if module.Type == "repository-root" {
			hasRepositoryRoot = true
		} else {
			hasOtherModules = true
		}
	}

	// If repository-root is present with other modules, exclude it
	if hasRepositoryRoot && hasOtherModules {
		filteredModules := make([]*modules.ModuleContract, 0, len(matchingModules)-1)
		for _, module := range matchingModules {
			if module.Type != "repository-root" {
				filteredModules = append(filteredModules, module)
			}
		}
		return filteredModules
	}

	return matchingModules
}

// GetRepositoryFilesWithModules is a convenience function that combines
// GetRepositoryFiles and EnrichFilesWithModules in one call.
//
// Parameters:
//   - repo: GitRepository interface for git operations
//   - trackedOnly: if true, only return files tracked by Git
//   - includeIgnored: if true, include files ignored by .gitignore
//   - stagedOnly: if true, only return files currently staged in Git index
//     (no version - repository config is unversioned)
//
// Returns:
//   - List of files with module ownership information
//   - Error if repository operations or module loading fails
//
// Example:
//
//	repo, _ := git.Open("/workspace")
//	files, err := repository.GetRepositoryFilesWithModules(repo, true, false, false)
//	for _, f := range files {
//	    if len(f.Modules) > 1 {
//	        fmt.Printf("Multi-ownership: %s -> %v\n", f.Name, f.Modules)
//	    }
//	}
func GetRepositoryFilesWithModules(repo git.GitRepository, trackedOnly, includeIgnored, stagedOnly bool) ([]RepositoryFileWithModule, error) {
	logModules.Debug("GetRepositoryFilesWithModules: start")
	rootPath := repo.RootPath()

	// Get all repository files (exclude Git internal files by default)
	logModules.Debug("GetRepositoryFilesWithModules: calling GetRepositoryFiles")
	files, err := GetRepositoryFiles(repo, trackedOnly, includeIgnored, false, stagedOnly)
	if err != nil {
		return nil, err
	}
	logModules.Debug("GetRepositoryFilesWithModules: GetRepositoryFiles complete")

	// Enrich with module ownership
	logModules.Debug("GetRepositoryFilesWithModules: calling EnrichFilesWithModules")
	result, err := EnrichFilesWithModules(files, rootPath)
	logModules.Debug("GetRepositoryFilesWithModules: EnrichFilesWithModules complete")
	return result, err
}

// GetFilesByModule groups files by their owning module(s).
// Returns a map of module moniker -> list of file paths.
//
// Files with multiple owners will appear in multiple module lists.
// Files with no owners will not appear in the result.
//
// Example:
//
//	repo, _ := git.Open("/workspace")
//	files, _ := repository.GetRepositoryFilesWithModules(repo, true, false, false, "0.1.0")
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
//	files, _ := repository.GetRepositoryFilesWithModules(true, false, "")
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
//	files, _ := repository.GetRepositoryFilesWithModules(true, false, "")
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
