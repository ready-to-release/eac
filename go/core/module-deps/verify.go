// Package moduledeps provides module dependency verification
package moduledeps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

// Verify checks if a module dependency is available.
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

// VerifyAll checks multiple module dependencies.
func VerifyAll(dependencies []string) []Result {
	results := make([]Result, len(dependencies))
	for i, dep := range dependencies {
		results[i] = Verify(dep)
	}
	return results
}

// IsAvailable quickly checks if a module dependency is available.
func IsAvailable(dependency string) bool {
	result := Verify(dependency)
	return result.Available
}

// GetMissingDependencies returns list of unavailable module dependencies.
func GetMissingDependencies(dependencies []string) []string {
	missing := []string{}
	for _, dep := range dependencies {
		if !IsAvailable(dep) {
			missing = append(missing, dep)
		}
	}
	return missing
}

// ModuleChecker checks if an internal module has been built.
type ModuleChecker struct {
	moniker   string
	repoRoot  string
	eacConfig *config.EACConfig
}

func (c *ModuleChecker) GetName() string {
	return fmt.Sprintf("Module: %s", c.moniker)
}

// init initializes the checker with repo root and EAC config.
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

	// Get build artifacts from module-level config
	artifacts := c.eacConfig.GetBuildArtifacts(c.moniker, true)
	if len(artifacts) == 0 {
		// No build artifacts defined: check if source root exists
		return c.checkSourceRootExists(module)
	}

	// Verify build artifacts exist
	buildDir := paths.BuildOutputPath(c.repoRoot, c.moniker)

	// For dependency checking, we accept any platform's artifacts
	// (unlike build validation which checks specific platforms)
	// This allows Docker containers to recognize modules built on different platforms
	for _, artifact := range artifacts {
		if artifact.Type == config.ArtifactTypeExecutable {
			// Check if ANY platform's artifacts exist for executables
			if c.checkAnyPlatformExists(artifact, buildDir, module.Metadata) {
				return true
			}
		} else {
			// Non-executables: use standard verification
			resolver := config.NewArtifactResolverWithMetadata(c.moniker, buildDir, module.Metadata)
			result := resolver.VerifyArtifact(artifact)
			if result.Exists {
				return true
			}
		}
	}

	return false
}

// checkAnyPlatformExists checks if ANY platform's artifacts exist for an executable
// This is used for dependency checking to allow cross-platform builds.
func (c *ModuleChecker) checkAnyPlatformExists(artifact config.Artifact, buildDir string, metadata map[string]string) bool {
	// Get platforms to check
	platforms := artifact.Platforms
	if len(platforms) == 0 {
		platforms = []string{"linux", "windows", "darwin"}
	}

	// Check each platform - return true if ANY exist
	for _, platform := range platforms {
		// Try both amd64 and arm64 for each platform
		archs := []string{"amd64"}
		if platform != "windows" {
			archs = append(archs, "arm64")
		}

		for _, arch := range archs {
			resolver := config.NewArtifactResolverFull(c.moniker, buildDir, platform, arch, metadata)
			result := resolver.VerifyArtifact(artifact)
			if result.Exists {
				return true
			}
		}
	}

	return false
}

// checkSourceRootExists verifies that at least one of the module's package directories exists.
func (c *ModuleChecker) checkSourceRootExists(module *modules.ModuleContract) bool {
	// Check if any package root exists
	for _, root := range module.GetComponentRoots() {
		sourcePath := filepath.Join(c.repoRoot, root)
		if _, err := os.Stat(sourcePath); err == nil {
			return true
		}
	}
	return false
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

	// Get build artifacts from module-level config
	artifacts := c.eacConfig.GetBuildArtifacts(c.moniker, true)
	if len(artifacts) == 0 {
		// No build artifacts: return first package root
		roots := module.GetComponentRoots()
		if len(roots) > 0 {
			for _, root := range roots {
				return fmt.Sprintf("Module source: %s", filepath.Join(c.repoRoot, root)), nil
			}
		}
		return fmt.Sprintf("Module: %s (no packages)", c.moniker), nil
	}

	// Return info about build artifacts
	buildDir := paths.BuildOutputPath(c.repoRoot, c.moniker)

	// For dependency checking, find any available artifact across platforms
	for _, artifact := range artifacts {
		if artifact.Type == config.ArtifactTypeExecutable {
			// Check all platforms
			platforms := artifact.Platforms
			if len(platforms) == 0 {
				platforms = []string{"linux", "windows", "darwin"}
			}

			for _, platform := range platforms {
				archs := []string{"amd64"}
				if platform != "windows" {
					archs = append(archs, "arm64")
				}

				for _, arch := range archs {
					resolver := config.NewArtifactResolverFull(c.moniker, buildDir, platform, arch, module.Metadata)
					result := resolver.VerifyArtifact(artifact)
					if result.Exists {
						absPath, err := filepath.Abs(result.Path)
						if err != nil {
							absPath = result.Path // Fallback to relative path
						}
						return fmt.Sprintf("Built: %s", absPath), nil
					}
				}
			}
		} else {
			// Non-executables: check normally
			resolver := config.NewArtifactResolverWithMetadata(c.moniker, buildDir, module.Metadata)
			result := resolver.VerifyArtifact(artifact)
			if result.Exists {
				absPath, err := filepath.Abs(result.Path)
				if err != nil {
					absPath = result.Path // Fallback to relative path
				}
				return fmt.Sprintf("Built: %s", absPath), nil
			}
		}
	}

	// No artifacts found
	return "", fmt.Errorf("no build artifacts found in: %s", buildDir)
}

// loadModuleContract loads the module contract from the registry.
func (c *ModuleChecker) loadModuleContract() (*modules.ModuleContract, error) {
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Load module registry
	registry, err := modules.LoadFromWorkspace(repoRoot)
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
