# Pipeline Commands

{{ page_breadcrumb() }}

Command reference for EAC's pipeline orchestration system.

## Quick Reference

| Command                         | Description                                      |
| ------------------------------- | ------------------------------------------------ |
| `pipeline run`                  | Execute module pipelines respecting dependencies |
| `pipeline wait`                 | Wait for GitHub workflow runs to complete        |
| `pipeline status`               | Show CI status for the head of trunk             |
| `pipeline ci dispatch-and-wait` | Dispatch workflow and wait for completion        |
| `pipeline ci summary-link`      | Generate diagnostic markdown for CI summaries    |

---

## pipeline run

Execute module pipelines respecting dependencies.

### Synopsis

```bash
r2r eac pipeline run [module1] [module2] ... [options]
```

### Description

Analyzes module dependency graph and executes pipelines in the correct order. Modules are grouped into dependency layers, with parallel execution within each layer.

### Arguments

| Argument    | Required | Description                              |
| ----------- | -------- | ---------------------------------------- |
| `module...` | No       | Module monikers to run (defaults to all) |

### Flags

| Flag             | Short | Type   | Default | Description                            |
| ---------------- | ----- | ------ | ------- | -------------------------------------- |
| `--changed-only` | `-c`  | bool   | `false` | Only run pipelines for changed modules |
| `--ref`          | `-r`  | string | current | Git ref to use                         |
| `--dry-run`      | `-n`  | bool   | `false` | Show execution plan without running    |

### Examples

```bash
# Run all module pipelines
r2r eac pipeline run

# Run single module pipeline
r2r eac pipeline run eac-commands

# Run specific modules
r2r eac pipeline run r2r-cli eac-core

# Run only changed modules
r2r eac pipeline run --changed-only

# Run against specific branch
r2r eac pipeline run --ref=develop

# Preview execution plan
r2r eac pipeline run --dry-run
```

### Output

```text
Pipeline Execution Plan:

Layer 1 (2 modules - no dependencies):
  - eac-core
  - src-contracts

Layer 2 (3 modules - depends on Layer 1):
  - eac-commands
  - src-registry
  - src-repository

Layer 3 (2 modules - depends on Layer 2):
  - r2r-cli
  - src-mcp

Executing Layer 1...
  ✓ eac-core (build: 45s, test: 12s)
  ✓ src-contracts (validate: 3s)

Executing Layer 2...
  ✓ eac-commands (build: 38s, test: 15s)
  ✓ src-registry (build: 22s, test: 8s)
  ✓ src-repository (build: 31s, test: 11s)

Executing Layer 3...
  ✓ r2r-cli (build: 55s, test: 18s)
  ✓ src-mcp (build: 42s, test: 14s)

✓ All pipelines completed successfully
Total time: 3m 45s
```

### Execution Rules

1. All modules in Layer 1 run in parallel
2. Each layer waits for the previous layer to complete
3. If any module fails, subsequent layers are cancelled
4. Within each layer, modules run in parallel

### Exit Codes

| Code | Description                          |
| ---- | ------------------------------------ |
| 0    | All pipelines completed successfully |
| 1    | One or more pipelines failed         |
| 2    | Configuration or validation error    |

---

## pipeline wait

Wait for GitHub workflow runs to complete.

### Synopsis

```bash
r2r eac pipeline wait <run-id> [run-id...] [options]
```

### Description

Polls GitHub API to wait for workflow runs to complete. Displays live progress and status changes.

### Arguments

| Argument | Required | Description                  |
| -------- | -------- | ---------------------------- |
| `run-id` | Yes      | One or more workflow run IDs |

### Flags

| Flag         | Short | Type | Default | Description                  |
| ------------ | ----- | ---- | ------- | ---------------------------- |
| `--timeout`  | `-t`  | int  | `1800`  | Maximum wait time in seconds |
| `--interval` | `-i`  | int  | `10`    | Polling interval in seconds  |

### Examples

```bash
# Wait for single run
r2r eac pipeline wait 1234567890

# Wait for multiple runs
r2r eac pipeline wait 1234567890 1234567891

# Custom timeout (1 hour)
r2r eac pipeline wait 1234567890 --timeout=3600

# Poll every 30 seconds
r2r eac pipeline wait 1234567890 --interval=30
```

