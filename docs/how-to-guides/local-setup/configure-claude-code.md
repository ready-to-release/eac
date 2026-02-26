# Working with Claude Code

Learn how to use Claude Code effectively for Go CLI development with this repository's specialized agents, skills, and commands.

## Prerequisites

- **Claude Code CLI** installed and configured
- **MCP servers** configured (`.mcp.json`) - Commands server provides 100+ project-specific commands
- **code-simplifier plugin** installed (mandatory for end-of-session cleanup)

## Quick Start

### Your First Session

> **Note**: Project context (`CLAUDE.md`) loads automatically when you open
> Claude Code — no manual initialization needed. Run `/go:status` if you want
> to confirm MCP server availability or check what a prior session left in progress.

```bash
# 1. Start Claude Code
claude

# 2. (Optional) Verify your environment
> /go:status

# 3. Plan your feature
> /go:plan add a 'status' command that shows repository health

# 4. Implement with TDD
> /go:implement

# 5. Review changes
> /go:review

# 6. End session (MANDATORY)
> /go:session-end
```

## Available Tools

### Slash Commands (Quick Access)

| Command           | Purpose                                   | When to Use                          |
| ----------------- | ----------------------------------------- | ------------------------------------ |
| `/go:status`      | Verify MCP, worktree, prior session       | Optional — confirm environment       |
| `/go:plan`        | Plan feature/change                       | Before implementing                  |
| `/go:implement`   | Implement with TDD                        | After planning                       |
| `/go:test`        | Write/debug tests                         | Testing phase                        |
| `/go:review`      | Review code + run code-simplifier         | Before committing                    |
| `/go:fix`         | Restore broken pipeline                   | CI/build/vet/test failures           |
| `/go:cli-docs`    | Update CLI docs                           | When CLI changes                     |
| `/go:release`     | Check release readiness                   | Before tagging release               |
| `/go:debug`       | Debug failures                            | When things break                    |
| `/go:session-end` | **End session cleanup + session summary** | **End of every session (MANDATORY)** |

### Sub-Agents (Specialized Tasks)

Invoke via: "Use the go-architect agent to design..."

| Agent                   | Purpose               | Example Use                          |
| ----------------------- | --------------------- | ------------------------------------ |
| **go-architect**        | Architecture & design | "Design a caching layer for the API" |
| **go-cli-ux**           | CLI command design    | "Design flags for the new command"   |
| **go-test-engineer**    | Testing               | "Write tests for config validation"  |
| **go-debugger**         | Debug issues          | "Debug why TestParse is failing"     |
| **go-security-release** | Security & release    | "Run pre-release checks"             |

### Skills (Complete Workflows)

Invoke via: "Run the go-cli-feature skill for..."

| Skill                    | Purpose                        | Steps                                                                     |
| ------------------------ | ------------------------------ | ------------------------------------------------------------------------- |
| **go-cli-feature**       | End-to-end feature development | Plan → Specify → Design → Test → Implement → Verify → Simplify → Document |
| **go-cli-release-check** | Pre-release validation         | CI → Security → Build → Changelog → Dependencies → Tests → Docs           |
| **go-cli-refactor-safe** | Safe refactoring               | Baseline → Plan → Refactor incrementally → Test each step → Simplify      |

## Common Workflows

### Adding a New CLI Command

**Goal**: Add a `validate config` command

**Workflow**:

```text
/go:plan add a 'validate config' command that checks configuration files
  ↓
Claude delegates to go-architect → provides architecture plan
  ↓
/go:implement
  ↓
Claude:
1. Creates Gherkin spec (specs/validate-config.feature)
2. Writes failing tests (validate_config_test.go)
3. Implements command (validate_config.go)
4. Runs tests until all pass
  ↓
/go:review
  ↓
Claude runs code-simplifier + validates quality
  ↓
/go:cli-docs
  ↓
Claude updates help text and creates how-to guide
  ↓
/go:session-end
  ↓
Claude:
- Runs code-simplifier again
- Verifies tests pass
- Provides session summary
```

**Result**: Complete, tested, documented feature ready to commit

### Debugging a Test Failure

**Goal**: Fix failing test

**Workflow**:

