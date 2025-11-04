# Establish Ubiquitous Language

> **Goal**: Build shared domain vocabulary that flows from workshops through to code

## Prerequisites

- Access to domain experts, developers, QA, and product owners
- Understanding of [Ubiquitous Language concepts](../../explanation/specifications/ubiquitous-language.md)
- Materials: Sticky notes and packing paper (Event Storming), index cards (Example Mapping)

---

## The Flow

```text
Event Storming → Document Glossary → Example Mapping → specification.feature → Code
  (discover)       (capture terms)      (apply terms)     (executable specs)   (implement)
```

---

## Step 1: Run Event Storming Workshop

**Purpose**: Discover domain vocabulary collaboratively

**See**: [Run Event Storming Workshop](./run-event-storming.md) for detailed facilitation guide

**Quick summary**:

1. **Map domain events** (orange): "Order Placed", "Payment Received"
2. **Identify actors** (yellow): "Customer", "Manager"
3. **Capture commands** (blue): "Place Order", "Approve Request"
4. **Document policies** (pink): "Orders over $10K require approval"
5. **Surface definitions** (white): Clarify terms when confusion arises
6. **Identify bounded contexts**: Note where same word means different things

**Duration**: 2-4 hours

**Key output**: Domain vocabulary and glossary

---

## Step 2: Document the Glossary

**Purpose**: Create reference of agreed-upon terminology

### Create Glossary File

**Location**: `docs/glossary.md` or `specs/<module>/glossary.md`

**Format**:

```markdown
# Domain Glossary

## Order
**Context**: Sales, Fulfillment
**Definition**: Customer's request to purchase products
**States**: Draft, Submitted, Awaiting Approval, Approved, Fulfilled, Cancelled

## Manager
**Definition**: Employee with approval authority for high-value transactions

## Customer (Sales Context)
**Definition**: Potential buyer, including leads and prospects

## Customer (Support Context)
**Definition**: Paying subscriber with active support contract
```

### Document Bounded Contexts

```markdown
## Bounded Contexts

**Sales**: Customer (prospect), Lead, Quote, Sales Order
**Support**: Customer (subscriber), Ticket, SLA
**Billing**: Customer (payer), Invoice, Payment
```

---

## Step 3: Apply Vocabulary in Example Mapping

**Purpose**: Use established terms for specific features

**See**: [Run Example Mapping Workshop](./run-example-mapping.md) for detailed guide

**Key principle**: Use exact terms from glossary—no translation

**Example cards**:

- **Yellow** (Story): "As a **Manager**, I want to **approve large orders**..."
- **Blue** (Rule): "**Orders** over $10,000 require **Manager** **approval**"
- **Green** (Example): "Given **order** of $15,000, when **Manager** **approves**..."
- **Red** (Question): "What if **Manager** is unavailable?"

**Duration**: 15-25 minutes per feature

---

## Step 4: Write specification.feature Using Domain Terms

**Purpose**: Create executable specifications with Ubiquitous Language

**File**: `specs/<module>/<feature>/specification.feature`

**Template using domain terms**:

```gherkin
@sales @critical @order-approval
Feature: sales_order-approval

  As a manager
  I want to approve large orders
  So that we control high-value transactions

  Rule: Orders over $10,000 must be approved by a manager

    @success @ac1
    Scenario: Manager approves large order
      Given an order with amount $15,000
      And the order status is "Awaiting Approval"
      When the manager approves the order
      Then the order status should be "Approved"
      And the order should be eligible for fulfillment

    @error @ac1
    Scenario: Non-manager cannot approve orders
      Given an order with amount $15,000
      When a sales representative attempts to approve the order
      Then the system should reject the approval
      And the order status should remain "Awaiting Approval"
```

**Key**: Every term—"Order", "Manager", "Approved", "Awaiting Approval"—matches the glossary.

**See**: [Create Specifications](./create-specifications.md) for complete guide

