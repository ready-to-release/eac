# Check CI Before Release

## What You'll Accomplish

Verify CI pipeline passes before creating a release to ensure quality.

## Prerequisites

- CI/CD configured (GitHub Actions, etc.)
- Commit to check has completed or is running

## Steps

### 1. Check Current Status

```bash
r2r eac pipeline status
```

**What happens**: Shows CI status for head of trunk

### 2. Check Specific Commit

```bash
r2r eac release check-ci $(git rev-parse HEAD)
```

**What happens**: Verifies CI passed for the commit

### 3. Wait for CI if Running

```bash
r2r eac pipeline wait
```

**What happens**: Blocks until CI workflows complete

### 4. Proceed with Release

If CI passes, continue with release:

```bash
r2r eac release this
```

## Example Scenario

Checking CI before release:

```bash
# Get current commit
COMMIT=$(git rev-parse HEAD)
echo $COMMIT
# abc123...

# Check CI status
r2r eac release check-ci $COMMIT

# Output:
# Checking CI for commit abc123...
# ✓ Build workflow: passed
# ✓ Test workflow: passed
# ✓ Security scan: passed
# ✓ All checks passed

# Safe to release
r2r eac release this

# If CI is still running:
r2r eac release check-ci $COMMIT
# ⏳ Build workflow: in progress
# ✓ Test workflow: passed

# Wait for completion
r2r eac pipeline wait
# Waiting for workflows...
# ✓ All workflows completed

# Check again
r2r eac release check-ci $COMMIT
# ✓ All checks passed
```

## CI Status Indicators

- ✓ **Passed** - Safe to release
- ✗ **Failed** - Do not release, fix issues
- ⏳ **In Progress** - Wait for completion
- ○ **Pending** - Not started yet

## Automation

```bash
# Script for safe release
#!/bin/bash
set -e

COMMIT=$(git rev-parse HEAD)

# Check CI
if ! r2r eac release check-ci $COMMIT; then
  echo "CI not passing, aborting release"
  exit 1
fi

# Proceed with release
r2r eac release this
```

## Next Steps

- [Prepare Module Release](./prepare-module-release.md) → Complete workflow

## Related Commands

- [`release check-ci`](../../../../reference/commands/release/check-ci.md) - Check CI status
- [`pipeline status`](../../../../reference/commands/pipeline/status.md) - Current status
- [`pipeline wait`](../../../../reference/commands/pipeline/wait.md) - Wait for completion
