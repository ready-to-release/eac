# Example Mapping

Workshop technique for discovering requirements through collaborative conversation.

---

## What is Example Mapping?

**[Example Mapping](https://cucumber.io/blog/bdd/example-mapping-introduction/)** is a collaborative workshop technique that uses colored index cards to discover and refine requirements before development begins. It's a time-boxed (15-25 minutes) structured conversation that produces concrete examples and acceptance criteria.

### Core Purpose

Example Mapping answers three critical questions:

1. **What does this feature need to do?** (Acceptance criteria)
2. **What are concrete examples?** (Test scenarios)
3. **What don't we know yet?** (Questions and risks)

### Prerequisites: Establishing Domain Language

Before running Example Mapping workshops, teams should establish a basic **Ubiquitous Language** through **Event Storming** to ensure everyone speaks the same domain vocabulary.

**Event Storming establishes the vocabulary**:

- Domain events: "Order Placed", "Payment Received"
- Actors: "Customer", "Manager"
- Policies: "Large orders require approval"
- **See**: [Event Storming](./event-storming.md) for domain discovery

**Ubiquitous Language ensures shared understanding**:

- Business and technical teams use the same terms
- Specifications reflect actual domain concepts
- No translation errors between business and code
- **See**: [Ubiquitous Language](./ubiquitous-language.md) for DDD foundation

**Example Mapping then applies this vocabulary** to specific features through collaborative workshops.

---

### Why Use Example Mapping?

#### Discover Requirements Early

**Problem**: Teams start coding without shared understanding
**Solution**: 15-25 minute conversation surfaces misunderstandings before any code is written

#### Concrete Over Abstract

**Problem**: Abstract requirements like "user-friendly" are subjective
**Solution**: Concrete examples like "creates 3 directories" are measurable

#### Fast Feedback Loop

**Problem**: Requirements issues discovered days/weeks into development
**Solution**: Issues discovered in 20 minutes, before development starts

#### Collaborative Understanding

**Problem**: Different stakeholders have different mental models
**Solution**: Cards create shared artifacts everyone can see and discuss

---

## Building the Shared Language First

Before running Example Mapping workshops, teams benefit from understanding the **domain language**—the specific vocabulary of the business domain.

**Domain-Driven Design (DDD)** provides techniques for discovering and formalizing this language:

- **Ubiquitous Language** - A rigorous, shared vocabulary between business and technical teams
- **Event Storming** - Collaborative workshops that surface domain concepts and vocabulary
- **Bounded Contexts** - Define where specific terms apply and mean specific things

**The flow**:

1. **Event Storming** - Discover domain language (days/weeks before)
2. **Example Mapping** - Apply that language to specific features (15-25 minutes)
3. **Gherkin Specifications** - Write using the shared language

When Example Mapping workshops use this established vocabulary:

- Blue cards (rules) use precise domain terms
- Green cards (examples) reflect actual domain scenarios
- Red cards (questions) often reveal where domain understanding is still unclear

See: [Ubiquitous Language](./ubiquitous-language.md) and [Event Storming](./event-storming.md)

---

## The Four Card Colors

Example Mapping uses four colored index cards, each with a specific purpose:

| Color | Represents | Purpose | Maps to Gherkin | Quantity |
|-------|------------|---------|-----------------|----------|
| 🟡 **Yellow** | User Story | Captures WHO, WHAT, WHY | Feature description | 1 per feature |
| 🔵 **Blue** | Rules/Acceptance Criteria | Defines success conditions | `Rule:` blocks (ATDD) | 2-6 per story |
| 🟢 **Green** | Concrete Examples | Specific test scenarios | `Scenario:` blocks (BDD) | 2-4 per criterion |
| 🔴 **Red** | Questions/Unknowns | Tracks blockers | issues.md | 0-N (fewer is better) |

### Yellow Cards - User Stories

**Purpose**: Define the feature from the user's perspective

**Format**: "As a [role] / I want [capability] / So that [business value]"

**Example**:

```text
As a developer
I want to initialize a CLI project with one command
So that I can quickly start development
```

**Quantity**: Always exactly 1 per feature

### Blue Cards - Rules/Acceptance Criteria

**Purpose**: Define what the feature must do to be considered "done"

**Format**: Measurable business rules, one per card

**Examples**:

```text
[BLUE-1] Creates project directory structure
[BLUE-2] Generates valid configuration file
[BLUE-3] Command completes in under 2 seconds
```

**Quantity**: 2-6 per feature

- **<2**: Feature might be too simple or poorly understood
- **>6**: Feature is too large, split it up

### Green Cards - Concrete Examples

**Purpose**: Show the rule in action with specific data

**Format**: [Context] → [Action] → [Result]

**Examples** (for Blue Card "Creates directory structure"):

```text
[GREEN-1a] Empty folder → init → creates src/, tests/, docs/
[GREEN-1b] Existing project → init → error "already initialized"
```

**Quantity**: 2-4 per Blue Card

- Include both success and error cases
- Use real data, not abstract descriptions

### Red Cards - Questions and Unknowns

**Purpose**: Capture uncertainties without derailing the workshop

**Format**: Question or concern that needs resolution

**Examples**:

```text
[RED-1] What if cc.yaml already exists?
[RED-2] Should we support a --force flag?
[RED-3] Do we need Windows compatibility?
```

**Action**:

- **Immediate blockers**: Stop feature work, need answers first
- **Minor questions**: Track in issues.md, resolve before release
- **Future enhancements**: Create backlog items

---

## Workshop Participants

### Required Roles

**Product Owner**:

- Defines business value and priorities
- Provides domain expertise
- Makes decisions on scope

**Developer**:

- Provides technical feasibility input
- Identifies technical constraints
- Estimates complexity

**Tester/QA**:

- Identifies edge cases and scenarios
- Questions assumptions
- Ensures testability

### Time Commitment

**Duration**: 15-25 minutes (strictly time-boxed)

**Why time-boxed?**

- Forces focus on essentials
- Prevents over-analysis
- Creates urgency to resolve unknowns later
- If you can't finish in 25 minutes, the feature is too large

---

## Workshop Process

### Step 1: Place Yellow Card (2 minutes)

**Action**: Write the user story

**Format**:

```text
As a [role]
I want [capability]
So that [business value]
```

**Example**:

```text
As a developer
I want to initialize a CLI project with one command
So that I can quickly start development
```

**Tips**:

- Keep it focused on user value
- One feature per workshop
- Place at the top of the table/board

---

### Step 2: Generate Blue Cards (8-12 minutes)

**Action**: Brainstorm rules and acceptance criteria

**Each Blue Card answers**: "What does this feature need to do?"

**Guidelines**:

- Aim for 2-6 Blue Cards
- Each must be measurable (not subjective)
- Include both functional and non-functional requirements
- Keep them concise (one rule per card)
- Use domain language established in Event Storming

**Example Blue Cards**:

```text
[BLUE-1] Creates project directory structure
[BLUE-2] Generates valid configuration file
[BLUE-3] Command completes in under 2 seconds
```

**Warning Signs**:

❌ **>6 Blue Cards**: Feature is too large, split it up

Example: If you have Blue Cards for "creates dirs", "creates config", "creates docs", "handles errors", "validates input", "logs activity", "sends notifications", "updates registry" → Split into 2-3 features

❌ **<2 Blue Cards**: Feature might be trivial or poorly understood

Example: If you only have "Command works" → Dig deeper, what does "works" mean?

**Good Examples**:

```text
✅ Creates 3 directories (src/, tests/, docs/)
✅ Config file contains valid YAML with 5 required keys
✅ Command completes in under 2 seconds
```

**Bad Examples**:

```text
❌ Interface is user-friendly (not measurable)
❌ Performance is good (not measurable)
❌ Error handling works correctly (too vague)
```

---

### Step 3: Create Green Cards (5-10 minutes)

**Action**: For each Blue Card, create concrete examples

**Format**: [Context] → [Action] → [Result]

**Example Green Cards** (for Blue Card 1: "Creates directory structure"):

```text
[GREEN-1a] Empty folder → init → creates src/, tests/, docs/
[GREEN-1b] Existing project → init → error "already initialized"
[GREEN-1c] No write permissions → init → error "permission denied"
```

**Guidelines**:

- Aim for 2-4 Green Cards per Blue Card
- Include both success and error scenarios
- Use actual data/commands, not abstract descriptions
- Think of edge cases

**Good Examples**:

```text
✅ Given: Empty directory
   When: Run "cc init my-project"
   Then: Creates my-project/src/, my-project/tests/, my-project/docs/

✅ Given: Directory with cc.yaml
   When: Run "cc init"
   Then: Exits with code 1, stderr shows "already initialized"
```

**Bad Examples**:

```text
❌ It creates directories (too abstract)
❌ Command works (not specific)
❌ Error is handled (doesn't say how)
```

**Tips**:

- Green Cards become Gherkin scenarios later
- Think like a tester: "How could this go wrong?"
- Real examples > hypothetical examples

---

### Step 4: Capture Red Cards (Ongoing)

**Action**: Whenever a question arises, write it on a Red Card and set it aside

**Important**: Don't try to resolve Red Cards during the workshop

**Example Red Cards**:

```text
[RED-1] What if cc.yaml already exists?
[RED-2] Should we support a --force flag?
[RED-3] Do we need Windows compatibility?
[RED-4] What's the maximum project name length?
```

**After the Workshop**:

Create issues for Red Cards in `specs/<module>/<feature>/issues.md`:

```markdown
## RED-1: What if cc.yaml already exists?

**Status**: Open
**Raised**: 2025-11-04
**Decision needed by**: Product Owner
**Priority**: Blocker
**Resolution**: TBD

## RED-2: Should we support a --force flag?

**Status**: Open
**Raised**: 2025-11-04
**Decision needed by**: Product Owner
**Priority**: Enhancement (not blocking)
**Resolution**: TBD
```

---

### Step 5: Assess Readiness (2 minutes)

**Action**: Determine if the feature is ready for implementation

#### ✅ Ready to Implement

Requirements:

- 2-6 Blue Cards
- Each Blue Card has 2-4 Green Cards
- Few or no Red Cards (or Red Cards have owners and due dates)

**Next step**: Write specifications

#### ⚠️ Too Large

Indicators:
>
- >6 Blue Cards
- Workshop exceeded 25 minutes

**Next step**: Split into multiple stories

**How to split**:

- Group related Blue Cards
- Each group becomes a separate feature
- Prioritize which to implement first

#### ❌ Too Uncertain

Indicators:

- Many Red Cards (>3 blocking questions)
- Fundamental uncertainties about requirements
- Cannot define measurable criteria

**Next step**: Research spike, then re-run workshop

---

## Visual Layout

Physical or virtual board layout:

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
[GREEN 1c]

[RED CARDS - TO THE SIDE]
[RED 1] What if config exists?
[RED 2] Support --force flag?
```

**Tips for Physical Workshops**:

- Use actual colored index cards
- Spread out on a table
- Everyone can see and touch the cards
- Easy to move and reorganize

**Tips for Virtual Workshops**:

- Use Miro, Mural, or similar tools
- Use colored virtual sticky notes
- Enable everyone to add cards
- Take a screenshot at the end

---

## From Example Mapping to Gherkin

After the workshop, convert the cards into a specification file.

### Card Mapping

| Example Mapping Card | Gherkin Element | Location |
|---------------------|-----------------|----------|
| 🟡 Yellow Card (User Story) | Feature description | `specs/<module>/<feature>/specification.feature` |
| 🔵 Blue Card (Acceptance Criterion) | `Rule:` block | `specs/<module>/<feature>/specification.feature` |
| 🟢 Green Card (Example) | `Scenario:` block under Rule | `specs/<module>/<feature>/specification.feature` |
| 🔴 Red Card (Question) | Issue in issues.md | `specs/<module>/<feature>/issues.md` |

### Complete Example

#### Workshop Cards

**Yellow Card**:

```text
As a developer
I want to initialize a CLI project with one command
So that I can quickly start development
```

**Blue Cards**:

```text
[BLUE-1] Creates project directory structure
[BLUE-2] Generates valid configuration file
```

**Green Cards**:

```text
[GREEN-1a] Empty folder → init → creates src/, tests/, docs/
[GREEN-1b] Existing project → init → error "already initialized"
[GREEN-2a] Valid YAML generated with default values
[GREEN-2b] YAML contains required keys: project.name, project.version
```

**Red Cards**:

```text
[RED-1] What if cc.yaml already exists?
[RED-2] Should we support a --force flag?
```

#### Resulting Gherkin File

**File**: `specs/cli/init-project/specification.feature`

```gherkin
# Feature ID: cli_init-project
# Module: CLI

@cli @critical @init
Feature: cli_init-project

  As a developer
  I want to initialize a CLI project with one command
  So that I can quickly start development

  Rule: Creates project directory structure

    @success @ac1
    Scenario: Initialize in empty directory creates structure
      Given I am in an empty folder
      When I run "cc init my-project"
      Then a directory named "my-project/src/" should exist
      And a directory named "my-project/tests/" should exist
      And a directory named "my-project/docs/" should exist

    @error @ac1
    Scenario: Initialize in existing project shows error
      Given I am in a directory with "cc.yaml"
      When I run "cc init"
      Then the command should fail
      And stderr should contain "already initialized"

  Rule: Generates valid configuration file

    @success @ac2
    Scenario: Generated YAML has default values
      Given I am in an empty folder
      When I run "cc init my-project"
      Then a file named "my-project/cc.yaml" should be created
      And the file should contain valid YAML

    @success @ac2
    Scenario: YAML contains required keys
      Given I am in an empty folder
      When I run "cc init my-project"
      Then the YAML should have key "project.name"
      And the YAML should have key "project.version"
```

**File**: `specs/cli/init-project/issues.md`

```markdown
# Open Questions

## RED-1: What if cc.yaml already exists?

**Status**: Open
**Raised**: 2025-11-04
**Owner**: Product Owner
**Decision**: TBD

## RED-2: Should we support a --force flag?

**Status**: Open
**Raised**: 2025-11-04
**Owner**: Product Owner
**Priority**: Enhancement (not blocking MVP)
**Decision**: Defer to v1.1
```

---

## Best Practices

### Do's ✅

- **Time-box strictly** - If you can't finish in 25 minutes, split the feature
- **Use domain language** - Terms from Event Storming or business vocabulary
- **Be specific** - "Creates 3 directories" not "Creates directories"
- **Include error cases** - At least one error Green Card per Blue Card
- **Set Red Cards aside** - Don't solve them in the workshop
- **Invite right people** - Product Owner + Developer + Tester minimum
- **Start small** - Better to split features than have huge workshops

### Don'ts ❌

- **Don't skip the Yellow Card** - Context matters
- **Don't make Blue Cards subjective** - "Fast" → "Completes in <2 seconds"
- **Don't use abstract Green Cards** - "It works" → specific input/output
- **Don't debate Red Cards** - Write them down, resolve later
- **Don't exceed 6 Blue Cards** - Split the feature instead
- **Don't run workshops alone** - Collaboration is the point
- **Don't skip edge cases** - Error scenarios are important

---

## Anti-Patterns to Avoid

### ❌ Anti-Pattern: Solution in User Story

**Bad**:

```text
As a developer
I want a YAML configuration file
So that settings are stored
```

**Good**:

```text
As a developer
I want to configure my CLI tool
So that it behaves according to my preferences
```

**Why**: User story should focus on the problem, not the solution. YAML is an implementation detail.

### ❌ Anti-Pattern: Unmeasurable Blue Cards

**Bad**:

```text
[BLUE-1] Interface is intuitive
[BLUE-2] Performance is acceptable
[BLUE-3] Errors are helpful
```

**Good**:

```text
[BLUE-1] Command uses standard CLI conventions (--help, --version)
[BLUE-2] Command completes in under 2 seconds for projects with <100 files
[BLUE-3] Error messages include the problem and next action
```

### ❌ Anti-Pattern: Too Many Green Cards per Blue

**Problem**: 10+ Green Cards for one Blue Card

**Symptom**: The Blue Card is actually multiple acceptance criteria

**Fix**: Split the Blue Card into multiple Blue Cards

**Example**:

```text
❌ [BLUE-1] Handles all errors
   [GREEN-1a] File not found
   [GREEN-1b] Permission denied
   [GREEN-1c] Invalid format
   ... (10 more error cases)

✅ [BLUE-1] Handles file system errors
   [GREEN-1a] File not found
   [GREEN-1b] Permission denied

✅ [BLUE-2] Handles validation errors
   [GREEN-2a] Invalid YAML format
   [GREEN-2b] Missing required fields
```

### ❌ Anti-Pattern: Implementation Details in Green Cards

**Bad**:

```text
[GREEN-1a] Parser loads YAML → deserializes → validates schema
```

**Good**:

```text
[GREEN-1a] Valid YAML file → init succeeds → config loaded
```

**Why**: Green Cards should focus on observable behavior, not internal implementation.

---

## When to Split Features

### Scenario 1: Too Many Blue Cards (>6)

**Example**: Init Project feature with 10 Blue Cards

**Solution**: Split into 2-3 features

- Feature 1: "Initialize Basic Project" (3 Blue Cards)
- Feature 2: "Initialize with Templates" (4 Blue Cards)
- Feature 3: "Initialize with Git Setup" (3 Blue Cards)

### Scenario 2: Workshop Exceeds Time Limit

**Example**: 40 minutes into workshop, still adding Blue Cards

**Solution**: Stop and split

- Document what you have
- Prioritize Blue Cards
- Create separate feature for lower-priority cards

### Scenario 3: Green Cards Reveal Complexity

**Example**: Every Blue Card has 6+ Green Cards

**Symptom**: Feature is more complex than initially thought

**Solution**:

- Keep current feature as "MVP" (minimum viable)
- Create separate features for edge cases and advanced scenarios

---

## Related Documentation

### For Understanding

- [ATDD and BDD with Gherkin](./atdd-bdd-with-gherkin.md) - Concepts behind the specifications
- [Three-Layer Testing Approach](./three-layer-approach.md) - How Example Mapping fits the architecture
- [Ubiquitous Language](./ubiquitous-language.md) - Domain-Driven Design and shared vocabulary
- [Event Storming](./event-storming.md) - Domain discovery workshops

### For Doing

- [Create Feature Spec](../../how-to-guides/specifications/create-specifications.md) - Step-by-step guide to convert cards to Gherkin
- [Run Example Mapping Workshop](../../how-to-guides/specifications/run-example-mapping.md) - How to facilitate a workshop

### For Reference

- [Gherkin Format](../../reference/specifications/gherkin-format.md) - Complete Gherkin syntax reference
- [Example Mapping Introduction (External)](https://cucumber.io/blog/bdd/example-mapping-introduction/) - Original technique by Matt Wynne
