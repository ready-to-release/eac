package modules

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/hash"
)

// ModuleContract represents a deployable module contract.
type ModuleContract struct {
	domain.BaseContract `yaml:",inline"`

	// Cached computed values
	workspaceRoot string
}

// NewModuleContract creates a new module contract with workspace context.
func NewModuleContract(base domain.BaseContract, workspaceRoot string) *ModuleContract {
	return &ModuleContract{
		BaseContract:  base,
		workspaceRoot: workspaceRoot,
	}
}

// GetComponentRoot returns the root for a specific component type.
func (m *ModuleContract) GetComponentRoot(compType string) string {
	return m.Components.GetComponentRoot(compType)
}

// GetComponentRoots returns all component roots as a map.
func (m *ModuleContract) GetComponentRoots() map[string]string {
	return m.Components.GetAllRoots()
}

// GetGlobPatterns returns GitHub Actions compatible glob patterns for this module.
// Collects patterns from all components. Results are sorted for determinism
// since Components is a map with non-deterministic iteration order. This ensures
// the hash cache (which uses slices.Equal on patterns) can match across runs.
func (m *ModuleContract) GetGlobPatterns() []string {
	var patterns []string

	// Collect component names and sort for deterministic iteration
	compNames := make([]string, 0, len(m.Components))
	for name := range m.Components {
		compNames = append(compNames, name)
	}
	sort.Strings(compNames)

	for _, name := range compNames {
		comp := m.Components[name]
		if comp == nil {
			continue
		}
		root := normalizePathSeparators(comp.Root)

		// Collect patterns from this component
		if comp.Patterns != nil {
			for _, pattern := range comp.Patterns.Source {
				patterns = append(patterns, joinPattern(root, pattern))
			}
			for _, pattern := range comp.Patterns.Tests {
				patterns = append(patterns, joinPattern(root, pattern))
			}
			for _, pattern := range comp.Patterns.Config {
				patterns = append(patterns, joinPattern(root, pattern))
			}
			for _, pattern := range comp.Patterns.Data {
				patterns = append(patterns, joinPattern(root, pattern))
			}
		}
	}

	// Add changelog
	if changelog := m.GetChangelogPath(); changelog != "" {
		patterns = append(patterns, normalizePathSeparators(changelog))
	}

	return patterns
}

// joinPattern joins a root and pattern, handling empty roots.
func joinPattern(root, pattern string) string {
	pattern = normalizePathSeparators(pattern)
	if root == "" || root == "/" {
		return pattern
	}
	return root + "/" + pattern
}

// GetAbsolutePaths returns absolute file system paths for this module's sources.
func (m *ModuleContract) GetAbsolutePaths() []string {
	if m.workspaceRoot == "" {
		return []string{}
	}

	var paths []string
	for _, comp := range m.Components {
		if comp == nil || comp.Patterns == nil {
			continue
		}
		root := comp.Root
		for _, pattern := range comp.Patterns.Source {
			if root != "" {
				paths = append(paths, filepath.Join(m.workspaceRoot, root, pattern))
			} else {
				paths = append(paths, filepath.Join(m.workspaceRoot, pattern))
			}
		}
	}

	return paths
}

// MatchesFile returns true if the given file path matches any component's patterns.
func (m *ModuleContract) MatchesFile(filePath string) bool {
	path := normalizePathSeparators(filePath)

	for _, comp := range m.Components {
		if comp == nil {
			continue
		}
		root := normalizePathSeparators(comp.Root)

		// Check if file is under this component's root
		isUnderRoot := root == "" || root == "/" ||
			strings.HasPrefix(path, root+"/") || path == root

		if !isUnderRoot {
			continue
		}

		// Collect all declared patterns across source, tests, config, and data.
		var allPatterns []string
		if comp.Patterns != nil {
			allPatterns = append(allPatterns, comp.Patterns.Source...)
			allPatterns = append(allPatterns, comp.Patterns.Tests...)
			allPatterns = append(allPatterns, comp.Patterns.Config...)
			allPatterns = append(allPatterns, comp.Patterns.Data...)
		}

		if len(allPatterns) > 0 {
			// Component has explicit patterns - match against them.
			for _, pattern := range allPatterns {
				fullPattern := joinPattern(root, pattern)
				if matchWithFallback(path, fullPattern) {
					return true
				}
			}
		} else if root != "" && root != "/" {
			// No patterns defined - component owns all files under its root.
			// This covers any component type that declares a root without
			// restricting by patterns (e.g., static content, books).
			return true
		}
	}

	return false
}

