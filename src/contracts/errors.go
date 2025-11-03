package contracts

import "fmt"

// ContractError represents a contract-related error
type ContractError struct {
	Op      string // Operation that failed (e.g., "load", "parse", "validate")
	Path    string // File path related to the error
	Err     error  // Underlying error
	Message string // Additional context
}

func (e *ContractError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("contract %s failed for %s: %s", e.Op, e.Path, e.Message)
	}
	return fmt.Sprintf("contract %s failed: %s", e.Op, e.Message)
}

func (e *ContractError) Unwrap() error {
	return e.Err
}

// NewContractError creates a new ContractError
func NewContractError(op, path string, err error, message string) *ContractError {
	return &ContractError{
		Op:      op,
		Path:    path,
		Err:     err,
		Message: message,
	}
}

// IsNotFound returns true if the error is a "not found" error
func IsNotFound(err error) bool {
	if ce, ok := err.(*ContractError); ok {
		return ce.Op == "load" && ce.Message == "contract not found"
	}
	return false
}
