<!-- EDITOR
# Editor: how-to-guides/specifications/working-with-specifications.md

## Soul

Practical guide for writing and organizing Gherkin specifications using the unified Rule/Scenario format.

## Sections

1. Quick Start
2. Writing Rule Blocks
3. Writing Scenario Blocks
4. File Organization
5. From Example Mapping Cards to Gherkin
6. Updating Specifications
7. Common Patterns
8. Related Documentation

## Notes

- Focused how-to guide with practical steps
- Links to explanation docs for conceptual content
- ~150-200 lines with actionable guidance
-->

# Working with specifications

Practical guide for writing and organizing Gherkin specifications.

---

## Quick Start

**Location**: All specifications live in `specs/<module>/<feature>/specification.feature`

**Format**: Gherkin with Rules (acceptance criteria) and Scenarios (executable examples)

**Basic structure**:

```gherkin
Feature: feature-name

  Rule: Acceptance criterion in business language

    @ov
    Scenario: Specific example of the rule
      Given [precondition]
      When [action]
      Then [expected outcome]
```

**See**: [Understanding BDD & Specifications](../../explanation/specifications/working-with-specifications.md) for conceptual foundation.

---

## Writing Rule Blocks

Rules define **acceptance criteria** - what the system must do.

### Rule Format

```gherkin
Rule: [Clear statement of business requirement]
```

### Best Practices

**Write measurable acceptance criteria**:

```gherkin
# Good - measurable
Rule: Creates 3 directories (src/, tests/, docs/)
Rule: Command completes in under 2 seconds
Rule: Error message contains "already initialized" text

# Bad - subjective
Rule: The interface is user-friendly
Rule: Performance is good
Rule: Error messages are helpful
```

**Use business language**:

```gherkin
# Good - domain language
Rule: Users must provide verified contact information

# Bad - technical jargon
Rule: Email field must match regex pattern
```

**One Rule per acceptance criterion** - If you have multiple distinct requirements, use multiple Rule blocks.

---

## Writing Scenario Blocks

Scenarios provide **executable examples** that verify Rules.

### Scenario Format

```gherkin
@ov
Scenario: [Describe the specific example]
  Given [setup - what's already true]
  When [action - what happens]
  Then [outcome - what we expect]
  And [additional outcomes]
```

### Step Guidelines

**Focus on behavior, not implementation**:

```gherkin
# Good - observable behavior
When I run "r2r init"
Then a file named "r2r.yaml" should be created

# Bad - implementation details
When the ConfigManager loads the file
Then the Config struct is populated
```

**Use concrete examples**:

```gherkin
# Good - specific
When I provide my email "user@example.com"

# Bad - vague
When I provide my email
```

**Keep scenarios focused** - One scenario tests one path through one rule.

### Required Tags

- `@ov` - Operational Verification (marks executable scenarios)
- Additional tags: `@risk-control:<name>-<id>`, `@gxp`, etc. (See: [Tag Reference](../../explanation/specifications/tag-reference.md))

---

## File Organization

### Directory Structure

```
specs/
└── <module>/
    └── <feature>/
        ├── specification.feature    # Rules & Scenarios
        └── issues.md               # Questions/blockers (pink cards)

src/
└── <module>/
    └── tests/
        └── steps_test.go           # Step implementations
```

**Separation**: Specifications (`specs/`) are business-readable. Implementations (`src/`) are technical.

### Example File

```gherkin
# specs/eac-commands/init/specification.feature

Feature: eac_init

  Rule: Initializes EAC workspace structure

    @ov
    Scenario: Initialize new workspace
      Given I am in an empty directory
      When I run "eac init"
      Then a directory named "src/" should be created
      And a directory named "tests/" should be created
      And a directory named "docs/" should be created
      And I should see "Workspace initialized"

    @ov
    Scenario: Prevent double initialization
      Given I have already run "eac init"
      When I run "eac init" again
      Then I should see an error "already initialized"
      And no new directories should be created
```

**See**: [Gherkin File Organization](../../explanation/specifications/gherkin-concepts.md) for detailed structure.

---

## From Example Mapping Cards to Gherkin

After running an Example Mapping workshop, translate cards to Gherkin:

