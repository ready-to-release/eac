# Event Storming Facilitation

> **Running effective Event Storming workshops**

Learn how to facilitate successful Event Storming sessions.

---

## When to Use

- **New domains**: When starting a new project or module
- **Complex workflows**: Understanding intricate business processes
- **Team alignment**: Building shared understanding across stakeholders

---

## Materials Needed

### Required

- Brown packing paper (6-8 meters long)
- Sticky notes in multiple colors:
  - Orange (domain events)
  - Blue (commands/actions)
  - Yellow (actors/personas)
  - Pink (policies/business rules)
  - Purple (external systems)
  - Red (problems/questions)
  - Green (opportunities)
  - White (definitions)
- Markers (black, fine tip)
- Wall space (long, uninterrupted)

### Optional

- Timer (for time-boxing discussions)
- Camera (for documentation)
- Whiteboard (for capturing definitions)

---

## Workshop Structure

### Opening (10-15 minutes)

1. Explain the format and rules
2. Clarify the scope (Big Picture vs Process Modeling)
3. Set ground rules (no laptops, stay engaged, park questions)

### Discovery Phase (varies by format)

1. **Silent storm**: Everyone writes domain events individually
2. **Timeline**: Arrange events on paper in rough chronological order
3. **Clustering**: Group related events
4. **Fill gaps**: Add commands, actors, policies

### Discussion Phase

1. Walk through the timeline together
2. Ask clarifying questions
3. Surface definitions and terminology
4. Identify conflicts and assumptions

### Closing (15-20 minutes)

1. Highlight key insights
2. Identify open questions
3. Capture glossary of terms
4. Define next steps

---

## Time Boxing

### Big Picture: 4-8 hours

Can split across multiple sessions:

- **Session 1**: Silent storm + timeline (2-4 hours)
- **Session 2**: Discussion + refinement (2-4 hours)

### Process Modeling: 2-4 hours per process

- Focus on one process at a time
- Take breaks every 90 minutes

### Software Design: 2-4 hours

- Requires prior Process Modeling output
- More focused technical discussion

---

## Facilitation Tips

### Start Broad, Then Focus

- Begin with high-level events
- Drill down into details progressively
- Don't get stuck on edge cases early

### Encourage Questions

- Red sticky for every "What if...?"
- Don't try to answer all questions immediately
- Some questions surface later understanding

### Maintain Energy

- Stand, don't sit
- Take regular breaks
- Keep discussions moving
- Park off-topic conversations

### Capture Everything

- Take photos of the final timeline
- Document definitions on whiteboard
- Record open questions
- Note follow-up actions

### Stay Visual

- Sticky notes only—no laptops during discovery
- Use the paper as the source of truth
- Arrange spatially to show relationships

---

## Common Pitfalls

### Going Too Detailed Too Early

**Problem**: Team gets stuck on edge cases

**Solution**: "Park it" on a red sticky, move forward

**Why it happens**: Developers want to understand everything completely

**Prevention**: Remind team this is discovery, not implementation planning

### Skipping Domain Experts

**Problem**: Developers make assumptions

**Solution**: Require domain expert participation

**Why it happens**: Domain experts are "too busy"

**Prevention**: Position Event Storming as essential, not optional

### Focusing on Implementation

**Problem**: Discussion drifts to "how we'll build it"

**Solution**: Stay focused on "what happens in the business"

**Why it happens**: Developers naturally think about implementation

**Prevention**: Facilitate with business-focused questions

### Not Capturing Definitions

**Problem**: Team uses terms differently

**Solution**: Use white stickies for definitions immediately

**Why it happens**: Assumed shared understanding

**Prevention**: Ask "What exactly do we mean by...?" frequently

### Insufficient Wall Space

**Problem**: Timeline feels cramped

**Solution**: Book a room with long walls, use multiple sheets

**Why it happens**: Underestimating space needed

**Prevention**: 6-8 meters minimum for Big Picture

---

## Detailed Session Steps

### Step 1: Set the Stage (5 minutes)

Explain the goal:

> "We're going to discover what happens in our system by telling its story through events."

Explain the rules:

- Events are past tense ("Order Placed", not "Place Order")
- No wrong answers
- Questions go on pink/red notes
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
- Mark hotspots with pink/red notes

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
- Write on purple/pink notes
- Connect to triggering event

```text
[Orange: Payment Failed] → [Purple: Retry 3 times] → [Blue: Cancel Order]
```

