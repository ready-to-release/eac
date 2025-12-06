// Package definitions handles repository-wide definitions and constraints
// loaded from .r2r/eac/repository.yml
package definitions

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"gopkg.in/yaml.v3"
)

// RepositoryConfigPath is the relative path to repository config file
const RepositoryConfigPath = ".r2r/eac/repository.yml"

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
// Reads versioning constraints from .r2r/eac/repository.yml
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

	fullPath := filepath.Join(workspaceRoot, RepositoryConfigPath)

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

	// Load repository contract
	var repoContract contracts.RepositoryContract
	if err := yaml.Unmarshal(data, &repoContract); err != nil {
		return nil, err
	}

	// Extract versioning constraint from repository config
	constraint := VersionConstraint(repoContract.Repository.Versioning.Constraint)
	if constraint == "" {
		constraint = Unrestricted
	}

	defs := &Definitions{
		Versioning: Versioning{
			Constraint: constraint,
		},
	}

	cachedDefs = defs
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
