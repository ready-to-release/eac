package orchestrator

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/ansi"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// StartBackgroundCacheDetection spawns goroutines to verify cache status in parallel.
// Uses I/O-appropriate parallelism (NOT the weighted semaphore - this is lightweight).
// Updates TUI and marks items as early-cached for worker short-circuit.
//
// The verifier interface checks cache status for a component and returns
// whether it's cached and when. This allows different commands (build, test, etc.)
// to provide their own cache verification logic via the execution.CacheVerifier interface.
//
// Background detection does NOT report to summaryBuilder - workers are the sole
// source of truth for summarization. This prevents duplicate results.
func (us *UnitScheduler) StartBackgroundCacheDetection(
	work []workunit.UnitSpec,
	cachedModules map[string]bool,
	cacheTimes map[string]time.Time,
	verifier CacheVerifier,
) {
	if verifier == nil || len(work) == 0 {
		return
	}

	go func() {
		// Panic recovery - don't crash if detection fails
		// Workers will still check cache normally
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "[cache-detect] panic recovered: %v\n", r)
			}
		}()

		// I/O-appropriate parallelism (NOT weighted semaphore)
		// Cache checks are file reads + hash computation, not heavy builds
		maxParallel := 32
		if len(work) < maxParallel {
			maxParallel = len(work)
		}
		sem := make(chan struct{}, maxParallel)

		var wg sync.WaitGroup

		// Create a context for cache verification
		// Background context is appropriate since detection runs to completion
		ctx := context.Background()

		for _, spec := range work {
			wg.Add(1)
			go func(s workunit.UnitSpec) {
				defer wg.Done()

				// Acquire I/O slot
				sem <- struct{}{}
				defer func() { <-sem }()

				// Use provided verification logic via interface
				result, err := verifier.Verify(ctx, s)
				if err != nil {
					// Log error but continue - workers will still check cache normally
					logging.C().Debugf("[cache-detect] verification error for %s: %v", s.ID.Longname(), err)
					return
				}

				if result.Cached {
					moniker := s.ID.Longname()

					// 1. Emit event to observers - tab lights up blue
					us.emit(core.UnitCompletedEvent{
						Time:     time.Now(),
						ID:       moniker,
						ExitCode: -1, // Exit code -1 = cached
					})

					// 2. Mark as early-cached so worker can short-circuit
					us.earlyCached.Store(moniker, EarlyCacheInfo{
						Module:    s.ID.Module,
						Component: s.ID.ComponentName,
						Handler:   s.ID.Tool,
						CacheTime: result.CacheTime,
					})

					// NOTE: Do NOT increment tuiCompleted here!
					// Workers are the sole source of truth for completion counting.
					// This prevents a race where background marks item cached after
					// worker started, causing worker to skip its counter increment.
				}

				// NOTE: Do NOT report to summaryBuilder here!
				// Workers are the sole source of truth for summarization.
				// This prevents duplicate results.
			}(spec)
		}

		wg.Wait()
	}()
}