---

## From Events to Specifications

After Event Storming:

1. **Identify user stories** from command-event pairs
2. **Run Example Mapping** on each story
3. **Create Gherkin specifications** from example maps

```text
Event Storming → Example Mapping → Gherkin
```

This creates traceability from domain discovery to executable tests.

---

## Follow-Up Actions

### Immediately After

1. Photograph the entire timeline
2. Transcribe glossary terms
3. Identify first features to specify
4. Schedule Example Mapping sessions

### Within a Week

1. Digitize the timeline (optional—photos often sufficient)
2. Create/update bounded context documentation
3. Share photos and glossary with stakeholders
4. Begin Example Mapping for first features

---

## Ongoing Event Storming

Event Storming is not a one-time workshop. The domain evolves, and your understanding must evolve with it.

### When to Re-run Event Storming

**Scheduled (Quarterly)**:

- Revisit domain model with full team
- Validate current understanding
- Identify evolution in business processes
- Discover new domain concepts

**Triggered (As needed)**:

- Major feature additions
- Business process changes
- New product lines
- Team onboarding (Big Picture format)
- Architecture decisions (Software Design format)

---

## Propagating Event Storming Insights

When Event Storming reveals new understanding, update affected specifications:

### 1. Identify Impacted Features

```bash
# Find specifications related to discovered domain events
grep -r "OrderPlaced\|OrderShipped" specs/
```

### 2. Review Existing Scenarios

- Do they reflect the current domain model?
- Are there new events/commands to test?
- Has the process flow changed?

### 3. Update Specifications

- Add scenarios for new domain events
- Refine language to match Event Storming terminology
- Update Rules to reflect evolved acceptance criteria
- Remove scenarios for deprecated processes

### 4. Refactor Implementation

- Update step definitions
- Refactor code to match new domain model
- Run tests to verify changes

---

## Facilitation Checklist

Use this checklist when preparing to facilitate:

**Before the Workshop**:

- [ ] Book room with long wall space
- [ ] Purchase brown paper (6-8 meters)
- [ ] Get sticky notes in all colors
- [ ] Get markers (one per participant)
- [ ] Identify and invite domain experts
- [ ] Set clear scope and objectives
- [ ] Block calendar (no interruptions)

**During the Workshop**:

- [ ] Explain format and rules (10-15 min)
- [ ] Silent storm phase
- [ ] Timeline creation
- [ ] Discussion and refinement
- [ ] Capture definitions (white stickies)
- [ ] Take photos throughout
- [ ] Park questions (red stickies)
- [ ] Maintain energy and pace

**After the Workshop**:

- [ ] Photograph final timeline
- [ ] Transcribe glossary
- [ ] Document open questions
- [ ] Share results with team
- [ ] Schedule Example Mapping sessions
- [ ] Follow up on action items

---

## Example Facilitation Script

### Opening

> "Welcome! Today we're using Event Storming to understand [scope]. This is a collaborative workshop—everyone's input matters. Here are the rules:
>
> 1. **Orange stickies** = things that happened (past tense)
> 2. **Stay visual** = sticky notes only, no laptops
> 3. **Park questions** = red stickies for 'what if' questions
> 4. **Capture definitions** = white stickies when we clarify terms
>
> We'll start with silent storming—everyone writes events independently for 15 minutes. Ready?"

### During Silent Storm

> "Write every significant event you can think of. Use past tense: 'Order Placed', not 'Place Order'. Don't worry about order yet—we'll organize them next."

### Timeline Creation

> "Let's arrange these events in rough chronological order. Who wants to start? What happens first in this process?"

### Discussion

> "Walk me through this section. What triggers this event? Who's involved? What information do we need?"

### Capturing Definitions

> "I'm hearing 'approved' used in different ways. Let's clarify—what exactly do we mean by 'approved'? [Write on white sticky]"

### Closing

> "Great work! We've discovered [key insights]. Red stickies show questions we need to answer. Next steps: [Example Mapping for feature X]. Thanks everyone!"

---

## Related Documentation

- [Event Storming Overview](./event-storming-overview.md) - What and why
- [Event Storming Formats](./event-storming-formats.md) - Big Picture, Process, Design
- [Example Mapping](./example-mapping.md) - Next step after Event Storming
- [Ubiquitous Language](../concepts/ubiquitous-language.md) - Using discovered vocabulary
