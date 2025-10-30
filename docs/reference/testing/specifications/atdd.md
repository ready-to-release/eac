# ATDD: Acceptance Test-Driven Development

**[<- Back to Testing Overview](./index.md)**

## What is ATDD?

**Acceptance Test-Driven Development (ATDD)** is a collaborative approach where business stakeholders, developers, and testers define acceptance criteria **before** development begins. It focuses on capturing business value and measurable success criteria from the customer's perspective.

### Key Characteristics

| Aspect | Description |
|--------|-------------|
| **Who** | Product owner, business stakeholders, developers, testers |
| **When** | Before feature work begins |
| **Format** | User story ("As a/I want/So that") + acceptance criteria checklist |
| **Location** | Feature files in `requirements/<module>/` |
| **Purpose** | Define business value and measurable success criteria |

## ATDD in Feature Files

ATDD content appears at the **top of each `.feature` file**, providing business context for the technical specifications that follow.

### Template Structure

```gherkin
@cli @critical
Feature: [Feature Name]

  # ATDD Layer: Business context and value
  As a [user role]
  I want [capability]
  So that [business value]

  Acceptance Criteria:
  - [ ] [Measurable criterion 1]
  - [ ] [Measurable criterion 2]
  - [ ] [Measurable criterion 3]
  - [ ] [Measurable criterion 4]
```

### Component Breakdown

#### 1. User Story (3 Lines)

**Format**: As a / I want / So that

**Purpose**: Captures WHO needs WHAT and WHY

**Example**:

```gherkin
As a developer
I want to initialize a CLI project with a single command
So that I can quickly start development with proper structure
```

**Guidelines**:

- **As a [role]**: Specify the user persona or stakeholder
- **I want [capability]**: Describe the desired functionality
- **So that [value]**: Explain the business benefit or outcome

#### 2. Acceptance Criteria (2-6 Items)

**Format**: Checkbox list with measurable outcomes

**Purpose**: Defines what "done" means from a business perspective

**Example**:

```gherkin
Acceptance Criteria:
- [ ] Creates project directory structure
- [ ] Generates valid configuration file
- [ ] Command completes in under 2 seconds
- [ ] Works on Linux, macOS, and Windows
- [ ] Exits with clear success/error messages
- [ ] Handles existing projects gracefully
```

## Workflow

### ATDD Development Process (with Example Mapping)

```text
ATDD Workflow

1. Stakeholder Meeting
   +- Gather business requirements
   +- Identify user personas
   +- Understand problem domain

2. Define Business Value
   +- Write "As a/I want/So that" user story (YELLOW CARD)
   +- Focus on WHY (business value)
   +- Get stakeholder agreement

3. Example Mapping Workshop (15-25 min)
   +- Gather: Product Owner, Developer, Tester
   +- Place Yellow Card (user story) at top
   +- Discover Blue Cards (rules/acceptance criteria)
   +- Create Green Cards (concrete examples)
   +- Capture Red Cards (questions/blockers)
   +- Assess: Ready, Too Large, or Too Uncertain?

4. Document Rules as Acceptance Criteria
   +- Blue Cards -> Acceptance Criteria in feature file
   +- Make each criterion testable (pass/fail)
   +- Include functional AND non-functional requirements
   +- Keep list focused (2-6 items)

5. Resolve Questions
   +- Address Red Cards from workshop
   +- Research or defer as needed
   +- Create follow-up stories if necessary

6. Review with Stakeholders
   +- Validate acceptance criteria
   +- Ensure clarity and completeness
   +- Confirm definition of "done"

7. Create Feature File (ATDD Layer)
   +- Choose appropriate module folder
   +- Use descriptive feature file name
   +- Add ATDD layer with user story and criteria
   +- Apply appropriate tags (@critical, etc.)

8. Proceed to BDD
   +- Use Green Cards to write scenarios
   +- See: BDD Workflow ->
```

### Example Mapping Workshop (Collaborative Discovery)

