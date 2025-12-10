# Effective BDD Scenarios

{{ page_breadcrumb() }}

**Status:** Placeholder - Content coming soon

**Estimated time:** 45 minutes
**Prerequisites:** [Your First Specification](../getting-started/first-specification.md), [TDD with Specifications](../core-workflows/tdd-with-specifications.md)

## Planned Content

This tutorial teaches advanced BDD techniques including Example Mapping for specification discovery, writing effective scenarios, and avoiding common anti-patterns.

### What You'll Learn

- Discover specifications through Example Mapping workshops
- Write concrete, independent, and focused scenarios
- Avoid BDD anti-patterns (imperative, vague, interdependent)
- Refactor specifications for clarity
- Use scenario outlines for data-driven tests
- Organize scenarios effectively
- Review and iterate on specifications

### Tutorial Structure

1. **Example Mapping technique**
   - Collaborative discovery workshop
   - User story → Rules → Examples → Questions
   - Color-coded cards (story, rules, examples, questions)
   - Facilitation techniques

2. **Example Mapping workshop**
   - Exercise: Map a user story
   - Identify rules and edge cases
   - Convert examples to scenarios
   - Resolve questions with stakeholders

3. **Writing effective scenarios**
   - CONCRETE: Specific values, not abstractions
   - INDEPENDENT: No order dependencies
   - FOCUSED: One behavior per scenario
   - DECLARATIVE: What, not how

4. **Common anti-patterns to avoid**
   - Imperative scenarios (UI-focused)
   - Vague scenarios (lacking specifics)
   - Interdependent scenarios (order matters)
   - Overly complex scenarios (too many steps)

5. **Scenario outlines**
   - Data-driven testing with Examples table
   - When to use outlines vs. separate scenarios
   - Keeping outlines readable

6. **Refactoring specifications**
   - Extract common setup to Background
   - Split complex scenarios
   - Improve step clarity
   - Remove duplication

7. **Organizing scenarios**
   - Group by Rules
   - Name scenarios descriptively
   - Order: happy path → edge cases → error cases
   - File organization strategies

8. **Review and iteration**
   - Specification review checklist
   - Collaborate with stakeholders
   - Refine based on feedback
   - Keep specifications living documents

### Example Mapping Exercise

The tutorial includes a complete Example Mapping session:

**User Story:**
```
As a customer
I want to apply discount codes to my order
So that I can save money on purchases
```

**Rules discovered:**
- Discount codes must be valid and not expired
- Only one discount code per order
- Discount applies before tax
- Some products excluded from discounts

**Examples (become scenarios):**
- Valid code with 20% discount → price reduced
- Expired code → error message
- Apply second code → first code removed
- Discount on excluded product → discount not applied

**Questions:**
- Can codes be combined? (No → Rule)
- What if code format is invalid? (Error → Scenario)

### Scenario Quality Checklist

Good scenarios are:
- [ ] Concrete (specific values)
- [ ] Independent (any order)
- [ ] Focused (one behavior)
- [ ] Declarative (what, not how)
- [ ] Readable (business language)
- [ ] Valuable (demonstrate requirement)

### Key Concepts Covered

- Example Mapping facilitation
- Scenario quality characteristics
- BDD anti-patterns
- Scenario outlines for data-driven tests
- Specification refactoring
- Collaborative discovery

### Before and After Examples

**BEFORE (Poor scenario):**
```gherkin
Scenario: Discounts work
  Given there are discounts
  When I use one
  Then it works
```

**AFTER (Effective scenario):**
```gherkin
Scenario: Apply 20% discount to eligible order
  Given I have items worth $100.00 in my cart
  And I have a valid discount code "SAVE20" for 20% off
  When I apply the discount code
  Then my subtotal should be $80.00
  And my total including tax should be $96.00
```

### Scenario Outline Example

```gherkin
Rule: Different discount types calculate correctly

  Scenario Outline: Apply discount to order
    Given I have items worth $<subtotal> in my cart
    And I have a discount code for <discount>% off
    When I apply the discount code
    Then my discounted subtotal should be $<discounted>

    Examples:
      | subtotal | discount | discounted |
      | 100.00   | 10       | 90.00      |
      | 100.00   | 20       | 80.00      |
      | 100.00   | 50       | 50.00      |
      | 50.00    | 20       | 40.00      |
```

### Best Practices

- Use Example Mapping before writing specifications
- Write scenarios in business language
- Keep scenarios short (< 10 steps)
- Use Background for common setup
- Review specifications with stakeholders
- Refactor specifications like code
- Make scenarios executable

### Tools and Techniques

- Example Mapping (sticky notes or digital)
- Scenario quality checklist
- Specification review sessions
- Living documentation
- Scenario refactoring patterns

### Next Steps

After completing this tutorial, you'll write high-quality specifications. Explore other specialized topics based on your needs: [Architecture Documentation](./architecture-documentation.md), [Security Scanning](./security-scanning-workflow.md), or [TypeScript Setup](./typescript-module-setup.md).

{{ diataxis_footer() }}
