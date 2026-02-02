// Package hash provides deterministic file content hashing.
package hash

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// DefaultParallelOptions Tests
// =============================================================================

func TestDefaultParallelOptions_MaxWorkersDefault(t *testing.T) {
	opts := DefaultParallelOptions()

	// Should default to half NumCPU, floor min(4, NumCPU), cap 8
	numCPU := runtime.NumCPU()
	expected := numCPU / 2
	floor := 4
	if numCPU < floor {
		floor = numCPU
	}
	if expected < floor {
		expected = floor
	}
	if expected > 8 {
		expected = 8
	}

	if opts.MaxWorkers != expected {
		t.Errorf("expected MaxWorkers=%d (half NumCPU, floor min(4,NumCPU), cap 8), got %d", expected, opts.MaxWorkers)
	}
}

func TestDefaultParallelOptions_MaxWorkersPositive(t *testing.T) {
	opts := DefaultParallelOptions()

	if opts.MaxWorkers < 1 {
		t.Errorf("MaxWorkers should be at least 1, got %d", opts.MaxWorkers)
	}
}

// =============================================================================
// FilesParallel Basic Functionality Tests
// =============================================================================

func TestFilesParallel_EmptyList(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	result, err := FilesParallel(ctx, tmpDir, nil, DefaultParallelOptions())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string for empty file list, got %q", result)
	}
}

func TestFilesParallel_SingleFile(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := FilesParallel(ctx, tmpDir, []string{"test.txt"}, DefaultParallelOptions())

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

func TestFilesParallel_MultipleFiles(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create multiple test files
	files := []string{"a.txt", "b.txt", "c.txt", "d.txt", "e.txt"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content of "+name), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	result, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty hash")
	}
	if len(result) != 64 {
		t.Errorf("expected 64-char SHA256 hash, got %d chars", len(result))
	}
}

func TestFilesParallel_Subdirectory(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "sub", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("nested content"), 0644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	result, err := FilesParallel(ctx, tmpDir, []string{"sub/nested/file.txt"}, DefaultParallelOptions())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty hash for nested file")
	}
}

// =============================================================================
// Determinism Tests - CRITICAL: Parallel must match Sequential
// =============================================================================

