# test merge-results

<!-- book:cmd test merge-results -->

## Synopsis

```bash
test merge-results --module <module> --version <version>
```

## Description

Transforms manual test results into test entries and writes them as a UoW manifest with a `manual-tests.json` artifact. If a manual UoW already exists, it is replaced with the new results.

This command integrates manual test results into the same UoW-based output structure used for automated tests, providing a unified view of all test results across automated and manual testing efforts.

## Input

Reads from: `test-results/<module>/<version>/manual-results.json`

This file must have been created by `test import-manual` command.

## Output

Creates or updates: `out/test/<module>/manual-merge/uow.manifest.json` and `manual-tests.json`

Output message:

```text
Merged manual test results for eac-commands v1.2.0
  Location: out/test/eac-commands/manual-merge/
  Manual tests: 5 passed, 1 failed, 0 skipped
```

## Transformation Logic

Each `ManualTestResult` becomes a `TestEntry` in the manifest:

**Input** (from manual-results.json):

```json
{
  "scenario_id": "eac-commands/feature1/manual-scenario",
  "status": "passed",
  "duration_seconds": 30.2,
  "notes": "Test completed successfully",
  "error": ""
}
```

**Output** (in manual-tests.json, referenced by uow.manifest.json):

```json
{
  "name": "manual-scenario",
  "package": "manual",
  "type": "manual",
  "suite": "manual",
  "status": "passed",
  "duration_ms": 30200,
  "tags": [],
  "file_path": "",
  "error": ""
}
```

### Field Mappings

| Source Field     | Target Field | Transformation                      |
| ---------------- | ------------ | ----------------------------------- |
| scenario_id      | name         | Extract last path component         |
| status           | status       | Direct copy (passed/failed/skipped) |
| duration_seconds | duration_ms  | Convert seconds to milliseconds     |
| error            | error        | Direct copy (if present)            |
| notes            | (not stored) | Not preserved in manifest           |
| evidence         | (not stored) | Not preserved in manifest           |
| -                | package      | Set to "manual"                     |
| -                | type         | Set to "manual"                     |
| -                | suite        | Set to "manual"                     |
| -                | tags         | Empty array                         |
| -                | file_path    | Empty string                        |

## UoW Manifest Output

The command writes a UoW manifest at `out/test/<module>/manual-merge/uow.manifest.json` with a `manual-tests.json` artifact containing the manual test entries. The `testview` aggregation system reads this alongside automated test UoW manifests to provide unified test summaries.

## Error Conditions

| Exit Code | Condition                     |
| --------- | ----------------------------- |
| 1         | Module flag missing           |
| 1         | Version flag missing          |
| 1         | Manual results file not found |
| 1         | Invalid JSON in results file  |
| 1         | Unknown module                |
| 1         | Invalid version format        |

## Version Format Validation

Accepts both semver and calver formats:

**Semver**: `v1.2.3`, `v1.2.3-alpha.1`, `v2.0.0-rc.1`

**Calver**: `v2024.01.19`, `v2026.12.31-hotfix`

## Workflow Integration

This command is the fourth and final step in the manual testing workflow:

1. **Export** → `test export-manual` generates scenarios
2. **Execute** → Human tester fills in results
3. **Import** → `test import-manual` validates and stores results
4. **Merge** → `test merge-results` adds to test manifest ← **You are here**

## Viewing Merged Results

### Test Summary

```bash
eac show test-summary eac-commands
```

Shows aggregated statistics including manual tests.

### Manual Suite Details

```bash
eac show suite manual --module eac-commands
```

Shows only manual test results.

### All Tests

```bash
eac show tests eac-commands
```

Lists all tests including manual entries with type "manual".

## Common Scenarios

### First-Time Merge

If no test manifest exists:

```bash
eac test merge-results --module new-module --version v1.0.0
```

Creates new manifest with only manual tests.

### Update Existing Manual Results

To replace previous manual test results:

```bash
# Re-import new results
eac test import-manual --input updated-results.json --release v1.2.0 --force

# Re-merge
eac test merge-results --module eac-commands --version v1.2.0
```

Old manual tests are removed, new ones added.

### Merge After Automated Tests

Typical workflow:

```bash
# Run automated tests (creates manifest)
eac test eac-commands

# Merge manual results (updates manifest)
eac test merge-results --module eac-commands --version v1.2.0
```

Manifest now contains both automated and manual test results.

## Idempotency

The merge command is **idempotent** - running it multiple times with the same input produces the same result:

```bash
# First merge
eac test merge-results --module eac-commands --version v1.2.0

# Second merge (produces identical manifest)
eac test merge-results --module eac-commands --version v1.2.0
```

This is safe because:

1. All existing manual tests are removed before merge
2. New manual tests are added from current import file
3. Summary is recalculated from scratch

## Best Practices

### Run After Import

Always run merge immediately after successful import:

```bash
eac test import-manual --input results.json --release v1.2.0 && \
eac test merge-results --module eac-commands --version v1.2.0
```

### Version Consistency

Use the same version for export, import, and merge:

```bash
VERSION="v1.2.0"
eac test export-manual --module eac-commands --release $VERSION
eac test import-manual --input results.json --release $VERSION
eac test merge-results --module eac-commands --version $VERSION
```

### CI Integration

In CI pipelines, merge manual results before generating test reports:

```bash
# Automated tests
eac test eac-commands

# Merge manual results (if available)
if [ -f "test-results/eac-commands/v1.2.0/manual-results.json" ]; then
  eac test merge-results --module eac-commands --version v1.2.0
fi

# Generate reports
eac show test-summary eac-commands
```

## Troubleshooting

### Manual Results File Not Found

```text
manual results file not found: test-results/eac-commands/v1.2.0/manual-results.json
```

**Solution**: Run `test import-manual` first to create the results file.

### Invalid JSON

```text
parsing manual results JSON: invalid character '}' looking for beginning of object key string
```

**Solution**: Check JSON syntax in manual-results.json file.

### Unknown Module

```text
unknown module: eac-commands-typo
```

**Solution**: Verify module moniker with `eac show modules`.

### Invalid Version Format

```text
invalid version format: version must be in semver (v1.2.3) or calver (v2024.01.19) format
```

**Solution**: Use valid semver or calver format for `--version` flag.

## See Also

- [test export-manual](./export-manual.md) - Export manual test scenarios
- [test import-manual](./import-manual.md) - Import manual test results
- [show test-summary](../show/test-summary.md) - View aggregated results
- [Execute Manual Tests](../../../../how-to-guides/eac/commands/build-test-validate/execute-manual-tests.md) - Full workflow guide
- [test Commands](../test/index.md)
