// Package moduledeps provides module dependency verification
package moduledeps

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/repository"
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
	moniker string
}

func (c *ModuleChecker) GetName() string {
	return fmt.Sprintf("Module: %s", c.moniker)
}

func (c *ModuleChecker) IsAvailable() bool {
	// Load module contract
	module, err := c.loadModuleContract()
	if err != nil {
		return false
	}

	// Check based on module type
	switch module.Type {
	case "go-cli":
		// CLI modules: check if binary exists
		path := c.getExecutablePath(module)
		if path == "" {
			return false
		}
		_, err := os.Stat(path)
		return err == nil

	case "go-commands":
		// Commands modules: always run with 'go run', check if main.go exists
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return false
		}
		mainGoPath := filepath.Join(repoRoot, module.Files.Root, "main.go")
		_, err = os.Stat(mainGoPath)
		return err == nil

	case "go-library":
		// Library modules: check if go.mod exists (module is properly set up)
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return false
		}
		goModPath := filepath.Join(repoRoot, module.Files.Root, "go.mod")
		_, err = os.Stat(goModPath)
		return err == nil

	case "go-mcp":
		// MCP modules: check if main.go exists
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return false
		}
		mainGoPath := filepath.Join(repoRoot, module.Files.Root, "main.go")
		_, err = os.Stat(mainGoPath)
		return err == nil

	case "repository-root":
		// Repository root module: check if repository tests directory exists
		testImplPath := module.GetTestImplementationPath()
		if testImplPath == "" {
			return false
		}
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return false
		}
		testsPath := filepath.Join(repoRoot, testImplPath)
		_, err = os.Stat(testsPath)
		return err == nil

	case "scripts-package":
		// Scripts modules: check if source root directory exists
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return false
		}
		scriptsPath := filepath.Join(repoRoot, module.Files.Root)
		_, err = os.Stat(scriptsPath)
		return err == nil

	default:
		// Unknown module type
		return false
	}
}

func (c *ModuleChecker) GetVersion() (string, error) {
	// Load module contract
	module, err := c.loadModuleContract()
	if err != nil {
		return "", fmt.Errorf("failed to load module contract: %w", err)
	}

	// Check based on module type
	switch module.Type {
	case "go-cli":
		// CLI modules: check for built executable
		path := c.getExecutablePath(module)
		if path == "" {
			return "", fmt.Errorf("module executable path not found (repo root not detected)")
		}

		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("module executable not found at %s", path)
		}

		absPath, _ := filepath.Abs(path)
		return fmt.Sprintf("Built executable: %s", absPath), nil

	case "go-commands":
		// Commands modules: run with 'go run'
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return "", err
		}
		modulePath := filepath.Join(repoRoot, module.Files.Root)
		return fmt.Sprintf("Commands module (go run): %s", modulePath), nil

	case "go-library":
		// Library modules: return module path
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return "", err
		}
		modulePath := filepath.Join(repoRoot, module.Files.Root)
		return fmt.Sprintf("Go library: %s", modulePath), nil

	case "go-mcp":
		// MCP modules: return module path
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return "", err
		}
		modulePath := filepath.Join(repoRoot, module.Files.Root)
		return fmt.Sprintf("MCP module: %s", modulePath), nil

	case "repository-root":
		// Repository root module: return tests path
		testImplPath := module.GetTestImplementationPath()
		if testImplPath == "" {
			return "", fmt.Errorf("module '%s' has no test_impl path defined", module.Moniker)
		}
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return "", err
		}
		testsPath := filepath.Join(repoRoot, testImplPath)
		return fmt.Sprintf("Repository validation tests: %s", testsPath), nil

	case "scripts-package":
		// Scripts modules: return scripts path
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return "", err
		}
		scriptsPath := filepath.Join(repoRoot, module.Files.Root)
		return fmt.Sprintf("Scripts: %s", scriptsPath), nil

	default:
		return "", fmt.Errorf("module type '%s' verification not implemented", module.Type)
	}
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

// getExecutablePath returns the full path to the module's executable
// For multi-platform builds, looks for platform-specific binaries using metadata from module contract.
// Metadata keys: exe-linux, exe-windows, exe-darwin-amd64, exe-darwin-arm64
func (c *ModuleChecker) getExecutablePath(module *modules.ModuleContract) string {
	// Find repository root using the centralized utility
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return ""
	}

	// Try to get executable name from module metadata
	exeName := c.getExecutableNameFromMetadata(module)
	buildDir := repository.BuildOutputPath(repoRoot, c.moniker)

	if exeName != "" {
		// Use metadata-defined executable name
		exePath := filepath.Join(buildDir, exeName)
		if _, err := os.Stat(exePath); err == nil {
			return exePath
		}
	}

	// Fallback: derive executable name from moniker
	var ext string
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	// Try platform-specific binary (format: {moniker}-{os}{ext})
	platformBinary := fmt.Sprintf("%s-%s%s", c.moniker, runtime.GOOS, ext)
	platformPath := filepath.Join(buildDir, platformBinary)
	if _, err := os.Stat(platformPath); err == nil {
		return platformPath
	}

	// Try architecture-specific variants for darwin (darwin-amd64, darwin-arm64)
	if runtime.GOOS == "darwin" {
		archBinary := fmt.Sprintf("%s-%s-%s", c.moniker, runtime.GOOS, runtime.GOARCH)
		archPath := filepath.Join(buildDir, archBinary)
		if _, err := os.Stat(archPath); err == nil {
			return archPath
		}
	}

	// Return the primary expected path (even if it doesn't exist, for error messages)
	return platformPath
}

// getExecutableNameFromMetadata looks up the executable name from module metadata.
// Uses keys like exe-linux, exe-windows, exe-darwin-amd64, exe-darwin-arm64.
func (c *ModuleChecker) getExecutableNameFromMetadata(module *modules.ModuleContract) string {
	if module.Metadata == nil {
		return ""
	}

	// For darwin, try architecture-specific key first (exe-darwin-amd64, exe-darwin-arm64)
	if runtime.GOOS == "darwin" {
		archKey := fmt.Sprintf("exe-%s-%s", runtime.GOOS, runtime.GOARCH)
		if name, ok := module.Metadata[archKey]; ok {
			return name
		}
	}

	// Try OS-specific key (exe-linux, exe-windows, exe-darwin)
	osKey := fmt.Sprintf("exe-%s", runtime.GOOS)
	if name, ok := module.Metadata[osKey]; ok {
		return name
	}

	return ""
}
