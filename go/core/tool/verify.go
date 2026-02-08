// Package tool provides tool verification functionality.
package tool

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	container "github.com/ready-to-release/eac/contracts/container-runtime/0.1.0"
)

// verifyContainerPort holds an optional ContainerPort for Docker availability checks.
// When set, verifyDockerAvailable uses this instead of shelling out to "docker info".
var (
	verifyContainerPort   container.ContainerPort
	verifyContainerPortMu sync.RWMutex
)

// SetContainerPortForVerification injects a ContainerPort for tool verification.
// When set, verifyDockerAvailable() uses ContainerPort.IsAvailable() instead of
// exec.Command("docker", "info"). Pass nil to revert to the default CLI behavior.
func SetContainerPortForVerification(cp container.ContainerPort) {
	verifyContainerPortMu.Lock()
	defer verifyContainerPortMu.Unlock()
	verifyContainerPort = cp
}

// getVerifyContainerPort returns the injected ContainerPort, or nil if not set.
func getVerifyContainerPort() container.ContainerPort {
	verifyContainerPortMu.RLock()
	defer verifyContainerPortMu.RUnlock()
	return verifyContainerPort
}

// VerifyToolDefinition checks if a tool is available and meets version requirements.
// This is the core verification logic used by the registry.
func VerifyToolDefinition(tool *ToolDefinition) VerifyResult {
	result := VerifyResult{
		ToolID:          tool.ID,
		Available:       false,
		RequiredVersion: tool.Version,
	}

	// Check platform compatibility first
	if len(tool.Platforms) > 0 {
		currentPlatform := runtime.GOOS
		supported := false
		for _, p := range tool.Platforms {
			if p == currentPlatform {
				supported = true
				break
			}
		}
		if !supported {
			result.Skipped = true
			result.Error = fmt.Errorf("tool %q not available on platform %s (supports: %v)",
				tool.ID, currentPlatform, tool.Platforms)
			return result
		}
	}

	// If tool has explicit verify config, use it
	if tool.Verify != nil {
		if tool.Verify.IsCommandBased() {
			return verifyCommand(result, tool)
		}
		if tool.Verify.IsEnvBased() {
			return verifyEnvVars(result, tool)
		}
	}

	// Default verification based on tool type
	switch tool.Type {
	case ToolTypeSystem:
		return verifyBinaryExists(result, tool)
	case ToolTypeContainer:
		return verifyDockerAvailable(result, tool)
	default:
		result.Error = fmt.Errorf("unknown tool type: %s", tool.Type)
		return result
	}
}

// verifyCommand runs a command and extracts version using pattern.
func verifyCommand(result VerifyResult, tool *ToolDefinition) VerifyResult {
	verify := tool.Verify

	// Parse command into parts
	parts := strings.Fields(verify.Command)
	if len(parts) == 0 {
		result.Error = fmt.Errorf("empty verify command for %s", tool.ID)
		return result
	}

	cmd := exec.Command(parts[0], parts[1:]...) //nolint:gosec // G204: command from tool-config.yml
	output, err := cmd.CombinedOutput()
	if err != nil {
		// On Windows, try WSL fallback for compatible tools
		if runtime.GOOS == "windows" && isWSLCompatibleTool(parts[0]) && wslAvailable() {
			wslArgs := append([]string{parts[0]}, parts[1:]...)
			wslCmd := exec.Command("wsl", wslArgs...)
			wslOutput, wslErr := wslCmd.CombinedOutput()
			if wslErr == nil {
				output = wslOutput
				err = nil
			}
		}
		if err != nil {
			result.Error = fmt.Errorf("verification command failed: %w", err)
			return result
		}
	}

	result.Available = true
	outputStr := string(output)

	// Extract version using pattern
	if verify.Pattern != "" {
		re, err := regexp.Compile(verify.Pattern)
		if err != nil {
			result.Error = fmt.Errorf("invalid pattern: %w", err)
			return result
		}

		matches := re.FindStringSubmatch(outputStr)
		if len(matches) > 1 {
			result.Version = matches[1]
		}
	}

	// Check version meets requirement
	if tool.Version != "" && tool.Version != "any" && result.Version != "" {
		meetsVersion, err := compareVersions(result.Version, tool.Version)
		if err != nil {
			result.Error = fmt.Errorf("version comparison failed: %w", err)
			return result
		}
		if !meetsVersion {
			result.Available = false
			result.Error = fmt.Errorf("version %s does not meet requirement %s", result.Version, tool.Version)
		}
	}

	return result
}

// verifyEnvVars checks if required environment variables are set.
func verifyEnvVars(result VerifyResult, tool *ToolDefinition) VerifyResult {
	verify := tool.Verify
	requireMode := verify.Require
	if requireMode == "" {
		requireMode = "any"
	}

	var found []string
	for _, envVar := range verify.EnvVars {
		if val := os.Getenv(envVar); val != "" {
			found = append(found, envVar)
		}
	}

	switch requireMode {
	case "any":
		result.Available = len(found) > 0
	case "all":
		result.Available = len(found) == len(verify.EnvVars)
	}

	if !result.Available {
		result.Error = fmt.Errorf("required environment variable(s) not set")
	}

	return result
}