Example Mapping is a collaborative technique to explore user stories before writing code. It helps teams discover acceptance criteria and concrete examples through structured conversation.

#### Workshop Structure

Conduct a **15-25 minute timeboxed session** with:

- **Product Owner** (story author, business perspective)
- **Developer(s)** (technical perspective, implementer)
- **Tester(s)** (quality perspective, edge cases)

#### The Four Card Types

**YELLOW CARD - User Story** (1 per session):

- Place at top of workspace (physical or digital)
- The "As a/I want/So that" user story
- Represents the single feature being explored

**BLUE CARDS - Rules** (Acceptance Criteria):

- Each blue card is one business rule or constraint
- Place horizontally under the yellow card
- Maps directly to ATDD acceptance criteria
- Examples: "Must work offline", "Response < 2 seconds", "Creates directory structure"
- Aim for 2-6 blue cards (if more, story may be too large)

**GREEN CARDS - Examples** (Concrete Scenarios):

- Illustrate how rules work in practice
- Place vertically under the relevant blue card
- Maps directly to BDD scenarios (Given/When/Then)
- Include both happy paths AND edge cases
- Typical: 2-4 examples per rule

**RED CARDS - Questions** (Uncertainties):

- Capture unknowns, dependencies, or blockers
- Place to the side of the main mapping
- Must be resolved before implementation
- May spawn new stories if too complex

#### Workshop Process

**1. Setup** (2 min):

- Write user story on yellow card
- Place at top of workspace

**2. Discover Rules** (5-10 min):

- Ask: "What are the acceptance criteria for this story?"
- Ask: "What rules constrain this feature?"
- Write each rule on a blue card
- Place horizontally under user story

**3. Create Examples** (10-15 min):

- For each blue card, ask: "Can you give me an example?"
- Write concrete examples on green cards
- Place under the relevant blue card
- Include happy path AND edge cases
- Ask: "What could go wrong?" to find error cases

**4. Capture Questions** (ongoing):

- When uncertainty arises, write on red card
- Place to the side
- Decide: resolve now, research later, or split story?

**5. Assess Readiness**:

- **Ready**: < 6 blue cards, clear examples, few/no red cards -> Proceed
- **Too Large**: > 6 blue cards -> Split the story
- **Too Uncertain**: Many red cards -> Need more research

**6. Output**:

- Blue cards -> ATDD Acceptance Criteria
- Green cards -> BDD Scenarios (next phase)
- Red cards -> Follow-up actions or new stories

#### Visual Layout

```text
+---------------------------------------------------------------+
|  YELLOW CARD: User Story                                      |
|  As a [role], I want [capability], So that [value]            |
+---------------------------------------------------------------+
        |
        +-- BLUE CARD: Rule 1 (e.g., Creates directory structure)
        |       |
        |       +-- GREEN: Example 1.1 (happy path: empty folder -> success)
        |       +-- GREEN: Example 1.2 (edge case: existing project -> error)
        |
        +-- BLUE CARD: Rule 2 (e.g., Generates config file)
        |       |
        |       +-- GREEN: Example 2.1 (new project -> default config)
        |       +-- GREEN: Example 2.2 (custom name -> custom config)
        |
        +-- BLUE CARD: Rule 3 (e.g., Works cross-platform)
                |
                +-- GREEN: Example 3.1 (Linux -> Unix paths)
                +-- GREEN: Example 3.2 (Windows -> Windows paths)

SIDE (Questions):
    RED CARD: What if config file already exists?
    RED CARD: Should we support --force flag?
```

#### Example Mapping Template

**Before Workshop**:

```text
YELLOW CARD:
As a [user role]
I want [capability]
So that [business value]
```

**During Workshop**:

```text
YELLOW: As a developer, I want to initialize a CLI project with one command,
        so that I can quickly start development

BLUE: Creates directory structure
  GREEN: Empty folder -> run init -> creates src/, tests/, docs/
  GREEN: Existing project -> run init -> error "already initialized"

BLUE: Generates valid configuration file
  GREEN: New project -> init -> creates cc.yaml with defaults
  GREEN: With --name flag -> cc.yaml contains custom name

BLUE: Works cross-platform
  GREEN: Linux -> creates Unix-style paths
  GREEN: Windows -> creates Windows-style paths

RED: What happens if cc.yaml already exists?
RED: Should we support --force flag to overwrite?
```

