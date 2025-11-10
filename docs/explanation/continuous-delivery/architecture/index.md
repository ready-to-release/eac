# Architecture

Infrastructure and organizational decisions that shape how you implement the Continuous Delivery Model. These articles cover environment architecture and repository organization patterns.

---

## [Environments](environments.md)

Understanding environment types and their role in the CD Model.

Environments provide the execution context for different CD Model stages. Each environment type has specific characteristics, access controls, and purposes.

**The Six Environment Types:**

**1. DevBox (Local Development)**:

- Developer's local workstation
- Stage 1 (Authoring) and Stage 2 (Pre-commit)
- Full access, no restrictions
- Fast feedback loop

**2. Build Agents (CI/CD Automation)**:

- Automated build and test execution
- Stages 3-4 (Merge Request, Commit)
- Ephemeral, isolated per build
- Parallel execution capability

**3. PLTE (Production-Like Test Environment)**:

- Mirrors production configuration
- Stages 5-6 (Acceptance, Extended Testing)
- Real infrastructure, test data
- Critical for L3 end-to-end tests

**4. Demo (Stakeholder Validation)**:

- Longer-lived than PLTE
- Stage 7 (Exploration)
- Accessible to non-technical stakeholders
- Manual testing and UAT

**5. Deploy Agents (Production Gateway)**:

- Segregated production access
- Stage 10 (Production Deployment)
- Limited, audited access
- Security boundary

**6. Production (Live Environment)**:

- Real user traffic
- Stages 11-12 (Live, Release Toggling)
- Monitoring and observability
- Highest security and access controls

**Topics covered:**

- Detailed characteristics of each environment type
- Access controls and security boundaries
- Stage mapping (which stages run in which environments)
- Infrastructure as Code integration
- Cost optimization strategies
- PLTE requirements and architecture

**Read this article to understand**: What environments you need, how they differ, and which stages execute in each environment.

---

## [Repository Patterns](repository-patterns.md)

How repository structure affects CD Model implementation.

Repository organization is a critical architectural decision that impacts developer workflow, deployment pipelines, and dependency management.

**Two Primary Patterns:**

**Monorepo (Single Repository):**

- All code in one repository
- Multiple deployable units in one trunk
- Shared dependencies and tooling
- Atomic cross-service changes possible

**Characteristics:**

- Simplified dependency management
- Easier to refactor across boundaries
- Single CI/CD pipeline configuration
- Requires tooling for selective testing

**Polyrepo (Multiple Repositories):**

- Each deployable unit in separate repository
- Independent release cadences
- Clear ownership boundaries
- Explicit dependency versioning

**Characteristics:**

- Strong isolation between services
- Independent deployment pipelines
- Explicit versioning between dependencies
- Coordination overhead for cross-service changes

**Topics covered:**

- Detailed comparison of monorepo vs polyrepo
- Benefits and tradeoffs of each approach
- Anti-patterns to avoid (monolith in monorepo, nano-repos)
- Best practices for module boundaries
- Impact on CD Model stages
- Versioning strategies by pattern
- When to use each pattern

**Read this article to understand**: How to structure your repositories to support trunk-based development and continuous delivery.

---

## Complete Architecture Overview

The complete architecture flows from code organization through production deployment across multiple layers:

### Layer 1 - Code Organization

**Trunk (Repository)** contains one or more **Deployable Units**:

- **Monorepo**: Multiple deployable units in single repository
- **Polyrepo**: One deployable unit per repository
- **Path filters** (monorepo) or **repository boundaries** (polyrepo) define unit boundaries

### Layer 2 - Build and Test

**DevBox** (local development):

- Executes Stages 1-2 (Authoring, Pre-commit)
- L0-L1 tests (unit tests, fast feedback)
- Developer-controlled environment

**Build Agents** (CI/CD automation):

- Executes Stages 3-4 (Merge Request, Commit)
- L0-L2 tests (unit, integration tests)
- Ephemeral, isolated per build
- No production access (Zone A)

### Layer 3 - Validation

**PLTE** (Production-Like Test Environment):

- Executes Stages 5-6 (Acceptance, Extended Testing)
- L3 end-to-end tests (vertical, with test doubles)
- Production-like infrastructure
- Ephemeral or short-lived

**Demo** (Stakeholder Environment):

- Executes Stage 7 (Exploration)
- L4 exploratory tests and UAT
- Longer-lived for stakeholder access

### Layer 4 - Deployment Gateway

**Deploy Agents** (Segregated Production Access):

- Executes Stage 10 (Production Deployment)
- Network boundary between development and production
- Zone C: Access to both artifact repos (Zone A) and production (Zone B)
- Comprehensive audit logging
- Production credentials stored in secure vaults

**Network Segregation:**

- **Zone A**: DevBox, Build Agents, PLTE, Demo (no production access)
- **Zone B**: Production (isolated from development)
- **Zone C**: Deploy Agents (gateway between A and B)

### Layer 5 - Production

**Production** (Live Environment):

- Executes Stages 11-12 (Live, Release Toggling)
- L4 tests (synthetic monitoring, production validation)
- Real user traffic
- Deployment strategies (blue-green, canary, rolling)
- Feature flags for release control

### Complete Flow

**Code Flow:**

1. Developer → DevBox (Stage 1-2)
2. Commit → Build Agents (Stage 3-4) → Artifacts created
3. Artifacts → PLTE (Stage 5-6) → Validation
4. Release candidate → Demo (Stage 7) → Stakeholder approval
5. Approved → Deploy Agents (Stage 10) → Retrieve artifacts
6. Deploy Agents → Production (Stage 11-12) → Live deployment

**Artifact Flow:**

1. Build Agents create immutable artifacts
2. Artifacts published to artifact repository
3. Deploy Agents retrieve artifacts
4. Deploy Agents deploy to Production
5. Production never pulls from development zones directly

**Network Boundaries:**

- Build Agents (Zone A) **cannot** access Production (Zone B)
- Deploy Agents (Zone C) **can** access both
- Production credentials **never** leave Zone C
- All production deployments audited via Deploy Agents

---

## Next Steps

- **Need environment details?** Read [Environments](environments.md)
- **Need repository guidance?** Read [Repository Patterns](repository-patterns.md)
- **Want to understand stages?** See [CD Model](../cd-model/index.md)
- **Ready for workflow?** Explore [Workflow](../workflow/index.md)

Return to **[Continuous Delivery Overview](../index.md)** for complete navigation.
