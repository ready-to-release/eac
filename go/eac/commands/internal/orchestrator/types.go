package orchestrator

import (
	"io"
	"time"
)

// WorkItem represents a single unit of work to be processed
type WorkItem struct {
	// Moniker is the unique identifier for this work item
	Moniker string
	// Index is the original position in the input slice (for ordering results)
	Index int
}

// WorkResult represents the outcome of processing a work item
type WorkResult struct {
	// Moniker is the unique identifier for this work item
	Moniker string
	// Index is the original position in the input slice
	Index int
	// ExitCode is the exit code from the worker function (0 = success)
	ExitCode int
	// LogPath is the relative path to the detailed log file
	LogPath string
	// Warnings are any warning messages collected during processing
	Warnings []string
	// Errors are any error messages collected during processing
	Errors []string
	// Duration is the time taken to process this work item
	Duration time.Duration
	// Type is the module/work type (e.g., "go-cli", "mkdocs", "go")
	Type string
}

// WorkerFunc is a function that processes a single work item
// It receives the moniker and should return an exit code (0 for success)
// All output should go to the provided logWriter (not stdout/stderr)
type WorkerFunc func(moniker string, logWriter io.Writer) int

// Config holds orchestrator configuration options
type Config struct {
	// WorkspaceRoot is the root directory of the repository
	WorkspaceRoot string
	// OutputBaseDir is the base directory for all output (e.g., "out/build" or "out/test")
	OutputBaseDir string
	// LogFileName is the name of the log file for each module (e.g., "build.log" or "test.log")
	LogFileName string
	// OrchestratorLogName is the name of the orchestrator summary log
	OrchestratorLogName string
	// ActionVerb is the present continuous verb for status messages (e.g., "building", "testing")
	ActionVerb string
	// MaxConcurrency is the maximum number of concurrent workers (0 = number of CPUs)
	MaxConcurrency int
	// StatusUpdateInterval is how often to show status updates (default: 2 seconds)
	StatusUpdateInterval int // in seconds
	// ModuleTypes maps moniker to type string (e.g., "go-cli", "mkdocs") for display
	ModuleTypes map[string]string
	// ShowTimings enables the timing summary section (use --timings flag)
	ShowTimings bool
}