// verifyBinaryExists checks if a binary exists in PATH.
func verifyBinaryExists(result VerifyResult, tool *ToolDefinition) VerifyResult {
	if tool.Binary == "" {
		result.Error = fmt.Errorf("no binary specified for system tool %s", tool.ID)
		return result
	}

	_, err := exec.LookPath(tool.Binary)
	if err != nil {
		// On Windows, try WSL fallback for common Unix tools
		if runtime.GOOS == "windows" && isWSLCompatibleTool(tool.Binary) {
			if wslAvailable() {
				result.Available = true
				result.Version = "wsl"
				return result
			}
		}
		result.Error = fmt.Errorf("binary %q not found in PATH", tool.Binary)
		return result
	}

	result.Available = true
	return result
}

// verifyDockerAvailable checks if Docker is available for container tools.
// When a ContainerPort has been injected via SetContainerPortForVerification,
// it uses ContainerPort.IsAvailable() instead of shelling out to "docker info".
func verifyDockerAvailable(result VerifyResult, tool *ToolDefinition) VerifyResult {
	// Use injected ContainerPort if available
	if cp := getVerifyContainerPort(); cp != nil {
		if !cp.IsAvailable() {
			result.Error = fmt.Errorf("container runtime not available")
			return result
		}
		result.Available = true
		return result
	}

	// Fallback: direct CLI check (no ContainerPort injected)
	_, err := exec.LookPath("docker")
	if err != nil {
		result.Error = fmt.Errorf("docker not found in PATH")
		return result
	}

	cmd := exec.Command("docker", "info") //nolint:gosec // G204: legacy fallback, to be removed when all callers inject ContainerPort
	if err := cmd.Run(); err != nil {
		result.Error = fmt.Errorf("docker daemon not running: %w", err)
		return result
	}

	result.Available = true
	return result
}

// compareVersions checks if actual >= required.
// Handles version strings like "1.21", "1.21.0", ">=1.21".
func compareVersions(actual, required string) (bool, error) {
	// Strip comparison operators from required
	required = strings.TrimPrefix(required, ">=")
	required = strings.TrimPrefix(required, ">")
	required = strings.TrimPrefix(required, "=")

	actualParts := strings.Split(actual, ".")
	requiredParts := strings.Split(required, ".")

	// Need at least major.minor
	if len(actualParts) < 2 {
		return false, fmt.Errorf("invalid actual version format: %s", actual)
	}
	if len(requiredParts) < 2 {
		return false, fmt.Errorf("invalid required version format: %s", required)
	}

	actualMajor, err := strconv.Atoi(actualParts[0])
	if err != nil {
		return false, err
	}
	actualMinor, err := strconv.Atoi(actualParts[1])
	if err != nil {
		return false, err
	}

	requiredMajor, err := strconv.Atoi(requiredParts[0])
	if err != nil {
		return false, err
	}
	requiredMinor, err := strconv.Atoi(requiredParts[1])
	if err != nil {
		return false, err
	}

	if actualMajor > requiredMajor {
		return true, nil
	}
	if actualMajor == requiredMajor && actualMinor >= requiredMinor {
		return true, nil
	}

	return false, nil
}

// Package-level convenience functions using global registry

// Verify checks if a tool is available (convenience function using global registry).
func Verify(toolID string) VerifyResult {
	return GlobalRegistry().VerifyTool(toolID)
}

// VerifyAll checks multiple tools (convenience function using global registry).
func VerifyAll(toolIDs []string) []VerifyResult {
	return GlobalRegistry().VerifyAll(toolIDs)
}

// IsAvailable checks if a tool is available (convenience function using global registry).
func IsAvailable(toolID string) bool {
	return GlobalRegistry().IsAvailable(toolID)
}

// GetMissingDependencies returns list of unavailable tools (convenience function).
// This mirrors the old systemdeps API for easier migration.
func GetMissingDependencies(toolIDs []string) []string {
	return GlobalRegistry().GetMissingTools(toolIDs)
}

// wslCompatibleTools lists Unix tools that can run via WSL on Windows.
var wslCompatibleTools = map[string]bool{
	"bash": true,
	"sh":   true,
}

// isWSLCompatibleTool checks if a binary can be run via WSL.
func isWSLCompatibleTool(binary string) bool {
	return wslCompatibleTools[binary]
}

// wslAvailable checks if WSL is available and the tool can be run through it.
// Caches the result after first check.
var wslChecked bool
var wslIsAvailable bool

func wslAvailable() bool {
	if wslChecked {
		return wslIsAvailable
	}
	wslChecked = true

	// Check if wsl.exe exists
	_, err := exec.LookPath("wsl")
	if err != nil {
		wslIsAvailable = false
		return false
	}

	// Verify WSL can run bash
	cmd := exec.Command("wsl", "bash", "--version")
	if err := cmd.Run(); err != nil {
		wslIsAvailable = false
		return false
	}

	wslIsAvailable = true
	return true
}

// IsPlatformSupported checks if a tool is supported on the current platform.
// Returns true if the tool has no platform restrictions or if the current platform is in the list.
func IsPlatformSupported(tool *ToolDefinition) bool {
	if tool == nil || len(tool.Platforms) == 0 {
		return true
	}
	currentPlatform := runtime.GOOS
	for _, p := range tool.Platforms {
		if p == currentPlatform {
			return true
		}
	}
	return false
}

// FilterPlatformSupported filters a list of tool IDs to only include those supported on the current platform.
func FilterPlatformSupported(toolIDs []string) []string {
	registry := GlobalRegistry()
	var supported []string
	for _, toolID := range toolIDs {
		tool, ok := registry.Get(toolID)
		if !ok {
			// Unknown tool - include it so verification can report the error
			supported = append(supported, toolID)
			continue
		}
		if IsPlatformSupported(tool) {
			supported = append(supported, toolID)
		}
	}
	return supported
}
