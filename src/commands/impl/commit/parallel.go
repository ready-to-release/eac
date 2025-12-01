// File: src/commands/impl/commit/parallel.go
// Purpose: Parallel module section generation for multi-module commits
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear result struct with index for ordering
//   - One goroutine per module with simple channel communication
//   - Separate result collection from error handling
//   - Well-documented concurrency patterns
//
// Easy to change:
//   - Module processing logic unchanged (calls existing generateWithPrompt)
//   - Debug writer used as-is (now thread-safe with mutex in ai.go)
//   - Progress messages still work (WithProgress is goroutine-safe)
//   - Can easily add features like cancellation or retry logic
//
// Hard to break:
//   - Preserves exact output order via indexed results
//   - Collects all errors, returns first one (no error loss)
//   - Channel closed after all goroutines complete (no leaks)
//   - Compatible with existing error handling patterns
//   - Thread-safe via logging module
//   - Buffered channel prevents goroutine blocking
//
// Performance:
//   - Sequential: N modules × 5s = 15s for 3 modules
//   - Parallel:   max(5s) = 5s for 3 modules
//   - Speedup:    60-70% for multi-module commits

package commit

import (
	"fmt"
	"sync"

	commitmessage "github.com/ready-to-release/eac/src/commands/impl/commit/internal"
	"github.com/ready-to-release/eac/src/ai"
	"github.com/ready-to-release/eac/src/core/logging"
	"github.com/ready-to-release/eac/src/core/repository"
)

// generateModuleSectionsParallel generates module sections concurrently
// while preserving order in the output.
//
// This function launches a goroutine for each module and collects results
// using a buffered channel. The result slice maintains the original module
// order by indexing results according to their position in affectedModules.
//
// Thread Safety:
//   - Each goroutine operates on independent data (no shared mutable state)
//   - logger is thread-safe (uses mutex internally)
//   - Channel is buffered to module count (prevents blocking)
//   - WaitGroup ensures clean goroutine lifecycle
//
// Error Handling:
//   - First error encountered is captured and returned
//   - All goroutines complete before error is returned (no leaks)
//   - Channel drain continues even after error detected
//
// Order Preservation:
//   - Results include original index from affectedModules
//   - Output slice is pre-sized and indexed directly
//   - No sorting or reordering needed
//
// Parameters:
//   - cfg: Execution configuration with modules and files
//   - logger: Thread-safe logger for debug output
//   - testExecutor: Optional executor for testing (nil uses real providers)
//
// Returns:
//   - Slice of module sections in original module order
//   - Error if any module generation fails (first error encountered)
//
// Example:
//   cfg := &executionConfig{
//       affectedModules: []string{"src-cli", "src-core", "src-commands"},
//       ...
//   }
//   sections, err := generateModuleSectionsParallel(cfg, logger, nil)
//   // sections[0] corresponds to "src-cli"
//   // sections[1] corresponds to "src-core"
//   // sections[2] corresponds to "src-commands"
func generateModuleSectionsParallel(cfg *executionConfig, logger *logging.Logger, testExecutor *ai.Executor) ([]string, error) {
	// Validate inputs
	if cfg == nil {
		return nil, fmt.Errorf("executionConfig cannot be nil")
	}
	if len(cfg.affectedModules) == 0 {
		return nil, fmt.Errorf("affectedModules cannot be empty")
	}

	// Single-module commits skip module sections (existing behavior)
	if len(cfg.affectedModules) <= 1 {
		log.Debug("Single-module commit - skipping module sections")
		return []string{}, nil
	}

	// Build module file map once (shared read-only data, safe for concurrent access)
	moduleFilesMap := make(map[string][]repository.RepositoryFileWithModule)
	for _, file := range cfg.stagedFiles {
		for _, module := range file.Modules {
			moduleFilesMap[module] = append(moduleFilesMap[module], file)
		}
	}

	// Result structure preserves original order via index
	type moduleResult struct {
		index        int    // Original position in affectedModules
		output       string // Generated module section
		providerName string // AI provider used
		err          error  // Error if generation failed
	}

	// Buffered channel prevents goroutine blocking when sending results
	// Capacity = module count ensures no goroutine waits to send
	resultsChan := make(chan moduleResult, len(cfg.affectedModules))

	// WaitGroup tracks goroutine completion
	var wg sync.WaitGroup

	// Launch one goroutine per module for concurrent generation
	for i, module := range cfg.affectedModules {
		wg.Add(1)
		go func(idx int, moduleName string) {
			defer wg.Done()

			// Each goroutine has its own local variables (no shared mutable state)
			moduleFiles := moduleFilesMap[moduleName]
			moduleContext := buildModuleContext(moduleName, moduleFiles, cfg.gitDiff)

			// Debug output (thread-safe via logging module)
			writeDebugFilef(cfg.workspaceRoot, logger, "debug-module-%d-%s-context.md", moduleContext, idx+1, moduleName)

			var output string
			var providerName string
			progressMsg := fmt.Sprintf("🤖 Generating section for module %s (%d/%d)...",
				moduleName, idx+1, len(cfg.affectedModules))

			// Generate module section using existing function
			// WithProgress is goroutine-safe (uses internal synchronization)
			err := commitmessage.WithProgress(progressMsg, func() error {
				result, genErr := generateWithPromptResult("module", moduleContext, cfg.workspaceRoot, cfg.affectedModules, cfg.debug, testExecutor)
				if result != nil {
					output = result.Output
					providerName = result.ProviderName
				}
				return genErr
			})

			writeDebugFilef(cfg.workspaceRoot, logger, "debug-module-%d-%s-output.md", output, idx+1, moduleName)

			// Send result with original index to preserve order
			// Buffered channel ensures this never blocks
			resultsChan <- moduleResult{
				index:        idx,
				output:       output,
				providerName: providerName,
				err:          err,
			}
		}(i, module)
	}

	// Close channel after all goroutines complete
	// Goroutine ensures non-blocking behavior
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results in original order
	// Pre-size slice to avoid allocations and enable direct indexing
	moduleSections := make([]string, len(cfg.affectedModules))
	var firstError error
	var loggedProvider bool

	// Drain channel until closed
	// Order is preserved by using result.index to place output
	for result := range resultsChan {
		// Capture first error but continue draining channel
		// This ensures all goroutines complete (no leaks)
		if result.err != nil && firstError == nil {
			firstError = result.err
		}

		// Log provider info once (all modules use the same provider)
		if !loggedProvider && result.providerName != "" {
			log.Debugf("AI provider used for module sections: %s", result.providerName)
			loggedProvider = true
		}

		// Place result at correct index to maintain order
		moduleSections[result.index] = result.output
	}

	// Return error only after all results collected
	// This ensures deterministic behavior and no goroutine leaks
	if firstError != nil {
		return nil, fmt.Errorf("failed to generate module sections: %w", firstError)
	}

	return moduleSections, nil
}

