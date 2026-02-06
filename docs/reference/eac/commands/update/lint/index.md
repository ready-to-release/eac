# update lint Commands

Run linters on modules based on their capabilities. The lint system automatically selects the appropriate linter for each module type.

## Commands in this Category

| Command                                 | Purpose                                 |
| --------------------------------------- | --------------------------------------- |
| [update lint](../lint.md)               | Run linters on one or more modules      |
| [update lint (Go)](./go.md)             | Go linting with golangci-lint           |
| [update lint (Markdown)](./markdown.md) | Markdown linting with markdownlint-cli2 |

## Linter Selection

Linters are selected based on component types defined in `component-types.yml`:

| Capability | Linter            | Module Types           |
| ---------- | ----------------- | ---------------------- |
| `go`       | golangci-lint     | go, cli, library       |
| `markdown` | markdownlint-cli2 | static, docs           |

## Quick Examples

```bash
# Lint a single module
eac update lint eac-commands

# Lint all modules
eac update lint

# Lint with auto-fix
eac update lint --fix

# Lint without TUI (CI mode)
eac update lint --no-tui
```

## Output Structure

```text
out/lint/<module>/
├── <component>-<tool>/     # UoW directory (e.g., go-golangci-lint)
│   ├── uow.manifest.json    # Execution metadata
│   ├── lint.log              # Human-readable output
│   └── lint.json             # Structured results (success/failure)
```

## See Also

- [build Commands](../../build/index.md) - Build modules
- [test Commands](../../test/index.md) - Test modules
- [validate Commands](../../validate/index.md) - Validate contracts
