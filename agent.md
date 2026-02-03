# Agent

## Session Initialization

**IMPORTANT**: At the start of every session, you MUST:

1. **Detect current workspace** from the environment context (current branch and working directory are provided in `<env>` and `gitStatus`)
2. **Verify workspace context**: Check if the current directory path matches the detected branch (may be a mismatch in multi-worktree setups)
3. **Read this file** (`/agent.md`) to load project context
4. **Junie-Specific Initialization**: If you are Junie, you MUST:
   - Read `./junie/README.md` and all files in `./junie/` following the load order defined in `junie/README.md`
   - Apply `./junie/` instructions as overrides to this file (`agent.md`) where contradictions exist
5. **Verify MCP server availability**:
   - Check your available tool list for `mcp__commands__*` tools (Commands Server)
   - Check your available tool list for `mcp__github__*` tools (GitHub Server)
   - Determine connection status for each: CONNECTED or NOT CONNECTED
   - Set execution mode accordingly (MCP-First or CLI Fallback)
6. **Internalize all constraints and guidelines** defined below
7. **Apply these instructions** throughout the entire session
8. **Confirm initialization** with a flashy initialization report using this format:

```text
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃  ⚡ SYSTEM INITIALIZED ⚡                                    ┃
┃  Project context loaded from agent.md                         ┃
┃  MCP servers: [✅ CONNECTED / ⚠️ NOT CONNECTED]              ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

Workspace Context:
- Current branch: [branch name]
- Working directory: [current path]
- Status: [✓ Match / ⚠ MISMATCH - Expected path: [expected path for this branch]]

Project Context Loaded:
- Go modular monorepo architecture
- Three Rules of Vibe Coding: Easy to understand, Easy to change, Hard to break
- Three-Phase Development: Specifications → TDD → Validation
- Go version: ≥ 1.21
- Claude Code tools: 6 agents, 3 skills, 8 commands available

Active Constraints:
- Git: READ-ONLY by default. No commits/pushes/branches without explicit user request
- Multi-Worktree Aware: Operating in [current directory] ([branch])
- File Organization: Modules in /go, intermediate files in /out
- Execution Mode: [MCP-First / CLI Fallback] based on MCP server availability
- Code-simplifier: Must run at end of every session

MCP Server Status:
Commands Server (mcp__commands__*):
  [✅ CONNECTED - XX tools available / ⚠️ NOT CONNECTED - Using fallback: go run ./go/cli/eac]

GitHub Server (mcp__github__*):
  [✅ CONNECTED - XX tools available / ⚠️ NOT CONNECTED - Using fallback: gh CLI]

[If Commands Server NOT CONNECTED, include this troubleshooting section:]
⚠️ Commands MCP Troubleshooting:
- Server configured in: .mcp.json
- Commands server: go run ./go/mcp/commands/main.go
- Verify: go run ./go/mcp/commands/main.go < /dev/null
- Check Claude Code MCP server logs for errors
- Fallback: Commands operations will use direct CLI commands (go run ./go/cli/eac)

[If GitHub Server NOT CONNECTED, include this troubleshooting section:]
⚠️ GitHub MCP Troubleshooting:
- Server configured in: .mcp.json
- GitHub server: docker run -i --rm -e GITHUB_PERSONAL_ACCESS_TOKEN ghcr.io/github/github-mcp-server:v0.21.0
- Authentication: Set GITHUB_PERSONAL_ACCESS_TOKEN environment variable
  - Windows: $env:GITHUB_PERSONAL_ACCESS_TOKEN = "ghp_your_token_here"
  - Linux/Mac: export GITHUB_PERSONAL_ACCESS_TOKEN="ghp_your_token_here"
- Verify Docker: docker --version
- Verify token: echo $env:GITHUB_PERSONAL_ACCESS_TOKEN (Windows) or echo $GITHUB_PERSONAL_ACCESS_TOKEN (Linux/Mac)
- Fallback: GitHub operations will use gh CLI or git commands

Ready to assist with project tasks.
```

**Git Worktree Context**: This repository uses git worktrees for parallel development. The initialization process automatically detects the current branch based on the working directory. If there's a mismatch between the expected worktree path and the current directory, it will be highlighted in the initialization report.

---

## MCP Server Configuration

This project uses **MCP (Model Context Protocol) servers** to provide specialized commands for managing the modular monorepo architecture.

### Commands Server

**Status**: ✅ Active (when connected)
**Type**: Local Go application
**Configuration**: `.mcp.json` → `go run ./go/mcp/commands/main.go`

**Available Commands** (100+):

