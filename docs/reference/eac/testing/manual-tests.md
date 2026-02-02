# Manual Testing Reference

> **Comprehensive technical reference for EAC manual testing system**

Manual tests are Gherkin scenarios tagged with `@Manual` that require human execution and observation. This reference provides complete technical documentation for the manual testing workflow, data schemas, file formats, CI/CD integration, and troubleshooting.

---

## Overview

### What Are Manual Tests?

Manual tests are test scenarios that cannot be easily automated and require human verification, judgment, or interaction. In the EAC testing system, these are Gherkin scenarios tagged with `@Manual` that are exported for human execution, then imported back into the test system with results.

Unlike automated tests that run on every commit, manual tests are executed at specific milestones (typically during release validation) and require:

- Human observation and decision-making
- Physical hardware interaction
- Subjective quality assessment (UX, accessibility)
- Third-party system integration without test APIs
- Compliance sign-off requiring human approval

### When to Use Manual Tests

| Use Manual Test ✅                                                                  | Automate Instead ❌                                            |
| ----------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| **Hardware verification** - Physical device interaction, hardware-specific behavior | **API response validation** - HTTP responses, JSON payloads    |
| **Usability/UX evaluation** - Subjective judgment, aesthetic assessment             | **Database query results** - Data validation, CRUD operations  |
| **Accessibility testing** - Screen readers, keyboard navigation, WCAG compliance    | **File content verification** - File parsing, content matching |
| **Third-party system integration** - No test API, production-only systems           | **Math calculations** - Deterministic computations             |
| **Compliance sign-off** - Human approval required, regulatory compliance            | **Unit logic** - Pure functions, business logic                |
| **Exploratory testing** - Ad-hoc investigation, edge case discovery                 | **Regression tests** - Repeatable tests, frequent execution    |
| **Visual verification** - Cross-browser rendering, responsive design                | **Command output** - CLI output validation, exit codes         |

### System Architecture

The manual testing system follows a four-phase workflow:

```text
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Phase 1    │    │   Phase 2    │    │   Phase 3    │    │   Phase 4    │
│              │    │              │    │              │    │              │
│    Export    │ →  │   Execute    │ →  │    Import    │ →  │    Merge     │
│   Scenarios  │    │    Tests     │    │   Results    │    │  to Manifest │
│              │    │              │    │              │    │              │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
       CLI                Human              CLI                  CLI

Input: .feature files with @Manual tags
Output: manual-test-scenarios.json

                      Input: scenarios.json
                      Output: results.json

                                          Input: results.json
                                          Output: manual-results.json

                                                              Input: manual-results.json
                                                              Output: test.manifest.json
```

**Data Flow**:

1. **Export**: CLI scans Gherkin specs, extracts `@Manual` scenarios, generates stable IDs
2. **Execute**: Human tester executes scenarios, records results, collects evidence
3. **Import**: CLI validates results JSON, checks schema compliance, stores in repository
4. **Merge**: CLI integrates manual results into test manifest with automated test results

---

## Workflow Architecture

### Phase 1: Export (test export-manual)

**Input**: Gherkin specs with `@Manual` tags
**Output**: `manual-test-scenarios.{json|csv|md}`
**Purpose**: Generate executable test scenarios for human testers

**Process**:

1. Scan `specs/<module>/` directory for `.feature` files
2. Parse Gherkin and extract scenarios tagged with `@Manual`
3. Generate stable scenario IDs (`module/feature/scenario-slug`)
4. Validate against export schema
5. Export in requested format (JSON, CSV, or Markdown)

**Command**:

```bash
eac test export-manual --module <module> --release <version> [--format json|csv|markdown]
```

---

### Phase 2: Execute (Human)

**Input**: Exported scenarios file
**Output**: Results file (tester fills in)
**Purpose**: Human executes tests and records outcomes

**Process**:

