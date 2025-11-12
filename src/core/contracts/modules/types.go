package modules

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/gobwas/glob"
	"github.com/ready-to-release/eac/src/core/contracts"
)

// ModuleContract represents a module deployable unit contract
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

// GetGlobPatterns returns GitHub Actions compatible glob patterns for this module
// These patterns can be used in workflow path filters
func (m *ModuleContract) GetGlobPatterns() []string {
	var patterns []string

	// Use patterns from contract source includes
	for _, include := range m.Source.Includes {
		// If pattern starts with "/", it's an absolute path from repository root
		// Strip the leading "/" and use as-is
		if strings.HasPrefix(include, "/") {
			patterns = append(patterns, normalizePathSeparators(strings.TrimPrefix(include, "/")))
		} else if strings.HasPrefix(include, m.Source.Root) {
			// If pattern already starts with root, it's already absolute
			patterns = append(patterns, normalizePathSeparators(include))
		} else if m.Source.Root != "" && m.Source.Root != "/" {
			// Combine root with pattern (skip if root is "/" which means repository root)
			combined := filepath.Join(m.Source.Root, include)
			patterns = append(patterns, normalizePathSeparators(combined))
		} else {
			// Root is empty or "/" - use pattern as-is
			patterns = append(patterns, normalizePathSeparators(include))
		}
	}

	return patterns
}

// GetAbsolutePaths returns absolute file system paths for this module's sources
func (m *ModuleContract) GetAbsolutePaths() []string {
	if m.workspaceRoot == "" {
		return []string{}
	}

	var paths []string
	for _, include := range m.Source.Includes {
		if m.Source.Root != "" {
			paths = append(paths, filepath.Join(m.workspaceRoot, m.Source.Root, include))
		} else {
			paths = append(paths, filepath.Join(m.workspaceRoot, include))
		}
	}

	return paths
}

// MatchesFile returns true if the given file path matches this module's patterns
func (m *ModuleContract) MatchesFile(filePath string) bool {
	// Normalize the file path
	normalizedPath := normalizePathSeparators(filePath)

	// Check each source pattern individually
	for _, include := range m.Source.Includes {
		// If pattern starts with "/", it's absolute from repository root
		// Match against it directly without root prefix check
		if strings.HasPrefix(include, "/") {
			absolutePattern := strings.TrimPrefix(include, "/")
			if matchGlobPattern(normalizedPath, absolutePattern) {
				return true
			}
			// Also handle ** patterns for absolute paths
			if strings.HasPrefix(absolutePattern, "**/") {
				patternWithoutDoubleStar := strings.TrimPrefix(absolutePattern, "**/")
				if matchGlobPattern(normalizedPath, patternWithoutDoubleStar) {
					return true
				}
			}
			continue
		}
	}

	// Check if file is under module root (skip check if root is "/" which means repository root)
	if m.Source.Root != "" && m.Source.Root != "/" && !strings.HasPrefix(normalizedPath, m.Source.Root) {
		return false
	}

	// Check against source includes patterns (non-absolute)
	for _, pattern := range m.GetGlobPatterns() {
		if matchGlobPattern(normalizedPath, pattern) {
			return true
		}
	}

	// Handle patterns that start with ** - they don't match root-level files
	// due to gobwas/glob behavior, so we need to also check without the **/ prefix
	for _, include := range m.Source.Includes {
		// Skip absolute patterns (already handled above)
		if strings.HasPrefix(include, "/") {
			continue
		}

		if strings.HasPrefix(include, "**/") {
			// Try matching without the **/ prefix for root-level files
			patternWithoutDoubleStar := strings.TrimPrefix(include, "**/")

			// Build the full pattern with module root
			var fullPattern string
			if m.Source.Root != "" && m.Source.Root != "/" {
				// Ensure we don't create double slashes
				rootNormalized := strings.TrimSuffix(m.Source.Root, "/")
				fullPattern = rootNormalized + "/" + patternWithoutDoubleStar
			} else {
				// Root is empty or "/" - use pattern as-is
				fullPattern = patternWithoutDoubleStar
			}

			if matchGlobPattern(normalizedPath, fullPattern) {
				return true
			}
		} else if strings.HasPrefix(include, "**") && !strings.HasPrefix(include, "**/") {
			// Handle patterns like "**/*.go" by also checking "*.go"
			patternWithoutDoubleStar := strings.TrimPrefix(include, "**")

			var fullPattern string
			if m.Source.Root != "" && m.Source.Root != "/" {
				// Ensure we don't create double slashes
				rootNormalized := strings.TrimSuffix(m.Source.Root, "/")
				fullPattern = rootNormalized + patternWithoutDoubleStar
			} else {
				// Root is empty or "/" - use pattern as-is
				fullPattern = patternWithoutDoubleStar
			}

			if matchGlobPattern(normalizedPath, fullPattern) {
				return true
			}
		}
	}

	return false
}

// GetDependencies returns the list of module dependencies
func (m *ModuleContract) GetDependencies() []string {
	return m.DependsOn
}

// GetUsedBy returns the list of modules that depend on this module
func (m *ModuleContract) GetUsedBy() []string {
	return m.UsedBy
}

// IsDefinitionsFile returns true if this contract represents a definitions file
func (m *ModuleContract) IsDefinitionsFile() bool {
	return m.Moniker == "definitions" || m.Type == "definitions-type"
}

// normalizePathSeparators converts Windows backslashes to forward slashes
// GitHub Actions and glob patterns use forward slashes
func normalizePathSeparators(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// Glob pattern cache for performance
var (
	globCache      = make(map[string]glob.Glob)
	globCacheMutex sync.RWMutex
)

// matchGlobPattern performs glob pattern matching using github.com/gobwas/glob
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

	// Try to get compiled glob from cache
	globCacheMutex.RLock()
	g, exists := globCache[pattern]
	globCacheMutex.RUnlock()

	if !exists {
		// Compile the pattern
		var err error
		g, err = glob.Compile(pattern, '/')
		if err != nil {
			// If pattern is invalid, return false
			return false
		}

		// Cache the compiled pattern
		globCacheMutex.Lock()
		globCache[pattern] = g
		globCacheMutex.Unlock()
	}

	// Match the path against the compiled glob pattern
	return g.Match(path)
}
