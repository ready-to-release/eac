# test merge-results

<!-- book:cmd test merge-results -->

Merge manual test results from `test-results/<module>/<version>/manual-results.json` into the test manifest at `out/test/<module>/test.manifest.json`.

## Synopsis

```bash
test merge-results --module <module> --version <version>
```

## Description

Transforms manual test results into test entries and updates the test manifest with aggregated statistics. If no manifest exists, a new one is created. If a manual suite already exists, it is replaced with the new results.

This command integrates manual test results into the same test manifest used for automated tests, providing a unified view of all test results across automated and manual testing efforts.

## Flags

- `--module <moniker>` (required) - Module moniker to merge results for
- `--version <version>` (required) - Release version

## Input

Reads from: `test-results/<module>/<version>/manual-results.json`

This file must have been created by `test import-manual` command.

## Output

Creates or updates: `out/test/<module>/test.manifest.json`

Output message:

```text
Merged manual test results for eac-commands v1.2.0
  Location: out/test/eac-commands/test.manifest.json
  Manual tests: 5 passed, 1 failed, 0 skipped
  Total tests in manifest: 785
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

**Output** (in test.manifest.json):

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

| Source Field          | Target Field   | Transformation                           |
| --------------------- | -------------- | ---------------------------------------- |
| scenario_id           | name           | Extract last path component              |
| status                | status         | Direct copy (passed/failed/skipped)      |
| duration_seconds      | duration_ms    | Convert seconds to milliseconds          |
| error                 | error          | Direct copy (if present)                 |
| notes                 | (not stored)   | Not preserved in manifest                |
| evidence              | (not stored)   | Not preserved in manifest                |
| -                     | package        | Set to "manual"                          |
| -                     | type           | Set to "manual"                          |
| -                     | suite          | Set to "manual"                          |
| -                     | tags           | Empty array                              |
| -                     | file_path      | Empty string                             |

## Manifest Updates

### New Manifest Creation

If manifest doesn't exist:

1. Creates new manifest with metadata from import_metadata
2. Adds manual suite to suites object
3. Populates tests array with manual test entries
4. Calculates summary statistics

### Existing Manifest Update

If manifest exists:

1. Removes all existing manual test entries
2. Appends new manual test entries
3. Replaces manual suite statistics
4. Recalculates summary statistics

### Manual Suite Statistics

The `manual` suite in the manifest includes:

```json
{
  "suites": {
    "manual": {
      "run_time": "2026-01-19T12:00:00Z",
      "duration_seconds": 120.5,
      "tests": {
        "total": 6,
        "passed": 5,
        "failed": 1,
        "skipped": 0
      }
    }
  }
}
```

### Summary Updates

The manifest summary aggregates all tests (automated + manual):

```json
{
  "summary": {
    "total": 785,
    "passed": 779,
    "failed": 1,
    "skipped": 5
  }
}
```

## Examples

### Merge Manual Test Results

```bash
r2r eac test merge-results --module eac-commands --version v1.2.0
```

Output:

```text
Merged manual test results for eac-commands v1.2.0
  Location: out/test/eac-commands/test.manifest.json
  Manual tests: 5 passed, 1 failed, 0 skipped
  Total tests in manifest: 785
```

### Verify Merged Results

```bash
# View test summary
r2r eac show test-summary eac-commands

# View manual suite details
r2r eac show suite manual --module eac-commands
```

## Error Conditions

| Exit Code | Condition                                   |
| --------- | ------------------------------------------- |
| 1         | Module flag missing                         |
| 1         | Version flag missing                        |
| 1         | Manual results file not found               |
| 1         | Invalid JSON in results file                |
| 1         | Unknown module                              |
| 1         | Invalid version format                      |

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
r2r eac show test-summary eac-commands
```

Shows aggregated statistics including manual tests.

### Manual Suite Details

```bash
r2r eac show suite manual --module eac-commands
```

Shows only manual test results.

### All Tests

```bash
r2r eac show tests eac-commands
```

Lists all tests including manual entries with type "manual".

## Common Scenarios

### First-Time Merge

If no test manifest exists:

```bash
r2r eac test merge-results --module new-module --version v1.0.0
```

Creates new manifest with only manual tests.

### Update Existing Manual Results

To replace previous manual test results:

```bash
# Re-import new results
r2r eac test import-manual --input updated-results.json --release v1.2.0 --force

# Re-merge
r2r eac test merge-results --module eac-commands --version v1.2.0
```

Old manual tests are removed, new ones added.

### Merge After Automated Tests

Typical workflow:

```bash
# Run automated tests (creates manifest)
r2r eac test eac-commands

# Merge manual results (updates manifest)
r2r eac test merge-results --module eac-commands --version v1.2.0
```

Manifest now contains both automated and manual test results.

## Idempotency

The merge command is **idempotent** - running it multiple times with the same input produces the same result:

```bash
# First merge
r2r eac test merge-results --module eac-commands --version v1.2.0

# Second merge (produces identical manifest)
r2r eac test merge-results --module eac-commands --version v1.2.0
```

This is safe because:

1. All existing manual tests are removed before merge
2. New manual tests are added from current import file
3. Summary is recalculated from scratch

## Best Practices

### Run After Import

Always run merge immediately after successful import:

```bash
r2r eac test import-manual --input results.json --release v1.2.0 && \
r2r eac test merge-results --module eac-commands --version v1.2.0
```

### Version Consistency

Use the same version for export, import, and merge:

```bash
VERSION="v1.2.0"
r2r eac test export-manual --module eac-commands --release $VERSION
r2r eac test import-manual --input results.json --release $VERSION
r2r eac test merge-results --module eac-commands --version $VERSION
```

### CI Integration

In CI pipelines, merge manual results before generating test reports:

```bash
# Automated tests
r2r eac test eac-commands

# Merge manual results (if available)
if [ -f "test-results/eac-commands/v1.2.0/manual-results.json" ]; then
  r2r eac test merge-results --module eac-commands --version v1.2.0
fi

# Generate reports
r2r eac show test-summary eac-commands
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

**Solution**: Verify module moniker with `r2r eac show modules`.

### Invalid Version Format

```text
invalid version format: version must be in semver (v1.2.3) or calver (v2024.01.19) format
```

**Solution**: Use valid semver or calver format for `--version` flag.

## See Also

- [test export-manual](./export-manual.md) - Export manual test scenarios
- [test import-manual](./import-manual.md) - Import manual test results
- [show test-summary](../show/test-summary.md) - View aggregated results
- [Execute Manual Tests](../../../how-to-guides/eac/commands/build-test-validate/execute-manual-tests.md) - Full workflow guide
- [test Commands](../categories/test.md)
