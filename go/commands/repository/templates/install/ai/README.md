# ai

Implements the `templates install ai` command that installs AI prompt templates by copying files as-is from `templates/ai/` to `.eac/templates/ai/` without variable substitution.

## Key Types

None (command-only package).

## Key Functions

- **`TemplatesInstallAI()`** -- Entry point for the `templates install ai` command

## Patterns

- `init()` registration: registers command function with the global registry
- Copy-only installation: copies template files without Go template variable substitution
- Fixed source and destination: `templates/ai/` to `.eac/templates/ai/`

## Internal Structure

| File | Responsibility |
| --- | --- |
| ai.go | AI prompt template installation with file copying |

## Dependencies

- `cli/eac/impl/templates/internal` -- template rendering engine (used in copy mode)
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `core/config` -- configuration loading
- `core/logging` -- structured logging
- `core/paths` -- template source path resolution
- `core/repository` -- repository root discovery

## Role in System

The `ai` install sub-package provides one-time installation of AI prompt templates into a project. The installed templates contain placeholder variables that users customize after installation.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
