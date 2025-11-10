package gomod

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/src/internal/contracts/modules"
)

// Mapper handles mapping between go.mod module paths and module contract monikers
type Mapper struct {
	registry       *modules.Registry
	baseModulePath string // e.g., "github.com/ready-to-release/eac"
	pathToMoniker  map[string]string
	monikerToPath  map[string]string
}

// NewMapper creates a new mapper with the given registry and base module path
func NewMapper(registry *modules.Registry, baseModulePath string) *Mapper {
	mapper := &Mapper{
		registry:       registry,
		baseModulePath: baseModulePath,
		pathToMoniker:  make(map[string]string),
		monikerToPath:  make(map[string]string),
	}

	mapper.buildMappings()
	return mapper
}

// buildMappings builds bidirectional lookup maps between module paths and monikers
func (m *Mapper) buildMappings() {
	for _, module := range m.registry.All() {
		// Build expected module path from source root
		// Example: source.root = "src/cli" -> module path = "github.com/ready-to-release/eac/src/cli"
		modulePath := m.baseModulePath + "/" + module.Source.Root

		// Store mappings
		m.pathToMoniker[modulePath] = module.Moniker
		m.monikerToPath[module.Moniker] = modulePath
	}
}

// GetMonikerFromPath converts a go.mod module path to a module moniker
// Example: "github.com/ready-to-release/eac/src/cli" -> "src-cli"
func (m *Mapper) GetMonikerFromPath(modulePath string) (string, error) {
	if moniker, ok := m.pathToMoniker[modulePath]; ok {
		return moniker, nil
	}

	// Try to find by extracting relative path
	relPath := strings.TrimPrefix(modulePath, m.baseModulePath+"/")

	// Look for module with matching source root
	for _, module := range m.registry.All() {
		if module.Source.Root == relPath {
			return module.Moniker, nil
		}
	}

	return "", fmt.Errorf("no module contract found for path: %s", modulePath)
}

// GetPathFromMoniker converts a module moniker to a go.mod module path
// Example: "src-cli" -> "github.com/ready-to-release/eac/src/cli"
func (m *Mapper) GetPathFromMoniker(moniker string) (string, error) {
	if path, ok := m.monikerToPath[moniker]; ok {
		return path, nil
	}

	// Try to find module and construct path
	module, exists := m.registry.Get(moniker)
	if !exists {
		return "", fmt.Errorf("module not found: %s", moniker)
	}

	return m.baseModulePath + "/" + module.Source.Root, nil
}

// GetMonikerFromModuleDir converts a module directory to a moniker
// Example: "src/cli" -> "src-cli"
func (m *Mapper) GetMonikerFromModuleDir(moduleDir string) (string, error) {
	// Look for module with matching source root
	for _, module := range m.registry.All() {
		if module.Source.Root == moduleDir {
			return module.Moniker, nil
		}
	}

	return "", fmt.Errorf("no module contract found for directory: %s", moduleDir)
}

// MapInternalDependenciesToMonikers converts a list of internal module paths to monikers
// Example: ["github.com/ready-to-release/eac/src/internal"] -> ["src-internal"]
func (m *Mapper) MapInternalDependenciesToMonikers(modulePaths []string) ([]string, []error) {
	var monikers []string
	var errors []error

	for _, path := range modulePaths {
		moniker, err := m.GetMonikerFromPath(path)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to map %s: %w", path, err))
			continue
		}
		monikers = append(monikers, moniker)
	}

	return monikers, errors
}

// IsInternalModulePath checks if a module path belongs to this repository
func (m *Mapper) IsInternalModulePath(modulePath string) bool {
	return IsInternalModule(modulePath, m.baseModulePath)
}

// GetBaseModulePath returns the base module path for this repository
func (m *Mapper) GetBaseModulePath() string {
	return m.baseModulePath
}

// GetAllMappings returns all path-to-moniker mappings
func (m *Mapper) GetAllMappings() map[string]string {
	// Return a copy to prevent external modification
	result := make(map[string]string, len(m.pathToMoniker))
	for k, v := range m.pathToMoniker {
		result[k] = v
	}
	return result
}
