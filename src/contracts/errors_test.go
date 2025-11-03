package contracts

import (
	"errors"
	"strings"
	"testing"
)

func TestContractError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ContractError
		expected string
	}{
		{
			name: "error with path",
			err: &ContractError{
				Op:      "load",
				Path:    "/test/file.yml",
				Message: "file not found",
			},
			expected: "contract load failed for /test/file.yml: file not found",
		},
		{
			name: "error without path",
			err: &ContractError{
				Op:      "parse",
				Message: "invalid YAML",
			},
			expected: "contract parse failed: invalid YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestContractError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &ContractError{
		Op:      "test",
		Path:    "/test",
		Err:     underlying,
		Message: "test message",
	}

	unwrapped := err.Unwrap()
	if unwrapped != underlying {
		t.Errorf("Unwrap() returned wrong error")
	}
}

func TestNewContractError(t *testing.T) {
	underlying := errors.New("test error")
	err := NewContractError("load", "/test/path", underlying, "test message")

	if err.Op != "load" {
		t.Errorf("Expected Op 'load', got '%s'", err.Op)
	}

	if err.Path != "/test/path" {
		t.Errorf("Expected Path '/test/path', got '%s'", err.Path)
	}

	if err.Err != underlying {
		t.Error("Expected Err to be underlying error")
	}

	if err.Message != "test message" {
		t.Errorf("Expected Message 'test message', got '%s'", err.Message)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "not found error",
			err: &ContractError{
				Op:      "load",
				Message: "contract not found",
			},
			expected: true,
		},
		{
			name: "different error",
			err: &ContractError{
				Op:      "load",
				Message: "some other error",
			},
			expected: false,
		},
		{
			name: "parse error",
			err: &ContractError{
				Op:      "parse",
				Message: "contract not found",
			},
			expected: false,
		},
		{
			name:     "non-contract error",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.err)
			if got != tt.expected {
				t.Errorf("IsNotFound() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestContractError_ErrorMessage_Contains_AllFields(t *testing.T) {
	err := &ContractError{
		Op:      "load",
		Path:    "/test/contract.yml",
		Err:     errors.New("underlying"),
		Message: "failed to load",
	}

	errorMsg := err.Error()

	// Check that error message contains key information
	if !strings.Contains(errorMsg, "load") {
		t.Error("Error message should contain operation")
	}

	if !strings.Contains(errorMsg, "/test/contract.yml") {
		t.Error("Error message should contain path")
	}

	if !strings.Contains(errorMsg, "failed to load") {
		t.Error("Error message should contain message")
	}
}