- **Module Discovery**: `get-modules`, `show-modules`, `show-components`, `get-files`, `show-files`
- **Dependency Management**: `get-dependencies`, `show-dependencies`, `validate-dependencies`, `get-execution-order`
- **Architecture**: `create-design`, `update-design`, `validate-design`, `serve-design`
- **Build and Test**: `build`, `test`, `pipeline-run`, `get-test-results`, `show-test-summary`
- **Documentation**: `serve` (MkDocs), `update-docs`, `create-spec`, `validate-specs`
- **Git Operations**: `work-commit`, `show-files-changed`, `show-files-staged`, `get-changed-modules`
- **Release Management**: `release-*`, `get-changelog`, `get-release-notes`
- **Security**: `scan`, `scan-zap`, `validate-risk-catalog`, `validate-risk-profile`

See full list: Use MCP `show-valid-commands` or `go run ./go/cli/eac show valid-commands`

### MCP Execution Policy

**When CONNECTED** (MCP tools available):

- ✅ Use `mcp__commands__*` tools for all operations
- Faster, native integration with Claude Code

**When NOT CONNECTED** (fallback mode):

- ✅ Use direct CLI: `go run ./go/cli/eac <command> [args]`
- All functionality remains available

**IMPORTANT during boot**: DO NOT call data-heavy commands like `get-files`, `get-modules`, `get-dependencies` during initialization. Only verify MCP availability, don't invoke large data operations.

### GitHub MCP Server

**Status**: ✅ Configured (requires GITHUB_PERSONAL_ACCESS_TOKEN environment variable)
**Type**: Docker container (GitHub official MCP server)
**Configuration**: `.mcp.json` → `docker run -i --rm -e GITHUB_PERSONAL_ACCESS_TOKEN ghcr.io/github/github-mcp-server:v0.21.0`

**Available Tools** (when connected):

- **Workflow Management**: `list_workflows`, `get_workflow_runs`, `get_workflow_run_logs`
- **Repository**: `get_repository`, `get_file_contents`, `search_code`
- **Issues & PRs**: `list_issues`, `create_issue`, `list_pull_requests`
- **Commits**: `list_commits`, `get_commit`

**Setup Required**:

```powershell
# Windows PowerShell
$env:GITHUB_PERSONAL_ACCESS_TOKEN = "ghp_your_personal_access_token_here"

# Or add to your PowerShell profile for persistence
Add-Content $PROFILE '$env:GITHUB_PERSONAL_ACCESS_TOKEN = "ghp_your_token_here"'
```

**When CONNECTED** (GitHub MCP tools available):

- ✅ Use `mcp__github__*` tools for workflow analysis, issue management
- Used by `go-workflow-engineer` agent for CI/CD analysis

**When NOT CONNECTED** (fallback mode):

- ✅ Use `gh` CLI: `gh workflow list`, `gh run view`, `gh issue list`
- All GitHub functionality remains available via CLI

**IMPORTANT during boot**: Test GitHub MCP connectivity to determine execution mode. Check for `mcp__github__*` tools in your available tool list.

**Token Permissions Required**:

- `repo` - Full repository access
- `workflow` - Update GitHub Action workflows
- `read:org` - Read org data (if analyzing org repos)

---

## Project Constraints

### Git Operations

- **DO NOT** perform ANY git operations unless explicitly requested by the user
- **READ-ONLY git operations** (`log`, `status`, `diff`) are permitted ONLY when explicitly needed for information gathering
- **NEVER** run git modifying operations (`commit`, `push`, `add`, `stash`, `checkout`, `branch`, `merge`, `rebase`, `worktree`, etc.)
- **USER CONTROLS GIT**: The user manages all git operations manually, especially in multi-worktree setups
- **USE** MCP `work-commit` ONLY when the user explicitly requests commit message generation

#### Git Worktree Awareness

This repository may be running multiple Claude Code sessions in parallel using **git worktrees**. Each worktree is an isolated working directory on a different branch.

**When operating in a worktree environment:**

1. **Identify your location**: Check `git worktree list` (read-only) ONLY if needed to understand context
2. **Stay in your lane**: Work only on files in your current worktree directory
3. **No cross-worktree operations**: Never attempt to modify files in other worktrees
4. **User coordinates branches**: The user manages branch switching, merging, and synchronization
5. **Report your worktree**: When asked about git state, mention which worktree you're operating in

Each Claude session works independently. The user handles all git coordination.

### File Organization

- **Modules**: All Go modules are placed in the `go/` directory (`go/cli/eac` for EAC CLI, `go/cli/r2r` for R2R CLI, `go/core` for shared libraries, `go/adapters/*` for external integrations)
- **Result files**: DO NOT create result markdown files except in:
  - Module directories (as identified by module contracts or `get-files`)
  - `/out/<my-result-file>.md` for intermediate/temporary files
