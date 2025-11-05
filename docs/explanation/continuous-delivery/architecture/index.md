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

## How These Articles Work Together

**Environments** establish:

- Where code executes at each stage
- What infrastructure is needed
- Security boundaries and access controls

**Repository Patterns** establish:

- How code is organized
- How deployable units relate to repositories
- Dependency management approach

**Together they define:**

- Complete architecture for CD Model implementation
- How code flows from DevBox to Production
- Organizational and technical boundaries

---

## Integration with CD Model Stages

**Environment Mapping by Stage:**

| Stage | Environment | Purpose |
|-------|------------|---------|
| 1. Authoring | DevBox | Local development |
| 2. Pre-commit | DevBox | Local validation (L0/L1) |
| 3. Merge Request | Build Agents | Automated validation (L0-L2) |
| 4. Commit | Build Agents | Build artifacts, integration tests |
| 5. Acceptance | PLTE | End-to-end testing (L3) |
| 6. Extended | PLTE | Performance, security, compliance |
| 7. Exploration | Demo | Stakeholder validation (L4) |
| 8. Start Release | Build Agents | Version tagging |
| 9. Release Approval | Demo/PLTE | Approval validation |
| 10. Production Deploy | Deploy Agents | Controlled production access |
| 11. Live | Production | Real user traffic |
| 12. Release Toggling | Production | Feature flag control |

**Repository Pattern Impact:**

**Monorepo:**

- Single pipeline configuration
- Selective testing based on changed files
- Shared infrastructure across deployable units
- Single Demo and PLTE instance can host all services

**Polyrepo:**

- Separate pipeline per repository
- Independent testing and deployment
- Dedicated infrastructure per deployable unit
- Multiple Demo/PLTE instances may be needed

See **[CD Model](../cd-model/index.md)** for complete stage details.

---

## Integration with Other Sections

**[Core Concepts](../core-concepts/index.md)**:

- Trunk (repository) contains one or more Deployable Units
- Monorepo vs polyrepo affects Unit of Flow implementation
- Repository pattern affects versioning strategy

**[Workflow](../workflow/index.md)**:

- Repository pattern affects branching strategy
- Monorepo requires careful path filtering in pipelines
- Polyrepo requires cross-repo coordination for dependencies

**[Testing](../testing/index.md)**:

- PLTE required for L3 end-to-end tests
- Build Agents execute L0-L2 tests
- Demo environment hosts L4 exploratory tests
- Repository pattern affects test execution strategy

**[Security](../security/index.md)**:

- Environment boundaries enforce security controls
- Deploy Agents provide segregated production access
- Repository pattern affects dependency scanning approach

---

## Infrastructure as Code

Both environment and repository architecture should be defined as code:

**Environment Infrastructure:**

```text
infrastructure/
├── devbox/           # Local development setup scripts
├── build-agents/     # CI/CD agent configuration
├── plte/             # PLTE environment definition
├── demo/             # Demo environment definition
├── deploy-agents/    # Production deployment config
└── production/       # Production infrastructure
```

**Benefits:**

- Version controlled with application code
- Reproducible environments
- Tested in lower environments before production
- Audit trail for infrastructure changes

See **[Everything as Code](../../everything-as-code/index.md)** for comprehensive approach.

---

## Decision Factors

**Choosing Repository Pattern:**

**Use Monorepo when:**

- Services share significant code
- Need atomic cross-service changes
- Small to medium team size
- Want simplified tooling

**Use Polyrepo when:**

- Services are loosely coupled
- Multiple teams with clear ownership boundaries
- Independent release cadences desired
- Different technology stacks

**Environment Decisions:**

**PLTE Investment:**

- Required for L3 end-to-end tests
- Should mirror production configuration
- Cost can be optimized with ephemeral environments

**Demo Environment:**

- Longer-lived for stakeholder access
- May be optional for mature CDE pattern
- Valuable for UAT and business validation

---

## Best Practices

**Environment Management:**

✅ **DO:**

- Define all environments as Infrastructure as Code
- Keep PLTE configuration identical to production
- Use ephemeral environments when possible (cost optimization)
- Monitor environment costs and usage
- Automate environment provisioning

❌ **DON'T:**

- Manually configure environments (leads to drift)
- Allow PLTE to diverge from production
- Share environments between teams (causes conflicts)
- Skip PLTE (L3 tests require production-like environment)

**Repository Organization:**

✅ **DO:**

- Choose pattern based on coupling and team structure
- Keep related code together (avoid artificial boundaries)
- Version shared dependencies explicitly (polyrepo)
- Use monorepo tooling for selective testing (monorepo)
- Document module boundaries clearly

❌ **DON'T:**

- Create monorepo without proper tooling
- Split repositories too finely (nano-repos)
- Mix patterns without reason
- Ignore dependency management strategy

---

## Next Steps

- **Need environment details?** Read [Environments](environments.md)
- **Need repository guidance?** Read [Repository Patterns](repository-patterns.md)
- **Want to understand stages?** See [CD Model](../cd-model/index.md)
- **Ready for workflow?** Explore [Workflow](../workflow/index.md)

Return to **[Continuous Delivery Overview](../index.md)** for complete navigation.
