# Real Progress Implementation - Ultra Enhancement Plan

## Executive Summary

This document presents a comprehensive enhancement roadmap for the VSCode Extension ↔ Go MCP Server progress notification system. The analysis identifies **10 core architectural limitations** and proposes **60+ enhancements** across 8 strategic categories.

**Current State:** Unidirectional progress notifications with 6 static stages
**Target State:** Bidirectional, hierarchical, cancellable, observable, production-grade progress system

---

## Architecture Analysis

### Current Strengths
✓ Clean JSON-RPC protocol over stdin/stdout
✓ Real-time streaming of progress notifications
✓ Graceful fallback to simulated progress
✓ Simple and understandable stage model

### Core Limitations

| # | Limitation | Impact | Severity |
|---|------------|--------|----------|
| 1 | **No bidirectional communication** | Cannot cancel, pause, or configure operations | High |
| 2 | **No progress quantification** | User can't estimate time remaining | Medium |
| 3 | **Single-threaded progress model** | Can't show parallel operations (e.g., multiple git commands) | Medium |
| 4 | **No error streaming** | Errors only surface at the end, poor UX for long operations | High |
| 5 | **No cancellation protocol** | User must kill process, leaves dirty state | High |
| 6 | **No progress persistence** | Lost on crash/restart, hard to debug | Medium |
| 7 | **Limited context in messages** | Generic messages don't show what's being processed | Low |
| 8 | **No sub-operation tracking** | Can't drill down into complex operations | Medium |
| 9 | **No timing metadata** | Can't predict ETAs or identify performance bottlenecks | Low |
| 10 | **Static stage model** | Stages hardcoded, can't adapt to different workflows | Medium |

---

## Enhancement Categories

### 1. Progress Tracking & Quantification

#### 1.1 Percentage-Based Progress

**Current:**
```json
{
  "method": "$/progress",
  "params": {
    "stage": "git",
    "message": "Gathering git context..."
  }
}
```

**Enhanced:**
```json
{
  "method": "$/progress",
  "params": {
    "stage": "git",
    "message": "Gathering git context...",
    "progress": {
      "current": 3,
      "total": 12,
      "percentage": 25.0,
      "unit": "operations"
    }
  }
}
```

**Benefits:**
- Accurate progress bar positioning
- User can estimate completion time
- Better UX for long-running operations

**Implementation:**
- Add progress tracking to each operation in Go
- Count total operations during planning phase
- Emit progress fraction with each update

#### 1.2 Hierarchical Progress (Parent/Child Operations)

**Concept:**
```
├─ Generating commit message (0/100%)
   ├─ Loading agent configuration (100%)
   ├─ Gathering git context (50%)
   │  ├─ git status (100%)
   │  ├─ git diff (100%)
   │  └─ git log (0%)
   └─ Reading documentation (0%)
```

**Protocol:**
```json
{
  "method": "$/progress",
  "params": {
    "id": "op-git-001",
    "parentId": "op-root",
    "stage": "git",
    "message": "Running git diff...",
    "progress": { "current": 2, "total": 3, "percentage": 66.6 }
  }
}
```

**VSCode Display:**
- Show current operation in progress message
- Show sub-operations in expandable tree view
- Calculate aggregate progress from hierarchy

**Use Cases:**
- Multi-file analysis
- Batch commit operations
- Complex git workflows (rebase, merge)

#### 1.3 Parallel Operation Tracking

**Scenario:** Processing multiple files concurrently

```json
[
  {
    "method": "$/progress",
    "params": {
      "id": "analyze-file-1",
      "message": "Analyzing src/main.go",
      "parallel": true,
      "slot": 1,
      "progress": { "current": 45, "total": 100 }
    }
  },
  {
    "method": "$/progress",
    "params": {
      "id": "analyze-file-2",
      "message": "Analyzing src/utils.go",
      "parallel": true,
      "slot": 2,
      "progress": { "current": 78, "total": 100 }
    }
  }
]
```

**VSCode Display:**
- Show multiple progress bars (up to 4 slots)
- Or show aggregated "3/10 files analyzed"
- Display parallelism level: "Processing 4 operations..."

