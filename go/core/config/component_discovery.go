package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ComponentDiscoveryRule defines a single YAML-driven rule for auto-discovering
// a component based on filesystem conventions.
type ComponentDiscoveryRule struct {
	// Component is the name of the component to discover (e.g., "gherkin", "structurizr").
	Component string `yaml:"component"`

	// Path is a direct path pattern with variable expansion (e.g., "{specs_root}/{moniker}").
	Path string `yaml:"path,omitempty"`

	// DeriveFrom lists component names to derive path from (first match wins).
	// Deprecated: use DeriveFromType for type-based lookup that works with bundles.
	DeriveFrom []string `yaml:"derive_from,omitempty"`

	// DeriveFromType lists component TYPES to derive path from (first match wins).
	// Unlike DeriveFrom (which matches by name), this finds the first component
	// whose type matches, making it work correctly for bundle modules where
	// components have names like "core", "contracts" instead of "go".
	DeriveFromType []string `yaml:"derive_from_type,omitempty"`

	// DeriveSubdirs maps source component type/name to subdirectory suffix.
	// e.g., {"go": "specs", "typescript": "features"} → {go_root}/specs
	DeriveSubdirs map[string]string `yaml:"derive_subdirs,omitempty"`

	// DeriveSubdir is a single subdirectory suffix (simpler alternative to DeriveSubdirs).
	// e.g., "specs" → {derived_root}/specs
	DeriveSubdir string `yaml:"derive_subdir,omitempty"`

	// FallbackPath is used when no source component exists for derive_from.
	FallbackPath string `yaml:"fallback_path,omitempty"`

	// Check is the filesystem check to perform: dir_exists, required_file, has_files_in_subdirs, has_matching_files.
	Check string `yaml:"check,omitempty"`

	// RequiredFile is the file that must exist (for check: required_file).
	RequiredFile string `yaml:"required_file,omitempty"`

	// CheckPattern is the filename to check in subdirectories (for check: has_files_in_subdirs).
	CheckPattern string `yaml:"check_pattern,omitempty"`

	// Patterns lists glob patterns for file matching (for check: has_matching_files).
	Patterns []string `yaml:"patterns,omitempty"`

	// SetPatterns copies Patterns into the component's Patterns.Source when true.
	SetPatterns bool `yaml:"set_patterns,omitempty"`

	// OnlyTemplates is a glob filter on module template name. Rule is skipped if template doesn't match.
	OnlyTemplates []string `yaml:"only_templates,omitempty"`
}

// discoverComponentsFromRules applies YAML-driven discovery rules to a module.
// For each rule: check template filter → skip if component already exists with Root →
// compute path (direct or derived) → filesystem check → create ComponentEntry.
func discoverComponentsFromRules(mod *Module, rules []ComponentDiscoveryRule, repoRoot string, vars map[string]string) {
	if len(rules) == 0 {
		return
	}

	if mod.Components == nil {
		mod.Components = make(ModuleComponents)
	}

	for _, rule := range rules {
		// Check template filter
		if !ruleMatchesTemplate(rule.OnlyTemplates, mod.Template) {
			continue
		}

		compName := rule.Component

		// Skip if component already exists with a Root (non-nil with content)
		if existing, ok := mod.Components[compName]; ok && existing != nil && existing.Root != "" {
			continue
		}

		// Compute path and verify via filesystem check
		path := computeRulePath(rule, mod.Components, vars)
		if path == "" || !checkRulePath(rule, repoRoot, path, vars) {
			// Discovery failed; clean up nil marker if present
			if existing, ok := mod.Components[compName]; ok && existing == nil {
				delete(mod.Components, compName)
			}
			continue
		}

		// Create component entry
		entry := &ComponentEntry{Root: path}
		if rule.SetPatterns && len(rule.Patterns) > 0 {
			entry.Patterns = &ComponentPatterns{
				Source: make([]string, len(rule.Patterns)),
			}
			copy(entry.Patterns.Source, rule.Patterns)
		}

		mod.Components[compName] = entry
	}
}

