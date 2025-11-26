# Orchestrator Package

A shared parallel execution orchestrator for build and test commands that solves the Windows terminal output interleaving problem.

## Problem Statement

When running modules in parallel with goroutines, multiple workers writing to stdout simultaneously causes:

- Garbled/interleaved output on Windows terminals
- Lines split across multiple rows
- Horizontal scrolling issues
- Difficult-to-read status updates

## Solution

The orchestrator uses a **single output goroutine** pattern:

- Only one goroutine writes to stdout/stderr (the display manager)
- Worker goroutines send updates through channels
- Display manager handles all console output serialization
- Results in clean, non-interleaved output

## Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│                      Orchestrator                           │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │   Worker 1   │  │   Worker 2   │  │   Worker N   │       │
│  │  (goroutine) │  │  (goroutine) │  │  (goroutine) │       │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘       │
│         │                 │                 │               │
│         └─────────────────┴─────────────────┘               │
│                           │                                 │
│                           ▼                                 │
│                  ┌─────────────────┐                        │
│                  │ Completion Chan │                        │
│                  └────────┬────────┘                        │
│                           │                                 │
│                           ▼                                 │
│                  ┌─────────────────┐                        │
│                  │ Display Manager │                        │
│                  │  (goroutine)    │                        │
│                  └────────┬────────┘                        │
│                           │                                 │
│                           ▼                                 │
│                      stdout/stderr ─────────────────────────┤
└─────────────────────────────────────────────────────────────┘
```

## Key Features

1. **Parallel Execution**: Configurable concurrency with semaphore-based control
2. **Status Updates**: Periodic updates every 2 seconds showing which modules are running
3. **Clean Completion Lines**: Each module prints exactly one completion line (no interleaving)
4. **Progress Tracking**: Shows elapsed time, completed count, and currently running modules
5. **Windows Compatible**: No horizontal scrolling or garbled output

## Usage

### Basic Usage

```go
import "github.com/ready-to-release/eac/src/commands/internal/orchestrator"

// Configure orchestrator
config := orchestrator.Config{
    WorkspaceRoot:        "/path/to/repo",
    OutputBaseDir:        "out/build",
    LogFileName:          "build.log",
    OrchestratorLogName:  "orchestrator.log",
    ActionVerb:           "building",
    MaxConcurrency:       0, // 0 = use number of CPUs
    StatusUpdateInterval: 2, // Update every 2 seconds
}

// Create worker function
worker := func(moniker string, logWriter io.Writer) int {
    // Do work here - write all output to logWriter, not stdout
    fmt.Fprintf(logWriter, "Processing %s\n", moniker)
    return 0 // Return exit code
}

// Run orchestrator
orch := orchestrator.New(config, worker)
results, err := orch.Run([]string{"module1", "module2", "module3"})
if err != nil {
    log.Fatal(err)
}

// Print summary
orch.PrintSummary(results)

// Get exit code
return orchestrator.GetExitCode(results)
```

### Example Output

```text
Building 3 modules in parallel: [src-core src-cli src-commands]

Status: 2s elapsed, 0/3 completed. 3 running (src-cli, src-commands, src-core)
[building] src-core (See out/build/src-core/build.log for details) ........ Done
Status: 4s elapsed, 1/3 completed. 2 running (src-cli, src-commands)
[building] src-cli (See out/build/src-cli/build.log for details) ........ Done
[building] src-commands (See out/build/src-commands/build.log for details) ........ Done

===========================================
Building Summary
===========================================
Total modules: 3
Passed: 3
Failed: 0
Warnings: 0 (in 0 modules)

Orchestrator log: out/build/orchestrator.log
Module logs: out/build/<module>/build.log
```

## Components

### orchestrator.go

Main orchestrator logic including:

- `Orchestrator` struct and `New()` constructor
- `Run()` method for executing work items
- `executeParallel()` for parallel execution with semaphore
- `processWorkItem()` for individual work item handling
- `PrintSummary()` for summary output
- `GetExitCode()` helper

### display.go

Display manager for console output:

- `displayManager` struct
- Single-goroutine display loop
- Status update ticker
- Completion message handling
- Duration formatting

### types.go

Type definitions:

- `WorkItem` - unit of work
- `WorkResult` - outcome of work
- `WorkerFunc` - worker function signature
- `Config` - orchestrator configuration

### parser.go

Log parsing utilities:

- `parseLogForIssues()` - extract warnings and errors from log files

## Integration Examples

### Build Command

```go
// Create worker function that builds a single module
worker := func(moniker string, logWriter io.Writer) int {
    module, exists := moduleReport.Registry.Get(moniker)
    if !exists {
        fmt.Fprintf(logWriter, "Error: module not found: %s\n", moniker)
        return 1
    }

    moduleOutputDir := filepath.Join(workspaceRoot, "out", "build", moniker)
    return runModuleBuild(module, workspaceRoot, moduleOutputDir, logWriter, tidyFirst)
}
```

### Test Command

```go
// Create worker function that tests a single module
worker := func(moniker string, logWriter io.Writer) int {
    module, exists := moduleReport.Registry.Get(moniker)
    if !exists {
        fmt.Fprintf(logWriter, "Error: module not found: %s\n", moniker)
        return 1
    }

    moduleOutputDir := filepath.Join(workspaceRoot, "out", "test", moniker)
    return runModuleTest(module, workspaceRoot, moduleOutputDir, logWriter, reportFormat, suiteName)
}
```

## Benefits

1. **Correctness**: No interleaved output on any platform
2. **Clarity**: Clean status updates and completion messages
3. **Performance**: True parallel execution with configurable concurrency
4. **Reusability**: Shared code between build and test commands
5. **Maintainability**: Single source of truth for orchestration logic

## Testing

The package includes comprehensive tests in `orchestrator_test.go`:

- Basic orchestration with multiple modules
- Failure handling and exit codes
- Log parsing for warnings/errors
- Duration formatting
- Display manager functionality
- Summary generation

Run tests with:

```bash
go test ./src/commands/internal/orchestrator/...
```

## Future Enhancements

Possible improvements:

- Real-time log streaming for verbose mode
- Progress bars for long-running operations
- Colorized output based on module status
- Dependency-aware scheduling (respect module dependencies)
- Resource-aware concurrency (adjust based on system load)