```text
/go:debug [paste test failure output]  OR  /go:fix [for pipeline failures]
  ↓
Claude delegates to go-debugger:
- Analyzes stack trace
- Identifies root cause
- Proposes fix
  ↓
/go:test write regression test for this bug
  ↓
Claude creates test that reproduces the bug
  ↓
/go:implement the fix
  ↓
Claude implements fix, all tests now pass
  ↓
/go:session-end
```

**Result**: Bug fixed with regression test to prevent recurrence

### Safe Refactoring

**Goal**: Simplify a 300-line function

**Workflow**:

```text
Use the go-cli-refactor-safe skill to refactor the config parser
  ↓
Claude:
1. Baseline: Records current test status (245 tests passing)
2. Plan: Breaks parser into stages
3. Refactor incrementally:
   - Extract validation (commit, tests pass)
   - Extract parsing (commit, tests pass)
   - Extract transformation (commit, tests pass)
   - Simplify main function (commit, tests pass)
4. Verify: All 245 tests still passing
5. Simplify: Runs code-simplifier
  ↓
/go:session-end
```

**Result**: Code refactored safely, all tests passing, improved structure

### Preparing a Release

**Goal**: Release commands module v1.5.0

**Workflow**:

```text
/go:release check if commands module is ready for v1.5.0
  ↓
Claude runs go-cli-release-check skill:
1. ✅ CI: All checks passing
2. ✅ Security: No vulnerabilities
3. ✅ Build: Success
4. ✅ Changelog: v1.5.0 documented
5. ✅ Dependencies: Valid
6. ✅ Tests: 245 passed, 0 failed
7. ✅ Documentation: Current
  ↓
Claude: ✅ APPROVED FOR RELEASE v1.5.0
```

**Result**: Confidence to tag and release

## Understanding the Setup

### CLAUDE.md — Auto-loaded Context

A `CLAUDE.md` file at the repository root is loaded automatically by Claude Code
at the start of every session. It contains stable project facts: module layout,
the clie isolation constraint, MCP server names, and the skills list.

You do not need to run any command to load it — it is always present.

For module-specific context, additional `CLAUDE.md` files exist in:

- `go/cli/clie/` — isolation constraint and environment constants
- `go/core/` — dependency position and key packages
- `go/cli/eac/` — command entrypoints and build info

These load automatically when Claude works in those directories.

### Stop Hook — Automatic go vet Gate

After every Claude response that modifies Go files, `go vet ./...` runs
automatically at the workspace root. This catches obvious errors (unreachable
code, printf format mismatches, suspicious constructs) before you see the result.

If the hook fires with errors, use `/go:fix` to address vet failures
systematically with guided exit criteria.

No configuration needed — this runs via `.claude/settings.json` automatically.

### MCP Tools

The repository has an MCP server providing commands like:

**Module Operations**:

- `show-modules`: List all modules
- `get-dependencies`: Show dependency graph
- `build`, `test`: Build and test modules

**Specifications**:

- `create-spec`: Generate Gherkin from description
- `validate-specs`: Validate specifications

**Release Management**:

- `release-check-ci`: Verify CI status
- `get-changelog`: View changelog

**See full list**: Use MCP `show-valid-commands`

**Fallback**: If MCP not connected, all commands work via CLI:

```bash
go run ./go/cli/eac <command> [args]
```

### MCP Server Execution Modes

The MCP server supports two execution modes:

#### Production Mode (Default)

**Configuration** (`.mcp.json`):

```json
{
  "mcpServers": {
    "commands": {
      "command": "eac-mcp-commands"
    }
  }
}
```

**Execution**: `eac <command>` (Docker container)
**Startup**: ~2s (Docker overhead)
**Use Case**: CI/CD, production, consistent environments

#### Development Mode (Local Override)

**Configuration** (`.mcp.json`):

```json
{
  "mcpServers": {
    "commands": {
      "command": "go",
      "args": ["run", "./go/mcp/commands/main.go"],
      "env": {
        "EAC_USE_DIRECT_BINARY": "true"
      }
    }
  }
}
```

**Execution**: Direct binary via `paths.CommandsBinaryPath()`
**Startup**: ~100ms (no Docker)
**Use Case**: Local development, fast iteration

#### Choosing a Configuration

**For Local Development**: Use `go run` with `EAC_USE_DIRECT_BINARY=true`
**For Production**: Use `eac-mcp-commands` (default)

### Code-Simplifier Plugin

