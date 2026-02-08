# Render Package

> Migrated from `go/clibase/render/README.md`.

The `render` package provides utilities for rendering data structures as markdown, supporting both tables and YAML-formatted struct output.

## Features

- **Markdown Tables**: Render data as well-formatted markdown tables
- **Struct to Markdown**: Convert Go structs to YAML code blocks
- **Fluent Builder API**: Chain methods for easy table construction
- **Multiple Output Formats**: Tables, lists, key-value pairs

## Installation

```go
import "github.com/ready-to-release/eac/go/clibase/render"
```

## Usage

### Markdown Tables

#### Simple Table

```go
headers := []string{"Name", "Age", "City"}
rows := [][]interface{}{
    {"Alice", 30, "NYC"},
    {"Bob", 25, "LA"},
    {"Charlie", 35, "Chicago"},
}

result := render.SimpleMarkdownTable(headers, rows)
fmt.Println(result)
```

Output:

```text
| Name    | Age | City    |
| ------- | --: | ------- |
| Alice   |  30 | NYC     |
| Bob     |  25 | LA      |
| Charlie |  35 | Chicago |
```

#### Table Builder (Fluent API)

```go
result := render.NewTableBuilder().
    WithHeaders("Module", "Type", "Status").
    WithAutoIndex().
    AddRow("cli", "application", "active").
    AddRow("contracts", "library", "active").
    AddRow("mcp", "server", "active").
    WithFooter("", "Total", "3 modules").
    Build()

fmt.Println(result)
```

Output:

```text
|   # | Module    | Type        | Status     |
| --: | --------- | ----------- | ---------- |
|   1 | cli       | application | active     |
|   2 | contracts | library     | active     |
|   3 | mcp       | server      | active     |
|     |           | Total       | 3 modules  |
```

#### Key-Value Table

```go
data := map[string]interface{}{
    "Name":    "MyProject",
    "Version": "1.0.0",
    "Author":  "Team",
    "License": "MIT",
}

result := render.RenderKeyValueTable("Property", "Value", data)
fmt.Println(result)
```

#### Compact List

```go
items := []string{"Initialize database", "Load configuration", "Start server"}
result := render.RenderCompactList("Startup Tasks", items)
fmt.Println(result)
```

### Struct to Markdown (YAML)

#### Basic Struct

```go
type Config struct {
    Host    string `yaml:"host"`
    Port    int    `yaml:"port"`
    Enabled bool   `yaml:"enabled"`
}

config := Config{
    Host:    "localhost",
    Port:    8080,
    Enabled: true,
}

markdown, err := render.RenderStructAsMarkdown(config)
```

#### Multiple Sections

```go
sections := map[string]interface{}{
    "Database": dbConfig,
    "Cache":    cacheConfig,
}

markdown, err := render.RenderMultipleStructs(sections)
```

### Multi-Format Serialization

The package supports rendering structs as JSON and TOML using YAML as the
single source of truth for serialization:

```go
jsonStr, err := render.RenderAsJSON(myStruct)
tomlStr, err := render.RenderAsTOML(myStruct)
```

## API Reference

### Table Functions

- `SimpleMarkdownTable(headers []string, rows [][]interface{}) string` - Quick table creation
- `RenderMarkdownTable(config *MarkdownTableConfig) string` - Full control over table rendering
- `NewTableBuilder() *TableBuilder` - Fluent API for building tables
- `RenderKeyValueTable(keyHeader, valueHeader string, data map[string]interface{}) string` - Two-column key-value table
- `RenderCompactList(header string, items []string) string` - Single-column indexed list

### Table Builder Methods

- `WithHeaders(headers ...string) *TableBuilder` - Set column headers
- `WithAutoIndex() *TableBuilder` - Add automatic row numbering
- `AddRow(cells ...interface{}) *TableBuilder` - Add a single row
- `AddRows(rows [][]interface{}) *TableBuilder` - Add multiple rows
- `WithFooter(cells ...interface{}) *TableBuilder` - Set footer row
- `Build() string` - Generate the markdown table

### Struct Rendering Functions

- `RenderStructAsMarkdown(v interface{}) (string, error)` - Convert struct to YAML code block
- `RenderStructAsMarkdownOrPanic(v interface{}) string` - Panic on error variant
- `RenderStructWithTitle(title string, v interface{}) (string, error)` - Add markdown title
- `RenderStructSliceAsMarkdown(slice interface{}) (string, error)` - Render slice of structs
- `RenderMultipleStructs(sections map[string]interface{}) (string, error)` - Multiple sections

## Dependencies

- `github.com/jedib0t/go-pretty/v6/table` - Table rendering
- `gopkg.in/yaml.v3` - YAML serialization
- `github.com/pelletier/go-toml/v2` - TOML serialization
