// Package oscal provides OSCAL document writing with file locking for concurrent access.
package oscal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

const (
	// LockTimeout is the maximum time to wait for a file lock.
	LockTimeout = 30 * time.Second
)

// WriteProfile writes an OSCAL profile to a file.
func WriteProfile(path string, profile *Profile) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Update last-modified timestamp
	profile.Profile.Metadata.LastModified = time.Now().UTC().Format(time.RFC3339)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	return nil
}

// WriteAssessmentResults writes OSCAL assessment-results with file locking.
// Uses exclusive file locking to prevent concurrent write corruption.
func WriteAssessmentResults(path string, ar *AssessmentResults) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Acquire exclusive lock
	lockPath := path + ".lock"
	lock := flock.New(lockPath)

	// Try to acquire lock with timeout
	ctx, cancel := context.WithTimeout(context.Background(), LockTimeout)
	defer cancel()

	locked, err := lock.TryLockContext(ctx, 100*time.Millisecond)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("assessment-results.json is locked by another process (timeout after %s)", LockTimeout)
	}
	defer func() {
		lock.Unlock()
		os.Remove(lockPath) // Clean up lock file
	}()

	// Update last-modified timestamp
	ar.AssessmentResults.Metadata.LastModified = time.Now().UTC().Format(time.RFC3339)

	// Write to temporary file first (atomic write)
	tmpPath := path + ".tmp"
	if err := writeAssessmentResultsJSON(tmpPath, ar); err != nil {
		os.Remove(tmpPath) // Clean up on error
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // Clean up on error
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// writeAssessmentResultsJSON marshals and writes OSCAL JSON.
func writeAssessmentResultsJSON(path string, ar *AssessmentResults) error {
	data, err := json.MarshalIndent(ar, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal assessment-results: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// WriteAssessmentResultsWithoutLock writes without file locking.
// Use only when you know concurrent access is not possible.
func WriteAssessmentResultsWithoutLock(path string, ar *AssessmentResults) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Update last-modified timestamp
	ar.AssessmentResults.Metadata.LastModified = time.Now().UTC().Format(time.RFC3339)

	return writeAssessmentResultsJSON(path, ar)
}

// CleanStaleLocksIfNeeded removes lock files older than 5 minutes.
// Call this before starting operations that may have been interrupted.
func CleanStaleLocksIfNeeded(lockPath string) error {
	info, err := os.Stat(lockPath)
	if os.IsNotExist(err) {
		return nil // No lock file
	}
	if err != nil {
		return err
	}

	if time.Since(info.ModTime()) > 5*time.Minute {
		return os.Remove(lockPath)
	}

	return nil
}

// EnsureRiskOutputDir ensures the risk output directory exists for a module.
func EnsureRiskOutputDir(workspaceRoot, moduleName string) (string, error) {
	outputDir := filepath.Join(workspaceRoot, "out", "risk", moduleName)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create risk output directory: %w", err)
	}
	return outputDir, nil
}

// EnsureProfilesDir ensures the profiles directory exists.
func EnsureProfilesDir(workspaceRoot string) (string, error) {
	profileDir := filepath.Join(workspaceRoot, "specs", "risk-controls")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create profiles directory: %w", err)
	}
	return profileDir, nil
}

// WriteRiskReport writes a human-readable risk report in markdown format.
func WriteRiskReport(path string, content string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write risk report: %w", err)
	}

	return nil
}

// WriteHistoricalSnapshot writes a timestamped snapshot of assessment results.
func WriteHistoricalSnapshot(workspaceRoot, moduleName string, ar *AssessmentResults) (string, error) {
	outputDir, err := EnsureRiskOutputDir(workspaceRoot, moduleName)
	if err != nil {
		return "", err
	}

	// Generate timestamp filename
	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05Z")
	snapshotPath := filepath.Join(outputDir, timestamp+".json")

	// Write snapshot (no locking needed for unique timestamped files)
	if err := WriteAssessmentResultsWithoutLock(snapshotPath, ar); err != nil {
		return "", err
	}

	return snapshotPath, nil
}
