# template

Go template rendering utilities for generating files from templates.
Wraps `text/template` with convenience functions for string and file output.

## Key Types

- `Renderer` -- wraps a parsed Go template with methods for rendering to strings and files; supports custom function maps and configurable missing-key behavior

## Key Functions

- `NewRenderer` -- creates a `Renderer` from a template file path with a default function map including string, math, percentage, and formatting helpers
- `WithFuncs` -- adds custom template functions to the renderer (builder pattern)
- `WithMissingKeyMode` -- sets the missing-key behavior (zero, invalid, error)
- `RenderToString` -- executes the template with data and returns the result as a string
- `RenderToFile` -- executes the template with data and writes the result to a file, creating directories as needed
- `NormalizeSpecPath` -- normalizes a spec file path for use in template data (strips `../` prefixes, ensures `specs/` prefix)

## Patterns

- **Template wrapping**: `Renderer` encapsulates template parsing at render time, keeping the API simple
- **Builder pattern**: `WithFuncs` and `WithMissingKeyMode` return `*Renderer` for chaining
- **File output**: `RenderToFile` handles directory creation and error wrapping
- **Rich default function map**: includes `join`, `split`, `upper`, `lower`, `add`, `sub`, `mul`, `div`, `percent`, `percentf`, `riskBadge`, `truncate`, `first`, `last`, and float-precision math

## Internal Structure

| File          | Purpose                                                                                                    |
| ------------- | ---------------------------------------------------------------------------------------------------------- |
| `renderer.go` | `Renderer` with `NewRenderer`, `RenderToString`, `RenderToFile`, `NormalizeSpecPath`, and `defaultFuncMap` |

## Dependencies

None (leaf package within clibase).

## Role in System

Used by commands that generate files from templates, such as spec creation and report generation. Provides a consistent rendering interface that handles file I/O concerns so callers only need to supply template paths and data.

## Code Health

### Tech Debt

- None identified

### Pain Points

- `renderer_test.go` (393 lines) exceeds 300 lines and is almost twice the size of `renderer.go` (215 lines)

### Optimization Opportunities

- None identified
