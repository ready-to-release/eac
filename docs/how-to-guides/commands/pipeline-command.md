# Pipeline Command

**Problem**: Orchestrating GitHub Actions workflows across modules requires respecting dependencies, tracking status, and waiting for completion.

**Solution**: Use `pipeline` commands to run workflows in dependency order, check CI status, and wait for completion with live progress.

## Key Benefits

- Respects module dependency order for safe execution
- Parallel execution within dependency layers
- Live progress display with GitHub CLI integration
- Status tracking for trunk commits
- Workflow run monitoring with timeout support

## Quick Start

```bash
# Run pipelines for changed modules only
r2r eac pipeline run --changed-only

# Run pipelines for specific modules
r2r eac pipeline run src-commands src-core

# Check CI status for trunk HEAD
r2r eac pipeline status

# Check status for specific ref
r2r eac pipeline status --ref=develop

# Wait for workflow runs to complete
r2r eac pipeline wait 1234567890 1234567891

# Wait with custom timeout
r2r eac pipeline wait 1234567890 --timeout=3600
```

## Command Reference

### pipeline run

Execute module pipelines respecting dependencies.

```bash
r2r eac pipeline run [module1] [module2] ... [options]

# Options:
--changed-only         # Only run pipelines for changed modules
--ref <branch>         # Git ref to use (default: current branch)

# Examples:
r2r eac pipeline run                      # Run all module pipelines
r2r eac pipeline run src-commands         # Run single module pipeline
r2r eac pipeline run src-cli src-core     # Run specific modules
r2r eac pipeline run --changed-only       # Run only changed modules
r2r eac pipeline run --ref=develop        # Run against develop branch
```

**How it works:**

1. Analyzes module dependency graph
2. Groups modules into dependency layers
3. Executes pipelines in parallel within each layer
4. Waits for each layer to complete before starting next
5. Reports success/failure for each module

**Exit codes:**

- `0` - All pipelines completed successfully
- `1` - One or more pipelines failed
- `2` - Configuration or validation error

### pipeline status

Show CI status for the head of trunk.

```bash
r2r eac pipeline status [options]

# Options:
--ref <branch>         # Branch to check (default: main)
--commit <sha>         # Specific commit SHA to check

# Examples:
r2r eac pipeline status                   # Status for main HEAD
r2r eac pipeline status --ref=develop     # Status for develop HEAD
r2r eac pipeline status --commit=abc123   # Status for specific commit
```

**Output includes:**

- Commit SHA and message
- Workflow run status (queued, in_progress, completed)
- Conclusion (success, failure, cancelled, skipped)
- Run ID and URL
- Duration and timestamps

**Exit codes:**

- `0` - All workflows successful or no workflows found
- `1` - One or more workflows failed or in progress
- `2` - Error retrieving status

### pipeline wait

Wait for GitHub workflow runs to complete.

```bash
r2r eac pipeline wait <run-id> [run-id...] [options]

# Options:
--timeout <seconds>    # Maximum wait time in seconds (default: 1800)
--interval <seconds>   # Polling interval in seconds (default: 10)

# Examples:
r2r eac pipeline wait 1234567890                    # Wait for single run
r2r eac pipeline wait 1234567890 1234567891         # Wait for multiple runs
r2r eac pipeline wait 1234567890 --timeout=3600     # Custom timeout (1 hour)
r2r eac pipeline wait 1234567890 --interval=30      # Poll every 30 seconds
```

**How it works:**

1. Polls GitHub API at specified interval
2. Displays live progress for each run
3. Shows status changes (queued → in_progress → completed)
4. Exits when all runs complete or timeout reached
5. Reports final status for each run

**Exit codes:**

- `0` - All runs completed successfully
- `1` - One or more runs failed or were cancelled
- `2` - Timeout reached before completion
- `3` - Error accessing GitHub API

## Typical Workflows

### CI/CD Pipeline Orchestration

```bash
# In GitHub Actions workflow
name: Multi-Module Pipeline

on: [push, pull_request]

jobs:
  orchestrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run changed module pipelines
        run: r2r eac pipeline run --changed-only

      - name: Check status
        run: r2r eac pipeline status
```

### Pre-Release CI Check

```bash
# Before creating a release tag
# 1. Check CI status for current commit
r2r eac pipeline status --commit=$(git rev-parse HEAD)

# 2. If successful, proceed with release
if [ $? -eq 0 ]; then
  r2r eac release calver src-cli
else
  echo "CI checks not passing, aborting release"
  exit 1
fi
```

### Manual Pipeline Trigger

