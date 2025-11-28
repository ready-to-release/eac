// Package systemdeps provides system dependency verification
package systemdeps

import (
	"fmt"
	"os"
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
	case "@deps:ai":
		return &AIChecker{}
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

// GitChecker checks for Git command-line tool
type GitChecker struct{}

func (c *GitChecker) GetName() string { return "Git" }

func (c *GitChecker) IsAvailable() bool {
	cmd := exec.Command("git", "version")
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

// AIChecker checks for ANY available AI provider
type AIChecker struct{}

func (c *AIChecker) GetName() string { return "AI Provider" }

func (c *AIChecker) IsAvailable() bool {
	// Check for test AI mock (for unit/integration tests)
	if os.Getenv("TEST_AI_MOCK") != "" {
		return true
	}

	// Check for Claude CLI
	if exec.Command("claude", "--version").Run() == nil {
		return true
	}

	// Check for API keys (any provider)
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return true
	}
	if os.Getenv("OPENAI_API_KEY") != "" {
		return true
	}
	if os.Getenv("GOOGLE_API_KEY") != "" {
		return true
	}

	return false
}

func (c *AIChecker) GetVersion() (string, error) {
	// Check for test AI mock first
	if mockValue := os.Getenv("TEST_AI_MOCK"); mockValue != "" {
		return fmt.Sprintf("test-ai-mock (%s)", mockValue), nil
	}

	// Return info about first available provider
	if output, err := exec.Command("claude", "--version").Output(); err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return "claude-api (ANTHROPIC_API_KEY)", nil
	}
	if os.Getenv("OPENAI_API_KEY") != "" {
		return "openai (OPENAI_API_KEY)", nil
	}
	if os.Getenv("GOOGLE_API_KEY") != "" {
		return "gemini (GOOGLE_API_KEY)", nil
	}

	return "", fmt.Errorf("no AI provider available")
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
