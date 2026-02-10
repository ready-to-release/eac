# render

Multi-format output rendering for CLI commands. Provides markdown tables,
console-aware tables, JSON, TOML, YAML struct rendering, and a custom
renderer registry for extensible output formats.

## Key Types

- **`TableBuilder`** -- Fluent builder for markdown tables
- **`ConsoleTableBuilder`** -- Builder for terminal-width-aware console tables
- **`MarkdownTableConfig`** -- Configuration struct for markdown tables
- **`ConsoleTableConfig`** -- Configuration struct for console tables
- **`AlignedTable`** -- Table with explicit column alignment control
- **`OrderedMap`** -- JSON-serializable map preserving insertion order

## Patterns

- YAML as single source of truth: JSON and TOML renderers marshal to YAML
  first, then convert, ensuring consistent field ordering
- Fluent builder: `TableBuilder` chains `WithHeaders`, `AddRow`, `Build`
  for ergonomic table construction
- Console-aware rendering: `ConsoleTableBuilder` auto-detects terminal
  width and distributes column widths proportionally
- Custom renderer registry: `custom.Register` allows commands to add
  format handlers looked up by name and filtered by command

## Internal Structure

| File               | Responsibility                                     |
| ---                | ---                                                |
| `markdown_table.go`  | `MarkdownTableConfig`, `TableBuilder`, `SimpleMarkdownTable` |
| `console_table.go`   | `ConsoleTableConfig`, `ConsoleTableBuilder`, terminal width  |
| `formatter.go`       | `FormatMarkdownTable` post-processor for alignment  |
| `struct_renderer.go` | Struct-to-YAML-code-block rendering                 |
| `json.go`            | `RenderAsJSON` with order-preserving `OrderedMap`   |
| `toml.go`            | `RenderAsTOML` via YAML intermediate                |
| `custom.go`          | Bridge to `render/custom` registry                  |

## Dependencies

- `render/custom` -- pluggable renderer registry sub-package

## Role in System

This package is the shared output layer of the `clibase` module. Every CLI
command that produces structured output (tables, serialized data, custom
formats) routes through `render` rather than formatting inline. The
`custom` sub-package allows downstream commands to register format handlers
without modifying the render package itself.

## Code Health

### Tech Debt
- `custom/registry.go:18` mutable package-level `registry` map; concurrent command registration would race without external synchronization
- Limited test coverage: only `markdown_table_test.go`, `struct_renderer_test.go`, and `examples_test.go` exist; `console_table.go`, `json.go`, and `toml.go` have no dedicated tests

### Pain Points
- YAML-first serialization path (JSON/TOML marshal via YAML intermediate) adds latency and subtle ordering bugs if YAML tags differ from JSON tags

### Optimization Opportunities
- Add unit tests for `json.go` and `toml.go` round-trip fidelity (low effort)
- Guard `custom/registry.go` global map with `sync.RWMutex` or switch to `sync.Map` if concurrent registration becomes realistic (low effort)
