# Initiation

The initiation process evaluates whether good ideas are truly great ideas using **Design Thinking**, where user researchers and designers collaborate iteratively to assess three key criteria:

* **Feasibility:** Can it be developed with available technology, skills, and resources?
* **Desirability:** Do users genuinely need and want this solution?
* **Viability:** Can it deliver sustainable value or return on investment?

## Key Activities

* Create Git repository
* Register system in relevant tracking systems
* Develop initial documentation: design overview, intended use, implementation plan, specifications, and risk assessment

---

## Design

### Architecture Diagrams

We recommend the [C4 Model](https://c4model.com/) for architecture visualization, maintained as code using [Structurizr](https://structurizr.com/):

* **Level 1 - System Context:** How the system fits in its environment
* **Level 2 - Container Diagram:** High-level technology choices
* **Level 3 - Component Diagram:** Components within each container
* **Level 4 - Code Diagram:** Detailed implementation (optional)

### Infrastructure and Tooling

Document in markdown format:

* Configured infrastructure
* Tools and technologies used

**Regulatory Note:** In regulated industries (medical devices, pharmaceuticals, finance), infrastructure and tooling often require validation/verification to demonstrate they reliably produce consistent, compliant results and don't introduce unintended risks into the final product.

---

## Decision Records

Document significant architectural and design choices that impact the system. Decision records capture the context, options considered, and rationale behind key design decisions, ensuring knowledge is preserved and can be revisited as the system evolves.

Templates: [Decision Record](../../../templates/design/dr.md)

---

## Threat Modeling

Threat modeling identifies security risks early in the design phase, enabling teams to build security into the architecture rather than retrofitting it later.

### Integration into Development

* **When:** During initial design and whenever architecture changes significantly (new features, integrations, data flows)
* **How Often:** At initiation, before major releases, and during design reviews
* **Design Impact:** Influences component boundaries, authentication/authorization patterns, data protection strategies, and infrastructure choices

### Tools

* [Microsoft Threat Modeling Tool](https://www.microsoft.com/en-us/securityengineering/sdl/threatmodeling)
* [OWASP Threat Dragon](https://owasp.org/www-project-threat-dragon/)
* [OWASP Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Threat_Modeling_Cheat_Sheet.html)

---

## Intended Use

**For regulatory environments only** (medical devices, pharmaceuticals, finance, etc.)

The **intended use** is a regulatory requirement that formally defines the purpose, target users, and operational context of the software. It establishes the scope for validation, risk management, and regulatory submissions.

**Why it matters:**

* Defines regulatory classification and compliance obligations
* Sets boundaries for risk assessment and validation activities
* Provides basis for regulatory approval and audit documentation
* Ensures software is used only for its approved purpose

**Specifications** then translate the intended use into specific, testable, and compliant requirements that guide development, validation, and regulatory adherence.

Templates: [Intended Use Template](../../../templates/intended-use.md)
