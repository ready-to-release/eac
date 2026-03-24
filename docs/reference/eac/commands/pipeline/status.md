# pipeline status

<!-- book:cmd pipeline status -->

## Output

- Commit SHA being checked
- List of workflow names with status icons (success, failure, cancelled, skipped, in progress)
- Exit code 0 on success (even if workflows are failing), 1 on errors (missing gh CLI, invalid ref)

## See Also

- [pipeline run](./run.md)
- [release check-ci](../release/check-ci.md)
- [pipeline Commands](../pipeline/index.md)
