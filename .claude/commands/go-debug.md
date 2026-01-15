# Debug

```text
description: "Debug Go code, test failures, or runtime issues"
```

You are debugging Go code, test failures, or runtime issues.

## Process

1. **Understand the problem**:
   - What is the error or unexpected behavior?
   - Can it be reproduced reliably?
   - Get error messages, stack traces, or test output

2. **Investigate**:
   - Delegate to go-debugger agent
   - Use Task tool with subagent_type="go-debugger"
   - Provide all error information for analysis

3. **For test failures**:
   - Run with verbose output: `go test -v ./path`
   - Check for race conditions: `go test -race`
   - Isolate failing test: `go test -run TestName`

4. **For runtime issues**:
   - Check logs for error context
   - Verify goroutine cleanup (context cancellation)
   - Profile if performance-related: `go test -cpuprofile`

5. **Propose fix**:
   - Explain root cause
   - Provide code fix
   - Add regression test
   - Verify fix with tests

## Debugging Tools

- `go test -v`: Verbose test output
- `go test -race`: Race detector
- `go test -run TestName`: Run specific test
- `go test -count=N`: Run test N times (catch flaky tests)
- `dlv debug`: Interactive debugger (delve)

## Example Usage

- `/go:debug failing test output: [paste output]`
- `/go:debug panic in CLI command execution`
- `/go:debug why is this function running slowly?`
