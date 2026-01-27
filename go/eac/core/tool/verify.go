// Package tool provides tool verification functionality.
package tool

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// VerifyToolDefinition checks if a tool is available and meets version requirements.
// This is the core verification logic used by the registry.
func VerifyToolDefinition(tool *ToolDefinition) VerifyResult {
	result := VerifyResult{
		ToolID:          tool.ID,
		Available:       false,
		RequiredVersion: tool.Version,
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
		result.Error = fmt.Errorf("verification command failed: %w", err)
		return result
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
		result.Error = fmt.Errorf("binary %q not found in PATH", tool.Binary)
		return result
	}

	result.Available = true
	return result
}

// verifyDockerAvailable checks if Docker is available for container tools.
func verifyDockerAvailable(result VerifyResult, tool *ToolDefinition) VerifyResult {
	// Check if docker command exists
	_, err := exec.LookPath("docker")
	if err != nil {
		result.Error = fmt.Errorf("docker not found in PATH")
		return result
	}

	// Check if Docker daemon is running
	cmd := exec.Command("docker", "info")
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
