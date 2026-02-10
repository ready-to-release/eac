# claude

Implements the `templates install claude` command that installs Claude Code configuration templates by copying workflow files from `templates/claude/` to `.claude/` without variable substitution.

## Key Types

None (command-only package).

## Key Functions

- **`TemplatesInstallClaude()`** -- Entry point for the `templates install claude` command

## Patterns

- `init()` registration: registers command function with the global registry
- Copy-only installation: copies template files without Go template variable substitution
- Fixed source and destination: `templates/claude/` to `.claude/`

## Internal Structure

| File | Responsibility |
| --- | --- |
| claude.go | Claude Code template installation (agents, commands, skills, setup files) |

## Dependencies

- `cli/eac/impl/templates/internal` -- template rendering engine (used in copy mode)
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `core/config` -- configuration loading
- `core/logging` -- structured logging
- `core/paths` -- template source path resolution
- `core/repository` -- repository root discovery

## Role in System

The `claude` install sub-package provisions Claude Code integration templates into a project, including agent definitions, command templates, skill workflows, and MCP setup configuration.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
