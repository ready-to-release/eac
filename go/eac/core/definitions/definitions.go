// Package definitions handles repository-wide definitions and constraints
// loaded from .r2r/eac/repository.yml with defaults from contracts
package definitions

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"gopkg.in/yaml.v3"
)

// RepositoryConfigPath is the relative path to repository config file
const RepositoryConfigPath = ".r2r/eac/repository.yml"

// DefaultsPath is the relative path to contract defaults
const DefaultsPath = "contracts/eac-core/0.1.0/defaults/repository.yml"

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

// cached definitions and mutex for thread-safe loading
var (
	cachedDefs *Definitions
	cacheMutex sync.RWMutex
	cacheRoot  string
)

// defaultsRoot returns the root for loading contract defaults (container-aware)
func defaultsRoot(repoRoot string) string {
	if containerRoot := os.Getenv("R2R_CONTAINER_ROOT"); containerRoot != "" {
		return containerRoot
	}
	return repoRoot
}

// loadDefaults loads defaults from contract defaults file
func loadDefaults(workspaceRoot string) (*Definitions, error) {
	root := defaultsRoot(workspaceRoot)
	defaultsPath := filepath.Join(root, DefaultsPath)

	data, err := os.ReadFile(defaultsPath)
	if err != nil {
		return nil, fmt.Errorf("reading defaults: %w", err)
	}

	var repoContract contracts.RepositoryContract
	if err := yaml.Unmarshal(data, &repoContract); err != nil {
		return nil, fmt.Errorf("parsing defaults: %w", err)
	}

	return &Definitions{
		Versioning: Versioning{
			Constraint: VersionConstraint(repoContract.Repository.Versioning.Constraint),
		},
	}, nil
}

// Load loads definitions from the workspace root
// Reads versioning constraints from .r2r/eac/repository.yml merged with contract defaults
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

	// Load defaults from contract
	defaults, err := loadDefaults(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("loading defaults: %w", err)
	}

	// Try to load user config
	fullPath := filepath.Join(workspaceRoot, RepositoryConfigPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Use defaults only
			cachedDefs = defaults
			cacheRoot = workspaceRoot
			return cachedDefs, nil
		}
		return nil, err
	}

	// Load user repository contract
	var repoContract contracts.RepositoryContract
	if err := yaml.Unmarshal(data, &repoContract); err != nil {
		return nil, err
	}

	// Merge: user config overrides defaults
	defs := defaults
	if repoContract.Repository.Versioning.Constraint != "" {
		defs.Versioning.Constraint = VersionConstraint(repoContract.Repository.Versioning.Constraint)
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
