# Run Example Mapping Workshop

Facilitate an Example Mapping workshop to discover requirements collaboratively.

---

## What is Example Mapping?

[Example Mapping](https://cucumber.io/blog/bdd/example-mapping-introduction/) is a time-boxed workshop (15-25 minutes) that uses colored index cards to discover requirements through structured conversation.

**Goal**: Transform abstract requirements into concrete examples that can become automated tests.

---

## Prerequisites

### Materials

- **Colored index cards** (or sticky notes):
  - 🟡 Yellow (user stories)
  - 🔵 Blue (acceptance criteria)
  - 🟢 Green (concrete examples)
  - 🔴 Red (questions/blockers)
- **Table or whiteboard** for layout
- **Markers/pens**

### Participants (Required)

- **Product Owner** - Defines business value and priorities
- **Developer** - Provides technical feasibility
- **Tester/QA** - Identifies edge cases and scenarios

**Total**: 3-5 people maximum (more = less efficient)

### Time

- **Duration**: 15-25 minutes (strictly timeboxed)
- **Preparation**: 5 minutes before
- **Follow-up**: 10 minutes after

---

## Before the Workshop

### Step 1: Identify the Feature

Clearly state what feature you'll discuss.

**Example**: "Initialize CLI project command"

### Step 2: Prepare Blank Cards

Lay out blank cards in piles by color:

- Yellow (1 blank)
- Blue (10 blanks)
- Green (20 blanks)
- Red (10 blanks)

### Step 3: Set the Timer

Set a timer for 20 minutes. Enforce the time limit.

---

## During the Workshop (20 minutes)

### Part 1: Yellow Card - User Story (2 minutes)

**Who leads**: Product Owner

**Task**: Write the user story on the Yellow Card

**Format**:

```text
As a [user role]
I want [capability]
So that [business value]
```

**Example**:

```text
As a developer
I want to initialize a CLI project with one command
So that I can quickly start development
```

**Action**: Place the Yellow Card at the top of the table

---

### Part 2: Blue Cards - Acceptance Criteria (10 minutes)

**Who leads**: All participants brainstorm together

**Task**: Generate rules and acceptance criteria

**Guidelines**:

- Each Blue Card = one acceptance criterion
- Must be **measurable** (not subjective)
- Include functional AND non-functional requirements
- Aim for 2-6 Blue Cards

**Examples**:

```text
[BLUE-1] Creates project directory structure

[BLUE-2] Generates valid configuration file

[BLUE-3] Handles errors gracefully

[BLUE-4] Command completes in under 2 seconds

[BLUE-5] Works on Linux, macOS, and Windows
```

**Action**: Place Blue Cards below the Yellow Card

**Warning signs**:

- **>6 Blue Cards**: Feature is too large → Split it
- **<2 Blue Cards**: Feature is trivial or poorly understood

---

### Part 3: Green Cards - Concrete Examples (8 minutes)

**Who leads**: Tester/QA with input from all

**Task**: For each Blue Card, create concrete examples

**Format**: [Context] → [Action] → [Result]

**Guidelines**:

- Each Green Card = one specific example
- Include both success and error scenarios
- Aim for 2-4 Green Cards per Blue Card
- Use actual data/commands (not abstract)

**Examples** (for Blue Card 1):

```text
[GREEN-1a] Empty folder → run init → creates src/, tests/, docs/

[GREEN-1b] Existing project → run init → error "already initialized"

[GREEN-1c] Read-only folder → run init → error "permission denied"
```

**Examples** (for Blue Card 2):

```text
[GREEN-2a] New project → run init → creates r2r.yaml with defaults

[GREEN-2b] With --name flag → run init → r2r.yaml contains custom name
```

**Action**: Place Green Cards below their corresponding Blue Card

**Layout**:

```text
+---------------------+
| [YELLOW]            |
| User story          |
+---------------------+
          |
          v
+--------+  +--------+  +--------+
| BLUE-1 |  | BLUE-2 |  | BLUE-3 |
+--------+  +--------+  +--------+
    |           |           |
    v           v           v
 [GREEN-1a]  [GREEN-2a]  [GREEN-3a]
 [GREEN-1b]  [GREEN-2b]  [GREEN-3b]
 [GREEN-1c]
```

---

### Part 4: Red Cards - Questions (Ongoing)

**When**: Throughout the workshop

**Task**: Capture questions and uncertainties as they arise

**Guidelines**:

- Don't try to answer during the workshop
- Place Red Cards to the side
- Assign an owner to each Red Card

**Examples**:

```text
[RED-1] What if r2r.yaml already exists?
Owner: Product Owner

[RED-2] Should we support a --force flag?
Owner: Development Team

[RED-3] Do we need Windows compatibility?
Owner: Product Owner
```

**Action**: Place Red Cards to the side of the table

---

## Assess Readiness (2 minutes)

At the end, evaluate if the feature is ready for development:

### ✅ Ready to Implement

- 2-6 Blue Cards
- Each Blue Card has 2-4 Green Cards
- Few or no Red Cards (or Red Cards have owners)

**Action**: Proceed to creating specification.feature file

### ⚠️ Too Large

- \>6 Blue Cards

**Action**: Split into multiple features/stories

### ❌ Too Uncertain

- Many Red Cards without owners
- Fundamental questions unanswered

**Action**: Research/spike needed before implementation

---

## After the Workshop (10 minutes)

### Step 1: Take a Photo

Take a photo of the card layout for reference.

### Step 2: Convert Cards to specification.feature

Follow [Create Feature Spec](./create-specifications.md) guide to convert:

- **Yellow Card** → `specification.feature` Feature description (user story)
- **Blue Cards** → `Rule:` blocks (acceptance criteria)
- **Green Cards** → `Scenario:` blocks nested under Rules
- **Red Cards** → `issues.md` tracker

**File location**: `specs/<module>/<feature>/specification.feature`

### Step 3: Track Red Cards

For each Red Card:

1. Create entry in `specs/<module>/<feature>/issues.md`
2. Assign owner
3. Set deadline
4. Track resolution

**Example** (`issues.md`):

```markdown
# Open Questions

## RED-1: What if r2r.yaml already exists?

**Status**: Open
**Raised**: 2025-11-03
**Owner**: Product Owner
**Deadline**: Before implementation
**Resolution**: TBD
```

---

## Example Workshop Output

### Workshop Result

```text
[YELLOW]
As a developer, I want to initialize a CLI project with one command,
so that I can quickly start development

[BLUE-1] Creates project directory structure
  [GREEN-1a] Empty folder → init → creates src/, tests/, docs/
  [GREEN-1b] Existing project → init → error "already initialized"

[BLUE-2] Generates valid configuration file
  [GREEN-2a] New project → init → creates r2r.yaml with defaults
  [GREEN-2b] With --name flag → r2r.yaml contains custom name

[BLUE-3] Command completes in under 2 seconds
  [GREEN-3a] Standard project → measure time → <2s

[RED-1] What if r2r.yaml already exists? (Owner: Product Owner)
```

### Converted to specification.feature

**File**: `specs/cli/init-project/specification.feature`

```gherkin
@cli @critical @init
Feature: cli_init-project

  As a developer
  I want to initialize a CLI project with a single command
  So that I can quickly start development with proper structure

  Rule: Creates project directory structure

    # Green Card 1a
    @success @ac1
    Scenario: Initialize in empty directory creates structure
      Given I am in an empty folder
      When I run "r2r init"
      Then directories "src/", "tests/", "docs/" should exist

    # Green Card 1b
    @error @ac1
    Scenario: Initialize in existing project shows error
      Given I am in a directory with "r2r.yaml"
      When I run "r2r init"
      Then the command should fail
      And stderr should contain "already initialized"

  Rule: Generates valid configuration file

    # Green Card 2a
    @success @ac2
    Scenario: Initialize creates configuration with defaults
      Given I am in an empty folder
      When I run "r2r init"
      Then a file named "r2r.yaml" should be created
      And the file should contain valid YAML

    # Green Card 2b
    @flag @success @ac2
    Scenario: Initialize with custom name flag
      Given I am in an empty folder
      When I run "r2r init --name my-project"
      Then the file should contain "name: my-project"

  Rule: Command completes in under 2 seconds

    # Green Card 3a
    @success @ac3 @PV
    Scenario: Initialize completes within performance threshold
      Given I am in an empty folder
      When I run "r2r init"
      Then the command should complete within 2 seconds
```

**Card mapping**:

- **Yellow Card** → Feature description (lines 5-7)
- **Blue Card 1** → Rule: "Creates project directory structure" (line 9)
- **Green Cards 1a, 1b** → Scenarios under Rule 1 (lines 12-23)
- **Blue Card 2** → Rule: "Generates valid configuration file" (line 25)
- **Green Cards 2a, 2b** → Scenarios under Rule 2 (lines 28-41)
- **Blue Card 3** → Rule: "Command completes in under 2 seconds" (line 43)
- **Green Card 3a** → Scenario under Rule 3 (lines 46-50)

---

## Tips for Success

### Do

✅ **Keep it timeboxed** - Stop at 25 minutes regardless of progress
✅ **Focus on examples** - Concrete beats abstract
✅ **Include all perspectives** - Product, Dev, QA all contribute
✅ **Capture Red Cards** - Don't ignore uncertainties
✅ **Split large features** - If >6 Blue Cards, break it up

### Don't

❌ **Don't solve problems** - Workshop discovers requirements, doesn't design solutions
❌ **Don't argue** - If disagreement, write a Red Card
❌ **Don't over-detail** - Keep examples concise
❌ **Don't skip steps** - Yellow → Blue → Green order matters
❌ **Don't exceed time** - Time pressure forces clarity

---

## Troubleshooting

### Problem: Too many Blue Cards (>6)

**Solution**: Feature is too large

- Split into multiple user stories
- Keep related criteria together
- Each story should be independently deliverable

### Problem: Too many Red Cards

**Solution**: Feature is poorly understood

- Schedule research/spike
- Don't start implementation
- Answer Red Cards before workshop retry

### Problem: Can't think of Green Cards

**Solution**: Blue Card is too abstract

- Rewrite Blue Card more concretely
- Ask "Can you give me an example?"
- If still stuck, it might not be a real requirement

### Problem: Workshop takes too long

**Solution**: Not timeboxed strictly enough

- Set timer and enforce it
- Table discussions for later
- Focus on capturing, not solving

---

## Next Steps

- ✅ Workshop completed with cards
- **Next**: [Create Feature Spec](./create-specifications.md) to convert cards to files
- **Then**: Implement and [Run Tests](./run-tests.md)

---

## Related Documentation

- [Example Mapping](../../explanation/specifications/example-mapping.md) - Understanding the workshop technique
- [Three-Layer Approach](../../explanation/specifications/three-layer-approach.md) - How it all fits together
- [Example Mapping (External)](https://cucumber.io/blog/bdd/example-mapping-introduction/) - Original technique
