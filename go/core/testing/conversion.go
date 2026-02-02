package testing

import (
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// ConvertToEntries converts TestReferences to SuiteTestEntries with full metadata.
// This is the canonical conversion function used by get/show commands.
//
// Parameters:
//   - tests: discovered and enriched test references
//   - fileModuleMap: maps file paths to module monikers (from git ownership)
//   - moduleRegistry: for looking up module types (can be nil)
//   - repoRoot: repository root path for path normalization
func ConvertToEntries(
	tests []TestReference,
	fileModuleMap map[string]string,
	moduleRegistry *modules.Registry,
	repoRoot string,
) []SuiteTestEntry {
	entries := make([]SuiteTestEntry, len(tests))

	for i, test := range tests {
		// Extract module from multiple sources (in priority order):
		// 1. Module dependencies from test tags (@depm:)
		// 2. File path inference from git ownership
		module := ""
		if len(test.ModuleDependencies) > 0 {
			module = test.ModuleDependencies[0]
		}
		if module == "" {
			module = extractModuleFromPath(test.FilePath, fileModuleMap, repoRoot)
		}

		// Look up the owning module's package types
		moduleType := ""
		if module != "" && moduleRegistry != nil {
			if mod, exists := moduleRegistry.Get(module); exists {
				moduleType = mod.GetComponentTypesDisplay()
			}
		}

		// Generate test moniker
		moniker := GenerateTestMoniker(test, module)

		// Extract package/feature from file path
		pkg := extractPackageFromPath(test.FilePath, repoRoot)

		// Extract tag categories
		levelTags := filterTagsByPrefix(test.Tags, "@L")
		verificationTags := filterTagsByPatterns(test.Tags, []string{"@ov", "@iv", "@pv", "@piv", "@ppv"})
		systemDeps := filterTagsByPrefix(test.Tags, "@deps:")
		moduleDeps := filterTagsByPrefix(test.Tags, "@depm:")

		entries[i] = SuiteTestEntry{
			Moniker:          moniker,
			TestName:         test.TestName,
			Type:             test.Type,
			FilePath:         test.FilePath,
			Package:          pkg,
			Module:           module,
			ModuleType:       moduleType,
			Level:            levelTags,
			Verification:     verificationTags,
			SystemDeps:       systemDeps,
			ModuleDeps:       moduleDeps,
			SourceTags:       test.SourceTags,
			InferredTags:     test.InferredTags,
			InferredDeps:     test.InferredDeps,
			InferredDepm:     test.InferredDepm,
			IsIgnored:        test.IsIgnored,
			SkipReason:       test.SkipReason,
			IsManual:         test.IsManual,
			IsSequential:     test.IsSequential,
			RiskControls:     test.RiskControls,
			IsGxP:            test.IsGxP,
			IsCriticalAspect: test.IsCriticalAspect,
		}
	}

	return entries
}

// extractPackageFromPath extracts the package/feature name from a file path.
// For specs: specs/MODULE/PACKAGE/... -> PACKAGE
// For go tests: go/eac/MODULE/PACKAGE/..._test.go -> PACKAGE.
func extractPackageFromPath(filePath, repoRoot string) string {
	normalized := normalizePathSeparators(filePath)
	repoRootNorm := normalizePathSeparators(repoRoot)

	// Get relative path
	relative := normalized
	if len(normalized) > len(repoRootNorm) && normalized[:len(repoRootNorm)] == repoRootNorm {
		relative = normalized[len(repoRootNorm):]
		if len(relative) > 0 && relative[0] == '/' {
			relative = relative[1:]
		}
	}

	parts := splitPath(relative)
	if len(parts) < 3 {
		return ""
	}

	// For specs/MODULE/PACKAGE/... -> return PACKAGE
	if parts[0] == "specs" && len(parts) >= 3 {
		return parts[2]
	}

	// For go/X/Y/PACKAGE/..._test.go -> look for test file and return parent
	// This is tricky - for now return the directory containing the test file
	if parts[0] == "go" && len(parts) >= 4 {
		// Return second-to-last directory before file
		return parts[len(parts)-2]
	}

	return ""
}

// splitPath splits a path by forward slashes.
func splitPath(path string) []string {
	parts := []string{}
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}

// extractModuleFromPath extracts module moniker from file path using the file-module map.
func extractModuleFromPath(filePath string, fileModuleMap map[string]string, repoRoot string) string {
	if fileModuleMap == nil {
		return ""
	}

	// Normalize path for lookup
	normalized := normalizePathSeparators(filePath)

	// Try direct lookup first
	if module, ok := fileModuleMap[normalized]; ok {
		return module
	}

	// Try relative path (remove repo root prefix)
	repoRootNorm := normalizePathSeparators(repoRoot)
	if len(normalized) > len(repoRootNorm) && normalized[:len(repoRootNorm)] == repoRootNorm {
		relative := normalized[len(repoRootNorm):]
		if len(relative) > 0 && relative[0] == '/' {
			relative = relative[1:]
		}
		if module, ok := fileModuleMap[relative]; ok {
			return module
		}
	}

	return ""
}

// normalizePathSeparators converts backslashes to forward slashes.
func normalizePathSeparators(path string) string {
	result := make([]byte, len(path))
	for i := 0; i < len(path); i++ {
		if path[i] == '\\' {
			result[i] = '/'
		} else {
			result[i] = path[i]
		}
	}
	return string(result)
}

// filterTagsByPrefix returns tags that start with the given prefix.
func filterTagsByPrefix(tags []string, prefix string) []string {
	result := []string{}
	for _, tag := range tags {
		if len(tag) >= len(prefix) && tag[:len(prefix)] == prefix {
			result = append(result, tag)
		}
	}
	return result
}

// filterTagsByPatterns returns tags that exactly match any of the given patterns.
func filterTagsByPatterns(tags, patterns []string) []string {
	result := []string{}
	for _, tag := range tags {
		for _, pattern := range patterns {
			if tag == pattern {
				result = append(result, tag)
				break
			}
		}
	}
	return result
}
