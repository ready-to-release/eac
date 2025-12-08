<!-- EDITOR
# Editor: reference/specifications/example-mapping-cards.md

## Soul

Example Mapping card color reference: Yellow=Story, Blue=Acceptance Criteria/Rules, Green=Examples/Scenarios, Pink=Questions.

## Sections

1. Card Color Reference
2. Yellow Card (User Story)
3. Blue Card (Acceptance Criteria)
4. Green Card (Examples)
5. Pink Card (Questions)
6. Visual Layout
7. Workshop Readiness Signals
8. Related
-->

# Example Mapping Card Colors

Reference for card colors and meanings in Example Mapping workshops.

## Card Color Reference

| Color  | Represents          | Maps To             | Quantity Limit                |
| ------ | ------------------- | ------------------- | ----------------------------- |
| Yellow | User Story          | Feature description | 1 per feature                 |
| Blue   | Acceptance Criteria | `Rule:` blocks      | ≤ 4 per story                 |
| Green  | Concrete Examples   | `Scenario:` blocks  | ≤ 4 per criterion             |
| Pink   | Questions/Unknowns  | issues.md           | Resolve before implementation |

## Yellow Card (User Story)

**Purpose:** Define WHO, WHAT, and WHY.

**Format:**

```text
As a [role]
I want [capability]
So that [business value]
```

**Quantity:** 1 per feature/workshop session.

**Maps to:** Feature description in Gherkin.

## Blue Card (Acceptance Criteria)

**Purpose:** Define measurable success conditions.

**Good Examples:**

- "Creates 3 directories (src/, tests/, docs/)"
- "Completes in < 2 seconds"
- "Returns error code 400 for invalid input"

**Bad Examples:**

- "Interface is user-friendly" (not measurable)
- "Performance is good" (subjective)

**Quantity:** ≤ 4 per Yellow card. If more, split the feature.

**Maps to:** `Rule:` blocks in Gherkin.

## Green Card (Examples)

**Purpose:** Provide concrete scenarios illustrating the rule.

**Format:** [Context] → [Action] → [Result]

**Good Examples:**

- "Empty folder → init → creates src/, tests/, docs/"
- "Invalid email → submit → shows 'Invalid email format'"

**Bad Examples:**

- "It creates directories" (too abstract)
- "Works correctly" (not specific)

**Quantity:** ≤ 4 per Blue card. If more, consider splitting the rule.

**Maps to:** `Scenario:` blocks under corresponding `Rule:`.

## Pink Card (Questions)

**Purpose:** Capture unknowns without solving during workshop.

**Examples:**

- "What if r2r.yaml already exists?"
- "Should we support a --force flag?"
- "What happens with no network?"

**Handling:**

- Write down immediately
- Do not debate during workshop
- Resolve before implementation starts
- Document in issues.md if deferred

**Maps to:** Issues to resolve or backlog items.

## Visual Layout

```
        [Yellow: User Story]
               |
    +----------+----------+
    |          |          |
[Blue 1]   [Blue 2]   [Blue 3]
    |          |          |
+---+---+  +---+---+  +---+---+
|   |   |  |   |   |  |   |   |
G1a G1b    G2a G2b    G3a G3b

[Pink questions placed to the side]
```

## Workshop Readiness Signals

| Signal        | Cards                                    | Action             |
| ------------- | ---------------------------------------- | ------------------ |
| Ready         | ≤4 Blue, ≤4 Green each, no blocking Pink | Proceed to Gherkin |
| Too Large     | >4 Blue cards or >25 min                 | Split feature      |
| Too Uncertain | Blocking Pink cards                      | Research spike     |

## Related

- [Running Example Mapping](../../how-to-guides/eac/specifications/running-example-mapping.md)
- [Translating to Gherkin](../../how-to-guides/eac/specifications/translating-to-gherkin.md)
- [Example Mapping](../../explanation/specifications/example-mapping.md)