// buildDiscoveryVars creates the variable map for discovery rule expansion.
// Includes repo-level path variables plus module-specific values.
func buildDiscoveryVars(mod *Module, repoCfg *RepositoryConfig) map[string]string {
	vars := map[string]string{
		"moniker": mod.Moniker,
	}

	// Add repo-level path variables
	if repoCfg != nil {
		vars["specs_root"] = repoCfg.Paths.SpecsRoot
		vars["containers_root"] = repoCfg.Paths.ContainersRoot
		vars["design_dir"] = repoCfg.Conventions.DesignDir
		vars["workspace_dsl"] = repoCfg.Conventions.WorkspaceDSL
		vars["specification"] = repoCfg.Conventions.Specification
		vars["godog_test"] = repoCfg.Conventions.GodogTest
	}

	// Add component roots as variables (e.g., go component root → {go} or accessed via derive_from)
	if mod.Components != nil {
		for name, entry := range mod.Components {
			if entry != nil && entry.Root != "" {
				// Make component roots available as variables
				vars[name+"_root"] = entry.Root
			}
		}
	}

	return vars
}

// expandVars replaces {placeholder} patterns in a string using the vars map.
// Delegates to substituteParams which implements the same logic for template expansion.
func expandVars(s string, vars map[string]string) string {
	return substituteParams(s, vars)
}

// computeRulePath computes the filesystem path for a discovery rule.
// Handles direct path, derive_from_type (by component type), and derive_from (by name).
func computeRulePath(rule ComponentDiscoveryRule, components ModuleComponents, vars map[string]string) string {
	// Direct path
	if rule.Path != "" {
		return expandVars(rule.Path, vars)
	}

	// Derive from existing components by TYPE (preferred — works for bundles)
	if len(rule.DeriveFromType) > 0 {
		for _, typeName := range rule.DeriveFromType {
			if name, entry := findFirstComponentByType(components, typeName); entry != nil {
				subdir := rule.DeriveSubdir
				if subdir == "" {
					// Fall back to DeriveSubdirs map (keyed by type name)
					subdir = rule.DeriveSubdirs[typeName]
					if subdir == "" {
						// Also try by component name for backward compat
						subdir = rule.DeriveSubdirs[name]
					}
				}
				if subdir != "" {
					return filepath.ToSlash(filepath.Join(entry.Root, subdir))
				}
				return entry.Root
			}
		}

		// No source found; try fallback
		if rule.FallbackPath != "" {
			return expandVars(rule.FallbackPath, vars)
		}
	}

	// Derive from existing components by NAME (legacy — kept for backward compat)
	if len(rule.DeriveFrom) > 0 {
		for _, sourceName := range rule.DeriveFrom {
			if entry, ok := components[sourceName]; ok && entry != nil && entry.Root != "" {
				if subdir, ok := rule.DeriveSubdirs[sourceName]; ok {
					return filepath.ToSlash(filepath.Join(entry.Root, subdir))
				}
				return entry.Root
			}
		}

		// No source found; try fallback
		if rule.FallbackPath != "" {
			return expandVars(rule.FallbackPath, vars)
		}
	}

	return ""
}

// findFirstComponentByType finds the first component whose type matches.
// Returns the component name and entry, or ("", nil) if not found.
// Components are checked in sorted name order for deterministic results,
// since map iteration order in Go is non-deterministic across processes.
func findFirstComponentByType(components ModuleComponents, typeName string) (string, *ComponentEntry) {
	if components == nil {
		return "", nil
	}
	// Sort component names for deterministic selection.
	// Without sorting, map iteration order varies between process invocations,
	// causing derived components (like assets) to get different roots and patterns,
	// which breaks hash-based cache invalidation across commands.
	names := make([]string, 0, len(components))
	for name := range components {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		entry := components[name]
		if entry == nil || entry.Root == "" {
			continue
		}
		// Component type: explicit Type field, or name is the type
		compType := name
		if entry.Type != "" {
			compType = entry.Type
		}
		if compType == typeName {
			return name, entry
		}
	}
	return "", nil
}

// checkRulePath applies the filesystem check specified by the rule.
func checkRulePath(rule ComponentDiscoveryRule, repoRoot, path string, vars map[string]string) bool {
	switch rule.Check {
	case "dir_exists":
		return pathExists(repoRoot, path)
	case "required_file":
		reqFile := expandVars(rule.RequiredFile, vars)
		return pathHasRequiredFile(repoRoot, path, reqFile)
	case "has_files_in_subdirs":
		pattern := expandVars(rule.CheckPattern, vars)
		return pathHasFilesInSubdirs(repoRoot, path, pattern)
	case "has_matching_files":
		return hasMatchingFiles(filepath.Join(repoRoot, path), rule.Patterns)
	default:
		// No check specified: just verify the path exists
		return pathExists(repoRoot, path)
	}
}

// ruleMatchesTemplate checks if a rule's template filter matches the module's template.
// Returns true if no filter is specified (rule applies to all templates).
func ruleMatchesTemplate(onlyTemplates []string, moduleTemplate string) bool {
	if len(onlyTemplates) == 0 {
		return true
	}

	for _, pattern := range onlyTemplates {
		matched, err := filepath.Match(pattern, moduleTemplate)
		if err == nil && matched {
			return true
		}
	}

	return false
}

