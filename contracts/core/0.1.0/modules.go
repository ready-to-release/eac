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

// ModuleContractPort provides access to a single module's contract.
type ModuleContractPort interface {
	// Identity
	GetMoniker() string
	GetName() string
	GetDescription() string

	// Components
	HasComponent(componentType string) bool
	GetComponentRoot(componentType string) string
	GetComponentRoots() map[string]string
	GetComponentTypesDisplay() string
	GetComponentAmp(componentName, operation string) float64

	// Dependencies
	GetDependsOn() []string

	// Versioning
	GetVersioningScheme() string
	GetReleaseType() string
	GetChangelog() string
	HasVersioning() bool

	// Metadata
	GetMetadata() map[string]interface{}

	// Content
	// GetContentHash returns a SHA256 hash of the module's owned files.
	// Used for content-addressed caching (e.g., Docker image tags).
	// Returns short hash (8 chars) and any error encountered.
	GetContentHash() (string, error)
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
