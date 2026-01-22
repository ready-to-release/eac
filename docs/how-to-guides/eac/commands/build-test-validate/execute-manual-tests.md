# Execute Manual Tests

## What You'll Accomplish

Execute manual tests for a module release, collect evidence, and integrate results into the test manifest for unified reporting.

## Prerequisites

### Required Setup

- Module has Gherkin specifications with `@Manual` tags
- Module and release version determined
- Test environment prepared
- Tester has email address for traceability

### What Are Manual Tests?

Manual tests are scenarios that require human verification, judgment, or interaction that cannot be easily automated:

- UI/UX verification (visual design, layout, responsiveness)
- Physical hardware testing (device interactions)
- Accessibility testing (screen readers, keyboard navigation)
- Exploratory testing (ad-hoc investigation)
- Regulatory compliance testing (legal requirements)
- User acceptance testing (stakeholder sign-off)

These tests are tagged with `@Manual` in Gherkin specifications and exported for human execution.

## Complete Workflow

### Overview

```text
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Export    │ →  │   Execute   │ →  │   Import    │ →  │    Merge    │
│  Scenarios  │    │    Tests    │    │   Results   │    │  to Manifest│
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
     CLI                Human             CLI                  CLI
```

### Step 1: Export Manual Test Scenarios

Export scenarios tagged with `@Manual` for the release:

```bash
r2r eac test export-manual --module eac-commands --release v1.2.0 --format json
```

**What happens**:

- Scans `specs/eac-commands/` for `.feature` files
- Extracts scenarios tagged with `@Manual`
- Generates stable scenario IDs
- Creates `manual-test-scenarios.json` with all test details

**Output file**: `manual-test-scenarios.json`

```json
{
  "export_metadata": {
    "export_time": "2026-01-19T12:00:00Z",
    "module": "eac-commands",
    "release_version": "v1.2.0",
    "git_commit": "a1b2c3d1234567890...",
    "schema_version": "1.0"
  },
  "scenarios": [
    {
      "scenario_id": "eac-commands/authentication/login-with-valid-credentials",
      "feature_name": "eac-commands_authentication",
      "scenario_name": "Login with valid credentials",
      "tags": ["@Manual", "@L2", "@Critical"],
      "steps": [
        "Given the login page is displayed",
        "When I enter valid username and password",
        "Then I should be redirected to dashboard"
      ],
      "description": "Verify successful login flow",
      "file_path": "specs/eac-commands/authentication/spec.feature"
    }
  ]
}
```

#### Alternative: Export as CSV or Markdown

**CSV** (for spreadsheet tools):

```bash
r2r eac test export-manual --module eac-commands --release v1.2.0 --format csv
```

**Markdown** (for human-readable checklist):

```bash
r2r eac test export-manual --module eac-commands --release v1.2.0 --format markdown
```

### Step 2: Execute Manual Tests

Human tester performs each scenario and records results.

#### 2a. Prepare Results File

Copy the export file and rename it:

```bash
cp manual-test-scenarios.json manual-test-results.json
```

#### 2b. Update Metadata

Modify the top-level structure:

```json
{
  "import_metadata": {
    "test_time": "2026-01-19T14:30:00Z",
    "tester": "jane.smith@company.com",
    "module": "eac-commands",
    "release_version": "v1.2.0",
    "duration_seconds": 1800.0,
    "schema_version": "1.0"
  },
  "results": [
    ...
  ]
}
```

**Important**: Change `export_metadata` to `import_metadata` and update fields:

- `test_time` - When tests were executed
- `tester` - Your email address
- `duration_seconds` - Total time spent (optional)

#### 2c. Fill In Test Results

For each scenario, update with execution results:

**Passed Test**:

```json
{
  "scenario_id": "eac-commands/authentication/login-with-valid-credentials",
  "status": "passed",
  "duration_seconds": 45.0,
  "notes": "Login successful, dashboard loaded in 2 seconds",
  "error": "",
  "evidence": [
    {
      "url": "https://example.com/screenshots/login-success.png",
      "type": "screenshot",
      "description": "Dashboard after successful login"
    }
  ]
}
```

**Failed Test**:

```json
{
  "scenario_id": "eac-commands/authentication/login-with-invalid-password",
  "status": "failed",
  "duration_seconds": 30.0,
  "notes": "Error message not displayed clearly",
  "error": "Expected error message 'Invalid password' but got generic 'Login failed'",
  "evidence": [
    {
      "url": "https://github.com/org/repo/issues/456",
      "type": "issue",
      "description": "Bug report for unclear error message"
    },
    {
      "url": "https://example.com/screenshots/error-message.png",
      "type": "screenshot",
      "description": "Generic error message displayed"
    }
  ]
}
```

**Skipped Test**:

```json
{
  "scenario_id": "eac-commands/payment/process-payment-with-stripe",
  "status": "skipped",
  "duration_seconds": 0,
  "notes": "Skipped - test environment doesn't have Stripe API keys configured",
  "error": "",
  "evidence": []
}
```

#### 2d. Collect Evidence

**Types of Evidence**:

- `screenshot` - UI screenshots, visual evidence
- `log` - Application logs, error traces
- `recording` - Screen recordings, video walkthroughs
- `document` - Test reports, compliance documents
- `issue` - Bug reports, GitHub issues

**Best Practices**:

- Upload screenshots/recordings to accessible location
- Link to GitHub issues for bugs
- Include SHA-256 hash for tamper detection
- Keep descriptions concise (200 char max)

### Step 3: Import Manual Test Results

Validate and store results:

```bash
r2r eac test import-manual --input manual-test-results.json --release v1.2.0
```

**What happens**:

- Validates JSON against schema
- Checks email format
- Verifies release version matches
- Cross-validates scenario IDs against export (if available)
- Checks for existing results
- Stores at `test-results/eac-commands/v1.2.0/manual-results.json`

**Output**:

```text
Imported manual test results for eac-commands v1.2.0
  Location: test-results/eac-commands/v1.2.0/manual-results.json
  Tests: 8 passed, 2 failed, 1 skipped
```

#### Handle Import Errors

**Validation Error**:

```text
schema validation failed: validation failed
- at '/results/0': missing property 'error'
```

**Fix**: Add error message for failed test:

```json
{
  "status": "failed",
  "error": "Expected behavior not observed"
}
```

**Conflict Error**:

```text
manual test results already exist for eac-commands v1.2.0
  To overwrite, use --force flag
```

**Fix**: Use `--force` to overwrite:

```bash
r2r eac test import-manual --input manual-test-results.json --release v1.2.0 --force
```

### Step 4: Merge Results into Test Manifest

Integrate manual results with automated test results:

```bash
r2r eac test merge-results --module eac-commands --version v1.2.0
```

**What happens**:

- Reads manual results from `test-results/eac-commands/v1.2.0/manual-results.json`
- Transforms into test entries
- Updates or creates `out/test/eac-commands/test.manifest.json`
- Replaces existing manual suite
- Recalculates summary statistics

**Output**:

```text
Merged manual test results for eac-commands v1.2.0
  Location: out/test/eac-commands/test.manifest.json
  Manual tests: 8 passed, 2 failed, 1 skipped
  Total tests in manifest: 788
```

## Example Scenario: Release Testing

You're testing the v1.2.0 release of `eac-commands` which includes new authentication features.

### Initial Setup

```bash
# Set release version
VERSION="v1.2.0"
MODULE="eac-commands"

# Export scenarios
r2r eac test export-manual --module $MODULE --release $VERSION --format json
```

### Execute Tests

Open `manual-test-scenarios.json` and:

1. Save as `manual-test-results.json`
2. Change `export_metadata` to `import_metadata`
3. Update metadata:

   ```json
   "import_metadata": {
     "test_time": "2026-01-19T14:30:00Z",
     "tester": "your.email@company.com",
     "module": "eac-commands",
     "release_version": "v1.2.0",
     "duration_seconds": 1800.0,
     "schema_version": "1.0"
   }
   ```

4. Execute each test and record results:
   - 8 tests passed ✅
   - 2 tests failed ❌ (filed bugs #456, #457)
   - 1 test skipped ⏭️ (missing test data)

5. Collect evidence:
   - Screenshots uploaded to internal wiki
   - Bug reports created in GitHub

### Import and Merge

```bash
# Import results
r2r eac test import-manual --input manual-test-results.json --release $VERSION

# Output:
# Imported manual test results for eac-commands v1.2.0
#   Location: test-results/eac-commands/v1.2.0/manual-results.json
#   Tests: 8 passed, 2 failed, 1 skipped

# Merge into manifest
r2r eac test merge-results --module $MODULE --version $VERSION

# Output:
# Merged manual test results for eac-commands v1.2.0
#   Location: out/test/eac-commands/test.manifest.json
#   Manual tests: 8 passed, 2 failed, 1 skipped
#   Total tests in manifest: 788
```

### View Results

```bash
# View overall test summary
r2r eac show test-summary $MODULE

# View manual suite details
r2r eac show suite manual --module $MODULE

# View all tests (including manual)
r2r eac show tests $MODULE
```

---

## Next Steps

- [test export-manual](../../../../reference/eac/commands/test/export-manual.md) → Detailed export options
- [test import-manual](../../../../reference/eac/commands/test/import-manual.md) → Validation rules
- [test merge-results](../../../../reference/eac/commands/test/merge-results.md) → Merge behavior
- [Debug Test Failures](./debug-test-failures.md) → Fix failed tests
- [Run Test Suites](./run-test-suites.md) → Execute automated tests

## Related Commands

- [`test export-manual`](../../../../reference/eac/commands/test/export-manual.md) - Export scenarios
- [`test import-manual`](../../../../reference/eac/commands/test/import-manual.md) - Import results
- [`test merge-results`](../../../../reference/eac/commands/test/merge-results.md) - Merge to manifest
- [`show test-summary`](../../../../reference/eac/commands/show/test-summary.md) - View results
- [`show suite`](../../../../reference/eac/commands/show/suite.md) - Suite details
- [`test`](../../../../reference/eac/commands/test/test.md) - Run automated tests
