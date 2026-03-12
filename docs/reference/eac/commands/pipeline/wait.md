# Pipeline wait

<!-- book:cmd pipeline wait -->

Waits for one or more GitHub Actions workflow runs to complete with a live progress display. Polls each run at a configurable interval and shows status icons for each workflow.

Status icons: `✓` success, `✗` failure, `◐` in progress, `○` queued.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--timeout` | int | `1800` | Maximum wait time in seconds (default: 30 minutes) |
| `--interval` | int | `10` | Poll interval in seconds |

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<run-id>...` | Yes | One or more GitHub Actions run IDs |

## Output

- Live-updating status display with elapsed time
- Exit code 0 if all workflows succeed
- Exit code 1 if any workflow fails or timeout is reached

## Examples

```bash
eac pipeline wait 12345 12346 12347
eac pipeline wait 12345 --timeout 600        # Wait up to 10 minutes
eac pipeline wait 12345 --interval 5         # Poll every 5 seconds
```

## See Also

- [pipeline ci](./ci.md)
- [pipeline status](./status.md)
- [pipeline Commands](../categories/pipeline.md)
