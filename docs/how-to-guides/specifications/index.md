# Specifications How-to Guides

Step-by-step guides for working with executable specifications.

---

## Overview

These guides provide task-oriented instructions for working with the three-layer specification approach (ATDD, BDD, TDD).

---

## Available Guides

### [Setup Godog](setup-godog.md)

Install and configure Godog for BDD specification testing.

**When to use**: Before creating your first specification.feature file

**What you'll learn**:

- Install Godog
- Configure Godog for your project
- Verify installation
- Create initial test harness

### [Establish Ubiquitous Language](establish-ubiquitous-language.md)

Build shared domain vocabulary that flows from workshops through to code.

**When to use**: Starting a new project or major feature area

**What you'll learn**:

- Run Event Storming to discover vocabulary
- Document glossary and bounded contexts
- Apply vocabulary in Example Mapping
- Write specifications using domain terms
- Implement code with domain language
- Evolve language continuously

### [Run Event Storming Workshop](run-event-storming.md)

Facilitate an Event Storming workshop to discover domain language and business processes.

**When to use**: Starting a project, major feature area, or when domain understanding is unclear

**What you'll learn**:

- Choose the right Event Storming format
- Set up space and materials
- Facilitate the workshop (silent storm, timeline, commands, actors, policies)
- Capture domain vocabulary and glossary
- Handle common problems

### [Run Example Mapping Workshop](run-example-mapping.md)

Facilitate an Example Mapping workshop to discover requirements for specific features.

**When to use**: After Event Storming, before implementing any specific feature

**What you'll learn**:

- Prepare for the workshop
- Run the 15-25 minute session
- Use colored cards effectively
- Convert cards to specification.feature
- Handle questions and blockers

### [Create Specifications](create-specifications.md)

Create specification.feature file with Rules and Scenarios for a new feature.

**When to use**: After Example Mapping, before implementation

**What you'll learn**:

- Create feature directory structure
- Write specification.feature with Feature → Rule → Scenario structure
- Map Example Mapping cards to Gherkin
- Link scenarios to acceptance criteria with @ac tags
- Implement step definitions

### [Link Risk Controls](link-risk-controls.md)

Link specification scenarios to risk control requirements using @risk tags.

**When to use**: When implementing or validating risk control requirements from assessments

**What you'll learn**:

- Define risk controls in specs/risk-controls/
- Tag implementation scenarios with @risk<ID>
- Verify traceability between controls and implementations
- Generate coverage reports
- Maintain audit trail

### [Run Tests](run-tests.md)

Execute tests at all three layers (ATDD, BDD, TDD).

**When to use**: During development and in CI/CD

**What you'll learn**:

- Run all tests
- Run specific feature tests
- Run tests by tags
- Generate reports
- Integrate with CI/CD

---

## Workflow Summary

```text
1. Setup Godog
   ↓
2. Event Storming → Discover vocabulary → Document glossary
   ↓
3. Example Mapping → Apply vocabulary → Produce cards
   ↓
4. Create specification.feature
   - Yellow Card → Feature description
   - Blue Cards → Rule blocks
   - Green Cards → Scenario blocks
   ↓
5. Implement step definitions (src/<module>/tests/)
   ↓
6. Run Tests → Validate → Iterate
```

---

## Related Documents

- **Want to understand WHY?** See [Specifications Explanation](../../explanation/specifications/index.md)
- **Need technical reference?** See [Specifications Reference](../../reference/specifications/index.md)