### Output

```text
Waiting for workflow runs to complete...

Run #1234567890: ci
  Status: in_progress → completed
  Conclusion: success
  Duration: 4m 32s

Run #1234567891: deploy
  Status: queued → in_progress → completed
  Conclusion: success
  Duration: 2m 15s

✓ All runs completed successfully
Total wait time: 5m 47s
```

### Status Progression

Workflows progress through states:

1. `queued` - Waiting to start
2. `in_progress` - Currently running
3. `completed` - Finished (check conclusion)

Conclusions:

- `success` - All jobs passed
- `failure` - One or more jobs failed
- `cancelled` - Run was cancelled
- `skipped` - Run was skipped

### Exit Codes

| Code | Description                          |
| ---- | ------------------------------------ |
| 0    | All runs completed successfully      |
| 1    | One or more runs failed or cancelled |
| 2    | Timeout reached before completion    |
| 3    | Error accessing GitHub API           |

---

## pipeline status

Show CI status for the head of trunk.

### Synopsis

```bash
r2r eac pipeline status [options]
```

### Description

Queries GitHub to display the status of all workflows for a specific commit or branch.

### Flags

| Flag       | Short | Type   | Default | Description         |
| ---------- | ----- | ------ | ------- | ------------------- |
| `--ref`    | `-r`  | string | `main`  | Branch to check     |
| `--commit` | `-c`  | string | -       | Specific commit SHA |
| `--json`   |       | bool   | `false` | Output as JSON      |

### Examples

```bash
# Status for main HEAD
r2r eac pipeline status

# Status for develop branch
r2r eac pipeline status --ref=develop

# Status for specific commit
r2r eac pipeline status --commit=abc123

# JSON output
r2r eac pipeline status --json
```

### Output

```text
Pipeline Status for main (abc1234)
═══════════════════════════════════════════════════════════════

Commit: abc1234 - feat: add new validation command
Author: developer@example.com
Date: 2025-12-01 10:30:00

Workflow Runs:
─────────────────────────────────────────────────────────────
│ Workflow       │ Status      │ Conclusion │ Duration │ URL │
├────────────────┼─────────────┼────────────┼──────────┼─────┤
│ ci             │ completed   │ success    │ 5m 23s   │ #123│
│ build-and-test │ completed   │ success    │ 3m 45s   │ #124│
│ security       │ in_progress │ -          │ 2m 10s   │ #125│

Summary:
  ✓ Successful: 2
  ⏳ In Progress: 1
  ✗ Failed: 0
```

### Exit Codes

| Code | Description                                 |
| ---- | ------------------------------------------- |
| 0    | All workflows successful (or no workflows)  |
| 1    | One or more workflows failed or in progress |
| 2    | Error retrieving status                     |

---

## Common Workflows

### CI/CD Pipeline Orchestration

```bash
# In GitHub Actions workflow
- name: Run changed module pipelines
  run: r2r eac pipeline run --changed-only

- name: Check status
  run: r2r eac pipeline status
```

### Pre-Release CI Check

```bash
# 1. Check CI status for current commit
r2r eac pipeline status --commit=$(git rev-parse HEAD)

# 2. If successful, proceed with release
if [ $? -eq 0 ]; then
  r2r eac release generate-module-calver eac-commands --create --push
fi
```

### Manual Pipeline Trigger

```bash
# 1. Trigger workflows via GitHub CLI
gh workflow run ci.yml --ref=main

# 2. Get run IDs
RUN_ID=$(gh run list --workflow=ci.yml --limit=1 --json databaseId -q '.[].databaseId')

# 3. Wait for completion
r2r eac pipeline wait $RUN_ID

# 4. Check final status
r2r eac pipeline status
```

### Dependency-Aware Execution

```bash
# Preview execution order
r2r eac pipeline run --dry-run

# Example output:
# Layer 1: eac-core, src-contracts
# Layer 2: eac-commands, src-registry
# Layer 3: r2r-cli

# Run with full dependency resolution
r2r eac pipeline run
```

