// Package systemdeps provides system dependency verification
package systemdeps

import (
	"fmt"
	"os/exec"
	"strings"
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

// getChecker returns the appropriate checker for a system dependency tag
func getChecker(dependency string) Checker {
	switch dependency {
	case "@deps:docker":
		return &DockerChecker{}
	case "@deps:git":
		return &GitChecker{}
	case "@deps:go":
		return &GoChecker{}
	case "@deps:claude":
		return &ClaudeChecker{}
	case "@deps:az-cli":
		return &AzureChecker{}
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
