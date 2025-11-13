// Package systemdeps provides system dependency verification
package systemdeps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/src/core/repository"
)

// Verify checks if a system dependency is available
func Verify(dependency string) Result {
	result := Result{
		Dependency: dependency,
		Available:  false,
	}

	checker := getChecker(dependency)
	if checker == nil {
		result.Error = fmt.Errorf("unknown dependency: %s", dependency)
		return result
	}

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

// VerifyAll checks multiple dependencies
func VerifyAll(dependencies []string) []Result {
	results := make([]Result, len(dependencies))
	for i, dep := range dependencies {
		results[i] = Verify(dep)
	}
	return results
}

// IsAvailable quickly checks if a dependency is available
func IsAvailable(dependency string) bool {
	result := Verify(dependency)
	return result.Available
}

// GetMissingDependencies returns list of unavailable dependencies
func GetMissingDependencies(dependencies []string) []string {
	missing := []string{}
	for _, dep := range dependencies {
		if !IsAvailable(dep) {
			missing = append(missing, dep)
		}
	}
	return missing
}

// getChecker returns the appropriate checker for a dependency tag
func getChecker(dependency string) Checker {
	switch dependency {
	case "@dep:docker":
		return &DockerChecker{}
	case "@dep:git":
		return &GitChecker{}
	case "@dep:go":
		return &GoChecker{}
	case "@dep:claude":
		return &ClaudeChecker{}
	case "@dep:az-cli":
		return &AzureChecker{}
	case "@dep:internal-src-cli":
		return &InternalSrcCLIChecker{}
	default:
		return nil
	}
}

// DockerChecker checks for Docker
type DockerChecker struct{}

func (c *DockerChecker) GetName() string { return "Docker" }

func (c *DockerChecker) IsAvailable() bool {
	// Check if Docker daemon is running, not just if CLI is installed
	cmd := exec.Command("docker", "ps")
	return cmd.Run() == nil
}

func (c *DockerChecker) GetVersion() (string, error) {
	cmd := exec.Command("docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GitChecker checks for Git
type GitChecker struct{}

func (c *GitChecker) GetName() string { return "Git" }

func (c *GitChecker) IsAvailable() bool {
	cmd := exec.Command("git", "--version")
	return cmd.Run() == nil
}

func (c *GitChecker) GetVersion() (string, error) {
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GoChecker checks for Go
type GoChecker struct{}

func (c *GoChecker) GetName() string { return "Go" }

func (c *GoChecker) IsAvailable() bool {
	cmd := exec.Command("go", "version")
	return cmd.Run() == nil
}

func (c *GoChecker) GetVersion() (string, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// ClaudeChecker checks for Claude CLI access
type ClaudeChecker struct{}

func (c *ClaudeChecker) GetName() string { return "Claude CLI" }

func (c *ClaudeChecker) IsAvailable() bool {
	// Check if claude CLI tool is available
	// Note: This uses subscription auth, NOT API key
	// See docs/reference/modules/src-commands/claude-constraints.md
	cmd := exec.Command("claude", "--version")
	return cmd.Run() == nil
}

func (c *ClaudeChecker) GetVersion() (string, error) {
	cmd := exec.Command("claude", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// AzureChecker checks for Azure CLI
type AzureChecker struct{}

func (c *AzureChecker) GetName() string { return "Azure CLI" }

func (c *AzureChecker) IsAvailable() bool {
	cmd := exec.Command("az", "--version")
	return cmd.Run() == nil
}

func (c *AzureChecker) GetVersion() (string, error) {
	cmd := exec.Command("az", "version", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// InternalSrcCLIChecker checks for built CLI executable
type InternalSrcCLIChecker struct{}

func (c *InternalSrcCLIChecker) GetName() string { return "Internal SRC CLI" }

func (c *InternalSrcCLIChecker) IsAvailable() bool {
	path := c.getExecutablePath()
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func (c *InternalSrcCLIChecker) GetVersion() (string, error) {
	path := c.getExecutablePath()
	if path == "" {
		return "", fmt.Errorf("CLI executable path not found (repo root not detected)")
	}

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("CLI executable not found at %s", path)
	}

	absPath, _ := filepath.Abs(path)
	return fmt.Sprintf("Built executable: %s", absPath), nil
}

// getExecutablePath returns the full path to the CLI executable
func (c *InternalSrcCLIChecker) getExecutablePath() string {
	// Find repository root using the centralized utility
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return ""
	}

	// Determine executable name based on OS
	var exeName string
	if runtime.GOOS == "windows" {
		exeName = "r2r-cli.exe"
	} else {
		exeName = "r2r-cli"
	}

	return filepath.Join(repoRoot, "out", "src-cli", exeName)
}
