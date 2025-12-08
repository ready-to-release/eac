# Running Event Storming Sessions

How to facilitate Event Storming workshops for domain discovery.

## When to Use

- **New domains**: When starting a new project or module
- **Complex workflows**: Understanding intricate business processes
- **Team alignment**: Building shared understanding across stakeholders

## Session Setup

### Time

- **Big Picture**: 2-4 hours
- **Process Level**: 1-2 hours per bounded context

### Participants

- Domain experts (business stakeholders)
- Developers
- Product owners
- Facilitator

### Materials

- Large wall space or virtual whiteboard
- Orange sticky notes (events)
- Blue sticky notes (commands)
- Yellow sticky notes (actors)
- Pink sticky notes (hotspots/questions)
- Purple sticky notes (policies)
- Green sticky notes (read models)

## Session Steps

### Step 1: Set the Stage (5 minutes)

Explain the goal:

> "We're going to discover what happens in our system by telling its story through events."

Explain the rules:

- Events are past tense ("Order Placed", not "Place Order")
- No wrong answers
- Questions go on pink notes
- Everyone participates

### Step 2: Chaotic Exploration (20-30 minutes)

**Everyone writes events simultaneously**:

- Write one event per orange sticky
- Use past tense: "Order Placed", "Payment Failed", "User Registered"
- Place on wall in rough time order (left to right)
- Don't worry about duplicates or order yet

**Facilitator**:

- Keep energy high
- Encourage participation
- Don't correct or organize yet

### Step 3: Enforce Timeline (15-20 minutes)

**Organize events chronologically**:

- Move events left (earlier) or right (later)
- Identify pivotal events (major state changes)
- Find gaps: "What happens between X and Y?"
- Mark hotspots with pink notes

### Step 4: Add Commands and Actors (20-30 minutes)

**For each event, ask**:

- "What triggers this event?" → Blue command note
- "Who triggers it?" → Yellow actor note

```text
[Yellow: Customer] → [Blue: Place Order] → [Orange: Order Placed]
```

### Step 5: Identify Aggregates and Bounded Contexts (20 minutes)

**Group related events**:

- Draw boundaries around clusters
- Name the bounded contexts
- Identify aggregates (key domain objects)

### Step 6: Capture Policies (10 minutes)

**Identify automatic reactions**:

- "When X happens, then Y must happen"
- Write on purple notes
- Connect to triggering event

```text
[Orange: Payment Failed] → [Purple: Retry 3 times] → [Blue: Cancel Order]
```

## Sticky Note Colors

| Color  | Represents       | Example                     |
| ------ | ---------------- | --------------------------- |
| Orange | Domain Event     | "Order Placed"              |
| Blue   | Command          | "Place Order"               |
| Yellow | Actor            | "Customer"                  |
| Pink   | Hotspot/Question | "What if payment fails?"    |
| Purple | Policy           | "Retry 3 times then cancel" |
| Green  | Read Model       | "Order Summary View"        |

## Common Pitfalls

**Avoid**:

- Starting with commands (start with events)
- Using present tense ("Order is placed")
- Technical jargon only developers understand
- One person dominating
- Going too deep too fast

**Do**:

- Start with domain events
- Use past tense consistently
- Use ubiquitous language
- Encourage everyone to participate
- Mark unknowns with pink notes

## From Events to Specifications

After Event Storming:

1. **Identify user stories** from command-event pairs
2. **Run Example Mapping** on each story
3. **Create Gherkin specifications** from example maps

```text
Event Storming → Example Mapping → Gherkin
```

## Virtual Sessions

For remote teams:

- Miro or Mural with color-coded stickies
- Template with zones and legend
- Breakout rooms for parallel discovery
- Clear timer visible to all

## Related

- [Running Example Mapping](running-example-mapping.md)
- [Event Storming Concepts](../../../explanation/specifications/event-storming.md)
- [Ubiquitous Language](../../../explanation/specifications/ubiquitous-language.md)