#### 1.4 Dynamic Context in Messages

**Current:**
```
"Gathering git context..."
```

**Enhanced:**
```
"Gathering git context (12 modified files, 3 commits)..."
"Reading documentation (guide/vscode-extension/...)..."
"Generating commit message (estimated 30s)..."
```

**Implementation:**
- Inject runtime metadata into message templates
- Pass context data through progress params
- Format in extension for display

#### 1.5 Time Tracking & ETA Prediction

**Protocol Extension:**
```json
{
  "method": "$/progress",
  "params": {
    "stage": "claude",
    "message": "Generating commit message...",
    "timing": {
      "startedAt": "2025-10-24T10:30:00Z",
      "estimatedDuration": 45000,
      "estimatedCompletion": "2025-10-24T10:30:45Z"
    }
  }
}
```

**VSCode Display:**
```
"Generating commit message... (30s remaining)"
```

**ETA Algorithm:**
- Track historical operation durations
- Use rolling average for predictions
- Adjust dynamically as operation progresses
- Store in local cache: `~/.vscode-mcp/timing-history.json`

---

### 2. Bidirectional Communication Protocol

#### 2.1 Cancellation Support

**Extension → Server (stdin):**
```json
{
  "jsonrpc": "2.0",
  "method": "$/cancel",
  "params": {
    "operationId": "op-commit-123",
    "reason": "user_requested"
  }
}
```

**Server → Extension (stdout):**
```json
{
  "method": "$/progress",
  "params": {
    "stage": "cancelled",
    "message": "Operation cancelled by user",
    "metadata": {
      "cleanupPerformed": true,
      "partialResults": null
    }
  }
}
```

**Implementation in Go:**
- Use context.Context with cancellation
- Monitor stdin for cancel messages in goroutine
- Cancel context when message received
- Cleanup temp files, restore state
- Send cancellation confirmation

**VSCode UX:**
- Show "Cancel" button during progress
- Confirm cancellation for destructive operations
- Show cleanup progress after cancel

#### 2.2 Pause/Resume Support

**Use Case:** Long analysis, user needs to check something

**Protocol:**
```json
// Pause request
{
  "method": "$/pause",
  "params": { "operationId": "op-001" }
}

// Resume request
{
  "method": "$/resume",
  "params": { "operationId": "op-001" }
}

// Pause notification
{
  "method": "$/progress",
  "params": {
    "stage": "paused",
    "message": "Operation paused (can be resumed)",
    "pausedAt": "2025-10-24T10:30:15Z"
  }
}
```

**Implementation Challenges:**
- Cannot pause external processes (git, Claude API)
- Can pause between operations
- Save state to resume later

#### 2.3 Interactive Prompts

**Scenario:** Conflict resolution, user input needed

**Server → Extension:**
```json
{
  "method": "$/prompt",
  "params": {
    "promptId": "prompt-001",
    "type": "choice",
    "message": "Conflict detected in main.go. How to proceed?",
    "choices": [
      { "id": "manual", "label": "Resolve manually" },
      { "id": "skip", "label": "Skip this file" },
      { "id": "abort", "label": "Abort operation" }
    ]
  }
}
```

**Extension → Server:**
```json
{
  "method": "$/promptResponse",
  "params": {
    "promptId": "prompt-001",
    "choiceId": "manual"
  }
}
```

**VSCode UX:**
- Show modal dialog or quick pick
- Block progress until user responds
- Timeout after 60s with default choice

#### 2.4 Configuration Updates During Execution

**Use Case:** User wants to adjust verbosity mid-operation

```json
{
  "method": "$/configure",
  "params": {
    "operationId": "op-001",
    "config": {
      "verbosity": "debug",
      "showTimings": true
    }
  }
}
```

---

### 3. Error Handling & Resilience

#### 3.1 Real-Time Error Streaming

**Current:** Errors only in final response

**Enhanced:**
```json
{
  "method": "$/error",
  "params": {
    "severity": "warning",
    "stage": "git",
    "message": "Could not read .gitignore (file not found)",
    "code": "ERR_FILE_NOT_FOUND",
    "recoverable": true,
    "timestamp": "2025-10-24T10:30:12Z"
  }
}
```

