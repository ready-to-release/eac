// Testing utilities for security scanners
// This file provides test helpers that are accessible from outside the package
package scan

import (
	"github.com/ready-to-release/eac/go/core/evidence"
)

// Evidence represents a security scan evidence file.
type Evidence = evidence.File

// ReadEvidence reads and parses an evidence file.
func ReadEvidence(filepath string) (*Evidence, error) {
	return evidence.ReadEvidence(filepath)
}

// SetupMocks configures mock outputs for all security tools.
// Security scanners now use the tool executor which mocks at the Docker client level.
// Set R2R_MOCK_DOCKER=true environment variable to enable mocking.
func SetupMocks() {
	// No-op: mocking is now handled at the Docker client level in adapters/docker.
	// Set R2R_MOCK_DOCKER=true to enable mock Docker client.
}

// ResetMocks clears all mock outputs.
// Security scanners now use the tool executor which mocks at the Docker client level.
func ResetMocks() {
	// No-op: mocking is now handled at the Docker client level in adapters/docker.
}
