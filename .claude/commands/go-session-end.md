# Session End

```text
description: "End-of-session cleanup and code simplification"
```

You are completing a Claude Code session and running mandatory cleanup.

**IMPORTANT**: This command MUST be run at the end of EVERY Claude Code session.

## Process

1. **MANDATORY: Run code-simplifier agent**:
   - Use Task tool with subagent_type="code-simplifier:code-simplifier"
   - This MUST be run at the end of EVERY session
   - Review all suggested changes carefully
   - Apply changes that improve clarity and simplicity

   **Note**: The code-simplifier is a Claude Code agent that analyzes code and suggests simplifications for clarity and maintainability

2. **Run final tests**:
   - `go test ./...` to verify nothing broken
   - Fix any test failures before ending session

3. **Run lint checks** (if golangci-lint available):
   - `golangci-lint run`
   - Address any critical issues

4. **Verify build** (if code was modified):
   - For affected modules, run: MCP `build <module>` or `go run ./go/cli/eac build <module>`
   - Ensure clean build with no errors

5. **Write session summary file**:
   - Determine filename: `out/session-summary-YYYY-MM-DD-HH-MM.md` (use current date/time)
   - Write the file using the template below
   - Confirm: "Session summary written to out/session-summary-<timestamp>.md"
   - This file is picked up by `/go:status` at the start of the next session

6. **Provide session summary** (print to terminal):

## Output Format

Write to `out/session-summary-YYYY-MM-DD-HH-MM.md` and print to terminal:

```
# Session Summary — YYYY-MM-DD HH:MM

## Accomplished
- [list]

## Files Modified
- [list]

## Code Simplification
- [changes applied or "none needed"]

## Tests
- go test ./...       [PASS / FAIL + details]
- go test -race ./... [PASS / FAIL + details]

## Build
- [PASS / FAIL + details]

## Lint
- [PASS / no critical issues / FAIL + details]

## Known TODOs
- [list or "none"]

## Next Session Should
- [1-3 actionable starting points]
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
