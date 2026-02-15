//go:build L1 && ov
// +build L1,ov

package commitmessage

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/repository"
)

func TestLoadContractFromConfig(t *testing.T) {
	// Get repository root dynamically
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		t.Fatalf("Failed to find repository root: %v", err)
	}

	contract, err := LoadContractFromConfig(repoRoot)
	if err != nil {
		t.Fatalf("Failed to load contract from config: %v", err)
	}

	// Verify basic structure
	if contract.Version != "0.1.0" {
		t.Errorf("Expected version 0.1.0, got %s", contract.Version)
	}

	if len(contract.SemanticTypes) != 8 {
		t.Errorf("Expected 8 semantic types, got %d", len(contract.SemanticTypes))
	}

	if contract.SubjectLineFormat != "<module>: <type>: <description>" {
		t.Errorf("Unexpected subject line format: %s", contract.SubjectLineFormat)
	}

	// Verify constraints exist
	if contract.Constraints == nil {
		t.Error("Expected constraints to be defined")
	}

	if _, exists := contract.Constraints["max_line_length"]; !exists {
		t.Error("Expected max_line_length constraint")
	}
}
