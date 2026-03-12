# Show lint-summary

<!-- book:cmd show lint-summary -->

Generate a lint summary for a module, formatted as Markdown suitable for `$GITHUB_STEP_SUMMARY`.

Reads UoW manifests from `out/lint/<module>/` to derive lint status from exit codes. Shows provider names, duration, and issue counts.

## Usage

```
eac show lint-summary <module> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module name (required) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--artifact-name` | string | Name of the artifact containing lint results (default: `lint-results-<module>`) |

## Output Sections

- **Header**: pass/fail status with module name
- **Status message**: "No lint issues found" or number of providers with issues
- **Metrics table**: duration and list of lint providers that were run
- **Artifact**: artifact name for downloading full results

## Examples

```bash
# Generate lint summary
eac show lint-summary core

# With custom artifact name
eac show lint-summary core --artifact-name=lint-core-linux

# Redirect to GitHub Actions step summary
eac show lint-summary core >> "$GITHUB_STEP_SUMMARY"
```

## See Also

- [show build-summary](./build-summary.md) - Build summary
- [show scan-summary](./scan-summary.md) - Security scan summary
