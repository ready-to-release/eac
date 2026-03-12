# Show ci-summary

<!-- book:cmd show ci-summary -->

Generate a CI workflow summary for a module, formatted as Markdown suitable for `$GITHUB_STEP_SUMMARY`.

Outputs a job results table covering build, test, scan, and evidence steps. The overall status is derived from individual job results -- build and test failures cause an overall failure, while scan failures are treated as warnings.

## Usage

```
eac show ci-summary [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--build` | string | | Build job result (`success`, `failure`, `skipped`) |
| `--container` | bool | `false` | Whether this is a container module |
| `--container-test` | string | | Container test result (container modules) |
| `--container-test-enabled` | bool | `false` | Whether container test was enabled |
| `--test-linux` | string | | Linux test result (binary modules) |
| `--test-windows` | string | | Windows test result (binary modules) |
| `--test-macos` | string | | macOS test result (binary modules) |
| `--test-on-windows` | bool | `false` | Whether Windows tests were enabled |
| `--test-on-macos` | bool | `false` | Whether macOS tests were enabled |
| `--scan` | string | | Security scan result |
| `--scans-enabled` | bool | `false` | Whether scans were enabled |
| `--evidence` | string | | Evidence build result |
| `--evidence-enabled` | bool | `false` | Whether evidence building was enabled |

## Output Sections

- **Header**: overall pass/fail status
- **Job Results Table**: one row per enabled job (Build, Test per platform, Container Test, Security Scan, Evidence) with status icons

For binary modules, test rows are shown per platform (Linux, Windows, macOS). For container modules, a single Container Test row is shown instead.

## Examples

```bash
# Binary module with multi-platform tests
eac show ci-summary \
  --build=success \
  --test-linux=success \
  --test-on-windows --test-windows=success \
  --scans-enabled --scan=success

# Container module
eac show ci-summary \
  --build=success \
  --container \
  --container-test-enabled --container-test=success

# Redirect to GitHub Actions step summary
eac show ci-summary --build=success --test-linux=success >> "$GITHUB_STEP_SUMMARY"
```

## See Also

- [show](show.md)
- [pipeline](../pipeline/index.md)
