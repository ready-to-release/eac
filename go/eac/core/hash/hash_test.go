// Package hash provides deterministic file content hashing.
package hash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashFiles_EmptyList(t *testing.T) {
	result, err := Files("/any/path", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string for empty file list, got %q", result)
	}
}

func TestHashFiles_SingleFile(t *testing.T) {
	// Setup: create temp directory with a file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Act
	result, err := Files(tmpDir, []string{"test.txt"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty hash")
	}
	// SHA256 hex is 64 chars
	if len(result) != 64 {
		t.Errorf("expected 64-char SHA256 hash, got %d chars", len(result))
	}
}

func TestHashFiles_MultipleFiles_Deterministic(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("content A"), 0644); err != nil {
		t.Fatalf("failed to create a.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("content B"), 0644); err != nil {
		t.Fatalf("failed to create b.txt: %v", err)
	}

	// Act: hash twice with same files
	hash1, err := Files(tmpDir, []string{"a.txt", "b.txt"})
	if err != nil {
		t.Fatalf("first hash failed: %v", err)
	}

	hash2, err := Files(tmpDir, []string{"a.txt", "b.txt"})
	if err != nil {
		t.Fatalf("second hash failed: %v", err)
	}

	// Assert: same result
	if hash1 != hash2 {
		t.Errorf("hash not deterministic: %s != %s", hash1, hash2)
	}
}

func TestHashFiles_OrderIndependent(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("content A"), 0644); err != nil {
		t.Fatalf("failed to create a.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("content B"), 0644); err != nil {
		t.Fatalf("failed to create b.txt: %v", err)
	}

	// Act: hash with different input order
	hashAB, err := Files(tmpDir, []string{"a.txt", "b.txt"})
	if err != nil {
		t.Fatalf("AB hash failed: %v", err)
	}

	hashBA, err := Files(tmpDir, []string{"b.txt", "a.txt"})
	if err != nil {
		t.Fatalf("BA hash failed: %v", err)
	}

	// Assert: order doesn't matter (files are sorted internally)
	if hashAB != hashBA {
		t.Errorf("hash depends on input order: AB=%s, BA=%s", hashAB, hashBA)
	}
}

func TestHashFiles_IncludesFilename(t *testing.T) {
	// Setup: two files with identical content but different names
	tmpDir := t.TempDir()
	content := []byte("same content")
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), content, 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.txt"), content, 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	// Act
	hash1, err := Files(tmpDir, []string{"file1.txt"})
	if err != nil {
		t.Fatalf("file1 hash failed: %v", err)
	}

	hash2, err := Files(tmpDir, []string{"file2.txt"})
	if err != nil {
		t.Fatalf("file2 hash failed: %v", err)
	}

	// Assert: different hashes because filenames are included
	if hash1 == hash2 {
		t.Error("expected different hashes for files with different names")
	}
}

func TestHashFiles_DifferentContent(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("version 1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	hash1, err := Files(tmpDir, []string{"test.txt"})
	if err != nil {
		t.Fatalf("first hash failed: %v", err)
	}

	// Modify content
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("version 2"), 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	hash2, err := Files(tmpDir, []string{"test.txt"})
	if err != nil {
		t.Fatalf("second hash failed: %v", err)
	}

	// Assert: different content = different hash
	if hash1 == hash2 {
		t.Error("expected different hashes for different content")
	}
}

func TestHashFiles_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Files(tmpDir, []string{"nonexistent.txt"})

	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestHashFiles_Subdirectory(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("nested"), 0644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	// Act: use forward slash path (normalized)
	result, err := Files(tmpDir, []string{"sub/file.txt"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty hash for nested file")
	}
}

// Tests for UncommittedState

func TestUncommittedState_EmptyList(t *testing.T) {
	result := UncommittedState("/any/path", nil)
	if result != "" {
		t.Errorf("expected empty string for empty file list, got %q", result)
	}
}

func TestUncommittedState_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "modified.txt"), []byte("uncommitted"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result := UncommittedState(tmpDir, []string{"modified.txt"})

	// Short hash is 16 chars
	if len(result) != 16 {
		t.Errorf("expected 16-char short hash, got %d chars: %q", len(result), result)
	}
}

func TestUncommittedState_DeletedFile(t *testing.T) {
	tmpDir := t.TempDir()

	// File doesn't exist - simulates deleted file
	result := UncommittedState(tmpDir, []string{"deleted.txt"})

	// Should still produce a hash (marking it as deleted)
	if result == "" {
		t.Error("expected hash even for deleted file")
	}
	if len(result) != 16 {
		t.Errorf("expected 16-char hash, got %d chars", len(result))
	}
}

func TestUncommittedState_Deterministic(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	hash1 := UncommittedState(tmpDir, []string{"file.txt"})
	hash2 := UncommittedState(tmpDir, []string{"file.txt"})

	if hash1 != hash2 {
		t.Errorf("hash not deterministic: %s != %s", hash1, hash2)
	}
}

func TestUncommittedState_OrderIndependent(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("A"), 0644); err != nil {
		t.Fatalf("failed to create a.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("B"), 0644); err != nil {
		t.Fatalf("failed to create b.txt: %v", err)
	}

	hashAB := UncommittedState(tmpDir, []string{"a.txt", "b.txt"})
	hashBA := UncommittedState(tmpDir, []string{"b.txt", "a.txt"})

	if hashAB != hashBA {
		t.Errorf("hash depends on order: AB=%s, BA=%s", hashAB, hashBA)
	}
}

// Tests for ExpandGlobPatterns

func TestExpandGlobPatterns_EmptyPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := ExpandGlobPatterns(tmpDir, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d files", len(result))
	}
}

