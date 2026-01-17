// Package systemdeps provides system dependency verification
package systemdeps

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// Verifier checks system dependencies using configuration.
type Verifier struct {
	config *config.SystemDependenciesConfig
}

// NewVerifier creates a new verifier with the given config.
func NewVerifier(cfg *config.SystemDependenciesConfig) *Verifier {
	return &Verifier{config: cfg}
}

// Verify checks if a system dependency is available and meets version requirements
// The dependency can be specified as either "go" or "@deps:go".
func (v *Verifier) Verify(dependency string) Result {
	// Normalize: strip @deps: prefix if present
	moniker := strings.TrimPrefix(dependency, "@deps:")

	result := Result{
		Dependency: dependency,
		Moniker:    moniker,
		Available:  false,
	}

	// Look up in config
	dep := v.config.Get(moniker)
	if dep == nil {
		result.Error = fmt.Errorf("unknown dependency: %s", moniker)
		return result
	}

	result.Name = dep.Name
	result.RequiredVersion = dep.Version

	// Run verification
	if dep.Verify.IsOSPlatformBased() {
		result = v.verifyOSPlatform(result, dep)
	} else if dep.Verify.IsCommandBased() {
		result = v.verifyCommand(result, dep)
	} else if dep.Verify.IsEnvBased() {
		result = v.verifyEnvVars(result, dep)
	} else {
		result.Error = fmt.Errorf("no verification method defined for %s", moniker)
	}

	return result
}

// verifyCommand runs a command and extracts version using pattern.
func (v *Verifier) verifyCommand(result Result, dep *config.SystemDependency) Result {
	// Parse command into parts
	parts := strings.Fields(dep.Verify.Command)
	if len(parts) == 0 {
		result.Error = fmt.Errorf("empty verify command for %s", dep.Moniker)
		return result
	}

	cmd := exec.Command(parts[0], parts[1:]...) //nolint:gosec // G204: command from system-dependencies.yml config
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Error = fmt.Errorf("command failed: %w", err)
		return result
	}

	result.Available = true
	outputStr := string(output)

	// Extract version using pattern
	if dep.Verify.Pattern != "" {
		re, err := regexp.Compile(dep.Verify.Pattern)
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
	if dep.Version != "any" && result.Version != "" {
		meetsVersion, err := compareVersions(result.Version, dep.Version)
		if err != nil {
			result.Error = fmt.Errorf("version comparison failed: %w", err)
			return result
		}
		if !meetsVersion {
			result.Available = false
			result.Error = fmt.Errorf("version %s does not meet requirement %s", result.Version, dep.Version)
		}
	}

	return result
}

// verifyOSPlatform checks if the current OS matches the required platform.
func (v *Verifier) verifyOSPlatform(result Result, dep *config.SystemDependency) Result {
	requiredOS := dep.Verify.OSPlatform
	currentOS := runtime.GOOS

	result.Available = currentOS == requiredOS
	result.Version = currentOS // Report current OS as "version"

	if !result.Available {
		result.Error = fmt.Errorf("requires %s, running on %s", requiredOS, currentOS)
	}

	return result
}

// verifyEnvVars checks if required environment variables are set.
func (v *Verifier) verifyEnvVars(result Result, dep *config.SystemDependency) Result {
	requireMode := dep.Verify.Require
	if requireMode == "" {
		requireMode = "any"
	}

	var found []string
	for _, envVar := range dep.Verify.EnvVars {
		if val := os.Getenv(envVar); val != "" {
			found = append(found, envVar)
		}
	}

	switch requireMode {
	case "any":
		result.Available = len(found) > 0
	case "all":
		result.Available = len(found) == len(dep.Verify.EnvVars)
	}

	if !result.Available {
		result.Error = fmt.Errorf("required environment variable(s) not set")
	}

	return result
}

// VerifyAll checks multiple dependencies.
func (v *Verifier) VerifyAll(dependencies []string) []Result {
	results := make([]Result, len(dependencies))
	for i, dep := range dependencies {
		results[i] = v.Verify(dep)
	}
	return results
}

// IsAvailable quickly checks if a dependency is available.
func (v *Verifier) IsAvailable(dependency string) bool {
	result := v.Verify(dependency)
	return result.Available
}

// GetMissingDependencies returns list of unavailable dependencies.
func (v *Verifier) GetMissingDependencies(dependencies []string) []string {
	missing := []string{}
	for _, dep := range dependencies {
		if !v.IsAvailable(dep) {
			missing = append(missing, dep)
		}
	}
	return missing
}

// Package-level convenience functions that auto-load config

// Verify checks if a system dependency is available (convenience function)
// Loads config automatically. For repeated checks, use NewVerifier instead.
func Verify(dependency string) Result {
	cfg, err := loadConfig()
	if err != nil {
		return Result{
			Dependency: dependency,
			Available:  false,
			Error:      fmt.Errorf("failed to load config: %w", err),
		}
	}
	return NewVerifier(cfg).Verify(dependency)
}

// VerifyAll checks multiple dependencies (convenience function).
func VerifyAll(dependencies []string) []Result {
	cfg, err := loadConfig()
	if err != nil {
		results := make([]Result, len(dependencies))
		for i, dep := range dependencies {
			results[i] = Result{
				Dependency: dep,
				Available:  false,
				Error:      fmt.Errorf("failed to load config: %w", err),
			}
		}
		return results
	}
	return NewVerifier(cfg).VerifyAll(dependencies)
}

// IsAvailable quickly checks if a dependency is available (convenience function).
func IsAvailable(dependency string) bool {
	return Verify(dependency).Available
}

// GetMissingDependencies returns list of unavailable dependencies (convenience function).
func GetMissingDependencies(dependencies []string) []string {
	cfg, err := loadConfig()
	if err != nil {
		return dependencies // Assume all missing if config fails
	}
	return NewVerifier(cfg).GetMissingDependencies(dependencies)
}

// loadConfig loads the system dependencies config.
func loadConfig() (*config.SystemDependenciesConfig, error) {
	eacCfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return nil, err
	}
	if eacCfg.SystemDependencies == nil {
		return nil, fmt.Errorf("system dependencies config not loaded")
	}
	return eacCfg.SystemDependencies, nil
}

// compareVersions checks if actual >= required (both in major.minor format).
func compareVersions(actual, required string) (bool, error) {
	actualParts := strings.Split(actual, ".")
	requiredParts := strings.Split(required, ".")

	if len(actualParts) < 2 || len(requiredParts) < 2 {
		return false, fmt.Errorf("invalid version format")
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
