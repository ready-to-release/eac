# Templates

Document templates for various project artifacts.

---

## Available Templates

### [Implementation Report](implementation-report.md)

Template for creating implementation reports with test results grouped by verification type (IV/OV/PV).

**Use when**: Documenting feature implementation completion with test traceability.

---

### [Design Templates](design/index.md)

Templates for design documentation and decision records.

**Available:**
- Decision Record (DR) template

---

### [Requirements Templates](specs/index.md)

Templates for test specifications and feature requirements.

**Available:**
- Acceptance specification template (acceptance.spec)
- Behavior scenarios template (behavior.feature)

---

## Using Templates

**1. Copy the template file:**

```bash
cp docs/templates/implementation-report.md docs/my-implementation-report.md
```

**2. Fill in the sections:**
- Replace placeholder text with your content
- Follow the structure and format provided
- Keep the same heading hierarchy

**3. Link to related documentation:**
- Reference acceptance.spec and behavior.feature files
- Link to test results and coverage reports
- Include Feature ID for traceability

---

## Template Categories

| Category | Purpose | Files |
|----------|---------|-------|
| **Design** | Design decisions and architecture | Decision records |
| **Requirements** | Test specifications and acceptance criteria | ATDD/BDD templates |
| **Reports** | Implementation and verification documentation | Implementation reports |

---

**Need help with testing?** See [Testing How-to Guides](../how-to-guides/testing/index.md)

**Understanding test formats?** See [Testing Reference](../reference/testing/index.md)
