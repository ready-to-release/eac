# How to Establish Ubiquitous Language for Your Project

> **Goal**: Build a shared language that bridges business and technical teams through collaborative domain discovery

## Prerequisites

- Access to domain experts, developers, QA, and product owners
- Understanding of [Ubiquitous Language concepts](../../explanation/everything-as-code/building-shared-language.md)
- Materials: Colored sticky notes, brown packing paper (for Event Storming)
- Materials: Colored index cards (for Example Mapping)

## Overview

This guide walks you through establishing a Ubiquitous Language for your project, from initial domain discovery through to executable specifications. The process follows this flow:

```text
Event Storming → Document Language → Example Mapping → Write Specifications → Implement Code
```

---

## Step 1: Run Event Storming Workshops

### Start with Big Picture Event Storming

**Purpose**: Discover the domain vocabulary and understand how the business works.

**Who to invite**:

- Domain experts (business stakeholders who know the domain deeply)
- Developers (who will implement the system)
- QA engineers (who will test it)
- Product owners (who define requirements)

**What to do**:

1. **Set up the space**:
   - Tape brown packing paper on a wall (3-5 meters long)
   - Provide colored sticky notes (orange, blue, pink, yellow, purple)
   - Give everyone markers

2. **Map domain events** (orange stickies):
   - Ask: "What are the key things that happen in this business?"
   - Write events in past tense: "Order Placed", "Payment Received", "Customer Registered"
   - Place them on the timeline (left to right)

3. **Identify actors** (yellow stickies):
   - Ask: "Who or what triggers these events?"
   - Label: "Customer", "Manager", "Payment System"

4. **Capture commands** (blue stickies):
   - Ask: "What actions cause these events?"
   - Write: "Place Order", "Approve Request", "Process Payment"

5. **Document policies** (purple stickies):
   - Ask: "What business rules govern behavior?"
   - Write: "Orders over $10K require approval"

6. **Surface definitions** (when confusion arises):
   - When people use terms differently, write a definition sticky
   - Clarify: "What do we mean by 'Customer' in this context?"

7. **Identify bounded contexts**:
   - Notice where the same word means different things
   - Draw context boundaries on the paper

**Duration**: 2-4 hours for initial Big Picture session

**Key output**: The Ubiquitous Language emerges from these discussions. Capture all domain terms and their meanings.

---

## Step 2: Document the Ubiquitous Language

**Purpose**: Create a reference of agreed-upon terminology.

### Create a Glossary

Document each term discovered in Event Storming:

**Format**:

```markdown
## Glossary

### Order
**Context**: Sales, Fulfillment
**Definition**: A customer's request to purchase products, containing line items, quantities, and pricing.
**States**: Draft, Submitted, Awaiting Approval, Approved, Fulfilled, Cancelled

### Customer (Sales Context)
**Definition**: A potential buyer, including leads and prospects who haven't purchased yet.

### Customer (Support Context)
**Definition**: A paying subscriber with an active support contract.
```

**Where to store**:

- In your project documentation (e.g., `docs/glossary.md`)
- In Event Storming photos/notes
- In team wiki or knowledge base

### Capture Bounded Contexts

Document which terms belong to which contexts:

```markdown
## Bounded Contexts

### Sales Context
Uses: Sales Customer, Lead, Prospect, Quote, Sales Order

### Support Context
Uses: Support Customer, Ticket, SLA, Escalation

### Accounting Context
Uses: Billing Customer, Invoice, Payment, Account
```

---

## Step 3: Use Example Mapping for Features

**Purpose**: Apply the established vocabulary to specific features.

### Run Example Mapping Workshop

**Duration**: 15-25 minutes per feature

**Materials**:

- Yellow index cards (User Stories)
- Blue index cards (Rules/Acceptance Criteria)
- Green index cards (Examples/Scenarios)
- Red index cards (Questions)

**Process**:

1. **Start with the User Story** (yellow card):
   - Use Ubiquitous Language for the actor: "As a **Manager**..."
   - Use domain terms for capability: "...I want to **approve orders**..."
   - Use business language for value: "...so that we control high-value transactions"

2. **Define Rules** (blue cards):
   - Use exact domain terminology: "**Orders** over $10,000 must be **approved** by a **Manager**"
   - Reference domain concepts from Event Storming

3. **Create Examples** (green cards):
   - Use concrete scenarios with domain terms:
     - "Given an **order** of $15,000"
     - "When **manager** **approves** the **order**"
     - "Then **order** status is '**Approved**'"

4. **Flag Questions** (red cards):
   - Identify gaps in domain understanding
   - Note: "What happens if manager is unavailable?"
   - Follow up with Event Storming if needed

**Key principle**: Use the same terms from your Event Storming glossary. No translation.

---

## Step 4: Write Specifications Using Domain Terms

### Write ATDD Acceptance Specifications

**File**: `requirements/<module>/<feature>/acceptance.spec`

Use the Ubiquitous Language in your acceptance criteria:

```markdown
# Manager Order Approval

> **Feature ID**: sales_order_approval
> **Module**: Sales
> **Context**: Sales, Fulfillment

## User Story

* As a manager
* I want to approve large orders
* So that we control high-value transactions

## Acceptance Criteria

* Orders over $10,000 must be approved by a manager
* Approved orders can proceed to fulfillment
* Managers can see all orders awaiting approval

## Acceptance Tests

### AC1: Manager approval required for large orders

* Create order with amount $15,000
* Verify order status is "Awaiting Approval"
* Manager approves the order
* Verify order status is "Approved"
* Verify order is eligible for fulfillment
```

**Notice**: Every term matches your glossary—"Order", "Manager", "Approved", "Awaiting Approval".

### Write BDD Behavior Scenarios

**File**: `requirements/<module>/<feature>/behavior.feature`

Use Given/When/Then with domain terms:

```gherkin
# Feature ID: sales_order_approval
# Context: Sales, Fulfillment

@sales @critical @order_approval
Feature: Manager Order Approval

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
    And the order status is "Awaiting Approval"
    When a sales representative attempts to approve the order
    Then the system should reject the approval
    And the order status should remain "Awaiting Approval"
```

**Notice**: The vocabulary flows directly from Event Storming → Example Mapping → Specifications.

### Handle Multiple Contexts

When a feature spans multiple contexts, be explicit:

```gherkin
Scenario: Convert sales prospect to paying customer
  Given a sales prospect with contact information
  When the prospect signs a contract
  Then the prospect becomes a Support customer
  And a Billing customer record is created
```

This shows the transition between Sales, Support, and Billing contexts.

---

## Step 5: Implement Code Using Domain Terms

**Purpose**: Ensure code reflects the exact business concepts.

### Name Classes and Methods with Ubiquitous Language

**Example** (Go):

```go
// Package sales implements the Sales bounded context
package sales

// Order represents a customer's purchase request
type Order struct {
    ID              string
    Amount          Money
    Status          OrderStatus
    RequiresApproval bool
}

// OrderStatus represents the state of an order
type OrderStatus string

const (
    OrderStatusDraft            OrderStatus = "Draft"
    OrderStatusAwaitingApproval OrderStatus = "Awaiting Approval"
    OrderStatusApproved         OrderStatus = "Approved"
    OrderStatusFulfilled        OrderStatus = "Fulfilled"
)

// Approve marks the order as approved by a manager
func (o *Order) Approve(manager Manager) error {
    if o.Status != OrderStatusAwaitingApproval {
        return ErrCannotApprove
    }
    o.Status = OrderStatusApproved
    return nil
}
```

**Key principles**:

- Type names match domain concepts: `Order`, `Manager`, `OrderStatus`
- Method names use domain verbs: `Approve()`, not `SetStatusToApproved()`
- States use domain language: `"Awaiting Approval"`, not `"PENDING_AUTH"`
- Comments clarify domain meaning when needed

### Organize by Domain Concepts

Structure code around bounded contexts:

```text
src/
  sales/           # Sales context
    order.go       # Order aggregate
    manager.go     # Manager entity
  support/         # Support context
    customer.go    # Support customer (different from sales)
    ticket.go      # Support ticket
  billing/         # Billing context
    customer.go    # Billing customer
    invoice.go     # Invoice
```

---

## Step 6: Evolve the Language Continuously

**Purpose**: Keep the language aligned as understanding deepens.

### Regular Event Storming Sessions

Schedule periodic sessions to:

- Explore new domain areas
- Refine existing understanding
- Identify new bounded contexts
- Update terminology

**Frequency**: Quarterly, or when starting major new features

### Update Specifications When Language Changes

**Scenario**: Team realizes "Validated" means two different things.

1. **Event Storming identifies the issue**:
   - Split into "Verified" (data correctness) and "Approved" (business acceptance)

2. **Update the glossary**:

   ```markdown
   ### Verified
   **Definition**: Data has been checked for correctness and completeness

   ### Approved
   **Definition**: Business has accepted and authorized to proceed
   ```

3. **Refactor specifications**:
   - Search for "Validated" in all `.spec` and `.feature` files
   - Replace with appropriate term based on context
   - Review scenarios for accuracy

4. **Refactor code**:
   - Use IDE refactoring tools to rename types/methods
   - Update constants and enums
   - Update tests

### Maintain the Glossary

Keep your glossary current:

- Add new terms as they're discovered
- Refine definitions as understanding improves
- Archive deprecated terms with explanations
- Note evolution: "Previously called 'Validated', now split into 'Verified' and 'Approved'"

---

## Related Guides

- [Run Example Mapping Workshop](run-example-mapping.md) - Detailed Example Mapping facilitation
- [Create Feature Specifications](create-feature-spec.md) - Write ATDD/BDD specifications
- [Understanding Ubiquitous Language](../../explanation/everything-as-code/building-shared-language.md) - Conceptual background

---

## Summary

Establishing Ubiquitous Language is a continuous practice:

1. **Discover** vocabulary through Event Storming
2. **Document** terms in a glossary with bounded contexts
3. **Apply** language in Example Mapping workshops
4. **Preserve** language in specifications (ATDD/BDD)
5. **Implement** using exact domain terms in code
6. **Evolve** language as understanding deepens

The result: Specifications that business can validate, code that reflects business concepts, and a shared vocabulary that eliminates translation errors.