```bash
# 1. Trigger workflows via GitHub CLI
gh workflow run build.yml --ref=main

# 2. Get run IDs
RUN_IDS=$(gh run list --workflow=build.yml --limit=1 --json databaseId -q '.[].databaseId')

# 3. Wait for completion
r2r eac pipeline wait $RUN_IDS --timeout=3600

# 4. Check final status
r2r eac pipeline status
```

### Dependency-Aware Execution

```bash
# Run pipelines respecting module dependencies
r2r eac pipeline run

# Example execution order:
# Layer 1 (no dependencies): src-core, src-contracts
# Layer 2 (depends on Layer 1): src-commands, src-registry
# Layer 3 (depends on Layer 2): src-cli
```

## Integration Patterns

### GitHub Actions Integration

```yaml
name: Dependency-Aware CI

on:
  push:
    branches: [main, develop]
  pull_request:

jobs:
  pipeline-orchestration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Full history for change detection

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install GitHub CLI
        run: |
          type -p gh >/dev/null || (
            curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
            sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg && \
            echo "deb [signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] \
            https://cli.github.com/packages stable main" | \
            sudo tee /etc/apt/sources.list.d/github-cli.list && \
            sudo apt update && sudo apt install gh
          )

      - name: Authenticate GitHub CLI
        run: echo "${{ secrets.GITHUB_TOKEN }}" | gh auth login --with-token

      - name: Run pipelines for changed modules
        run: r2r eac pipeline run --changed-only

      - name: Verify CI status
        run: r2r eac pipeline status --commit=${{ github.sha }}
```

### Local Development

```bash
# Check what pipelines would run
r2r eac get-changed-modules

# Run pipelines locally before pushing
r2r eac pipeline run --changed-only

# Push only if pipelines succeed
if [ $? -eq 0 ]; then
  git push origin main
else
  echo "Pipelines failed, fix issues before pushing"
fi
```

### Release Automation

```bash
#!/bin/bash
# scripts/release.sh

MODULE=$1
VERSION_TYPE=${2:-calver}

# 1. Ensure we're on main
git checkout main
git pull origin main

# 2. Check CI status
echo "Checking CI status..."
r2r eac pipeline status
if [ $? -ne 0 ]; then
  echo "❌ CI checks not passing"
  exit 1
fi

# 3. Create release
echo "Creating release..."
r2r eac release $VERSION_TYPE $MODULE

# 4. Push tag
git push origin --tags

echo "✅ Release complete"
```

### Continuous Deployment

```yaml
name: Continuous Deployment

on:
  push:
    tags:
      - 'v*'

jobs:
  verify-ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Check CI status for tag
        run: |
          COMMIT=$(git rev-list -n 1 ${{ github.ref }})
          r2r eac pipeline status --commit=$COMMIT

  deploy:
    needs: verify-ci
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to production
        run: echo "Deploying..."
```

## Dependency Layer Execution

### How Layers Work

The pipeline run command analyzes module dependencies and executes in layers:

```text
Layer 1 (Foundation - no dependencies):
├── src-core
└── src-contracts

Layer 2 (Core Services - depends on Layer 1):
├── src-commands
├── src-registry
└── src-repository

Layer 3 (Integration - depends on Layer 2):
├── src-cli
└── src-mcp

Layer 4 (Distribution - depends on Layer 3):
└── distribution
```

**Execution rules:**

1. All modules in Layer 1 run in parallel
2. Layer 2 starts only after Layer 1 completes successfully
3. If any module fails, subsequent layers are cancelled
4. Within each layer, modules run in parallel for speed

### Example Output

```text
Pipeline Execution Plan:

Layer 1 (2 modules):
  - src-core
  - src-contracts

Layer 2 (3 modules):
  - src-commands
  - src-registry
  - src-repository

Layer 3 (2 modules):
  - src-cli
  - src-mcp

Executing Layer 1...
  ✓ src-core (build: 45s, test: 12s)
  ✓ src-contracts (validate: 3s)

Executing Layer 2...
  ✓ src-commands (build: 38s, test: 15s)
  ✓ src-registry (build: 22s, test: 8s)
  ✓ src-repository (build: 31s, test: 11s)

Executing Layer 3...
  ✓ src-cli (build: 55s, test: 18s)
  ✓ src-mcp (build: 42s, test: 14s)

✅ All pipelines completed successfully
Total time: 3m 45s
```

## GitHub CLI Requirements

### Installation

**Windows:**

```bash
winget install GitHub.cli
```

**macOS:**

```bash
brew install gh
```

**Linux:**

```bash
# Debian/Ubuntu
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
  sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] \
  https://cli.github.com/packages stable main" | \
  sudo tee /etc/apt/sources.list.d/github-cli.list
sudo apt update && sudo apt install gh
```

