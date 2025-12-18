# Event Storming Overview
>
> **Collaborative technique for discovering Ubiquitous Language through domain exploration**

---

## What is Event Storming?

> Event Storming is a group of collaborative modeling techniques that help teams understand complex domains by visually mapping out key events. It is designed to help people from different parts of an organization create a shared understanding of the problem space they are going to work with.

Event Storming uses a game-like format with rules, a "board" (brown packing paper), and grammar (colored sticky notes). This collaborative approach surfaces the domain vocabulary through structured conversation.

---

## Why Event Storming Matters for Specifications

Event Storming creates the foundation for all specifications work:

- **Discovers domain vocabulary** that flows into Example Mapping and Gherkin
- **Surfaces terminology conflicts** before they become specification problems
- **Identifies bounded contexts** where different languages apply
- **Reveals business processes** that need to be specified
- **Builds shared understanding** across business and technical teams

---

## How Event Storming Helps

### Discovers Domain Vocabulary

Ubiquitous Language doesn't emerge spontaneously—it's **discovered** through collaborative domain exploration. Event Storming is a powerful technique for this discovery.

When domain experts explain processes, they use their natural vocabulary. When developers ask clarifying questions, misunderstandings emerge and get resolved.

### Surfaces Critical Information

Event Storming workshops surface critical information for specifications:

**Domain Events**:

- Things that happen in the business
- Examples: "Order Placed", "Payment Received", "Approval Granted"
- Become the foundation for Given/When/Then scenarios

**Actors/Personas**:

- Who initiates or responds to events
- Examples: "Customer", "Manager", "System Administrator"
- Appear in user stories ("As a customer...")

**Commands**:

- Actions that cause events
- Examples: "Place Order", "Approve Request", "Cancel Subscription"
- Become the "When" in scenarios

**Policies**:

- Business rules that govern behavior
- Examples: "Orders over $10,000 require manager approval"
- Become acceptance criteria (Rule blocks)

**Bounded Contexts**:

- Where different terminologies apply
- Examples: "Customer" in Sales context vs Support context
- Prevents terminology confusion in specifications

---

## From Event Storming to Specifications

### Event Storming → Example Mapping

Event Storming provides the vocabulary for Example Mapping workshops:

**Event Storming discovered**:

- Domain event: "Order Approved"
- Actor: "Manager"
- Policy: "Large orders require approval"

**Example Mapping applies this**:

- Yellow card: "As a manager, I want to approve large orders..."
- Blue card: "Orders over $10,000 must be approved by a manager"
- Green card: "Given an order of $15,000, when manager approves..."

**See**: [Example Mapping](./example-mapping.md) for detailed workshop process.

### Event Storming → Gherkin Specifications

Event Storming vocabulary flows directly into Gherkin:

**Event Storming discovered**:

- Domain terms: "Order", "Manager", "Approved"
- Process flow: Place order → Check amount → Require approval → Manager approves

**Gherkin specification**:

```gherkin
Feature: Order Approval Process

  Rule: Large orders require manager approval

    @ov
    Scenario: Manager approves large order
      Given an order with amount $15,000
      And the order status is "Awaiting Approval"
      When the manager approves the order
      Then the order status should be "Approved"
      And the order should be eligible for fulfillment
```

**Notice**: Every term comes from Event Storming—"order", "manager", "approved", "awaiting approval".

---

## Maintaining Glossaries

**During Event Storming**: Use white definition stickies to capture term meanings

**After Event Storming**: Create a glossary document

**Example glossary format**:

```markdown
# Domain Glossary: Order Management Context

**Order**: A customer request to purchase products
- Status values: "Pending", "Awaiting Approval", "Approved", "Fulfilled"
- Bounded context: Order Management

**Manager**: An employee with approval authority
- Can: Approve orders, reject orders, request more information
- Bounded context: Order Management

**Approved**: Business acceptance of an order
- Trigger: Manager approval action
- Result: Order eligible for fulfillment
- NOT the same as "Verified" (data correctness)
```

---

## Further Reading

### Books

- [Introduction to Event Storming](https://leanpub.com/introducing_eventstorming) - Alberto Brandolini's book on how to do Event Storming
- [Domain-Driven Design](https://www.domainlanguage.com/ddd/) - Eric Evans' foundational work

### Online Resources

- [eventstorming.com](https://www.eventstorming.com) - A site full of resources
- [Awesome EventStorming](https://github.com/mariuszgil/awesome-eventstorming) - A curated list of material and links
- [Alberto Brandolini's Blog](https://blog.avanscoperta.it/author/ziobrando/) - Original creator's insights

---

## Related Documentation

- [Event Storming Formats](./event-storming-formats.md) - The three Event Storming formats
- [Event Storming Facilitation](./event-storming-facilitation.md) - How to run workshops
- [Ubiquitous Language](../concepts/ubiquitous-language.md) - DDD foundation for shared vocabulary
- [Example Mapping](./example-mapping.md) - Requirements discovery using Event Storming vocabulary