**Severity Levels:**
- `debug` - Development info
- `info` - Informational messages
- `warning` - Non-fatal issues
- `error` - Fatal errors (stops operation)
- `critical` - System-level failures

**VSCode Display:**
- Show warnings in progress subtitle
- Show errors in notification toast
- Collect all errors in output channel
- Critical errors stop progress immediately

#### 3.2 Recoverable Error Handling

**Scenario:** API rate limit hit

```json
{
  "method": "$/error",
  "params": {
    "severity": "warning",
    "message": "Claude API rate limit (retry in 5s)",
    "code": "ERR_RATE_LIMIT",
    "recoverable": true,
    "retryStrategy": {
      "type": "exponential_backoff",
      "retryIn": 5000,
      "maxRetries": 3,
      "currentRetry": 1
    }
  }
}
```

**VSCode UX:**
```
"Generating commit message... (rate limited, retrying in 5s)"
```

#### 3.3 Partial Failure Handling

**Scenario:** Batch commit, 2/5 succeed

```json
{
  "method": "$/progress",
  "params": {
    "stage": "complete",
    "message": "Completed with partial failures",
    "results": {
      "total": 5,
      "succeeded": 2,
      "failed": 3,
      "failures": [
        {
          "operation": "commit-repo-3",
          "error": "Authentication failed",
          "code": "ERR_AUTH"
        }
      ]
    }
  }
}
```

**VSCode UX:**
- Show partial success notification
- Offer to retry failures
- Show detailed failure report

#### 3.4 Validation Errors (Pre-flight)

**Concept:** Validate before starting expensive operation

```json
{
  "method": "$/progress",
  "params": {
    "stage": "validation",
    "message": "Validating prerequisites...",
    "validations": [
      { "check": "git_installed", "passed": true },
      { "check": "git_repo", "passed": true },
      { "check": "uncommitted_changes", "passed": false, "error": "No changes to commit" }
    ]
  }
}
```

**Benefits:**
- Fail fast
- Clear error messages
- Prevent wasted API calls

---

### 4. User Experience Enhancements

#### 4.1 Multi-Operation Progress

**Use Case:** Batch commit multiple repositories

**VSCode UI:**
```
Overall Progress: 3/10 repositories
Current: repo-alpha (Gathering git context...)
├─ repo-alpha: Generating commit...
├─ repo-beta: Complete ✓
├─ repo-gamma: Complete ✓
└─ repo-delta: Pending...
```

**Protocol:**
```json
{
  "method": "$/progress",
  "params": {
    "batchId": "batch-001",
    "batchProgress": { "current": 3, "total": 10 },
    "currentOperation": {
      "id": "repo-alpha",
      "stage": "git",
      "message": "Gathering git context..."
    }
  }
}
```

#### 4.2 Operation History & Replay

**Concept:** Store completed operations for review/replay

**Storage:** `~/.vscode-mcp/history/2025-10-24-10-30-15.json`

```json
{
  "operationId": "op-commit-123",
  "startedAt": "2025-10-24T10:30:00Z",
  "completedAt": "2025-10-24T10:31:05Z",
  "duration": 65000,
  "stages": [
    { "stage": "init", "duration": 100, "timestamp": "..." },
    { "stage": "git", "duration": 2500, "timestamp": "..." },
    { "stage": "claude", "duration": 60000, "timestamp": "..." }
  ],
  "result": "success",
  "output": "feat: add user authentication..."
}
```

**VSCode Commands:**
- `MCP: View Operation History`
- `MCP: Replay Last Operation`
- `MCP: Show Operation Timeline`

#### 4.3 System Notifications

**Concept:** Notify user when operation completes in background

```typescript
vscode.window.showInformationNotification(
  'Commit message generated!',
  'View', 'Copy'
).then(selection => {
  if (selection === 'View') {
    // Show result
  } else if (selection === 'Copy') {
    vscode.env.clipboard.writeText(result);
  }
});
```

**User Setting:**
```json
{
  "mcp.notifications.enabled": true,
  "mcp.notifications.minDuration": 10000  // Only notify if >10s
}
```

