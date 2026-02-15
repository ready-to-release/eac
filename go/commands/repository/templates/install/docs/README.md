# docs

Implements the `templates install docs` command that installs documentation templates by copying files from `templates/docs/` to a configurable destination directory (default: `docs/reference/`).

## Key Types

None (command-only package).

## Key Functions

- **`TemplatesInstallDocs()`** -- Entry point for the `templates install docs` command

## Patterns

- `init()` registration: registers command function with the global registry
- Copy-only installation: copies template files without Go template variable substitution
- Configurable destination: `--destination` flag overrides default `docs/reference/` output directory

## Internal Structure

| File | Responsibility |
| --- | --- |
| docs.go | Documentation template installation with configurable destination |

## Dependencies

- `cli/eac/impl/templates/internal` -- template rendering engine (used in copy mode)
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `core/config` -- configuration loading
- `core/logging` -- structured logging
- `core/paths` -- template source path resolution
- `core/repository` -- repository root discovery

## Role in System

The `docs` install sub-package provisions documentation templates into a project, providing starting-point documentation files that can be customized after installation.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
