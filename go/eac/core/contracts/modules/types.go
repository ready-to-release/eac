package modules

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
)

// ModuleContract represents a deployable module contract
type ModuleContract struct {
	contracts.BaseContract `yaml:",inline"`

	// Cached computed values
	workspaceRoot string
}

// NewModuleContract creates a new module contract with workspace context
func NewModuleContract(base contracts.BaseContract, workspaceRoot string) *ModuleContract {
	return &ModuleContract{
		BaseContract:  base,
		workspaceRoot: workspaceRoot,
	}
}

// getRelativePatterns returns all patterns relative to files.root
func (m *ModuleContract) getRelativePatterns() []string {
	var patterns []string
	patterns = append(patterns, m.Files.Source...)
	patterns = append(patterns, m.Files.Config...)
	patterns = append(patterns, m.Files.Assets...)
	patterns = append(patterns, m.Files.Tests...)

	// Include changelog file for ownership
	if m.Files.Changelog != "" {
		patterns = append(patterns, m.Files.Changelog)
	}

	return patterns
}

// hasExplicitPatterns returns true if the module has user-specified patterns.
// This is used to determine whether to fall back to "all files under root" matching.
// Default files like Changelog don't count as explicit patterns.
func (m *ModuleContract) hasExplicitPatterns() bool {
	return len(m.Files.Source) > 0 ||
		len(m.Files.Config) > 0 ||
		len(m.Files.Assets) > 0 ||
		len(m.Files.Tests) > 0
}

// getRepoPatterns returns all patterns relative to repo root
// Includes specs, other, workflows, and converts test_impl/design paths to glob patterns
func (m *ModuleContract) getRepoPatterns() []string {
	var patterns []string
	patterns = append(patterns, m.Files.Repo.Specs...)
	patterns = append(patterns, m.Files.Repo.Other...)

	// Include workflow files for ownership
	if m.Files.Workflows.CI != "" {
		patterns = append(patterns, m.Files.Workflows.CI)
	}
	if m.Files.Workflows.Release != "" {
		patterns = append(patterns, m.Files.Workflows.Release)
	}

	// Convert test_impl directory path to glob pattern for file ownership
	if m.Files.Repo.TestImpl != "" {
		patterns = append(patterns, m.Files.Repo.TestImpl+"/**")
	}

	// Convert design directory path to glob pattern for file ownership
	if m.Files.Repo.Design != "" {
		patterns = append(patterns, m.Files.Repo.Design+"/**")
	}

	return patterns
}

// GetGlobPatterns returns GitHub Actions compatible glob patterns for this module
// These patterns can be used in workflow path filters
func (m *ModuleContract) GetGlobPatterns() []string {
	var patterns []string

	// Add relative patterns with root prefix
	root := normalizePathSeparators(m.Files.Root)
	for _, pattern := range m.getRelativePatterns() {
		pattern = normalizePathSeparators(pattern)
		if root != "" && root != "/" {
			patterns = append(patterns, root+"/"+pattern)
		} else {
			patterns = append(patterns, pattern)
		}
	}

	// Add repo-root patterns as-is
	for _, pattern := range m.getRepoPatterns() {
		patterns = append(patterns, normalizePathSeparators(pattern))
	}

	return patterns
}

// GetAbsolutePaths returns absolute file system paths for this module's sources
func (m *ModuleContract) GetAbsolutePaths() []string {
	if m.workspaceRoot == "" {
		return []string{}
	}

	var paths []string
	for _, pattern := range m.getRelativePatterns() {
		if m.Files.Root != "" {
			paths = append(paths, filepath.Join(m.workspaceRoot, m.Files.Root, pattern))
		} else {
			paths = append(paths, filepath.Join(m.workspaceRoot, pattern))
		}
	}

	return paths
}

// MatchesFile returns true if the given file path matches this module's patterns
func (m *ModuleContract) MatchesFile(filePath string) bool {
	path := normalizePathSeparators(filePath)
	root := normalizePathSeparators(m.Files.Root)

	// 1. Quick check: is file under module root?
	isUnderRoot := root == "" || root == "/" || strings.HasPrefix(path, root+"/") || path == root

	if isUnderRoot {
		// 2a. If no explicit patterns defined, root ownership implies all files under root match
		// Default files like Changelog don't count as explicit patterns.
		// This fallback is disabled when flags.explicit_ownership is true.
		if !m.hasExplicitPatterns() && !m.Flags.ExplicitOwnership && root != "" && root != "/" {
			return !m.isExcluded(path)
		}

		// 2b. File is under root - check relative patterns (most common case)
		relativePatterns := m.getRelativePatterns()
		for _, pattern := range relativePatterns {
			pattern = normalizePathSeparators(pattern)

			// Build full pattern with root
			var fullPattern string
			if root != "" && root != "/" {
				fullPattern = root + "/" + pattern
			} else {
				fullPattern = pattern
			}

			if matchWithFallback(path, fullPattern) {
				return !m.isExcluded(path)
			}
		}
	}

	// 2c. Check repo-root patterns (specs, other) - these match files anywhere
	// Only check if file didn't match relative patterns above
	for _, pattern := range m.getRepoPatterns() {
		if matchWithFallback(path, normalizePathSeparators(pattern)) {
			return !m.isExcluded(path)
		}
	}

	return false
}