1. Copy export file → rename to results file
2. Update metadata (change `export_metadata` to `import_metadata`)
3. Add tester email and test execution timestamp
4. For each scenario:
   - Execute test steps manually
   - Record status (`passed`, `failed`, `skipped`)
   - Collect evidence (screenshots, logs, GitHub issues)
   - Add notes with observations
5. Save completed results file

**Evidence Requirements**:

- Store evidence at accessible URLs (GitHub Issues, internal wiki, screenshot hosting)
- Supported types: `screenshot`, `log`, `recording`, `document`, `issue`
- Optional: Include SHA-256 hash for integrity verification

---

### Phase 3: Import (test import-manual)

**Input**: Completed results JSON file
**Output**: `test-results/<module>/<version>/manual-results.json`
**Purpose**: Validate and store results in repository

**Validation Steps**:

1. **Schema validation** - Structure, types, required fields
2. **Email format validation** - Tester field must be valid email
3. **Release version matching** - File version must match `--release` flag
4. **Module validation** - Module must exist in repository.yml
5. **Scenario ID cross-validation** - IDs must exist in export file (if available)
6. **Conflict detection** - Check if results already exist for module/version
7. **Status validation** - Failed status requires error message

**Command**:

```bash
eac test import-manual --input <file> --release <version> [--force]
```

Use `--force` to overwrite existing results.

---

### Phase 4: Merge (test merge-results)

**Input**: Imported results at `test-results/<module>/<version>/manual-results.json`
**Output**: Updated `out/test/<module>/test.manifest.json`
**Purpose**: Integrate manual results with automated test results

**Merge Behavior**:

1. Read manual results from test-results directory
2. Transform to test manifest entries (TestAssertion format)
3. Create or update "manual" suite in manifest
4. Replace existing manual suite completely (idempotent operation)
5. Recalculate summary statistics (passed/failed/skipped totals)
6. Preserve automated test results (unit, integration, acceptance suites)

**Manifest Structure**:

```json
{
  "module": "eac-commands",
  "version": "v1.2.0",
  "timestamp": "2026-01-29T15:00:00Z",
  "summary": {
    "total": 800,
    "passed": 788,
    "failed": 10,
    "skipped": 2
  },
  "suites": {
    "unit": { "passed": 150, "failed": 2, "skipped": 0 },
    "integration": { "passed": 80, "failed": 1, "skipped": 0 },
    "acceptance": { "passed": 548, "failed": 5, "skipped": 1 },
    "manual": {
      "name": "manual",
      "passed": 10,
      "failed": 2,
      "skipped": 1,
      "total": 13,
      "tests": [
        {
          "name": "eac-commands/authentication/login-with-valid-credentials",
          "status": "passed",
          "duration_seconds": 45.0,
          "tags": ["@Manual", "@L2", "@ov"],
          "evidence": [...]
        }
      ]
    }
  }
}
```

**Command**:

```bash
eac test merge-results --module <module> --version <version>
```

---

## Data Schemas

### Export Schema (manual-test-export.schema.json)

**Contract**: `contracts/eac-core/0.1.0/manual-test-export.schema.json`

#### export_metadata Object

| Field           | Type   | Required | Format              | Description                    |
| --------------- | ------ | -------- | ------------------- | ------------------------------ |
| export_time     | string | ✅       | ISO 8601            | When scenarios were exported   |
| module          | string | ✅       | `^[a-z][a-z0-9-]*$` | Module moniker                 |
| release_version | string | ✅       | -                   | Release version (e.g., v1.2.0) |
| git_commit      | string | ✅       | 40 hex chars        | Git SHA at export time         |
| schema_version  | string | ✅       | "1.0"               | Schema version                 |

#### exported_scenario Object

