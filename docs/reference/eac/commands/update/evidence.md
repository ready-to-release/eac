# evidence

<!-- book:cmd update evidence -->

Generates evidence PDFs from a module's evidence-book components. Evidence books are markdown-based documentation packages that aggregate test results, security scans, and other compliance artifacts.

Output is written to `out/evidence/<module>/`.

## Usage

```bash
eac update evidence <module> [flags]
eac update evidence --all [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module moniker to build evidence for |

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--all` | | Build evidence for all modules with evidence-book components |
| `--verbose` | `-v` | Show detailed progress |

## Examples

```bash
eac update evidence eac              # Build evidence for a single module
eac update evidence --all            # Build evidence for all modules
eac update evidence eac -v           # Verbose output
```

## See Also

- [scan](../scan/scan.md)
- [test](../test/test.md)
- [update Commands](../../categories/update.md)