#### 4.4 Detailed Logs on Demand

**VSCode UI:**
- Progress notification shows summary
- "Show Details" button opens output channel
- Output channel shows full logs in real-time

**Output Channel:**
```
[10:30:00] Starting commit message generation
[10:30:00] Loading agent: vscode-extension-commit-button.md
[10:30:00] Running: git status
[10:30:01] Found 12 modified files
[10:30:01] Running: git diff
[10:30:02] Diff size: 4523 bytes
[10:30:02] Reading docs: guide/vscode-extension/USAGE.md
[10:30:45] Claude API response received (42 tokens/s)
[10:31:05] Complete! Generated 234 char message
```

**Protocol:**
```json
{
  "method": "$/log",
  "params": {
    "level": "debug",
    "message": "Running: git status",
    "timestamp": "2025-10-24T10:30:00Z",
    "metadata": { "command": "git status" }
  }
}
```

#### 4.5 Dry-Run Mode

**Use Case:** Preview what will happen without executing

**Protocol:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "commit_message_generator",
    "arguments": {
      "dryRun": true
    }
  }
}
```

**Server Response:**
```json
{
  "method": "$/progress",
  "params": {
    "stage": "dry_run",
    "message": "Dry-run complete",
    "preview": {
      "filesAnalyzed": 12,
      "estimatedTokens": 450,
      "estimatedCost": "$0.002",
      "estimatedDuration": "45s"
    }
  }
}
```

---

### 5. Observability & Debugging

#### 5.1 Structured Logging

**Log File:** `~/.vscode-mcp/logs/2025-10-24.jsonl`

**Format (JSONL):**
```json
{"timestamp":"2025-10-24T10:30:00Z","level":"info","stage":"init","message":"Starting operation","operationId":"op-123"}
{"timestamp":"2025-10-24T10:30:01Z","level":"debug","stage":"git","message":"Executing git command","command":"git status"}
{"timestamp":"2025-10-24T10:30:02Z","level":"info","stage":"git","message":"Git status complete","filesChanged":12,"duration":1500}
```

**Benefits:**
- Easy to parse for analysis
- Can pipe to log aggregators
- Machine-readable for debugging tools

#### 5.2 Performance Metrics

**Protocol:**
```json
{
  "method": "$/metrics",
  "params": {
    "operationId": "op-123",
    "metrics": {
      "totalDuration": 65000,
      "stageBreakdown": {
        "init": 100,
        "git": 2500,
        "docs": 1500,
        "prompt": 500,
        "claude": 60000,
        "complete": 400
      },
      "resourceUsage": {
        "cpuPercent": 5.2,
        "memoryMB": 45,
        "diskReadMB": 2.3,
        "diskWriteMB": 0.1
      },
      "apiCalls": {
        "claude": {
          "count": 1,
          "duration": 60000,
          "tokensIn": 1250,
          "tokensOut": 234,
          "cost": 0.0023
        }
      }
    }
  }
}
```

**VSCode Command:** `MCP: Show Performance Report`

**Performance Dashboard:**
```
Last 10 Operations:
┌─────────────────────────┬──────────┬─────────┬────────┐
│ Timestamp               │ Duration │ Tokens  │ Cost   │
├─────────────────────────┼──────────┼─────────┼────────┤
│ 2025-10-24 10:30:00     │ 65s      │ 1484    │ $0.002 │
│ 2025-10-24 09:15:22     │ 42s      │ 892     │ $0.001 │
│ 2025-10-23 16:45:10     │ 78s      │ 2103    │ $0.004 │
└─────────────────────────┴──────────┴─────────┴────────┘

Average duration: 61.6s
Total API cost today: $0.007
```

#### 5.3 Operation Traces (OpenTelemetry)

**Concept:** Instrument with distributed tracing

```go
import "go.opentelemetry.io/otel"

