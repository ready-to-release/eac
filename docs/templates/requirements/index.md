# Requirements Templates

Templates for creating feature specifications using the three-layer testing approach.

---

## Available Templates

### [acceptance.spec](acceptance.spec)

Template for ATDD (Acceptance Test-Driven Development) specifications.

**Tool**: [Gauge](https://gauge.org/)
**Format**: Markdown
**Purpose**: Define business requirements and acceptance criteria

**Includes:**
- User story (As a/I want/So that)
- Acceptance criteria (measurable outcomes)
- Acceptance tests (Gauge steps)
- Feature ID and metadata

### [behavior.feature](behavior.feature)

Template for BDD (Behavior-Driven Development) scenarios.

**Tool**: [Godog](https://github.com/cucumber/godog) / Cucumber
**Format**: Gherkin
**Purpose**: Executable scenarios in Given/When/Then format

**Includes:**
- Feature description and metadata
- Gherkin scenarios (Given/When/Then)
- Scenario tags (@success, @error, @ac1, etc.)
- Verification tags (@IV, @PV, or OV default)

---

## Quick Start

**1. Create feature directory:**

```bash
mkdir -p specs/<module>/<feature_name>
```

**2. Copy templates:**

```bash
cp docs/templates/specs/acceptance.spec specs/<module>/<feature_name>/
cp docs/templates/specs/behavior.feature specs/<module>/<feature_name>/
```

**3. Fill in placeholders:**
- Replace `<module>_<feature_name>` with your Feature ID
- Replace `[Feature Name]` with your feature title
- Fill in user story, acceptance criteria, and scenarios

---

## Three-Layer Approach

| Layer | File | Tool | Purpose |
|-------|------|------|---------|
| **ATDD** | acceptance.spec | Gauge | Business requirements and acceptance criteria |
| **BDD** | behavior.feature | Godog | Executable scenarios in Given/When/Then format |
| **TDD** | *_test.go | Go test | Unit tests for implementation |

**Complete guide**: See [Three-Layer Testing Approach](../../explanation/testing/three-layer-approach.md)

---

## Related Documentation

- **How-to Guide**: [Create Feature Spec](../../how-to-guides/testing/create-feature-spec.md)
- **Explanation**: [Three-Layer Testing Approach](../../explanation/testing/three-layer-approach.md)
- **Reference**:
  - [ATDD Format](../../reference/testing/atdd-format.md)
  - [BDD Format](../../reference/testing/bdd-format.md)

---

**Back to:** [Templates](../index.md)
