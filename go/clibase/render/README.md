# render

Multi-format output rendering for CLI commands. Provides markdown tables,
console-aware tables, JSON, TOML, YAML struct rendering, and a custom
renderer registry for extensible output formats.

## Usage

```go
import "github.com/ready-to-release/eac/go/clibase/render"
```

### Markdown Tables

Quick table from headers and rows:

```go
result := render.SimpleMarkdownTable(
    []string{"Name", "Age", "City"},
    [][]interface{}{
        {"Alice", 30, "NYC"},
        {"Bob", 25, "LA"},
    },
)
```

Full control via config struct:

```go
result := render.RenderMarkdownTable(&render.MarkdownTableConfig{
    Headers: []string{"Name", "Status"},
    Rows:    [][]interface{}{{"cli", "active"}},
})
```

### Table Builder (Fluent API)

```go
result := render.NewTableBuilder().
    WithHeaders("Module", "Type", "Status").
    WithAutoIndex().
    AddRow("cli", "application", "active").
    AddRow("contracts", "library", "active").
    WithFooter("", "Total", "2 modules").
    Build()
```

Builder methods: `WithHeaders`, `WithAutoIndex`, `AddRow`, `AddRows`,
`AddSeparator`, `WithFooter`, `WithColumnMaxWidth`, `WithMarkdown`,
`Build`, `BuildMarkdown`.

### Console Tables

Terminal-width-aware tables that auto-detect width and distribute columns
proportionally:

```go
result := render.NewConsoleTableBuilder().
    WithHeaders("Name", "Description").
    WithMaxWidth(120).
    WithTruncate().
    AddRow("cli", "Main CLI tool").
    Build()
```

Builder methods: `WithHeaders`, `WithMaxWidth`, `WithColumnMaxWidth`,
`WithTruncate`, `AddRow`, `AddSeparator`, `Build`.

`FormatListForColumn` formats a string slice for display inside a single
table cell, truncating to `maxItems`.

### Aligned Tables

`AlignedTable` wraps `go-pretty` directly for explicit column alignment:

```go
at := render.NewAlignedTable()
at.SetHeaders("Key", "Value")
at.AddRow("host", "localhost")
result := at.RenderMarkdown()
```

### Key-Value Tables and Lists

```go
render.RenderKeyValueTable("Property", "Value", map[string]interface{}{
    "Name": "MyProject",
    "Version": "1.0.0",
})

render.RenderCompactList("Tasks", []string{"Init", "Load", "Start"})
```

### Struct Rendering (YAML code blocks)

```go
md, err := render.RenderStructAsMarkdown(config)
md, err  = render.RenderStructWithTitle("Database", config)
md, err  = render.RenderMultipleStructs(map[string]interface{}{
    "Database": dbConfig,
    "Cache":    cacheConfig,
})
md, err  = render.RenderStructSliceAsMarkdown(items)
md       = render.RenderStructAsMarkdownOrPanic(config)
```

### JSON and TOML

YAML is the single source of truth: JSON and TOML renderers marshal to YAML
first, then convert, ensuring consistent field ordering.

```go
jsonStr, err := render.RenderAsJSON(v)       // direct JSON marshal
jsonStr, err  = render.RenderAsJSONViaYAML(v) // YAML-first, order-preserving
jsonStr       = render.RenderAsJSONOrPanic(v)

tomlStr, err := render.RenderAsTOML(v)       // direct TOML marshal
tomlStr, err  = render.RenderAsTOMLViaYAML(v) // YAML-first
tomlStr       = render.RenderAsTOMLOrPanic(v)
```

`OrderedMap` is a JSON-serializable map that preserves insertion order,
used internally by the YAML-to-JSON pipeline.

### Custom Renderers

The `render/custom` sub-package provides a pluggable renderer registry.
Commands register format handlers that are looked up by name and filtered
by command:

```go
render.RenderAsCustom(data, "my-format", "show-modules")
render.ListCustomRenderers("show-modules")
```

### Utilities

- `FormatMarkdownTable(raw string) string` -- post-processes raw markdown
  table text with proper alignment and padding.
- `TrimMarkdownTable(md string) string` -- trims whitespace from table cells.
- `GetTerminalWidth() int` -- returns current terminal width (default 80).

## Internal Structure

| File                 | Responsibility                                               |
| -------------------- | ------------------------------------------------------------ |
| `markdown_table.go`  | `MarkdownTableConfig`, `TableBuilder`, `AlignedTable`, helpers |
| `console_table.go`   | `ConsoleTableConfig`, `ConsoleTableBuilder`, terminal width  |
| `formatter.go`       | `FormatMarkdownTable` post-processor for alignment           |
| `struct_renderer.go` | Struct-to-YAML-code-block rendering                          |
| `json.go`            | `RenderAsJSON`, `RenderAsJSONViaYAML`, `OrderedMap`          |
| `toml.go`            | `RenderAsTOML`, `RenderAsTOMLViaYAML`                        |
| `custom.go`          | Bridge to `render/custom` registry                           |

## Dependencies

- `github.com/jedib0t/go-pretty/v6/table` -- table rendering
- `gopkg.in/yaml.v3` -- YAML serialization
- `github.com/pelletier/go-toml/v2` -- TOML serialization
- `render/custom` -- pluggable renderer registry sub-package
