# Error Handling

> **Feature ID**: src_mcp_vscode_error_handling
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: src-mcp-vscode
> **Tags**: mcp, vscode, error, critical

## User Story

* As a VSCode extension user
* I want clear, actionable error messages when operations fail
* So that I understand what went wrong and how to fix it

## Acceptance Criteria

* Handles git command failures gracefully
* Handles Claude CLI execution errors
* Handles file system errors (missing files, permissions)
* Handles JSON parsing errors
* Handles validation failures with clear messages
* Returns errors in JSON-RPC error format
* Includes error code, message, and details
* Never crashes or panics
* Logs errors for debugging
* Provides actionable error messages

## Acceptance Tests

### AC1: Handles git command failures gracefully
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, error, git

* Execute git command that fails (e.g., not a git repo)
* Catch execution error
* Return JSON-RPC error response
* Include descriptive error message
* Do not crash or panic
* Log error details

### AC2: Handles Claude CLI execution errors
**Validated by**: behavior.feature -> @ac2 scenarios

Tags: critical, error, agents

* Execute Claude CLI command that fails
* Catch execution error
* Return error in tool result
* Error message starts with "ERROR:"
* Include details about what failed
* Suggest corrective action if possible

### AC3: Handles file system errors
**Validated by**: behavior.feature -> @ac3 scenarios

Tags: error, filesystem

* Attempt to read non-existent file
* Catch file not found error
* Return descriptive error message
* Include file path in error
* Handle permission denied errors
* Handle directory not found errors

### AC4: Handles JSON parsing errors
**Validated by**: behavior.feature -> @ac4 scenarios

Tags: critical, error, jsonrpc

* Receive malformed JSON via stdin
* Catch JSON parse error
* Return JSON-RPC parse error (-32700)
* Error message is "Parse error"
* Continue listening for next request
* Do not crash

### AC5: Handles validation failures with clear messages
**Validated by**: behavior.feature -> @ac5 scenarios

Tags: error, validation

* Generate commit with validation errors
* Collect validation errors
* Format errors as readable message
* Include specific line numbers
* Include error codes (MD041, etc.)
* Explain what is wrong
* Suggest how to fix

### AC6: Returns errors in JSON-RPC error format
**Validated by**: behavior.feature -> @ac6 scenarios

Tags: critical, error, jsonrpc

* Encounter error during request handling
* Build JSON-RPC error response
* Include "error" object in response
* Include "code" field (integer)
* Include "message" field (string)
* Include "data" field with details (optional)
* Follow JSON-RPC 2.0 error spec

### AC7: Includes error code, message, and details
**Validated by**: behavior.feature -> @ac7 scenarios

Tags: error, jsonrpc

* Return error response
* Verify error object has "code" field
* Verify error object has "message" field
* Verify error object has "data" field with additional details
* Error code follows JSON-RPC spec
* Message is human-readable

### AC8: Never crashes or panics
**Validated by**: behavior.feature -> @ac8 scenarios

Tags: critical, error, reliability

* Trigger various error conditions
* Verify process continues running
* Verify no panics occur
* Verify server remains responsive
* Verify subsequent requests are handled
* Verify graceful degradation

### AC9: Logs errors for debugging
**Validated by**: behavior.feature -> @ac9 scenarios

Tags: error, debugging

* Encounter error during operation
* Verify error is logged to stderr
* Log includes timestamp
* Log includes error details
* Log includes stack trace if applicable
* Logs help developers diagnose issues

### AC10: Provides actionable error messages
**Validated by**: behavior.feature -> @ac10 scenarios

Tags: critical, error, usability

* Generate various error messages
* Verify each message describes the problem
* Verify each message suggests corrective action
* Examples: "Not a git repository. Run 'git init' or use a git repository."
* Messages are clear and helpful
* Users can resolve issues based on messages
