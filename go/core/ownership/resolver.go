package ownership

import (
	"path/filepath"
	"sort"
	"strings"
)

// Resolver resolves file ownership using the simplified model.
// Ownership is determined by directory roots: the most-specific-root wins.
// When multiple components share the same root, extension-based tie-breaking
// selects the component whose type matches the file extension.
type Resolver struct {
	// entries sorted by root specificity (most specific first).
	// File entries come before root entries at the same depth.
	entries []ownerEntry
}

type ownerEntry struct {
	module        string
	component     string
	root          string   // normalized, always forward slashes
	depth         int      // number of path segments (more = more specific)
	isFile        bool     // true if File ownership (not Root)
	componentType string   // for extension tie-breaking
	extensions    []string // file extensions from component type
}

// NewResolver creates an ownership resolver from module definitions.
func NewResolver(modules []ModuleDefinition) *Resolver {
	var entries []ownerEntry
	for _, m := range modules {
		for compName, comp := range m.Components {
			if comp == nil {
				continue
			}
			if comp.Root != "" {
				root := toSlash(comp.Root)
				entries = append(entries, ownerEntry{
					module:        m.Moniker,
					component:     compName,
					root:          root,
					depth:         countSegments(root),
					componentType: comp.ComponentType,
					extensions:    comp.Extensions,
				})
			}
			if comp.File != "" {
				entries = append(entries, ownerEntry{
					module:        m.Moniker,
					component:     compName,
					root:          toSlash(comp.File),
					isFile:        true,
					depth:         1000, // Files always win over directories
					componentType: comp.ComponentType,
					extensions:    comp.Extensions,
				})
			}
		}
	}

	// Sort by depth descending (most specific first).
	// File entries sort before root entries at the same effective depth.
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].depth != entries[j].depth {
			return entries[i].depth > entries[j].depth
		}
		if entries[i].isFile != entries[j].isFile {
			return entries[i].isFile
		}
		return entries[i].root < entries[j].root
	})

	return &Resolver{entries: entries}
}

// Resolve returns the owner of a file path.
// The filePath should be relative to the repo root using forward slashes,
// but backslashes are also accepted and normalized.
// Returns nil if no owner found.
func (r *Resolver) Resolve(filePath string) *Owner {
	path := toSlash(filePath)

	// Check for exact file matches (highest priority)
	for i := range r.entries {
		e := &r.entries[i]
		if e.isFile && path == e.root {
			return &Owner{Module: e.module, Component: e.component, Root: e.root}
		}
	}

	candidates := r.findCandidates(path)
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) == 1 {
		c := &candidates[0]
		return &Owner{Module: c.module, Component: c.component, Root: c.root}
	}

	return r.resolveByExtension(path, candidates)
}

// ResolveModule returns all files from allFiles that are owned by the named module.
func (r *Resolver) ResolveModule(moduleName string, allFiles []string) []string {
	var result []string
	for _, f := range allFiles {
		if owner := r.Resolve(f); owner != nil && owner.Module == moduleName {
			result = append(result, f)
		}
	}
	return result
}

// ResolveComponent returns all files from allFiles that are owned by a specific
// module:component pair.
func (r *Resolver) ResolveComponent(moduleName, componentName string, allFiles []string) []string {
	var result []string
	for _, f := range allFiles {
		if owner := r.Resolve(f); owner != nil && owner.Module == moduleName && owner.Component == componentName {
			result = append(result, f)
		}
	}
	return result
}

// Validate checks that every file in allFiles is owned by exactly one module.
// Returns validation errors for unowned or multi-owned files.
func (r *Resolver) Validate(allFiles []string) []ValidationError {
	var errs []ValidationError

	for _, f := range allFiles {
		path := toSlash(f)

		// Check file entries first
		var fileOwnerCount int
		for i := range r.entries {
			e := &r.entries[i]
			if e.isFile && path == e.root {
				fileOwnerCount++
			}
		}
		if fileOwnerCount > 1 {
			errs = append(errs, ValidationError{FilePath: f, Issue: "multi-owned"})
			continue
		}
		if fileOwnerCount == 1 {
			continue
		}

		// Check root candidates for cross-module overlap
		candidates := r.findCandidates(path)
		if len(candidates) == 0 {
			errs = append(errs, ValidationError{FilePath: f, Issue: "unowned"})
			continue
		}

		// Count distinct modules among candidates
		modules := make(map[string]bool)
		for _, c := range candidates {
			modules[c.module] = true
		}
		if len(modules) > 1 {
			owners := make([]Owner, len(candidates))
			for i, c := range candidates {
				owners[i] = Owner{Module: c.module, Component: c.component, Root: c.root}
			}
			errs = append(errs, ValidationError{FilePath: f, Issue: "multi-owned", Owners: owners})
		}
	}

	return errs
}

// findCandidates returns all root entries at the most-specific depth that match path.
// Entries are sorted by depth (most specific first), so we collect all entries
// at the first matching depth and stop.
func (r *Resolver) findCandidates(path string) []ownerEntry {
	var bestDepth = -1
	var candidates []ownerEntry

	for i := range r.entries {
		e := &r.entries[i]
		if e.isFile {
			continue
		}
		if !rootContains(e.root, path) {
			continue
		}
		if bestDepth == -1 {
			bestDepth = e.depth
		}
		if e.depth < bestDepth {
			break
		}
		candidates = append(candidates, *e)
	}

	return candidates
}

// resolveByExtension picks the best owner when multiple components share the same root.
// Components with extensions matching the file's extension win.
// If no extension match, the component with no extensions (catch-all) wins.
func (r *Resolver) resolveByExtension(path string, candidates []ownerEntry) *Owner {
	ext := filepath.Ext(path)

	var catchAll *ownerEntry
	for i := range candidates {
		c := &candidates[i]
		for _, e := range c.extensions {
			if e == ext {
				return &Owner{Module: c.module, Component: c.component, Root: c.root}
			}
		}
		if len(c.extensions) == 0 && catchAll == nil {
			catchAll = c
		}
	}

	if catchAll != nil {
		return &Owner{Module: catchAll.module, Component: catchAll.component, Root: catchAll.root}
	}

	// Fallback: return first candidate
	c := &candidates[0]
	return &Owner{Module: c.module, Component: c.component, Root: c.root}
}

// rootContains returns true if the file path is under the given root.
func rootContains(root, path string) bool {
	if root == "/" {
		return true
	}
	return strings.HasPrefix(path, root+"/")
}

// countSegments returns the number of path segments.
func countSegments(path string) int {
	if path == "/" {
		return 0
	}
	return strings.Count(path, "/") + 1
}

// toSlash normalizes path separators to forward slashes.
func toSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
