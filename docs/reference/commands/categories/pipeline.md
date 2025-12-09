# pipeline Commands

{{ page_breadcrumb() }}

## Overview

The **pipeline** category contains 7 commands for CI/CD orchestration and diagnostics.

## Commands

| Command | Purpose |
|---------|---------|
| [pipeline run](../pipeline/run.md) | Execute module pipelines respecting dependencies |
| [pipeline status](../pipeline/status.md) | Show CI status for head of trunk |
| [pipeline wait](../pipeline/wait.md) | Wait for GitHub workflow runs to complete |
| [pipeline ci](../pipeline/ci.md) | CI orchestration and diagnostics |
| [pipeline ci-dispatch-and-wait](../pipeline/ci-dispatch-and-wait.md) | Wait for GitHub workflow runs |
| [pipeline ci-summary-link](../pipeline/ci-summary-link.md) | Generate diagnostic markdown for CI summaries |

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

- [build](../other/build.md)
- [test Commands](./test.md)
- [get execution-order](../get/execution-order.md)

{{ diataxis_footer() }}
