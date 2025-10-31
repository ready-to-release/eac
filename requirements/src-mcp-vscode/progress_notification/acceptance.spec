# Progress Notification

> **Feature ID**: src_mcp_vscode_progress_notification
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: src-mcp-vscode
> **Tags**: mcp, vscode, progress

## User Story

* As a VSCode extension user
* I want to see real-time progress updates during commit generation
* So that I understand what stage the process is at and how long it's taking

## Acceptance Criteria

* Sends progress notifications via $/progress method
* Includes stage identifier in notification
* Includes descriptive message in notification
* Includes elapsed time in [Xs] format
* Sends notifications for each pipeline stage
* Sends completion notification when done
* Uses JSON-RPC 2.0 format for notifications
* Does not wait for response (notification, not request)

## Acceptance Tests

### AC1: Sends progress notifications via $/progress method
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, progress, jsonrpc

* Execute operation that triggers progress notifications
* Capture JSON-RPC messages sent to stdout
* Verify method field is "$/progress"
* Verify message follows JSON-RPC 2.0 format
* Verify no "id" field (notification, not request)

### AC2: Includes stage identifier in notification
**Validated by**: behavior.feature -> @ac2 scenarios

Tags: critical, progress

* Send progress notification for stage "init"
* Verify params object includes stage field
* Verify stage value is "init"
* Repeat for other stages: git, docs, gen-claude, etc.

### AC3: Includes descriptive message in notification
**Validated by**: behavior.feature -> @ac3 scenarios

Tags: progress

* Send progress notification for stage "git"
* Verify params object includes message field
* Verify message is human-readable
* Example: "Gathering git context..."
* Verify message describes current activity

### AC4: Includes elapsed time in [Xs] format
**Validated by**: behavior.feature -> @ac4 scenarios

Tags: progress, timing

* Start progress tracking
* Wait 2 seconds
* Send progress notification
* Verify message includes time like "[2s]"
* Verify time format matches pattern \[\d+s\]

### AC5: Sends notifications for each pipeline stage
**Validated by**: behavior.feature -> @ac5 scenarios

Tags: critical, progress, pipeline

* Execute full semantic commit generation
* Verify notification sent for "Loading generator agent"
* Verify notification sent for "Gathering git context"
* Verify notification sent for "Reading documentation"
* Verify notification sent for "Generating initial commit message"
* Verify notification sent for "Validating file completeness"
* Verify notification sent for "Reviewing commit message"
* Verify notification sent for "Approving commit message"
* Verify notification sent for "Generating commit title"
* Verify notification sent for "Validating commit message"

### AC6: Sends completion notification when done
**Validated by**: behavior.feature -> @ac6 scenarios

Tags: critical, progress

* Execute commit generation to completion
* Verify final notification indicates completion
* Verify completion message is distinct from progress messages
* Example: "Commit generation complete [5s]"

### AC7: Uses JSON-RPC 2.0 format for notifications
**Validated by**: behavior.feature -> @ac7 scenarios

Tags: critical, jsonrpc

* Send progress notification
* Verify "jsonrpc" field equals "2.0"
* Verify "method" field equals "$/progress"
* Verify "params" object is present
* Verify no "id" field is present
* Verify message is valid JSON

### AC8: Does not wait for response (fire-and-forget)
**Validated by**: behavior.feature -> @ac8 scenarios

Tags: progress, performance

* Send progress notification
* Verify execution continues immediately
* Verify no response is expected
* Verify notification is one-way communication
* Verify pipeline does not block on notification
