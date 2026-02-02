// Package books provides book preprocessing for MkDocs sites.
package books

import (
	"bytes"
	"os"
	"runtime"
	"sync"
)

// FileProcessor is a function that processes file content and returns modified content.
// If no modification is needed, it should return the original content unchanged.
type FileProcessor func(path string, content []byte) ([]byte, error)

// ProcessFilesParallel processes multiple files in parallel using a worker pool.
// It reads each file, applies the processor function, and writes back if modified.
// The number of workers is limited to the number of CPU cores.
func ProcessFilesParallel(files []string, process FileProcessor) error {
	if len(files) == 0 {
		return nil
	}

	// Limit concurrency to CPU cores
	numWorkers := runtime.NumCPU()
	if numWorkers > len(files) {
		numWorkers = len(files)
	}

	// Create job channel
	jobs := make(chan string, len(files))
	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	// Collect errors
	var mu sync.Mutex
	var errors []error

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				if err := processOneFile(path, process); err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// Return first error if any
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

// processOneFile reads, processes, and writes a single file.
func processOneFile(path string, process FileProcessor) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	modified, err := process(path, content)
	if err != nil {
		return err
	}

	// Only write if content changed
	if !bytes.Equal(content, modified) {
		if err := os.WriteFile(path, modified, 0o644); err != nil {
			return err
		}
	}

	return nil
}

// ProcessFilesParallelWithCache processes files in parallel, using the hash cache
// to skip unchanged files. Returns the number of files processed (not skipped).
func ProcessFilesParallelWithCache(
	files []string,
	cache *FileHashCache,
	process FileProcessor,
) (int, error) {
	if len(files) == 0 {
		return 0, nil
	}

	// Limit concurrency to CPU cores
	numWorkers := runtime.NumCPU()
	if numWorkers > len(files) {
		numWorkers = len(files)
	}

	// Create job channel
	jobs := make(chan string, len(files))
	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	// Collect errors and count
	var mu sync.Mutex
	var errors []error
	var processedCount int

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				processed, err := processOneFileWithCache(path, cache, process)
				mu.Lock()
				if err != nil {
					errors = append(errors, err)
				}
				if processed {
					processedCount++
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Return first error if any
	if len(errors) > 0 {
		return processedCount, errors[0]
	}
	return processedCount, nil
}

// processOneFileWithCache reads, checks cache, processes if needed, and writes.
func processOneFileWithCache(path string, cache *FileHashCache, process FileProcessor) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	// Check if file needs processing
	if cache != nil && !cache.ShouldProcessFile(path, content) {
		return false, nil // Cache hit - skip
	}

	modified, err := process(path, content)
	if err != nil {
		return false, err
	}

	// Only write if content changed
	if !bytes.Equal(content, modified) {
		if err := os.WriteFile(path, modified, 0o644); err != nil {
			return false, err
		}
	}

	return true, nil
}
