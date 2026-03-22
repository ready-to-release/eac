# Pipeline status

<!-- book:cmd pipeline status -->

Displays the status of GitHub Actions workflows for a specific commit or branch. By default shows the status for the current branch HEAD.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ref` | string | `HEAD` | Git ref (branch/tag) to check status for |
| `--commit` | string | | Specific commit SHA to check |

## Output

- Commit SHA being checked
- List of workflow names with status icons (success, failure, cancelled, skipped, in progress)
- Exit code 0 on success (even if workflows are failing), 1 on errors (missing gh CLI, invalid ref)

## Examples

```bash
eac pipeline status                    # Current branch HEAD
eac pipeline status --ref develop      # Check develop branch
eac pipeline status --commit abc123    # Check specific commit
```

## See Also

- [pipeline run](./run.md)
- [release check-ci](../release/check-ci.md)
- [pipeline Commands](../../categories/pipeline.md)
