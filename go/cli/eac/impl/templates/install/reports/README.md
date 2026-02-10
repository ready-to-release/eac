# reports

Implements the `templates install reports` command that installs report templates by copying files from `templates/reports/` to `.clie/templates/reports/` without variable substitution.

## Key Types

None (command-only package).

## Key Functions

- **`TemplatesInstallReports()`** -- Entry point for the `templates install reports` command

## Patterns

- `init()` registration: registers command function with the global registry
- Copy-only installation: copies template files without Go template variable substitution
- Fixed source and destination: `templates/reports/` to `.clie/templates/reports/`

## Internal Structure

| File | Responsibility |
| --- | --- |
| reports.go | Report template installation with file copying |

## Dependencies

- `cli/eac/impl/templates/internal` -- template rendering engine (used in copy mode)
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `core/config` -- configuration loading
- `core/logging` -- structured logging
- `core/paths` -- template source path resolution
- `core/repository` -- repository root discovery

## Role in System

The `reports` install sub-package provisions report templates into a project, providing starting-point report templates that can be customized for project-specific reporting needs.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
