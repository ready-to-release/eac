# Get ci-results

<!-- book:cmd get ci-results -->

Returns CI workflow run results from GitHub Actions, including job details, artifacts, and diagnostic links. Queries are enriched with job-level durations and downloadable artifact info.

## Usage

```bash
eac get ci-results [sha-or-run-id] [module...] [flags]
```

## Arguments

The first positional argument is auto-classified:

| Input | Detection |
|-------|-----------|
| 40-char hex or 7+ hex prefix | Commit SHA |
| Numeric value | GitHub Actions run ID |
| Omitted | Auto-detect (GITHUB_SHA, origin/main, or HEAD) |

Remaining positional arguments filter to specific modules.

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--as-yaml` | bool | Output as YAML (default) |
| `--as-json` | bool | Output as JSON |
| `--as-toml` | bool | Output as TOML |

## Output Structure

```yaml
head_sha: abc1234...
orchestrator:
  module: ci-orchestrator
  workflow: change-trigger.yaml
  run_id: 12345678
  status: completed
  conclusion: success
runs:
  - module: core
    workflow: ci-core.yaml
    run_id: 12345679
    status: completed
    conclusion: success
    jobs:
      - name: build
        status: completed
        conclusion: success
        duration: 2m30s
    artifacts:
      - name: core-evidence
        size_bytes: 1048576
        expired: false
    links:
      web_url: https://github.com/org/repo/actions/runs/12345679
      view_logs: gh run view 12345679 --repo org/repo --log
      view_failed_logs: gh run view 12345679 --repo org/repo --log-failed
      download_all: gh run download 12345679 --repo org/repo
total_runs: 5
passed: 4
failed: 1
```

## Examples

```bash
# Current HEAD
eac get ci-results

# Specific commit
eac get ci-results abc1234

# Specific modules at a commit
eac get ci-results abc1234 core clibase

# Specific run ID
eac get ci-results 12345678

# JSON output
eac get ci-results --as-json
```

