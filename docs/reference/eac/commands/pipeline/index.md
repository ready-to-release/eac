# Pipeline Commands

The **pipeline** category contains commands for CI/CD orchestration and diagnostics.

**Key Features**:

- Module-aware build orchestration
- GitHub Actions integration
- CI status monitoring
- Workflow diagnostics

## Commands in this Category

| Command                                                    | Purpose                                          |
| ---------------------------------------------------------- | ------------------------------------------------ |
| [pipeline](./pipeline.md)                                  | Base pipeline command                            |
| [pipeline ci](./ci.md)                                     | CI orchestration and diagnostics                 |
| [pipeline ci-dispatch-and-wait](./ci-dispatch-and-wait.md) | Wait for GitHub workflow runs                    |
| [pipeline ci-summary-link](./ci-summary-link.md)           | Generate diagnostic markdown for CI summaries    |
| [pipeline run](./run.md)                                   | Execute module pipelines respecting dependencies |
| [pipeline status](./status.md)                             | Show CI status for head of trunk                 |
| [pipeline wait](./wait.md)                                 | Wait for GitHub workflow runs to complete        |

## Common Use Cases

### Local Pipeline Execution

```bash
eac pipeline run src-auth
```

### CI Status Monitoring

```bash
eac pipeline status
eac pipeline wait
```

### CI Orchestration

```bash
eac pipeline ci-dispatch-and-wait
```

## See Also

- [build](../build/build.md)
- [test Commands](../test/index.md)
