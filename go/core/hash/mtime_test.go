package hash

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Tests for collectMtimes

func TestCollectMtimes_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()
	result := collectMtimes(tmpDir, nil)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestCollectMtimes_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result := collectMtimes(tmpDir, []string{"test.txt"})
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if _, ok := result["test.txt"]; !ok {
		t.Error("expected entry for test.txt")
	}
}

func TestCollectMtimes_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"a.go", "b.go", "c.go"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	result := collectMtimes(tmpDir, []string{"a.go", "b.go", "c.go"})
	if len(result) != 3 {
		t.Errorf("expected 3 entries, got %d", len(result))
	}
}

func TestCollectMtimes_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "exists.txt"), []byte("here"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result := collectMtimes(tmpDir, []string{"exists.txt", "missing.txt"})
	if len(result) != 1 {
		t.Errorf("expected 1 entry (skipping missing), got %d", len(result))
	}
	if _, ok := result["exists.txt"]; !ok {
		t.Error("expected entry for exists.txt")
	}
}

// Tests for mtimesUnchanged

func TestMtimesUnchanged_AllMatch(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Collect current mtimes
	mtimes := collectMtimes(tmpDir, []string{"file.txt"})

	// Should match immediately
	if !mtimesUnchanged(tmpDir, mtimes) {
		t.Error("expected mtimes to be unchanged")
	}
}

func TestMtimesUnchanged_MtimeChanged(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Collect current mtimes
	mtimes := collectMtimes(tmpDir, []string{"file.txt"})

	// Modify the file with a different mtime
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(filePath, future, future); err != nil {
		t.Fatalf("failed to change mtime: %v", err)
	}

	if mtimesUnchanged(tmpDir, mtimes) {
		t.Error("expected mtimes to be detected as changed")
	}
}

func TestMtimesUnchanged_FileMissing(t *testing.T) {
	tmpDir := t.TempDir()

	// Cache with a file that doesn't exist
	cached := map[string]int64{"missing.txt": 12345}

	if mtimesUnchanged(tmpDir, cached) {
		t.Error("expected false when file is missing")
	}
}

func TestMtimesUnchanged_EmptyCache(t *testing.T) {
	tmpDir := t.TempDir()

	// Empty cache should return true (vacuously true)
	if !mtimesUnchanged(tmpDir, map[string]int64{}) {
		t.Error("expected true for empty cache")
	}
}
