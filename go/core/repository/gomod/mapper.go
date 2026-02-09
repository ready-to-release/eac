package gomod

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// Mapper handles mapping between go.mod module paths and module contract monikers.
type Mapper struct {
	registry       *modules.Registry
	baseModulePath string // e.g., "github.com/ready-to-release/eac"
	pathToMoniker  map[string]string
	monikerToPath  map[string]string
}

// NewMapper creates a new mapper with the given registry and base module path.
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

// buildMappings builds bidirectional lookup maps between module paths and monikers.
func (m *Mapper) buildMappings() {
	for _, module := range m.registry.All() {
		// Build expected module path from go package root
		// Example: go package root = "go/cli/clie" -> module path = "github.com/ready-to-release/eac/go/cli/clie"
		goRoot := module.GetComponentRoot("go")
		if goRoot == "" {
			continue // Skip modules without go package
		}
		modulePath := m.baseModulePath + "/" + goRoot

		// Store mappings
		m.pathToMoniker[modulePath] = module.Moniker
		m.monikerToPath[module.Moniker] = modulePath

		// Also index additional go-type components within a module
		for _, entry := range module.Components.GetComponentsByType("go") {
			if entry == nil || entry.Root == "" || entry.Root == goRoot {
				continue
			}
			m.pathToMoniker[m.baseModulePath+"/"+entry.Root] = module.Moniker
		}
	}
}

// GetMonikerFromPath converts a go.mod module path to a module moniker
// Example: "github.com/ready-to-release/eac/go/cli/clie" -> "clie".
func (m *Mapper) GetMonikerFromPath(modulePath string) (string, error) {
	if moniker, ok := m.pathToMoniker[modulePath]; ok {
		return moniker, nil
	}

	// Try to find by extracting relative path
	relPath := strings.TrimPrefix(modulePath, m.baseModulePath+"/")

	// Single pass: check exact match and track best prefix match
	var bestMatch *modules.ModuleContract
	bestMatchLen := 0

	for _, module := range m.registry.All() {
		for _, root := range allGoRoots(module) {
			if root == relPath {
				return module.Moniker, nil
			}
			if root != "" && root != "." && strings.HasPrefix(relPath, root+"/") {
				if len(root) > bestMatchLen {
					bestMatch = module
					bestMatchLen = len(root)
				}
			}
		}
	}

	if bestMatch != nil {
		return bestMatch.Moniker, nil
	}

	return "", fmt.Errorf("no module contract found for path: %s", modulePath)
}

// GetPathFromMoniker converts a module moniker to a go.mod module path
// Example: "clie" -> "github.com/ready-to-release/eac/go/cli/clie".
func (m *Mapper) GetPathFromMoniker(moniker string) (string, error) {
	if path, ok := m.monikerToPath[moniker]; ok {
		return path, nil
	}

	// Try to find module and construct path
	module, exists := m.registry.Get(moniker)
	if !exists {
		return "", fmt.Errorf("module not found: %s", moniker)
	}

	goRoot := module.GetComponentRoot("go")
	if goRoot == "" {
		return "", fmt.Errorf("module %s has no go package", moniker)
	}

	return m.baseModulePath + "/" + goRoot, nil
}

// GetMonikerFromModuleDir converts a module directory to a moniker
// Example: "go/cli/clie" -> "clie"
// Also handles subdirectories: "go/eac/core/ai" -> "core".
func (m *Mapper) GetMonikerFromModuleDir(moduleDir string) (string, error) {
	// Normalize path separators
	normalizedDir := strings.ReplaceAll(moduleDir, "\\", "/")

	// Single pass: check exact match and track best prefix match
	var bestMatch *modules.ModuleContract
	bestMatchLen := 0

	for _, module := range m.registry.All() {
		for _, root := range allGoRoots(module) {
			if root == normalizedDir {
				return module.Moniker, nil
			}
			if root != "" && root != "." && strings.HasPrefix(normalizedDir, root+"/") {
				if len(root) > bestMatchLen {
					bestMatch = module
					bestMatchLen = len(root)
				}
			}
		}
	}

	if bestMatch != nil {
		return bestMatch.Moniker, nil
	}

	return "", fmt.Errorf("no module contract found for directory: %s", moduleDir)
}

// MapInternalDependenciesToMonikers converts a list of internal module paths to monikers
// Example: ["github.com/ready-to-release/eac/go/core"] -> ["core"].
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

// IsInternalModulePath checks if a module path belongs to this repository.
func (m *Mapper) IsInternalModulePath(modulePath string) bool {
	return IsInternalModule(modulePath, m.baseModulePath)
}

// GetBaseModulePath returns the base module path for this repository.
func (m *Mapper) GetBaseModulePath() string {
	return m.baseModulePath
}

// GetAllMappings returns all path-to-moniker mappings.
func (m *Mapper) GetAllMappings() map[string]string {
	// Return a copy to prevent external modification
	result := make(map[string]string, len(m.pathToMoniker))
	for k, v := range m.pathToMoniker {
		result[k] = v
	}
	return result
}

// allGoRoots returns all root paths for go-type components in a module.
// This includes both the primary "go" component and any additional components
// with type "go".
func allGoRoots(module *modules.ModuleContract) []string {
	goComps := module.Components.GetComponentsByType("go")
	roots := make([]string, 0, len(goComps))
	for _, entry := range goComps {
		if entry != nil && entry.Root != "" {
			roots = append(roots, entry.Root)
		}
	}
	return roots
}