- **Intermediate files**: CREATE all intermediate files, shell scripts, analysis results in `/out/` directory
- **Before modifications**: USE `get-files` **on-demand** to understand file ownership
  - WARNING: This command loads ~2690 files (~19k tokens). Only call when you need to determine module ownership before making changes.
  - Alternative: Use `show-files-changed` or `show-files-staged` for smaller, targeted queries

---

## Claude Code Tools

This repository has a complete Claude Code setup with specialized agents, skills, and commands for Go CLI development.

### Available Sub-Agents

Specialized agents for specific tasks (invoke via Task tool):

| Agent                    | Purpose                                | When to Use                                                               |
| ------------------------ | -------------------------------------- | ------------------------------------------------------------------------- |
| **go-architect**         | Design architecture and plan modules   | Planning features, designing interfaces, evaluating trade-offs            |
| **go-cli-ux**            | Design CLI commands, flags, and output | Adding commands, improving output, designing UX                           |
| **go-test-engineer**     | Write comprehensive tests              | Writing tests (TDD), debugging test failures, improving coverage          |
| **go-debugger**          | Debug failures and investigate issues  | Test failures, runtime panics, performance issues                         |
| **go-security-release**  | Security scanning and release checks   | Pre-release validation, security audits, release readiness                |
| **go-workflow-engineer** | Analyze GitHub workflows and CI/CD     | Debugging workflows, optimizing pipelines, validating CD model compliance |

### Available Skills

Orchestrated workflows combining multiple agents. See `.claude/skills/` for detailed documentation.

| Skill                       | Purpose                                     | Key Agents                                | When to Use                                |
| --------------------------- | ------------------------------------------- | ----------------------------------------- | ------------------------------------------ |
| **go-cli-feature**          | End-to-end feature development (TDD)        | go-architect, go-cli-ux, go-test-engineer | Building new CLI commands or features      |
| **go-cli-refactor-safe**    | Safe refactoring with continuous validation | go-architect, go-test-engineer            | Refactoring code without breaking behavior |
| **go-cli-release-check**    | Pre-release validation checklist            | go-security-release                       | Before tagging releases or deployment      |
| **go-comprehensive-review** | Multi-perspective code review               | All agents                                | Important features, security changes       |
| **drawio-editor**           | DrawIO diagram editing                      | None (standalone)                         | Architecture diagrams, visualizations      |

**Workflows**:

- `go-cli-feature`: Plan → Specify → Design UX → Test → Implement → Verify → Simplify → Document
- `go-cli-refactor-safe`: Baseline → Plan → Refactor → Test → Simplify
- `go-cli-release-check`: CI → Security → Build → Changelog → Dependencies → Tests → Docs → Final Check
- `go-comprehensive-review`: Context → Multi-Agent Analysis → Aggregate Findings
- `drawio-editor`: Decode → Edit → Encode → Embed

**Quick Selection**:

- New features → `go-cli-feature`
- Refactoring → `go-cli-refactor-safe`
- Release prep → `go-cli-release-check`
- Code review → `go-comprehensive-review`
- Diagrams → `drawio-editor`

### How to Use Skills

**Method 1: Via Slash Commands** (Recommended)

Use slash commands that automatically load skill instructions:

```text
/go:plan          # Loads go-plan skill
/go:implement     # Loads go-implement skill
/go:review        # Loads go-review skill
```

**Method 2: Reference Workflow Skills**:

Request Claude to follow a specific workflow skill:

```text
Follow the go-cli-feature skill to implement the new command
```

**Method 3: Agent Delegation**:

Commands delegate to agents, which may use workflow skills:

```text
/go:release → go-security-release agent → go-cli-release-check workflow
```

### Available Slash Commands

Quick-access commands for common workflows:

| Command               | Purpose                            | Use Case                              |
| --------------------- | ---------------------------------- | ------------------------------------- |
| `/boot`               | Initialize session                 | Start every session                   |
| `/go:plan`            | Plan feature or change             | Before implementing                   |
| `/go:implement`       | Implement using TDD                | After planning                        |
| `/go:test`            | Write or debug tests               | Testing phase                         |
| `/go:review`          | Review code (runs code-simplifier) | Before committing                     |
| `/go:cli-docs`        | Update CLI documentation           | When CLI surface changes              |
| `/go:release`         | Prepare for release                | Release readiness check               |
| `/go:debug`           | Debug issues                       | When things break                     |
| **`/go:session-end`** | **End session cleanup**            | **MANDATORY at end of every session** |

### Recommended Workflows

#### Feature Development

```text
/boot → /go:plan → /go:implement → /go:test → /go:review → /go:session-end
```

#### Bug Fix

```text
/boot → /go:debug → /go:test (regression) → /go:implement → /go:review → /go:session-end
```

#### Release Preparation

