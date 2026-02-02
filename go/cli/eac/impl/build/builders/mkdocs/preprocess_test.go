package mkdocs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPreprocessHandler_Name(t *testing.T) {
	h := NewPreprocessHandler("/workspace")
	if got := h.Name(); got != "mkdocs-preprocess" {
		t.Errorf("Name() = %q, want %q", got, "mkdocs-preprocess")
	}
}

func TestPreprocessHandler_Requirements(t *testing.T) {
	h := NewPreprocessHandler("/workspace")
	reqs := h.Requirements()
	if reqs != nil && len(reqs) != 0 {
		t.Errorf("Requirements() = %v, want nil or empty (no Docker required)", reqs)
	}
}

func TestPreprocessHandler_IsContainer(t *testing.T) {
	h := NewPreprocessHandler("/workspace")
	if h.IsContainer() {
		t.Error("IsContainer() = true, want false (preprocessing runs in pure Go)")
	}
}

func TestPreprocessHandler_IsHostInstalled(t *testing.T) {
	h := NewPreprocessHandler("/workspace")
	if !h.IsHostInstalled() {
		t.Error("IsHostInstalled() = false, want true")
	}
}

func TestPreprocessHandler_ListArtifacts(t *testing.T) {
	h := NewPreprocessHandler("/workspace")
	artifacts := h.ListArtifacts(nil, "/workspace")

	expected := []string{"staged/", "mkdocs.yml", ".manifest.json"}
	if len(artifacts) != len(expected) {
		t.Fatalf("ListArtifacts() returned %d artifacts, want %d", len(artifacts), len(expected))
	}

	for i, a := range expected {
		if artifacts[i] != a {
			t.Errorf("ListArtifacts()[%d] = %q, want %q", i, artifacts[i], a)
		}
	}
}

func TestComputeStagingHash_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mkdocs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hash, err := computeStagingHash(tmpDir)
	if err != nil {
		t.Fatalf("computeStagingHash() error: %v", err)
	}

	// Empty directory should produce a consistent hash
	if hash == "" {
		t.Error("computeStagingHash() returned empty hash for empty directory")
	}
}

func TestComputeStagingHash_Deterministic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mkdocs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some files
	if err := os.WriteFile(filepath.Join(tmpDir, "index.md"), []byte("# Test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "page.md"), []byte("content"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Compute hash twice
	hash1, err := computeStagingHash(tmpDir)
	if err != nil {
		t.Fatalf("computeStagingHash() error: %v", err)
	}

	hash2, err := computeStagingHash(tmpDir)
	if err != nil {
		t.Fatalf("computeStagingHash() error: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("computeStagingHash() not deterministic: %s != %s", hash1, hash2)
	}
}

func TestComputeStagingHash_ChangesOnContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mkdocs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create initial file
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("original"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	hash1, err := computeStagingHash(tmpDir)
	if err != nil {
		t.Fatalf("computeStagingHash() error: %v", err)
	}

	// Modify file
	if err := os.WriteFile(testFile, []byte("modified"), 0o644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	hash2, err := computeStagingHash(tmpDir)
	if err != nil {
		t.Fatalf("computeStagingHash() error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("computeStagingHash() should produce different hash when content changes")
	}
}

func TestLogln(t *testing.T) {
	var buf bytes.Buffer
	logln(&buf, "Hello %s", "World")

	expected := "Hello World\n"
	if got := buf.String(); got != expected {
		t.Errorf("logln() wrote %q, want %q", got, expected)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 1, -1},
		{0, 0, 0},
	}

	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
