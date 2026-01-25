# test export-manual

<!-- book:cmd test export-manual -->

Export manual test scenarios (tagged with `@Manual`) from Gherkin specifications for human execution and evidence collection.

## Synopsis

```bash
test export-manual --module <module> --release <version> [--format <format>]
```

## Description

Scans module specifications, extracts scenarios tagged with `@Manual`, generates stable scenario IDs, and exports them in JSON, CSV, or Markdown format. The exported file includes scenario metadata (name, tags, steps), feature context, and release information for traceability.

Manual test scenarios are tests that require human verification, judgment, or interaction that cannot be easily automated. Examples include:

- UI/UX verification
- Physical hardware testing
- Accessibility testing
- Exploratory testing
- Regulatory compliance testing

## Flags

- `--module <moniker>` (required) - Module moniker to export manual tests from
- `--release <version>` (required) - Release version being tested
- `--format <format>` - Export format: `json`, `csv`, `markdown` (default: `json`)

## Output

Creates file in current directory:

- JSON: `manual-test-scenarios.json`
- CSV: `manual-test-scenarios.csv`
- Markdown: `manual-test-scenarios.md`

## Export Formats

### JSON (Canonical Format)

Validated against `contracts/eac-core/0.1.0/manual-test-export.schema.json`

```json
{
  "export_metadata": {
    "export_time": "2026-01-19T12:00:00Z",
    "module": "eac-commands",
    "release_version": "v1.2.0",
    "git_commit": "a1b2c3d...",
    "schema_version": "1.0"
  },
  "scenarios": [
    {
      "scenario_id": "eac-commands/feature1/manual-test-scenario",
      "feature_name": "eac-commands_feature1",
      "scenario_name": "Manual test scenario",
      "tags": ["@Manual", "@L2", "@ov"],
      "steps": [
        "Given a precondition",
        "When I perform an action",
        "Then I expect a result"
      ],
      "description": "Optional description",
      "file_path": "specs/eac-commands/feature1/spec.feature"
    }
  ]
}
```

### CSV Format

Comma-separated values with snake_case headers:

```csv
scenario_id,feature_name,scenario_name,tags,steps,description,file_path
eac-commands/feature1/manual-test-scenario,eac-commands_feature1,Manual test scenario,@Manual @L2 @ov,Given... | When... | Then...,,specs/.../spec.feature
```

### Markdown Format

Human-readable format with H2 sections per scenario:

```markdown
# Manual Test Scenarios

**Module**: eac-commands
**Release**: v1.2.0
**Exported**: 2026-01-19T12:00:00Z
**Git Commit**: a1b2c3d...

---

## Scenario: Manual test scenario

**Scenario ID**: `eac-commands/feature1/manual-test-scenario`

**Feature**: eac-commands_feature1

**Tags**: @Manual @L2 @ov

**Steps**:
- Given a precondition
- When I perform an action
- Then I expect a result

*Source*: `specs/eac-commands/feature1/spec.feature`

---
```

## Scenario ID Format

**Pattern**: `<module>/<feature>/<scenario-slug>`

**Example**: `eac-commands/authentication/login-with-valid-credentials`

**Generation Logic**:

1. Extract feature name from file path: `specs/eac-commands/feature1/spec.feature` → `feature1`
2. Slugify scenario name: `"Manual Test Scenario: With Special!"` → `"manual-test-scenario-with-special"`
3. Combine: `eac-commands/feature1/manual-test-scenario`

## Tag Extraction

- Extracts tags from both feature-level and scenario-level
- Automatically deduplicates combined tags
- Preserves tag order (feature tags first, then scenario tags)

## Schema Validation

JSON exports are automatically validated against the schema:

- Required fields present (export_metadata, scenarios)
- Git commit SHA format (40 hex chars)
- Scenario ID format (module/feature/scenario-slug)
- Minimum 1 scenario
- All scenario fields present

## Examples

### Export as JSON

```bash
r2r eac test export-manual --module eac-commands --release v1.2.0
```

Output: `manual-test-scenarios.json` in current directory

### Export as CSV

```bash
r2r eac test export-manual --module eac-commands --release v1.2.0 --format csv
```

Output: `manual-test-scenarios.csv`

### Export as Markdown

```bash
r2r eac test export-manual --module eac-commands --release v1.2.0 --format markdown
```

Output: `manual-test-scenarios.md`

## Error Conditions

| Exit Code | Condition                             |
| --------- | ------------------------------------- |
| 1         | Module flag missing                   |
| 1         | Release flag missing                  |
| 1         | Unknown module                        |
| 1         | No @Manual scenarios found            |
| 1         | Invalid format specified              |
| 1         | Schema validation failed (JSON only)  |

## Workflow Integration

This command is the first step in the manual testing workflow:

1. **Export** → `test export-manual` generates scenarios
2. **Execute** → Human tester fills in results
3. **Import** → `test import-manual` validates and stores results
4. **Merge** → `test merge-results` adds to test manifest

## See Also

- [test import-manual](./import-manual.md) - Import manual test results
- [test merge-results](./merge-results.md) - Merge results into manifest
- [Execute Manual Tests](../../../../how-to-guides/eac/commands/build-test-validate/execute-manual-tests.md) - Full workflow guide
- [test Commands](../categories/test.md)