> **MANDATORY at end of every session**

**What it does**:

- Simplifies code for clarity
- Removes unnecessary complexity
- Improves naming
- Preserves all functionality

**How it runs**:

1. Automatically via `/go:review`
2. Automatically via `/go:session-end`
3. Within skills at appropriate steps

**Expected output**:

```text
Running code-simplifier...
Files analyzed: 5
Files simplified: 3

Changes applied:
- config_parser.go: Simplified logic (3 changes)
- validator.go: Improved naming (2 changes)

Tests: ✅ All still pass
```

## Best Practices

### Always

- ✅ Project context loads automatically — run `/go:status` to verify MCP availability
- ✅ Run `/go:session-end` at end of every session
- ✅ Write tests first (TDD)
- ✅ Keep commits small and focused
- ✅ Use slash commands for common tasks
- ✅ Delegate to specialized agents
- ✅ Run code-simplifier before committing

### Never

- ❌ Skip `/go:session-end`
- ❌ Commit without running tests
- ❌ Skip code-simplifier
- ❌ Make large changes without planning
- ❌ Implement without tests

### Three Rules of Vibe Coding

All code must be:

1. **Easy to understand**: Clear, simple, explicit
2. **Easy to change**: Small functions, stable boundaries
3. **Hard to break**: Tests, validation, clear errors

## Quality Checklist

Before marking work complete:

**Tests**:

- ✅ All tests pass (`go test ./...`)
- ✅ No race conditions (`go test -race`)
- ✅ Coverage maintained/improved

**Code Quality**:

- ✅ Follows Go conventions (`gofmt`, `go vet`)
- ✅ Lint passing (`golangci-lint run`)
- ✅ Functions small (< 40 lines ideal)
- ✅ Clear naming

**Documentation**:

- ✅ CLI help text updated
- ✅ How-to guides current

**Process**:

- ✅ Code-simplifier ran
- ✅ Ready for commit/PR

## Troubleshooting

### MCP Not Connected

**Symptoms**: Commands like `show-modules` not available

**Solution**:

1. Verify config: Check `.mcp.json`
2. Test server: `go run ./go/mcp/commands/main.go < /dev/null`
3. Check Claude Code MCP logs
4. **Fallback**: Use direct CLI: `go run ./go/cli/eac <command>`

### code-simplifier Not Running

**Symptoms**: `/go:session-end` doesn't simplify code

**Solution**:

1. Verify plugin installed
2. Manually invoke: Use Task tool with subagent_type="code-simplifier"
3. Check Claude Code plugin list

### Tests Failing After Refactoring

**Symptoms**: Tests pass at start, fail after changes

**Solution**:

1. **Revert last change**: `git revert HEAD` or undo edits
2. Make smaller incremental changes
3. Run tests after EACH change
4. Use `/go:debug` to identify root cause

## Advanced Usage

### Parallel Agent Workflows

For complex features, use multiple agents in parallel:

```text
User: "Work on authentication feature using parallel agents"

Claude spawns in parallel:
- Design agent: Architecture and interfaces
- Test agent: Test plan and scaffolding
- Analysis agent: Impact analysis
- Docs agent: Documentation structure

Then synthesizes results into unified plan
```

### Custom Workflows

Combine commands for custom workflows:

**Example: Quick bug fix**:

```text
/go:debug → /go:test → /go:review → /go:session-end
```

**Example: Documentation update**:

```text
/go:cli-docs → /go:review → /go:session-end
```

## Learning More

- **Full agent details**: See `.claude/agents/` directory
- **Command reference**: See `.claude/commands/` directory
- **Project guidelines**: Loaded automatically from `CLAUDE.md` (stable facts) and `agent.md` (deep reference, on demand)
- **Module operations**: Use MCP `show-valid-commands`

## Summary

**Every session**:

1. (Optional) Verify environment: `/go:status`
2. Work: Use appropriate commands/agents
3. End: `/go:session-end` (MANDATORY)

**Every feature**:

1. Plan: `/go:plan`
2. Implement: `/go:implement` (TDD)
3. Review: `/go:review` (runs code-simplifier)

**Every commit**:

- ✅ Tests pass
- ✅ Code simplified
- ✅ Documentation updated

With Claude Code and this setup, you have a complete toolkit for fast, safe Go CLI development.
