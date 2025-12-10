# Your First Feature Specification

{{ page_breadcrumb() }}

Learn how to write Gherkin feature specifications that describe system behavior in a structured, testable format using the Given/When/Then pattern.

**Estimated time:** 20 minutes
**Prerequisites:** [Quick Start Guide](./quick-start.md), basic understanding of software requirements

## What You'll Learn

By the end of this tutorial, you'll be able to:

- Understand the structure of Gherkin feature specifications
- Write specifications using the Given/When/Then pattern
- Use Features, Rules, and Scenarios to organize requirements
- Apply appropriate tags for test levels and verification types
- Validate specifications against quality standards
- Use AI to generate specifications from natural language
- Write effective, concrete scenarios

## What is a Feature Specification?

Feature specifications use Gherkin syntax to describe how software should behave. They follow the **Given/When/Then** pattern:

- **Given** - Initial context/preconditions
- **When** - Action or event
- **Then** - Expected outcome

## Step 1: Understand the Structure

A well-formed specification has three key components:

1. **Feature** - High-level description of the capability
2. **Rule** - Business rule or requirement (one or more per feature)
3. **Scenario** - Concrete example demonstrating the rule

Example structure:

```gherkin
Feature: Module_Feature-Name

  As a [user type]
  I want [capability]
  So that [benefit]

  Rule: [Business rule description]

    Scenario: [Specific example]
      Given [context]
      When [action]
      Then [outcome]
```

## Step 2: Choose Your Approach

You can create specifications in two ways:

### Option A: Manual Creation

Create a new file in the `specs/` directory following the naming convention:

```text
specs/[module-name]/[feature-name]/specification.feature
```

### Option B: AI-Assisted Generation

Use the `create spec` command to generate a specification from a description:

```bash
r2r create spec "Add user authentication with email and password"
```

The command will:

- Use AI to generate a proper Gherkin specification
- Infer the module from your description
- Save to the appropriate `specs/` directory
- Validate against the specification contract

For this tutorial, we'll create one manually to understand the structure.

## Step 3: Create Your First Specification

Create a new directory and file:

```bash
mkdir -p specs/my-app/user-greeting
touch specs/my-app/user-greeting/specification.feature
```

Open the file and add this content:

```gherkin
@L2 @ov @deps:go
Feature: my-app_user-greeting

  As a user
  I want to receive a personalized greeting
  So that I feel welcomed by the application

  Rule: Greeting must include the user's name

    Scenario: Display greeting with valid username
      Given I have a username "Alice"
      When I request a greeting
      Then I should see "Hello, Alice!"

    Scenario: Display default greeting without username
      Given I have no username
      When I request a greeting
      Then I should see "Hello, Guest!"

  Rule: Greeting must be displayed on application start

    Scenario: Show greeting when application starts
      Given the application is starting
      When the startup sequence completes
      Then the greeting is displayed to the user
```

## Step 4: Understand the Tags

Tags at the top of the feature provide metadata:

- `@L2` - Test level (L0=unit, L1=integration, L2=component, L3=system, L4=smoke)
- `@ov` - Verification type (ov=operational, iv=interface, pv=performance)
- `@deps:go` - Technology dependency (go, node, python, etc.)
- `@skip:wip` - Skip this test (wip=work in progress, broken, etc.)

Common verification tags:

- `@iv` - Interface Verification (tests contracts/APIs)
- `@ov` - Operational Verification (tests functionality)
- `@pv` - Performance Verification (tests performance/scalability)

## Step 5: Write Effective Scenarios

**Good scenarios are:**

- **Concrete** - Use specific examples, not abstract concepts
- **Independent** - Can run in any order
- **Focused** - Test one thing at a time

Example of a good scenario:

```gherkin
Scenario: Calculate total price with discount
  Given I have items worth $100 in my cart
  And I have a 20% discount code
  When I apply the discount code
  Then the total price should be $80
```

Example of a poor scenario (too vague):

```gherkin
Scenario: Discounts work correctly
  Given there are discounts
  When I use a discount
  Then it works
```