### Authentication

```bash
# Interactive authentication
gh auth login

# Token authentication
echo "YOUR_TOKEN" | gh auth login --with-token

# Verify authentication
gh auth status
```

### Required Permissions

The GitHub token needs these scopes:

- `repo` - Full repository access
- `workflow` - Workflow access for triggering and checking runs

## Best Practices

- **Use --changed-only in CI**: Optimize pipeline execution time
- **Set appropriate timeouts**: Default 1800s (30 min), adjust based on workflow duration
- **Monitor run IDs**: Save workflow run IDs for debugging
- **Check status before release**: Verify CI passes before creating tags
- **Use dependency order**: Let pipeline run handle execution sequence
- **Poll efficiently**: Default 10s interval balances responsiveness and API rate limits

## Troubleshooting

| Problem | Solution |
|---------|----------|
| gh command not found | Install GitHub CLI: `winget install GitHub.cli` or `brew install gh` |
| Authentication failed | Run `gh auth login` to authenticate |
| Run not found | Verify run ID is correct with `gh run list` |
| Timeout reached | Increase `--timeout` value or check workflow for hangs |
| Rate limit exceeded | Increase `--interval` to reduce API calls |
| Dependency cycle detected | Fix circular dependencies in module contracts |
| Pipeline failed | Check individual module logs in GitHub Actions |
| Status shows old commit | Ensure you're checking the correct branch/ref |

## Advanced Usage

### Custom Pipeline Configurations

```bash
# Run specific modules in custom order
r2r eac get-execution-order src-cli | while read module; do
  r2r eac pipeline run $module
done

# Run with different refs
r2r eac pipeline run --ref=feature/new-api
r2r eac pipeline status --ref=feature/new-api

# Parallel execution of independent modules
r2r eac pipeline run src-docs src-contracts  # No dependencies, runs parallel
```

### Monitoring and Logging

```bash
# Capture pipeline output
r2r eac pipeline run --changed-only 2>&1 | tee pipeline.log

# Check status and save to file
r2r eac pipeline status --ref=main > status.txt

# Wait and log progress
r2r eac pipeline wait 1234567890 --interval=5 2>&1 | tee wait.log
```

### Integration with Other Tools

```bash
# Slack notification on completion
r2r eac pipeline wait $RUN_ID && \
  curl -X POST -H 'Content-type: application/json' \
  --data '{"text":"Pipeline completed successfully"}' \
  $SLACK_WEBHOOK_URL

# Email notification
r2r eac pipeline status || \
  echo "Pipeline failed" | mail -s "CI Alert" admin@example.com

# Custom reporting
r2r eac pipeline status --ref=main | \
  jq '.workflows[] | select(.conclusion != "success")' | \
  # ... process failed workflows
```

## Exit Codes

All pipeline commands use consistent exit codes:

| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | All operations completed successfully |
| 1 | Failure | One or more pipelines/runs failed |
| 2 | Error | Configuration error, validation failure, or API error |
| 3 | Timeout | Wait command exceeded timeout (pipeline wait only) |

## Performance Considerations

### Parallel Execution

```text
Sequential execution: ~15 minutes
Layer 1: 5 min
Layer 2: 7 min (waits for Layer 1)
Layer 3: 3 min (waits for Layer 2)

Parallel execution within layers: ~8 minutes
Layer 1: 5 min (all modules parallel)
Layer 2: 2 min (all modules parallel)
Layer 3: 1 min (all modules parallel)
```

### API Rate Limits

GitHub API rate limits:

- Authenticated: 5,000 requests/hour
- Default polling (10s interval): 360 requests/hour per workflow
- Safe for monitoring up to 10 concurrent workflows

Reduce API calls:

```bash
# Increase polling interval
r2r eac pipeline wait $RUN_ID --interval=30  # 120 requests/hour
```

## Summary

**Pipeline orchestration workflow:**

1. **Plan**: `r2r eac get-changed-modules` - Identify affected modules
2. **Execute**: `r2r eac pipeline run --changed-only` - Run in dependency order
3. **Monitor**: `r2r eac pipeline wait <run-id>` - Track progress
4. **Verify**: `r2r eac pipeline status` - Confirm success
5. **Release**: Proceed with deployment if all checks pass

**Key features:**

- Dependency-aware execution prevents breaking changes
- Parallel execution within layers maximizes speed
- Live progress tracking provides real-time feedback
- GitHub CLI integration ensures reliable workflow control
- Consistent exit codes enable CI/CD automation

Use `pipeline run` for orchestration, `pipeline status` for verification, and `pipeline wait` for monitoring.
