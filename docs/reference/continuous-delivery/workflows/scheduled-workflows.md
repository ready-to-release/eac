# Scheduled Workflows
Reference for scheduled (cron) workflows.

## Overview

Scheduled workflows run on a time-based schedule to perform periodic validation and maintenance tasks. These workflows catch issues that incremental CI might miss and ensure the full codebase remains healthy over time.

## Scheduled Full CI Workflow

**File:** `.github/workflows/cron-full-trigger.yaml`

**Name:** `Scheduled Full CI`

**Purpose:** Runs a full CI rebuild every 2 hours to detect deviations between incremental CI and full builds.

### Triggers

```yaml
on:
  schedule:
    - cron: '0 */2 * * *'  # Every 2 hours (00:00, 02:00, 04:00, etc.) UTC
  workflow_dispatch:        # Manual trigger for testing
```

**Schedule:** Every 2 hours, at the top of even hours (UTC)

**Manual trigger:** Allows testing the change-detection logic on demand

### Permissions

```yaml
permissions:
  actions: write    # Trigger workflows and wait for completion
  contents: read    # Read repository content
```

### Purpose and Rationale

**Problem:** Incremental CI may miss integration issues that only appear in full builds

**Examples of issues caught:**

- **Change detection bugs:** Incremental CI missed rebuilding affected modules
- **Flaky tests:** Tests pass individually but fail in full builds
- **Integration issues:** Modules work separately but fail together
- **Environment drift:** CI environment changed between runs
- **Dependency updates:** External dependency changes affecting builds

**Solution:** Periodic full rebuilds detect deviations from incremental CI

### Job: check-and-trigger

Checks if a full rebuild is needed and triggers it if necessary.

**Outputs:**

- `skipped` - Whether the rebuild was skipped (recent success exists)
- `run_id` - Run ID of triggered CI workflow
- `result` - Result of the triggered workflow (success/failure)

#### Step 1: Check if Full Rebuild Needed

```yaml
- name: Check if full rebuild needed
  id: check
  run: |
    # Check for successful runs in the last 2 hours
    TWO_HOURS_AGO=$(date -u -d '2 hours ago' '+%Y-%m-%dT%H:%M:%SZ')

    RECENT_SUCCESS=$(gh run list \
      --workflow "cron-full-trigger.yaml" \
      --branch main \
      --status success \
      --json headSha,createdAt \
      | jq --arg cutoff "$TWO_HOURS_AGO" --arg sha "⟪ github.sha ⟫" \
        '[.[] | select(.createdAt > $cutoff and .headSha == $sha)] | length')

    if [ "${RECENT_SUCCESS:-0}" -gt 0 ]; then
      echo "skip=true"
    else
      echo "skip=false"
    fi
```

**Logic:**

1. Calculate 2 hours ago timestamp
2. Query GitHub API for successful `cron-full-trigger.yaml` runs
3. Filter to runs within last 2 hours for current HEAD
4. Skip if recent success found, otherwise trigger rebuild

**Optimization:** Prevents duplicate full rebuilds for the same commit

#### Step 2: Trigger Full Rebuild

```yaml
- name: Trigger full rebuild
  id: trigger
  if: steps.check.outputs.skip != 'true'
  run: |
    gh workflow run change-trigger.yaml \
      --repo ⟪ github.repository ⟫ \
      --ref main \
      -f trigger-all=true

    # Wait for run to be created
    sleep 10

    # Get the run ID
    RUN_ID=$(gh run list \
      --workflow "change-trigger.yaml" \
      --branch main \
      --limit 1 \
      --json databaseId \
      --jq '.[0].databaseId')

    echo "run_id=$RUN_ID"
```

**Behavior:**

- Triggers `change-trigger.yaml` with `trigger-all=true` input
- Bypasses incremental change detection
- Builds all modules regardless of changes
- Captures run ID for monitoring

**Key Parameter:** `trigger-all=true` - Forces full rebuild

#### Step 3: Wait for Full Rebuild to Complete

```yaml
- name: Wait for full rebuild to complete
  id: wait
  if: steps.check.outputs.skip != 'true' && steps.trigger.outputs.run_id != ''
  run: |
    RUN_ID="⟪ steps.trigger.outputs.run_id ⟫"

    if gh run watch "$RUN_ID" --exit-status; then
      echo "result=success"
    else
      echo "result=failure"
      echo "::error::Full rebuild FAILED - deviation detected"
    fi
  timeout-minutes: 120
```

**Behavior:**

- Waits for triggered workflow to complete (up to 2 hours)
- Uses `gh run watch` to monitor execution
- Reports success or failure
- Errors on failure to indicate deviation

**Timeout:** 120 minutes (2 hours) - Accounts for full build of all modules

