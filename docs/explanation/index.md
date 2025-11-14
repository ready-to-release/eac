# Explanation

**Understanding-oriented**: In-depth explanations of concepts, architecture, and design decisions.

---

## Getting Started

- **New to this approach?** Start with [Everything as Code](#everything-as-code) to understand the foundational principles
- **Want to understand the practices?** Read [Specifications Concepts](#specifications-concepts) to learn ATDD/BDD/TDD
- **Need the big picture?** See [Software Development Lifecycle](#software-development-lifecycle) for the complete lifecycle
- **Looking for delivery details?** Explore [Continuous Delivery](#continuous-delivery) for the 12-stage CD model
- **In a regulated environment?** Also review [Compliance Transformation](#compliance-transformation) for optimization strategies

---

## Available Explanations

### [Everything as Code](everything-as-code/index.md)

Understanding the problem space in regulated industries and why "Everything as Code" principles are essential for modern software delivery while maintaining compliance.

**Topics covered:**

- The compliance-velocity paradox facing regulated industries
- Understanding the problem domain through Cynefin framework
- What "Everything as Code" means and why it matters
- Single timeline for traceability
- Collaboration through shared language and executable specifications
- Measuring and improving flow with DORA metrics and Value Stream Mapping
- Why automation tooling is essential

---

### [Specifications Concepts](specifications/index.md)

Understanding the three-layer testing approach and test-driven development practices. These concepts are foundational to how requirements are defined and tested throughout the [Software Development Lifecycle](#software-development-lifecycle) and [Continuous Delivery](#continuous-delivery) model.

**Topics covered:**

- Three-layer testing approach (ATDD, BDD, TDD)
- ATDD concepts and Example Mapping workshops
- BDD concepts and Gherkin language
- How the layers interact and when to use each
- Living documentation and traceability
- Ubiquitous Language and Domain-Driven Design
- Event Storming workshops
- Risk controls for compliance

---

### [Software Development Lifecycle](lifecycle/index.md)

Understanding the complete software lifecycle from initiation through operations and end-of-life, balancing regulatory compliance with continuous delivery in a DevOps model. The Development phase implements the [Continuous Delivery](#continuous-delivery) model, and all phases use [Specifications](#specifications-concepts) for requirements definition.

**Topics covered:**

- **Lifecycle Overview**: DevOps approach eliminating handovers between development and operations
- **Initiation**: Feasibility assessment (feasibility, desirability, viability), design documentation (C4 model), decision records, threat modeling, intended use (regulatory)
- **Development**: Implementation planning, specifications, risk controls, implementation reports, testing strategy, automated documentation
- **Operations**: Maintenance plans, periodic evaluation, user access reviews, defect and incident management, blameless post-mortems
- **End of Life**: Decommissioning and knowledge transfer

---

### [Continuous Delivery](continuous-delivery/index.md)

Understanding the 12-stage Continuous Delivery Model - a comprehensive framework for delivering software from development through production with built-in quality, traceability, and compliance. This model is the core implementation mechanism for the Development phase in the [Software Development Lifecycle](#software-development-lifecycle), using [Specifications](#specifications-concepts) for testing.

**Topics covered:**

- **CD Model**: 12-stage framework with visual notation, stages 1-6 (development/testing), stages 7-12 (release/production)
- **Implementation Patterns**: Release Approval (RA) for regulated systems, Continuous Deployment (CDE) for fast iteration
- **Architecture**: Environments (DevBox, Build Agents, PLTE, Demo, Deploy Agents, Production), Repository patterns (monorepo vs polyrepo)
- **Testing Strategy**: Five test levels (L0-L4), shift-left approach, integration with CD Model stages, ATDD/BDD/TDD alignment
- **Security**: SAST, DAST, dependency scanning, container security using open-source tools (OWASP ZAP, Trivy, Dependabot)
- **Workflow**: Trunk-based development practices

---

### [Compliance Transformation](transformation/index.md)

Understanding how to transform compliance from a blocking activity into a continuous, automated capability that enables faster, safer software delivery. Builds on [Everything as Code](#everything-as-code) principles and integrates with the [Continuous Delivery](#continuous-delivery) model using [Specifications](#specifications-concepts) for risk controls.

**Topics covered:**

- Why traditional compliance fails and the opportunity for transformation
- Compliance-as-code principles (everything-as-code, continuous validation, shift-left, automated evidence)
- Four-phase transformation framework (Assessment, Pilot, Automation, Rollout)
- Risk control specifications in Gherkin format
- Evidence automation architecture
- Shift-left compliance strategy
- Success factors and common pitfalls
- ROI analysis and readiness assessment

---

## What is Explanation?

Explanation documentation is **understanding-oriented** content that clarifies and illuminates topics. It provides background, context, and discussion of alternatives.

**Characteristics:**

- Discusses concepts and ideas
- Provides context and background
- Explains design decisions
- Compares alternative approaches
- Deepens understanding

**Not finding what you need?**

- **Just getting started?** See [Tutorials](../tutorials/index.md)
- **Need to solve a problem?** See [How-to Guides](../how-to-guides/index.md)
- **Looking for technical details?** See [Reference](../reference/index.md)