func generateCommitMessage(ctx context.Context) {
    ctx, span := tracer.Start(ctx, "generateCommitMessage")
    defer span.End()

    // Child spans
    _, gitSpan := tracer.Start(ctx, "gatherGitContext")
    // ... do git operations
    gitSpan.End()

    _, claudeSpan := tracer.Start(ctx, "callClaudeAPI")
    // ... call API
    claudeSpan.End()
}
```

**Export:** Send traces to Jaeger/Zipkin for visualization

**Benefits:**
- Visualize operation flow
- Identify bottlenecks
- Compare performance across runs

#### 5.4 Debug Mode

**Enable:**
```json
{
  "mcp.debug.enabled": true,
  "mcp.debug.verbosity": "trace",
  "mcp.debug.saveStdio": true
}
```

**Behavior:**
- Save all stdin/stdout to debug log
- Show all JSON-RPC messages in output channel
- Include stack traces in errors
- Disable timeouts for debugging
- Save intermediate artifacts

**Debug Log:** `~/.vscode-mcp/debug/2025-10-24-10-30-15/`
```
stdin.jsonl      # All messages sent to server
stdout.jsonl     # All messages from server
stderr.log       # Error output
artifacts/       # Temp files, prompts, responses
```

#### 5.5 Progress Snapshots

**Concept:** Serialize entire operation state for replay

**Snapshot File:** `~/.vscode-mcp/snapshots/op-123-snapshot.json`

```json
{
  "operationId": "op-123",
  "snapshotTime": "2025-10-24T10:30:30Z",
  "currentStage": "claude",
  "progress": { "current": 5, "total": 6 },
  "context": {
    "workingDirectory": "/path/to/repo",
    "gitStatus": { /* ... */ },
    "modifiedFiles": [ /* ... */ ]
  },
  "history": [
    { "stage": "init", "completed": true, "duration": 100 },
    { "stage": "git", "completed": true, "duration": 2500 }
  ]
}
```

**Use Cases:**
- Crash recovery: Resume from snapshot
- Debugging: Replay from specific point
- Testing: Use snapshots as test fixtures

---

### 6. Performance Optimizations

#### 6.1 Progressive Diff Rendering

**Problem:** Large diffs (>1MB) slow down Claude API

**Solution:**
- Stream diff in chunks
- Summarize large hunks
- Paginate file-by-file

```go
// Instead of:
diff := runGitDiff() // 5MB

// Do:
for file := range modifiedFiles {
    fileDiff := runGitDiff(file)
    if len(fileDiff) > 100KB {
        fileDiff = summarizeDiff(fileDiff) // Just show function signatures
    }
    sendProgress("git", fmt.Sprintf("Analyzing %s...", file))
    aggregateDiff = append(aggregateDiff, fileDiff)
}
```

#### 6.2 Caching Git Operations

**Cache:** `~/.vscode-mcp/cache/git-cache.json`

```json
{
  "repo": "/path/to/repo",
  "lastCommit": "abc123",
  "cachedAt": "2025-10-24T10:25:00Z",
  "ttl": 60000,
  "data": {
    "recentCommits": [ /* ... */ ],
    "branchName": "main",
    "remoteUrl": "https://github.com/..."
  }
}
```

**Strategy:**
- Cache `git log` results (TTL 1min)
- Cache branch name until it changes
- Invalidate on commit/checkout

**Performance Gain:** -500ms on repeated operations

#### 6.3 Incremental Processing

**Use Case:** User commits frequently in same session

**Track:**
- Last processed commit SHA
- Only diff since last commit
- Reuse context from previous run

```go
var lastProcessedCommit string

func getChanges() Diff {
    if lastProcessedCommit != "" {
        return git.DiffRange(lastProcessedCommit, "HEAD")
    }
    return git.Diff() // Full diff
}
```

#### 6.4 Lazy Documentation Loading

**Current:** Load all docs upfront (slow)

**Optimized:**
- Load docs index first
- Only load relevant docs based on file types
- Stream doc content as needed

```go
// Phase 1: Quick index (100ms)
docIndex := loadDocIndex() // Just filenames + metadata

// Phase 2: Filter relevance (50ms)
relevantDocs := filterByFileTypes(docIndex, modifiedFiles)