// pathExists checks if a path exists in the repo.
func pathExists(repoRoot, path string) bool {
	fullPath := filepath.Join(repoRoot, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// pathHasRequiredFile checks if a path contains the required file.
func pathHasRequiredFile(repoRoot, path, requiredFile string) bool {
	if requiredFile == "" {
		return pathExists(repoRoot, path)
	}
	fullPath := filepath.Join(repoRoot, path, requiredFile)
	_, err := os.Stat(fullPath)
	return err == nil
}

// pathHasFilesInSubdirs checks if a directory has subdirectories containing files matching the pattern.
// This is used for gherkin discovery where features are in specs/{moniker}/{feature-name}/{specification}.
func pathHasFilesInSubdirs(repoRoot, path, filename string) bool {
	fullPath := filepath.Join(repoRoot, path)

	info, err := os.Stat(fullPath)
	if err != nil || !info.IsDir() {
		return false
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			filePath := filepath.Join(fullPath, entry.Name(), filename)
			if _, err := os.Stat(filePath); err == nil {
				return true
			}
		}
	}

	return false
}

// hasMatchingFiles checks if a directory (or its subdirectories) contains files
// matching any of the given glob patterns.
func hasMatchingFiles(dir string, patterns []string) bool {
	var found bool

	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || found {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := d.Name()
		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, name)
			if err == nil && matched {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})

	return found
}

// DiscoverContainerComponents scans {containersRoot}/*/Dockerfile and creates
// dockerfile components within the given module. Components get docker_build config
// matching container template structure. Push is derived from the local_only parameter.
//
// Parameters:
//   - mod: the module to add discovered components to
//   - repoRoot: repository root directory
//   - containersRoot: containers root directory name (e.g., "containers")
//   - owner: repository owner for tag generation
func DiscoverContainerComponents(mod *Module, repoRoot, containersRoot, owner string) {
	if containersRoot == "" {
		containersRoot = "containers"
	}
	containersDir := filepath.Join(repoRoot, containersRoot)

	// Check if containers directory exists
	if _, err := os.Stat(containersDir); os.IsNotExist(err) {
		return
	}

	// Parse local_only from discover_components config
	localOnly := make(map[string]bool)
	var nameSuffix string
	if mod.DiscoverComponents != nil {
		for _, name := range mod.DiscoverComponents.LocalOnly {
			localOnly[strings.TrimSpace(name)] = true
		}
		nameSuffix = mod.DiscoverComponents.NameSuffix
	}

	// Find all Dockerfiles
	pattern := filepath.Join(containersDir, "*", "Dockerfile")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	// Sort for deterministic ordering
	sort.Strings(matches)

	if mod.Components == nil {
		mod.Components = make(ModuleComponents)
	}

	for _, dockerfile := range matches {
		// Extract container name from path (parent directory name)
		dir := filepath.Dir(dockerfile)
		name := filepath.Base(dir)

		// Skip containers that don't match the required name suffix
		if nameSuffix != "" && !strings.HasSuffix(name, nameSuffix) {
			continue
		}

		// Skip if component already exists
		if _, exists := mod.Components[name]; exists {
			continue
		}

		// Determine if this container should be pushed
		isLocalOnly := localOnly[name]
		shouldPush := !isLocalOnly

		// Build docker_build config
		root := filepath.ToSlash(filepath.Join(containersRoot, name))
		dockerBuild := &DockerBuildConfig{
			Container:  name,
			Context:    root,
			Dockerfile: filepath.ToSlash(filepath.Join(root, "Dockerfile")),
			Platforms:  []string{"linux/amd64"},
			Push:       BoolPtr(shouldPush),
		}

		// Only add tags, registry, and cache for pushed containers
		if shouldPush && owner != "" {
			dockerBuild.Tags = []string{
				"ghcr.io/" + owner + "/" + name + ":latest",
				"ghcr.io/" + owner + "/" + name + ":sha-{short_sha}",
			}
			dockerBuild.Registry = "ghcr.io"
			dockerBuild.Cache = &DockerCacheConfig{
				Type: "gha",
			}
		}

		// Create component entry
		entry := &ComponentEntry{
			Type:        "dockerfile",
			Root:        root,
			DockerBuild: dockerBuild,
			Patterns: &ComponentPatterns{
				Source: []string{"**/*"},
			},
		}

		mod.Components[name] = entry
	}
}
