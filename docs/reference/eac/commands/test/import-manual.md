# test import-manual

<!-- book:cmd test import-manual -->

Import manual test results from a JSON file and validate against schemas, exported scenarios, and repository configuration.

## Synopsis

```bash
test import-manual --input <file> --release <version> [--force]
```

## Description

Validates manual test results, checks for conflicts, and stores them in the repository for later merging into test manifests. This command ensures that manual test results are properly formatted, reference valid scenarios, and contain required metadata.

The import process performs comprehensive validation to maintain data integrity and traceability:

- JSON schema compliance
- Release version matching
- Module validation
- Scenario ID cross-validation against exports
- Conflict detection (existing results)
- Email format validation
- Required field checking

## Flags

- `--input <file>` (required) - Path to manual test results JSON file
- `--release <version>` (required) - Release version being tested
- `--force` - Overwrite existing results if present

## Input Format

JSON file validated against `contracts/eac-core/0.1.0/manual-test-results.schema.json`:

```json
{
  "import_metadata": {
    "test_time": "2026-01-19T12:00:00Z",
    "tester": "tester@example.com",
    "module": "eac-commands",
    "release_version": "v1.2.0",
    "duration_seconds": 120.5,
    "schema_version": "1.0"
  },
  "results": [
    {
      "scenario_id": "eac-commands/feature1/manual-scenario",
      "status": "passed",
      "duration_seconds": 30.2,
      "notes": "Test completed successfully",
      "error": "Error message if failed",
      "evidence": [
        {
          "url": "https://example.com/screenshot.png",
          "type": "screenshot",
          "description": "UI screenshot",
          "sha256": "optional-hash"
        }
      ]
    }
  ]
}
```

### Required Fields

**import_metadata**:

- `test_time` - ISO 8601 timestamp when tests were executed
- `tester` - Email of person who executed tests
- `module` - Module moniker being tested
- `release_version` - Release version tested (must match `--release` flag)
- `schema_version` - Schema version (currently "1.0")

**results** (each entry):

- `scenario_id` - Must match exported scenario ID
- `status` - One of: `passed`, `failed`, `skipped`
- `error` - Required if status is `failed`

### Optional Fields

**import_metadata**:

- `duration_seconds` - Total time spent executing manual tests

**results** (each entry):

- `duration_seconds` - Time spent executing this scenario
- `notes` - Tester observations, failure reasons, or skip rationale (max 5000 chars)
- `evidence` - Array of evidence artifacts

**evidence**:

- `url` - HTTP(S) URL to evidence (required)
- `type` - One of: `screenshot`, `log`, `recording`, `document`, `issue` (required)
- `description` - Brief description (max 200 chars)
- `sha256` - SHA-256 hash for integrity verification (64 hex chars)

## Status Values

- **passed** - Test executed successfully, expected behavior observed
- **failed** - Test did not pass, requires `error` field with explanation
- **skipped** - Test not executed, use `notes` to explain why

Note: `pending` and `undefined` statuses are not allowed for manual tests - all manual tests must be executed.

## Output

Stores results at: `test-results/<module>/<version>/manual-results.json`

Example: `test-results/eac-commands/v1.2.0/manual-results.json`

## Validation Process

### 1. JSON Schema Validation

Validates structure against `manual-test-results.schema.json`:

- Required fields present
- Data types correct
- String lengths within limits
- Negative durations rejected
- Empty results array rejected

### 2. Metadata Validation

- **Email format**: Validates tester email with regex pattern
- **Required fields**: Ensures all mandatory fields present
- **Release version**: Must match `--release` flag value

### 3. Module Validation

- Checks module exists in repository configuration
- Ensures all scenario IDs reference same module
- Detects mixed modules in results (not allowed)

### 4. Scenario ID Cross-Validation

If export file exists at `manual-test-exports/<module>/<version>.json`:

- Validates each scenario_id exists in export
- Prevents results for non-existent scenarios
- Ensures traceability to original scenarios

