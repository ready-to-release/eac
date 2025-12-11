# Event Storming Formats

{{ page_breadcrumb() }}

> **Big Picture, Process Modeling, and Software Design formats**

Event Storming has three formats, each serving a different purpose.

---

## Big Picture Event Storming

**Purpose**: Align key stakeholders from different parts of an organization

**Activities**:

- Map out how the business works at a high level
- Identify major domain events, actors, and boundaries
- Surface terminology conflicts and context boundaries

**Duration**: 4-8 hours (can be split across multiple sessions)

**Participants**: Cross-functional representation from all areas

**Outcome**: Shared understanding of the business domain

### When to Use Big Picture

- Starting a new project or initiative
- Onboarding new team members to complex domains
- Identifying bounded contexts and system boundaries
- Before planning major architectural changes

### Big Picture Example Output

After a Big Picture session, you'll have:

- **Timeline of major events** spanning the entire business domain
- **Bounded context boundaries** showing where terminology changes
- **Key actors** who interact with the system
- **External systems** that integrate with your domain
- **Questions and conflicts** surfaced but not yet resolved

---

## Process Modeling Event Storming

**Purpose**: Create detailed understanding of specific business processes

**Activities**:

- Map information flows, decision points, and policies
- Identify who or what makes decisions and what data is needed
- Detail the sequence and dependencies of domain events

**Duration**: 2-4 hours per process

**Participants**: Domain experts, developers, QA for the specific process

**Outcome**: Clear process flows with domain vocabulary

### When to Use Process Modeling

- Before implementing a specific feature
- Refining understanding of a business process
- Before Example Mapping workshops
- Troubleshooting process-related bugs

### Process Modeling Example Output

After a Process Modeling session, you'll have:

- **Detailed event flow** for one specific process
- **Decision points** (pink stickies showing business rules)
- **Information requirements** (what data is needed when)
- **Actor interactions** (who does what, when)
- **Policy definitions** (business rules that govern the process)

**Example**: "Order Approval Process"

```text
[Customer] → Places Order → [Order Placed]
                                    ↓
                            [Check Amount Policy]
                                    ↓
                        If > $10k → [Approval Required]
                                    ↓
                [Manager] → Approves → [Order Approved]
                                    ↓
                            [Order Fulfilled]
```

---

## Software Design Event Storming

**Purpose**: Distill technical solutions from business knowledge

**Activities**:

- Map domain concepts to software structures
- Identify aggregates, commands, and events
- Design bounded contexts and integration points

**Duration**: 2-4 hours

**Participants**: Development team, technical leads, architects

**Outcome**: Technical design aligned with domain language

### When to Use Software Design

- After Process Modeling, before implementation
- Designing system architecture
- Refactoring existing systems
- Planning technical migrations

### Software Design Example Output

After a Software Design session, you'll have:

- **Aggregate boundaries** (which entities belong together)
- **Command handlers** (who processes each command)
- **Event publishers** (when and where events are emitted)
- **Read models** (what data views are needed)
- **Integration points** (where bounded contexts communicate)

**Example**: "Order Aggregate"

```text
Order Aggregate
├── Commands:
│   ├── PlaceOrder
│   ├── ApproveOrder
│   └── FulfillOrder
├── Events:
│   ├── OrderPlaced
│   ├── OrderApproved
│   └── OrderFulfilled
└── Business Rules:
    └── Large orders require approval
```

---

## Choosing the Right Format

### Start with Big Picture When

- Team is new to the domain
- Organization has multiple departments/teams
- You need to understand the whole system
- You're identifying bounded contexts

**Output leads to**: Process Modeling sessions for specific processes

### Use Process Modeling When

- You have a specific feature to implement
- You need detailed understanding of one process
- Big Picture revealed an area needing deeper exploration
- You're preparing for Example Mapping

**Output leads to**: Example Mapping workshops for that feature

### Use Software Design When

- You have Process Modeling output
- You're ready to design technical solution
- You need to refactor existing code
- You're defining system architecture

**Output leads to**: Implementation and coding

---

## Progressive Refinement

Event Storming formats build on each other:

```text
Big Picture Event Storming
    ↓
Identifies bounded contexts and major processes
    ↓
Process Modeling Event Storming
    ↓
Details specific business processes
    ↓
Example Mapping
    ↓
Creates concrete examples and rules
    ↓
Software Design Event Storming
    ↓
Designs technical implementation
    ↓
Gherkin Specifications
    ↓
Executable tests
```

You don't always need all formats. Choose based on your needs:

- **New project**: Big Picture → Process Modeling → Software Design
- **Specific feature**: Process Modeling → Example Mapping
- **Refactoring**: Software Design
- **Bug fixing**: Process Modeling (focused on problem area)

---

## Sticky Note Grammar

Event Storming uses color-coded sticky notes as a visual language:

| Color      | Represents       | Format                                 | Example                                  |
| ---------- | ---------------- | -------------------------------------- | ---------------------------------------- |
| **Orange** | Domain Event     | Past tense, something that happened    | "Order Placed", "Payment Received"       |
| **Blue**   | Command/Action   | Imperative, action that triggers event | "Place Order", "Approve Request"         |
| **Yellow** | Actor/Persona    | Who initiates commands                 | "Customer", "Manager", "System"          |
| **Pink**   | Policy/Rule      | Business rule that triggers commands   | "When order >$10k → require approval"    |
| **Purple** | External System  | Systems outside your control           | "Payment Gateway", "Email Service"       |
| **Red**    | Problem/Question | Uncertainty or blocker                 | "What if...?", "How does...?"            |
| **Green**  | Opportunity      | Improvement ideas                      | "Could automate...", "Might simplify..." |
| **White**  | Definition       | Glossary term clarification            | "Approved = manager signed off"          |

---

## Related Documentation

- [Event Storming Overview](./event-storming-overview.md) - What and why
- [Event Storming Facilitation](./event-storming-facilitation.md) - Running workshops
- [Example Mapping](./example-mapping.md) - Next step after Process Modeling

{{ diataxis_footer() }}
