// Package interfaces defines the security scanning contract interfaces.
// These interfaces define how security scanners are configured and accessed.
package interfaces

import "time"

// SecurityConfigPort provides access to security scanner configuration.
// Implementations load scanner definitions and policies from configuration files.
type SecurityConfigPort interface {
	// GetScanner returns a scanner definition by ID.
	// Returns false if the scanner is not found.
	GetScanner(id string) (ScannerPort, bool)

	// ListScanners returns all available scanner IDs.
	ListScanners() []string

	// GetDefaultScanners returns the default scanner IDs for a component type.
	// Falls back to "default" policy if no specific policy exists.
	GetDefaultScanners(componentType string) []string

	// ShouldSkipModule returns true if the module should be skipped during scanning.
	ShouldSkipModule(moniker string) bool
}

// ScannerPort defines a security scanner's configuration.
// Scanners are typically Docker containers that perform security analysis.
type ScannerPort interface {
	// ID returns the unique scanner identifier (e.g., "trivy-sbom").
	ID() string

	// Category returns the scanner category (e.g., "sbom", "vuln", "sast").
	Category() string

	// Image returns the Docker image name (e.g., "ghcr.io/aquasecurity/trivy").
	Image() string

	// Tag returns the Docker image tag (e.g., "0.68.2").
	Tag() string

	// FullImage returns the complete image reference (image:tag).
	FullImage() string

	// Command returns the scanner command arguments.
	Command() []string

	// Timeout returns the maximum execution duration.
	Timeout() time.Duration

	// Description returns a human-readable description.
	Description() string
}
