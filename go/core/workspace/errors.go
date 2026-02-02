package workspace

import (
	"errors"
	"fmt"
)

// Sentinel errors for workspace detection.
var (
	// ErrNotFound is returned when no valid workspace can be detected.
	ErrNotFound = errors.New("workspace not found")

	// ErrInvalidPath is returned when a path exists but is not a valid workspace.
	ErrInvalidPath = errors.New("invalid workspace path")
)

// DetectionError provides detailed context about workspace detection failures.
type DetectionError struct {
	Op      string // "detect", "validate", "resolve"
	Path    string // Path that was checked (if applicable)
	Source  string // "env:R2R_REPO_ROOT", "git", "docker", etc.
	Message string
	Err     error
}

func (e *DetectionError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("workspace %s failed at %q (%s): %s", e.Op, e.Path, e.Source, e.Message)
	}
	return fmt.Sprintf("workspace %s failed (%s): %s", e.Op, e.Source, e.Message)
}

func (e *DetectionError) Unwrap() error { return e.Err }
