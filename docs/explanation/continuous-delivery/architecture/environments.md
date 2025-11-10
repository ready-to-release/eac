# Environments Architecture

## Introduction

Environments are the foundation of the Continuous Delivery Model. Each environment serves a specific purpose in the software delivery pipeline, from local development through production deployment. Understanding environment types, their characteristics, and their relationships is essential for implementing an effective CD Model.

Traditional approaches often rely on long-lived, shared environments that created inter-team bottlenecks and inconsistencies. The CD Model reimagines environments as purpose-built, often ephemeral resources that enable parallel execution, rapid feedback, and consistent infrastructure.

Doing this, we are "controlling the variables" of our verification processes, making our verifications more deterministic and thereby trustworthy.

### Traditional vs Modern Approach

**Traditional Environment Model:**

Traditional software delivery relies on a linear progression through fixed, long-lived environments:

- **Development** → **Test** → **Validation** → **Production**
- Shared environments create resource contention
- Configuration drift between environments
- Manual environment setup and maintenance
- Bottlenecks when multiple teams share resources
- Environment-specific bugs discovered late

**CD Model Environment Approach:**

The CD Model uses purpose-built environments that match specific validation needs:

- **DevBox** for local development
- **Build Agents** for CI/CD automation
- **PLTE** (Production-Like Test Environments) for realistic validation
- **Demo** environments for stakeholder feedback
- **Deploy Agents** with segregated production access
- **Production** with phased rollout capabilities

Key improvements:

- Infrastructure as Code ensures consistency
- Ephemeral environments eliminate drift
- Parallel execution without conflicts
- Purpose-built for specific validation
- Production-like characteristics from early stages

---

## Environment Types Explained

### DevBox

The DevBox is the developer's local development environment where all changes begin.

**Characteristics:**

- Full control and isolation
- Fast iteration without network dependencies
- Immediate feedback loops
- No impact on other developers

**Tools and Resources:**

- IDE or code editor
- Local build tools and compilers
- Unit testing frameworks
- Local security scanners (Trivy)
- Version control (Git)
- Container runtime (Docker)

**Purpose in CD Model:**

- Stage 1 (Authoring): Create and develop changes
- Stage 2 (Pre-commit): Fast local validation

**Best Practices:**

- Mirror production configuration where possible
- Run pre-commit checks locally
- Use containerization for consistency
- Maintain clean, reproducible setup

### Build Agents

Build Agents are dedicated CI/CD pipeline runners that provide consistent, reproducible build environments.

**Characteristics:**

- Isolated execution for each build
- Consistent configuration across runs
- Access to artifact repositories
- Network access to test infrastructure
- No production credentials

**Purpose in CD Model:**

- Stage 2 (Pre-commit): CI/CD validation
- Stage 3 (Merge Request): Peer review automation
- Stage 4 (Commit): Integration testing
- Stage 8 (Start Release): Release candidate creation

**Security Considerations:**

- Isolated from production networks
- Limited credentials (read artifact repos, write test results)
- No direct deployment capabilities
- Audit logging for all actions

**Infrastructure:**

- Containerized runners (Docker, Kubernetes or vendor PasS)
- Ephemeral execution environments
- Infrastructure as Code for consistency

### Production-Like Test Environments (PLTE)

PLTEs are ephemeral, isolated environments that emulate production characteristics for realistic testing.

![PLTE Legend](../../../assets/cd-model/legend-env-plte.drawio.png)

**Legend for PLTE notation:** Shows the symbol used in CD Model diagrams to represent Production-Like Test Environments - ephemeral, isolated environments for vertical testing (L3) with test doubles for all external dependencies.

**Characteristics:**

- Production-like infrastructure (OS, database versions, network topology)
- Production-like configuration (without production credentials)
- Realistic test data (anonymized if necessary)
- Isolated per feature branch or release candidate
- Ephemeral - created on-demand, destroyed after testing
- Can come as a series of environments with different toggle settings

**Purpose in CD Model:**

- Stage 5 (Acceptance Testing): Functional validation (IV, OV, PV)
- Stage 6 (Extended Testing): Performance, security, compliance

**Benefits:**

- Realistic testing without production risk
- No resource contention between features
- Catch environment-specific issues early
- Parallel testing for multiple branches
- Infrastructure as Code validation

**Implementation:**

- Infrastructure as Code (Terraform, CloudFormation, Bicep etc.)
- Automated provisioning and teardown
- Database snapshots or seed data
- Network isolation for security

**Cost Management:**

- Short-lived (hours, not days)
- Automated cleanup after testing
- Resource limits to prevent overprovisioning
- Cloud-native auto-scaling

### Demo Environment

The Demo (or "Trunk Demo") environment provides a stable, production-like environment for stakeholder validation and exploratory testing.

