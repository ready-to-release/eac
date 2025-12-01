// Testing utilities for security scanners
// This file provides test helpers that are accessible from outside the package
package security

import (
	"github.com/ready-to-release/eac/src/commands/impl/security/internal"
)

// Evidence represents a security scan evidence file
type Evidence = internal.EvidenceFile

// ReadEvidence reads and parses an evidence file
func ReadEvidence(filepath string) (*Evidence, error) {
	return internal.ReadEvidence(filepath)
}

// SetupMocks configures mock outputs for all security tools
func SetupMocks() {
	// Set mock output for all Trivy-based scanners
	mockTrivyData := map[string]interface{}{
		"SchemaVersion": 2,
		"ArtifactName":  "test-module",
		"ArtifactType":  "filesystem",
		"Results": []map[string]interface{}{
			{
				"Target": "test-module",
				"Type":   "test",
			},
		},
	}
	internal.SetMockTrivyOutput(mockTrivyData)

	// Set mock output for Semgrep
	mockSemgrepData := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"check_id": "test-rule",
				"path":     "test.go",
			},
		},
	}
	internal.SetMockSemgrepOutput(mockSemgrepData)

	// Set mock output for ZAP
	mockZAPData := map[string]interface{}{
		"site": []map[string]interface{}{
			{
				"@name":  "http://localhost:8080",
				"alerts": []string{},
			},
		},
	}
	internal.SetMockZAPOutput(mockZAPData)
}

// ResetMocks clears all mock outputs
func ResetMocks() {
	internal.ResetMockTrivyOutput()
	internal.ResetMockSemgrepOutput()
	internal.ResetMockZAPOutput()
}
