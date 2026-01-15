# Session End

```text
description: "End-of-session cleanup and code simplification"
```

You are completing a Claude Code session and running mandatory cleanup.

**IMPORTANT**: This command MUST be run at the end of EVERY Claude Code session.

## Process

1. **MANDATORY: Run code-simplifier plugin**:
   - Use the Task tool with subagent_type="code-simplifier"
   - This MUST be run at the end of EVERY session
   - Review all suggested changes carefully
   - Apply changes that improve clarity and simplicity

2. **Run final tests**:
   - `go test ./...` to verify nothing broken
   - Fix any test failures before ending session

3. **Run lint checks** (if golangci-lint available):
   - `golangci-lint run`
   - Address any critical issues

4. **Verify build** (if code was modified):
   - For affected modules, run: MCP `build <module>` or `go run ./go/eac/commands build <module>`
   - Ensure clean build with no errors

5. **Provide session summary**:
   - What was accomplished
   - Files modified
   - Code-simplifier changes applied
   - Tests status (all passing)
   - Any remaining TODOs

## Output Format

```
Session Summary:
================

Completed:
- [List accomplishments]

Files Modified:
- [List files]

Code Simplification:
- ✅ code-simplifier ran successfully
- [List simplifications applied]

Tests:
- ✅ All tests passing

Build:
- ✅ Clean build

Lint:
- ✅ No critical issues

Ready for commit: [Yes/No]

Remaining TODOs (if any):
- [List items]
```

## Why This Matters

Running `/go:session-end` ensures:

- Code stays simple and maintainable
- No broken tests left behind
- Clean state for next session
- Follows project quality standards

## Example Usage

User: `/go:session-end`

This is the last command you should run before closing Claude Code.
