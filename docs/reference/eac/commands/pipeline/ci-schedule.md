# Pipeline ci schedule

Schedule and dispatch CI workflows with concurrency limits and dependency-aware execution.

<!-- book:cmd pipeline ci schedule -->

## Overview

`pipeline ci schedule` is a pull-based scheduler that orchestrates CI workflow dispatch with concurrency control and dependency awareness. It replaces wave-based dispatch with intelligent scheduling that:

1. **Filters modules** - Uses CI cache check (same as `get ci-dispatch`)
2. **Builds dependency graph** - Analyzes module dependencies from repository config
3. **Dispatches intelligently** - Dispatches modules as capacity allows and dependencies are satisfied
4. **Monitors completion** - Polls for workflow completion and dispatches next ready modules
5. **Reports results** - Exits 0 when all complete successfully, 1 on any failure

This is the CI-level analog of the local DependencyScheduler, handling the full dispatch lifecycle.

## How It Works

### Scheduling Algorithm

```text
┌─────────────────┐
│ Filter Modules  │  Check CI cache, identify what needs dispatch
└────────┬────────┘
         │
┌────────▼────────┐
│ Build Dep Graph │  Analyze dependencies from repository config
└────────┬────────┘
         │
┌────────▼────────┐
│ Initial Batch   │  Dispatch modules with no pending dependencies
└────────┬────────┘
         │
    ┌────▼────┐
    │  Poll   │◄─────┐
    └────┬────┘      │
         │           │
    ┌────▼────┐      │
    │Complete?│──No──┘
    └────┬────┘
         Yes
    ┌────▼────┐
    │ Report  │  Summary of completed/failed/cached modules
    └─────────┘
```

### Concurrency Control

- **Default**: 6 concurrent workflows
- **Configurable**: Adjust with `--max-concurrent`
- **Dependency-aware**: Only dispatches when dependencies complete
- **Cascade handling**: Skips dependent modules if parent fails

### CI Cache Integration

Uses the same caching logic as `get ci-dispatch`:

- Checks for successful CI runs at the target SHA
- Skips dispatch for modules with valid cached builds
- Reports cached modules separately

## Usage

```bash
# Basic usage with changed modules
pipeline ci schedule \
  --directly-changed "core" \
  --invalidated "eac docs" \
  --head-sha abc123

# With custom concurrency
pipeline ci schedule \
  --directly-changed "core auth" \
  --max-concurrent 20 \
  --timeout 3600

# Full options
pipeline ci schedule \
  --directly-changed "core" \
  --invalidated "eac" \
  --head-sha abc123 \
  --dispatch-ref main \
  --max-concurrent 10 \
  --timeout 7200 \
  --poll-interval 60 \
  --trigger-run-id 123456
```

## Flags

| Flag                 | Type   | Default        | Description                                               |
| -------------------- | ------ | -------------- | --------------------------------------------------------- |
| `--directly-changed` | string | -              | Space-separated list of directly changed modules          |
| `--invalidated`      | string | -              | Space-separated list of invalidated (dependent) modules   |
| `--head-sha`         | string | auto-detected  | Commit SHA to dispatch CI for                             |
| `--dispatch-ref`     | string | current branch | Git ref to dispatch workflows on                          |
| `--max-concurrent`   | int    | 6              | Maximum number of concurrent CI dispatches                |
| `--timeout`          | int    | 3600           | Maximum time in seconds to wait for all CI                |
| `--poll-interval`    | int    | 30             | How often to check for completed workflows (seconds)      |
| `--trigger-run-id`   | string | -              | Run ID of the triggering workflow (for artifact download) |
| `--mock`             | string | -              | Mock CI cache status (JSON format) for testing            |

## Output

### Console Output

```text
=== CI Schedule Results ===
  Completed: core auth api
  Failed: payment
  Cascade-skipped: order checkout
  Cached (not dispatched): docs website
  Total time: 15m 42s
```

### Exit Codes

- **0** - All dispatched workflows completed successfully
- **1** - One or more workflows failed, or scheduling error occurred

### GitHub Actions Summary

When running in GitHub Actions (`GITHUB_STEP_SUMMARY` set), writes a formatted summary:

```markdown
## CI Scheduler Results

Dispatched **8** module(s) with concurrency-limited scheduling.

**Completed**: `core`, `auth`, `api`

**Failed**: `payment`

**Cascade-skipped**: `order`, `checkout`

**Cached** (valid CI at HEAD): `docs`, `website`

**Total time**: 15m 42s
```

## Examples

### Standard CI Workflow