// executeWorker runs the actual work for a component.
// Called by dispatcher after capacity is acquired and dependencies satisfied.
// This is the core execution extracted from processComponent without dep-waiting or semaphore handling.
func (us *UnitScheduler) executeWorker(spec workunit.UnitSpec, worker UnitWorkerFunc) UnitResult {
	module := spec.ID.Module
	component := spec.ID.ComponentName
	tool := spec.ID.Tool

	result := UnitResult{
		Longname:  spec.ID.Longname(),
		Module:    module,
		Component: component,
		Handler:   tool,
	}

	moniker := spec.ID.Longname()
	displayName := spec.DisplayName()

	// FAST PATH: Check if background already verified cached
	// This enables fast termination - workers don't re-do cache checks
	// TUI was already updated by background detection - just return result for summary
	if cacheInfo, ok := us.earlyCached.Load(moniker); ok {
		info := cacheInfo.(EarlyCacheInfo)
		return UnitResult{
			Longname:  moniker,
			Module:    info.Module,
			Component: info.Component,
			Handler:   info.Handler,
			ExitCode:  -1, // Cached
			Duration:  0,  // Instant - no actual work done
		}
	}

	// Start timing - duration measures actual execution time
	startTime := time.Now()
	result.StartedAt = startTime

	// Track active tool usage
	isContainer := spec.GetPoolAllocation().IsContainer()
	us.addActiveTool(tool, isContainer, moniker)

	// Create output directory for this component
	// Structure: out/<context>/<module>/<dirname> (e.g., out/build/books/howto, out/test/eac/go_gotest_impl-build)
	// Uses DirName() which includes tool and Extra values for unique directory names
	sanitizedModule := sanitizePathForFS(output.PackageDisplayName(module))
	sanitizedDirName := sanitizePathForFS(spec.ID.DirName())
	componentOutputDir := filepath.Join(us.config.WorkspaceRoot, us.config.OutputBaseDir, sanitizedModule, sanitizedDirName)

	// Relative log path for result reporting: out/<context>/<module>/<dirname>/<logfile>
	relLogPath := filepath.Join(us.config.OutputBaseDir, sanitizedModule, sanitizedDirName, us.config.LogFileName())

	if err := os.MkdirAll(componentOutputDir, 0o755); err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create directory: %v", err)}
		result.LogPath = relLogPath
		result.Duration = time.Since(startTime)
		us.removeActiveTool(tool, isContainer, moniker)
		return result
	}

	// Create log file
	logPath := filepath.Join(componentOutputDir, us.config.LogFileName())
	logFile, err := os.Create(logPath)
	if err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create log file: %v", err)}
		result.LogPath = relLogPath
		result.Duration = time.Since(startTime)
		us.removeActiveTool(tool, isContainer, moniker)
		return result
	}

	// Wrap log file with bad ANSI filter - preserves colors, strips control sequences
	filteredLogFile := ansi.NewBadOnlyFilter(logFile, displayName)

	// Create writer for worker - use writerFactory if available (e.g., TUIObserver)
	// Use moniker (longname) to match TUI tab identification
	var workerWriter io.Writer
	if us.writerFactory != nil {
		workerWriter = us.writerFactory.NewWriter(moniker, filteredLogFile)
	} else {
		workerWriter = filteredLogFile
	}

	// Execute work with memory instrumentation
	memBefore := GetMemoryStats()
	fmt.Fprintf(logFile, "[memory] before: used=%s avail=%s total=%s (%.1f%%)\n",
		FormatBytes(memBefore.UsedBytes), FormatBytes(memBefore.AvailableBytes),
		FormatBytes(memBefore.TotalBytes), memBefore.UsedPercent)

	// Get worker timeout from config (default 5m if not set)
	timeout := config.WorkerTimeout()
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	// Docker image builds need significantly more time than the generic worker timeout.
	// Cold builds (no layer cache) involve pulling base images, installing system packages,
	// and downloading large dependencies which easily exceed the default worker timeout.
	// Once layers are cached, subsequent builds complete in seconds.
	if tool == "buildx" || tool == "docker" {
		timeout = config.DockerImageBuildTimeout()
	}

	// Create cancellable context with timeout — propagated to worker and its subprocesses
	workerCtx, workerCancel := context.WithTimeout(context.Background(), timeout)
	defer workerCancel()

	// Run worker with timeout enforcement
	var exitCode int
	resultCh := make(chan int, 1)
	go func() {
		// Pass full spec to worker - no more string joining/parsing
		resultCh <- worker(workerCtx, spec, workerWriter)
	}()

	select {
	case exitCode = <-resultCh:
		// Worker completed normally
	case <-workerCtx.Done():
		// Worker timed out — context cancellation kills subprocesses via exec.CommandContext
		exitCode = workunit.ExitCodeTimeout
		fmt.Fprintf(os.Stderr, "TIMEOUT: %s killed after %v (limit: %v)\n", displayName, time.Since(startTime), timeout)
		fmt.Fprintf(logFile, "\n[TIMEOUT] Worker killed after %v (limit: %v)\n", time.Since(startTime), timeout)
	}

	memAfter := GetMemoryStats()
	memDelta := int64(memAfter.UsedBytes) - int64(memBefore.UsedBytes)
	deltaSign := "+"
	if memDelta < 0 {
		deltaSign = ""
	}
	fmt.Fprintf(logFile, "[memory] after: used=%s avail=%s total=%s (%.1f%%) delta=%s%s\n",
		FormatBytes(memAfter.UsedBytes), FormatBytes(memAfter.AvailableBytes),
		FormatBytes(memAfter.TotalBytes), memAfter.UsedPercent,
		deltaSign, FormatBytes(uint64(abs64(memDelta))))

	// Close TUI writer first (flushes pipe), then log file
	if closer, ok := workerWriter.(io.Closer); ok {
		closer.Close()
	}
	logFile.Close()

	// Parse log for warnings/errors
	warnings, errors := parseLogForIssues(logPath)

	result.ExitCode = exitCode
	result.Warnings = warnings
	result.Errors = errors
	result.LogPath = relLogPath
	result.Duration = time.Since(startTime)

	// Merge any extras set by the worker (e.g., test counts)
	if extras, ok := us.getUnitExtras(module, component); ok {
		result.TestsTotal = extras.TestsTotal
		result.TestsPassed = extras.TestsPassed
		result.TestsFailed = extras.TestsFailed
		result.TestsSkipped = extras.TestsSkipped
	}

	// Remove tool from active list
	us.removeActiveTool(tool, isContainer, moniker)

	return result
}
