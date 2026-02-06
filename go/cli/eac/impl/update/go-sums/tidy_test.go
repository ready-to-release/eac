package gosums

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGoWork(t *testing.T) {
	dir := t.TempDir()

	content := `go 1.24.4

use (
	./contracts
	./contracts/core/0.1.0/interfaces
	./go/cli/eac
	./go/core
)
`
	if err := os.WriteFile(filepath.Join(dir, "go.work"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	modules, err := parseGoWork(dir)
	if err != nil {
		t.Fatalf("parseGoWork failed: %v", err)
	}

	expected := []string{
		filepath.Join(dir, "contracts"),
		filepath.Join(dir, "contracts", "core", "0.1.0", "interfaces"),
		filepath.Join(dir, "go", "cli", "eac"),
		filepath.Join(dir, "go", "core"),
	}

	if len(modules) != len(expected) {
		t.Fatalf("expected %d modules, got %d: %v", len(expected), len(modules), modules)
	}

	for i, got := range modules {
		if got != expected[i] {
			t.Errorf("module[%d]: expected %q, got %q", i, expected[i], got)
		}
	}
}

func TestParseGoWork_Empty(t *testing.T) {
	dir := t.TempDir()

	content := `go 1.24.4
`
	if err := os.WriteFile(filepath.Join(dir, "go.work"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	modules, err := parseGoWork(dir)
	if err != nil {
		t.Fatalf("parseGoWork failed: %v", err)
	}

	if len(modules) != 0 {
		t.Fatalf("expected 0 modules, got %d: %v", len(modules), modules)
	}
}

func TestParseGoWork_MissingFile(t *testing.T) {
	dir := t.TempDir()

	_, err := parseGoWork(dir)
	if err == nil {
		t.Fatal("expected error for missing go.work, got nil")
	}
}