---

## Step 5: Implement Code Using Domain Terms

**Purpose**: Ensure code reflects business concepts exactly

### Name Classes and Methods with Domain Language

```go
// Package sales implements the Sales bounded context
package sales

// Order represents a customer's purchase request (from glossary)
type Order struct {
    ID     string
    Amount Money
    Status OrderStatus
}

// OrderStatus represents order states (from glossary)
type OrderStatus string

const (
    OrderStatusDraft            OrderStatus = "Draft"
    OrderStatusAwaitingApproval OrderStatus = "Awaiting Approval"
    OrderStatusApproved         OrderStatus = "Approved"
)

// Approve marks order as approved (domain verb from Event Storming)
func (o *Order) Approve(manager Manager) error {
    if o.Status != OrderStatusAwaitingApproval {
        return ErrCannotApprove
    }
    o.Status = OrderStatusApproved
    return nil
}
```

**Key principles**:
- Type names match domain concepts: `Order`, `Manager`
- Method names use domain verbs: `Approve()`, not `SetStatusToApproved()`
- States use exact glossary terms: `"Awaiting Approval"`

### Organize by Bounded Context

```text
src/
  sales/           # Sales context
    order.go       # Order aggregate
    manager.go     # Manager entity
  support/         # Support context (different Customer definition)
    customer.go
    ticket.go
  billing/         # Billing context (different Customer definition)
    customer.go
    invoice.go
```

---

## Step 6: Evolve Language Continuously

**Purpose**: Keep language aligned as understanding deepens

### When Language Changes

**Scenario**: Team discovers "Validated" means two different things

1. **Event Storming** identifies issue → Split into "Verified" (data correct) and "Approved" (business accepts)
2. **Update glossary** with new terms
3. **Refactor specifications**: Search/replace "Validated" with appropriate term
4. **Refactor code**: Rename types/methods using IDE tools

### Regular Updates

- **Event Storming**: Quarterly or when starting major features
- **Glossary**: Add terms as discovered, refine definitions
- **Archive deprecated terms**: Note evolution ("Previously 'Validated', now 'Verified' or 'Approved'")

---

## Complete Flow Example

**Event Storming discovers**:
- Event: "Order Approved"
- Actor: "Manager"
- Policy: "Large orders require approval"

**Glossary documents**:
```markdown
## Order
States: Draft, Awaiting Approval, Approved, Fulfilled
```

**Example Mapping applies**:
- Blue card: "Orders over $10K require Manager approval"
- Green card: "$15K order → Manager approves → Approved"

**specification.feature preserves**:
```gherkin
Rule: Orders over $10,000 must be approved by a manager
  Scenario: Manager approves large order
    Given an order with amount $15,000
    When the manager approves the order
    Then the order status should be "Approved"
```

**Code implements**:
```go
func (o *Order) Approve(manager Manager) error {
    o.Status = OrderStatusApproved
    return nil
}
```

**Result**: Same vocabulary from workshop → specs → code. No translation errors.

---

## Related Guides

- [Run Event Storming Workshop](./run-event-storming.md) - Domain discovery facilitation
- [Run Example Mapping Workshop](./run-example-mapping.md) - Feature requirements discovery
- [Create Specifications](./create-specifications.md) - Write specification.feature files
- [Understanding Ubiquitous Language](../../explanation/specifications/ubiquitous-language.md) - Conceptual foundation

---

## Summary

**Six steps to establish Ubiquitous Language**:

1. **Event Storming** - Discover domain vocabulary collaboratively
2. **Document glossary** - Capture terms and bounded contexts
3. **Example Mapping** - Apply vocabulary to specific features
4. **Write specifications** - Use domain terms in specification.feature
5. **Implement code** - Name classes/methods with domain language
6. **Evolve continuously** - Update as understanding deepens

**Outcome**: Specifications business can validate, code reflecting business concepts, shared vocabulary eliminating translation errors.
