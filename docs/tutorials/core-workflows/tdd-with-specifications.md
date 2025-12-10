# TDD with Specifications

{{ page_breadcrumb() }}

**Status:** Placeholder - Content coming soon

**Prerequisites:** [Your First Specification](../getting-started/first-specification.md), [Your First Module](../getting-started/first-module.md)

## Planned Content

This tutorial teaches the complete test-driven development loop using executable Gherkin specifications with Godog.

### What You'll Learn

- Start with a feature idea and write a Gherkin specification
- Generate step definition scaffolding
- Implement step definitions that connect specs to code
- Follow the TDD cycle: Write failing test → Implement → Tests pass
- Run specifications with `r2r test suite acceptance`
- Understand the relationship between specifications and unit tests

### Tutorial Structure

1. **Choose a feature to implement**
   - Example: "Calculate tax on invoice"
   - Define the business rules
   - Identify edge cases

2. **Write the Gherkin specification**
   - Create `specs/billing/calculate-tax/specification.feature`
   - Define Feature, Rules, and Scenarios
   - Use concrete examples with real values
   - Tag appropriately (@L2 @ov @deps:go)

3. **Generate step definition scaffolding**
   - Run the specification to get undefined steps
   - Copy generated step signatures
   - Create `go/billing/tests/tax_steps.go`

4. **Implement step definitions (RED)**
   - Wire Gherkin steps to test code
   - Set up test context and state
   - Assert expected outcomes
   - Steps fail because implementation doesn't exist

5. **Write unit tests (RED)**
   - Create `go/billing/tax_test.go`
   - Write table-driven tests for tax calculation
   - Tests fail because implementation doesn't exist

6. **Implement the feature (GREEN)**
   - Create `go/billing/tax.go`
   - Implement tax calculation logic
   - Run tests: `r2r test billing`
   - All tests pass

7. **Refactor and validate**
   - Improve code clarity
   - Extract helper functions
   - Run full test suite
   - Run specification: `r2r test suite acceptance`

### Complete Example

The tutorial will implement a tax calculation feature end-to-end:

**Specification:**
```gherkin
@L2 @ov @deps:go
Feature: billing_calculate-tax

  As a billing system
  I want to calculate tax on invoices
  So that customers are charged correctly

  Rule: Standard tax rate is 20%

    Scenario: Calculate tax on taxable invoice
      Given an invoice with subtotal $100.00
      When I calculate tax
      Then the tax amount should be $20.00
      And the total should be $120.00
```

**Step Definitions:**
```go
func (s *TaxContext) anInvoiceWithSubtotal(amount string) error {
    s.invoice = &Invoice{Subtotal: parseAmount(amount)}
    return nil
}

func (s *TaxContext) iCalculateTax() error {
    s.result = CalculateTax(s.invoice)
    return nil
}

func (s *TaxContext) theTaxAmountShouldBe(expected string) error {
    if s.result.Tax != parseAmount(expected) {
        return fmt.Errorf("expected %s, got %s", expected, s.result.Tax)
    }
    return nil
}
```

**Implementation:**
```go
func CalculateTax(invoice *Invoice) *TaxResult {
    tax := invoice.Subtotal * 0.20
    total := invoice.Subtotal + tax
    return &TaxResult{Tax: tax, Total: total}
}
```

### Key Concepts Covered

- The TDD cycle: Red → Green → Refactor
- Executable specifications with Godog
- Step definition patterns and context management
- Relationship between specifications and unit tests
- Running acceptance tests vs. unit tests

### Three Rules of Vibe Coding Applied

The tutorial demonstrates how this approach satisfies:

- **Easy to understand**: Specification describes behavior clearly
- **Easy to change**: Tests enable safe refactoring
- **Hard to break**: Tests catch regressions immediately

### Next Steps

After completing this tutorial, you'll understand the core TDD workflow. Continue to [Building and Testing Changes](./building-and-testing.md) to learn how to efficiently build and validate your changes.

{{ diataxis_footer() }}