**After Workshop** - Maps to Feature File:

```gherkin
@cli @critical
Feature: Initialize a new project

  # ATDD Layer (from Yellow + Blue Cards)
  As a developer
  I want to initialize a CLI project with one command
  So that I can quickly start development

  Acceptance Criteria:
  - [ ] Creates directory structure (src/, tests/, docs/)
  - [ ] Generates valid configuration file
  - [ ] Works on Linux, macOS, and Windows

  # BDD Layer (from Green Cards - see BDD guide)
  @success
  Scenario: Initialize in empty folder
    Given I am in an empty folder
    When I run "cc init"
    Then directories "src/", "tests/", "docs/" should exist

  @error
  Scenario: Initialize in existing project
    Given I am in a directory with a "cc.yaml" file
    When I run "cc init"
    Then the command should fail
    And stderr should contain "Project already initialized"

  # ... additional scenarios from other green cards
```

**Red Card Resolution**:

```text
RED CARD: What if cc.yaml already exists?
  -> RESOLVED: Added GREEN CARD (error scenario above)

RED CARD: Should we support --force flag?
  -> DEFERRED: Created new story "Add --force flag to init command"
```

#### Tools for Example Mapping

**Physical** (in-person workshops):

- Index cards in 4 colors (yellow, blue, green, red)
- Whiteboard or wall space
- Markers

**Digital** (remote/hybrid workshops):

- Miro (has Example Mapping template)
- Mural (has Example Mapping template)
- FigJam (Figma whiteboard)
- Google Jamboard
- Plain markdown/text files (for async)

#### Benefits

- **Time-boxed**: 15-25 minutes keeps meetings focused
- **Collaborative**: Brings together different perspectives
- **Visual**: Everyone sees the same structure
- **Right-sizing**: Reveals if story is too large (>6 blues = split)
- **Early questions**: Surfaces blockers before coding
- **Shared understanding**: Team aligns on scope and examples

### Prerequisites

Before starting ATDD:

- Product vision or feature request exists
- Key stakeholders are identified
- Business problem is understood

### Outputs

After completing ATDD:

- `.feature` file created in `requirements/<module>/`
- User story captures WHO/WHAT/WHY
- 2-6 measurable acceptance criteria defined
- Stakeholders have approved criteria
- Clear definition of "done" established

## Style Rules

### Do

- **Use checkboxes** for trackable criteria: `- [ ]`
- **Focus on business value** and customer needs
- **Write from stakeholder perspective** (not developer)
- **Make criteria measurable** (can determine pass/fail)
- **Include non-functional requirements** (performance, compatibility, usability)
- **Keep criteria specific** and unambiguous
- **Limit to 2-6 criteria** (keep focused)

### Don't

- **Reference internal code** or implementation details
- **Use technical jargon** stakeholders won't understand
- **Write ambiguous criteria** ("should be fast", "user-friendly")
- **Include more than 6 criteria** (too complex)
- **Skip the "So that" clause** (always explain business value)
- **Make criteria untestable** (subjective or vague)

## Examples

### Example 1: CLI Feature

**File**: `requirements/cli/init_project.feature`

```gherkin
@cli @critical
Feature: Initialize a new project

  # ATDD Layer
  As a developer
  I want to initialize a CLI project with a single command
  So that I can quickly start development with proper structure

  Acceptance Criteria:
  - [ ] Creates project directory structure
  - [ ] Generates valid configuration file
  - [ ] Exits with clear success/error messages
  - [ ] Handles existing projects gracefully
  - [ ] Command completes in under 2 seconds
  - [ ] Works on Linux, macOS, and Windows
```

**Why this works**:

- User story clearly identifies WHO (developer), WHAT (initialize project), WHY (quick start)
- Criteria are measurable (can verify directory exists, config is valid, timing is under 2s)
- Includes functional requirements (creates structure, handles errors)
- Includes non-functional requirements (performance, cross-platform)
- All criteria can be independently verified

### Example 2: VS Code Extension Feature

**File**: `requirements/vscode/commit_button.feature`

```gherkin
@vscode @critical
Feature: Generate commit messages via button

  # ATDD Layer
  As a developer using VS Code
  I want to generate semantic commit messages by clicking a button
  So that I can create consistent, well-formatted commits without manual effort

  Acceptance Criteria:
  - [ ] Button appears in VS Code Source Control panel
  - [ ] Generated message follows semantic commit format
  - [ ] Message reflects actual code changes accurately
  - [ ] Generation completes within 5 seconds
  - [ ] Handles multi-file commits correctly
```

**Why this works**:

- Targets specific user persona (VS Code developer)
- Business value is clear (consistency without manual effort)
- Criteria cover UX (button placement), quality (accurate messages), performance (5s), and edge cases (multi-file)

### Example 3: Documentation Feature

**File**: `requirements/docs/build_docs.feature`

```gherkin
@cli @io @critical
Feature: Build documentation site

  # ATDD Layer
  As a documentation maintainer
  I want to build a static documentation site with one command
  So that I can deploy updated docs to hosting quickly

  Acceptance Criteria:
  - [ ] Builds complete in under 30 seconds
  - [ ] Generates valid HTML output
  - [ ] Works without internet connection (offline build)
  - [ ] Preserves images and assets correctly
  - [ ] Exits with build summary and error count
```

**Why this works**:

- Role is specific (documentation maintainer, not general developer)
- Business value emphasizes speed of deployment
- Performance criterion is realistic for doc builds (30s)
- Offline capability is a non-functional requirement
- Error reporting is part of acceptance

### Example 4: Error Handling Feature

**File**: `requirements/cli/handle_config_errors.feature`

```gherkin
@cli @error
Feature: Handle configuration file errors gracefully

  # ATDD Layer
  As a CLI user
  I want clear error messages when configuration is invalid
  So that I can quickly fix issues without debugging

  Acceptance Criteria:
  - [ ] Identifies specific line number with syntax error
  - [ ] Suggests valid syntax or correction
  - [ ] Exits with non-zero code (failure)
  - [ ] Error message appears on stderr (not stdout)
  - [ ] Validates config before executing commands
```

**Why this works**:

- Focuses on user experience during error scenarios
- Business value is explicit (quick fixes without debugging)
- Criteria include technical details (stderr, exit codes) that impact usability
- Validation timing is specified (before execution)

## Common Patterns

### Performance Requirements

```gherkin
Acceptance Criteria:
- [ ] Command completes in under 2 seconds for typical use
- [ ] Handles files up to 10MB without timeout
- [ ] Response time degrades gracefully with large inputs
```

### Cross-Platform Requirements

```gherkin
Acceptance Criteria:
- [ ] Works on Linux, macOS, and Windows
- [ ] Handles platform-specific path separators correctly
- [ ] Uses UTF-8 encoding consistently across platforms
```

### Usability Requirements

```gherkin
Acceptance Criteria:
- [ ] Provides progress indicators for long operations
- [ ] Exits with clear success/error messages
- [ ] Includes helpful examples in error messages
- [ ] Supports --help flag with usage examples
```

### Integration Requirements

```gherkin
Acceptance Criteria:
- [ ] Integrates with Docker without local installation
- [ ] Reads configuration from standard locations
- [ ] Respects environment variables for customization
- [ ] Works with CI/CD pipelines (non-interactive mode)
```

## Validation Checklist

Use this checklist when reviewing ATDD content:

### User Story

- [ ] Has "As a [role]" clause identifying the user
- [ ] Has "I want [capability]" clause describing the feature
- [ ] Has "So that [value]" clause explaining business benefit
- [ ] Written from user perspective (not technical)
- [ ] Clear and understandable to non-developers