### 5. Conflict Detection

- Checks if results already exist for module/version
- Blocks import without `--force` flag
- Preserves existing results unless forced

## Examples

### Import Manual Test Results

```bash
eac test import-manual --input results.json --release v1.2.0
```

Output:

```text
Imported manual test results for eac-commands v1.2.0
  Location: test-results/eac-commands/v1.2.0/manual-results.json
  Tests: 5 passed, 1 failed, 0 skipped
```

### Overwrite Existing Results

```bash
eac test import-manual --input results.json --release v1.2.0 --force
```

Output:

```text
Imported manual test results for eac-commands v1.2.0
  Location: test-results/eac-commands/v1.2.0/manual-results.json
  Tests: 5 passed, 1 failed, 0 skipped
```

## Error Conditions

| Exit Code | Condition                                    |
| --------- | -------------------------------------------- |
| 1         | Input flag missing                           |
| 1         | Release flag missing                         |
| 1         | Input file not found                         |
| 1         | Invalid JSON format                          |
| 1         | Schema validation failed                     |
| 1         | Required field missing (e.g., tester)        |
| 1         | Invalid email format                         |
| 1         | Release version mismatch                     |
| 1         | Unknown module                               |
| 1         | Mixed modules in results                     |
| 1         | Scenario ID not found in export              |
| 1         | Results already exist (without --force)      |
| 1         | Failed status without error message          |
| 1         | Invalid status value                         |

## Common Validation Errors

### Schema Validation Failed

```text
schema validation failed: validation failed: jsonschema validation failed
- at '/results/0': missing property 'error'
```

**Solution**: Add error message for failed test:

```json
{
  "scenario_id": "...",
  "status": "failed",
  "error": "Expected behavior not observed"
}
```

### Release Version Mismatch

```text
release version mismatch: file has v1.0.0, flag specifies v1.2.0
```

**Solution**: Ensure file's `release_version` matches `--release` flag.

### Scenario ID Not Found

```text
scenario not found in export: eac-commands/feature1/nonexistent-scenario
```

**Solution**: Use scenario IDs from export file or regenerate export.

### Results Already Exist

```text
manual test results already exist for eac-commands v1.2.0
  Location: test-results/eac-commands/v1.2.0/manual-results.json
  To overwrite, use --force flag
```

**Solution**: Use `--force` flag to overwrite or delete existing file.

## Workflow Integration

This command is the third step in the manual testing workflow:

1. **Export** → `test export-manual` generates scenarios
2. **Execute** → Human tester fills in results
3. **Import** → `test import-manual` validates and stores results ← **You are here**
4. **Merge** → `test merge-results` adds to test manifest

## Best Practices

### Tester Email

Use organization email address for traceability:

```json
"tester": "jane.smith@company.com"
```

### Evidence Collection

Include evidence for failed and critical tests:

```json
"evidence": [
  {
    "url": "https://github.com/org/repo/issues/123",
    "type": "issue",
    "description": "Bug report for login failure"
  },
  {
    "url": "https://example.com/screenshots/login-error.png",
    "type": "screenshot",
    "description": "Error message displayed to user"
  }
]
```

### Notes Field

Use notes to provide context:

```json
"notes": "Test passed but UI response time was slower than expected (3s vs 1s target)"
```

### Duration Tracking

Record accurate durations for workload estimation:

```json
"duration_seconds": 45.5,
"import_metadata": {
  "duration_seconds": 180.0
}
```

## See Also

- [Manual Testing Reference](../../testing/manual-tests.md) - Complete technical reference
- [test export-manual](./export-manual.md) - Export manual test scenarios
- [test merge-results](./merge-results.md) - Merge results into manifest
- [Execute Manual Tests](../../../../how-to-guides/eac/commands/build-test-validate/execute-manual-tests.md) - Full workflow guide
- [test Commands](../../categories/test.md)
