package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveAllWithRetry_RemovesDirectory(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")

	// Create a directory with files
	if err := os.MkdirAll(filepath.Join(target, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "file.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "sub", "nested.txt"), []byte("nested"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveAllWithRetry(target); err != nil {
		t.Fatalf("RemoveAllWithRetry failed: %v", err)
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("expected target to be removed, stat returned: %v", err)
	}
}

func TestRemoveAllWithRetry_NonexistentPath(t *testing.T) {
	err := RemoveAllWithRetry(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("expected nil error for nonexistent path, got: %v", err)
	}
}

func TestRemoveAllWithRetry_RemovesSingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")

	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveAllWithRetry(path); err != nil {
		t.Fatalf("RemoveAllWithRetry failed: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file to be removed")
	}
}
