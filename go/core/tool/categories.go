// Package tool provides category resolution functions.
// Categories are semantic groupings (sbom, vuln, etc.) that map to tool IDs.

package tool

import "sync"

// DefaultScannerCategoryMap maps security scanner categories to their default tool IDs.
// These mappings are used by ScannerToolIDForCategory() and align with the
// "security" namespace in tool-config.yml.
//
// To override at runtime, call OverrideScannerCategoryMap before first use.
var DefaultScannerCategoryMap = map[string]string{
	"sbom":       ToolTrivySBOM,
	"vuln":       ToolTrivyVuln,
	"secrets":    ToolTrivySecrets,
	"compliance": ToolTrivyCompliance,
	"iac":        ToolTrivyIaC,
	"sast":       ToolSemgrep,
	"zap":        ToolZap,
}

// DefaultServerTypeMap maps server types to their default tool IDs.
// These mappings are used by ServerToolIDForType() and align with the
// serve-capable tools in tool-config.yml.
//
// To override at runtime, call OverrideServerTypeMap before first use.
var DefaultServerTypeMap = map[string]string{
	ToolStaticSite:      ToolStaticSite,
	ToolMkDocsLive:      ToolMkDocsLive,
	ToolStructurizrLite: ToolStructurizrLite,
}

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

	// Copy from package-level defaults so callers can override
	// DefaultScannerCategoryMap / DefaultServerTypeMap before first access.
	r.scannerMap = make(map[string]string, len(DefaultScannerCategoryMap))
	for k, v := range DefaultScannerCategoryMap {
		r.scannerMap[k] = v
	}

	r.serverMap = make(map[string]string, len(DefaultServerTypeMap))
	for k, v := range DefaultServerTypeMap {
		r.serverMap[k] = v
	}

	r.initialized = true
}

// OverrideScannerCategoryMap replaces the active scanner category mappings.
// This allows config-driven overrides to take effect after initialization.
// Thread-safe: acquires write lock.
func OverrideScannerCategoryMap(m map[string]string) {
	globalCategoryResolver.mu.Lock()
	defer globalCategoryResolver.mu.Unlock()

	globalCategoryResolver.scannerMap = make(map[string]string, len(m))
	for k, v := range m {
		globalCategoryResolver.scannerMap[k] = v
	}
	globalCategoryResolver.initialized = true
}

// OverrideServerTypeMap replaces the active server type mappings.
// This allows config-driven overrides to take effect after initialization.
// Thread-safe: acquires write lock.
func OverrideServerTypeMap(m map[string]string) {
	globalCategoryResolver.mu.Lock()
	defer globalCategoryResolver.mu.Unlock()

	globalCategoryResolver.serverMap = make(map[string]string, len(m))
	for k, v := range m {
		globalCategoryResolver.serverMap[k] = v
	}
	globalCategoryResolver.initialized = true
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