```text
/boot → /go:release → Address issues → Ready to tag
```

### Code-Simplifier Integration

**MANDATORY**: The `code-simplifier` agent must run at the end of every session.

**How to invoke**:

The code-simplifier is invoked via Task tool with `subagent_type="code-simplifier:code-simplifier"`. It runs automatically in:

1. `/go:review` command (before commit/PR)
2. `/go:session-end` command (end of session)
3. Skills: `go-cli-feature` (step 7), `go-cli-refactor-safe` (step 5)

**What it does**:

- Simplifies code for clarity and maintainability
- Removes unnecessary complexity
- Improves naming and structure
- Preserves all functionality (refactors, doesn't change behavior)

**User responsibility**:

- MUST run `/go:session-end` at the end of EVERY session
- Review simplifications (don't blindly accept)
- Commit simplifications separately from feature work

**Agent availability**:

The code-simplifier agent is built into Claude Code. Use Task tool with `subagent_type="code-simplifier:code-simplifier"` to invoke it.

---

## Development Workflow

You are an AI coding agent contributing Go code to this repository following the **Three Rules of Vibe Coding**:

1. **Make the code easy to understand**: Clear, simple, explicit
2. **Make the code easy to change**: Small functions, stable boundaries, no hidden state
3. **Make it hard to break**: Tests, validation, clear errors

### Three-Phase Development Process

**MANDATORY** for all development tasks:

1. **Specifications First**: Write `.feature` files before code (use MCP `create-spec`)
2. **Test-Driven Development**: Write tests before implementation
3. **Validation**: Run all tests before reporting complete

**For detailed guidance**:

- Architecture & Design → Use **go-architect** agent
- CLI UX Design → Use **go-cli-ux** agent
- Test Writing → Use **go-test-engineer** agent
- Debugging → Use **go-debugger** agent
- Security & Release → Use **go-security-release** agent
- Workflow & CI/CD Analysis → Use **go-workflow-engineer** agent

**For complete workflows**:

- Feature development → Use **go-cli-feature** skill
- Safe refactoring → Use **go-cli-refactor-safe** skill
- Release readiness → Use **go-cli-release-check** skill

**Or use slash commands**:

- `/go:plan` → `/go:implement` → `/go:test` → `/go:review` → `/go:session-end`

---

## Go-Specific Guidelines

**Requirements**:

- Go version: **≥ 1.21**
- Enforce: `gofmt`, `go vet`, `golangci-lint` (config: `.golangci.yml`)
- Use idiomatic Go: standard library preferred, minimal exported APIs
- Avoid global mutable state

**Key Patterns**:

- **Errors**: Wrap with `%w` for context
- **Context**: Use `context.Context` for all I/O operations
- **Packages**: Organize by domain, use `internal/` for private code
- **Functions**: 20-40 lines ideal, <100 lines max
- **Files**: 200-400 lines ideal, <1000 lines max

**For detailed patterns and examples**, see:

- **go-architect** agent: Architecture and interfaces
- **go-cli-ux** agent: CLI-specific patterns
- **go-test-engineer** agent: Testing patterns

---

## Quality Bars

All code must meet these standards before completion:

### Tests

- ✅ All tests pass (`go test ./...`)
- ✅ No race conditions (`go test -race ./...`)
- ✅ Table-driven tests for multiple scenarios
- ✅ Test coverage maintained or improved

### Code Quality

- ✅ Follows Go conventions (`gofmt`, `go vet`)
- ✅ Passes linting (`golangci-lint run`)
- ✅ Functions are small (20-40 lines ideal, <100 lines max)
- ✅ Clear, intention-revealing names
- ✅ Errors wrapped with %w for context
- ✅ No unnecessary complexity

### Documentation

- ✅ Public APIs have doc comments
- ✅ CLI help text updated (if applicable)
- ✅ How-to guides updated if CLI surface changed

### Security

- ✅ No hardcoded secrets
- ✅ Input validation for user data
- ✅ Proper error handling (don't leak internals)

### End-of-Session Checklist

**MANDATORY**: Complete these steps at the end of EVERY session:

1. ✅ Run `/go:session-end` command
2. ✅ Code-simplifier executed and changes applied
3. ✅ All tests passing (`go test ./...`)
4. ✅ Clean build (no errors)
5. ✅ Lint passing (if golangci-lint available)
6. ✅ Ready for commit/PR

---

## Additional Resources

- **Full command reference**: Use MCP `show-valid-commands` or `go run ./go/cli/eac help`
- **How-to guide**: See `docs/how-to-guides/eac/claude-code-setup.md` for detailed workflows
- **Module structure**: Use MCP `show-modules` to see all modules
- **Dependencies**: Use MCP `show-dependencies <module>` to see dependency graph
