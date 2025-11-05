# Event Storming: Domain Discovery Workshop

> **Collaborative technique for discovering Ubiquitous Language through domain exploration**

---

## Overview

Ubiquitous Language doesn't emerge spontaneously—it's **discovered** through collaborative domain exploration. Event Storming is a powerful technique for this discovery.

### What is Event Storming?

> Event Storming is a group of collaborative modeling techniques that help teams understand complex domains by visually mapping out key events. It is designed to help people from different parts of an organization create a shared understanding of the problem space they are going to work with.

Event Storming uses a game-like format with rules, a "board" (brown packing paper), and grammar (colored sticky notes). This collaborative approach surfaces the domain vocabulary through structured conversation.

### Why Event Storming Matters for Specifications

Event Storming creates the foundation for all specifications work:

- **Discovers domain vocabulary** that flows into Example Mapping and Gherkin
- **Surfaces terminology conflicts** before they become specification problems
- **Identifies bounded contexts** where different languages apply
- **Reveals business processes** that need to be specified
- **Builds shared understanding** across business and technical teams

---

## The Three Formats

Event Storming has three formats, each serving a different purpose:

### Big Picture Event Storming

**Purpose**: Align key stakeholders from different parts of an organization

**Activities**:

- Map out how the business works at a high level
- Identify major domain events, actors, and boundaries
- Surface terminology conflicts and context boundaries

**Duration**: 4-8 hours (can be split across multiple sessions)

**Participants**: Cross-functional representation from all areas

**Outcome**: Shared understanding of the business domain

**When to use**:

- Starting a new project or initiative
- Onboarding new team members to complex domains
- Identifying bounded contexts and system boundaries
- Before planning major architectural changes

### Process Modeling Event Storming

**Purpose**: Create detailed understanding of specific business processes

**Activities**:

- Map information flows, decision points, and policies
- Identify who or what makes decisions and what data is needed
- Detail the sequence and dependencies of domain events

**Duration**: 2-4 hours per process

**Participants**: Domain experts, developers, QA for the specific process

**Outcome**: Clear process flows with domain vocabulary

**When to use**:

- Before implementing a specific feature
- Refining understanding of a business process
- Before Example Mapping workshops
- Troubleshooting process-related bugs

### Software Design Event Storming

**Purpose**: Distill technical solutions from business knowledge

**Activities**:

- Map domain concepts to software structures
- Identify aggregates, commands, and events
- Design bounded contexts and integration points

**Duration**: 2-4 hours

**Participants**: Development team, technical leads, architects

**Outcome**: Technical design aligned with domain language

**When to use**:

- After Process Modeling, before implementation
- Designing system architecture
- Refactoring existing systems
- Planning technical migrations

---

## Workshop Structure

### Materials Needed

**Required**:

- Brown packing paper (6-8 meters long)
- Sticky notes in multiple colors:
  - Orange (domain events)
  - Blue (commands/actions)
  - Yellow (actors/personas)
  - Pink (policies/business rules)
  - Purple (external systems)
  - Red (problems/questions)
  - Green (opportunities)
- Markers (black, fine tip)
- Wall space (long, uninterrupted)

**Optional**:

- Timer (for time-boxing discussions)
- Camera (for documentation)
- Whiteboard (for capturing definitions)

### Facilitation Approach

**Opening (10-15 minutes)**:

1. Explain the format and rules
2. Clarify the scope (Big Picture vs Process Modeling)
3. Set ground rules (no laptops, stay engaged, park questions)

**Discovery Phase (varies by format)**:

1. Silent storm: Everyone writes domain events individually
2. Timeline: Arrange events on paper in rough chronological order
3. Clustering: Group related events
4. Fill gaps: Add commands, actors, policies

**Discussion Phase**:

1. Walk through the timeline together
2. Ask clarifying questions
3. Surface definitions and terminology
4. Identify conflicts and assumptions

**Closing (15-20 minutes)**:

1. Highlight key insights
2. Identify open questions
3. Capture glossary of terms
4. Define next steps

### Sticky Note Grammar

Event Storming uses color-coded sticky notes as a visual language:

