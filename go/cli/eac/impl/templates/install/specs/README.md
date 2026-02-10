# specs

Implements the `templates install specs` command that installs specification templates by copying files from `templates/specs/` to `specs/risk-controls/` without variable substitution.

## Key Types

None (command-only package).

## Key Functions

- **`TemplatesInstallSpecs()`** -- Entry point for the `templates install specs` command

## Patterns

- `init()` registration: registers command function with the global registry
- Copy-only installation: copies template files without Go template variable substitution
- Fixed source and destination: `templates/specs/` to `specs/risk-controls/`

## Internal Structure

| File | Responsibility |
| --- | --- |
| specs.go | Specification template installation with file copying |

## Dependencies

- `cli/eac/impl/templates/internal` -- template rendering engine (used in copy mode)
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `core/config` -- configuration loading
- `core/logging` -- structured logging
- `core/paths` -- template source path resolution
- `core/repository` -- repository root discovery

## Role in System

The `specs` install sub-package provisions risk control specification templates into a project, providing Gherkin feature file templates for security control verification.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
