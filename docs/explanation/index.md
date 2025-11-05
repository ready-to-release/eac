# Explanation

**Understanding-oriented**: In-depth explanations of concepts, architecture, and design decisions.

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

### [Continuous Delivery](continuous-delivery/index.md)

Understanding the 12-stage Continuous Delivery Model - a comprehensive framework for delivering software from development through production with built-in quality, traceability, and compliance.

**Topics covered:**

- **CD Model**: 12-stage framework with visual notation, stages 1-6 (development/testing), stages 7-12 (release/production)
- **Implementation Patterns**: Release Approval (RA) for regulated systems, Continuous Deployment (CDE) for fast iteration
- **Architecture**: Environments (DevBox, Build Agents, PLTE, Demo, Deploy Agents, Production), Repository patterns (monorepo vs polyrepo)
- **Testing Strategy**: Five test levels (L0-L4), shift-left approach, integration with CD Model stages, ATDD/BDD/TDD alignment
- **Security**: SAST, DAST, dependency scanning, container security using open-source tools (OWASP ZAP, Trivy, Dependabot)
- **Workflow**: Trunk-based development practices

---

### [Testing](testing/index.md)

Understanding the three-layer testing approach and test-driven development practices.

**Topics covered:**

- Three-layer testing approach (ATDD, BDD, TDD)
- ATDD concepts and Example Mapping workshops
- BDD concepts and Gherkin language
- How the layers interact and when to use each
- Living documentation and traceability

---

### [VS Code Extension Architecture](vscode-extension-architecture.md)

Understanding how the VSCode extension integrates with Model Context Protocol (MCP) servers.

**Topics covered:**

- Client-server architecture with MCP
- JSON-RPC communication protocol
- Extension lifecycle and components
- Why MCP? Design benefits and trade-offs

---

### [Commit Agent Pipeline Architecture](commit-agent-pipeline.md)

Understanding the 5-agent system that generates semantic commit messages.

**Topics covered:**

- Architecture flow and agent interaction
- Why 5 agents? Design rationale
- Model selection (why Haiku)
- Git state validation
- Performance characteristics
- Security considerations
- Design trade-offs

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
