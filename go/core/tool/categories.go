// Package tool provides category resolution functions.
// Categories are semantic groupings (sbom, vuln, etc.) that map to tool IDs.

package tool

import "sync"

// newDefaultScannerCategoryMap returns a fresh scanner category map.
// Returned as a new map each time to prevent callers from mutating shared state.
func newDefaultScannerCategoryMap() map[string]string {
	return map[string]string{
		"sbom":       ToolTrivySBOM,
		"vuln":       ToolTrivyVuln,
		"secrets":    ToolTrivySecrets,
		"compliance": ToolTrivyCompliance,
		"iac":        ToolTrivyIaC,
		"sast":       ToolSemgrep,
		"zap":        ToolZap,
	}
}

// newDefaultServerTypeMap returns a fresh server type map.
// Returned as a new map each time to prevent callers from mutating shared state.
func newDefaultServerTypeMap() map[string]string {
	return map[string]string{
		ToolStaticSite:      ToolStaticSite,
		ToolMkDocsLive:      ToolMkDocsLive,
		ToolStructurizrLite: ToolStructurizrLite,
	}
}

// DefaultScannerCategories returns a fresh copy of the scanner category map.
func DefaultScannerCategories() map[string]string {
	return newDefaultScannerCategoryMap()
}

// DefaultServerTypes returns a fresh copy of the server type map.
func DefaultServerTypes() map[string]string {
	return newDefaultServerTypeMap()
}

// CategoryResolver maps scanner categories to default tool IDs.
// The mapping is derived from tool-config.yml conventions.
type CategoryResolver struct {
	mu          sync.RWMutex
	scannerMap  map[string]string // category -> tool ID
	serverMap   map[string]string // server type -> tool ID
	initialized bool
}

// globalCategoryResolverOnce ensures the singleton is created exactly once.
var globalCategoryResolverOnce sync.Once

// globalCategoryResolverInstance is the lazily-initialized singleton.
// Access only via getGlobalCategoryResolver().
var globalCategoryResolverInstance *CategoryResolver

// getGlobalCategoryResolver returns the singleton CategoryResolver, creating it on first call.
func getGlobalCategoryResolver() *CategoryResolver {
	globalCategoryResolverOnce.Do(func() {
		globalCategoryResolverInstance = &CategoryResolver{}
	})
	return globalCategoryResolverInstance
}

// initDefaultMappings sets up default category-to-tool mappings.
// Called lazily on first access.
func (r *CategoryResolver) initDefaultMappings() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.initialized {
		return
	}

	r.scannerMap = newDefaultScannerCategoryMap()
	r.serverMap = newDefaultServerTypeMap()
	r.initialized = true
}

// ScannerToolIDForCategory returns the default tool ID for a scanner category.
func (r *CategoryResolver) ScannerToolIDForCategory(category string) string {
	r.initDefaultMappings()
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.scannerMap[category]
}

// ServerToolIDForType returns the tool ID for a server type.
func (r *CategoryResolver) ServerToolIDForType(serverType string) string {
	r.initDefaultMappings()
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.serverMap[serverType]
}

// AllScannerCategories returns all known scanner categories.
func (r *CategoryResolver) AllScannerCategories() []string {
	r.initDefaultMappings()
	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make([]string, 0, len(r.scannerMap))
	for cat := range r.scannerMap {
		categories = append(categories, cat)
	}
	return categories
}

// AllServerTypes returns all known server types.
func (r *CategoryResolver) AllServerTypes() []string {
	r.initDefaultMappings()
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.serverMap))
	for t := range r.serverMap {
		types = append(types, t)
	}
	return types
}

// OverrideScannerCategoryMap replaces the active scanner category mappings.
// Thread-safe: acquires write lock.
func (r *CategoryResolver) OverrideScannerCategoryMap(m map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.scannerMap = make(map[string]string, len(m))
	for k, v := range m {
		r.scannerMap[k] = v
	}
	r.initialized = true
}

// OverrideServerTypeMap replaces the active server type mappings.
// Thread-safe: acquires write lock.
func (r *CategoryResolver) OverrideServerTypeMap(m map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.serverMap = make(map[string]string, len(m))
	for k, v := range m {
		r.serverMap[k] = v
	}
	r.initialized = true
}

// Package-level convenience functions using the global CategoryResolver singleton.

// ScannerToolIDForCategory returns the default tool ID for a scanner category.
// Categories: sbom, vuln, secrets, compliance, iac, sast, zap
// Returns empty string if category is unknown.
func ScannerToolIDForCategory(category string) string {
	return getGlobalCategoryResolver().ScannerToolIDForCategory(category)
}

// ServerToolIDForType returns the tool ID for a server type.
// Server types: nginx-oci, mkdocs-live, structurizr-lite
// Returns empty string if type is unknown.
func ServerToolIDForType(serverType string) string {
	return getGlobalCategoryResolver().ServerToolIDForType(serverType)
}

// AllScannerCategories returns all known scanner categories.
func AllScannerCategories() []string {
	return getGlobalCategoryResolver().AllScannerCategories()
}

// AllServerTypes returns all known server types.
func AllServerTypes() []string {
	return getGlobalCategoryResolver().AllServerTypes()
}

// OverrideScannerCategoryMap replaces the active scanner category mappings on the global resolver.
func OverrideScannerCategoryMap(m map[string]string) {
	getGlobalCategoryResolver().OverrideScannerCategoryMap(m)
}

// OverrideServerTypeMap replaces the active server type mappings on the global resolver.
func OverrideServerTypeMap(m map[string]string) {
	getGlobalCategoryResolver().OverrideServerTypeMap(m)
}

// IsScannerCategory returns true if the string is a known scanner category.
func IsScannerCategory(s string) bool {
	return ScannerToolIDForCategory(s) != ""
}

// IsServerType returns true if the string is a known server type.
func IsServerType(s string) bool {
	return ServerToolIDForType(s) != ""
}
