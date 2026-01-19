# update lint

<!-- book:cmd update lint -->

Run linters on one or more modules. The command automatically selects appropriate linters based on module capabilities.

## Synopsis

```bash
r2r eac update lint [modules...] [flags]
```

## Description

The `update lint` command runs language-specific linters on modules based on their declared capabilities. It supports parallel execution, automatic dependency resolution, and both TUI and CI output modes.

Linter selection is automatic based on module type:

- **Go modules** use [golangci-lint](./go.md)
- **Static/docs modules** use [markdownlint-cli2](./markdown.md)

## Flags

| Flag       | Description                              |
| ---------- | ---------------------------------------- |
| `--fix`    | Automatically fix issues where possible  |
| `--config` | Override default linter config file      |
| `--no-tui` | Disable TUI, use plain text output (CI)  |

## Examples

```bash
# Lint all modules
r2r eac update lint

# Lint specific modules
r2r eac update lint eac-commands eac-core

# Lint with auto-fix enabled
r2r eac update lint --fix

# CI mode (no TUI)
r2r eac update lint --no-tui

# Use custom config
r2r eac update lint --config .golangci-strict.yml
```

## Exit Codes

| Code | Meaning                     |
| ---- | --------------------------- |
| 0    | All modules passed linting  |
| 1    | One or more modules failed  |

## Output

Results are written to `out/lint/<module>/`:

```text
out/lint/eac-commands/
├── lint.log              # Full linter output
├── lint.json             # Structured results
└── lint.manifest.json    # Execution metadata
```

### lint.json Structure

```json
{
  "success": true,
  "tool": "golangci-lint"
}
```

## System Dependencies

The lint command requires language-specific tools to be installed:

| Linter            | Install Command                                                         |
| ----------------- | ----------------------------------------------------------------------- |
| golangci-lint     | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |
| markdownlint-cli2 | `npm install -g markdownlint-cli2`                                      |

Check availability with:

```bash
r2r eac show deps-setup-summary --module=eac-commands --deps=go
```

## See Also

- [update lint (Go)](./go.md) - Go linting details
- [update lint (Markdown)](./markdown.md) - Markdown linting details
- [build Commands](../../build/index.md) - Build modules
- [test Commands](../../test/index.md) - Test modules
