# ATDD Concepts

Understanding Acceptance Test-Driven Development with Gauge.

---

## What is ATDD?

**Acceptance Test-Driven Development (ATDD)** is a collaborative approach where business stakeholders, developers, and testers define acceptance criteria **before** development begins. It focuses on capturing business value and measurable success criteria from the customer's perspective.

### Core Purpose

ATDD answers the question: **"What does 'done' mean for this feature?"**

By defining acceptance criteria upfront, all stakeholders agree on:

- What business value the feature delivers
- How success will be measured
- What conditions must be satisfied for acceptance

---

## Why Use ATDD?

### Business Alignment

**Problem**: Developers build features that don't meet business needs
**Solution**: ATDD ensures everyone agrees on requirements before coding starts

### Reduced Rework

**Problem**: Discovering missing requirements after implementation
**Solution**: ATDD catches misunderstandings early through collaborative discussion

### Measurable Success

**Problem**: Subjective acceptance ("Does this look good?")
**Solution**: ATDD requires measurable criteria ("Creates 3 directories", "Completes in <2s")

### Stakeholder Collaboration

**Problem**: Product owners can't review technical test code
**Solution**: ATDD uses natural language (Markdown) that stakeholders can read and validate

---

## ATDD with Gauge

This project uses **[Gauge](https://gauge.org/)** to write executable ATDD specifications.

### Why Gauge?

- **Markdown format**: Natural language, easy for non-technical stakeholders
- **Executable**: Specifications become automated tests
- **Collaborative**: Business language maps directly to test steps
- **Clear reporting**: HTML/XML reports show which acceptance criteria pass/fail
- **Test data management**: Built-in support for tables and parameters

### Gauge Specifications

Gauge specifications are written in Markdown with executable test steps:

```markdown
# Initialize Project

## User Story

* As a developer
* I want to initialize a CLI project with a single command
* So that I can quickly start development

## Acceptance Criteria

* Creates project directory structure
* Generates valid configuration file

## Acceptance Tests

### AC1: Creates project directory structure

* Create empty test directory
* Run "cc init" command
* Verify "src/" directory exists
* Verify "tests/" directory exists
```

**Key Point**: These steps are **executable** - they map to actual Go code that validates the behavior.

---

## Example Mapping Workshop

**[Example Mapping](https://cucumber.io/blog/bdd/example-mapping-introduction/)** is the primary technique for ATDD requirement discovery. It's a time-boxed collaborative workshop (15-25 minutes) that uses colored index cards.

### The Four Card Colors

| Color | Represents | Purpose | Quantity |
|-------|------------|---------|----------|
| 🟡 **Yellow** | User Story | Captures WHO, WHAT, WHY | 1 per feature |
| 🔵 **Blue** | Rules/Acceptance Criteria | Defines success conditions | 2-6 per story |
| 🟢 **Green** | Concrete Examples | Specific test scenarios | 2-4 per criterion |
| 🔴 **Red** | Questions/Unknowns | Tracks blockers and uncertainties | 0-N (fewer is better) |

### Workshop Participants

**Product Owner**: Defines business value and priorities
**Developer**: Provides technical feasibility and constraints
**Tester/QA**: Identifies edge cases and scenarios

**Time commitment**: 15-25 minutes (strictly timeboxed)

### Workshop Process

#### Step 1: Place Yellow Card (2 minutes)

Write the user story in "As a / I want / So that" format:

```text
As a developer
I want to initialize a CLI project with one command
So that I can quickly start development
```

Place at the top of the table/board.

#### Step 2: Generate Blue Cards (8-12 minutes)

Brainstorm rules and acceptance criteria. Each Blue Card answers: "What does this feature need to do?"

**Guidelines**:

- Aim for 2-6 Blue Cards
- Each must be measurable (not subjective)
- Include both functional and non-functional requirements
- Keep them concise (one rule per card)

**Example Blue Cards**:

```text
[BLUE-1] Creates project directory structure
[BLUE-2] Generates valid configuration file
[BLUE-3] Command completes in under 2 seconds
```

**Warning Signs**:

- **>6 Blue Cards**: Feature is too large, split it up
- **<2 Blue Cards**: Feature might be trivial or poorly understood

#### Step 3: Create Green Cards (5-10 minutes)

For each Blue Card, create concrete examples showing the rule in action.

**Format**: [Context] → [Action] → [Result]

**Example Green Cards** (for Blue Card 1):

```text
[GREEN-1a] Empty folder → init → creates src/, tests/, docs/
[GREEN-1b] Existing project → init → error "already initialized"
```

**Guidelines**:

- Aim for 2-4 Green Cards per Blue Card
- Include both success and error scenarios
- Use actual data/commands, not abstract descriptions

#### Step 4: Capture Red Cards (Ongoing)

Whenever a question or uncertainty arises, write it on a Red Card and set it aside. Don't try to resolve during the workshop.

**Example Red Cards**:

```text
[RED-1] What if cc.yaml already exists?
[RED-2] Should we support a --force flag?
[RED-3] Do we need Windows compatibility?
```

**What to do with Red Cards**:

- **Immediate blockers**: Stop feature work, need answers first
- **Minor questions**: Track in issues.md, resolve before release
- **Future enhancements**: Create backlog items

#### Step 5: Assess Readiness (2 minutes)

Determine if the feature is ready for implementation:

**✅ Ready to implement**:

- 2-6 Blue Cards
- Each Blue Card has 2-4 Green Cards
- Few or no Red Cards (or Red Cards have owners/due dates)

**⚠️ Too large**:

- \>6 Blue Cards → Split into multiple stories

**❌ Too uncertain**:

- Many Red Cards → Needs research/spike first

### Visual Layout

```text
+---------------------------------------+
| [YELLOW CARD]                         |
| As a developer, I want to init        |
| project, so that I can start quickly  |
+---------------------------------------+
          |
          v
+-----------------+  +-----------------+  +-----------------+
| [BLUE CARD 1]   |  | [BLUE CARD 2]   |  | [BLUE CARD 3]   |
| Creates dirs    |  | Generates config|  | Handles errors  |
+-----------------+  +-----------------+  +-----------------+
  |                    |                    |
  v                    v                    v
[GREEN 1a]          [GREEN 2a]          [GREEN 3a]
[GREEN 1b]          [GREEN 2b]          [GREEN 3b]

[RED CARDS - TO THE SIDE]
[RED 1] What if config exists?
[RED 2] Support --force flag?
```

---

## From Cards to acceptance.spec

After the workshop, convert the cards into a Gauge specification file.

### Yellow Card → User Story Section

**Yellow Card**:

```text
As a developer
I want to initialize a CLI project with one command
So that I can quickly start development
```

**Becomes**:

```markdown
## User Story

* As a developer
* I want to initialize a CLI project with a single command
* So that I can quickly start development with proper structure
```

### Blue Cards → Acceptance Criteria + Tests

**Blue Cards**:

```text
[BLUE-1] Creates project directory structure
[BLUE-2] Generates valid configuration file
[BLUE-3] Command completes in under 2 seconds
```

**Becomes**:

```markdown
## Acceptance Criteria

* Creates project directory structure
* Generates valid configuration file
* Command completes in under 2 seconds

## Acceptance Tests

### AC1: Creates project directory structure
**Validated by**: behavior.feature -> @ac1 scenarios

* Create empty test directory
* Run "cc init" command
* Verify "src/" directory exists
* Verify "tests/" directory exists
* Verify "docs/" directory exists

### AC2: Generates valid configuration file
**Validated by**: behavior.feature -> @ac2 scenarios

* Create empty test directory
* Run "cc init" command
* Read "cc.yaml" file contents
* Verify YAML is valid
* Verify default values are present

### AC3: Command completes in under 2 seconds
**Validated by**: behavior.feature -> @ac3 scenarios

* Create empty test directory
* Start performance timer
* Run "cc init" command
* Stop performance timer
* Assert execution time is less than "2" seconds
```

### Green Cards → BDD Scenarios

Green Cards don't go in `acceptance.spec` - they are converted to Gherkin scenarios in `behavior.feature` (see [BDD Concepts](./bdd-concepts.md)).

### Red Cards → Questions Tracker

**Red Card**:

```text
What if cc.yaml already exists?
```

**Becomes** (in `issues.md`):

```markdown
## RED-1: What if cc.yaml already exists?

**Status**: Open
**Raised**: 2025-10-30
**Decision needed by**: Product Owner
**Resolution**: TBD
```

---

## Key ATDD Principles

### 1. Measurable Acceptance Criteria

**Bad** (subjective):

```markdown
* The interface is user-friendly
* Performance is good
* Error messages are helpful
```

**Good** (measurable):

```markdown
* Creates 3 directories (src/, tests/, docs/)
* Command completes in under 2 seconds
* Error message contains "already initialized" text
```

### 2. Collaboration Before Code

ATDD is **collaborative** - it requires:

- Product Owner (business perspective)
- Developer (technical perspective)
- Tester (quality perspective)

**Don't**: Have developers write acceptance criteria alone
**Do**: Run Example Mapping workshop with all roles present

### 3. Acceptance Criteria Drive Development

Acceptance criteria define "done":

- Development starts when criteria are clear
- Development ends when all criteria pass
- No "scope creep" mid-implementation

### 4. Living Documentation

Gauge specifications serve as:

- **Requirements documentation** (what the feature does)
- **Automated tests** (validation that it works)
- **Audit trail** (proof of testing for compliance)

All in one place, always up to date.

---

## Benefits of ATDD

### 1. Reduced Misunderstandings

**Before ATDD**:

- Developer: "I thought 'init' meant..."
- Product Owner: "No, I needed it to..."
- Result: Rework and delays

**With ATDD**:

- Example Mapping surfaces misunderstandings in 20 minutes
- Concrete examples clarify intent
- Everyone agrees before coding starts

### 2. Faster Feedback

**Traditional**: Discover issues after implementation (days/weeks later)
**ATDD**: Discover issues during 20-minute workshop (before any code)

### 3. Stakeholder Confidence

Business stakeholders can:

- Read acceptance criteria in plain language
- Review Gauge test results (green/red reports)
- Validate that business value was delivered

### 4. Prevents Gold-Plating

Clear acceptance criteria prevent:

- Adding features that weren't requested
- Over-engineering solutions
- Implementing "nice to have" instead of "must have"

**When all criteria pass, you're done. Ship it.**

---

## Common Questions

### Q: Who writes the acceptance.spec file?

**A**: Initially, the **developer** writes it after the Example Mapping workshop, based on the cards. Then the **Product Owner** reviews and approves it before development starts.

### Q: When do we write acceptance criteria?

**A**: **Before development begins**. ATDD is "test-driven" - write tests first, then implement.

### Q: How detailed should acceptance tests be?

**A**: Detailed enough to be executable. Each step should map to actual code that validates behavior. But focus on **what** to validate, not **how** to implement.

### Q: What if we discover new requirements during implementation?

**A**: Update the acceptance.spec file:

1. Add new acceptance criteria
2. Write new Gauge steps
3. Ensure they pass before considering feature complete

### Q: How many acceptance criteria per feature?

**A**: **2-6 criteria**. If you have more, the feature is too large - split it up.

---

## ATDD vs BDD vs TDD

### ATDD (acceptance.spec)

- **Focus**: Business requirements and value
- **Language**: Natural language (Markdown)
- **Stakeholder**: Product Owner
- **Question**: "Are we building the right thing?"

### BDD (behavior.feature)

- **Focus**: Observable user behavior
- **Language**: Gherkin (Given/When/Then)
- **Stakeholder**: QA and Developers
- **Question**: "Does it behave as expected?"

### TDD (unit tests)

- **Focus**: Implementation correctness
- **Language**: Go test code
- **Stakeholder**: Developers
- **Question**: "Does the code work correctly?"

**All three work together** - see [Three-Layer Approach](./three-layer-approach.md).

---

## Related Documentation

- [Three-Layer Testing Approach](./three-layer-approach.md) - How ATDD/BDD/TDD work together
- [BDD Concepts](./bdd-concepts.md) - Understanding BDD with Godog
- [ATDD Format Reference](../../reference/testing/atdd-format.md) - Specification format
- [Create Feature Spec](../../how-to-guides/testing/create-feature-spec.md) - Step-by-step guide
- [Example Mapping (External)](https://cucumber.io/blog/bdd/example-mapping-introduction/) - Original technique
