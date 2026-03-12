# Pipeline download-evidence-artifacts

<!-- book:cmd pipeline download-evidence-artifacts -->

Downloads test results and scan artifacts from CI runs for a module and all its dependencies. Uses `get evidence-ci-runs` internally to determine which CI runs to download from.

For each module in the dependency chain that has a `ci-{module}.yaml` workflow:

- Downloads `test-results-{module}*` artifacts to `{output-dir}/test/`
- Downloads `scan-results-{module}` artifacts to `{output-dir}/scan/`

Fails if any dependency with a CI workflow has no successful run.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output-dir` | string | `out` | Base output directory (test/ and scan/ subdirs are created) |

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<module>` | Yes | Module moniker to download evidence artifacts for |

## Output

- Per-module download progress with artifact counts
- Summary with total modules processed, test artifacts, and scan artifacts
- Exit code 0 if all downloads succeed, 1 on errors

## Examples

```bash
eac pipeline download-evidence-artifacts clie
eac pipeline download-evidence-artifacts eac-ext --output-dir custom/path
```

## See Also

- [get evidence-ci-runs](../get/evidence-ci-runs.md) - Get CI run IDs for evidence
- [update evidence](../update/evidence.md) - Build evidence documentation
- [pipeline](./index.md)
