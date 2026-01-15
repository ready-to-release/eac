# Test

```text
description: "Write or debug tests for Go code"
```

You are writing or debugging tests for Go code.

## Process

1. **Understand the context**:
   - Is this writing new tests (TDD) or debugging existing failures?
   - What code needs testing?

2. **For writing new tests**:
   - Delegate to go-test-engineer agent
   - Use Task tool with subagent_type="go-test-engineer"
   - Request table-driven tests covering edge cases

3. **For debugging test failures**:
   - Delegate to go-debugger agent
   - Use Task tool with subagent_type="go-debugger"
   - Provide test failure output for analysis

4. **Run tests and verify**:
   - Local: `go test ./...` or `go test -v ./path/to/package`
   - Race detector: `go test -race ./...`
   - Coverage: `go test -cover ./...`
   - MCP: `test <module-name>`

5. **Test best practices**:
   - Tests must be deterministic and fast
   - No external I/O in unit tests
   - Use interfaces/mocks for dependencies
   - Subtests with t.Run() for table-driven tests

## Example Usage

- `/go:test write tests for the config validation logic`
- `/go:test debug why TestParseCommand is failing`
