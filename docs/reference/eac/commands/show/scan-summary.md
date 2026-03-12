# Show scan-summary

<!-- book:cmd show scan-summary -->

Generate a security scan summary for a module, formatted as Markdown suitable for `$GITHUB_STEP_SUMMARY`.

Reads UoW manifests from `out/scan/<module>/` to derive scan status. Shows each scan type with its pass/fail/cached status and duration. If no scan manifests are found, outputs an informational message and exits successfully.

## Usage

```
eac show scan-summary <module> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module name (required) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--artifact-name` | string | Name of the artifact containing scan results (default: `scan-results-<module>`) |

## Output Sections

- **Header**: module name with lock icon
- **Status message**: all passed or some failed
- **Scan table**: one row per scan with name, status (Passed/Failed/Cached), and duration
- **Summary line**: counts of passed, failed, and skipped scans
- **Artifact**: artifact name for downloading full results

## Examples

```bash
# Generate scan summary
eac show scan-summary core

# With custom artifact name
eac show scan-summary core --artifact-name=security-scan-core

# Redirect to GitHub Actions step summary
eac show scan-summary core >> "$GITHUB_STEP_SUMMARY"
```

## See Also

- [show](show.md)
- [scan](../scan/index.md)
