# Discovery Techniques
> **Workshop methods for discovering requirements and building shared understanding**

Learn collaborative techniques for discovering business rules, acceptance criteria, and domain language before development begins.

---

## Overview

This section covers:

- **Event Storming** - Collaborative workshop for discovering domain events, processes, and Ubiquitous Language
- **Example Mapping** - Time-boxed technique for discovering acceptance criteria with colored cards
- **Card Reference** - Color meanings and quantities for Example Mapping workshops

Discovery happens BEFORE implementation. These workshops build shared understanding across business, development, and testing perspectives.

---

## In This Section

| Topic                                                           | Description                                              |
| --------------------------------------------------------------- | -------------------------------------------------------- |
| [Event Storming Overview](./event-storming-overview.md)         | What is Event Storming and when to use it                |
| [Event Storming Formats](./event-storming-formats.md)           | Big Picture, Process Modeling, Software Design           |
| [Event Storming Facilitation](./event-storming-facilitation.md) | Running an Event Storming session                        |
| [Example Mapping](./example-mapping.md)                         | Workshop for discovering acceptance criteria (15-25 min) |
| [Card Reference](./card-reference.md)                           | Example Mapping card colors and meanings                 |

---

## Discovery Workflow

```text
Event Storming (2-4 hours)
  Discover: Domain events, Ubiquitous Language, Bounded Contexts
    ↓
Example Mapping (15-25 min per feature)
  Discover: User stories, Acceptance criteria, Concrete examples
    ↓
Gherkin Specifications
  Document: Feature files with Rules and Scenarios
```

### When to Use Each Technique

**Event Storming** - Use when:

- Starting new project or major feature area
- Understanding complex domain with many stakeholders
- Discovering Ubiquitous Language for team alignment
- Identifying system boundaries and bounded contexts

**Example Mapping** - Use when:

- Planning specific feature before development
- Breaking down user stories into acceptance criteria
- Discovering concrete examples for testing
- Validating feature scope (should take <25 min)

---

## Key Principles

### 1. Discovery is collaborative

Both techniques require multiple perspectives:

- **Business**: Knows domain rules and desired outcomes
- **Development**: Understands technical constraints
- **Testing**: Identifies edge cases and failure modes

### 2. Discovery is time-boxed

**Event Storming**: 2-4 hours maximum

- Longer sessions lose focus and energy
- Split large domains across multiple sessions

**Example Mapping**: 15-25 minutes per feature

>

- > 25 min means feature is too large (split it)
- <15 min might mean missing edge cases

### 3. Capture questions, don't solve them

Use **pink cards** (Example Mapping) to document unknowns:

- "What if the external API is down?"
- "Should we support concurrent sessions?"

Resolve these AFTER the workshop, not during.

### 4. Use domain language

Establish and use **Ubiquitous Language** from Event Storming:

- Consistent terms across workshops, specs, and code
- Bridges business and technical conversations
- Reduces translation errors and misunderstandings

---

## Quick Start

### Running Your First Event Storming

**Preparation (30 min)**:

1. Define scope: "Order fulfillment process" not "entire e-commerce system"
2. Invite stakeholders: Domain experts + developers + facilitator
3. Materials: Sticky notes (orange, blue, yellow, pink), wall space, markers

**During Session (2-3 hours)**:

1. Chaotic exploration (30 min) - Everyone writes domain events
2. Enforce timeline (20 min) - Arrange events chronologically
3. Add commands and actors (30 min) - Who triggers events?
4. Identify policies (20 min) - Business rules connecting events
5. Capture outcomes (20 min) - Take photos, document language

**After Session**:

- Document Ubiquitous Language glossary
- Identify features for Example Mapping
- Update or create domain model

See: [Event Storming Facilitation](./event-storming-facilitation.md)

### Running Your First Example Mapping

**Preparation (10 min)**:

1. User story identified: "As a developer I want to initialize a project"
2. Materials: Yellow, blue, green, pink index cards
3. Attendees: Product Owner + Developer + Tester
4. Timer: Set to 25 minutes

**During Session (15-25 min)**:

1. Place yellow card with user story (2 min)
2. Generate blue cards for acceptance criteria (10 min)
3. Add green cards with concrete examples (10 min)
4. Capture pink cards for questions (ongoing)
5. Assess readiness (3 min)

**After Session**:

- Convert to Gherkin feature file (same day)
- Resolve pink card questions (1-2 days)
- Share with stakeholders for review

See: [Example Mapping](./example-mapping.md)

---

## Example Mapping Card Colors

Quick reference:

| Color     | Represents          | Maps To             | Quantity Limit                |
| --------- | ------------------- | ------------------- | ----------------------------- |
| 🟡 Yellow | User Story          | Feature description | 1 per feature                 |
| 🔵 Blue   | Acceptance Criteria | `Rule:` blocks      | ≤ 4 per story                 |
| 🟢 Green  | Concrete Examples   | `Scenario:` blocks  | ≤ 4 per criterion             |
| 🟣 Pink   | Questions/Unknowns  | issues.md           | Resolve before implementation |

See: [Card Reference](./card-reference.md) for detailed usage.

---

## Common Pitfalls

### Event Storming

❌ **Too broad scope** → Session never ends, participants lose focus
✅ **Bounded scope** → "User onboarding flow" not "entire application"

❌ **Only developers attend** → Miss business rules and edge cases
✅ **Cross-functional team** → Domain experts + developers + testers

❌ **Try to build software design** → Wrong format, wrong stage
✅ **Discover domain events first** → Design comes later

### Example Mapping

❌ **Solving pink card questions during workshop** → Wastes time, blocks progress
✅ **Capture and defer** → Document questions, resolve after

❌ **>4 blue cards** → Feature too large, will take weeks to implement
✅ **Split feature** → Multiple smaller features, each <25 min workshop

❌ **Abstract examples** → "User enters valid data" (not testable)
✅ **Concrete examples** → "User enters '<john@example.com>'" (testable)

---

## Related Documentation

### Core Concepts

- [BDD Fundamentals](../concepts/bdd-fundamentals.md) - Behavior-Driven Development principles
- [Ubiquitous Language](../concepts/ubiquitous-language.md) - Shared domain vocabulary
- [Three-Layer Approach](../concepts/three-layer-approach.md) - Rules/Scenarios/Unit Tests

### Organization

- [File Structure](../organization/file-structure.md) - Where to put feature files
- [Organizing Rules](../organization/organizing-rules.md) - Structuring acceptance criteria
- [Size Guidelines](../organization/size-guidelines.md) - Rule and scenario count limits

### Quality

- [Review and Iterate](../quality/review-and-iterate.md) - Maintaining living specifications
