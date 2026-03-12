# lint

<!-- book:cmd lint -->

Lints one or more modules using appropriate linters per module type. Runs modules in parallel by default with incremental caching.

Output is written to `out/lint/<module>/` including lint logs, structured results, and UoW manifests.

## Usage

```bash
eac lint [flags] [module...]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module...` | Module monikers to lint (all modules if omitted) |

## Flags

| Flag | Description |
|------|-------------|
| `--fix` | Auto-fix issues where possible |
| `--config PATH` | Override lint config file path |
| `--skip-cache` | Skip incremental cache, force full lint |
| `--dry-run` | Show what would be linted without running linters |
| `--turbo` | Enable turbo mode (increases parallelism) |
| `--sequential` | Run lints sequentially instead of in parallel |
| `--skip-deps` | Skip system dependency verification |
| `--timings` | Show detailed timing summary |
| `--debug` | Enable debug logs to console |
| `--tui` / `--no-tui` | Enable/disable TUI console |
| `--tui-height N` | Set TUI height (3-20, default: 6) |
| `--ascii` | Use ASCII-only characters in TUI |
| `--skip-tui-delay` | Skip TUI exit delay |

## Examples

```bash
eac lint                           # Lint all modules
eac lint eac                       # Lint a single module
eac lint --fix                     # Lint with auto-fix
eac lint --skip-cache              # Force full lint, ignore incremental state
```

## See Also

- [show lint-summary](../show/lint-summary.md) - Display lint summary
- [build](../build/build.md) - Build modules
