package core

// ModuleRegistryPort provides access to module domain.
type ModuleRegistryPort interface {
	// Get retrieves a module contract by moniker.
	Get(moniker string) (ModuleContractPort, bool)

	// Has checks if a module exists in the registry.
	Has(moniker string) bool

	// All returns all module domain.
	All() []ModuleContractPort

	// AllMonikers returns all module monikers sorted alphabetically.
	AllMonikers() []string

	// Count returns the number of modules.
	Count() int

	// FilterByComponent returns modules with the specified component type.
	FilterByComponent(componentType string) []ModuleContractPort

	// FindModulesForFile returns modules that own the given file path.
	FindModulesForFile(filePath string) []ModuleContractPort

	// WorkspaceRoot returns the workspace root path.
	WorkspaceRoot() string
}

// ModuleIdentity provides module identification.
type ModuleIdentity interface {
	GetMoniker() string
	GetName() string
	GetDescription() string
	GetModuleGroup() string
}

// ModuleComponentProvider provides component layout information.
type ModuleComponentProvider interface {
	HasComponent(componentType string) bool
	GetComponentRoot(componentType string) string
	GetComponentRoots() map[string]string
	GetComponentTypesDisplay() string
	GetComponentAmp(componentName, operation string) float64
	GetComponentGroup(componentName string) string
}

// ModuleReleaseInfo provides versioning and changelog information.
type ModuleReleaseInfo interface {
	GetVersioningScheme() string
	GetReleaseType() string
	GetChangelog() string
	HasVersioning() bool
}

// ModuleContent provides dependency and metadata information.
type ModuleContent interface {
	GetDependsOn() []string
	GetMetadata() map[string]interface{}
	// GetContentHash returns a SHA256 hash of the module's owned files.
	// Used for content-addressed caching (e.g., Docker image tags).
	// Returns short hash (8 chars) and any error encountered.
	GetContentHash() (string, error)
}

// ModuleContractPort provides access to a single module's contract.
// It composes the focused sub-interfaces for full module access.
type ModuleContractPort interface {
	ModuleIdentity
	ModuleComponentProvider
	ModuleReleaseInfo
	ModuleContent
}

// ComponentTypePort provides component type configuration.
type ComponentTypePort interface {
	GetName() string
	GetPatterns() ComponentPatternsPort
	HasBuild() bool
}

// ComponentPatternsPort provides file patterns for a component type.
type ComponentPatternsPort interface {
	GetSource() []string
	GetTests() []string
	GetConfig() []string
}
