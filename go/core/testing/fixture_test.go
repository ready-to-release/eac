//go:build L0

package testing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFixturePool(t *testing.T) {
	pool := NewFixturePool()
	if pool == nil {
		t.Fatal("NewFixturePool returned nil")
	}

	stats := pool.Stats()
	if stats.TemplateCount != 0 {
		t.Errorf("new pool should have 0 templates, got %d", stats.TemplateCount)
	}
	if stats.TempDirCount != 0 {
		t.Errorf("new pool should have 0 temp dirs, got %d", stats.TempDirCount)
	}
}

func TestFixturePool_GetTemplate_MissingReturnsNil(t *testing.T) {
	pool := NewFixturePool()
	template := pool.GetTemplate("/nonexistent")
	if template != nil {
		t.Error("GetTemplate for missing key should return nil")
	}
}

func TestFixturePool_TrackTempDir(t *testing.T) {
	pool := NewFixturePool()
	defer pool.Cleanup()

	tmpDir, err := os.MkdirTemp("", "fixture-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	pool.TrackTempDir(tmpDir)

	stats := pool.Stats()
	if stats.TempDirCount != 1 {
		t.Errorf("expected 1 tracked dir, got %d", stats.TempDirCount)
	}

	// Cleanup should remove it
	pool.Cleanup()
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Error("tracked temp dir should be removed after Cleanup")
	}
}

func TestFixturePool_CleanupIdempotent(t *testing.T) {
	pool := NewFixturePool()

	tmpDir, err := os.MkdirTemp("", "fixture-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	pool.TrackTempDir(tmpDir)

	// First cleanup
	pool.Cleanup()

	// Second cleanup should be safe
	pool.Cleanup()

	stats := pool.Stats()
	if stats.TempDirCount != 0 {
		t.Errorf("after cleanup, expected 0 temp dirs, got %d", stats.TempDirCount)
	}
}

func TestFixturePool_NewIsolationFromTemplate_NilTemplate(t *testing.T) {
	pool := NewFixturePool()
	defer pool.Cleanup()

	_, err := pool.NewIsolationFromTemplate(nil)
	if err == nil {
		t.Error("expected error for nil template")
	}
}

func TestFixturePool_NewIsolationFromTemplate_CopiesDir(t *testing.T) {
	pool := NewFixturePool()
	defer pool.Cleanup()

	// Create a fake template directory with a marker file
	templateDir, err := os.MkdirTemp("", "fixture-template-*")
	if err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}
	defer os.RemoveAll(templateDir)

	markerPath := filepath.Join(templateDir, "marker.txt")
	if err := os.WriteFile(markerPath, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to write marker: %v", err)
	}

	// Also need a .git directory for the isolation to be valid
	gitDir := filepath.Join(templateDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}

	template := &FixtureTemplate{
		TemplateDir:      templateDir,
		OriginalRepoRoot: "/original",
	}

	isolation, err := pool.NewIsolationFromTemplate(template)
	if err != nil {
		t.Fatalf("NewIsolationFromTemplate failed: %v", err)
	}
	defer isolation.Cleanup()

	// Verify marker file was copied
	copiedMarker := filepath.Join(isolation.IsolatedDir(), "marker.txt")
	content, err := os.ReadFile(copiedMarker)
	if err != nil {
		t.Fatalf("failed to read copied marker: %v", err)
	}
	if string(content) != "hello" {
		t.Errorf("expected 'hello', got %q", string(content))
	}

	// Stats should reflect the new temp dir
	stats := pool.Stats()
	if stats.TempDirCount != 1 {
		t.Errorf("expected 1 temp dir, got %d", stats.TempDirCount)
	}
}

func TestEstimatePerformanceGain(t *testing.T) {
	tests := []struct {
		name           string
		scenarios      int
		wantSavingsPos bool
	}{
		{"single scenario", 1, false}, // Template creation dominates for 1 scenario
		{"ten scenarios", 10, true},
		{"hundred scenarios", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			without, with, savings := EstimatePerformanceGain(tt.scenarios)
			if without <= 0 {
				t.Errorf("withoutFixtures should be positive, got %f", without)
			}
			if with <= 0 {
				t.Errorf("withFixtures should be positive, got %f", with)
			}
			if tt.wantSavingsPos && savings <= 0 {
				t.Errorf("expected positive savings for %d scenarios, got %f", tt.scenarios, savings)
			}
		})
	}
}

func BenchmarkFixturePool_NewIsolationFromTemplate(b *testing.B) {
	pool := NewFixturePool()
	defer pool.Cleanup()

	// Create a template directory with some files to simulate a real template
	templateDir, err := os.MkdirTemp("", "bench-template-*")
	if err != nil {
		b.Fatalf("failed to create template dir: %v", err)
	}
	defer os.RemoveAll(templateDir)

	// Create .git dir and several files
	if err := os.MkdirAll(filepath.Join(templateDir, ".git", "objects"), 0755); err != nil {
		b.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(templateDir, "contracts"), 0755); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		name := filepath.Join(templateDir, "contracts", "file"+string(rune('0'+i))+".yml")
		if err := os.WriteFile(name, []byte("content"), 0644); err != nil {
			b.Fatal(err)
		}
	}

	template := &FixtureTemplate{
		TemplateDir:      templateDir,
		OriginalRepoRoot: "/bench",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iso, err := pool.NewIsolationFromTemplate(template)
		if err != nil {
			b.Fatal(err)
		}
		iso.Cleanup()
	}
}

func BenchmarkFixturePool_Stats(b *testing.B) {
	pool := NewFixturePool()
	defer pool.Cleanup()

	// Track some directories
	for i := 0; i < 50; i++ {
		pool.TrackTempDir("/fake/dir/" + string(rune('A'+i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Stats()
	}
}
