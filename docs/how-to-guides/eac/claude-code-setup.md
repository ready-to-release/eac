# Working with Claude Code in the EAC Repository

Learn how to use Claude Code effectively for Go CLI development with this repository's specialized agents, skills, and commands.

## Prerequisites

- **Claude Code CLI** installed and configured
- **MCP servers** configured (`.mcp.json`) - Commands server provides 100+ project-specific commands
- **code-simplifier plugin** installed (mandatory for end-of-session cleanup)

## Quick Start

### Your First Session

```bash
# 1. Start Claude Code
claude-code

# 2. Initialize with project context
> /boot

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

| Command           | Purpose                           | When to Use                          |
| ----------------- | --------------------------------- | ------------------------------------ |
| `/boot`           | Initialize session                | **Start of every session**           |
| `/go:plan`        | Plan feature/change               | Before implementing                  |
| `/go:implement`   | Implement with TDD                | After planning                       |
| `/go:test`        | Write/debug tests                 | Testing phase                        |
| `/go:review`      | Review code + run code-simplifier | Before committing                    |
| `/go:cli-docs`    | Update CLI docs                   | When CLI changes                     |
| `/go:release`     | Check release readiness           | Before tagging release               |
| `/go:debug`       | Debug failures                    | When things break                    |
| `/go:session-end` | **End session cleanup**           | **End of every session (MANDATORY)** |

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

| Skill                    | Purpose                        | Steps                                                                                   |
| ------------------------ | ------------------------------ | --------------------------------------------------------------------------------------- |
| **go-cli-feature**       | End-to-end feature development | Plan → Specify → Design → Test → Implement → Verify → Simplify → Document               |
| **go-cli-release-check** | Pre-release validation         | CI → Security → Build → Changelog → Dependencies → Tests → Docs                         |
| **go-cli-refactor-safe** | Safe refactoring               | Baseline → Plan → Refactor incrementally → Test each step → Simplify                    |

## Common Workflows

### Adding a New CLI Command

**Goal**: Add a `validate config` command

**Workflow**:

```text
/boot
  ↓
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
/boot
  ↓
/go:debug [paste test failure output]
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
/boot
  ↓
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
/boot
  ↓
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

### MCP Tools (100+ Commands)

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
go run ./go/eac/commands <command> [args]
```

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

- ✅ Run `/boot` at start of every session
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
2. Test server: `go run ./go/eac/mcp/commands/main.go < /dev/null`
3. Check Claude Code MCP logs
4. **Fallback**: Use direct CLI: `go run ./go/eac/commands <command>`

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
/boot → /go:debug → /go:test → /go:review → /go:session-end
```

**Example: Documentation update**:

```text
/boot → /go:cli-docs → /go:review → /go:session-end
```

## Learning More

- **Full agent details**: See `.claude/agents/` directory
- **Skill workflows**: See `.claude/skills/` directory
- **Command reference**: See `.claude/commands/` directory
- **Project guidelines**: Read `/agent.md`
- **Module operations**: Use MCP `show-valid-commands`

## Summary

**Every session**:

1. Start: `/boot`
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
