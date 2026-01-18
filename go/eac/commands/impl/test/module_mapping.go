// module_mapping.go - Maps package paths to module monikers for test output organization
package test

import (
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

// ModuleMapper provides file-to-module mapping functionality.
type ModuleMapper struct {
	registry      *modules.Registry
	workspaceRoot string
}

// NewModuleMapper creates a mapper from EAC config.
func NewModuleMapper(eacCfg *config.EACConfig, workspaceRoot string) *ModuleMapper {
	registry := modules.NewRegistry("0.1.0", workspaceRoot)

	// Convert config modules to contract modules
	for i := range eacCfg.Repository.Modules {
		m := &eacCfg.Repository.Modules[i]

		// Convert config.ModuleComponents to contracts.ModuleComponents
		packages := make(contracts.ModuleComponents)
		for pkgName, entry := range m.Components {
			if entry == nil {
				packages[pkgName] = nil
			} else {
				packages[pkgName] = &contracts.ComponentEntry{
					Root: entry.Root,
				}
			}
		}

		base := contracts.BaseContract{
			Moniker:     m.Moniker,
			Name:        m.Name,
			Components:  packages,
			Description: m.Description,
			DependsOn:   m.DependsOn,
		}
		contract := modules.NewModuleContract(base, workspaceRoot)
		_ = registry.Add(contract) //nolint:errcheck // registry.Add only fails on duplicate which won't happen here
	}

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

	// Try with a synthetic file to handle directory paths
	if !strings.HasSuffix(relPath, ".go") && !strings.HasSuffix(relPath, ".feature") {
		// This looks like a directory path - try with a synthetic Go file
		syntheticFile := relPath + "/test.go"
		matches = m.registry.FindModulesForFile(syntheticFile)
		if len(matches) > 0 {
			return matches[0].Moniker
		}
	}

	return ""
}

// GetModuleForPackagePath returns the module moniker for a package path.
// Package paths are like "go/eac/core/contracts" or for godog:
// "featureName:go/eac/specs/impl/eac-commands:specs/eac-commands/..."
// Returns empty string if no module found - no fallback heuristics.
func (m *ModuleMapper) GetModuleForPackagePath(pkgPath string) string {
	// Handle godog BDD paths: "featureName:testRoot:featurePath"
	// The testRoot (second part) is where the test runner lives and determines the module
	parts := strings.SplitN(pkgPath, ":", 3)
	var actualPath string
	if len(parts) == 3 {
		// Godog format: use testRoot (second part) for module lookup
		actualPath = parts[1]
	} else if len(parts) == 2 {
		// Legacy format: "path:featurePath"
		actualPath = parts[0]
	} else {
		actualPath = pkgPath
	}

	// Try to find module via registry lookup
	if moniker := m.GetModuleForFile(actualPath); moniker != "" {
		return moniker
	}

	// Try with synthetic test file
	if moniker := m.GetModuleForFile(actualPath + "/*_test.go"); moniker != "" {
		return moniker
	}

	// No fallback - return empty string
	return ""
}

// BuildModuleOutputPath constructs the output path for a module's test results.
// Returns a path like "<module-moniker>/packages/<package-suffix>" or "<module-moniker>/packages/<feature-name>" for godog
// The new structure is: out/test/<module>/packages/<package>/.
func (m *ModuleMapper) BuildModuleOutputPath(pkgPath, moduleMoniker string) string {
	if moduleMoniker == "" {
		// No module found, use sanitized package path under "unknown" module
		return filepath.ToSlash(filepath.Join("unknown", "packages", sanitizePathForLog(pkgPath)))
	}

	// Handle godog paths: "featureName:testRoot:featurePath"
	parts := strings.SplitN(pkgPath, ":", 3)
	if len(parts) == 3 {
		// Godog format: use module/packages/featureName for cleaner output
		featureName := parts[0]
		return filepath.ToSlash(filepath.Join(moduleMoniker, "packages", featureName))
	}

	// Extract the suffix after the module's directory
	// e.g., for "go/eac/core/contracts" and module "eac-core", return "eac-core/packages/contracts"
	normalizedPath := filepath.ToSlash(pkgPath)

	// Handle legacy two-part paths
	featureSuffix := ""
	if idx := strings.Index(normalizedPath, ":"); idx >= 0 {
		featureSuffix = normalizedPath[idx+1:]
		normalizedPath = normalizedPath[:idx]
	}

	// Try to find the module root in the path
	suffix := extractPathSuffix(normalizedPath, moduleMoniker)

	// Build path with packages/ directory
	result := filepath.Join(moduleMoniker, "packages")
	if suffix != "" {
		result = filepath.Join(result, suffix)
	} else {
		// If no suffix, use "root" to avoid empty path
		result = filepath.Join(result, "root")
	}
	if featureSuffix != "" {
		result = filepath.Join(result, sanitizePathForLog(featureSuffix))
	}

	return filepath.ToSlash(result)
}

// extractPathSuffix extracts the path suffix after the module identifier.
func extractPathSuffix(pkgPath, moduleMoniker string) string {
	// Map module monikers to their typical path patterns
	// eac-core -> go/eac/core
	// eac-commands -> go/eac/commands
	// r2r-cli -> go/r2r/cli

	parts := strings.Split(moduleMoniker, "-")
	if len(parts) >= 2 {
		prefix := parts[0]                         // "eac" or "r2r"
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