| Color | Represents | Format | Example |
|-------|------------|--------|---------|
| **Orange** | Domain Event | Past tense, something that happened | "Order Placed", "Payment Received" |
| **Blue** | Command/Action | Imperative, action that triggers event | "Place Order", "Approve Request" |
| **Yellow** | Actor/Persona | Who initiates commands | "Customer", "Manager", "System" |
| **Pink** | Policy/Rule | Business rule that triggers commands | "When order >$10k → require approval" |
| **Purple** | External System | Systems outside your control | "Payment Gateway", "Email Service" |
| **Red** | Problem/Question | Uncertainty or blocker | "What if...?", "How does...?" |
| **Green** | Opportunity | Improvement ideas | "Could automate...", "Might simplify..." |
| **White** | Definition | Glossary term clarification | "Approved = manager signed off" |

---

## Key Outputs

Event Storming workshops surface critical information for specifications:

### Domain Events

**What**: Things that happen in the business

**Examples**: "Order Placed", "Payment Received", "Approval Granted"

**How it helps specifications**: Events become the foundation for Given/When/Then scenarios

### Actors/Personas

**What**: Who initiates or responds to events

**Examples**: "Customer", "Manager", "System Administrator"

**How it helps specifications**: Actors appear in user stories ("As a customer...")

### Commands

**What**: Actions that cause events

**Examples**: "Place Order", "Approve Request", "Cancel Subscription"

**How it helps specifications**: Commands become the "When" in scenarios

### Policies

**What**: Business rules that govern behavior

**Examples**: "Orders over $10,000 require manager approval", "Inactive accounts are archived after 90 days"

**How it helps specifications**: Policies become acceptance criteria (Rule blocks)

### Bounded Contexts

**What**: Where different terminologies apply

**Examples**: "Customer" in Sales context vs Support context

**How it helps specifications**: Prevents terminology confusion in specifications

### Domain Vocabulary

**Most importantly**: Event Storming surfaces the **Ubiquitous Language** through collaborative discussion. When domain experts explain processes, they use their natural vocabulary. When developers ask clarifying questions, misunderstandings emerge and get resolved.

**Capture vocabulary**:

- Use white definition stickies during the session
- Maintain a glossary on a separate board/document
- Document context boundaries and term meanings

---

## From Event Storming to Specifications

### Event Storming → Example Mapping

Event Storming provides the vocabulary for Example Mapping workshops:

**Event Storming discovered**:

- Domain event: "Order Approved"
- Actor: "Manager"
- Policy: "Large orders require approval"

**Example Mapping applies this**:

- Yellow card: "As a manager, I want to approve large orders..."
- Blue card: "Orders over $10,000 must be approved by a manager"
- Green card: "Given an order of $15,000, when manager approves..."

**See**: [Example Mapping](./example-mapping.md) for detailed workshop process.

### Event Storming → Gherkin Specifications

Event Storming vocabulary flows directly into Gherkin:

**Event Storming discovered**:

- Domain terms: "Order", "Manager", "Approved"
- Process flow: Place order → Check amount → Require approval → Manager approves

**Gherkin specification**:

```gherkin
Feature: Order Approval Process

  Rule: Large orders require manager approval

    @success @ac1
    Scenario: Manager approves large order
      Given an order with amount $15,000
      And the order status is "Awaiting Approval"
      When the manager approves the order
      Then the order status should be "Approved"
      And the order should be eligible for fulfillment
```

**Notice**: Every term comes from Event Storming—"order", "manager", "approved", "awaiting approval".

### Maintaining Glossaries

**During Event Storming**: Use white definition stickies to capture term meanings

**After Event Storming**: Create a glossary document

**Example glossary format**:

```markdown
# Domain Glossary: Order Management Context

**Order**: A customer request to purchase products
- Status values: "Pending", "Awaiting Approval", "Approved", "Fulfilled"
- Bounded context: Order Management

**Manager**: An employee with approval authority
- Can: Approve orders, reject orders, request more information
- Bounded context: Order Management

**Approved**: Business acceptance of an order
- Trigger: Manager approval action
- Result: Order eligible for fulfillment
- NOT the same as "Verified" (data correctness)
```

---