// Phase 3: Load only relevant (500ms instead of 2000ms)
for doc := range relevantDocs {
    sendProgress("docs", fmt.Sprintf("Reading %s...", doc))
    content := loadDoc(doc)
}
```

#### 6.5 Streaming Claude API Responses

**Current:** Wait for full response

**Enhanced:** Stream tokens as they arrive

**Protocol:**
```json
{
  "method": "$/progress",
  "params": {
    "stage": "claude",
    "message": "Generating commit message...",
    "stream": {
      "type": "token",
      "tokens": "feat: add user authentication\n\nImplement"
    }
  }
}
```

**VSCode Display:**
- Show partial message in preview pane
- Update in real-time as tokens arrive
- User can see quality immediately

---

### 7. Advanced Features

#### 7.1 Concurrent Multi-Repository Support

**Use Case:** Monorepo with multiple independent modules

**VSCode UI:**
```
╔═══════════════════════════════════════════════╗
║  Processing 3 repositories concurrently       ║
╠═══════════════════════════════════════════════╣
║  [████████░░] repo-frontend (80%)             ║
║  [██████████] repo-backend (100%) ✓           ║
║  [███░░░░░░░] repo-shared (30%)               ║
╚═══════════════════════════════════════════════╝
```

**Implementation:**
- Launch multiple Go processes (one per repo)
- Multiplex progress notifications
- Aggregate results
- Show unified view

#### 7.2 Smart Batching

**Problem:** 50 small commits vs 1 large commit

**Solution:** Intelligently batch related changes

```go
func detectBatches(changes []FileChange) []Batch {
    batches := []Batch{}

    // Group by:
    // 1. Same directory
    // 2. Same type (feature/fix/docs)
    // 3. Time proximity (<5min)

    for change := range changes {
        if batch := findRelatedBatch(batches, change) {
            batch.Add(change)
        } else {
            batches = append(batches, NewBatch(change))
        }
    }

    return batches
}
```

**User Prompt:**
```
Detected 3 logical change groups:
1. Frontend updates (12 files)
2. Backend API changes (5 files)
3. Documentation (2 files)

Create 3 separate commits? [Yes/No/Custom]
```

#### 7.3 Webhook Notifications

**Use Case:** Notify team channel when commits happen

**Configuration:**
```json
{
  "mcp.webhooks": [
    {
      "url": "https://hooks.slack.com/...",
      "events": ["commit.created", "commit.failed"],
      "template": {
        "text": "Commit created: {{message}}"
      }
    }
  ]
}
```

**Triggered on completion:**
```json
POST https://hooks.slack.com/...
{
  "text": "Commit created: feat: add user authentication",
  "author": "developer@example.com",
  "timestamp": "2025-10-24T10:31:05Z",
  "repository": "my-project"
}
```

#### 7.4 CI/CD Integration

**Use Case:** Generate commits via CI pipeline

**CLI Interface:**
```bash
vscode-mcp commit --agent vscode-extension-commit-button \
  --workspace /path/to/repo \
  --output commit-message.txt \
  --format json \
  --progress-file progress.jsonl
```

**Progress File (Live Updated):**
```jsonl
{"stage":"init","message":"Starting...","timestamp":"2025-10-24T10:30:00Z"}
{"stage":"git","message":"Gathering context...","progress":{"current":1,"total":6}}
{"stage":"complete","message":"Done","result":"success"}
```

**GitHub Actions Example:**
```yaml
- name: Generate commit message
  run: |
    vscode-mcp commit --agent my-agent --output message.txt
    git commit -F message.txt
    git push
```

#### 7.5 Custom Stage Injection

**Use Case:** Add project-specific validation stages

**Configuration:**
```json
{
  "mcp.stages.custom": [
    {
      "stage": "lint",
      "position": "before:claude",
      "command": "npm run lint",
      "required": true
    },
    {
      "stage": "test",
      "position": "after:claude",
      "command": "npm test",
      "required": false
    }
  ]
}
```

**Execution Flow:**
```
1. init
2. git
3. docs
4. prompt
5. lint (custom, required)
6. claude
7. test (custom, optional)
8. complete
```

**Progress:**
```json
{
  "method": "$/progress",
  "params": {
    "stage": "lint",
    "message": "Running linter...",
    "custom": true,
    "command": "npm run lint",
    "output": "✓ All files pass linting"
  }
}
```

---

### 8. Developer Experience

#### 8.1 Progress Testing Framework

**Mock Progress Server:**
```go
// test_server.go
type MockProgressServer struct {
    stages []ProgressStage
}