### Job: summary

Generates summary and reports deviations.

**Dependencies:** `check-and-trigger` (waits for completion)

**Condition:** `always()` - Runs even if previous job fails

#### Summary Generation

```yaml
- name: Generate Summary
  run: |
    SKIPPED="⟪ needs.check-and-trigger.outputs.skipped ⟫"
    RESULT="⟪ needs.check-and-trigger.outputs.result ⟫"
    RUN_ID="⟪ needs.check-and-trigger.outputs.run_id ⟫"

    if [ "$SKIPPED" = "true" ]; then
      echo "No action needed - recent full rebuild exists"
    elif [ "$RESULT" = "success" ]; then
      echo "Full rebuild completed successfully"
    elif [ "$RESULT" = "failure" ]; then
      echo "DEVIATION DETECTED - Full rebuild failed"
    fi
```

**Summary Content:**

**If skipped:**

- Current HEAD SHA
- Explanation: Recent full rebuild exists within 2 hours

**If successful:**

- Current HEAD SHA
- Run ID with link to triggered workflow
- Confirmation: Full rebuild passed

**If failed (DEVIATION DETECTED):**

- Current HEAD SHA
- Run ID with link to failed workflow
- Warning: Deviation detected
- Possible causes explanation
- Diagnostic commands for investigation

#### Step: Fail on Deviation

```yaml
- name: Fail on deviation
  if: needs.check-and-trigger.outputs.result == 'failure'
  run: |
    echo "::error::Scheduled full CI detected a deviation"
    exit 1
```

**Purpose:** Marks workflow as failed to alert on deviations

### Deviation Detection

**What is a deviation?**

A deviation occurs when:

- Incremental CI passed for commits on main
- But a full rebuild of the same commit fails

**This indicates:**

1. **Change detection issue:** Incremental CI didn't rebuild affected modules
2. **Flaky test:** Test passes alone but fails in full builds
3. **Integration issue:** Modules work separately but fail together
4. **Environment drift:** CI environment changed between runs

### Deviation Diagnosis

When a deviation is detected, the summary provides diagnostic commands:

```bash
# View the failed full rebuild run
gh run view <run-id> --repo <owner>/<repo>

# View failed step logs
gh run view <run-id> --log-failed

# Compare with recent incremental CI runs
gh run list \
  --workflow change-trigger.yaml \
  --branch main \
  --limit 5

# Check which modules failed
gh run view <run-id> --json jobs \
  --jq '.jobs[] | select(.conclusion != "success") | {name, conclusion}'
```

### Investigation Steps

1. **Identify failed module(s):**

   - Check which module CI workflow failed in the full rebuild
   - Review test failures or build errors

2. **Compare with incremental CI:**

   - Check if the same module passed in recent incremental CI
   - Look for differences in test execution or environment

3. **Reproduce locally:**

   ```bash
   # Build all modules
   r2r eac build

   # Test all modules with all suites
   r2r eac test --suite unit+integration+acceptance
   ```

4. **Check change detection:**

   ```bash
   # Review recent changes
   r2r eac get changed-modules-ci

   # Review dependency graph
   r2r eac get dependencies <module>
   ```

5. **Review test logs:**
   - Look for timing-dependent failures
   - Check for resource contention (parallel execution)
   - Identify environment-specific issues

### Manual Testing

Test the scheduled workflow manually:

```bash
# Trigger manually (bypasses schedule)
gh workflow run cron-full-trigger.yaml

# Monitor execution
gh run watch <run-id>

# View results
gh run view <run-id>
```

### Schedule Configuration

**Current schedule:** Every 2 hours

**Customization:**

```yaml
schedule:
  # Every hour
  - cron: '0 * * * *'

  # Every 6 hours
  - cron: '0 */6 * * *'

  # Daily at midnight
  - cron: '0 0 * * *'

  # Multiple schedules
  - cron: '0 */2 * * *'  # Every 2 hours
  - cron: '0 0 * * 1'    # Weekly on Monday
```

**Considerations:**

- **Frequency:** Balance between deviation detection and resource usage
- **Timing:** Align with low-activity periods to reduce impact
- **Cost:** More frequent runs consume more GitHub Actions minutes

### Performance Impact

**Resource usage:**

- Full rebuild takes 30-60 minutes (all modules)
- Runs 12 times per day (every 2 hours)
- Total: ~6-12 hours of CI time per day

**Optimization:**

- Skips rebuild if recent success exists
- Only runs when HEAD changes
- Allows manual trigger for testing

## References

- [Trigger Orchestration](./trigger-orchestration.md) - How change-trigger.yaml works
- Workflow file: `.github/workflows/cron-full-trigger.yaml`
- Trigger workflow: `.github/workflows/change-trigger.yaml`