// isExcluded checks if a path matches any exclude pattern
func (m *ModuleContract) isExcluded(path string) bool {
	root := normalizePathSeparators(m.Files.Root)

	// Check relative excludes
	for _, exclude := range m.Files.Exclude {
		exclude = normalizePathSeparators(exclude)
		var fullPattern string
		if root != "" && root != "/" {
			fullPattern = root + "/" + exclude
		} else {
			fullPattern = exclude
		}
		if matchWithFallback(path, fullPattern) {
			return true
		}
	}

	// Check repo-root excludes
	for _, exclude := range m.Files.Repo.Exclude {
		if matchWithFallback(path, normalizePathSeparators(exclude)) {
			return true
		}
	}

	return false
}

// matchWithFallback handles glob matching with fallback for ** patterns
func matchWithFallback(path, pattern string) bool {
	if matchGlobPattern(path, pattern) {
		return true
	}

	// Handle ** patterns that don't match root-level files
	if strings.HasPrefix(pattern, "**/") {
		// Try without **/ prefix
		if matchGlobPattern(path, strings.TrimPrefix(pattern, "**/")) {
			return true
		}
	} else if strings.Contains(pattern, "/**/") {
		// Handle patterns like "root/**/file.go"
		// Split and try matching the suffix
		parts := strings.SplitN(pattern, "/**/", 2)
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]
			// Try matching as if ** matches zero directories
			directPattern := prefix + "/" + suffix
			if matchGlobPattern(path, directPattern) {
				return true
			}
		}
	} else if strings.HasPrefix(pattern, "**") && !strings.HasPrefix(pattern, "**/") {
		// Handle patterns like "**.go"
		if matchGlobPattern(path, strings.TrimPrefix(pattern, "**")) {
			return true
		}
	}

	return false
}

// GetDependencies returns the list of module dependencies
func (m *ModuleContract) GetDependencies() []string {
	return m.DependsOn
}

// IsDefinitionsFile returns true if this contract represents a definitions file
func (m *ModuleContract) IsDefinitionsFile() bool {
	return m.Moniker == "definitions" || m.Type == "definitions-type"
}

// GetSpecsRoot returns the specs root directory for this module
// Uses the first pattern from Files.Repo.Specs, stripping the /** suffix
func (m *ModuleContract) GetSpecsRoot() string {
	if len(m.Files.Repo.Specs) > 0 {
		// Strip /** or /* suffix if present
		pattern := m.Files.Repo.Specs[0]
		pattern = strings.TrimSuffix(pattern, "/**")
		pattern = strings.TrimSuffix(pattern, "/*")
		return pattern
	}
	// Default: specs/<module-moniker>
	return filepath.Join("specs", m.Moniker)
}

// GetChangelogPath returns the changelog file path
func (m *ModuleContract) GetChangelogPath() string {
	if m.Files.Changelog != "" {
		if m.Files.Root != "" && m.Files.Root != "/" {
			return filepath.Join(m.Files.Root, m.Files.Changelog)
		}
		return m.Files.Changelog
	}
	// Default
	if m.Files.Root != "" && m.Files.Root != "/" {
		return filepath.Join(m.Files.Root, "CHANGELOG.md")
	}
	return "CHANGELOG.md"
}

// GetReleaseNotesPath returns the path to RELEASE-NOTES.md for this module
// Defaults to release/{moniker}/RELEASE-NOTES.md if not explicitly set
func (m *ModuleContract) GetReleaseNotesPath() string {
	if m.Files.ReleaseNotes != "" {
		if m.Files.Root != "" && m.Files.Root != "/" {
			return filepath.Join(m.Files.Root, m.Files.ReleaseNotes)
		}
		return m.Files.ReleaseNotes
	}
	// Default: release/{moniker}/RELEASE-NOTES.md
	return filepath.Join("release", m.Moniker, "RELEASE-NOTES.md")
}

// GetTestImplementationPath returns the test implementation directory path.
// This is where godog step definitions and test implementations live.
// Returns empty string if not defined in the module contract.
func (m *ModuleContract) GetTestImplementationPath() string {
	return m.Files.Repo.TestImpl
}

// GetDesignPath returns the design workspace directory path.
// This is where Structurizr DSL files (workspace.dsl) are stored.
// Returns default path specs/<moniker>/.design if not explicitly defined.
func (m *ModuleContract) GetDesignPath() string {
	if m.Files.Repo.Design != "" {
		return m.Files.Repo.Design
	}
	// Default: specs/<module-moniker>/.design
	return filepath.Join("specs", m.Moniker, ".design")
}

// normalizePathSeparators converts Windows backslashes to forward slashes
// GitHub Actions and glob patterns use forward slashes
func normalizePathSeparators(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// matchGlobPattern performs glob pattern matching using github.com/bmatcuk/doublestar
// Supports full glob syntax including:
// - ** (any depth of directories)
// - * (any characters within a segment)
// - ? (single character)
// - [abc] (character classes)
// - {a,b,c} (alternation/brace expansion)
func matchGlobPattern(path, pattern string) bool {
	// Normalize both paths to use forward slashes
	path = normalizePathSeparators(path)
	pattern = normalizePathSeparators(pattern)

	// Match the path against the pattern
	matched, _ := doublestar.Match(pattern, path)
	return matched
}