| Field         | Type          | Required | Format                | Description                              |
| ------------- | ------------- | -------- | --------------------- | ---------------------------------------- |
| scenario_id   | string        | ✅       | `module/feature/slug` | Stable scenario identifier               |
| feature_name  | string        | ✅       | -                     | Gherkin feature name                     |
| scenario_name | string        | ✅       | -                     | Gherkin scenario name                    |
| tags          | array[string] | ✅       | `@tag` format         | All scenario tags (must include @Manual) |
| steps         | array[string] | ✅       | -                     | Gherkin steps (Given/When/Then/And)      |
| description   | string        | ❌       | -                     | Optional feature/scenario description    |
| file_path     | string        | ❌       | -                     | Source .feature file path                |

**Validation Rules**:

- Minimum 1 scenario required
- Git commit must be 40 hexadecimal characters
- Module must exist in repository.yml
- Scenario ID format must be `<module>/<feature>/<scenario-slug>`

---

### Results Schema (manual-test-results.schema.json)

**Contract**: `contracts/eac-core/0.1.0/manual-test-results.schema.json`

#### import_metadata Object

| Field            | Type   | Required | Format              | Description                         |
| ---------------- | ------ | -------- | ------------------- | ----------------------------------- |
| test_time        | string | ✅       | ISO 8601            | When tests were executed            |
| tester           | string | ✅       | email               | Tester email address                |
| module           | string | ✅       | `^[a-z][a-z0-9-]*$` | Module moniker                      |
| release_version  | string | ✅       | -                   | Release version (must match export) |
| duration_seconds | number | ❌       | ≥ 0                 | Total time spent executing tests    |
| schema_version   | string | ✅       | "1.0"               | Schema version                      |

#### manual_test_result Object

| Field            | Type          | Required | Format                | Description                                   |
| ---------------- | ------------- | -------- | --------------------- | --------------------------------------------- |
| scenario_id      | string        | ✅       | `module/feature/slug` | Must match exported scenario ID               |
| status           | string        | ✅       | enum                  | `passed`, `failed`, `skipped`                 |
| duration_seconds | number        | ❌       | ≥ 0                   | Time spent executing this scenario            |
| notes            | string        | ❌       | ≤ 5000 chars          | Observations, failure reasons, skip rationale |
| error            | string        | ⚠️       | ≤ 2000 chars          | **Required if status=failed**                 |
| evidence         | array[object] | ❌       | ≤ 10 items            | Evidence artifacts                            |

**Conditional Requirements**:

- If `status: "failed"`, then `error` field is required
- No `pending` or `undefined` statuses allowed (all manual tests must be executed)

#### evidence_reference Object

| Field       | Type   | Required | Format       | Description                                           |
| ----------- | ------ | -------- | ------------ | ----------------------------------------------------- |
| url         | string | ✅       | HTTP(S) URI  | URL to evidence artifact                              |
| type        | string | ✅       | enum         | `screenshot`, `log`, `recording`, `document`, `issue` |
| description | string | ❌       | ≤ 200 chars  | Brief description of evidence                         |
| sha256      | string | ❌       | 64 hex chars | SHA-256 hash for integrity verification               |

**Validation Rules**:

- Minimum 1 result required
- Email must be valid format (RFC 5322)
- Release version must match `--release` flag value
- All scenario IDs must reference same module
- Scenario IDs must exist in export file (if export file present)
- Failed status requires error message

---

## File Formats

### JSON Format (Canonical)

**Purpose**: Primary format with schema validation. Required for import.

**Export File Example**:

```json
{
  "export_metadata": {
    "export_time": "2026-01-29T12:00:00Z",
    "module": "eac-commands",
    "release_version": "v1.2.0",
    "git_commit": "a1b2c3d1234567890abcdef1234567890abcdef12",
    "schema_version": "1.0"
  },
  "scenarios": [
    {
      "scenario_id": "eac-commands/authentication/login-with-valid-credentials",
      "feature_name": "eac-commands_authentication",
      "scenario_name": "Login with valid credentials",
      "tags": ["@Manual", "@L2", "@ov"],
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

**Results File Example**:

```json
{
  "import_metadata": {
    "test_time": "2026-01-29T14:30:00Z",
    "tester": "jane.smith@company.com",
    "module": "eac-commands",
    "release_version": "v1.2.0",
    "duration_seconds": 1800.0,
    "schema_version": "1.0"
  },
  "results": [
    {
      "scenario_id": "eac-commands/authentication/login-with-valid-credentials",
      "status": "passed",
      "duration_seconds": 45.0,
      "notes": "Login successful, dashboard loaded in 2 seconds",
      "evidence": [
        {
          "url": "https://example.com/screenshots/login-success.png",
          "type": "screenshot",
          "description": "Dashboard after successful login"
        }
      ]
    },
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
        }
      ]
    }
  ]
}
```

---

### CSV Format (Spreadsheet-Compatible)

**Purpose**: Import into Excel, Google Sheets for tracking and management.

**Format Details**:

- **Headers**: snake_case, matches JSON schema field names
- **Step Separator**: Pipe character (`|`) for multiple steps
- **Tag Separator**: Space-separated string (parsed to array on import)
- **Encoding**: UTF-8

---

### Markdown Format (Human-Readable Checklist)

**Purpose**: Print or email manual test checklists for offline execution.

**Use Case**: Print checklist → tester manually checks off scenarios → transcribe results to JSON for import.

---

## Scenario ID Format

### Pattern

**Format**: `<module>/<feature>/<scenario-slug>`

**Examples**:

- `eac-commands/authentication/login-with-valid-credentials`
- `eac-core/validation/validate-config-schema`
- `ext-eac/github/create-pull-request-with-ai-description`

### Generation Algorithm

1. **Extract feature name** from file path:
   - Path: `specs/eac-commands/authentication/login.feature`
   - Feature directory: `authentication`

2. **Slugify scenario name**:
   - Original: `"Login with Valid Credentials: OAuth Flow!"`
   - Slugified: `"login-with-valid-credentials-oauth-flow"`
   - Rules:
     - Convert to lowercase
     - Replace spaces and special characters with hyphens
     - Remove trailing punctuation
     - Collapse multiple consecutive hyphens into one

3. **Combine**: `eac-commands/authentication/login-with-valid-credentials-oauth-flow`

### Stability Requirements

- Scenario IDs **MUST remain stable** across releases for traceability
- Renaming scenarios changes the scenario ID (this is expected and acceptable)
- Use scenario ID in results files (not array index)
- If feature directory structure changes, scenario ID changes accordingly

**Best Practice**: Design scenario names to be long-lived and descriptive.

---

## Command Reference

### Quick Reference

| Command              | Purpose                    | Output File                                           |
| -------------------- | -------------------------- | ----------------------------------------------------- |
| `test export-manual` | Generate scenario list     | `manual-test-scenarios.{json\|csv\|md}`               |
| `test import-manual` | Validate and store results | `test-results/<module>/<version>/manual-results.json` |
| `test merge-results` | Update test manifest       | `out/test/<module>/test.manifest.json`                |

### Export Command

**Synopsis**:

```bash
eac test export-manual --module <module> --release <version> [--format <format>]
```

**Formats**: `json` (default), `csv`, `markdown`

**Details**: See [test export-manual](../commands/test/export-manual.md)

---

### Import Command

**Synopsis**:

```bash
eac test import-manual --input <file> --release <version> [--force]
```

**Validation**: Schema compliance, email format, version matching, module validation, scenario ID cross-validation, conflict detection

**Details**: See [test import-manual](../commands/test/import-manual.md)

---

### Merge Command

**Synopsis**:

```bash
eac test merge-results --module <module> --version <version>
```

**Behavior**: Replaces manual suite in test manifest (idempotent operation)

**Details**: See [test merge-results](../commands/test/merge-results.md)

---

## Test Manifest Integration

### Manifest Structure

**File**: `out/test/<module>/test.manifest.json`

**Purpose**: Unified test results combining automated tests (unit, integration, acceptance) with manual test results.

**Complete Structure Example**:

```json
{
  "module": "eac-commands",
  "version": "v1.2.0",
  "timestamp": "2026-01-29T15:00:00Z",
  "git_commit": "a1b2c3d1234567890abcdef1234567890abcdef12",
  "summary": {
    "total": 813,
    "passed": 798,
    "failed": 12,
    "skipped": 3
  },
  "suites": {
    "unit": {
      "name": "unit",
      "passed": 150,
      "failed": 2,
      "skipped": 0,
      "total": 152
    },
    "integration": {
      "name": "integration",
      "passed": 80,
      "failed": 1,
      "skipped": 0,
      "total": 81
    },
    "acceptance": {
      "name": "acceptance",
      "passed": 558,
      "failed": 7,
      "skipped": 2,
      "total": 567
    },
    "manual": {
      "name": "manual",
      "passed": 10,
      "failed": 2,
      "skipped": 1,
      "total": 13,
      "tests": [
        {
          "name": "eac-commands/authentication/login-with-valid-credentials",
          "status": "passed",
          "duration_seconds": 45.0,
          "executed_by": "jane.smith@company.com",
          "executed_at": "2026-01-29T14:30:00Z",
          "tags": ["@Manual", "@L2", "@ov"],
          "notes": "Login successful, dashboard loaded in 2 seconds",
          "evidence": [
            {
              "url": "https://example.com/screenshots/login-success.png",
              "type": "screenshot",
              "description": "Dashboard after successful login"
            }
          ]
        }
      ]
    }
  }
}
```

### Merge Behavior

**Idempotent**: Running merge multiple times with same results produces same output. Safe to re-run.

**Replacement**: Manual suite is completely replaced (not appended). Previous manual results are overwritten.

**Summary Recalculation**: Total counts (total, passed, failed, skipped) are recalculated after merge.

**Preservation**: Automated suite results (unit, integration, acceptance) are unchanged by manual merge.

---

## CI/CD Integration

### Manual Tests in the 12-Stage CD Model

Manual tests integrate into the EAC [12-stage CD Model](../../../explanation/continuous-delivery/cd-model/stages.md) at **Stage 8-9**, where test results are committed to the release branch and approved together with automated test results before production deployment.

```text
┌──────────────────────────────────────────────────────────────┐
│ Development Stages (1-7)                                     │
├──────────────────────────────────────────────────────────────┤
│ 1. Authoring      → Local development                        │
│ 2. Pre-commit     → Local validation (5-10 min)              │
│ 3. Merge Request  → Peer review + automated CI               │
│ 4. Commit         → Integration + automated tests            │
│ 5. Acceptance     → Functional tests in PLTE                 │
│ 6. Extended Test  → Performance, security, compliance        │
│ 7. Exploration    → Stakeholder demos, exploratory testing   │
└──────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────┐
│ Release Stages (8-12)                                        │
├──────────────────────────────────────────────────────────────┤
│ 8. Start Release     → Create release branch                 │
│                      → Export manual test scenarios          │
│                      → QA executes tests                      │
│                      → ⭐ Commit results to release branch   │
│ 9. Release Approval  → ⭐ Approve automated + manual tests   │
│ 10. Prod Deploy      → Deploy to production                  │
│ 11. Live             → Monitor production                     │
│ 12. Toggling         → Feature flags (optional)              │
└──────────────────────────────────────────────────────────────┘
```

**Rationale**: Manual test results are committed to the release branch during Stage 8 preparation. In Stage 9, both automated test results (from CI) and committed manual test results are validated together as part of the release approval gate.

---

### Release Branch Workflow

Manual tests follow the release branch workflow used in this repository:

**Stage 8: Start Release**:

1. Create release branch from main: `release/eac-commands/v1.2.0`
2. Export manual test scenarios
3. QA team executes tests offline
4. **QA team commits results to release branch**

**Stage 9: Release Approval**:

5. Release workflow validates both:
   - Automated tests (from CI runs)
   - Manual test results (committed to release branch)
6. If any tests fail, release is blocked
7. If all tests pass, release proceeds to production

---

### GitHub Actions Integration

**Example: Release Workflow Checking Committed Manual Tests**:

This workflow shows how manual tests committed to the release branch are validated at Stage 9 (Release Approval):

```yaml
name: "release-eac-commands (stage 8+)"

