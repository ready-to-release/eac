# Running Example Mapping Sessions

How to facilitate Example Mapping workshops for requirements discovery.

## Session Setup

### Time

**25 minutes** - strictly time-boxed

### Participants

- Product Owner (business perspective)
- Developer (technical perspective)
- Tester (quality perspective)

### Materials

- Yellow index cards (user stories)
- Blue index cards (rules/acceptance criteria)
- Green index cards (examples/scenarios)
- Red/Pink index cards (questions/unknowns)
- Timer

## Session Steps

### Step 1: Present the Story (2 minutes)

Write the user story on a **Yellow card**:

```text
As a [role]
I want to [capability]
So that [business value]
```

Place at top of workspace.

### Step 2: Discover Rules (8 minutes)

Ask: "What acceptance criteria must be satisfied?"

Write each on a **Blue card**:

- "Must validate email format"
- "Must send verification email"
- "Must prevent duplicate accounts"

Place Blue cards in a row below Yellow card.

### Step 3: Add Examples (12 minutes)

For each Blue card, ask: "Can you give me an example?"

Write each on a **Green card**:

- "User enters valid email, receives confirmation"
- "User enters invalid format, sees error"
- "User enters existing email, directed to login"

Place Green cards below their Blue card.

### Step 4: Capture Questions (ongoing)

When you encounter unknowns, write on **Red/Pink card**:

- "What happens if email service is down?"
- "How long is verification link valid?"

Place to the side. Red cards must be resolved before implementation.

### Step 5: Assess Readiness (3 minutes)

**Ready to implement if**:

- 1 Yellow card (story)
- ≤4 Blue cards (acceptance criteria)
- ≤4 Green cards per Blue card
- No blocking Red cards

**Not ready if**:

- Too many Red cards → stop and research
- Too many Blue cards → story too big, split it

## Output: Gherkin Translation

### Cards → Gherkin

| Card Color | Gherkin Element     |
| ---------- | ------------------- |
| Yellow     | Feature description |
| Blue       | `Rule:` block       |
| Green      | `Scenario:` block   |
| Red        | Issues to resolve   |

### Example Translation

**Blue card**: "Must validate email format"

**Green cards**:

- "Valid email succeeds"
- "Invalid email shows error"

```gherkin
Rule: Must validate email format

  @ov
  Scenario: Valid email succeeds
    Given I am registering
    When I enter email "user@example.com"
    Then registration should proceed

  @ov
  Scenario: Invalid email shows error
    Given I am registering
    When I enter email "not-an-email"
    Then I should see "Invalid email format"
```

## Session Signals

### Stop and Split

- More than 4 Blue cards
- Green cards are hard to think of
- Lots of disagreement

### Stop and Research

- Many Red cards accumulating
- Fundamental questions unresolved
- Scope unclear

### Ready to Go

- Clear acceptance criteria
- Concrete examples
- Questions resolved

## Virtual Sessions

For remote teams:

- Use Miro, Mural, or FigJam
- Color-coded sticky notes
- Screen share for all participants
- Timebox strictly (use visible timer)

## Related

- [Translating to Gherkin](translating-to-gherkin.md)
- [Creating Feature Files](creating-feature-files.md)
- [Example Mapping](../../explanation/specifications/example-mapping.md)
