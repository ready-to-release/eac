# Example Mapping

{{ page_breadcrumb() }}

Workshop technique for discovering requirements through collaborative conversation in 15-25 minutes.

---

## What is Example Mapping?

**[Example Mapping](https://cucumber.io/blog/bdd/example-mapping-introduction/)** is a time-boxed collaborative workshop that uses colored index cards to discover acceptance criteria and concrete examples before development begins.

**Core purpose**: Answers three questions:

1. What does this feature need to do? (Acceptance criteria)
2. What are concrete examples? (Test scenarios)
3. What don't we know yet? (Questions and risks)

**Prerequisites**: Establish **Ubiquitous Language** through **Event Storming** first. See: [Event Storming](./event-storming-overview.md) and [Ubiquitous Language](../concepts/ubiquitous-language.md)

**Why use it?** Discovers requirements early (less than 25 min conversation vs days/weeks of coding), produces concrete measurable criteria (not subjective), creates fast feedback loop, builds collaborative understanding.

---

## Quick Start

> **Running Your First Example Mapping Session:**
>
> 1. **Prepare (10 min)**: Get colored index cards (Yellow, Blue, Green, Pink) and invite Product Owner + Developer + Tester
> 2. **Place Yellow Card (2 min)**: Write user story "As a [role] I want [capability] So that [value]"
> 3. **Generate Blue Cards (10 min)**: Brainstorm 2-4 acceptance criteria (measurable rules)
> 4. **Add Green Cards (10 min)**: Write 2-4 concrete examples per Blue Card
> 5. **Capture Pink Cards (ongoing)**: Write down questions, DON'T solve them during workshop
> 6. **Assess Readiness (3 min)**: <4 Blue, <4 Green each, no blocking Pink = Ready to implement
>
> **Time limit**: 25 minutes maximum. If longer, split the feature.
>
> See [Workshop Process](#workshop-process) for detailed facilitation guide.

---

## The Four Card Colors

| Card Color                                                                                                                                            | Represents                              | Maps To             | Location          |
| ----------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------- | ------------------- | ----------------- |
| <svg width="24" height="24" xmlns="http://www.w3.org/2000/svg"><rect width="24" height="24" rx="2" fill="#F7E55A" stroke="#D6C645"/></svg> **Yellow** | User Story: WHO, WHAT, WHY              | Feature description | 1 per feature     |
| <svg width="24" height="24" xmlns="http://www.w3.org/2000/svg"><rect width="24" height="24" rx="2" fill="#5BA3F7" stroke="#4A89D6"/></svg> **Blue**   | Acceptance Criteria: Success conditions | `Rule:` blocks      | < 4 per story     |
| <svg width="24" height="24" xmlns="http://www.w3.org/2000/svg"><rect width="24" height="24" rx="2" fill="#7EDC7A" stroke="#68B666"/></svg> **Green**  | Concrete Examples: Specific scenarios   | `Scenario:` blocks  | < 4 per criterion |
| <svg width="24" height="24" xmlns="http://www.w3.org/2000/svg"><rect width="24" height="24" rx="2" fill="#FF7EBF" stroke="#D66A9F"/></svg> **Pink**   | Questions/Unknowns: Blockers            | issues.md           | See below         |

### Card Guidelines

**Yellow** (User Story):

```text
As a [role]
I want [capability]
So that [business value]
```

**Blue** (Rules): Measurable business rules

- ✅ "Creates 3 directories (src/, tests/, docs/)"
- ❌ "Interface is user-friendly" (not measurable)

**Green** (Examples): [Context] → [Action] → [Result]

- ✅ "Empty folder → init → creates src/, tests/, docs/"
- ❌ "It creates directories" (too abstract)

**Pink** (Questions): Capture without solving during workshop

- "What if r2r.yaml already exists?"
- "Should we support a --force flag?"

---

## Workshop Process

**Participants**: Product Owner + Engineer + Tester
**Duration**: 15-25 minutes (strictly time-boxed)

### Steps

1. **Place Yellow Card** (2 min) - Write user story
2. **Generate Blue Cards** (8-12 min) - Brainstorm 2-6 acceptance criteria
3. **Create Green Cards** (5-10 min) - Write 2-4 concrete examples per Blue Card
4. **Capture Pink Cards** (ongoing) - Write down questions, don't solve them
5. **Assess Readiness** (2 min) - Ready to implement, too large, or too uncertain?

### Readiness Assessment

**✅ Ready**: <4 Blue Cards, <4 Green Cards each, no Pink cards you can't answer within a day
**⚠️ Too Large**: >4 Blue Cards or >25 minutes → Split into multiple features
**❌ Too Uncertain**: Blocking Pink Cards → Research spike, then re-run

---

## Visual Layout

```mermaid
flowchart TD
    Story["User Story<br/>As an engineer..."]

    Story --> Rule1["AC 1<br/>Creates dirs"]
    Story --> Rule2["AC 2<br/>Generates config"]
    Story --> Rule3["AC 3<br/>Handles errors"]

    Rule1 --> Example1a["Example 1a"]
    Rule1 --> Example1b["Example 1b"]

    Rule2 --> Example2a["Example 2a"]
    Rule2 --> Example2b["Example 2b"]

    Rule3 --> Example3a["Example 3a"]
    Rule3 --> Example3b["Example 3b"]

    Question1["What if config exists?"]
    Question2["Support --force flag?"]

```

**Workshop tips**: Use physical cards on table (or Miro/Mural for virtual), everyone can add cards, take photos for documentation.

---

## Facilitation Guide

### Session Setup

**Time**: **25 minutes** - strictly time-boxed

**Participants**:

- Product Owner (business perspective)
- Developer (technical perspective)
- Tester (quality perspective)

**Materials**:

- Yellow index cards (user stories)
- Blue index cards (rules/acceptance criteria)
- Green index cards (examples/scenarios)
- Red/Pink index cards (questions/unknowns)
- Timer (visible to all)

### Detailed Steps

#### Step 1: Present the Story (2 minutes)

Write the user story on a **Yellow card**:

```text
As a [role]
I want to [capability]
So that [business value]
```

Place at top of workspace.

#### Step 2: Discover Rules (8 minutes)

Ask: "What acceptance criteria must be satisfied?"

Write each on a **Blue card**:

- "Must validate email format"
- "Must send verification email"
- "Must prevent duplicate accounts"

Place Blue cards in a row below Yellow card.

#### Step 3: Add Examples (12 minutes)

For each Blue card, ask: "Can you give me an example?"

Write each on a **Green card**:

- "User enters valid email, receives confirmation"
- "User enters invalid format, sees error"
- "User enters existing email, directed to login"

Place Green cards below their Blue card.

#### Step 4: Capture Questions (ongoing)

When you encounter unknowns, write on **Red/Pink card**:

- "What happens if email service is down?"
- "How long is verification link valid?"

Place to the side. Red cards must be resolved before implementation.

#### Step 5: Assess Readiness (3 minutes)

**Ready to implement if**:

- 1 Yellow card (story)
- ≤4 Blue cards (acceptance criteria)
- ≤4 Green cards per Blue card
- No blocking Red cards

**Not ready if**:

- Too many Red cards → stop and research
- Too many Blue cards → story too big, split it

### Session Signals

**Stop and Split**:

- More than 4 Blue cards
- Green cards are hard to think of
- Lots of disagreement

**Stop and Research**:

- Many Red cards accumulating
- Fundamental questions unresolved
- Scope unclear

**Ready to Go**:

- Clear acceptance criteria
- Concrete examples
- Questions resolved

### Virtual Sessions

For remote teams:

- Use Miro, Mural, or FigJam
- Color-coded sticky notes
- Screen share for all participants
- Timebox strictly (use visible timer)

---

## From Cards to Gherkin

```mermaid
flowchart LR
    subgraph Workshop["Example Mapping Cards"]
        Story["Yellow<br/>User Story"]
        Rule["Blue<br/>AC"]
        Example["Green<br/>Example"]
        Question["Pink<br/>Question"]
    end

    subgraph Spec["specification.feature"]
        Feature["Feature:<br/>description"]
        RuleBlock["Rule:<br/>AC block"]
        Scenario["Scenario:<br/>under Rule"]
    end

    subgraph Issues["issues.md"]
        Issue["## Questions"]
    end

    Story --> Feature
    Rule --> RuleBlock
    Example --> Scenario
    Question --> Issue

```

### Example Conversion

**Workshop Cards**:

```text
🟡 As an engineer, I want to initialize a CLI project with one command
🔵 [BLUE-1] Creates project directory structure
🟢 [GREEN-1a] Empty folder → init → creates src/, tests/, docs/
🟢 [GREEN-1b] Existing project → init → error "already initialized"
```

**Resulting Gherkin** (`specs/cli/init-project/specification.feature`):

```gherkin
Feature: cli_init-project

  As an engineer
  I want to initialize a CLI project with one command
  So that I can quickly start development

  Rule: Creates project directory structure

    @ov
    Scenario: Initialize in empty directory creates structure
      Given I am in an empty folder
      When I run "r2r init my-project"
      Then a directory named "my-project/src/" should exist
      And a directory named "my-project/tests/" should exist
      And a directory named "my-project/docs/" should exist

    @ov
    Scenario: Initialize in existing project shows error
      Given I am in a directory with "r2r.yaml"
      When I run "r2r init"
      Then the command should fail
      And stderr should contain "already initialized"
```

### Translation Guide

Follow these steps to convert Example Mapping cards to Gherkin:

### Step 1: Yellow Card → Feature

**Yellow Card**:

```text
As a developer
I want to initialize a project
So that I can start coding quickly
```

**Gherkin**:

```gherkin
Feature: cli_init-project

  As a developer
  I want to initialize a project
  So that I can start coding quickly
```

### Step 2: Blue Cards → Rules

**Blue Cards**:

- "Creates project structure"
- "Configuration has defaults"
- "Prevents overwriting existing projects"

**Gherkin**:

```gherkin
Rule: Creates project directory structure

Rule: Configuration file contains default values

Rule: Existing projects are not overwritten
```

### Step 3: Green Cards → Scenarios

**Green Cards under "Creates project structure"**:

- "Empty directory → creates files"
- "With subdirectory → creates in current"

**Gherkin**:

```gherkin
Rule: Creates project directory structure

  @ov @ac1
  Scenario: Initialize in empty directory
    Given I am in an empty folder
    When I run "r2r init"
    Then a file named "r2r.yaml" should be created
    And a directory named "specs/" should be created

  @ov @ac1
  Scenario: Initialize with existing subdirectories
    Given I am in a directory with "src/"
    When I run "r2r init"
    Then "r2r.yaml" should be created in current directory
    And "src/" should remain unchanged
```

### Step 4: Red Cards → Actions

**Red Cards**:

- "What if no write permission?"
- "What about .gitignore?"

**Actions**:

1. Resolve before implementation (ask product owner)
2. Add as new scenarios once resolved
3. Or document in `issues.md` if deferred

### Translation Checklist

- [ ] One Feature per file
- [ ] Feature name follows `module_feature-name` format
- [ ] User story in Feature description
- [ ] One Rule per Blue card
- [ ] Scenarios under their Rule
- [ ] All scenarios have `@ov` (or other verification tag)
- [ ] All scenarios have `@ac<n>` linking to Rule
- [ ] Red card questions resolved or documented

---

## After the Workshop

### Immediate (Same Day)

1. **Write specification.feature** - Convert cards to Gherkin while context is fresh
2. **Share for review** - Product Owner, Engineers, QA validate
3. **Document Pink Cards** - Create `issues.md` with ownership and deadlines

### Short-term (1-2 Days)

1. **Incorporate feedback** - Refine ambiguous steps, add missing criteria
2. **Resolve Pink Cards** - Research, mini-sessions, spike implementations
3. **Confirm scope** - All three amigos agree before implementation starts

### During Implementation (1 Week)

1. **Discover edge cases** - Unit tests reveals boundary conditions → Add scenarios
2. **Refine language** - Vague steps become precise → Update specification
3. **Keep synchronized** - Commit spec changes with code changes

### After Implementation

1. **Retrospective review** - Did spec match what was built?
2. **Refactor for clarity** - Consolidate duplicates, simplify verbose scenarios
3. **Document lessons learned** - Improve next workshop

**Remember**: Specifications are **living documents**. They evolve through discovery, feedback, refinement, and iteration. See: [Review and Iterate](../quality/review-and-iterate.md)

---

## Best Practices

### Do's ✅

- Time-box strictly (25 min max)
- Use domain language from Event Storming
- Be specific: "Creates 3 directories" not "Creates directories"
- Include error cases (at least one per Blue Card)
- Set Pink Cards aside (don't solve during workshop)
- Invite all three amigos (Product + Dev + Test)

### Don'ts ❌

- Don't make Blue Cards subjective → Use measurable criteria
- Don't use abstract Green Cards → Use specific input/output
- Don't debate Pink Cards → Write down, resolve later
- Don't exceed 6 Blue Cards → Split the feature
- Don't skip edge cases → Error scenarios matter

---

## Common Anti-Patterns

❌ **Solution in User Story**: "I want a YAML config file" → ✅ "I want to configure my CLI tool"
❌ **Unmeasurable criteria**: "Performance is good" → ✅ "Completes in <2 seconds"
❌ **Too many Green Cards**: 10+ examples for one Blue Card → ✅ Split into multiple Blue Cards
❌ **Implementation details**: "Parser loads YAML → deserializes" → ✅ "Valid YAML → init succeeds"

---

## When to Split Features

Split when:

- **>6 Blue Cards** - Group related criteria into separate features
- **Workshop >25 minutes** - Prioritize and split
- **Every Blue Card has 6+ Green Cards** - Keep MVP, defer edge cases

---

## See Also

- [BDD Fundamentals](../concepts/bdd-fundamentals.md) - Concepts behind specifications
- [Three-Layer Approach](../concepts/three-layer-approach.md) - How Example Mapping fits the workflow
- [Review and Iterate](../quality/review-and-iterate.md) - Maintaining living specifications

## Quick Reference

- [Example Mapping Cards](card-reference.md) - Card colors and quantities

{{ diataxis_footer() }}
