# Go Linting

Go modules are linted using [golangci-lint](https://golangci-lint.run/), a fast linters aggregator for Go.

## Configuration

The lint configuration is defined in `.golangci.yml` at the workspace root. This is the **single source of truth** for Go linting rules, used by both the CLI and IDE integrations.

### Config Location

```text
workspace/
├── .golangci.yml          # Lint rules (source of truth)
└── go/
    └── eac/
        └── commands/      # Module uses workspace config
```

### Enabled Linters

The default configuration enables these linter categories:

| Category    | Linters                                           |
| ----------- | ------------------------------------------------- |
| Standard    | errcheck, govet, ineffassign, staticcheck, unused |
| Quality     | misspell, revive, unconvert, unparam              |
| Security    | gosec                                             |
| Performance | prealloc, gocyclo                                 |
| Style       | godot, whitespace, gocritic                       |
| Formatting  | gofmt, goimports, gofumpt                         |

### Key Settings

```yaml
# .golangci.yml (excerpt)
version: "2"

run:
  timeout: 5m
  tests: true

linters:
  default: none
  enable:
    - errcheck      # Unchecked errors
    - govet         # Subtle bugs
    - staticcheck   # Comprehensive analysis
    - gosec         # Security issues
    - gocyclo       # Complexity
    # ... see full config
```

## Usage

```bash
# Lint Go modules
r2r eac update lint eac-commands

# Lint with auto-fix (formatters only)
r2r eac update lint eac-commands --fix

# Use custom config
r2r eac update lint eac-commands --config .golangci-strict.yml
```

## Auto-Fix Support

The `--fix` flag automatically fixes issues for these linters:

- **gofmt** - Formatting
- **goimports** - Import ordering
- **gofumpt** - Stricter formatting
- **whitespace** - Trailing whitespace

Other linters report issues but cannot auto-fix.

## IDE Integration

### VS Code

Install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) and configure:

```json
// .vscode/settings.json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"]
}
```

### GoLand/IntelliJ

1. Go to **Settings > Tools > File Watchers**
2. Add golangci-lint watcher
3. Point to workspace `.golangci.yml`

## Output Example

```text
=== Linting Go files ===
Running: golangci-lint run --config C:\projects\eac\.golangci.yml ./...

internal/docker/hosting-env.go:45:2: ineffectual assignment to err (ineffassign)
internal/docker/hosting-env.go:89:9: Error return value of `cmd.Run` is not checked (errcheck)

✗ Go lint issues found
```

## System Requirements

| Tool          | Minimum Version | Install Command                                                         |
| ------------- | --------------- | ----------------------------------------------------------------------- |
| golangci-lint | 2.0+            | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |

Verify installation:

```bash
golangci-lint version
# golangci-lint has version v2.x.x ...
```

## Troubleshooting

### Timeout errors

Increase timeout in `.golangci.yml`:

```yaml
run:
  timeout: 10m
```

### Linter disagreements

Suppress specific issues with `//nolint` directives:

```go
//nolint:errcheck // intentionally ignoring error for cleanup
_ = file.Close()
```

Prefer specific suppressions over blanket `//nolint` comments.

## See Also

- [update lint](./lint.md) - Main lint command
- [update lint (Markdown)](./markdown.md) - Markdown linting
- [golangci-lint docs](https://golangci-lint.run/)
