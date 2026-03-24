# pipeline ci-schedule

<!-- book:cmd pipeline ci schedule -->

## Overview

`pipeline ci schedule` is a pull-based scheduler that orchestrates CI workflow dispatch with concurrency control and dependency awareness. It replaces wave-based dispatch with intelligent scheduling that:

1. **Filters modules** - Uses CI cache check (same as `get ci-dispatch`)
2. **Builds dependency graph** - Analyzes module dependencies from repository config
3. **Dispatches intelligently** - Dispatches modules as capacity allows and dependencies are satisfied
4. **Monitors completion** - Polls for workflow completion and dispatches next ready modules
5. **Reports results** - Exits 0 when all complete successfully, 1 on any failure

This is the CI-level analog of the local DependencyScheduler, handling the full dispatch lifecycle.

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

## Common Workflows

### GitHub Actions Integration

{% raw %}

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

{% endraw %}

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
- [pipeline Commands](../pipeline/index.md) - All pipeline commands
