# Ubiquitous Language: Building Shared Vocabulary

> **How shared domain vocabulary creates the foundation for executable specifications**

---

## The Problem: Lost in Translation

Software teams often suffer from a fundamental communication problem: **business stakeholders and technical teams speak different languages**.

**Common misalignments:**

- Business says "customer" → Developers think "user"
- Product owner says "order" → Developers think "transaction"
- Domain expert says "approved" → Developers think "validated"
- Stakeholder says "account" → Developers think "record"

**The consequences:**

When teams use different terminology, specifications fail to bridge the gap between business intent and technical implementation:

- **Rework** - Teams discover they meant different things after implementation
- **Bugs** - Code implements the wrong concept because terminology was misunderstood
- **Failed validation** - Business stakeholders can't validate specifications written in technical jargon
- **Drift** - Specifications use one set of terms, code uses another, creating disconnect

**The root cause:**

Without a shared language, every handoff is a translation:

```text
Business Language → Requirements Document → Developer Interpretation → Code

Each translation introduces potential misalignment
```

---

## The Solution: Ubiquitous Language (DDD)

Domain-Driven Design (DDD) provides a solution: **Ubiquitous Language**.

### What is Ubiquitous Language?

> **Ubiquitous Language** is the term Eric Evans uses in Domain Driven Design for the practice of building up a common, rigorous language between developers and users.
>
> This language should be based on the Domain Model used in the software - hence the need for it to be rigorous, since software doesn't cope well with ambiguity.
>
> Evans makes clear that using the ubiquitous language in conversations with domain experts is an important part of testing it, and hence the domain model. He also stresses that the language (and model) should evolve as the team's understanding of the domain grows.
>
> — Martin Fowler, [Ubiquitous Language](https://martinfowler.com/bliki/UbiquitousLanguage.html)

### Key Characteristics

**Rigorous**: The language must be precise. Software doesn't handle ambiguity—"sort of like this" doesn't compile.

**Shared**: Everyone uses the same terms to mean the same things—business experts, developers, QA, product owners.

**Domain-based**: The language comes from the business domain, not from technical implementation.

**Evolving**: As the team's understanding of the domain deepens, the language refines.

### Why It Matters

When everyone uses the Ubiquitous Language:

- **Conversations are productive** - No translation required between business and technical discussions
- **Specifications are validatable** - Business stakeholders can read and confirm requirements
- **Code reflects intent** - Implementation uses the exact concepts the business understands
- **Tests document the domain** - Automated tests become living documentation in domain terms

---

## Discovering Language Through Collaboration

Ubiquitous Language doesn't emerge spontaneously—it's **discovered** through collaborative domain exploration.

### Event Storming: Domain Discovery

**Event Storming** is a collaborative workshop technique that discovers domain vocabulary by visually mapping business events and processes.

**Key outcomes**:

- Domain events, actors, commands, and policies
- Bounded contexts where different languages apply
- **Most importantly**: The Ubiquitous Language itself

**See**: [Event Storming](./event-storming.md) for comprehensive workshop guide including:

- Three Event Storming formats (Big Picture, Process Modeling, Software Design)
- Workshop facilitation and sticky note grammar
- Capturing domain vocabulary and glossaries
- Best practices and common pitfalls

### Example Mapping: Requirements Discovery

**Example Mapping** applies the established Ubiquitous Language to specific features through time-boxed workshops using colored cards.

**How it uses Ubiquitous Language**:

| Card Color | Purpose | Uses Domain Terms For |
|------------|---------|----------------------|
| **Yellow** | User Story | WHO (actor), WHAT (capability), WHY (value) |
| **Blue** | Rules/Acceptance Criteria | Business rules using domain vocabulary |
| **Green** | Examples/Scenarios | Concrete situations using domain concepts |
| **Red** | Questions/Unknowns | Gaps in domain understanding |

**Example flow**:

Event Storming discovered: "Order Approved" event, "Manager" actor, "Approval" policy

Example Mapping applies this:

- **Yellow card**: "As a manager, I want to approve large orders..."
- **Blue card**: "Orders over $10,000 must be approved by a manager"
- **Green card**: "Given an order of $15,000, when manager approves..."

**See**: [Example Mapping](./example-mapping.md) for detailed workshop process.

---

## From Shared Language to Executable Specifications

The Ubiquitous Language established through Event Storming and refined in Example Mapping flows directly into executable specifications.

### ATDD Uses the Ubiquitous Language

Acceptance Test-Driven Development (ATDD) acceptance criteria use the domain language in Gherkin `Rule:` blocks:

```gherkin
Feature: Order Approval Process

  As a manager
  I want to approve large orders
  So that we control high-value transactions

  Rule: Orders over $10,000 must be approved by a manager

  Rule: Approved orders can proceed to fulfillment
```

**Notice**: The specification uses "Order", "Approved", "Manager"—the same terms from domain discovery.

### BDD Uses the Ubiquitous Language

Behavior-Driven Development (BDD) scenarios use Given/When/Then with domain terms:

```gherkin
Feature: Order Approval Process

  Rule: Orders over $10,000 must be approved by a manager

    @success @ac1
    Scenario: Manager approves large order
      Given an order with amount $15,000
      And the order status is "Awaiting Approval"
      When the manager approves the order
      Then the order status should be "Approved"
      And the order should be eligible for fulfillment
```

**Notice**: Every term—"order", "manager", "approved", "awaiting approval"—comes from the Ubiquitous Language.

### The Complete Flow

```text
Event Storming: Discover domain vocabulary
      ↓
      Domain Events: "Order Approved"
      Actors: "Manager"
      Policies: "Large orders require approval"
      ↓
Example Mapping: Apply vocabulary to features
      ↓
      Blue card: "Orders over $10,000 must be approved by a manager"
      Green card: "Given $15,000 order, when manager approves..."
      ↓
Gherkin Specifications: Write using same terms
      ↓
      Rule: "Orders over $10,000 must be approved by a manager"
      Scenario: "When the manager approves the order"
      ↓
Code: Implement using same terms
      ↓
      class Order { ... }
      method approveOrder(Manager manager) { ... }
```

---

## Bounded Contexts: When Languages Diverge

Not every term means the same thing everywhere in a business. The Ubiquitous Language exists within **bounded contexts**.

### What is a Bounded Context?

> A **bounded context** is a defined boundary within which a particular model is applicable, ensuring clarity when different models might apply in different contexts.
>
> — Eric Evans, Domain-Driven Design

### When Terms Mean Different Things

**Example**: The word "Customer" might mean different things in different parts of a business:

- **Sales Context**: "Customer" = potential buyer, includes leads and prospects
- **Support Context**: "Customer" = paying subscriber with support contract
- **Accounting Context**: "Customer" = entity with payment history and invoices

**The solution**: Define bounded contexts and be explicit about which context applies.

### Handling Context Boundaries in Specifications

**Specify the context in the feature header**:

```gherkin
Feature: sales_lead-management

  # In this context, "Customer" means "potential buyer"
```

**Or be explicit in scenarios about context transitions**:

```gherkin
@success @ac1
Scenario: Convert sales prospect to paying customer
  Given a sales prospect with contact information
  When the prospect signs a contract
  Then the prospect becomes a Support customer
  And a Billing customer record is created
```

**Notice**: The specification is explicit about context transitions. "Sales prospect" becomes "Support customer" and "Billing customer".

### Why Bounded Contexts Matter

Bounded contexts prevent the mistake of forcing one language across incompatible domains:

- Acknowledges that different contexts need different models
- Prevents terminology conflicts and confusion
- Makes integration points explicit
- Allows each context to have its own Ubiquitous Language

---

## The Continuous Evolution of Language

Ubiquitous Language is **not static**. It evolves as the team's understanding of the domain deepens.

### How Language Evolves

**Discovery of new concepts**: Event Storming sessions reveal previously hidden domain concepts

**Refinement of existing terms**: Example Mapping clarifies ambiguous terminology

**Identification of boundaries**: Context conflicts surface, leading to bounded context definitions

**Elimination of technical jargon**: Domain terms replace programmer vocabulary

### When Language Changes

**Event Storming** identifies the change:

- Team realizes "Validated" actually means two different things
- Split into "Verified" (data is correct) and "Approved" (business accepts)

**Example Mapping** confirms the distinction:

- Blue cards use refined terminology
- Green cards show concrete examples of each

**Specifications** are refactored:

- Search and replace "Validated" with appropriate term
- Review scenarios to use correct term for context

**Code** is renamed:

- `Validated` state becomes `Verified` or `Approved`
- Method names updated: `Validate()` → `Verify()` or `Approve()`
- Refactoring tools ensure consistency

### Maintaining Glossaries

**In Event Storming**: Definition stickies capture term meanings

**In specifications**: Add definition sections when needed:

```gherkin
# Domain Glossary
#
# Order: A customer request to purchase products
# Manager: An employee with approval authority
# Approved: Business acceptance (different from "Verified" = data checked)
```

**In code**: Comments or documentation strings clarify domain concepts

**In team practices**: Regular domain discussion sessions to align understanding

---

## Why This Matters for Specifications

Building a shared language through DDD is fundamental to executable specifications because:

### Specifications Can't Drift from Business Intent

When specifications use the Ubiquitous Language, they **are** the business requirements. There's no separate "business document" that specifications try to reflect—the specifications use the exact vocabulary the business uses.

**Traditional approach**:

- Business writes requirements document with business terms
- Developers write specs with interpreted/translated terms
- Drift occurs between the two

**With Ubiquitous Language**:

- Specifications written using exact business vocabulary
- Business can validate specifications directly
- No drift possible—the same language throughout

### All Stakeholders Can Read and Validate Specifications

Specifications written in the Ubiquitous Language are readable by everyone:

- **Product owners** recognize the terms and can validate requirements
- **Domain experts** see their concepts reflected accurately
- **Developers** understand what to implement
- **QA** knows what to test
- **Auditors** can trace requirements to implementation

### Tests Become Living Documentation

When automated tests use the Ubiquitous Language:

- Tests document the domain in terms everyone understands
- Business rules are visible and verifiable
- New team members learn the domain from reading tests
- Documentation never drifts because it executes

### Code Implements Exact Business Concepts

When code uses the Ubiquitous Language:

- Classes, methods, and variables reflect domain concepts
- Code reviews can involve domain experts
- Refactoring preserves business concepts
- Technical debt is visible as linguistic drift

### Changes to Language Propagate Consistently

When domain understanding evolves and language changes:

- Changes identified in Event Storming sessions
- Example Mapping uses updated terminology
- Specifications refactored to use new terms
- Code renamed to match specifications
- Consistency maintained across all layers

---

## Next Steps

Ready to establish Ubiquitous Language in your project?

**See**: [How to Establish Ubiquitous Language](../../how-to-guides/specifications/establish-ubiquitous-language.md) for a comprehensive step-by-step guide that covers:

- Running Event Storming workshops to discover domain vocabulary
- Documenting the language in glossaries and bounded contexts
- Applying the language in Example Mapping sessions
- Writing specifications using exact domain terms
- Implementing code that reflects business concepts
- Evolving the language as understanding deepens

---

## Related Documentation

### For Understanding

- [Event Storming](./event-storming.md) - Domain discovery workshops
- [Example Mapping](./example-mapping.md) - Requirements discovery workshops
- [Three-Layer Testing Approach](./three-layer-approach.md) - How ATDD/BDD/TDD work together
- [ATDD and BDD with Gherkin](./atdd-bdd-with-gherkin.md) - Writing specifications

### For Doing

- [How to Establish Ubiquitous Language](../../how-to-guides/specifications/establish-ubiquitous-language.md) - Step-by-step guide
- [Create Feature Spec](../../how-to-guides/specifications/create-specifications.md) - Creating specifications

### For Reference

- [Gherkin Format](../../reference/specifications/gherkin-format.md) - Specification syntax
- [Domain-Driven Design](https://www.domainlanguage.com/ddd/) - Eric Evans' foundational work
- [Ubiquitous Language (Martin Fowler)](https://martinfowler.com/bliki/UbiquitousLanguage.html) - Concept explanation

---

## Key Takeaways

1. **Without shared language, specifications fail** - Terminology mismatches cause rework and bugs
2. **Ubiquitous Language creates rigorous shared vocabulary** - Based on the domain model, evolved through conversation
3. **Event Storming discovers the language** - Collaborative domain exploration surfaces terminology
4. **Example Mapping applies the language** - Features use established vocabulary
5. **Specifications preserve the language** - ATDD/BDD written in domain terms
6. **Code implements the language** - Classes and methods reflect domain concepts
7. **Bounded contexts handle divergence** - Different contexts can have different languages
8. **Language evolves continuously** - Understanding deepens, terminology refines
