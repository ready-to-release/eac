package commitmessage

import (
	"path/filepath"
	"testing"
)

func TestContractImplementation(t *testing.T) {
	// Path relative to this test file
	contractPath := filepath.Join("..", "..", "..", "contracts", "commit-message", "0.1.0", "structure.yml")

	errors := VerifyContractImplementation(contractPath)

	if len(errors) > 0 {
		t.Errorf("Contract implementation verification failed with %d error(s):", len(errors))
		for _, err := range errors {
			t.Errorf("  - [%s] %s", err.Code, err.Message)
		}
	}
}

func TestLoadContract(t *testing.T) {
	contractPath := filepath.Join("..", "..", "..", "contracts", "commit-message", "0.1.0", "structure.yml")

	contract, err := LoadContract(contractPath)
	if err != nil {
		t.Fatalf("Failed to load contract: %v", err)
	}

	// Verify basic structure
	if contract.Version != "0.1.0" {
		t.Errorf("Expected version 0.1.0, got %s", contract.Version)
	}

	if len(contract.Structure) != 4 {
		t.Errorf("Expected 4 structure sections, got %d", len(contract.Structure))
	}

	if len(contract.SemanticTypes) != 8 {
		t.Errorf("Expected 8 semantic types, got %d", len(contract.SemanticTypes))
	}

	if contract.SubjectLineFormat != "<module>: <type>: <description>" {
		t.Errorf("Unexpected subject line format: %s", contract.SubjectLineFormat)
	}

	// Verify constraints
	expectedConstraints := []string{
		"max_line_length",
		"no_trailing_periods",
		"code_blocks_closed",
		"module_header_no_colons",
	}

	for _, constraint := range expectedConstraints {
		if _, exists := contract.Constraints[constraint]; !exists {
			t.Errorf("Missing constraint: %s", constraint)
		}
	}

	// Verify markdown rules exist
	if len(contract.MarkdownRules) == 0 {
		t.Error("Expected markdown_rules to be defined")
	}
}
