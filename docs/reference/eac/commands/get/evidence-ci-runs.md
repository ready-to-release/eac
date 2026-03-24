# get evidence-ci-runs

<!-- book:cmd get evidence-ci-runs -->

## Output Fields

| Field                | Description                                      |
| -------------------- | ------------------------------------------------ |
| `module`             | The requested module moniker                     |
| `ci_runs`            | List of CI runs needed for evidence              |
| `ci_runs[].module`   | Dependency module moniker                        |
| `ci_runs[].workflow` | CI workflow filename (e.g. `ci-core.yaml`)       |
| `ci_runs[].run_id`   | GitHub Actions run ID to download artifacts from |
| `skipped`            | Modules without CI workflows (skipped)           |

## How It Works

1. Resolves all transitive dependencies for the given module
2. For each dependency with a `ci-{module}.yaml` workflow, queries GitHub for the last successful run
3. Fails if any dependency with a CI workflow has no successful run

## See Also

- [pipeline download-evidence-artifacts](../pipeline/download-evidence-artifacts.md) - Download artifacts using these CI runs
- [get build-deps](./build-deps.md) - Get build dependencies for a module
- [get changed-modules-ci](./changed-modules-ci.md) - Get modules requiring rebuild
