# Pipeline Commands

## Overview

The **pipeline** category contains 7 commands for CI/CD orchestration and diagnostics.

## Commands

<!-- book:category-commands pipeline -->

## Common Use Cases

### Local Pipeline Execution

```bash
r2r eac pipeline run src-auth
```

### CI Status Monitoring

```bash
r2r eac pipeline status
r2r eac pipeline wait
```

### CI Orchestration

```bash
r2r eac pipeline ci-dispatch-and-wait
```

## Key Features

- Module-aware build orchestration
- Dependency-based execution order
- GitHub Actions integration
- CI status monitoring
- Workflow diagnostics

## See Also

- [build](../build/build.md)
- [test Commands](./test.md)
- [get execution-order](../get/execution-order.md)