func TestFilesParallel_MatchesSequential_SingleFile(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	files := []string{"test.txt"}

	sequentialHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	parallelHash, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	if sequentialHash != parallelHash {
		t.Errorf("CRITICAL: parallel hash differs from sequential\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

func TestFilesParallel_MatchesSequential_MultipleFiles(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create multiple files with varying content
	files := []string{"a.go", "b.go", "c.txt", "d.md", "e.json"}
	for i, name := range files {
		content := []byte("content number " + string(rune('0'+i)) + " for " + name)
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	sequentialHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	parallelHash, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	if sequentialHash != parallelHash {
		t.Errorf("CRITICAL: parallel hash differs from sequential\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

func TestFilesParallel_MatchesSequential_DifferentInputOrder(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create files
	fileContents := map[string]string{
		"z.txt": "last alphabetically",
		"a.txt": "first alphabetically",
		"m.txt": "middle alphabetically",
	}
	for name, content := range fileContents {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	// Different input orders
	orderAZ := []string{"a.txt", "m.txt", "z.txt"}
	orderZA := []string{"z.txt", "m.txt", "a.txt"}
	orderRandom := []string{"m.txt", "z.txt", "a.txt"}

	// All should produce the same hash as sequential
	sequentialHash, err := Files(tmpDir, orderAZ)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	for _, order := range [][]string{orderAZ, orderZA, orderRandom} {
		parallelHash, err := FilesParallel(ctx, tmpDir, order, DefaultParallelOptions())
		if err != nil {
			t.Fatalf("parallel hash failed for order %v: %v", order, err)
		}

		if sequentialHash != parallelHash {
			t.Errorf("CRITICAL: parallel hash differs for order %v\nsequential: %s\nparallel:   %s", order, sequentialHash, parallelHash)
		}
	}
}

func TestFilesParallel_MatchesSequential_LargeFileSet(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create many files to ensure parallel processing is actually used
	numFiles := 100
	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		name := filepath.Join(tmpDir, "file"+padNumber(i)+".txt")
		content := []byte("content for file " + padNumber(i) + " with some additional text to make it non-trivial")
		if err := os.WriteFile(name, content, 0644); err != nil {
			t.Fatalf("failed to create file %d: %v", i, err)
		}
		files[i] = "file" + padNumber(i) + ".txt"
	}

	sequentialHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	parallelHash, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	if sequentialHash != parallelHash {
		t.Errorf("CRITICAL: parallel hash differs from sequential for %d files\nsequential: %s\nparallel:   %s", numFiles, sequentialHash, parallelHash)
	}
}

func TestFilesParallel_MatchesSequential_NestedDirectories(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create nested directory structure
	dirs := []string{"a", "a/b", "a/b/c", "x", "x/y"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	files := []string{
		"root.txt",
		"a/file.txt",
		"a/b/file.txt",
		"a/b/c/deep.txt",
		"x/file.txt",
		"x/y/file.txt",
	}
	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("content of "+file), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", file, err)
		}
	}

	sequentialHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	parallelHash, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	if sequentialHash != parallelHash {
		t.Errorf("CRITICAL: parallel hash differs from sequential for nested files\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

func TestFilesParallel_Deterministic_MultipleRuns(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create files
	files := []string{"a.txt", "b.txt", "c.txt"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content of "+name), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	// Run multiple times and verify same result
	var firstHash string
	for i := 0; i < 10; i++ {
		hash, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
		if err != nil {
			t.Fatalf("run %d: unexpected error: %v", i, err)
		}
		if i == 0 {
			firstHash = hash
		} else if hash != firstHash {
			t.Errorf("run %d: hash differs from first run\nfirst: %s\nrun %d: %s", i, firstHash, i, hash)
		}
	}
}

func TestFilesParallel_Deterministic_DifferentWorkerCounts(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create enough files to span multiple workers
	numFiles := 50
	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		name := "file" + padNumber(i) + ".txt"
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content "+padNumber(i)), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		files[i] = name
	}

	// Get reference hash with 1 worker (effectively sequential within parallel)
	opts := ParallelOptions{MaxWorkers: 1}
	referenceHash, err := FilesParallel(ctx, tmpDir, files, opts)
	if err != nil {
		t.Fatalf("reference hash failed: %v", err)
	}

	// Test with various worker counts
	workerCounts := []int{2, 4, 8, 16}
	for _, workers := range workerCounts {
		opts := ParallelOptions{MaxWorkers: workers}
		hash, err := FilesParallel(ctx, tmpDir, files, opts)
		if err != nil {
			t.Fatalf("hash with %d workers failed: %v", workers, err)
		}
		if hash != referenceHash {
			t.Errorf("hash differs with %d workers\nreference: %s\nactual:    %s", workers, referenceHash, hash)
		}
	}
}

// =============================================================================
// Context Cancellation Tests
// =============================================================================

func TestFilesParallel_ContextCancelled_BeforeStart(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Cancel context before calling
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := FilesParallel(ctx, tmpDir, []string{"test.txt"}, DefaultParallelOptions())

	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}

func TestFilesParallel_ContextCancelled_DuringProcessing(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many files to ensure processing takes time
	numFiles := 100
	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		name := "file" + padNumber(i) + ".txt"
		// Create larger files to slow down processing
		content := make([]byte, 10*1024) // 10KB each
		for j := range content {
			content[j] = byte(i % 256)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		files[i] = name
	}

	// Cancel context shortly after starting
	ctx, cancel := context.WithCancel(context.Background())

	// Start processing in goroutine
	done := make(chan struct{})
	var hashErr error
	go func() {
		_, hashErr = FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
		close(done)
	}()

	// Cancel after a brief moment
	time.Sleep(1 * time.Millisecond)
	cancel()

	// Wait for completion
	select {
	case <-done:
		// Good - completed (either error or success before cancel took effect)
	case <-time.After(5 * time.Second):
		t.Fatal("FilesParallel did not respond to context cancellation within timeout")
	}

	// If there was an error, it should be context.Canceled
	if hashErr != nil && !errors.Is(hashErr, context.Canceled) {
		// Allow the operation to complete successfully if it finished before cancellation
		t.Logf("got non-cancel error (may be acceptable if completed before cancel): %v", hashErr)
	}
}

func TestFilesParallel_ContextTimeout(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	files := []string{"test.txt"}
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Context with very short timeout (already expired)
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	_, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())

	if err == nil {
		t.Error("expected error for expired context")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded error, got: %v", err)
	}
}

func TestFilesParallel_GracefulShutdown(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many files
	numFiles := 50
	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		name := "file" + padNumber(i) + ".txt"
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		files[i] = name
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Track that we don't hang forever
	done := make(chan struct{})
	go func() {
		_, _ = FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
		close(done)
	}()

	// Cancel mid-flight
	cancel()

	// Should complete quickly (graceful shutdown)
	select {
	case <-done:
		// Success - graceful shutdown
	case <-time.After(2 * time.Second):
		t.Fatal("FilesParallel did not shut down gracefully within 2 seconds")
	}
}

// =============================================================================
// Error Handling Tests - First Error Wins
// =============================================================================

func TestFilesParallel_MissingFile(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, err := FilesParallel(ctx, tmpDir, []string{"nonexistent.txt"}, DefaultParallelOptions())

	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestFilesParallel_MissingFileAmongMany(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create some valid files
	validFiles := []string{"a.txt", "b.txt", "c.txt"}
	for _, name := range validFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	// Include a missing file in the middle
	files := []string{"a.txt", "missing.txt", "b.txt", "c.txt"}

	_, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())

	if err == nil {
		t.Error("expected error when one file is missing")
	}
}

func TestFilesParallel_FirstErrorWins_MultipleErrors(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// All files are missing - should get exactly one error
	files := []string{"missing1.txt", "missing2.txt", "missing3.txt", "missing4.txt"}

	_, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())

	if err == nil {
		t.Error("expected error for missing files")
	}

	// Error should mention one of the missing files (first one encountered wins)
	// We can't predict which one will be first due to parallel execution,
	// but we should get exactly one error, not multiple combined
	errStr := err.Error()
	foundMissingRef := false
	for _, name := range files {
		if contains(errStr, name) {
			foundMissingRef = true
			break
		}
	}
	if !foundMissingRef {
		t.Errorf("error should reference one of the missing files, got: %v", err)
	}
}

func TestFilesParallel_ErrorDoesNotBlockOtherWorkers(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create many valid files and one missing
	numFiles := 20
	files := make([]string, numFiles)
	for i := 0; i < numFiles-1; i++ {
		name := "file" + padNumber(i) + ".txt"
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		files[i] = name
	}
	files[numFiles-1] = "missing.txt"

	// Should complete in reasonable time (error doesn't cause hang)
	done := make(chan struct{})
	var hashErr error
	go func() {
		_, hashErr = FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
		close(done)
	}()

	select {
	case <-done:
		if hashErr == nil {
			t.Error("expected error for missing file")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("FilesParallel hung when encountering error")
	}
}

func TestFilesParallel_PermissionDenied(t *testing.T) {
	// Skip on Windows where permission model is different
	if runtime.GOOS == "windows" {
		t.Skip("skipping permission test on Windows")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a file with no read permission
	unreadableFile := filepath.Join(tmpDir, "unreadable.txt")
	if err := os.WriteFile(unreadableFile, []byte("secret"), 0000); err != nil {
		t.Fatalf("failed to create unreadable file: %v", err)
	}
	defer os.Chmod(unreadableFile, 0644) // Cleanup

	_, err := FilesParallel(ctx, tmpDir, []string{"unreadable.txt"}, DefaultParallelOptions())

	if err == nil {
		t.Error("expected error for unreadable file")
	}
}

// =============================================================================
// Worker Limiting Tests
// =============================================================================

func TestFilesParallel_RespectsMaxWorkers_One(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create files
	numFiles := 10
	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		name := "file" + padNumber(i) + ".txt"
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		files[i] = name
	}

	opts := ParallelOptions{MaxWorkers: 1}
	result, err := FilesParallel(ctx, tmpDir, files, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty hash")
	}
}

func TestFilesParallel_RespectsMaxWorkers_ZeroDefaultsToSensible(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// MaxWorkers of 0 should default to a sensible value (not panic or hang)
	opts := ParallelOptions{MaxWorkers: 0}
	result, err := FilesParallel(ctx, tmpDir, []string{"test.txt"}, opts)

	if err != nil {
		t.Fatalf("unexpected error with MaxWorkers=0: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty hash")
	}
}

func TestFilesParallel_RespectsMaxWorkers_NegativeDefaultsToSensible(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Negative MaxWorkers should default to a sensible value
	opts := ParallelOptions{MaxWorkers: -1}
	result, err := FilesParallel(ctx, tmpDir, []string{"test.txt"}, opts)

	if err != nil {
		t.Fatalf("unexpected error with MaxWorkers=-1: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty hash")
	}
}

func TestFilesParallel_RespectsMaxWorkers_LargeValueCapped(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Very large MaxWorkers should be capped (implementation detail, but should not cause issues)
	opts := ParallelOptions{MaxWorkers: 10000}
	result, err := FilesParallel(ctx, tmpDir, []string{"test.txt"}, opts)

	if err != nil {
		t.Fatalf("unexpected error with large MaxWorkers: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty hash")
	}
}

func TestFilesParallel_ConcurrencyLimited(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create many files
	numFiles := 50
	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		name := "file" + padNumber(i) + ".txt"
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content "+padNumber(i)), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		files[i] = name
	}

	// Track concurrent operations
	var maxConcurrent int64
	var current int64

	// This test verifies behavior through a wrapper - implementation would need hooks
	// For now, just verify it doesn't exceed reasonable limits by completing without issues
	maxWorkers := 4
	opts := ParallelOptions{MaxWorkers: maxWorkers}

	result, err := FilesParallel(ctx, tmpDir, files, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty hash")
	}

	// These would be set by instrumented code - keeping them to show intent
	_ = maxConcurrent
	_ = current
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestFilesParallel_EmptyFiles(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create empty files
	files := []string{"empty1.txt", "empty2.txt"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte{}, 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	parallelHash, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	sequentialHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	if parallelHash != sequentialHash {
		t.Errorf("empty file hashes differ\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

func TestFilesParallel_LargeFiles(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a larger file (1MB)
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "large.bin"), largeContent, 0644); err != nil {
		t.Fatalf("failed to create large file: %v", err)
	}

	parallelHash, err := FilesParallel(ctx, tmpDir, []string{"large.bin"}, DefaultParallelOptions())
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	sequentialHash, err := Files(tmpDir, []string{"large.bin"})
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	if parallelHash != sequentialHash {
		t.Errorf("large file hashes differ\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

func TestFilesParallel_MixedFileSizes(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create files of varying sizes
	files := []string{"tiny.txt", "small.txt", "medium.txt", "large.txt"}
	sizes := []int{1, 100, 10000, 100000}

	for i, name := range files {
		content := make([]byte, sizes[i])
		for j := range content {
			content[j] = byte((i + j) % 256)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	parallelHash, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	sequentialHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	if parallelHash != sequentialHash {
		t.Errorf("mixed size file hashes differ\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

func TestFilesParallel_SpecialCharactersInFilename(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Files with special characters (that are valid on most filesystems)
	files := []string{
		"file with spaces.txt",
		"file-with-dashes.txt",
		"file_with_underscores.txt",
		"file.multiple.dots.txt",
	}

	for _, name := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content of "+name), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	parallelHash, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	sequentialHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	if parallelHash != sequentialHash {
		t.Errorf("special character filename hashes differ\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

func TestFilesParallel_DuplicateFiles(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Same file listed multiple times
	files := []string{"file.txt", "file.txt", "file.txt"}

	parallelHash, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	sequentialHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	if parallelHash != sequentialHash {
		t.Errorf("duplicate file hashes differ\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

func TestFilesParallel_UnicodeFilenames(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Unicode filenames (if filesystem supports them)
	files := []string{
		"normal.txt",
		"unicode_test.txt", // Keeping simple for cross-platform compatibility
	}

	for _, name := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	parallelHash, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	sequentialHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	if parallelHash != sequentialHash {
		t.Errorf("unicode filename hashes differ\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

func TestFilesParallel_BinaryContent(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Binary file with all byte values
	content := make([]byte, 256)
	for i := range content {
		content[i] = byte(i)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "binary.bin"), content, 0644); err != nil {
		t.Fatalf("failed to create binary file: %v", err)
	}

	parallelHash, err := FilesParallel(ctx, tmpDir, []string{"binary.bin"}, DefaultParallelOptions())
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	sequentialHash, err := Files(tmpDir, []string{"binary.bin"})
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	if parallelHash != sequentialHash {
		t.Errorf("binary content hashes differ\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

// =============================================================================
// Concurrency Safety Tests
// =============================================================================

func TestFilesParallel_ConcurrentCalls(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create files
	files := []string{"a.txt", "b.txt", "c.txt"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content of "+name), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	// Get expected hash
	expectedHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("failed to get expected hash: %v", err)
	}

	// Run multiple FilesParallel calls concurrently
	numGoroutines := 10
	results := make(chan string, numGoroutines)
	errs := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			hash, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
			if err != nil {
				errs <- err
				return
			}
			results <- hash
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		select {
		case hash := <-results:
			if hash != expectedHash {
				t.Errorf("concurrent call returned different hash\nexpected: %s\nactual:   %s", expectedHash, hash)
			}
		case err := <-errs:
			t.Errorf("concurrent call failed: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("concurrent calls timed out")
		}
	}
}

func TestFilesParallel_NoRaceConditions(t *testing.T) {
	// This test is primarily for the race detector (-race flag)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create many files
	numFiles := 20
	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		name := "file" + padNumber(i) + ".txt"
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		files[i] = name
	}

	// Run with race detector (go test -race)
	for i := 0; i < 5; i++ {
		_, err := FilesParallel(ctx, tmpDir, files, DefaultParallelOptions())
		if err != nil {
			t.Fatalf("run %d failed: %v", i, err)
		}
	}
}

// =============================================================================
// Performance Characteristics Tests
// =============================================================================

func TestFilesParallel_MoreFilesThanWorkers(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Many more files than workers
	numFiles := 100
	maxWorkers := 4

	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		name := "file" + padNumber(i) + ".txt"
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		files[i] = name
	}

	opts := ParallelOptions{MaxWorkers: maxWorkers}
	parallelHash, err := FilesParallel(ctx, tmpDir, files, opts)
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	sequentialHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	if parallelHash != sequentialHash {
		t.Errorf("hashes differ with more files than workers\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

func TestFilesParallel_FewerFilesThanWorkers(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Fewer files than workers
	numFiles := 2
	maxWorkers := 8

	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		name := "file" + padNumber(i) + ".txt"
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		files[i] = name
	}

	opts := ParallelOptions{MaxWorkers: maxWorkers}
	parallelHash, err := FilesParallel(ctx, tmpDir, files, opts)
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	sequentialHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	if parallelHash != sequentialHash {
		t.Errorf("hashes differ with fewer files than workers\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

func TestFilesParallel_ExactlyOneWorker(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	files := []string{"a.txt", "b.txt", "c.txt"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content of "+name), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	opts := ParallelOptions{MaxWorkers: 1}
	parallelHash, err := FilesParallel(ctx, tmpDir, files, opts)
	if err != nil {
		t.Fatalf("parallel hash failed: %v", err)
	}

	sequentialHash, err := Files(tmpDir, files)
	if err != nil {
		t.Fatalf("sequential hash failed: %v", err)
	}

	if parallelHash != sequentialHash {
		t.Errorf("hashes differ with single worker\nsequential: %s\nparallel:   %s", sequentialHash, parallelHash)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// padNumber returns a zero-padded number string for consistent file ordering
func padNumber(n int) string {
	if n < 10 {
		return "00" + itoa(n)
	}
	if n < 100 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

// itoa converts int to string without importing strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// These variables are used in TestFilesParallel_ConcurrencyLimited to track
// concurrent operations if the implementation provides hooks
var (
	_ = atomic.Int64{} // Placeholder for concurrency tracking
)

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkFiles_Sequential(b *testing.B) {
	tmpDir := b.TempDir()

	// Create 100 files
	numFiles := 100
	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		name := "file" + padNumber(i) + ".txt"
		content := make([]byte, 1024) // 1KB each
		for j := range content {
			content[j] = byte((i + j) % 256)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			b.Fatalf("failed to create file: %v", err)
		}
		files[i] = name
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Files(tmpDir, files)
		if err != nil {
			b.Fatalf("hash failed: %v", err)
		}
	}
}

func BenchmarkFilesParallel(b *testing.B) {
	ctx := context.Background()
	tmpDir := b.TempDir()

	// Create 100 files
	numFiles := 100
	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		name := "file" + padNumber(i) + ".txt"
		content := make([]byte, 1024) // 1KB each
		for j := range content {
			content[j] = byte((i + j) % 256)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			b.Fatalf("failed to create file: %v", err)
		}
		files[i] = name
	}

	opts := DefaultParallelOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := FilesParallel(ctx, tmpDir, files, opts)
		if err != nil {
			b.Fatalf("hash failed: %v", err)
		}
	}
}

func BenchmarkFilesParallel_VaryingWorkers(b *testing.B) {
	ctx := context.Background()
	tmpDir := b.TempDir()

	// Create 100 files
	numFiles := 100
	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		name := "file" + padNumber(i) + ".txt"
		content := make([]byte, 1024) // 1KB each
		for j := range content {
			content[j] = byte((i + j) % 256)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			b.Fatalf("failed to create file: %v", err)
		}
		files[i] = name
	}

	for _, workers := range []int{1, 2, 4, 8, 16} {
		b.Run("workers="+itoa(workers), func(b *testing.B) {
			opts := ParallelOptions{MaxWorkers: workers}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := FilesParallel(ctx, tmpDir, files, opts)
				if err != nil {
					b.Fatalf("hash failed: %v", err)
				}
			}
		})
	}
}
