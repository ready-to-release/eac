# cache-clear

<!-- book:cmd update cache-clear -->

Removes cache files used by build, lint, test, and scan commands, forcing a full rebuild on the next run.

Default behavior (no `--type`) clears state and work caches. The state cache includes capacity semaphore files that coordinate parallel execution -- clear these if tests or builds hang.

## Usage

```bash
eac update cache-clear [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--type` | | Cache type to clear (default: `state+work`) |
| `--dry-run` | | Show what would be deleted without deleting |
| `--verbose` | `-v` | Show each file being deleted |

## Cache Types

| Type | Description |
|------|-------------|
| `state` | Incremental state (build/lint/test state.json, capacity semaphores) |
| `asset` | Rendered assets (mermaid, drawio, structurizr caches) |
| `work` | Ephemeral work directories (npm work dirs) |
| `registry` | Docker image cache (runs `docker image prune`) |
| `layer` | Docker builder cache (runs `docker builder prune`) |
| `all` | Everything |
| `local:state` | Fine-grained: local state only |

## Examples

```bash
eac update cache-clear                     # Clear state + work (default)
eac update cache-clear --type=all          # Clear everything including Docker
eac update cache-clear --type=asset        # Clear only asset caches
eac update cache-clear --dry-run           # Preview what would be deleted
```

## See Also

- [build](../build/build.md)
- [test](../test/test.md)
- [lint](../lint.md)
- [update Commands](../../categories/update.md)