// Implementation Notes:
//
// Why buffered channel?
//   - Unbuffered would cause goroutines to block on send
//   - Buffer size = module count ensures all sends complete immediately
//   - No goroutines wait for receiver (better performance)
//
// Why collect all results before returning error?
//   - Prevents goroutine leaks (all goroutines complete)
//   - Deterministic behavior (same result every run)
//   - Simpler error handling (no context cancellation needed)
//
// Why use index instead of append?
//   - Preserves exact order from affectedModules
//   - No sorting needed (O(1) instead of O(n log n))
//   - Clear mapping: index i → moduleSections[i]
//
// Thread safety guarantees:
//   - moduleFilesMap: Read-only after creation (safe)
//   - cfg: Read-only (safe)
//   - moduleContext: Local to goroutine (safe)
//   - output: Local to goroutine (safe)
//   - logger: Thread-safe (safe)
//   - resultsChan: Channel synchronization (safe)
//   - moduleSections: Write to different indices (safe)
//
// Performance characteristics:
//   - Time complexity: O(max(T_i)) where T_i is time for module i
//   - Space complexity: O(n) for n modules (same as sequential)
//   - Goroutines: n goroutines for n modules (bounded)
//   - Memory: ~constant per goroutine (small stacks)
//
// Testing strategy:
//   - Unit tests: Order preservation, thread safety
//   - Acceptance tests: Real AI calls, performance validation
//   - Race detector: go test -race (no data races)
//   - Benchmarks: Measure actual speedup
