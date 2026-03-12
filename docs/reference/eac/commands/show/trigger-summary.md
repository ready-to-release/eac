# Show trigger-summary

<!-- book:cmd show trigger-summary -->

Generate a release trigger summary, formatted as Markdown suitable for `$GITHUB_STEP_SUMMARY`.

Outputs a formatted summary when a release workflow is triggered, showing the workflow that was dispatched and key properties like version, run IDs, branch, and commit.

## Usage

```
eac show trigger-summary --module=<name> --workflow=<file> --ci-run-id=<id> --branch=<branch> --commit=<sha> [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--module` | string | | Module name (required) |
| `--workflow` | string | | Workflow filename (required) |
| `--workflow-desc` | string | `Release module` | Workflow description |
| `--version` | string | | Version being released |
| `--trigger-run-id` | string | | Original trigger workflow run ID |
| `--ci-run-id` | string | | CI workflow run ID (required) |
| `--branch` | string | | Git branch name (required) |
| `--commit` | string | | Git commit SHA (required) |

## Output Sections

- **Header**: "Release Triggered: `<module>`"
- **Workflow table**: workflow filename and description
- **Properties table**: version (if set), trigger run ID, CI run ID, branch, and commit

## Examples

```bash
eac show trigger-summary \
  --module=core \
  --workflow=release-core.yml \
  --version=1.2.3 \
  --ci-run-id=12345678 \
  --branch=main \
  --commit=abc1234def5678

# Redirect to GitHub Actions step summary
eac show trigger-summary --module=core --workflow=release-core.yml \
  --ci-run-id=12345678 --branch=main --commit=abc1234 >> "$GITHUB_STEP_SUMMARY"
```

## See Also

- [show](show.md)
- [release](../release/index.md)
