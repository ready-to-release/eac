// Package hash provides deterministic file content hashing for change detection.
package hash

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
)

// ParallelOptions configures parallel hashing behavior.
type ParallelOptions struct {
	// MaxWorkers limits concurrent file operations.
	// Default: runtime.NumCPU(), capped at 16.
	MaxWorkers int
}

// DefaultParallelOptions returns sensible defaults for parallel hashing.
// Uses half of available CPUs as a circuit breaker, with floor of min(4, NumCPU) and cap of 8.
func DefaultParallelOptions() ParallelOptions {
	numCPU := runtime.NumCPU()
	workers := numCPU / 2
	floor := 4
	if numCPU < floor {
		floor = numCPU
	}
	if workers < floor {
		workers = floor
	}
	if workers > 8 {
		workers = 8
	}
	return ParallelOptions{
		MaxWorkers: workers,
	}
}

// normalizeWorkers ensures MaxWorkers is within valid range.
// Uses half of available CPUs as default, floor of min(4, NumCPU), cap of 8.
func normalizeWorkers(opts ParallelOptions) int {
	numCPU := runtime.NumCPU()
	workers := opts.MaxWorkers
	if workers <= 0 {
		workers = numCPU / 2
		floor := 4
		if numCPU < floor {
			floor = numCPU
		}
		if workers < floor {
			workers = floor
		}
	}
	if workers > 8 {
		workers = 8
	}
	return workers
}

// fileHashResult holds the result of hashing a single file.
type fileHashResult struct {
	data []byte // The bytes written to hasher for this file (filename + content)
	err  error
}

// FilesParallel computes a SHA-256 hash of file contents using parallel workers.
// It produces IDENTICAL results to the sequential Files() function.
// Files are processed in parallel but combined in sorted order for determinism.
// Returns an empty string for an empty file list.
func FilesParallel(ctx context.Context, workspaceRoot string, files []string, opts ParallelOptions) (string, error) {
	if len(files) == 0 {
		return "", nil
	}

	// Check context before starting
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Sort files for determinism (matching sequential behavior)
	sorted := make([]string, len(files))
	copy(sorted, files)
	sort.Strings(sorted)

	// Pre-allocate results slice (ordered by sorted index)
	results := make([]fileHashResult, len(sorted))

	// Worker pool with semaphore
	workers := normalizeWorkers(opts)
	sem := make(chan struct{}, workers)

	var wg sync.WaitGroup
	var firstErr error
	var errOnce sync.Once

	// Context for cancellation on first error
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Fan out work to workers
	for i, file := range sorted {
		// Check context before scheduling
		select {
		case <-ctx.Done():
			// Wait for in-flight goroutines
			wg.Wait()
			if firstErr != nil {
				return "", firstErr
			}
			return "", ctx.Err()
		default:
		}

		// Acquire semaphore slot
		select {
		case <-ctx.Done():
			wg.Wait()
			if firstErr != nil {
				return "", firstErr
			}
			return "", ctx.Err()
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(idx int, path string) {
			defer func() {
				<-sem // Release semaphore
				wg.Done()
			}()

			// Check context before file operation
			select {
			case <-ctx.Done():
				return
			default:
			}

			data, err := hashSingleFile(workspaceRoot, path)
			if err != nil {
				results[idx] = fileHashResult{err: err}
				errOnce.Do(func() {
					firstErr = err
					cancel() // Cancel remaining work
				})
				return
			}
			results[idx] = fileHashResult{data: data}
		}(i, file)
	}

	wg.Wait()

	// Check for errors
	if firstErr != nil {
		return "", firstErr
	}

	// Check context one more time
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Combine hashes in order (matching sequential behavior)
	h := sha256.New()
	for _, r := range results {
		if r.err != nil {
			return "", r.err
		}
		h.Write(r.data)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// hashSingleFile reads a file and returns the bytes that would be written to the hasher.
// This matches the sequential Files() behavior: filename + "\n" + file content.
func hashSingleFile(workspaceRoot, file string) ([]byte, error) {
	path := filepath.Join(workspaceRoot, file)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", file, err)
	}
	defer f.Close()

	// Build the data that would be written to hasher
	// First: filename + newline
	header := []byte(file + "\n")

	// Then: file content
	content, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", file, err)
	}

	// Combine header and content
	result := make([]byte, len(header)+len(content))
	copy(result, header)
	copy(result[len(header):], content)

	return result, nil
}
