# EAC Monorepo

Go multi-module monorepo for the EAC/CLIE toolchain.
Go workspace: `go.work` at repository root (Go 1.24.4).

## Key Modules

| Module path | Binary | Description |
|-------------|--------|-------------|
| `go/cli/clie` | `clie` | Extension lifecycle manager — ZERO workspace imports |
| `go/cli/eac` | `eac` | Repository automation CLI |
| `go/core` | — | Shared domain libraries (60+ packages) |
| `go/mcp/commands` | — | MCP server (100+ tools) |

## Critical Isolation Constraint

`go/cli/clie` MUST NOT import any other workspace module in production code.
The module is distributed as a standalone binary. Test code may import local
modules for test infrastructure.

If you need environment variable constants in clie: use
`go/cli/clie/internal/envconsts/constants.go` (local copies, CLIE_ prefixed).

Enforced by: `specs/repository/module-isolation/specification.feature`

## MCP Servers

Two MCP servers are configured in `.mcp.json`:
- **commands** (`mcp__commands__*`): 100+ tools — build, test, lint, release, modules
- **github** (`mcp__github__*`): Docker-based GitHub integration

MCP tools are deferred — use `ToolSearch` to probe before use.
Fallback: `go run ./go/cli/eac <command>` and `gh` CLI.

## Git Policy

READ-ONLY git by default. Never commit, push, branch, or merge unless the user
explicitly requests it. This repo uses git worktrees for parallel sessions —
stay in the current directory only.

## File Organization

- Go modules: `go/` directory
- Intermediate/planning files: `out/` directory ONLY
- Never create `.md` result files outside module directories or `out/`

## Development Workflow

Three Rules of Vibe Coding: easy to understand, easy to change, hard to break.
Three-phase process: Specifications (`.feature`) → TDD → Validation.

Skills: `/go:status`, `/go:plan`, `/go:implement`, `/go:test`, `/go:review`,
        `/go:debug`, `/go:release`, `/go:session-end`, `/go:fix`, `/drawio`

Run `/go:session-end` at the end of EVERY session.
