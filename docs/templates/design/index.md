# Design Templates

Templates for design documentation and decision records.

---

## Available Templates

### [Decision Record (DR)](dr.md)

Template for documenting significant architectural and technical decisions.

**Use when:**
- Making architectural style changes (monolith to microservices, etc.)
- Selecting technology stack (databases, frameworks, languages)
- Defining integration patterns and protocols
- Making security architecture decisions
- Choosing between significant trade-offs

**Template includes:**
- Status tracking (Proposed/Accepted/Rejected/Deprecated/Superseded)
- Context and problem statement
- Decision description
- Consequences (positive and negative)
- Alternatives considered
- References

---

## Using Decision Records

**1. Copy the template:**

```bash
cp docs/templates/design/dr.md docs/reference/decision-records/dr-XXX-my-decision.md
```

**2. Fill in sections:**
- Title: Clear, concise decision description
- Status: Check appropriate status
- Context: Why this decision is needed
- Decision: What you decided to do
- Consequences: Trade-offs and impacts
- Alternatives: What else was considered

**3. Link to existing DRs:**
- Reference related decisions
- Note if this supersedes another DR
- Link to implementation details

---

## Example Decision Records

See existing decision records in [Reference/Decision Records](../../reference/decision-records/index.md):

- DR-001: Three-layer testing approach
- DR-002: Separate ATDD and BDD files
- DR-003: Test framework selection

---

**Back to:** [Templates](../index.md)
