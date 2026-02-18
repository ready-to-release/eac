// module_mapping.go - Maps package paths to module monikers for test output organization
package test

import (
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/testrunners"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// ModuleMapper provides file-to-module mapping functionality.
type ModuleMapper struct {
	registry      *modules.Registry
	workspaceRoot string
}

// NewModuleMapper creates a mapper from an existing module registry.
// The registry should have all component type defaults applied.
func NewModuleMapper(registry *modules.Registry, workspaceRoot string) *ModuleMapper {
	return &ModuleMapper{
		registry:      registry,
		workspaceRoot: workspaceRoot,
	}
}

// GetModuleForFile returns the module moniker that owns the given file path.
// Returns empty string if no module owns the file.
func (m *ModuleMapper) GetModuleForFile(filePath string) string {
	// Normalize to relative path
	relPath := filePath
	if filepath.IsAbs(filePath) {
		rel, err := filepath.Rel(m.workspaceRoot, filePath)
		if err == nil {
			relPath = rel
		}
	}
	relPath = filepath.ToSlash(relPath)

	// Find modules that own this file
	matches := m.registry.FindModulesForFile(relPath)
	if len(matches) > 0 {
		return matches[0].Moniker
	}

	// For directory paths, check if path is under any component root
	// This handles test directories that don't match specific file patterns
	if moniker := m.findModuleByComponentRoot(relPath); moniker != "" {
		return moniker
	}

	// Fallback: handle specs directory by convention (<specsRoot>/<moniker>/**)
	// This maps spec files to their module even if the module doesn't have a specs: component
	specsRoot := "specs"
	if cfg := config.Global(); cfg != nil {
		specsRoot = cfg.Repository.Paths.SpecsRoot
	}
	specsPrefix := specsRoot + "/"
	if strings.HasPrefix(relPath, specsPrefix) {
		afterSpecs := strings.TrimPrefix(relPath, specsPrefix)
		if idx := strings.Index(afterSpecs, "/"); idx > 0 {
			potentialMoniker := afterSpecs[:idx]
			if _, exists := m.registry.Get(potentialMoniker); exists {
				return potentialMoniker
			}
		}
	}

	return ""
}

// findModuleByComponentRoot finds the module that owns a path by checking component roots.
// This handles directory paths that don't match specific file patterns but are under a component root.
// Uses longest-prefix matching for deterministic results when multiple modules share a root.
// On ties, the alphabetically first moniker wins.
func (m *ModuleMapper) findModuleByComponentRoot(path string) string {
	var bestMoniker string
	var bestRootLen int
	for _, module := range m.registry.All() {
		for _, root := range module.GetComponentRoots() {
			if root == "" || root == "/" {
				continue
			}
			root = filepath.ToSlash(root)
			if strings.HasPrefix(path, root+"/") || path == root {
				rootLen := len(root)
				if rootLen > bestRootLen || (rootLen == bestRootLen && (bestMoniker == "" || module.Moniker < bestMoniker)) {
					bestRootLen = rootLen
					bestMoniker = module.Moniker
				}
			}
		}
	}
	return bestMoniker
}

// GetModuleForPackagePath returns the module moniker for a package path.
// Package paths are like "go/eac/core/contracts" or for BDD:
// "featureName:testRoot:featurePath" (using the BDDPackagePath protocol).
// Returns empty string if no module found.
func (m *ModuleMapper) GetModuleForPackagePath(pkgPath string) string {
	p := testrunners.ParseBDDPackagePath(pkgPath)

	// Runner file convention from godog descriptor for synthetic file matching
	runnerFile := ""
	for _, desc := range testrunners.AllDescriptors() {
		if desc.RunnerFileConvention != "" {
			runnerFile = desc.RunnerFileConvention
			break
		}
	}

	if p.IsBDD() {
		// BDD format: prefer featurePath for module lookup.
		// The spec file unambiguously determines module ownership, while the
		// testRoot may reside in a different module.
		if moniker := m.GetModuleForFile(p.FeaturePath); moniker != "" {
			return moniker
		}
		// Fall back to testRoot if feature path didn't resolve
		if moniker := m.GetModuleForFile(p.TestRoot); moniker != "" {
			return moniker
		}
		if runnerFile != "" {
			if moniker := m.GetModuleForFile(p.TestRoot + "/" + runnerFile); moniker != "" {
				return moniker
			}
		}
		return ""
	}

	// Unit test path (single part)
	if moniker := m.GetModuleForFile(p.TestRoot); moniker != "" {
		return moniker
	}

	// Try with synthetic runner file to trigger pattern matching
	if runnerFile != "" {
		if moniker := m.GetModuleForFile(p.TestRoot + "/" + runnerFile); moniker != "" {
			return moniker
		}
	}

	return ""
}

// BuildModuleOutputPath constructs the output path for a module's test results.
// Returns a path like "<module-moniker>/packages/<package-suffix>" or "<module-moniker>/packages/<feature-name>" for BDD.
// The structure is: out/test/<module>/packages/<package>/.
func (m *ModuleMapper) BuildModuleOutputPath(pkgPath, moduleMoniker string) string {
	if moduleMoniker == "" {
		return filepath.ToSlash(filepath.Join("unknown", "packages", sanitizePathForLog(pkgPath)))
	}

	p := testrunners.ParseBDDPackagePath(pkgPath)
	if p.IsBDD() {
		return filepath.ToSlash(filepath.Join(moduleMoniker, "packages", p.FeatureName))
	}

	// Unit test: extract the suffix after the module's component root
	normalizedPath := filepath.ToSlash(pkgPath)
	suffix := extractPathSuffix(normalizedPath, moduleMoniker)

	result := filepath.Join(moduleMoniker, "packages")
	if suffix != "" {
		result = filepath.Join(result, suffix)
	} else {
		result = filepath.Join(result, "root")
	}

	return filepath.ToSlash(result)
}

// extractPathSuffix extracts the path suffix after the module identifier.
func extractPathSuffix(pkgPath, moduleMoniker string) string {
	// Map module monikers to their typical path patterns
	// core -> go/eac/core
	// eac-cli -> go/cli/eac
	// clie -> go/cli/clie

	parts := strings.Split(moduleMoniker, "-")
	if len(parts) >= 2 {
		prefix := parts[0]                         // "eac" or "clie"
		modulePart := strings.Join(parts[1:], "-") // "core", "commands", etc.

		// Try to find "go/<prefix>/<module>/" in the path
		searchPattern := "go/" + prefix + "/" + modulePart + "/"
		if idx := strings.Index(pkgPath, searchPattern); idx >= 0 {
			return pkgPath[idx+len(searchPattern):]
		}

		// Try without trailing slash (for exact match)
		searchPattern = "go/" + prefix + "/" + modulePart
		if strings.HasSuffix(pkgPath, searchPattern) {
			return ""
		}
	}

	// For specs implementations, extract the feature path
	if strings.Contains(pkgPath, "/specs/impl/") {
		if idx := strings.Index(pkgPath, "/specs/impl/"); idx >= 0 {
			afterImpl := pkgPath[idx+len("/specs/impl/"):]
			// Skip the moniker directory
			if slashIdx := strings.Index(afterImpl, "/"); slashIdx >= 0 {
				return afterImpl[slashIdx+1:]
			}
		}
	}

	return ""
}
