# templates

Installs project template files for documentation, AI prompts, specifications, reports, and Claude Code workflows. Copies templates from the repository to designated project directories, with optional Go `text/template` value substitution or as-is copying that preserves placeholders.

## Key Types

- **`BaseConfig`** -- Common configuration with workspace root, debug flag, and logger
- **`Renderer`** -- Walks a template directory tree, rendering files with value substitution or copying as-is
- **`TemplateValues`** -- `map[string]interface{}` of key-value pairs for template variable replacement
- **`Config`** -- Per-subcommand configuration with destination path, workspace root, and debug flag (defined in each install sub-package)

## Patterns

- Registry-based subcommand dispatch: parent command, `templates install`, and each template type register independently via `init()`
- Dual mode rendering: `Renderer` either applies Go `text/template` substitution when values are provided, or copies files byte-for-byte preserving all `{{ .Variable }}` placeholders
- Path traversal prevention: `ValidatePath` rejects absolute paths, `..` escapes, and any resolved path outside the output directory
- Container-aware root resolution: checks `GetContainerRoot()` for Docker execution, falls back to workspace root
- Fixed destination convention: each template type writes to a predetermined directory (e.g., `.eac/templates/ai/`, `specs/risk-controls/`, `.claude/`)
- Available template discovery: scans command registry for `templates install *` entries to list valid template types

## Internal Structure

| File | Responsibility |
| --- | --- |
| templates.go | Parent command entry point, help display, subcommand dispatch |
| install.go | `templates install` base handler, unknown template detection, available template listing |
| config.go | `BaseConfig` with shared flag parsing, logger initialization, workspace root resolution |
| install/ai/ai.go | `templates install ai` -- AI prompt templates to `.eac/templates/ai/` |
| install/claude/claude.go | `templates install claude` -- Claude Code workflow templates to `.claude/` |
| install/docs/docs.go | `templates install docs` -- documentation templates to `docs/reference/` with optional `--destination` |
| install/reports/reports.go | `templates install reports` -- report templates to `.clie/templates/reports/` |
| install/specs/specs.go | `templates install specs` -- specification templates to `specs/risk-controls/` |
| internal/renderer.go | `Renderer` with directory walk, file rendering, string template paths, and byte copy |
| internal/security.go | `ValidatePath` and `SecureFilePath` for path traversal prevention |
| internal/values.go | `TemplateValues` type, JSON loading, and required-key validation |

## Dependencies

- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration and subcommand lookup
- `core/config` -- logs path resolution for debug output
- `core/logging` -- structured logging with debug/default modes
- `core/paths` -- template source paths, destination directory constants, and EAC directory references
- `core/repository` -- repository root discovery and container root detection

## Role in System

The `templates` package provides the `templates install` command family for `eac-cli`, enabling projects to bootstrap standard file structures for documentation, AI prompts, security specifications, compliance reports, and Claude Code integration. Its `Renderer` supports both raw copying for initial scaffolding and value-substituted rendering for automated generation, while the `internal/security` layer ensures template paths cannot escape their designated output directories.

## Code Health

### Tech Debt
- Install sub-packages (ai, docs, reports, specs) are each ~200 lines with near-identical structure: init registration, flag parsing, path resolution, renderer invocation
- Only install/claude has a dedicated unit test; install/ai, install/docs, install/reports, and install/specs have no tests
- Each install sub-package declares its own `var log = logging.C()` independently

### Pain Points
- High boilerplate duplication across the five install sub-packages; adding a new template type requires copying an entire file and changing paths/names
- No test coverage for install/ai, install/docs, install/reports, install/specs despite containing path resolution and rendering logic

### Optimization Opportunities
- Extract a generic install handler that takes template source path, destination path, and optional values, reducing each sub-package to a thin registration wrapper (high feasibility, all follow identical patterns)
- Good internal/ test coverage exists (security_test.go, templates_test.go); focus new tests on install sub-package path resolution edge cases (moderate feasibility)
