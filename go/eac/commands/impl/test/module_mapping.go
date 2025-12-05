// module_mapping.go - Maps package paths to module monikers for test output organization
package test

import (
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

// ModuleMapper provides file-to-module mapping functionality
type ModuleMapper struct {
	registry      *modules.Registry
	workspaceRoot string
}

// NewModuleMapper creates a mapper from EAC config
func NewModuleMapper(eacCfg *config.EACConfig, workspaceRoot string) *ModuleMapper {
	registry := modules.NewRegistry("0.1.0", workspaceRoot)

	// Convert config modules to contract modules
	for _, m := range eacCfg.Modules.Modules {
		base := contracts.BaseContract{
			Moniker:     m.Moniker,
			Name:        m.Name,
			Type:        m.Type,
			Description: m.Description,
			DependsOn:   m.DependsOn,
			Files: contracts.Files{
				Root:      m.Files.Root,
				Source:    m.Files.Source,
				Config:    m.Files.Config,
				Assets:    m.Files.Assets,
				Tests:     m.Files.Tests,
				Exclude:   m.Files.Exclude,
				Changelog: m.Files.Changelog,
				Repo: contracts.RepoPatterns{
					Specs:    m.Files.Repo.Specs,
					TestImpl: m.Files.Repo.TestImpl,
					Design:   m.Files.Repo.Design,
					Other:    m.Files.Repo.Other,
					Exclude:  m.Files.Repo.Exclude,
				},
			},
		}
		contract := modules.NewModuleContract(base, workspaceRoot)
		_ = registry.Add(contract)
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

	// Try to find a file in this directory to determine ownership
	// Use a synthetic test file path
	testFilePath := actualPath + "/*_test.go"

	// First try the directory itself
	if moniker := m.GetModuleForFile(actualPath); moniker != "" {
		return moniker
	}

	// Try with test file extension
	if moniker := m.GetModuleForFile(testFilePath); moniker != "" {
		return moniker
	}

	// Fallback to heuristic extraction
	return extractModuleMonikerFromPath(actualPath)
}

// extractModuleMonikerFromPath extracts module moniker from package path using heuristics.
// This is a fallback when registry lookup fails.
func extractModuleMonikerFromPath(pkgPath string) string {
	// Normalize path
	normalizedPath := filepath.ToSlash(pkgPath)

	// Handle godog paths: "go/eac/specs/impl/eac-commands:specs/..."
	if idx := strings.Index(normalizedPath, ":"); idx >= 0 {
		normalizedPath = normalizedPath[:idx]
	}

	// Check for Go module paths: go/eac/<module>/... or go/r2r/<module>/...
	for _, boundary := range []string{"go/eac/", "go/r2r/"} {
		idx := strings.Index(normalizedPath, boundary)
		if idx >= 0 {
			relativePath := normalizedPath[idx+len(boundary):]
			parts := strings.Split(relativePath, "/")
			if len(parts) >= 1 && parts[0] != "" {
				prefix := "eac"
				if boundary == "go/r2r/" {
					prefix = "r2r"
				}

				// Special case: specs/impl -> use directory name after impl
				if len(parts) >= 3 && parts[0] == "specs" && parts[1] == "impl" {
					return parts[2] // e.g., "eac-commands" from "specs/impl/eac-commands"
				}

				return prefix + "-" + parts[0]
			}
		}
	}

	// Check for TypeScript paths: typescript/<module>/...
	if idx := strings.Index(normalizedPath, "typescript/"); idx >= 0 {
		relativePath := normalizedPath[idx+len("typescript/"):]
		parts := strings.Split(relativePath, "/")
		if len(parts) >= 1 && parts[0] != "" {
			return parts[0]
		}
	}

	// Check for specs paths: specs/<module>/...
	if idx := strings.Index(normalizedPath, "specs/"); idx >= 0 {
		relativePath := normalizedPath[idx+len("specs/"):]
		parts := strings.Split(relativePath, "/")
		if len(parts) >= 1 && parts[0] != "" {
			return parts[0] + "-specs"
		}
	}

	// Fallback: use the package path itself
	return strings.ReplaceAll(normalizedPath, "/", "-")
}

// BuildModuleOutputPath constructs the output path for a module's test results.
// Returns a path like "<module-moniker>/<package-suffix>" or "<module-moniker>/<feature-name>" for godog
func (m *ModuleMapper) BuildModuleOutputPath(pkgPath, moduleMoniker string) string {
	if moduleMoniker == "" {
		// No module found, use sanitized package path
		return sanitizePathForLog(pkgPath)
	}

	// Handle godog paths: "featureName:testRoot:featurePath"
	parts := strings.SplitN(pkgPath, ":", 3)
	if len(parts) == 3 {
		// Godog format: use module/featureName for cleaner output
		featureName := parts[0]
		return filepath.ToSlash(filepath.Join(moduleMoniker, featureName))
	}

	// Extract the suffix after the module's directory
	// e.g., for "go/eac/core/contracts" and module "eac-core", return "eac-core/contracts"
	normalizedPath := filepath.ToSlash(pkgPath)

	// Handle legacy two-part paths
	featureSuffix := ""
	if idx := strings.Index(normalizedPath, ":"); idx >= 0 {
		featureSuffix = normalizedPath[idx+1:]
		normalizedPath = normalizedPath[:idx]
	}

	// Try to find the module root in the path
	suffix := extractPathSuffix(normalizedPath, moduleMoniker)

	result := moduleMoniker
	if suffix != "" {
		result = filepath.Join(moduleMoniker, suffix)
	}
	if featureSuffix != "" {
		result = filepath.Join(result, sanitizePathForLog(featureSuffix))
	}

	return filepath.ToSlash(result)
}

// extractPathSuffix extracts the path suffix after the module identifier
func extractPathSuffix(pkgPath, moduleMoniker string) string {
	// Map module monikers to their typical path patterns
	// eac-core -> go/eac/core
	// eac-commands -> go/eac/commands
	// r2r-cli -> go/r2r/cli

	parts := strings.Split(moduleMoniker, "-")
	if len(parts) >= 2 {
		prefix := parts[0] // "eac" or "r2r"
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
