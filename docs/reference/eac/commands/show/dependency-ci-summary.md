# Show dependency-ci-summary

<!-- book:cmd show dependency-ci-summary -->

Generate a dependency CI check summary, formatted as Markdown suitable for `$GITHUB_STEP_SUMMARY`.

Shows the results of verifying that a module's dependencies have passing CI. On success, displays counts of checked and skipped dependencies. On failure, shows a failure message.

## Usage

```
eac show dependency-ci-summary --module=<name> [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--module` | string | | Module name (required) |
| `--passed` | int | `0` | Number of dependencies that passed CI |
| `--skipped` | int | `0` | Number of dependencies skipped (no CI workflow) |
| `--status` | string | `success` | Overall status (`success` or `failure`) |

## Output Sections

**On success:**
- **Header**: "Dependency CI Check: `<module>`"
- **Status**: all checks passed
- **Metrics table**: dependencies checked count, skipped count (if any)

**On failure:**
- **Header**: "Dependency CI Check: `<module>`"
- **Status**: check failed with message to see logs

## Examples

```bash
# All dependencies passed
eac show dependency-ci-summary --module=eac --passed=5

# Some skipped (no CI workflow)
eac show dependency-ci-summary --module=eac --passed=3 --skipped=2

# Failed check
eac show dependency-ci-summary --module=eac --status=failure

# Redirect to GitHub Actions step summary
eac show dependency-ci-summary --module=eac --passed=5 >> "$GITHUB_STEP_SUMMARY"
```

## See Also

- [show](show.md)
- [dependencies](dependencies.md)
