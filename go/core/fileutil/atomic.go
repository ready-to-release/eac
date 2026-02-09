package fileutil

import (
	"encoding/json"
	"fmt"
	"os"
)

// AtomicWrite writes data to a file atomically using temp file + rename.
// Pattern extracted from oscal/writer.go and books/cache.go.
//
// This ensures that:
// - Either the entire write succeeds or nothing is written (no partial files)
// - Concurrent readers see old data or new data, never partial data
// - Failures leave the original file intact
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	tmpPath := path + ".tmp"

	// Write to temporary file
	if err := os.WriteFile(tmpPath, data, perm); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	// Atomic rename (OS-level atomicity)
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // Clean up on error
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// AtomicWriteJSON writes JSON data atomically with pretty formatting.
func AtomicWriteJSON(path string, v interface{}, perm os.FileMode) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	return AtomicWrite(path, data, perm)
}