// GetDependencies returns the list of module dependencies.
func (m *ModuleContract) GetDependencies() []string {
	return m.DependsOn
}

// IsDefinitionsFile returns true if this contract represents a definitions file.
func (m *ModuleContract) IsDefinitionsFile() bool {
	return m.Moniker == "definitions"
}

// GetSpecsRoot returns the specs root directory for this module.
// Uses the specs or gherkin component root, or defaults to specs/{moniker}.
func (m *ModuleContract) GetSpecsRoot() string {
	// Check specs component first
	if comp, ok := m.Components["specs"]; ok && comp != nil && comp.Root != "" {
		return comp.Root
	}
	// Check gherkin component
	if comp, ok := m.Components["gherkin"]; ok && comp != nil && comp.Root != "" {
		return comp.Root
	}
	// Default: specs/<module-moniker>
	return filepath.Join("specs", m.Moniker)
}

// GetChangelogPath returns the changelog file path.
// Delegates to BaseContract.GetChangelog() which uses versioning.changelog if set,
// otherwise defaults to release/<moniker>/CHANGELOG.md.
func (m *ModuleContract) GetChangelogPath() string {
	return m.GetChangelog()
}

// GetReleaseNotesPath returns the path to RELEASE-NOTES.md for this module.
// Defaults to release/{moniker}/RELEASE-NOTES.md.
func (m *ModuleContract) GetReleaseNotesPath() string {
	return filepath.Join("release", m.Moniker, "RELEASE-NOTES.md")
}

// GetTestImplementationPath returns the BDD test runner directory path.
// Checks for known BDD runner components (godog, cucumberjs).
// Returns empty string if no BDD runner component is defined.
func (m *ModuleContract) GetTestImplementationPath() string {
	for _, name := range []string{"godog", "cucumberjs"} {
		if comp, ok := m.Components[name]; ok && comp != nil && comp.Root != "" {
			return comp.Root
		}
	}
	return ""
}

// GetDesignPath returns the design workspace directory path.
// Uses the design component root, or defaults to specs/{moniker}/.design.
func (m *ModuleContract) GetDesignPath() string {
	if comp, ok := m.Components["design"]; ok && comp != nil && comp.Root != "" {
		return comp.Root
	}
	// Default: specs/<module-moniker>/.design
	return filepath.Join("specs", m.Moniker, ".design")
}

// GetCIWorkflowPath returns the CI workflow file path.
// Derives from workflows component patterns or convention.
func (m *ModuleContract) GetCIWorkflowPath() string {
	if comp, ok := m.Components["workflows"]; ok && comp != nil {
		root := comp.Root
		if root == "" {
			root = ".github/workflows"
		}
		// Check if patterns specify CI workflow
		if comp.Patterns != nil {
			for _, p := range comp.Patterns.Source {
				if strings.HasPrefix(p, "ci-") {
					return filepath.Join(root, p)
				}
			}
		}
		// Convention
		return filepath.Join(root, "ci-"+m.Moniker+".yaml")
	}
	// Default convention
	return filepath.Join(".github/workflows", "ci-"+m.Moniker+".yaml")
}

// GetReleaseWorkflowPath returns the release workflow file path.
// Derives from workflows component patterns or convention.
func (m *ModuleContract) GetReleaseWorkflowPath() string {
	if comp, ok := m.Components["workflows"]; ok && comp != nil {
		root := comp.Root
		if root == "" {
			root = ".github/workflows"
		}
		// Check if patterns specify release workflow
		if comp.Patterns != nil {
			for _, p := range comp.Patterns.Source {
				if strings.HasPrefix(p, "release-") {
					return filepath.Join(root, p)
				}
			}
		}
		// Convention
		return filepath.Join(root, "release-"+m.Moniker+".yaml")
	}
	// Default convention
	return filepath.Join(".github/workflows", "release-"+m.Moniker+".yaml")
}

// GetBooks returns the list of book names for this module.
// Books are defined in the book component configuration.
func (m *ModuleContract) GetBooks() []string {
	return m.BaseContract.GetBooks()
}

// HasBooks returns true if the module has books defined.
func (m *ModuleContract) HasBooks() bool {
	return m.Components.HasComponentType("book")
}