**Characteristics:**

- Reflects current state of main branch
- Longer-lived than PLTEs (days to weeks)
- Accessible to non-technical stakeholders
- Production-like without production data
- Represent "next release" features
- Can come as a series of demo environments with different toggle settings

**Purpose in CD Model:**

- Stage 7 (Exploration): Stakeholder validation, UAT, exploratory testing

**Use Cases:**

- Feature demonstrations to product owners
- User acceptance testing
- Documentation and training preparation
- Exploratory testing by QA teams
- Stakeholder feedback collection

**Access:**

- Product owners and stakeholders
- QA teams
- Documentation teams
- Support and training teams

**Update Cadence:**

- Typically updated from main branch after successful Stage 6
- May be updated daily or weekly
- Represents validated, release-ready features

### Deploy Agents

Deploy Agents are specialized CI/CD runners with segregated access to production networks and deployment credentials.

![Environment Agent Legend](../../../assets/cd-model/legend-env-agent.drawio.png)

**Legend for agent types:** Shows the symbols for Build Agents (no production access, run Stages 2-4) and Deploy Agents (segregated production access, run Stage 10). The diagram illustrates network boundaries and credential segregation between agent types.

**Characteristics:**

- Network access to production environments
- Production deployment credentials (stored securely in vaults)
- Strict access controls and audit logging
- Principle of least privilege
- Separate from Build Agents

**Purpose in CD Model:**

- Stage 10 (Production Deployment): Execute production deployments
- Stage 11 (Live): Health check validation
- Rollback execution if needed

**Security Measures:**

- Network segmentation from Build Agents
- Multi-factor authentication for credentials
- Comprehensive audit logging
- Time-limited sessions
- Change approval integration

**Approval Integration:**

- Manual approval gate (RA pattern)
- Automated approval gate (CDe pattern)
- Emergency break-glass procedures
- Rollback triggers

**Why Separate from Build Agents:**

- Principle of least privilege
- Reduce attack surface
- Prevent unauthorized production access
- Clear audit trail

### Production Environment

The Production environment is where software serves end users and delivers business value.

**Characteristics:**

- Live user traffic
- Real business data
- High availability requirements
- Performance monitoring
- Incident response procedures

**Purpose in CD Model:**

- Stage 10 (Production Deployment): Receive new releases
- Stage 11 (Live): Operational monitoring
- Stage 12 (Release Toggling): Feature flag management

**Deployment Strategies:**

- Hot deploy (in-place updates)
- Staged deploy (rolling updates)
- Blue-green deployment
- Canary deployment

**Monitoring:**

- Application performance metrics
- Business metrics (conversion, revenue)
- Error rates and types
- Resource utilization
- User behavior

**Rollback Capabilities:**

- Automated rollback on threshold breaches
- Manual rollback procedures
- Database rollback considerations
- Feature flag kill switches

---

## Architecture Visuals Explained

### Architectural Layers

![Architectural Layers](../../../assets/environment/layers.drawio.png)

**This diagram shows the architectural organization layers for environment infrastructure:** The diagram illustrates how environments are structured using **categories** (Production vs Dev/Test), **category instances** (individual subscriptions or account structures), **templates** (Infrastructure as Code definitions), and **environment instances** (deployed environments from templates). The layered architecture shows how **shared infrastructure** supports multiple **environment slot groups** (named horizontal environments like DEVELOPMENT, DEMO, PRODUCTION), which in turn contain **environment slots** (individual environment instances). This hierarchical organization enables consistent infrastructure definitions across all environments while allowing for appropriate isolation and access controls at each layer.

### Environment Slot Groups and Slots

![Environment Slots](../../../assets/environment/slots.drawio.png)

**This diagram shows environment slot groups and slots organization:** The diagram illustrates how environments are organized into **slot groups** and **slots**. An **environment slot group** is a named horizontal environment grouping (e.g., DEVELOPMENT, DEMO, ACCEPTANCE, PRODUCTION) used to organize related environments. Within each slot group, **environment slots** are logical constructs that map to infrastructure templates. Horizontal PLTEs are instantiated within a single slot group and can consist of one to many slots. Vertical isolated PLTEs are also instantiated within slot groups. Slots can be empty or filled with environment instances, and slot groups can be partially or completely filled. This organization enables teams to manage multiple environment instances with clear boundaries for horizontal end-to-end testing and vertical isolated testing.

### Environment Slot and Slot Group Naming

![Environment Naming](../../../assets/environment/units.drawio.png)

