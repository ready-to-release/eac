# Pipeline run

<!-- book:cmd pipeline run -->

## How It Works

Executes module pipelines respecting dependency order:

- **Dependency Resolution**: Builds topological sort of module dependencies
- **Parallel Execution**: Independent modules run concurrently
- **Artifact Management**: Ensures build artifacts available for dependents
- **Error Handling**: Stops on first failure, reports affected modules

Each module's pipeline includes build, test, and validation steps.

## See Also

- [pipeline status](./status.md)
- [pipeline ci](./ci.md)
- [pipeline Commands](../../categories/pipeline.md)
