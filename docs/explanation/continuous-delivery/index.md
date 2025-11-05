# Continuous Delivery

Understanding the Continuous Delivery Model, a comprehensive framework for delivering software from development through production with quality, traceability, and compliance built in.

**Topics covered:**

- Core concepts: Unit of Flow, Deployable Units, Trunk-Based Development, Branching
- 12-stage Continuous Delivery Model
- Implementation patterns (Release Approval vs Continuous Deployment)
- Environment architecture and Infrastructure as Code
- Repository organization patterns
- Testing strategy across all stages
- Security integration using open-source tools
- Workflow practices and branching strategies

---

## Core Concepts

Foundational concepts that underpin the Continuous Delivery Model.

### [Unit of Flow](core-concepts/unit-of-flow.md)

Understanding the four interconnected components that enable Continuous Delivery.

**Topics covered:**

- The four components: Trunk, Deployable Unit, Deployment Pipeline, Live
- Relationships between components
- Polyrepo vs monorepo patterns
- Integration with CD Model stages
- Common architectural patterns

### [Deployable Units](core-concepts/deployable-units.md)

Understanding what gets built, versioned, and deployed through the CD Model.

**Topics covered:**

- Definition and characteristics of deployable units
- Types: Runtime Systems vs Versioned Components
- Versioning strategies (Implicit, CalVer, Release Number, SemVer, API)
- Immutable artifacts and why they matter
- Dependency management (internal and external)
- Choosing the right granularity and boundaries
- Integration with CD Model stages

---

## CD Model

The complete 12-stage framework for delivering software with quality, traceability, and compliance.

### [Overview & Visual Notation](cd-model/cd-model-overview.md)

Introduction to the 12-stage Continuous Delivery Model and how to read the visual diagrams.

**Topics covered:**

- What is the CD Model and why it matters
- Traditional vs CD Model approaches
- Visual notation and legend explanations
- Overview of all 12 stages
- Key principles (shift-left, fail fast, automation, traceability)

### [Stages 1-6: Development to Testing](cd-model/cd-model-stages-1-6.md)

Detailed explanation of development and testing stages from code authoring through extended validation.

**Topics covered:**

- Stage 1: Authoring Changes with Requirements as Code
- Stage 2: Pre-commit validation (5-10 min time-box)
- Stage 3: Merge Request with peer review
- Stage 4: Commit with integration testing
- Stage 5: Acceptance Testing in PLTE (IV, OV, PV)
- Stage 6: Extended Testing (performance, security, compliance)

### [Stages 7-12: Release to Production](cd-model/cd-model-stages-7-12.md)

Detailed explanation of release and production stages from stakeholder validation through live monitoring.

**Topics covered:**

- Stage 7: Exploration with stakeholder validation and UAT
- Stage 8: Start Release with version tagging
- Stage 9: Release Approval (manual vs automated)
- Stage 10: Production Deployment (strategies and rollback)
- Stage 11: Live monitoring with phased rollout
- Stage 12: Release Toggling with feature flags

### [Implementation Patterns](cd-model/implementation-patterns.md)

Understanding when and how to use Release Approval (RA) vs Continuous Deployment (CDE) patterns.

**Topics covered:**

- When to use RA pattern (regulated, high-risk systems)
- When to use CDE pattern (non-regulated, fast iteration)
- Manual vs automated approval gates
- Traceability and audit evidence
- Compliance and signoffs
- Pattern selection decision tree

---

## Workflow Practices

Day-to-day development practices that enable Continuous Integration and Continuous Delivery.

### [Trunk-Based Development](workflow/trunk-based-development.md)

Comprehensive guide to trunk-based development practices enabling Continuous Integration and Continuous Delivery.

**Topics covered:**

- Core principles: Single source of truth, short-lived branches, small changes, continuous integration
- Branch types: Trunk, topic branches, release branches
- Commits and squash merging
- Daily development flow (7 steps)
- Feature hiding strategies (code-level, configuration, feature flags)
- Release flows for RA and CDE patterns
- Cherry-picking fixes between branches
- Best practices and anti-patterns
- Emergency fixes and conflict resolution

### [Branching Strategies](workflow/branching-strategies.md)

Detailed branching flows for Release Approval and Continuous Deployment patterns.

**Topics covered:**

- Release Approval (RA) pattern: Branching flow with release branches
- Continuous Deployment (CDE) pattern: Direct deployment from trunk
- Stage-by-stage flow for each pattern
- Release branch lifecycle and management
- Fixing bugs on release branches vs trunk
- Pipeline integration and separation
- Pin and stitch dependency management
- Timeline and best practices comparison
- When to use each pattern

---

## Testing

Comprehensive testing strategy integrated throughout all CD Model stages.

### [Testing Strategy Overview](testing/testing-strategy-overview.md)

Understanding test taxonomy and the shift-left testing approach.

**Topics covered:**

- Five test levels (L0-L4) explained
- L0: Unit tests for logic validation
- L1: Component integration with mocking
- L2: Integration tests with real dependencies
- L3: End-to-end tests in PLTE
- L4: Exploratory and UAT
- Hybrid E2E approach
- Shift-left strategy and benefits

### [Testing Integration with CD Model](testing/testing-strategy-integration.md)

How test levels integrate with CD Model stages.

**Topics covered:**

- Test level environment mapping
- Process isolation strategies (L0/L1, L2, L3)
- Test levels by CD Model stage
- Time-boxing per stage
- Integration with ATDD/BDD/TDD
- Test pyramid in practice

---

## Architecture

Infrastructure and organizational decisions that shape CD Model implementation.

### [Environments](architecture/environments.md)

Understanding environment types and their role in the CD Model.

**Topics covered:**

- DevBox for local development
- Build Agents for CI/CD automation
- Production-Like Test Environments (PLTE)
- Demo environments for stakeholder validation
- Deploy Agents with segregated production access
- Production environment architecture
- Infrastructure as Code integration

### [Repository Patterns](architecture/repository-patterns.md)

How repository structure affects CD Model implementation.

**Topics covered:**

- Monorepo pattern (characteristics, benefits, tradeoffs)
- Polyrepo pattern (characteristics, benefits, tradeoffs)
- Anti-patterns to avoid
- Best practices for module boundaries and versioning
- Impact on CD Model stages

---

## Security

Security integration throughout all stages using open-source tools.

### [Security in the CD Model](security/security.md)

Security integration throughout all stages using open-source tools.

**Topics covered:**

- SAST, DAST, dependency scanning, container security
- OWASP ZAP for dynamic testing
- Trivy for multi-purpose scanning
- Dependabot for dependency management
- Security by stage matrix
- Shift-left security practices
- Blocking strategies and remediation workflow

---

## See Also

### [Everything as Code](../everything-as-code/index.md)

Continuous Delivery in regulated industries requires "Everything as Code" principles. Understanding the problem space, Cynefin framework, and DORA metrics helps contextualize why continuous delivery practices work for compliance.

---

**Looking for technical details?** See [Continuous Delivery Reference](../../reference/continuous-delivery/index.md)
