# Update ai-summary

<!-- book:cmd update ai-summary -->

Analyzes modules using AI to generate comprehensive status summaries covering architecture (DSL), specifications (Gherkin), source code, and documentation.

Each analysis type produces a separate status file in `out/ai-summary/<module>/`.

## Usage

```bash
eac update ai-summary [flags] [module...]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module...` | Module monikers to analyze (all modules if omitted) |

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--type` | `-t` | Specific analysis type: `dsl`, `specs`, `source`, `docs` |
| `--skip-cache` | | Force regeneration even if cached |
| `--dry-run` | | Show what would be analyzed without executing |
| `--tui` / `--no-tui` | | Enable/disable TUI console |
| `--turbo` | | Enable turbo mode (+2 parallel workers) |
| `--debug` | `-d` | Enable debug logging |

## Analysis Types

- **dsl** -- Analyzes Structurizr DSL architecture files
- **specs** -- Analyzes Gherkin BDD specifications
- **source** -- Analyzes source code (depends on dsl and specs)
- **docs** -- Analyzes documentation files

## Examples

```bash
eac update ai-summary                    # Analyze all modules
eac update ai-summary eac                # Analyze single module
eac update ai-summary --type=dsl core    # Analyze only DSL for core
```

## See Also

- [update design](./design.md)
- [update docs](./docs.md)