func (m *MockProgressServer) SimulateProgress() {
    for _, stage := range m.stages {
        sendProgress(stage.Name, stage.Message)
        time.Sleep(stage.Duration)
    }
}

// Use in tests:
func TestProgressDisplay(t *testing.T) {
    server := MockProgressServer{
        stages: []ProgressStage{
            {Name: "init", Message: "Starting...", Duration: 100*time.Millisecond},
            {Name: "git", Message: "Git...", Duration: 500*time.Millisecond},
        },
    }
    server.SimulateProgress()
    // Assert VSCode extension displays correctly
}
```

#### 8.2 Progress Visualization Tool

**Web UI:** `http://localhost:8080/progress-visualizer`

**Features:**
- Upload progress JSONL file
- Visualize timeline
- Show stage durations
- Identify bottlenecks
- Compare multiple runs

**Timeline Visualization:**
```
0s    10s   20s   30s   40s   50s   60s   70s
|─────|─────|─────|─────|─────|─────|─────|
[init][git   ][docs][prompt][claude................][done]
      2.5s   1.5s  0.5s     60s
```

#### 8.3 Schema Validation

**JSON Schema for Progress Protocol:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Progress Notification",
  "type": "object",
  "required": ["jsonrpc", "method", "params"],
  "properties": {
    "jsonrpc": { "const": "2.0" },
    "method": { "const": "$/progress" },
    "params": {
      "type": "object",
      "required": ["stage", "message"],
      "properties": {
        "stage": { "type": "string" },
        "message": { "type": "string" },
        "progress": {
          "type": "object",
          "properties": {
            "current": { "type": "number" },
            "total": { "type": "number" },
            "percentage": { "type": "number", "minimum": 0, "maximum": 100 }
          }
        }
      }
    }
  }
}
```

**Usage:**
- Validate in Go before sending
- Validate in TypeScript on receive
- Catch protocol violations early

#### 8.4 TypeScript Protocol Types

**Generate from schema:**
```typescript
// protocol.ts (auto-generated)
export interface ProgressNotification {
  jsonrpc: "2.0";
  method: "$/progress";
  params: ProgressParams;
}

export interface ProgressParams {
  stage: string;
  message: string;
  progress?: ProgressInfo;
  timing?: TimingInfo;
  metadata?: Record<string, any>;
}

export interface ProgressInfo {
  current: number;
  total: number;
  percentage?: number;
  unit?: string;
}

