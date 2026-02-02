// Package interfaces provides type definitions for security scanning.
package interfaces

import "time"

// ScannerDefinition defines a security scanner configuration.
// This is the concrete implementation of ScannerPort.
type ScannerDefinition struct {
	// IDValue is the unique scanner identifier (e.g., "trivy-sbom").
	IDValue string `yaml:"id,omitempty"`

	// CategoryValue is the scanner category (e.g., "sbom", "vuln", "sast").
	CategoryValue string `yaml:"category,omitempty"`

	// ImageValue is the Docker image name.
	ImageValue string `yaml:"image"`

	// TagValue is the Docker image tag/version.
	TagValue string `yaml:"tag"`

	// CommandValue is the scanner command arguments.
	CommandValue []string `yaml:"command"`

	// TimeoutValue is the maximum execution duration as a string (e.g., "10m").
	TimeoutValue string `yaml:"timeout"`

	// DescriptionValue is a human-readable description.
	DescriptionValue string `yaml:"description,omitempty"`
}

// ID returns the scanner identifier.
func (s *ScannerDefinition) ID() string { return s.IDValue }

// Category returns the scanner category.
func (s *ScannerDefinition) Category() string { return s.CategoryValue }

// Image returns the Docker image name.
func (s *ScannerDefinition) Image() string { return s.ImageValue }

// Tag returns the Docker image tag.
func (s *ScannerDefinition) Tag() string { return s.TagValue }

// FullImage returns the complete image reference.
func (s *ScannerDefinition) FullImage() string {
	if s.TagValue == "" {
		return s.ImageValue
	}
	return s.ImageValue + ":" + s.TagValue
}

// Command returns the scanner command arguments.
func (s *ScannerDefinition) Command() []string { return s.CommandValue }

// Timeout returns the maximum execution duration.
func (s *ScannerDefinition) Timeout() time.Duration {
	d, err := time.ParseDuration(s.TimeoutValue)
	if err != nil {
		return 10 * time.Minute // Default timeout
	}
	return d
}

// Description returns the scanner description.
func (s *ScannerDefinition) Description() string { return s.DescriptionValue }

// ScannersConfig holds the scanner definitions from scanners.yml.
type ScannersConfig struct {
	// Scanners maps scanner ID to definition.
	Scanners map[string]*ScannerDefinition `yaml:"scanners"`
}

// PoliciesConfig holds the scanner policies from policies.yml.
type PoliciesConfig struct {
	// ComponentScanners maps component type to scanner IDs.
	ComponentScanners map[string][]string `yaml:"component_scanners"`

	// SkipModules lists modules to skip during scanning.
	SkipModules []string `yaml:"skip_modules"`

	// Default lists scanner IDs when no specific policy exists.
	Default []string `yaml:"default"`
}

// ScannerCategory constants for well-known scanner categories.
const (
	CategorySBOM       = "sbom"       // Software Bill of Materials
	CategoryVuln       = "vuln"       // Vulnerability scanning
	CategorySecrets    = "secrets"    // Secrets detection
	CategoryIaC        = "iac"        // Infrastructure as Code
	CategoryCompliance = "compliance" // Compliance checking
	CategorySAST       = "sast"       // Static Application Security Testing
	CategoryDAST       = "dast"       // Dynamic Application Security Testing
)
