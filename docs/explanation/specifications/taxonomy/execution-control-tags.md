# Execution Control Tags

> **Skipping and manual test execution**

Control how tests are executed in the test suite.

---

## `@ignore` - Exclude from Test Execution

**Purpose**: Exclude features or scenarios from all test suite runs (work-in-progress, temporarily disabled)

**Usage**: Feature level (excludes all scenarios) OR Scenario level (excludes only that scenario)

**Examples**:

```gherkin
# Feature-level: Entire feature excluded
@ignore @ov
Feature: new-feature_experimental-api
  Scenario: Create resource    # ❌ EXCLUDED
  Scenario: Delete resource    # ❌ EXCLUDED

# Scenario-level: Only OAuth scenario excluded
@ov
Feature: stable-feature_authentication
  Scenario: Valid login        # ✅ RUNS
  @ignore
  Scenario: OAuth (WIP)        # ❌ EXCLUDED
  Scenario: Session expiry     # ✅ RUNS
```

**Behavior**:

`@ignore` is evaluated before other selectors.

Ignored tests are excluded from all test suites regardless of other tags.

---

## `@Manual` - Manual Test Scenario

**Purpose**: Mark scenarios that must be executed manually (cannot be automated)

**Usage**: Scenario level (general use across all contexts)

### When to Use

- Requires human judgment or subjective evaluation
- Involves physical hardware interaction
- Depends on external systems not available in test environments
- Cost/complexity of automation exceeds benefit

### Documentation Requirements

- Include detailed test instructions as comments
- Document expected outcomes clearly
- Specify verification criteria
- Record results systematically

---

## Manual Test Examples

### Example - Usability Testing

```gherkin
@Manual @ov
Scenario: User finds the interface intuitive (usability test)
  Given I am a new user
  When I attempt to complete my first task
  Then I can complete it without consulting documentation
  # Manual observation by UX researcher
```

### Example - Performance Testing

```gherkin
@Manual @pv @deps:load-generator
Scenario: System handles 10,000 concurrent users (load test)
  Given the production environment is scaled to maximum capacity
  When 10,000 simulated users access the system simultaneously
  Then average response time is under 200ms
  And no errors occur
  # Manual execution using third-party load testing service
```

### Example - Hardware Integration

```gherkin
@Manual @ov @deps:hardware
Scenario: Printer outputs receipt correctly
  Given a transaction is completed
  When the receipt is printed
  Then the receipt contains all transaction details
  And the receipt is legible
  # Manual verification of physical printout
  # Test with: Epson TM-T88V printer
```

---

## Pipeline Execution and Evidence Collection

When a test suite encounters scenarios tagged with `@Manual`, the pipeline must **pause execution** to allow manual test execution:

### Manual Test Workflow

1. **Pipeline Pause**: Test runner stops at `@Manual` scenarios and waits for manual completion
2. **Manual Execution**: Tester executes the scenario following documented instructions
3. **Evidence Collection**: Tester collects evidence (screenshots, photos, logs, signatures, timestamps)
4. **Evidence Storage**: Evidence MUST be committed to git repository
5. **Pipeline Resume**: After evidence is committed, pipeline continues

### Critical Requirement

Evidence **cannot** be stored in separate systems (Excel spreadsheets, SharePoint, paper logbooks only).

All evidence must live in the git repository to maintain:

- Full traceability from requirement → test → evidence
- Version control of test results
- Auditability and compliance
- Single source of truth

---

## Best Practices

### Using `@ignore`

✅ **DO**:

- Use `@ignore` temporarily for work-in-progress features
- Document why tests are ignored (comments or issue links)
- Review ignored tests regularly (weekly in active development)
- Remove `@ignore` as soon as tests are stable

❌ **DON'T**:

- Use `@ignore` as permanent solution for broken tests (fix or remove them)
- Leave ignored tests without tracking (use issue numbers)
- Ignore tests for extended periods (>1 sprint)

**Example with documentation**:

```gherkin
# TODO: Remove @ignore after fixing test data issue (#1234)
@ignore @ov
Scenario: Process large dataset
  Given I have 1 million records
  When I process the dataset
  Then processing should complete within 1 hour
```

### Using `@Manual`

✅ **DO**:

- Use `@Manual` for tests that cannot be reasonably automated
- Include detailed test instructions as comments
- Document expected outcomes and verification criteria
- Store manual tests alongside automated tests in feature files
- **Commit evidence to git** (execution records, screenshots, signatures)
- Pause pipeline execution when manual tests are encountered
- Record manual test results systematically in markdown format
- Periodically review if manual tests can be automated

❌ **DON'T**:

- Use `@Manual` to avoid writing automation (automation-first approach)
- Leave manual tests without clear step-by-step instructions
- **Store evidence in separate systems** (Excel, SharePoint, paper only) - breaks traceability
- Continue pipeline without collecting and committing manual test evidence
- Forget to document why automation isn't feasible
- Let manual tests grow without review (aim to automate when possible)

---

## Evidence Storage Format

### Manual Test Evidence Structure

```text
specs/.evidence/
└── <module>/
    └── <feature>/
        └── <scenario>/
            ├── execution-record.md      # Test execution record
            ├── screenshot-001.png       # Visual evidence
            ├── screenshot-002.png
            └── data-output.log          # Log files
```

### Execution Record Template

```markdown
# Manual Test Execution Record

**Feature**: Authentication
**Scenario**: User finds the interface intuitive
**Executed By**: Jane Doe
**Date**: 2024-03-15
**Duration**: 15 minutes

## Test Steps

1. ✅ Given I am a new user
   - Opened application for first time
   - No prior experience

2. ✅ When I attempt to complete my first task
   - Attempted to create account
   - Screenshot: screenshot-001.png

3. ✅ Then I can complete it without consulting documentation
   - Successfully created account
   - Did not need help documentation
   - Screenshot: screenshot-002.png

## Result

**PASS** - User completed task intuitively

## Notes

- User hesitated slightly on password requirements
- Consider adding inline password hints
```

---

## Related Documentation

- [Verification Tags](./verification-tags.md) - Types of validation
- [Test Suites](./test-suites.md) - How ignored tests are excluded
- [GxP Tagging](../compliance/gxp-tagging.md) - Regulatory compliance evidence