// normalizePathSeparators converts Windows backslashes to forward slashes.
func normalizePathSeparators(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// matchWithFallback handles glob matching with fallback for ** patterns.
// The doublestar library handles most cases, but root-level files sometimes
// need explicit fallback matching when ** is used at different positions.
func matchWithFallback(path, pattern string) bool {
	if matchGlobPattern(path, pattern) {
		return true
	}

	// Fallback: try collapsing ** to match root-level files directly.
	switch {
	case strings.HasPrefix(pattern, "**/"):
		// "**/foo.go" should also match "foo.go" (root-level)
		return matchGlobPattern(path, strings.TrimPrefix(pattern, "**/"))
	case strings.HasPrefix(pattern, "**"):
		// "**.go" should also match ".go" suffix at root level
		return matchGlobPattern(path, strings.TrimPrefix(pattern, "**"))
	case strings.Contains(pattern, "/**/"):
		// "src/**/foo.go" should also match "src/foo.go" (no intermediate dirs)
		parts := strings.SplitN(pattern, "/**/", 2)
		if len(parts) == 2 {
			return matchGlobPattern(path, parts[0]+"/"+parts[1])
		}
	}

	return false
}

// matchGlobPattern performs glob pattern matching using doublestar.
// Supports shorthand like "**.go" which is expanded to "**/*.go".
func matchGlobPattern(path, pattern string) bool {
	path = normalizePathSeparators(path)
	pattern = normalizePathSeparators(pattern)

	// Convert shorthand "**.ext" to "**/*.ext" for doublestar compatibility
	if strings.HasPrefix(pattern, "**") && !strings.HasPrefix(pattern, "**/") && len(pattern) > 2 {
		pattern = "**/*" + pattern[2:]
	}

	matched, err := doublestar.Match(pattern, path)
	if err != nil {
		return false
	}
	return matched
}

// HasComponent returns true if the given component type is enabled for this module.
func (m *ModuleContract) HasComponent(compType string) bool {
	if m.Components == nil {
		return false
	}
	return m.Components.HasComponent(compType)
}

// GetEnabledComponents returns the list of enabled component types for this module.
func (m *ModuleContract) GetEnabledComponents() []string {
	if m.Components == nil {
		return nil
	}
	return m.Components.GetEnabled()
}

// GetComponentTypesDisplay returns a comma-separated list of enabled component types.
// Uses ComponentOrder for YAML declaration order when available, otherwise sorted alphabetically.
func (m *ModuleContract) GetComponentTypesDisplay() string {
	comps := m.GetOrderedComponentNames()
	if len(comps) == 0 {
		return "(no components)"
	}
	return strings.Join(comps, ", ")
}

// GetOrderedComponentNames returns component names in YAML declaration order.
// Falls back to sorted alphabetical order if ComponentOrder is not set.
func (m *ModuleContract) GetOrderedComponentNames() []string {
	if m.Components == nil {
		return nil
	}

	// Use declared order if available
	if len(m.ComponentOrder) > 0 {
		// Return ordered names, including any that were added after initial parse
		// (e.g., by discovery rules) appended at the end
		seen := make(map[string]bool, len(m.ComponentOrder))
		var result []string
		for _, name := range m.ComponentOrder {
			if m.Components.HasComponent(name) {
				result = append(result, name)
				seen[name] = true
			}
		}
		// Append any components not in the original order (discovered later)
		extras := m.Components.GetEnabled()
		sort.Strings(extras)
		for _, name := range extras {
			if !seen[name] {
				result = append(result, name)
			}
		}
		return result
	}

	// Fallback: sorted alphabetically
	comps := m.Components.GetEnabled()
	sort.Strings(comps)
	return comps
}

// GetContentHash returns a SHA256 hash of the module's owned files.
// Used for content-addressed caching (e.g., Docker image tags).
// Uses the core hash utility via GetGlobPatterns() for consistency.
// Returns short hash (8 chars) and any error encountered.
func (m *ModuleContract) GetContentHash() (string, error) {
	if m.workspaceRoot == "" {
		return "", nil
	}

	// Use core hash utility - module implements GlobPatternGetter via GetGlobPatterns()
	fullHash, err := hash.ComputeFromPatterns(m.workspaceRoot, m)
	if err != nil {
		return "", err
	}

	if len(fullHash) >= 8 {
		return fullHash[:8], nil
	}
	return fullHash, nil
}

