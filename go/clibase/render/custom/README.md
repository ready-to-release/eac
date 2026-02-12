# custom

Pluggable custom renderer registry for extensible output formats.
Commands register named renderers that transform YAML data into custom string formats,
filtered by which commands they apply to.

## Key Types

- `CustomRenderer` -- function type that takes YAML bytes and returns a formatted string
- `RendererRegistration` -- holds a renderer and its command filters (which commands it supports)

## Key Functions

- `Register` -- registers a named custom renderer with an optional list of supported commands; typically called from `init()` functions
- `Get` -- retrieves a renderer by name, checking that it supports the given command
- `List` -- returns renderer names that support a given command

## Patterns

- **Plugin registry**: renderers self-register via `init()` functions, and the registry dispatches by name with command-level filtering
- **Command filtering**: each renderer declares which commands it supports (e.g., only `get-modules`); wildcard `*` matches all commands
- **YAML-in, string-out**: all custom renderers consume YAML bytes and produce formatted strings, providing a uniform contract

## Internal Structure

| File          | Purpose                                                                                                          |
| ------------- | ---------------------------------------------------------------------------------------------------------------- |
| `registry.go` | `CustomRenderer`, `RendererRegistration`, `Register()`, `Get()`, `List()`, `SupportsCommand()`                   |
| `count.go`    | `RenderCount` renderer: produces a simple module count (registered for `get-modules` only)                       |
| `summary.go`  | `RenderSummary` renderer: produces a human-readable module summary with statistics (registered for all commands) |

## Dependencies

None (leaf package; uses only `gopkg.in/yaml.v3` for YAML parsing).

## Role in System

Extends the `render` package with named custom output formats. When a user passes `--format count` or `--format summary`, the render layer looks up the renderer from this registry and delegates formatting. New custom formats can be added by creating a file with an `init()` function that calls `Register()`.

## Code Health

### Tech Debt

- None identified

### Pain Points

- Test files are larger than implementation: `registry_test.go` (255 lines), `summary_test.go` (211 lines)
- Implementation files are small: `registry.go` (102 lines), `summary.go` (77 lines), `count.go` (26 lines)

### Optimization Opportunities

- None identified