**This diagram shows how environment slots and slot groups are identified through naming conventions:** The diagram illustrates how infrastructure components are named to indicate their **slot group** and **slot** membership. In cloud providers like Azure, a **slot** and **slot group** are identified by specific parts of the infrastructure component naming. For example, infrastructure components that exist in slot groups include App Services, Function Apps, Databases, Key Vaults, and Storage Accounts. **Shared infrastructure** (App Plans, Networks, DNS, Gateways, SQL Servers, Container Registries) has different naming patterns as it supports multiple slot groups. The naming convention enables clear identification of which environment instance a resource belongs to, facilitating automated provisioning and lifecycle management through Infrastructure as Code.

---

## Traditional vs CD Model Comparison

### Traditional Model: Dev → Test → Val → Prod

**Development Environment:**

- Shared by multiple developers
- Frequent conflicts and contention
- Configuration often differs from production
- Manual setup and maintenance

**Test Environment:**

- Shared by QA team
- Test results affected by concurrent testing
- Environment state inconsistent
- Configuration drift from production

**Validation Environment:**

- Pre-production validation
- Limited capacity creates bottleneck
- Often differs from production
- Manual approvals delay releases

**Production Environment:**

- Live user traffic
- Issues discovered late
- Risky deployments due to environment differences

**Problems with Traditional Approach:**

- Environment drift causes late discovery of bugs
- Shared resources create bottlenecks
- Manual environment management is error-prone
- Configuration differences hide issues until production
- Long feedback loops (days to weeks)

### CD Model Approach

**DevBox (Local):**

- Isolated, developer-controlled
- Consistent through containerization
- Fast feedback (seconds to minutes)
- No resource contention

**Build Agents (Ephemeral):**

- Consistent, reproducible
- Parallel execution
- Isolated per build
- Infrastructure as Code

**PLTE (On-Demand):**

- Production-like from Stage 5
- Isolated per feature/release
- Catch issues early
- Ephemeral, no drift

**Production (Controlled):**

- Phased rollout (canary, rings)
- Feature flags for control
- Automated monitoring and rollback
- High confidence from earlier validation

**Benefits of CD Model Approach:**

- Consistency eliminates environment-specific bugs
- Parallel execution removes bottlenecks
- Infrastructure as Code prevents drift
- Early production-like validation reduces risk
- Fast feedback loops (minutes to hours)
- High confidence in production deployments

### Migration Path

**Moving from Traditional to CD Model:**

1. **Infrastructure as Code**: Define environments as code for consistency
2. **Ephemeral PLTEs**: Implement on-demand PLTE creation and automated testing
3. **Agent Segregation**: Separate Build and Deploy Agents with network isolation, if required
4. **Production Readiness**: Add deployment strategies, monitoring, and automated rollback

---

## Infrastructure as Code Integration

Infrastructure as Code (IaC) ensures all environments are created from the same definitions, providing consistency, version control, and automated provisioning. Use Terraform/CloudFormation/Bicep for cloud infrastructure, Docker for packaging, docker compose for emulation and Kubernetes or vendor PaaS for PLTE and production orchestration.

**Ephemeral PLTE Lifecycle:**

1. **Trigger**: Merge to main or release candidate creation
2. **Provision**: Create from IaC (5-10 min)
3. **Deploy**: Install application and seed test data
4. **Test**: Run acceptance/extended tests (1-4 hours)
5. **Destroy**: Tear down infrastructure

Benefits: No configuration drift, parallel testing without conflicts, cost-effective (pay only when used).

---

## Summary

Environments in the CD Model are purpose-built, often ephemeral resources that enable rapid, parallel validation with production-like characteristics:

- **DevBox**: Local development with fast feedback
- **Build Agents**: Consistent CI/CD automation
- **PLTE**: Ephemeral, production-like testing
- **Demo**: Stakeholder validation and exploration
- **Deploy Agents**: Segregated production access
- **Production**: Monitored, controlled deployment

The CD Model eliminates traditional environment bottlenecks through:

- Infrastructure as Code for consistency
- Ephemeral environments to prevent drift
- Network segregation for security
- Purpose-built environments for specific validation

This architecture enables fast feedback, high quality, and confident production deployments.

## Next Steps

- [CD Model Overview](../cd-model/cd-model-overview.md) - Understand the 12 stages
- [Stages 1-6](../cd-model/cd-model-stages-1-6.md) - See how environments support development stages
- [Stages 7-12](../cd-model/cd-model-stages-7-12.md) - See how environments support release stages
- [Repository Patterns](repository-patterns.md) - Understand repository organization
- [Implementation Patterns](../cd-model/implementation-patterns.md) - Choose RA or CDE pattern

## References

- [CD Model Overview](../cd-model/cd-model-overview.md)
- [Stages 1-6](../cd-model/cd-model-stages-1-6.md)
- [Stages 7-12](../cd-model/cd-model-stages-7-12.md)
- [Testing Strategy Overview](../testing/testing-strategy-overview.md)