### Release Automation Script

```bash
#!/bin/bash
MODULE=$1

# 1. Run pipelines for the module
r2r eac pipeline run $MODULE

# 2. Check status
r2r eac pipeline status
if [ $? -ne 0 ]; then
  echo "❌ Pipeline failed"
  exit 1
fi

# 3. Create release
r2r eac release generate-module-calver $MODULE --create --push
echo "✓ Release complete"
```

---

## Integration Patterns

### GitHub Actions Integration

```yaml
name: Dependency-Aware CI

on:
  push:
    branches: [main, develop]
  pull_request:

jobs:
  pipeline:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

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
r2r eac get changed-modules

# Run pipelines locally before pushing
r2r eac pipeline run --changed-only

# Push only if pipelines succeed
if [ $? -eq 0 ]; then
  git push origin main
fi
```

### Continuous Deployment

```yaml
name: Continuous Deployment

on:
  push:
    tags:
      - 'v*'

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Check CI status for tag
        run: |
          COMMIT=$(git rev-list -n 1 ${{ github.ref }})
          r2r eac pipeline status --commit=$COMMIT

  deploy:
    needs: verify
    runs-on: ubuntu-latest
    steps:
      - name: Deploy
        run: echo "Deploying..."
```

---

## Dependency Layer Execution

### How Layers Work

The pipeline run command analyzes module dependencies:

```text
Layer 1 (Foundation - no dependencies):
├── eac-core
└── src-contracts

Layer 2 (Core Services - depends on Layer 1):
├── eac-commands
├── src-registry
└── src-repository

Layer 3 (Integration - depends on Layer 2):
├── r2r-cli
└── src-mcp

Layer 4 (Distribution - depends on Layer 3):
└── distribution
```

### Performance Benefits

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

---

## GitHub CLI Requirements

### Installation

```bash
# Windows
winget install GitHub.cli

# macOS
brew install gh

# Linux (Debian/Ubuntu)
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
  sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] \
  https://cli.github.com/packages stable main" | \
  sudo tee /etc/apt/sources.list.d/github-cli.list
sudo apt update && sudo apt install gh
```

### Authentication

```bash
# Interactive
gh auth login

# Token
echo "YOUR_TOKEN" | gh auth login --with-token

# Verify
gh auth status
```

### Required Permissions

Token scopes needed:

- `repo` - Full repository access
- `workflow` - Workflow trigger and status

---

## Best Practices

- **Use --changed-only in CI** - Optimize execution time
- **Set appropriate timeouts** - Default 30 min, adjust as needed
- **Monitor run IDs** - Save workflow IDs for debugging
- **Check status before release** - Verify CI passes first
- **Use dependency order** - Let pipeline run handle sequencing
- **Poll efficiently** - Default 10s balances responsiveness and API limits

---

## Troubleshooting

| Problem                 | Solution                                |
| ----------------------- | --------------------------------------- |
| gh command not found    | Install GitHub CLI                      |
| Authentication failed   | Run `gh auth login`                     |
| Run not found           | Verify run ID with `gh run list`        |
| Timeout reached         | Increase `--timeout` or check for hangs |
| Rate limit exceeded     | Increase `--interval`                   |
| Dependency cycle        | Fix circular dependencies in contracts  |
| Pipeline failed         | Check module logs in GitHub Actions     |
| Status shows old commit | Verify correct branch/ref               |

---

## Pipeline CI Commands

For CI orchestration and diagnostic commands, see the dedicated [Pipeline CI Commands](pipeline-ci-commands.md) reference which covers:

- `pipeline ci dispatch-and-wait` - Dispatch workflow and wait for completion
- `pipeline ci summary-link` - Generate diagnostic markdown for CI summaries

---

## Related Documentation

- [Pipeline Overview](pipeline-overview.md) - Concepts and architecture
- [Pipeline Configuration](pipeline-configuration.md) - Configuration reference
- [Pipeline CI Commands](pipeline-ci-commands.md) - CI orchestration and diagnostics
- [Release Commands](release-commands.md) - CI integration for releases

{{ diataxis_footer() }}
