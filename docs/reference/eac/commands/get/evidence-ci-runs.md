# Get evidence-ci-runs

<!-- book:cmd get evidence-ci-runs -->

Returns CI run IDs for a module and all its transitive dependencies, used for downloading test and scan artifacts during evidence building. Queries GitHub for the last successful CI run per dependency.

## Usage

```bash
eac get evidence-ci-runs <module> [--format json]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `module` | Yes | Module moniker to get evidence CI runs for |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--format` | string | `json` outputs just the `ci_runs` array; otherwise standard get formats |

## Output Fields

| Field | Description |
|-------|-------------|
| `module` | The requested module moniker |
| `ci_runs` | List of CI runs needed for evidence |
| `ci_runs[].module` | Dependency module moniker |
| `ci_runs[].workflow` | CI workflow filename (e.g. `ci-core.yaml`) |
| `ci_runs[].run_id` | GitHub Actions run ID to download artifacts from |
| `skipped` | Modules without CI workflows (skipped) |

## How It Works

1. Resolves all transitive dependencies for the given module
2. For each dependency with a `ci-{module}.yaml` workflow, queries GitHub for the last successful run
3. Fails if any dependency with a CI workflow has no successful run

## Examples

```bash
# Get CI runs for evidence building
eac get evidence-ci-runs clie

# JSON format for shell parsing
CI_RUNS=$(eac get evidence-ci-runs clie --format json)
echo "$CI_RUNS" | jq -c '.[]' | while read entry; do
  module=$(echo "$entry" | jq -r '.module')
  run_id=$(echo "$entry" | jq -r '.run_id')
  gh run download "$run_id" --pattern "test-results-${module}*"
done
```

## See Also

- [pipeline download-evidence-artifacts](../pipeline/download-evidence-artifacts.md) - Download artifacts using these CI runs
- [get build-deps](./build-deps.md) - Get build dependencies for a module
- [get changed-modules-ci](./changed-modules-ci.md) - Get modules requiring rebuild
