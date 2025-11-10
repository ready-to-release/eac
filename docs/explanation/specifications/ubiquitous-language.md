# Ubiquitous Language: Building Shared Vocabulary

> **How shared domain vocabulary creates the foundation for executable specifications**

---

## The Problem: Lost in Translation

Software teams suffer from a fundamental communication problem: **business stakeholders and technical teams speak different languages**.

**Common misalignments**: Business says "customer" → Developers think "user" | Product owner says "order" → Developers think "transaction" | Domain expert says "approved" → Developers think "validated"

**Consequences**: Rework (teams discover they meant different things after implementation), Bugs (code implements wrong concept), Failed validation (business can't validate technical jargon), Drift (specs use one set of terms, code uses another)

**Root cause**: Without shared language, every handoff is a translation introducing potential misalignment.

---

## The Solution: Ubiquitous Language (DDD)

Domain-Driven Design (DDD) provides **Ubiquitous Language**: a common, rigorous language between developers and users, based on the domain model used in software.

> Evans makes clear that using the ubiquitous language in conversations with domain experts is an important part of testing it, and hence the domain model. He also stresses that the language (and model) should evolve as the team's understanding of the domain grows.
> — Martin Fowler, [Ubiquitous Language](https://martinfowler.com/bliki/UbiquitousLanguage.html)

### Key Characteristics

- **Rigorous**: Precise - software doesn't handle ambiguity
- **Shared**: Everyone uses same terms with same meanings
- **Domain-based**: From business domain, not technical implementation
- **Evolving**: Refines as team's understanding deepens

### Why It Matters

- Conversations are productive (no translation required)
- Specifications are validatable (business can read and confirm)
- Code reflects intent (implementation uses exact business concepts)
- Tests document domain (automated tests as living documentation)

---

## Discovering Language Through Collaboration

Ubiquitous Language is **discovered** through collaborative domain exploration.

### Event Storming: Domain Discovery

Collaborative workshop technique that discovers domain vocabulary by visually mapping business events and processes.

**Outcomes**: Domain events, actors, commands, policies, bounded contexts, and the Ubiquitous Language itself.

**See**: [Event Storming](./event-storming.md) for workshop formats, facilitation, and best practices.

### Example Mapping: Requirements Discovery

Applies established Ubiquitous Language to specific features through time-boxed workshops.

**Uses domain terms in**:
- **Yellow cards**: WHO (actor), WHAT (capability), WHY (value)
- **Blue cards**: Business rules using domain vocabulary
- **Green cards**: Concrete situations using domain concepts
- **Red cards**: Gaps in domain understanding

**See**: [Example Mapping](./example-mapping.md) for detailed workshop process.

---

## From Shared Language to Executable Specifications

### The Flow

```
Event Storming: Discover vocabulary
    → Domain Events, Actors, Policies
Example Mapping: Apply to features
    → Rules and Examples in domain terms
Gherkin Specifications: Write using same terms
    → Feature, Rule, Scenario
Code: Implement using same terms
    → class Order, method approveOrder()
```

### Example

**Event Storming** discovered: "Order Approved" event, "Manager" actor, "Approval" policy

**Example Mapping** applied:
- Blue card: "Orders over $10,000 must be approved by a manager"
- Green card: "Given $15,000 order, when manager approves..."

**Gherkin Specification**:

```gherkin
Feature: Order Approval Process

  Rule: Orders over $10,000 must be approved by a manager

    @success @ac1
    Scenario: Manager approves large order
      Given an order with amount $15,000
      And the order status is "Awaiting Approval"
      When the manager approves the order
      Then the order status should be "Approved"
```

**Code**: Uses same terms - `Order`, `Manager`, `Approved`

---

## Bounded Contexts: When Languages Diverge

Not every term means the same thing everywhere. Ubiquitous Language exists within **bounded contexts**.

**Example**: "Customer" has different meanings:
- **Sales Context**: Potential buyer, includes leads
- **Support Context**: Paying subscriber with support contract
- **Accounting Context**: Entity with payment history

### Handling in Specifications

**Specify context in feature**:

```gherkin
Feature: sales_lead-management
  # In this context, "Customer" means "potential buyer"
```

**Or be explicit about transitions**:

```gherkin
Scenario: Convert sales prospect to paying customer
  Given a sales prospect with contact information
  When the prospect signs a contract
  Then the prospect becomes a Support customer
  And a Billing customer record is created
```

**Why it matters**: Prevents forcing one language across incompatible domains, makes integration points explicit, allows each context its own language.

---

## Continuous Language Evolution

Ubiquitous Language evolves as understanding deepens.

### How Language Changes

**Event Storming** identifies changes (e.g., "Validated" actually means "Verified" vs "Approved")
**Example Mapping** confirms distinctions with concrete examples
**Specifications** are refactored to use refined terms
**Code** is renamed to match (classes, methods, variables)

### Review Cadences

**Weekly**: Review new scenarios for consistency, identify terms needing clarification
**Monthly**: Check for language drift, consolidate synonyms, update older specs
**Quarterly**: Run Event Storming, major terminology refactoring

### Propagating Changes

When domain language evolves:

1. **Update glossary** - Document new/refined term
2. **Refactor specifications** - Use new language
3. **Update step definitions** - Match terminology
4. **Rename production code** - Align types, functions
5. **Update documentation** - Consistency everywhere

### Red Flags 🚩

- Same concept has multiple names in different specs
- Code uses different terminology than specifications
- New team members confused by terminology
- Stakeholders and developers use different words for same thing
- Glossary not updated >6 months

### Green Indicators ✅

- Consistent terms across all specifications
- Code identifiers match specification language
- Glossary reflects current understanding
- Team explains terms clearly to newcomers
- Stakeholders recognize terminology in specifications

**See**: [Review and Iterate](review-and-iterate.md) for ongoing specification maintenance practices.

---

## Why This Matters for Specifications

### Prevents Drift

Specifications use exact business vocabulary - no separate "business document" that drift can occur between. Business validates specifications directly.

### All Stakeholders Can Validate

Written in Ubiquitous Language means:
- Product owners recognize terms and validate requirements
- Domain experts see their concepts reflected
- Developers understand what to implement
- QA knows what to test
- Auditors trace requirements to implementation

### Tests as Living Documentation

Automated tests using Ubiquitous Language:
- Document domain in terms everyone understands
- Business rules visible and verifiable
- New team members learn domain from reading tests
- Documentation never drifts (it executes)

### Code Implements Business Concepts

Code using Ubiquitous Language:
- Classes, methods, variables reflect domain concepts
- Code reviews can involve domain experts
- Refactoring preserves business concepts
- Technical debt visible as linguistic drift

---

## Key Takeaways

1. **Without shared language, specifications fail** - Terminology mismatches cause rework and bugs
2. **Ubiquitous Language = rigorous shared vocabulary** - Based on domain model, evolved through conversation
3. **Event Storming discovers the language** - Collaborative domain exploration
4. **Example Mapping applies the language** - Features use established vocabulary
5. **Specifications preserve the language** - ATDD/BDD in domain terms
6. **Code implements the language** - Classes and methods reflect domain
7. **Bounded contexts handle divergence** - Different contexts have different languages
8. **Language evolves continuously** - Understanding deepens, terminology refines

---

## See Also

- [Event Storming](./event-storming.md) - Domain discovery workshops
- [Example Mapping](./example-mapping.md) - Requirements discovery
- [Review and Iterate](review-and-iterate.md) - How specifications evolve with language
- [ATDD and BDD with Gherkin](./atdd-bdd-with-gherkin.md) - Writing specifications
- [Three-Layer Approach](./three-layer-approach.md) - How ATDD/BDD/TDD work together
