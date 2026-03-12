# Show CI Results

<!-- book:cmd show ci-results -->

Displays a formatted summary of CI workflow run results, including per-workflow status, failed job names, and links for diagnostics.

## Usage

```bash
eac show ci-results [sha-or-run-id] [module...]
```

## Arguments

The first positional argument is auto-classified:

| Input | Detection |
|-------|-----------|
| 40-char hex or 7+ hex prefix | Treated as a commit SHA |
| Numeric value | Treated as a specific GitHub Actions run ID |
| Omitted | Auto-detects SHA from `GITHUB_SHA`, `origin/main`, or git HEAD |

Additional positional arguments filter to specific modules.

## Output Sections

A table with columns:

| Column | Description |
|--------|-------------|
| Workflow | Workflow name (orchestrator listed first) |
| Status | Conclusion or in-progress status; failures show `FAILED: <job>` |
| Link | URL to the run or the specific failed job |

Summary line shows the short SHA with pass/fail counts.

## Examples

```bash
# CI results for current HEAD
eac show ci-results

# CI results for a specific commit
eac show ci-results abc1234

# CI results for specific modules at a commit
eac show ci-results abc1234 core clibase

# CI results for a specific run ID
eac show ci-results 12345678
```

## See Also

- [get ci-results](../get/ci-results.md) - JSON output
- [show ci-summary](./ci-summary.md)