| Card Color | Maps To | Action |
|------------|---------|--------|
| Yellow | Feature description | Write `Feature:` line |
| Blue | Acceptance criteria | Write `Rule:` blocks |
| Green | Concrete examples | Write `Scenario:` blocks under Rules |
| Pink | Questions/unknowns | Create `issues.md` |

### Step-by-Step

1. **Create feature file**: `specs/<module>/<feature>/specification.feature`
2. **Add Feature line**: Use yellow card content
3. **For each blue card**: Create a `Rule:` block
4. **For each green card**: Create a `Scenario:` under the relevant Rule
5. **Add tags**: At minimum `@ov` for each executable scenario
6. **Document questions**: Create `issues.md` for pink cards

**Example**:

```
Yellow: "User Registration"
Blue: "Users must provide contact information"
Green: "User provides email user@example.com"
Green: "User provides invalid email format"
```

Becomes:

```gherkin
Feature: cli_user-registration

  Rule: Users must provide contact information

    @ov
    Scenario: User provides valid email
      Given I am registering a new account
      When I provide my email "user@example.com"
      Then my account should be created

    @ov
    Scenario: User provides invalid email format
      Given I am registering a new account
      When I provide my email "not-an-email"
      Then I should see an error "Invalid email format"
```

**See**: [Example Mapping Guide](../../explanation/specifications/example-mapping.md) for workshop process.

---

## Updating Specifications

Specifications evolve as understanding deepens. Update them when:

| Trigger | Action |
|---------|--------|
| Implementation reveals edge case | Add new scenario |
| Stakeholder feedback | Refine Rule/Scenario wording |
| Production bug | Add regression scenario |
| Domain language evolves | Refactor to use new terms |

### Update Process

1. **Edit specification file** in `specs/`
2. **Update step definitions** in `src/` if needed
3. **Run tests** to verify changes
4. **Commit together** - spec and code in same commit

**Important**: Keep specifications synchronized with code. Don't accumulate "specification debt."

**See**: [Review and Iterate](../../explanation/specifications/review-and-iterate.md) for detailed evolution practices.

---

## Common Patterns

### Multiple Examples of Same Rule

```gherkin
Rule: Command validates input format

  @ov
  Scenario: Valid input accepted
    When I run "cmd --flag value"
    Then the command should succeed

  @ov
  Scenario: Missing flag rejected
    When I run "cmd value"
    Then I should see "flag required"

  @ov
  Scenario: Invalid flag rejected
    When I run "cmd --wrong value"
    Then I should see "unknown flag"
```

### Background (Shared Setup)

```gherkin
Feature: file-operations

  Background:
    Given I have a file "test.txt"

  Rule: Can read files

    @ov
    Scenario: Read existing file
      When I read "test.txt"
      Then I should see the file contents
```

**Use sparingly** - Explicit Given steps are often clearer.

### Scenario Outlines (Data-Driven)

```gherkin
Rule: Validates email format

  @ov
  Scenario Outline: Email validation
    When I provide email "<email>"
    Then I should see "<result>"

    Examples:
      | email              | result  |
      | user@example.com   | success |
      | invalid            | error   |
      | user@              | error   |
```

**Use when**: Testing same logic with different data.

---

## Related Documentation

### Explanation (Concepts)

- [Understanding BDD & Specifications](../../explanation/specifications/working-with-specifications.md) - Why BDD, key principles, evolution
- [Three-Layer Testing](../../explanation/specifications/three-layer-approach.md) - Rule/Scenario/Unit test relationship
- [Ubiquitous Language](../../explanation/specifications/ubiquitous-language.md) - Domain vocabulary
- [Example Mapping](../../explanation/specifications/example-mapping.md) - Discovery workshops
- [Gherkin Concepts](../../explanation/specifications/gherkin-concepts.md) - Structure and organization

### How-To Guides

- [BDD Development Workflow](./bdd-development-workflow.md) - Execute and implement tests
- [Reviewing Specifications](./reviewing-specifications.md) - Conduct reviews
- [Creating Feature Files](./creating-feature-files.md) - Start new specifications

### Reference

- [Tag Reference](../../reference/specifications/tag-reference.md) - All available tags
- [Gherkin Limits](../../reference/specifications/gherkin-limits.md) - Syntax constraints and best practices