## Step 6: Validate Your Specification

Check that your specification follows the project's quality standards:

```bash
r2r validate specs specs/my-app/user-greeting/specification.feature
```

This validates:

- Proper Gherkin syntax
- Required structure (Feature/Rule/Scenario hierarchy)
- Tag formatting and requirements
- Step formatting

If validation fails, you'll see specific error messages indicating what needs to be fixed.

## Step 7: Run Your Specification

While this specification doesn't have implementation code yet, you can see how it would be discovered by the test runner:

```bash
r2r show tests
```

This lists all tests in the repository, including your new specification.

## Step 8: Explore Real Examples

Look at existing specifications in the repository for inspiration:

```bash
# Simple example
cat specs/eac-commands/show-help/specification.feature

# Complex example with multiple rules
cat specs/eac-commands/init/specification.feature
```

## Advanced: Using Background

If multiple scenarios share the same setup, use `Background`:

```gherkin
Feature: my-app_shopping-cart

  As a customer
  I want to manage items in my cart
  So that I can purchase products

  Background:
    Given I am logged in as "customer@example.com"
    And I have an empty shopping cart

  Rule: Items can be added to cart

    Scenario: Add single item
      When I add product "Widget" to cart
      Then my cart contains 1 item

    Scenario: Add multiple items
      When I add product "Widget" to cart
      And I add product "Gadget" to cart
      Then my cart contains 2 items
```

The `Background` steps run before each scenario.

## Best Practices

1. **Use Rules to organize scenarios** - Group related scenarios under descriptive rules
2. **Start with the user story** - The Feature description should explain the user value
3. **Be specific in scenarios** - Use concrete values, not placeholders
4. **Write declaratively** - Focus on *what*, not *how*
5. **Keep scenarios independent** - Don't rely on order or state from other scenarios

**Good (declarative):**

```gherkin
When I search for "laptop"
Then I should see 10 search results
```

**Bad (imperative):**

```gherkin
When I click the search box
And I type "laptop"
And I press enter
Then I should see 10 results in the results div
```

## What You Learned

Congratulations! You've successfully:

- ✅ Understood the structure of Gherkin specifications (Feature/Rule/Scenario)
- ✅ Written specifications using Given/When/Then pattern
- ✅ Applied tags for test levels and verification types
- ✅ Created concrete, independent, and focused scenarios
- ✅ Validated specifications against quality standards
- ✅ Learned both manual and AI-assisted specification creation

## Key Concepts Covered

- **Gherkin syntax** - Feature, Rule, Scenario, Given/When/Then structure
- **Specification tags** - Test levels (@L0-L4), verification types (@ov, @iv, @pv), dependencies (@deps:go)
- **Scenario quality** - Concrete examples, independence, focus, declarative style
- **AI-assisted generation** - Using `r2r create spec` to generate from descriptions
- **Validation** - Ensuring specifications meet quality standards with `r2r validate specs`
- **Background sections** - Sharing common setup across scenarios
- **Best practices** - Writing effective, maintainable specifications

## Next Steps

### Continue Learning
- **Next tutorial:** [Your First Module](./first-module.md) - Create a Go module with tests (30 min)
- **Advanced:** [TDD with Specifications](../core-workflows/tdd-with-specifications.md) - Full TDD development loop (45 min)

### Apply What You Learned
Now that you can write specifications, you can accomplish these tasks:

- **[Create Specifications](../../how-to-guides/eac/commands/documentation/create-specifications.md)** - Generate specs from descriptions with AI
- **[Validate Specifications](../../how-to-guides/eac/commands/build-test-validate/validate-specifications.md)** - Check specs against quality standards
- **Browse existing specs** - Look in the `specs/` directory for real-world examples

### Dive Deeper
- [Specifications Explanation](../../explanation/specifications/) - Understand BDD concepts and philosophy
- [Gherkin Best Practices](../../explanation/specifications/quality/) - Writing effective scenarios
- [Tag Taxonomy](../../explanation/specifications/taxonomy/) - Complete guide to test tags

{{ diataxis_footer() }}
