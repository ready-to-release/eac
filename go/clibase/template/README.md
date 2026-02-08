# template

Go template rendering utilities for generating files from templates.
Wraps `text/template` with convenience functions for string and file output.

## Key Types

- `Renderer` -- wraps a parsed Go template with methods for rendering to strings and files

## Key Functions

- `NewRenderer` -- creates a `Renderer` from a template string with custom function map
- `RenderToString` -- executes the template with data and returns the result as a string
- `RenderToFile` -- executes the template with data and writes the result to a file
- `NormalizeSpecPath` -- normalizes a spec file path for use in template data

## Patterns

- **Template wrapping**: `Renderer` encapsulates template parsing errors at construction time, keeping render calls simple
- **File output**: `RenderToFile` handles directory creation, atomic writing, and error wrapping

## Internal Structure

| File | Purpose |
|---|---|
| `renderer.go` | `Renderer` with `NewRenderer`, `RenderToString`, `RenderToFile`, and path utilities |

## Dependencies

None (leaf package within clibase).

## Role in System

Used by commands that generate files from templates, such as spec creation and report generation. Provides a consistent rendering interface that handles file I/O concerns so callers only need to supply template strings and data.

## Code Health

### Tech Debt
- None identified; no TODO/FIXME markers, no mutable global state, all functions are short

### Pain Points
- None identified; compact leaf package with strong test coverage (393 test lines vs 215 source lines)

### Optimization Opportunities
- None identified; the package is well-structured and thoroughly tested
