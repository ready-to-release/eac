// Package systemdeps provides system dependency verification
package systemdeps

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/go-git/go-git/v5"
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

// GitChecker checks for Git functionality via go-git library
type GitChecker struct{}

func (c *GitChecker) GetName() string { return "Git (go-git)" }

func (c *GitChecker) IsAvailable() bool {
	// go-git is always available as it's a pure Go implementation
	// We verify by checking that the library is importable and functional
	// This is effectively always true since we compiled successfully
	return true
}

func (c *GitChecker) GetVersion() (string, error) {
	// Return go-git version info
	// go-git doesn't expose its version directly, so we return a descriptive string
	return fmt.Sprintf("go-git/v5 (pure Go, %s/%s)", runtime.GOOS, runtime.GOARCH), nil
}

// Ensure go-git is used (prevents unused import error)
var _ = git.ErrRepositoryNotExists

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