## Workshop Best Practices

### Time Boxing

**Big Picture**: 4-8 hours (can split across multiple sessions)

- Session 1: Silent storm + timeline (2-4 hours)
- Session 2: Discussion + refinement (2-4 hours)

**Process Modeling**: 2-4 hours per process

- Focus on one process at a time
- Take breaks every 90 minutes

**Software Design**: 2-4 hours

- Requires prior Process Modeling output
- More focused technical discussion

### Facilitation Tips

**Start broad, then focus**:

- Begin with high-level events
- Drill down into details progressively
- Don't get stuck on edge cases early

**Encourage questions**:

- Red sticky for every "What if...?"
- Don't try to answer all questions immediately
- Some questions surface later understanding

**Maintain energy**:

- Stand, don't sit
- Take regular breaks
- Keep discussions moving
- Park off-topic conversations

**Capture everything**:

- Take photos of the final timeline
- Document definitions on whiteboard
- Record open questions
- Note follow-up actions

**Stay visual**:

- Sticky notes only—no laptops during discovery
- Use the paper as the source of truth
- Arrange spatially to show relationships

### Common Pitfalls

**Going too detailed too early**:

- Problem: Team gets stuck on edge cases
- Solution: "Park it" on a red sticky, move forward

**Skipping domain experts**:

- Problem: Developers make assumptions
- Solution: Require domain expert participation

**Focusing on implementation**:

- Problem: Discussion drifts to "how we'll build it"
- Solution: Stay focused on "what happens in the business"

**Not capturing definitions**:

- Problem: Team uses terms differently
- Solution: Use white stickies for definitions immediately

**Insufficient wall space**:

- Problem: Timeline feels cramped
- Solution: Book a room with long walls, use multiple sheets

### Follow-Up Actions

**Immediately after**:

1. Photograph the entire timeline
2. Transcribe glossary terms
3. Identify first features to specify
4. Schedule Example Mapping sessions

**Within a week**:

1. Digitize the timeline (optional—photos often sufficient)
2. Create/update bounded context documentation
3. Share photos and glossary with stakeholders
4. Begin Example Mapping for first features

**Ongoing**:

1. Reference Event Storming output in Example Mapping
2. Update glossary as understanding evolves
3. Schedule follow-up Event Storming sessions (quarterly or when major changes occur)

---

## Further Reading

### Books

- [Introduction to Event Storming](https://leanpub.com/introducing_eventstorming) - Alberto Brandolini's book on how to do Event Storming
- [Domain-Driven Design](https://www.domainlanguage.com/ddd/) - Eric Evans' foundational work

### Online Resources

- [eventstorming.com](https://www.eventstorming.com) - A site full of resources
- [Awesome EventStorming](https://github.com/mariuszgil/awesome-eventstorming) - A curated list of material and links
- [Alberto Brandolini's Blog](https://blog.avanscoperta.it/author/ziobrando/) - Original creator's insights

---

## Related Documentation

### For Understanding

- [Ubiquitous Language](./ubiquitous-language.md) - DDD foundation for shared vocabulary
- [Three-Layer Testing Approach](./three-layer-approach.md) - How Event Storming fits the testing strategy

### For Doing

- [Example Mapping](./example-mapping.md) - Requirements discovery using Event Storming vocabulary
- [How to Establish Ubiquitous Language](../../how-to-guides/specifications/establish-ubiquitous-language.md) - Step-by-step guide

### For Reference

- [Gherkin Format](../../reference/specifications/gherkin-format.md) - How to write specifications using discovered vocabulary

---

## Key Takeaways

1. **Event Storming discovers domain vocabulary collaboratively** - Through visual mapping and structured conversation
2. **Three formats serve different purposes** - Big Picture, Process Modeling, and Software Design
3. **Sticky note grammar creates visual language** - Colors represent different domain concepts
4. **Key outputs include events, actors, commands, policies** - But most importantly: Ubiquitous Language
5. **Vocabulary flows to Example Mapping and specifications** - Same terms used throughout
6. **Glossaries maintain shared understanding** - Document definitions and context boundaries
7. **Regular sessions keep understanding current** - Schedule quarterly or when domain changes
