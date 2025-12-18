# pipeline Commands

CI/CD orchestration and diagnostics for module-aware build pipelines.

## Commands in this Category

| Command | Purpose |
|---------|---------|
| [pipeline](./pipeline.md) | Base pipeline command |
| [pipeline ci](./ci.md) | CI orchestration and diagnostics |
| [pipeline ci-dispatch-and-wait](./ci-dispatch-and-wait.md) | Wait for GitHub workflow runs |
| [pipeline ci-summary-link](./ci-summary-link.md) | Generate diagnostic markdown for CI summaries |
| [pipeline run](./run.md) | Execute module pipelines respecting dependencies |
| [pipeline status](./status.md) | Show CI status for head of trunk |
| [pipeline wait](./wait.md) | Wait for GitHub workflow runs to complete |

## Quick Examples

```bash
# Run pipeline for module
r2r eac pipeline run src-auth

# Check CI status
r2r eac pipeline status
```

## See Also

- [Category Overview](../categories/pipeline.md)
- [build](../build/build.md)
