// Package moduledeps provides module dependency verification
package moduledeps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// Verify checks if a module dependency is available
func Verify(dependency string) Result {
	result := Result{
		Dependency: dependency,
		Available:  false,
	}

	// Extract moniker from @depm:moniker format
	if !strings.HasPrefix(dependency, "@depm:") {
		result.Error = fmt.Errorf("invalid module dependency format: %s (expected @depm:<moniker>)", dependency)
		return result
	}

	moniker := strings.TrimPrefix(dependency, "@depm:")
	checker := &ModuleChecker{moniker: moniker}

	result.Available = checker.IsAvailable()
	if result.Available {
		version, err := checker.GetVersion()
		if err != nil {
			result.Error = err
		} else {
			result.Version = version
		}
	}

	return result
}

// VerifyAll checks multiple module dependencies
func VerifyAll(dependencies []string) []Result {
	results := make([]Result, len(dependencies))
	for i, dep := range dependencies {
		results[i] = Verify(dep)
	}
	return results
}

// IsAvailable quickly checks if a module dependency is available
func IsAvailable(dependency string) bool {
	result := Verify(dependency)
	return result.Available
}

// GetMissingDependencies returns list of unavailable module dependencies
func GetMissingDependencies(dependencies []string) []string {
	missing := []string{}
	for _, dep := range dependencies {
		if !IsAvailable(dep) {
			missing = append(missing, dep)
		}
	}
	return missing
}

// ModuleChecker checks if an internal module has been built
type ModuleChecker struct {
	moniker   string
	repoRoot  string
	eacConfig *config.EACConfig
}

func (c *ModuleChecker) GetName() string {
	return fmt.Sprintf("Module: %s", c.moniker)
}

// init initializes the checker with repo root and EAC config
func (c *ModuleChecker) init() error {
	if c.repoRoot != "" {
		return nil // Already initialized
	}

	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}
	c.repoRoot = repoRoot

	// Load EAC config to get module types
	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        repoRoot,
		ValidateSchemas: false,
		LazyLoad:        false, // Need module types loaded immediately
	})
	if err != nil {
		return fmt.Errorf("failed to load EAC config: %w", err)
	}
	c.eacConfig = cfg

	return nil
}

func (c *ModuleChecker) IsAvailable() bool {
	if err := c.init(); err != nil {
		return false
	}

	// Load module contract
	module, err := c.loadModuleContract()
	if err != nil {
		return false
	}

	// Get module type definition (with nil safety)
	var typeDef *config.ModuleTypeDef
	if c.eacConfig.ModuleTypes != nil {
		typeDef = c.eacConfig.ModuleTypes.Get(module.Type)
	}
	if typeDef == nil {
		// Fallback for unknown types or missing config: check if source root exists
		return c.checkSourceRootExists(module)
	}

	// If module type has no build artifacts, check source root exists
	if !typeDef.HasArtifacts() {
		return c.checkSourceRootExists(module)
	}

	// Verify build artifacts exist
	buildDir := repository.BuildOutputPath(c.repoRoot, c.moniker)
	resolver := config.NewArtifactResolverWithMetadata(c.moniker, buildDir, module.Metadata)
	results := resolver.VerifyArtifacts(typeDef.GetArtifacts())

	return config.AllSuccessful(results)
}

// checkSourceRootExists verifies the module's source directory exists
func (c *ModuleChecker) checkSourceRootExists(module *modules.ModuleContract) bool {
	// For modules without build artifacts, check source exists
	sourcePath := filepath.Join(c.repoRoot, module.Files.Root)
	_, err := os.Stat(sourcePath)
	return err == nil
}

func (c *ModuleChecker) GetVersion() (string, error) {
	if err := c.init(); err != nil {
		return "", err
	}

	// Load module contract
	module, err := c.loadModuleContract()
	if err != nil {
		return "", fmt.Errorf("failed to load module contract: %w", err)
	}

	// Get module type definition (with nil safety)
	var typeDef *config.ModuleTypeDef
	if c.eacConfig.ModuleTypes != nil {
		typeDef = c.eacConfig.ModuleTypes.Get(module.Type)
	}
	if typeDef == nil {
		sourcePath := filepath.Join(c.repoRoot, module.Files.Root)
		return fmt.Sprintf("Module source: %s", sourcePath), nil
	}

	// If no build artifacts, return source path
	if !typeDef.HasArtifacts() {
		sourcePath := filepath.Join(c.repoRoot, module.Files.Root)
		return fmt.Sprintf("%s: %s", typeDef.Description, sourcePath), nil
	}

	// Return info about build artifacts
	buildDir := repository.BuildOutputPath(c.repoRoot, c.moniker)
	resolver := config.NewArtifactResolverWithMetadata(c.moniker, buildDir, module.Metadata)
	results := resolver.VerifyArtifacts(typeDef.GetArtifacts())

	if !config.AllSuccessful(results) {
		failures := config.GetFailures(results)
		return "", fmt.Errorf("missing build artifacts:\n%s", config.FormatVerificationResults(failures))
	}

	// Return first successful artifact path as version info
	for _, r := range results {
		if r.Exists {
			absPath, _ := filepath.Abs(r.Path)
			return fmt.Sprintf("Built: %s", absPath), nil
		}
	}

	return fmt.Sprintf("Build directory: %s", buildDir), nil
}

// loadModuleContract loads the module contract from the registry
func (c *ModuleChecker) loadModuleContract() (*modules.ModuleContract, error) {
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Load module registry
	registry, err := modules.LoadFromWorkspaceLatest(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load module registry: %w", err)
	}

	// Get module by moniker
	module, found := registry.Get(c.moniker)
	if !found {
		return nil, fmt.Errorf("module not found: %s", c.moniker)
	}

	return module, nil
}
