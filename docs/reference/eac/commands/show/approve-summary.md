# Show approve-summary

<!-- book:cmd show approve-summary -->

Generate a release approval summary, formatted as Markdown suitable for `$GITHUB_STEP_SUMMARY`.

Outputs a check status table covering version, tag, commit, changelog validation, existing release check, and CI status. On failure with a run ID, includes a note pointing to diagnostic details.

## Usage

```
eac show approve-summary --module=<name> --version=<ver> --tag=<tag> --commit=<sha> [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--module` | string | | Module name (required) |
| `--version` | string | | Release version (required) |
| `--tag` | string | | Full tag name (required) |
| `--commit` | string | | Commit SHA (required) |
| `--version-type` | string | `semver` | Version type (`semver` or `calver`) |
| `--ci-skipped` | bool | `false` | Whether CI check was skipped |
| `--status` | string | `success` | Overall status (`success` or `failure`) |
| `--run-id` | string | | Run ID for diagnostic links on failure |

## Output Sections

- **Header**: "Release Approval: `<module>`"
- **Check table**: version (with type), tag, commit (shortened), and on success:
  - Changelog status (updated for semver, N/A for calver)
  - Existing Release check
  - CI Check (passed or skipped warning)

## Examples

```bash
# Successful semver approval
eac show approve-summary \
  --module=core --version=1.2.3 \
  --tag=core/v1.2.3 --commit=abc1234def5678

# Calver with CI skipped
eac show approve-summary \
  --module=docs --version=2025.0312 \
  --tag=docs/v2025.0312 --commit=abc1234 \
  --version-type=calver --ci-skipped

# Redirect to GitHub Actions step summary
eac show approve-summary --module=core --version=1.2.3 \
  --tag=core/v1.2.3 --commit=abc1234 >> "$GITHUB_STEP_SUMMARY"
```

## See Also

- [show](show.md)
- [release](../release/index.md)