// Use in extension:
function handleProgress(notification: ProgressNotification) {
  // TypeScript knows the shape!
  const { stage, message, progress } = notification.params;
}
```

---

## Implementation Roadmap

### Phase 1: Foundation (Week 1-2)
**Goal:** Bidirectional communication + quantified progress

1. Add percentage-based progress
2. Implement cancellation protocol
3. Add real-time error streaming
4. Structured logging to file

**Deliverables:**
- Go: Progress quantification, cancellation support
- Extension: Cancel button, progress percentages
- Protocol v2 with progress fractions

**Success Metrics:**
- User can cancel long operations
- Progress bar shows accurate percentage
- Errors appear in real-time

### Phase 2: Observability (Week 3-4)
**Goal:** Debugging and performance insights

5. Performance metrics collection
6. Operation history storage
7. Debug mode implementation
8. Progress visualization tool

**Deliverables:**
- Metrics dashboard in VSCode
- Operation history viewer
- Debug log capture
- Standalone visualizer

**Success Metrics:**
- Can replay past operations
- Can identify performance bottlenecks
- Debug logs help troubleshoot issues

### Phase 3: UX Polish (Week 5-6)
**Goal:** Superior user experience

9. Hierarchical progress display
10. System notifications
11. Detailed logs on demand
12. Time estimates & ETAs

**Deliverables:**
- Tree view for sub-operations
- Background completion notifications
- Expandable log viewer
- ETA predictions

**Success Metrics:**
- Users understand what's happening
- Can work on other tasks during long operations
- Accurate time estimates

### Phase 4: Performance (Week 7-8)
**Goal:** Faster operations

13. Caching git operations
14. Lazy documentation loading
15. Progressive diff rendering
16. Streaming Claude responses

**Deliverables:**
- 50% faster repeated operations
- Reduced memory usage
- Real-time response preview

**Success Metrics:**
- <5s for cached git operations
- <30s average total time
- Smooth progress updates

### Phase 5: Advanced Features (Week 9-12)
**Goal:** Power user capabilities

17. Multi-repository support
18. Smart batching
19. Custom stage injection
20. CLI interface for CI/CD

**Deliverables:**
- Batch commit UI
- Custom pipeline stages
- CI/CD integration guide

**Success Metrics:**
- Can process multiple repos
- Integrates with build pipelines
- Extensible for custom workflows

---

## Risk Analysis

### Technical Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **Stdin/stdout buffering issues** | High | Medium | Use line-buffered mode, flush after each message |
| **Protocol version compatibility** | High | Medium | Version negotiation handshake, backward compatibility |
| **Performance overhead from logging** | Medium | Medium | Async logging, configurable verbosity |
| **Race conditions in cancellation** | High | Low | Proper context propagation, mutex protection |
| **Large progress state memory usage** | Medium | Low | Streaming, pruning old history |

### UX Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **Information overload** | Medium | High | Progressive disclosure, summary view by default |
| **Notification fatigue** | Medium | Medium | Smart defaults, user configuration |
| **Progress inaccuracy** | Medium | Medium | Conservative estimates, dynamic adjustment |

---

## Success Metrics

### Performance Metrics
- Average operation time: **<30s** (currently ~60s)
- P95 operation time: **<60s** (currently ~90s)
- Cache hit rate: **>70%** for repeated operations
- Memory usage: **<100MB** per operation

### User Experience Metrics
- Cancellation response time: **<500ms**
- Progress accuracy: **±10%** of actual completion
- Error visibility: **<1s** from occurrence to display

### Reliability Metrics
- Operation success rate: **>95%**
- Recovery from failures: **100%** (clean state)
- Crash recovery: **100%** (resume from snapshot)

---

## Open Questions

1. **Protocol Versioning:** How to handle breaking changes in JSON-RPC protocol?
   - Proposal: Handshake with version negotiation
   - Fallback to lowest common version

2. **Concurrency Model:** How many parallel operations to allow?
   - Proposal: Configurable limit (default 4)
   - Queue excess operations

3. **Storage Limits:** How much history/logs to retain?
   - Proposal: 30 days or 100MB, whichever comes first
   - User-configurable

4. **Security:** Should we encrypt sensitive data in logs?
   - Proposal: Redact credentials, API keys by default
   - Optional full encryption

5. **Cross-platform:** How to handle Windows vs Unix differences?
   - Proposal: Abstract file paths, use Go stdlib
   - Test on all platforms

---

## Appendix

### A. Complete Protocol Specification

See: `/out/mcp-progress-protocol-v2-spec.md` (to be created)

### B. Performance Benchmarks

See: `/out/progress-performance-benchmarks.md` (to be created)

### C. Migration Guide

See: `/out/progress-v1-to-v2-migration.md` (to be created)

---

## Conclusion

This ultra enhancement plan transforms the progress notification system from a basic status display into a **production-grade, observable, interactive operation management system**.

**Key Improvements:**
- **10x better UX** - Cancellation, accurate progress, real-time feedback
- **5x easier debugging** - Structured logs, metrics, traces
- **2x faster** - Caching, streaming, optimization
- **∞ more extensible** - Custom stages, webhooks, CI/CD

**Next Steps:**
1. Review and prioritize enhancements with stakeholders
2. Implement Phase 1 (Foundation) as MVP
3. Gather user feedback
4. Iterate on Phases 2-5

**Estimated Total Effort:** 10-12 weeks for full implementation
**Recommended Team:** 2 engineers (1 Go, 1 TypeScript)

---

*Document Version: 1.0*
*Last Updated: 2025-10-24*
*Author: Claude (Sonnet 4.5)*
