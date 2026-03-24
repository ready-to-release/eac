# pipeline ci

<!-- book:cmd pipeline ci -->

## How It Works

Orchestrates CI pipeline execution for affected modules:

- **Change Detection**: Identifies modules requiring rebuild
- **Dispatch Layers**: Groups modules by dependency depth for parallel execution
- **Workflow Triggering**: Dispatches GitHub Actions workflows for each module
- **Status Monitoring**: Tracks completion and reports failures
- **Artifact Management**: Ensures build artifacts are available for dependent modules

Coordinates distributed CI execution across multiple GitHub Actions runners.

## See Also

- [pipeline wait](./wait.md)
- [pipeline status](./status.md)
- [pipeline Commands](../pipeline/index.md)
