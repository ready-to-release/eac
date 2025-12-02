// Package definitions handles repository-wide definitions and constraints
// loaded from .r2r/definitions.yml
package definitions

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// DefinitionsPath is the relative path to definitions file
const DefinitionsPath = ".r2r/definitions.yml"

// VersionConstraint defines how version bumping is constrained
type VersionConstraint string

const (
	// Unrestricted allows major, minor, and patch bumps based on conventional commits
	Unrestricted VersionConstraint = "unrestricted"
	// PatchOnly limits semver modules to patch bumps only (0.0.x → 0.0.x+1)
	PatchOnly VersionConstraint = "patch-only"
	// CalverOnly forces all modules to use calendar versioning
	CalverOnly VersionConstraint = "calver-only"
)

// Versioning contains versioning configuration
type Versioning struct {
	// Constraint controls how version bumping is constrained
	Constraint VersionConstraint `yaml:"constraint"`
}

// Definitions represents the repository-wide definitions
type Definitions struct {
	Versioning Versioning `yaml:"versioning"`
}

// Default returns default definitions when file is missing
func Default() *Definitions {
	return &Definitions{
		Versioning: Versioning{
			Constraint: Unrestricted,
		},
	}
}

// cached definitions and mutex for thread-safe loading
var (
	cachedDefs *Definitions
	cacheMutex sync.RWMutex
	cacheRoot  string
)

// Load loads definitions from the workspace root
// Returns default definitions if file doesn't exist
func Load(workspaceRoot string) (*Definitions, error) {
	cacheMutex.RLock()
	if cachedDefs != nil && cacheRoot == workspaceRoot {
		defer cacheMutex.RUnlock()
		return cachedDefs, nil
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if cachedDefs != nil && cacheRoot == workspaceRoot {
		return cachedDefs, nil
	}

	fullPath := filepath.Join(workspaceRoot, DefinitionsPath)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return defaults if file doesn't exist
			cachedDefs = Default()
			cacheRoot = workspaceRoot
			return cachedDefs, nil
		}
		return nil, err
	}

	var defs Definitions
	if err := yaml.Unmarshal(data, &defs); err != nil {
		return nil, err
	}

	// Validate constraint value
	if defs.Versioning.Constraint == "" {
		defs.Versioning.Constraint = Unrestricted
	}

	cachedDefs = &defs
	cacheRoot = workspaceRoot
	return cachedDefs, nil
}

// ClearCache clears the cached definitions (for testing)
func ClearCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cachedDefs = nil
	cacheRoot = ""
}

// IsPatchOnly returns true if versioning is constrained to patch-only
func (d *Definitions) IsPatchOnly() bool {
	return d.Versioning.Constraint == PatchOnly
}

// IsCalverOnly returns true if versioning is forced to calver
func (d *Definitions) IsCalverOnly() bool {
	return d.Versioning.Constraint == CalverOnly
}

// IsUnrestricted returns true if versioning is unrestricted
func (d *Definitions) IsUnrestricted() bool {
	return d.Versioning.Constraint == Unrestricted || d.Versioning.Constraint == ""
}