### Acceptance Criteria

- [ ] Uses checkbox format: `- [ ]`
- [ ] Contains 2-6 measurable criteria
- [ ] Each criterion is testable (pass/fail, no ambiguity)
- [ ] Includes functional requirements (what it does)
- [ ] Includes non-functional requirements (how well it does it)
- [ ] Free from implementation details
- [ ] Uses stakeholder-friendly language
- [ ] All criteria are achievable in one feature iteration

### File Organization

- [ ] Saved in correct module folder (`requirements/cli/`, `requirements/vscode/`, etc.)
- [ ] File name is descriptive (`init_project.feature`, not `feature1.feature`)
- [ ] Feature has appropriate tags (`@critical`, `@cli`, etc.)

## Migration

### Enhancing Legacy BDD-Only Feature Files

Many existing `.feature` files contain only BDD scenarios without ATDD context.

#### Migration Strategy

1. **Identify legacy files**: Look for files missing user story
2. **Prioritize**: Start with `@critical` tagged features
3. **Consult stakeholders**: Gather business context if missing
4. **Add ATDD layer**: Insert at top of file (before scenarios)
5. **Preserve existing content**: Don't modify working BDD scenarios

#### Before (BDD-only)

**File**: `requirements/docs/build.feature`

```gherkin
@cli @success
Feature: Build documentation

  Scenario: Build docs successfully
    Given Docker is running
    When I call "build-docs"
    Then static site is created
```

#### After (ATDD + BDD)

**File**: `requirements/docs/build_docs.feature` (renamed for clarity)

```gherkin
@cli @critical
Feature: Build documentation site

  # ATDD Layer: Add business context
  As a documentation maintainer
  I want to build a static documentation site with one command
  So that I can deploy updated docs to hosting quickly

  Acceptance Criteria:
  - [ ] Builds complete in under 30 seconds
  - [ ] Generates valid HTML output
  - [ ] Works without internet connection

  # BDD Layer: Keep existing scenarios
  @success
  Scenario: Build docs successfully
    Given Docker is running
    When I call "build-docs"
    Then static site is created
    And site contains valid HTML files
```

**Changes made**:

- Added user story with clear WHO/WHAT/WHY
- Added measurable acceptance criteria
- Renamed file to be more descriptive
- Enhanced scenario with additional verification step
- Preserved original scenario logic

## Integration with BDD and TDD

### Flow from ATDD to BDD

Once ATDD acceptance criteria are defined:

1. **Extract testable behaviors** from each criterion
2. **Write BDD scenarios** that verify those behaviors
3. **See**: [BDD Workflow Guide](./bdd.md#workflow)

**Example**:

ATDD Criterion:

```text
- [ ] Creates project directory structure
```

BDD Scenarios:

```gherkin
Scenario: Initialize creates required directories
  Given I am in an empty folder
  When I run "cc init"
  Then a directory named "src/" should exist
  And a directory named "tests/" should exist
```

### Flow from ATDD through TDD

Acceptance criteria inform unit test coverage:

1. **Identify internal components** needed for criteria
2. **Write unit tests** for those components (TDD)
3. **Link tests to feature** with `// Feature:` comments
4. **See**: [TDD Workflow Guide](./tdd.md#workflow)

**Example**:

ATDD Criterion:

```text
- [ ] Command completes in under 2 seconds
```

Unit Test (pseudocode - see [TDD guide](./tdd.md) for language-specific examples):

```text
# Feature: init_project
test_init_performance() {
    start_time = current_time()
    init_project()
    duration = current_time() - start_time

    assert duration < 2_seconds, "Init took too long"
}
```

## Related Resources

- **[BDD Guide](./bdd.md)** - Write scenarios that verify acceptance criteria
- **[TDD Guide](./tdd.md)** - Implement features with unit tests
- **[Testing Overview](./index.md)** - Understand the complete testing strategy

---

**Next**: Learn how to translate acceptance criteria into [BDD scenarios](./bdd.md).
