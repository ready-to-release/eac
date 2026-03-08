# Manual Testing Reference

Manual tests are Gherkin scenarios tagged with `@Manual` that require human execution and observation.

---

## Workflow

```bash
# 1. Export manual test scenarios
eac test export-manual <module>
# Output: out/test/<module>/manual-test-scenarios.json

# 2. Execute tests manually (human)
# - Open scenarios.json
# - Execute each test scenario
# - Record results in results.json

# 3. Import results
eac test import-manual <module> --results results.json
# Output: out/test/<module>/manual-results.json

# 4. Merge results into test manifest
eac test merge-results <module>
# Updates: out/test/<module>/uow.manifest.json
```

---

## Scenario Export Format

**File**: `manual-test-scenarios.json`

```json
{
  "scenarios": [
    {
      "id": "scenario-001",
      "feature": "User Authentication",
      "name": "Login with valid credentials",
      "tags": ["@Manual", "@L3", "@ov"],
      "steps": [
        "Given user is on login page",
        "When user enters valid credentials",
        "Then user should be logged in"
      ],
      "location": "specs/auth/login.feature:12"
    }
  ]
}
```

---

## Results Import Format

**File**: `manual-test-results.json`

**Schema**: `contracts/core/0.1.0/schemas/manual-test-results.schema.json`

```json
{
  "metadata": {
    "tester": "John Doe",
    "executed_at": "2024-02-17T10:30:00Z",
    "environment": "PLTE",
    "version": "2024.02.1"
  },
  "results": [
    {
      "scenario_id": "scenario-001",
      "status": "passed",
      "executed_at": "2024-02-17T10:35:00Z",
      "duration_seconds": 45,
      "evidence": [
        {
          "type": "screenshot",
          "path": "evidence/login-success.png",
          "description": "Successful login screen"
        }
      ],
      "notes": "Login successful on first attempt"
    }
  ]
}
```

---

## Status Values

| Status     | Description                     |
| ---------- | ------------------------------- |
| `passed`   | Test passed all requirements    |
| `failed`   | Test failed verification        |
| `skipped`  | Test not executed               |
| `blocked`  | Test blocked by external factor |

---

## Evidence Collection

Evidence types supported:

| Type           | Purpose                    | Example                    |
| -------------- | -------------------------- | -------------------------- |
| `screenshot`   | Visual verification        | login-success.png          |
| `video`        | Screen recording           | checkout-flow.mp4          |
| `log`          | System logs                | application.log            |
| `document`     | Test report or sign-off    | compliance-approval.pdf    |
| `artifact`     | Generated output           | export-data.csv            |

---

## CI/CD Integration

Manual tests are typically executed during release milestones:

```yaml
# .github/workflows/release.yml
- name: Export Manual Tests
  run: eac test export-manual eac

- name: Upload Scenarios
  uses: actions/upload-artifact@v3
  with:
    name: manual-test-scenarios
    path: out/test/eac/manual-test-scenarios.json
```

After human execution:

```yaml
- name: Download Results
  uses: actions/download-artifact@v3
  with:
    name: manual-test-results

- name: Import Results
  run: eac test import-manual eac --results manual-test-results.json

- name: Merge to Manifest
  run: eac test merge-results eac
```

---

## Commands

| Command                                          | Purpose                  |
| ------------------------------------------------ | ------------------------ |
| `eac test export-manual <module>`                | Export scenarios to JSON |
| `eac test import-manual <module> --results file` | Import results from JSON |
| `eac test merge-results <module>`                | Merge into manifest      |

**See**: [Test Commands](../commands/test/index.md)

---

## Related Documentation

- [Test Suites](./test-suites.md) - Suite definitions
- [Test Commands](../commands/test/index.md) - CLI command reference
- [Test Levels](../../../explanation/specifications/taxonomy/test-levels.md) - Conceptual overview
