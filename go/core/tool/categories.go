// Package tool provides category resolution functions.
// Categories are semantic groupings (sbom, vuln, etc.) that map to tool IDs.

package tool

import "sync"

// CategoryResolver maps scanner categories to default tool IDs.
// The mapping is derived from tool-config.yml conventions.
type CategoryResolver struct {
	mu          sync.RWMutex
	scannerMap  map[string]string // category -> tool ID
	serverMap   map[string]string // server type -> tool ID
	initialized bool
}

// globalCategoryResolver is the singleton category resolver.
var globalCategoryResolver = &CategoryResolver{}

// initDefaultMappings sets up default category-to-tool mappings.
// Called lazily on first access.
func (r *CategoryResolver) initDefaultMappings() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.initialized {
		return
	}

	// Scanner category -> tool ID mapping
	// Categories are semantic names, tools are concrete implementations
	r.scannerMap = map[string]string{
		"sbom":       ToolTrivySBOM,
		"vuln":       ToolTrivyVuln,
		"secrets":    ToolTrivySecrets,
		"compliance": ToolTrivyCompliance,
		"iac":        ToolTrivyIaC,
		"sast":       ToolSemgrep,
		"zap":        ToolZap,
	}

	// Server type -> tool ID mapping
	// Server types are the canonical identifiers
	r.serverMap = map[string]string{
		ToolStaticSite:      ToolStaticSite,
		ToolMkDocsLive:      ToolMkDocsLive,
		ToolStructurizrLite: ToolStructurizrLite,
	}

	r.initialized = true
}

// ScannerToolIDForCategory returns the default tool ID for a scanner category.
// Categories: sbom, vuln, secrets, compliance, iac, sast, zap
// Returns empty string if category is unknown.
func ScannerToolIDForCategory(category string) string {
	globalCategoryResolver.initDefaultMappings()

	globalCategoryResolver.mu.RLock()
	defer globalCategoryResolver.mu.RUnlock()
	return globalCategoryResolver.scannerMap[category]
}

// ServerToolIDForType returns the tool ID for a server type.
// Server types: nginx-oci, mkdocs-live, structurizr-lite
// Returns empty string if type is unknown.
func ServerToolIDForType(serverType string) string {
	globalCategoryResolver.initDefaultMappings()

	globalCategoryResolver.mu.RLock()
	defer globalCategoryResolver.mu.RUnlock()
	return globalCategoryResolver.serverMap[serverType]
}

// AllScannerCategories returns all known scanner categories.
func AllScannerCategories() []string {
	globalCategoryResolver.initDefaultMappings()

	globalCategoryResolver.mu.RLock()
	defer globalCategoryResolver.mu.RUnlock()

	categories := make([]string, 0, len(globalCategoryResolver.scannerMap))
	for cat := range globalCategoryResolver.scannerMap {
		categories = append(categories, cat)
	}
	return categories
}

// AllServerTypes returns all known server types.
func AllServerTypes() []string {
	globalCategoryResolver.initDefaultMappings()

	globalCategoryResolver.mu.RLock()
	defer globalCategoryResolver.mu.RUnlock()

	types := make([]string, 0, len(globalCategoryResolver.serverMap))
	for t := range globalCategoryResolver.serverMap {
		types = append(types, t)
	}
	return types
}

// IsScannerCategory returns true if the string is a known scanner category.
func IsScannerCategory(s string) bool {
	return ScannerToolIDForCategory(s) != ""
}

// IsServerType returns true if the string is a known server type.
func IsServerType(s string) bool {
	return ServerToolIDForType(s) != ""
}

// CategoryToToolID converts a scanner category to its tool ID.
// This is an alias for ScannerToolIDForCategory for backward compatibility.
// Deprecated: Use ScannerToolIDForCategory instead.
func CategoryToToolID(category string) (string, bool) {
	toolID := ScannerToolIDForCategory(category)
	return toolID, toolID != ""
}