on:
  workflow_dispatch:
    inputs:
      version:
        description: "Version to release (semver format: x.y.z)"
        required: true
        type: string

concurrency:
  group: $<< github.workflow >>-$<< github.head_ref || github.ref >>
  cancel-in-progress: false

jobs:
  # Stage 9: Release Approval
  validate-manual-tests:
    name: Validate Manual Test Results
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - name: Checkout release branch
        uses: actions/checkout@v6
        with:
          ref: release/eac-commands/$<< inputs.version >>

      - name: Setup Commands Binary
        uses: ./.github/actions/setup-commands

      - name: Check manual test results exist
        run: |
          RESULTS_FILE="test-results/eac-commands/$<< inputs.version >>/manual-results.json"

          if [ ! -f "$RESULTS_FILE" ]; then
            echo "❌ Manual test results not found at $RESULTS_FILE"
            echo "QA team must execute manual tests and commit results before release"
            exit 1
          fi

          echo "✅ Manual test results found"

      - name: Merge manual results into test manifest
        run: |
          eac test merge-results \
            --module eac-commands \
            --version $<< inputs.version >>

      - name: Check for test failures
        run: |
          # Block release if manual tests failed
          if eac show test-summary eac-commands | grep -q "manual.*failed"; then
            echo "❌ Manual tests failed - blocking release"
            echo ""
            echo "Failed tests must be fixed before proceeding to production"
            eac show suite manual --module eac-commands
            exit 1
          fi

          echo "✅ All manual tests passed"

  # Stage 9: Release Approval - Automated Approval
  approve-release:
    name: Approve Release
    needs: validate-manual-tests
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - name: Checkout repository
        uses: actions/checkout@v6
        with:
          fetch-depth: 0
          ref: release/eac-commands/$<< inputs.version >>

      - name: Approve release
        uses: ./.github/actions/approve-release
        with:
          module: eac-commands
          skip-ci-check: false
          timeout: "600"

  # Stage 10: Production Deployment
  deploy-production:
    name: Deploy to Production
    needs: approve-release
    runs-on: ubuntu-latest
    permissions:
      contents: write
      packages: write
    steps:
      - name: Checkout repository
        uses: actions/checkout@v6
        with:
          ref: release/eac-commands/$<< inputs.version >>

      - name: Deploy to production
        run: |
          echo "Deploying eac-commands $<< inputs.version >> to production"
          # Production deployment logic here

  # Stage 11: Live - Update Evidence
  update-evidence:
    name: Update Evidence
    needs: deploy-production
    runs-on: ubuntu-latest
    permissions:
      contents: write
      actions: read
    steps:
      - name: Checkout repository
        uses: actions/checkout@v6

      - name: Update evidence documentation
        uses: ./.github/actions/update-evidence
        with:
          module: eac-commands
          version: $<< inputs.version >>
```

---

## Related Documentation

### Conceptual Documentation

- [Execution Control Tags](../../../explanation/specifications/taxonomy/execution-control-tags.md) - @Manual tag concepts and philosophy
- [Test Levels](../../../explanation/specifications/taxonomy/test-levels.md) - L0-L4 execution environments

### How-to Guides

- [Execute Manual Tests](../../../how-to-guides/eac/commands/build-test-validate/execute-manual-tests.md) - Step-by-step workflow for executing manual tests

### Command Reference

- [test export-manual](../commands/test/export-manual.md) - Export command detailed reference
- [test import-manual](../commands/test/import-manual.md) - Import command detailed reference
- [test merge-results](../commands/test/merge-results.md) - Merge command detailed reference

### Test Suite Reference

- [Test Suites](./test-suites.md) - All test suites including manual suite
