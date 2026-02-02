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
import "github.com/ready-to-release/eac/go/clibase/orchestrator"

// Configure orchestrator
config := orchestrator.Config{
    WorkspaceRoot:        "/path/to/repo",
    OutputBaseDir:        "out/build",
    LogFileName:          "build.log",
    OrchestratorLogName:  "orchestrator.log",
    ActionVerb:           "building",
    MaxConcurrency:       0, // 0 = use number of CPUs
    StatusUpdateInterval: 500, // Update every 500ms
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
Building 3 modules in parallel: [eac-core r2r-cli eac-commands]

Status: 2s elapsed, 0/3 completed. 3 running (r2r-cli, eac-commands, eac-core)
[building] eac-core (See out/build/eac-core/build.log for details) ........ Done
Status: 4s elapsed, 1/3 completed. 2 running (r2r-cli, eac-commands)
[building] r2r-cli (See out/build/r2r-cli/build.log for details) ........ Done
[building] eac-commands (See out/build/eac-commands/build.log for details) ........ Done

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
go test ./go/clibase/orchestrator/...
```

## Dynamic Capacity System

The orchestrator uses a dynamic capacity system that automatically adjusts parallelism based on available system resources.

### Capacity Calculation

```
Capacity = (Available RAM / 256MB) × turbo
```

| Mode | Formula | Example (6GB RAM) |
|------|---------|-------------------|
| Normal | RAM / 256MB × 1 | 22 slots |
| Turbo 2x | RAM / 256MB × 2 | 44 slots |
| Turbo 4x | RAM / 256MB × 4 | 64 slots (capped) |

The capacity is recalculated every 2 seconds to adapt to changing system conditions.

### Memory Detection

The system intelligently detects available memory based on the execution environment:

1. **Docker** (preferred): Queries `docker info --format {{.MemTotal}}`
   - Reflects WSL2 memory limit on Windows
   - Most accurate for containerized builds

2. **WSL**: Queries `wsl -e cat /proc/meminfo`
   - Gets actual memory available inside WSL
   - Used when Docker is not available

3. **Host fallback**: Uses `gopsutil` to query host available RAM
   - Used when neither Docker nor WSL is available

### Turbo Mode

Turbo mode multiplies the capacity roof for more aggressive parallelism:

```bash
eac build --turbo      # Uses default 4x multiplier
eac build --turbo=2    # Uses 2x multiplier
eac build --turbo=8    # Uses 8x multiplier
```

Use turbo mode when:
- Builds are I/O-bound (waiting on network, disk)
- Running lightweight containerized builds
- You want faster builds and have headroom

### Memory Instrumentation

Each build logs memory usage before and after execution:

```
[memory] before: used=2.89GB avail=2.89GB total=5.78GB (50.0%)
[memory] after: used=3.10GB avail=2.68GB total=5.78GB (53.6%) delta=+210MB
```

This data can be used to:
- Tune the 256MB slot size if needed
- Identify memory-hungry builds
- Validate that containerized builds have low memory impact

### Configuration

| Setting | Description | Default |
|---------|-------------|---------|
| `MaxConcurrency` | Hard ceiling (0 = dynamic) | 0 (dynamic) |
| `Turbo` | Capacity multiplier | 1 (normal) |

Example configuration:

```go
config := orchestrator.Config{
    MaxConcurrency: 0,  // Dynamic based on RAM
    Turbo:          4,  // 4x multiplier
}
```

### Files

| File | Purpose |
|------|---------|
| `memory_detection.go` | Docker/WSL/host memory detection |
| `component_scheduler.go` | Capacity calculation, LPT scheduling |

## LPT Scheduling Algorithm

The orchestrator uses **LPT (Longest Processing Time First)** scheduling to optimize job distribution and minimize total build time.

### Problem

Without intelligent scheduling, jobs might execute in arbitrary order, causing:
- All heavy jobs running at the end (poor parallelism)
- Light jobs finishing early, leaving heavy jobs queued
- Suboptimal resource utilization

### Solution

LPT sorts jobs by weight (processing time) in descending order before execution:

```
Before LPT: [docs●, core●, cli●, books●●●●, tests●●]
After LPT:  [books●●●●, tests●●, docs●, core●, cli●]
```

Heavy jobs start first, ensuring they run in parallel with lighter jobs rather than sequentially at the end.

### Algorithm

```go
// Sort by weight descending (heaviest first)
sort.Slice(work, func(i, j int) bool {
    return work[i].Weight > work[j].Weight
})
```

### Benefits

1. **Better parallelism**: Heavy jobs overlap with light jobs
2. **Shorter makespan**: Total build time is minimized
3. **Predictable behavior**: Heavy jobs always visible early in TUI
4. **4/3 approximation**: LPT guarantees makespan within 4/3 of optimal

### Example

With capacity 6 and jobs: `books(4)`, `docs(1)`, `core(1)`, `cli(1)`, `tests(2)`

**Without LPT** (arbitrary order):
```
Time 0-1: docs(1), core(1), cli(1), tests(2) = 5 slots
Time 1-2: books(4) waiting... then runs alone = 4 slots
Total: 2 time units, poor utilization
```

**With LPT** (heaviest first):
```
Time 0-1: books(4), tests(2) = 6 slots (full!)
Time 0-1: docs(1), core(1), cli(1) finish during books
Total: 1 time unit, optimal utilization
```

### TUI Visualization

The TUI displays job weights as dots in the tab bar:

```
│All ▶ books ●●●● ▶ tests ●● ▶ docs ● ▶ core ●              │
```

This shows at a glance which heavy jobs are running.

### References

- [LPT Scheduling - Wikipedia](https://en.wikipedia.org/wiki/Longest-processing-time-first_scheduling)
- [Multiprocessor Scheduling](https://en.wikipedia.org/wiki/Identical-machines_scheduling)

## Future Enhancements

Possible improvements:

- Real-time log streaming for verbose mode
- Progress bars for long-running operations
- Colorized output based on module status
- Dependency-aware scheduling (respect module dependencies)
- Component weight tuning based on instrumentation data
