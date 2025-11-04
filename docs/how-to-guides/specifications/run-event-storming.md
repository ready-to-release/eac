# Run Event Storming Workshop

Facilitate an Event Storming workshop to discover domain language and business processes collaboratively.

---

## What is Event Storming?

Event Storming is a collaborative workshop technique that uses sticky notes on a long paper timeline to discover domain vocabulary and business processes. It's the foundation for establishing Ubiquitous Language.

**Goal**: Discover domain events, actors, commands, policies, and most importantly—the shared vocabulary that will flow into your specifications.

**See**: [Event Storming Explanation](../../explanation/specifications/event-storming.md) for conceptual background.

---

## Choose Your Format

Event Storming has three formats. Choose based on your goal:

| Format | Goal | Duration | When to Use |
|--------|------|----------|-------------|
| **Big Picture** | Align stakeholders, discover domain | 4-8 hours | Starting new projects, onboarding, finding boundaries |
| **Process Modeling** | Detail specific processes | 2-4 hours | Before implementing features, refining processes |
| **Software Design** | Technical design from domain | 2-4 hours | After Process Modeling, before coding |

**This guide focuses on Process Modeling** (most common for specification work).

For Big Picture facilitation, see: [Event Storming Resources](https://www.eventstorming.com)

---

## Prerequisites

### Materials

**Required**:

- **Brown packing paper** (6-8 meters long, ideally on a roll)
- **Sticky notes** in specific colors:
  - Orange (domain events) - 100+
  - Blue (commands) - 50+
  - Yellow (actors) - 30+
  - Pink (policies/rules) - 30+
  - Purple (external systems) - 20+
  - Red (problems/questions) - 30+
  - White (definitions) - 30+
- **Markers** (black, fine tip, one per person)
- **Masking tape** (to attach paper to wall)

**Optional**:

- Camera (for documentation)
- Timer (for time-boxing)
- Whiteboard nearby (for glossary)

### Space

- **Long wall** (6-8 meters of uninterrupted wall space)
- **Standing height** (paper at eye level when standing)
- **Room for movement** (people need to walk along the timeline)
- **No tables** (standing workshop—keeps energy high)

### Participants (Required)

**For Process Modeling**:

- **Domain expert(s)** - 1-2 people who know the business process deeply
- **Developer(s)** - 2-3 developers who will implement
- **Facilitator** - 1 person who guides (can be a developer)
- **QA/Tester** - 1 person who will test

**Total**: 4-7 people (more = slower)

### Time

- **Duration**: 2-4 hours (take breaks every 90 minutes)
- **Preparation**: 30 minutes before
- **Follow-up**: 30-60 minutes after

---

## Before the Workshop

### Step 1: Identify the Process Scope

Clearly define what process you're exploring.

**Example**: "Order approval process from customer order to fulfillment"

**Be specific**:

- ✅ "Order approval for orders >$10k"
- ❌ "Everything about orders" (too broad)

### Step 2: Set Up the Space

**30 minutes before**:

1. Attach paper to wall at standing height
2. Use masking tape every meter to secure
3. Ensure paper is smooth and writable
4. Set up sticky note stations along the wall
5. Give each person a marker

### Step 3: Brief Participants (10 minutes)

Explain before starting:

**The Format**:

- We'll use sticky notes to map events
- Orange = things that happen ("Order Placed")
- Blue = actions that trigger events ("Place Order")
- Yellow = who does it ("Customer")
- Stay standing, stay engaged

**The Rules**:

- Write one thing per sticky note
- Use past tense for events ("Order Placed" not "Place Order")
- Keep it brief (2-4 words max)
- No laptops or phones during discovery
- Questions go on red stickies—we'll address them later

**The Goal**:

- Map the business process as it actually works
- Surface the vocabulary the business uses
- Identify questions and unknowns

---

## During the Workshop

### Part 1: Silent Storm (15-20 minutes)

**Who leads**: Everyone works independently

**Task**: Each person writes domain events individually on orange sticky notes

**Instructions**:

"For the next 15 minutes, write domain events—things that happen in this process. Use past tense. One event per sticky note. Write as many as you can think of."

**Example events**:

- "Order Placed"
- "Payment Received"
- "Manager Notified"
- "Approval Granted"
- "Order Fulfilled"

**Facilitator tips**:

- Circulate and help people who are stuck
- Remind: "What happens in this process?"
- Encourage: "Don't worry about order yet, just capture events"
- No discussion yet—silent brainstorming

**Common mistakes**:

- ❌ Writing actions: "Place Order" → ✅ "Order Placed"
- ❌ Technical details: "Database updated" → ✅ "Record Saved"
- ❌ Multiple events per sticky → One per sticky

---

### Part 2: Timeline (30-45 minutes)

**Who leads**: Facilitator guides, everyone participates

**Task**: Arrange events on the timeline in chronological order

**Instructions**:

"Let's build a timeline. One person at a time, place an orange sticky on the paper and briefly explain it. We'll arrange them in rough time order—left to right."

**Process**:

1. **First person** places an orange sticky and explains (30 seconds)
2. **Group discussion** (if needed): "Is that event named correctly?"
3. **Next person** places their sticky relative to existing ones
4. **Repeat** until all stickies are placed

**Facilitator techniques**:

**Clustering**: If multiple similar events, stack them vertically

```text
        "Order Submitted" ─────→ "Approval Granted"
        "Order Received"          "Approval Denied"
        "Order Confirmed"
```

**Parallel flows**: Use vertical space for things that happen at same time

```text
"Order Placed" ─→ "Inventory Checked"
                  "Payment Processed"  ─→ "Order Confirmed"
                  "Email Sent"
```

**Gaps**: Leave space where something is missing

```text
"Order Placed" ─────────[gap here]─────→ "Order Fulfilled"
```

**Facilitator phrases**:

- "What happens between X and Y?"
- "Who makes that decision?"
- "What would trigger this event?"
- "Is there another term the business uses for this?"

**When conflicts arise**:

- Domain expert says: "We call it 'Approved'"
- Developer says: "We call it 'Validated'"
- **Solution**: Write on WHITE sticky: "Approved ≠ Validated. Approved = business accepted, Validated = data checked"

---

### Part 3: Add Commands (20-30 minutes)

**Who leads**: Facilitator, with group participation

**Task**: Add blue stickies for commands (actions that trigger events)

**Instructions**:

"Now let's add what causes these events. Commands are actions—use imperative form."

**Format**: Place BLUE sticky to the left of the ORANGE event it triggers

```text
[Blue: "Place Order"] → [Orange: "Order Placed"]
[Blue: "Approve"] → [Orange: "Order Approved"]
```

**Example**:

```text
Customer          Manager
   ↓                ↓
[Place Order] → [Order Placed] → [Check Amount] → [Approval Required] → [Approve Order] → [Approval Granted]
```

**Key questions**:

- "What action causes this event?"
- "Who or what initiates this command?"

---

### Part 4: Add Actors (15-20 minutes)

**Who leads**: Facilitator

**Task**: Add yellow stickies for who executes commands

**Instructions**:

"Who executes each command? Add a yellow sticky above the command."

**Format**: Place YELLOW sticky above BLUE command

```text
[Yellow: "Customer"]
        ↓
[Blue: "Place Order"] → [Orange: "Order Placed"]
```

**Full example**:

```text
[Customer]              [System]           [Manager]
    ↓                      ↓                   ↓
[Place Order] → [Order Placed] → [Check Amount] → [Order Approved]
```

**Key actors**:

- People: "Customer", "Manager", "Operator"
- Systems: "Payment Gateway", "Inventory System"
- Roles: "Administrator", "Support Agent"

---

### Part 5: Add Policies (15-20 minutes)

**Who leads**: Facilitator, domain expert clarifies

**Task**: Add pink stickies for business rules

**Instructions**:

"What rules govern this process? Policies are 'whenever X, then Y' statements."

**Format**: Place PINK sticky near the event/command it affects

**Example**:

```text
                [Pink: "Whenever order > $10,000 → require manager approval"]
                                    ↓
[Order Placed] ──────────────→ [Approval Required]
```

**Policy format**:

- "Whenever [condition], then [action]"
- "If [condition], require [action]"

**Examples**:

- "Whenever order > $10k → require approval"
- "If payment fails 3 times → lock account"
- "When inventory < 10 → reorder"

---

### Part 6: Capture Questions (Ongoing)

**Who adds**: Anyone, anytime during the workshop

**Task**: Write questions on RED stickies

**Instructions**:

"Whenever you have a question or uncertainty, write it on a red sticky and place it near the relevant event. We'll address them at the end."

**Example questions**:

- "What if order is exactly $10,000?"
- "Can customer cancel after approval?"
- "How long until approval timeout?"

**Facilitator action**:

- Don't stop to answer questions immediately
- Acknowledge: "Good question—red sticky it"
- At end: Group reviews red stickies

---

### Part 7: Capture Definitions (Ongoing)

**Who adds**: Facilitator, as conflicts arise

**Task**: Write definitions on WHITE stickies

**Instructions**:

Whenever terminology confusion arises, capture the clarification.

**Format**:

```text
[White sticky]
"Approved = Manager signed off (business acceptance)
 Validated = Data format checked (technical verification)"
```

**Place near** the relevant events

**Also maintain** a glossary on whiteboard:

```text
GLOSSARY:
- Order: Customer request to purchase
- Approved: Business acceptance by manager
- Fulfilled: Items shipped to customer
```

---

## After the Workshop

### Step 1: Photograph the Timeline (5 minutes)

Take multiple photos:

1. **Wide shot** - Entire timeline
2. **Sections** - Each major segment in detail
3. **Definitions** - All white stickies clearly visible
4. **Glossary** - Whiteboard with terms

Label photos with date and process name.

### Step 2: Transcribe Glossary (15-20 minutes)

Create a document with domain vocabulary discovered:

```markdown
# Domain Glossary: Order Management

## Events
- **Order Placed**: Customer submitted purchase request
- **Order Approved**: Manager granted business acceptance
- **Order Fulfilled**: Items shipped to customer

## Actors
- **Customer**: Person placing order
- **Manager**: Employee with approval authority

## Commands
- **Place Order**: Customer initiates purchase
- **Approve Order**: Manager grants approval

## Policies
- Orders > $10,000 require manager approval
- Failed payments lock account after 3 attempts
```

**Save as**: `specs/<module>/glossary.md` or `specs/<module>/<feature>/glossary.md`

### Step 3: Identify First Features (10 minutes)

Review the timeline and identify first features to specify with Example Mapping.

**Look for**:

- Clear start/end boundaries
- Self-contained workflows
- Business value

**Example features** from order approval timeline:

- "High-value order approval" (Order >$10k → approval flow)
- "Standard order processing" (Order <$10k → direct fulfillment)

### Step 4: Schedule Example Mapping (5 minutes)

For each feature identified, schedule Example Mapping workshop:

**Timeline**:

- Event Storming: Today
- Example Mapping: Within 1 week
- Specification writing: Within 2 weeks

**Use glossary** from Event Storming during Example Mapping.

---

## Next Steps

### Immediate (Same Day)

- ✅ Photos taken and labeled
- ✅ Glossary transcribed
- ✅ First features identified
- **Next**: Schedule Example Mapping for first feature

### Within 1 Week

- Run Example Mapping workshops
- Use glossary vocabulary in cards
- Discover specific acceptance criteria

**See**: [Run Example Mapping Workshop](./run-example-mapping.md)

### Within 2 Weeks

- Write Gherkin specifications using discovered vocabulary
- Feature descriptions use event/actor names
- Scenarios use commands and events

**See**: [Create Feature Spec](./create-specifications.md)

---

## Workshop Checklist

**Before** (30 minutes):

- [ ] Paper attached to wall (6-8 meters)
- [ ] Sticky notes organized by color
- [ ] Markers distributed (one per person)
- [ ] Process scope clearly defined
- [ ] Participants briefed on format

**During** (2-4 hours):

- [ ] Silent storm: Everyone writes events (15-20 min)
- [ ] Timeline: Arrange events chronologically (30-45 min)
- [ ] Commands: Add what triggers events (20-30 min)
- [ ] Actors: Add who executes commands (15-20 min)
- [ ] Policies: Add business rules (15-20 min)
- [ ] Questions: Red stickies captured throughout
- [ ] Definitions: White stickies for terminology

**After** (30-60 minutes):

- [ ] Photos taken (wide, sections, definitions)
- [ ] Glossary transcribed to document
- [ ] First features identified
- [ ] Example Mapping workshops scheduled

---

## Related Guides

### For Understanding

- [Event Storming Explanation](../../explanation/specifications/event-storming.md) - Three formats, concepts, and benefits
- [Ubiquitous Language](../../explanation/specifications/ubiquitous-language.md) - DDD foundation for shared vocabulary
- [Three-Layer Approach](../../explanation/specifications/three-layer-approach.md) - How Event Storming fits the testing strategy

### For Doing

- [Run Example Mapping Workshop](./run-example-mapping.md) - Next step after Event Storming
- [Establish Ubiquitous Language](./establish-ubiquitous-language.md) - Complete flow from Event Storming to code
- [Create Feature Spec](./create-specifications.md) - Write Gherkin using discovered vocabulary

### External Resources

- [Event Storming Official Site](https://www.eventstorming.com) - Resources and examples
- [Introducing Event Storming](https://leanpub.com/introducing_eventstorming) - Alberto Brandolini's book
- [Awesome Event Storming](https://github.com/mariuszgil/awesome-eventstorming) - Curated resources

---

## Summary

Event Storming discovers domain vocabulary through collaborative visual mapping. The output—events, actors, commands, policies, and especially the glossary—flows directly into Example Mapping and Gherkin specifications.

**Key Success Factors**:

1. **Right participants** - Domain experts + developers + QA
2. **Adequate space** - Long wall with paper timeline
3. **Time management** - Stay focused, park deep questions
4. **Capture vocabulary** - White stickies and glossary are critical
5. **Follow through** - Schedule Example Mapping within a week

The shared language discovered today becomes the specifications written tomorrow.
