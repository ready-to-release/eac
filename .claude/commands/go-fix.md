# Fix

```text
description: "Restore a broken pipeline: vet, build, test — with module isolation exit criteria"
```

You are restoring a broken pipeline state in this Go workspace. Something is
broken. Diagnose and fix it, then verify all exit criteria before reporting done.

## When to Use

- `go vet ./...` is failing (Stop hook fired)
- CI checks failing on the current branch
- `go build ./...` compilation errors
- `go test ./...` test failures
- A prior session left the codebase in a broken state

## Process

1. **Assess the break**:
   - Run `go vet ./...` from workspace root — note all failures
   - Run `go build ./...` — note compilation errors
   - Run `go test ./...` — note test failures
   - Or use MCP: `build <module>`, `test <module>`, `lint <module>`
   - Group failures by module

2. **Check module isolation** (if `go/cli/clie` was touched):
   - Verify no new imports from workspace modules were added
   - Permitted: `contracts/*` only in production code
   - Check: `go list -deps github.com/ready-to-release/eac/go/cli/clie/...`
   - Flag any `go/core`, `go/cli/eac`, `go/mcp`, `go/clibase`, `go/adapters` imports

3. **Fix in order** (smallest blast radius first):
   - Fix compilation errors first (`go vet` / `go build`)
   - Fix unit test failures
   - Fix integration test failures last
   - Make the minimal change — do not refactor while fixing

4. **Verify exit criteria** (all must be green before reporting done):
   - [ ] `go vet ./...` — zero errors
   - [ ] `go build ./...` — succeeds
   - [ ] `go test ./...` — passes (or pre-existing failures documented)
   - [ ] `go test -race ./...` — clean (for modules with concurrency)
   - [ ] clie isolation intact (if `go/cli/clie` was touched)

5. **Report**:

```text
Fix Summary
===========
Root cause:   [one sentence]
Files changed: [list]

Exit criteria:
  go vet ./...        [PASS / FAIL]
  go build ./...      [PASS / FAIL]
  go test ./...       [PASS / N failures]
  go test -race ./... [PASS / FAIL]
  clie isolation:     [INTACT / N/A]

Status: [PIPELINE RESTORED / PARTIALLY FIXED — <what remains>]
```
