package validator

import (
	"regexp"
	"slices"
	"strings"
)

// Compiled regex patterns from R2R CLI schema
var (
	// Extension name pattern: lowercase kebab-case
	ExtensionNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

	// Environment variable name pattern: uppercase with underscores
	EnvVarNamePattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

	// Version patterns
	SchemaVersionPattern   = regexp.MustCompile(`^[0-9]+\.[0-9]+$`)
	SemanticVersionPattern = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+(-[a-z0-9.]+)?(\+[a-z0-9.]+)?$|^latest$`)
	VersionPinPattern      = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?(\.[0-9]+)?$|^latest$`)

	// Resource limit patterns
	MemoryLimitPattern = regexp.MustCompile(`^[0-9]+[kmg]?$`)
	CPULimitPattern    = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

	// Timeout pattern
	TimeoutPattern = regexp.MustCompile(`^[0-9]+[smh]$`)

	// Command/option patterns
	CommandNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	LongOptionPattern  = regexp.MustCompile(`^--[a-z0-9][a-z0-9-]*$`)
	ShortOptionPattern = regexp.MustCompile(`^-[a-zA-Z]$`)

	// PowerShell module pattern
	PowerShellModulePattern = regexp.MustCompile(`^pwsh-[a-z0-9][a-z0-9-]*$`)

	// Test level pattern
	TestLevelPattern = regexp.MustCompile(`^L[0-4]$`)
)

// Valid enum values from R2R CLI schema
var (
	// Image pull policy values
	ValidImagePullPolicies = []string{"Always", "IfNotPresent", "Never", "AutoDetect"}

	// Network mode values
	ValidNetworkModes = []string{"bridge", "host", "none"}

	// Test output format values
	ValidTestOutputFormats = []string{"standard", "json", "junit", "tap"}

	// Option type values
	ValidOptionTypes = []string{"string", "number", "boolean", "array"}

	// Pester verbosity values
	ValidPesterVerbosity = []string{"None", "Normal", "Detailed", "Diagnostic"}

	// Pester output formats
	ValidPesterFormats = []string{"NUnitXml", "JUnitXml", "ConsoleOnly"}
)

// PatternValidator provides pattern validation utilities
type PatternValidator struct{}

// IsValidExtensionName validates extension name format
func (p PatternValidator) IsValidExtensionName(name string) bool {
	return ExtensionNamePattern.MatchString(name)
}

// IsValidEnvVarName validates environment variable name format
func (p PatternValidator) IsValidEnvVarName(name string) bool {
	return EnvVarNamePattern.MatchString(name)
}

// IsValidMemoryLimit validates memory limit format (e.g., "512m", "2g")
func (p PatternValidator) IsValidMemoryLimit(limit string) bool {
	if limit == "" {
		return true // Optional field
	}
	return MemoryLimitPattern.MatchString(limit)
}

// IsValidCPULimit validates CPU limit format (e.g., "0.5", "2.0")
func (p PatternValidator) IsValidCPULimit(limit string) bool {
	if limit == "" {
		return true // Optional field
	}
	return CPULimitPattern.MatchString(limit)
}

// IsValidTimeout validates timeout format (e.g., "30s", "5m", "1h")
func (p PatternValidator) IsValidTimeout(timeout string) bool {
	if timeout == "" {
		return true // Optional field
	}
	return TimeoutPattern.MatchString(timeout)
}

// IsValidSchemaVersion validates schema version format (e.g., "1.0")
func (p PatternValidator) IsValidSchemaVersion(version string) bool {
	if version == "" {
		return true // Optional field
	}
	return SchemaVersionPattern.MatchString(version)
}

// IsValidVersionPin validates version pinning format
func (p PatternValidator) IsValidVersionPin(version string) bool {
	if version == "" {
		return true // Optional field
	}
	return VersionPinPattern.MatchString(version) || version == "latest"
}

// IsValidTestLevel validates test level format (L0-L4)
func (p PatternValidator) IsValidTestLevel(level string) bool {
	return TestLevelPattern.MatchString(level)
}

// IsInEnum checks if a value is in a list of valid values
func IsInEnum(value string, validValues []string) bool {
	if value == "" {
		return true // Optional field
	}
	for _, valid := range validValues {
		if value == valid {
			return true
		}
	}
	return false
}

// IsValidPort checks if a port number is in valid range
func IsValidPort(port int) bool {
	return port >= 1 && port <= 65535
}

// IsValidURL provides basic URL validation
func IsValidURL(url string) bool {
	if url == "" {
		return true // Optional field
	}
	// Basic check for URL format
	return len(url) > 0 && (strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "github.com/"))
}

// Helper to check if string is in a slice
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
