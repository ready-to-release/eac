# internal

Provides the template rendering engine, security validation, and template value management for the `templates install` command group. Handles Go template rendering with path traversal prevention.

## Key Types

- **`Renderer`** -- Template rendering engine that processes Go templates with security validation
- **`TemplateValues`** -- Key-value data used for template variable substitution during rendering

## Key Functions

- **`NewRenderer()`** -- Create a new `Renderer` with source template directory and destination output directory
- **`RenderTemplates()`** -- Process all template files from source to destination with variable substitution
- **`renderFile()`** -- Render a single template file with Go template execution
- **`renderString()`** -- Render a string containing Go template directives
- **`copyFile()`** -- Copy a non-template file from source to destination
- **`ValidatePath()`** -- Validate a file path against directory traversal attacks
- **`SecureFilePath()`** -- Sanitize a file path by removing traversal sequences and dangerous characters
- **`LoadValuesFromJSON()`** -- Load template values from a JSON configuration file
- **`ValidateValues()`** -- Validate that all required template values are present

## Patterns

- Security-first path handling: all output paths are validated against traversal attacks before writing
- Go template engine: uses standard `text/template` for variable substitution in template files
- File-type detection: distinguishes template files (`.tmpl`) from static files for copy-vs-render

## Internal Structure

| File | Responsibility |
| --- | --- |
| renderer.go | Template rendering engine with Go template execution and file copying |
| security.go | Path traversal prevention with validation and sanitization |
| values.go | Template value loading from JSON and required-value validation |

## Dependencies

None (standard library only).

## Role in System

The `templates/internal` package is the engine behind all `templates install` sub-commands. It provides the secure template rendering pipeline that takes template directories, substitutes values, validates paths, and writes output files. Every install sub-package (ai, claude, docs, reports, specs) delegates to this renderer.

## Code Health

### Tech Debt
- No test for renderer.go (215 lines) or values.go (74 lines)

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