func TestExpandGlobPatterns_SimplePattern(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create file1.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create file2.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# readme"), 0644); err != nil {
		t.Fatalf("failed to create readme.md: %v", err)
	}

	result, err := ExpandGlobPatterns(tmpDir, []string{"*.go"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 .go files, got %d: %v", len(result), result)
	}
}

func TestExpandGlobPatterns_Subdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create src/: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create src/main.go: %v", err)
	}

	result, err := ExpandGlobPatterns(tmpDir, []string{"src/*.go"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(result), result)
	}
	// Result should use forward slashes
	if result[0] != "src/main.go" {
		t.Errorf("expected 'src/main.go', got %q", result[0])
	}
}

func TestExpandGlobPatterns_ExcludesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a directory that matches the pattern
	if err := os.MkdirAll(filepath.Join(tmpDir, "test.go"), 0755); err != nil {
		t.Fatalf("failed to create test.go directory: %v", err)
	}
	// Create an actual file
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	result, err := ExpandGlobPatterns(tmpDir, []string{"*.go"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should only include the file, not the directory
	if len(result) != 1 {
		t.Errorf("expected 1 file (excluding directory), got %d: %v", len(result), result)
	}
}

func TestExpandGlobPatterns_Deduplicates(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	// Multiple patterns that match the same file
	result, err := ExpandGlobPatterns(tmpDir, []string{"*.go", "main.*", "main.go"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected deduplicated result of 1 file, got %d: %v", len(result), result)
	}
}

func TestExpandGlobPatterns_SortedOutput(t *testing.T) {
	tmpDir := t.TempDir()
	// Create files in non-alphabetical order
	for _, name := range []string{"z.go", "a.go", "m.go"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("package main"), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	result, err := ExpandGlobPatterns(tmpDir, []string{"*.go"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be sorted
	expected := []string{"a.go", "m.go", "z.go"}
	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("expected %s at position %d, got %s", exp, i, result[i])
		}
	}
}

func TestExpandGlobPatterns_InvalidPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Invalid glob pattern
	_, err := ExpandGlobPatterns(tmpDir, []string{"[invalid"})

	if err == nil {
		t.Error("expected error for invalid glob pattern")
	}
}

func TestExpandGlobPatterns_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a different type of file
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# readme"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result, err := ExpandGlobPatterns(tmpDir, []string{"*.go"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty result is fine, not an error
	if len(result) != 0 {
		t.Errorf("expected empty result for no matches, got %d: %v", len(result), result)
	}
}

func TestExpandGlobPatterns_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	// Absolute pattern
	absPattern := filepath.Join(tmpDir, "*.go")
	result, err := ExpandGlobPatterns(tmpDir, []string{absPattern})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(result), result)
	}
	// Result should still be relative
	if result[0] != "main.go" {
		t.Errorf("expected relative path 'main.go', got %q", result[0])
	}
}