```bash
# Detect changes and schedule CI
CHANGED=$(eac get changed-modules-ci | jq -r '.changed_modules | join(" ")')
INVALIDATED=$(eac get changed-modules-ci | jq -r '.invalidated_modules | join(" ")')

pipeline ci schedule \
  --directly-changed "$CHANGED" \
  --invalidated "$INVALIDATED" \
  --head-sha $(git rev-parse HEAD)
```

### High-Concurrency Build

```bash
# Allow up to 20 concurrent workflows
pipeline ci schedule \
  --directly-changed "core auth api" \
  --max-concurrent 20 \
  --timeout 7200
```

### Long-Running Pipelines

```bash
# Increase timeout and poll interval for slow builds
pipeline ci schedule \
  --directly-changed "ml-model data-pipeline" \
  --timeout 14400 \
  --poll-interval 120
```

### Testing with Mock Cache

```bash
# Test scheduling logic with mocked CI cache
pipeline ci schedule \
  --directly-changed "core" \
  --mock '{"core": {"has_ci": true, "ci_passed": true}}'
```

## Comparison with pipeline ci

| Feature               | `pipeline ci`         | `pipeline ci schedule`   |
| --------------------- | --------------------- | ------------------------ |
| Dispatch Strategy     | Wave-based (by layer) | Pull-based (by capacity) |
| Concurrency Control   | ❌ No                 | ✅ Yes (configurable)    |
| Dependency Awareness  | ✅ Yes (layers)       | ✅ Yes (graph)           |
| Early Failure Cascade | ⚠️ Partial            | ✅ Full                  |
| Resource Efficiency   | Lower                 | Higher                   |
| Use Case              | Simple pipelines      | Complex, large-scale CI  |

## When to Use

**Use `pipeline ci schedule` when:**

- Building many modules with complex dependencies
- Need to limit concurrent GitHub Actions usage
- Want resource-efficient CI execution
- Have long-running build pipelines
- Need fine-grained control over dispatch timing

**Use `pipeline ci` when:**

- Simple pipeline with few modules
- Prefer straightforward wave-based execution
- Don't need concurrency limits

## Common Workflows

### GitHub Actions Integration

```yaml
name: CI Scheduler

on:
  push:
    branches: [main]

jobs:
  schedule:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Detect Changes
        id: changes
        run: |
          CHANGED=$(eac get changed-modules-ci | jq -r '.changed_modules | join(" ")')
          INVALIDATED=$(eac get changed-modules-ci | jq -r '.invalidated_modules | join(" ")')
          echo "changed=$CHANGED" >> $GITHUB_OUTPUT
          echo "invalidated=$INVALIDATED" >> $GITHUB_OUTPUT

      - name: Schedule CI
        run: |
          eac pipeline ci schedule \
            --directly-changed "${{ steps.changes.outputs.changed }}" \
            --invalidated "${{ steps.changes.outputs.invalidated }}" \
            --head-sha ${{ github.sha }} \
            --max-concurrent 10
```

### Local Testing

```bash
# Test scheduling logic locally with mock cache
pipeline ci schedule \
  --directly-changed "core auth" \
  --invalidated "api" \
  --mock '{"core": {"has_ci": false}, "auth": {"has_ci": false}, "api": {"has_ci": true, "ci_passed": true}}' \
  --max-concurrent 2
```

## Troubleshooting

### No Modules Dispatched

**Symptom**: Scheduler completes immediately with no dispatches.

**Causes**:

- All modules have valid CI cache at HEAD SHA
- Module filtering excluded all candidates
- Empty `--directly-changed` and `--invalidated` lists

**Solution**: Check cache status with `get ci-dispatch --head-sha <sha>`

### Workflows Not Starting

**Symptom**: Scheduler dispatches but workflows don't start.

**Causes**:

- GitHub API rate limiting
- Invalid workflow dispatch configuration
- Missing GitHub Actions workflows for modules

**Solution**: Check GitHub Actions logs and verify workflow files exist

### Timeout Exceeded

**Symptom**: Scheduler exits with timeout error.

**Causes**:

- Workflows taking longer than `--timeout` value
- Stuck or hanging workflows
- Insufficient concurrent capacity

**Solution**:

```bash
# Increase timeout and concurrent capacity
pipeline ci schedule \
  --max-concurrent 15 \
  --timeout 7200 \
  ...
```

### Cascade Failures

**Symptom**: Many modules marked as "cascade-skipped".

**Causes**:

- Early dependency failure cascading to dependents
- Build failures in foundational modules

**Solution**: Fix root cause failures in base modules first

## See Also

- [pipeline ci](./ci.md) - Wave-based CI orchestration
- [pipeline status](./status.md) - Check pipeline status
- [pipeline wait](./wait.md) - Wait for pipeline completion
- [get ci-dispatch](../get/ci-dispatch.md) - Module filtering logic
- [pipeline Commands](../categories/pipeline.md) - All pipeline commands
