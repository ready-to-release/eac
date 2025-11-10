# Repository Patterns

## Introduction

Repository organization is a foundational decision that affects how teams collaborate, how code is versioned, and how the CD Model is implemented.

The way you structure your repositories influences build times, dependency management, deployment coordination, and team autonomy.

This article explains the two primary repository patterns - monorepo and polyrepo - and provides guidance on choosing the right approach for your organization and system architecture.

### Impact on CD Model

Repository structure directly impacts several CD Model stages:

- **Stage 3 (Merge Request)**: Code review scope and automation
- **Stage 4 (Commit)**: Build and test execution strategy
- **Stage 5 (Acceptance Testing)**: PLTE provisioning and test coordination
- **Stage 8 (Start Release)**: Release candidate creation and versioning
- **Stage 10 (Production Deployment)**: Deployment coordination and sequencing

Choosing the right pattern aligns repository organization with your system architecture and team structure.

---

## Monorepo Pattern

![Monorepo Structure](../../../assets/repository/single.drawio.png)

**This diagram illustrates the single-repository (mono-repository) pattern:** The diagram shows a single repository containing more than one deployable unit. In this pattern, multiple deployable units share a single version history and repository boundary. Path filters (glob patterns) define the boundaries of each deployable unit within the repository, allowing independent versioning and deployment of each unit despite sharing the same repository. This is also called a **single-repository** to distinguish it from large-scale organizational mono-repositories. The pattern enables atomic cross-cutting changes while maintaining independent deployment pipelines per deployable unit.

### Characteristics

**Single Repository:**

- All application code in one repository
- Shared infrastructure and tooling
- Unified version history
- Single source of truth

**Code Organization:**

- Organized by module or service directories
- Shared libraries and utilities
- Common configuration files
- Unified dependency management

**Example Structure:**

```
monorepo/
├── services/
│   ├── api/
│   ├── web/
│   └── worker/
├── shared/
│   ├── models/
│   └── utils/
├── infrastructure/
└── docs/
```

### Benefits

**Atomic Changes Across Boundaries:**

- Refactor multiple services in single commit
- Update shared libraries and consumers together
- No breaking changes across repository boundaries
- Single pull request for cross-cutting changes

**Simplified Dependency Management:**

- Shared dependencies at root level
- No version conflicts between modules
- Easier to ensure consistency
- Single dependency update affects all modules

**Code Sharing:**

- Shared utilities and libraries
- Reusable components
- Common infrastructure code
- Centralized documentation

**Unified Tooling:**

- Single CI/CD pipeline configuration
- Shared code quality tools
- Consistent build process
- Centralized security scanning

**Easier Refactoring:**

- Cross-module refactoring in one PR
- Immediate validation of changes
- No coordination between repositories
- Safe, atomic updates

### Tradeoffs

**Build Times:**

- Potentially longer build times, if not doing incremental builds etc.
- Need for selective builds (build only changed modules) locally and in pipelines
- Caching strategies essential
- Requires sophisticated build tooling

**Access Control:**

- Coarser-grained permissions
- All developers can see all code
- May not suit distributed teams with confidentiality needs
- CODEOWNERS file helps but limited

**Repository Size:**

- Can grow very large over time
- Git operations may slow down
- Requires Git LFS for large assets
- Clone times increase

**Cognitive Overhead:**

- Developers may be overwhelmed by scope
- Harder to navigate large codebase
- IDE performance considerations

### When to Use Monorepo

**Best for:**

- **Team coupled services**: Services that a team maintains together
- **Shared libraries**: Heavy code reuse across projects
- **Small to medium teams**: < 50-100 developers working in same domain, however it can scale to google sized orgs.
- **Rapid iteration**: Fast-moving products requiring frequent cross-cutting changes
- **Unified ownership**: Single team or organization owns all code

**Example Scenarios:**

- Startup with multiple microservices owned by one team
- Platform with shared component library
- Full-stack application with frontend, backend, and infrastructure code
- Internal tools suite with common dependencies

### CD Model Integration

**Stage 4 (Commit):**

- Use change detection to build only affected modules
- Run targeted test suites based on changed files
- Generate individual build artifacts pr. module

**Stage 5 (Acceptance Testing):**

- Single PLTE instance with a service vertical deployed
- Simplified service coordination
- Integrated end-to-end testing in extended testing stage 6

**Stage 8 (Start Release):**

- Create independent release tags for all modules
- Independent release notes

**Stage 10 (Production Deployment):**

- Multiple deployment pipelines are affecting production
- Deployment orchestration when multiple services changed in same change.
- Should use feature flags to decouple deployment from release

---

## Polyrepo Pattern

![Polyrepo Structure](../../../assets/repository/poly.drawio.png)

**This diagram illustrates the poly-repository pattern:** The diagram shows the pattern where a repository boundary perfectly aligns with a single deployable unit boundary. In this pattern, one repository contains exactly one deployable unit - whether that's a versioned component (library, container, package) or a runtime system (service, application). The version of any commit is directly equal to the version of the deployable unit, making versioning simple and straightforward. This pattern is commonly used in GitHub open-source projects and enforces decoupling through versioned modules, where dependencies between units are managed through published versioned artifacts consumed via package managers rather than direct code references.

### Characteristics

**Multiple Repositories:**

- One repository per service or versioned component (one repository pr. deployable unit)
- Independent version history
- Separate access controls
- Isolated tooling and configuration

**Code Organization:**

- Each repository is self-contained
- Dependencies managed per repository
- Independent documentation
- Service-specific configuration

**Example Structure:**

```text
organization/
├── api-service/
├── web-service/
├── worker-service/
├── shared-models/
├── shared-utils/
└── infrastructure/
```

### Benefits

**Team Autonomy:**

- Teams own and control their repositories
- Independent decision-making
- Different tech stacks possible
- Reduced coordination overhead

**Clear Boundaries:**

- Enforces service boundaries
- Prevents unintended coupling
- Clear ownership and responsibility
- API-first integration

**Independent Deployment:**

- Services deploy on separate schedules
- Faster deployment cycles per service
- Lower blast radius for changes
- Easier rollback of individual services

**Granular Access Control:**

- Fine-grained permissions per repository
- Suitable for distributed teams
- Supports confidential projects
- Compliance with data access requirements

**Smaller, Focused Repositories:**

- Faster clone times
- Simpler navigation
- Better IDE performance
- Focused code reviews

### Tradeoffs

**Cross-Repository Changes:**

- Changes spanning multiple repos require coordination
- Multiple pull requests needed
- Potential for breaking changes
- Version compatibility challenges
- Contract testing required
- Each repository must have standardized (versioned, named etc.) artifacts produced
- Repository to Repository bindings are NOT allowed ever. Instead formal version dependency menagement must be used

**Dependency Management Complexity:**

- Shared libraries versioned separately
- Version conflicts between repositories
- Need for dependency update coordination
- Breaking changes require careful management

**Tooling Duplication:**

- CI/CD configuration duplicated across repos
- Inconsistent tooling possible
- More maintenance overhead
- Potential for drift

**Discoverability:**

- Harder to find related code
- No unified search across repositories
- Documentation spread across repos
- Learning curve for new developers

### When to Use Polyrepo

**Best for:**

- **Loosely coupled services**: Microservices with independent lifecycles
- **Large organizations**: Multiple teams with separate ownership
- **Distributed teams**: Teams in different locations or with confidentiality needs
- **Independent deployment cadences**: Services that release on different schedules
- **Clear service boundaries**: Well-defined APIs between services (Contracts)

**Example Scenarios:**

- Large enterprise with independent product teams
- Microservices with separate deployment schedules
- Organization with confidential or regulated modules
- Platform with third-party integrations requiring isolation

### CD Model Integration

**Stage 4 (Commit):**

- Independent build pipelines per repository
- Parallel builds across repositories
- Repository-specific testing

**Stage 5 (Acceptance Testing):**

- PLTE must coordinate multiple repositories
- Version pinning for service dependencies
- Contract testing between services
- More complex environment setup

**Stage 8 (Start Release):**

- Independent release tags per repository
- Service-specific versioning
- Separate release notes per service

**Stage 10 (Production Deployment):**

- Deploy services independently
- Backward compatibility requirements
- Rolling deployment strategies
- API versioning for compatibility

---

## Repository Types

![Repository Types](../../../assets/repository/types.drawio.png)

**This diagram shows the repository type taxonomy:** The diagram categorizes repositories by the number of deployable units they contain. **Poly-repository** (left) contains exactly one deployable unit - the repository boundary perfectly aligns with the deployable unit boundary, making versioning simple (any commit = new version of the unit). **Mono-repository** (right) contains more than one deployable unit, with three subtypes: **Team mono-repository** (single-repository) where one team owns multiple deployable units, **Product mono-repository** (single-repository) where multiple teams collaborate on one product's deployable units, and **Organizational mono-repository** used by large organizations like Google/Facebook (not recommended for most teams). The diagram establishes the fundamental distinction: poly = one deployable unit, mono = multiple deployable units.

### Side-by-Side Analysis

```mermaid
flowchart LR
    subgraph Monorepo["Monorepo"]
        MR_All[All code<br/>one repo]
        MR_Atomic[Atomic changes]
        MR_Shared[Direct sharing]

        MR_All --> MR_Atomic
        MR_All --> MR_Shared
    end

    subgraph Polyrepo["Polyrepo"]
        PR_Many[Multiple repos]
        PR_Indep[Independent teams]
        PR_Deploy[Independent deploy]

        PR_Many --> PR_Indep
        PR_Many --> PR_Deploy
    end

    style MR_Atomic fill:#e8f5e9
    style MR_Shared fill:#e8f5e9
    style PR_Indep fill:#e3f2fd
    style PR_Deploy fill:#e3f2fd
```

| Factor                     | Monorepo                      | Polyrepo                    |
| -------------------------- | ----------------------------- | --------------------------- |
| **Atomic Changes**         | ✅ Excellent - single commit  | ❌ Difficult - multiple PRs |
| **Team Autonomy**          | ⚠️ Limited - shared decisions | ✅ Excellent - independent  |
| **Build Times**            | ⚠️ Potentially long           | ✅ Fast per repository      |
| **Dependency Management**  | ✅ Simple - unified           | ⚠️ Complex - versioned      |
| **Code Reuse**             | ✅ Easy - shared directly     | ⚠️ Requires versioning      |
| **Access Control**         | ⚠️ Coarse-grained             | ✅ Fine-grained             |
| **Discoverability**        | ✅ All code in one place      | ⚠️ Spread across repos      |
| **Independent Deployment** | ⚠️ Coordinated releases       | ✅ Independent cycles       |
| **Tooling**                | ✅ Unified                    | ⚠️ Duplicated               |

### Decision Factors

**Choose Monorepo if:**

- Services frequently change together
- Heavy code sharing between modules
- Small to medium team size
- Unified ownership and responsibility
- Need atomic cross-cutting changes

**Choose Polyrepo if:**

- Services have independent lifecycles
- Multiple teams with separate ownership
- Need fine-grained access control
- Services deploy on different schedules
- Clear service boundaries exist

---

## Poor Repository Design (Anti-Pattern)

![Repository Anti-Pattern](../../../assets/repository/bad.drawio.png)

**This diagram shows the anti-pattern of splitting repositories by technical boundary:** The diagram illustrates the problematic pattern of organizing repositories by technology type rather than by deployable unit boundaries. For example, creating separate repositories for **frontend/**, **backend/**, **scripts/**, **documentation/**, and **infrastructure/** - each representing a technical concern rather than a cohesive deployable unit. This creates dependency hell where there is no coherency or frontier of what one specific version consists of. Changes to a single feature require coordinating across multiple repositories (frontend repo, backend repo, scripts repo), with no clear version boundary for the complete system. This violates the principle that repositories should align with either poly-repository (one deployable unit) or single-repository (multiple deployable units owned together) patterns.

### How to Avoid This Anti-Pattern

**✅ DO**:

- Organize repositories around deployable unit boundaries, not technical boundaries
- Use poly-repository pattern (one deployable unit per repository) OR single-repository pattern (multiple deployable units per repository owned by same team)
- Keep all code for a deployable unit together (frontend, backend, scripts, docs, infrastructure) in the same repository
- Version Everything-as-Code (EaC) artifacts together, grouped by deployable unit boundaries

**❌ DO NOT**:

- Create separate repositories for frontend, backend, scripts, documentation, or infrastructure unless each is a distinct deployable unit
- Create loose gatherings of repositories with cross-repository dependencies
- Split a single deployable unit across multiple repositories

---

## Best Practices

**Module Boundaries**: Define explicit boundaries, document dependencies, enforce with tooling (linters, dependency checks)

**Versioning**:

- Monorepo: Unified or independent module versioning
- Polyrepo: Semantic versioning per repository, dependency manifests specify compatible versions

**Dependency Management**:

- Monorepo: Root-level with lock file for consistency
- Polyrepo: Per repository with published versioned libraries and contract testing

**Automation Tools**:

- Monorepo: Change detection, selective builds, distributed caching (Nx, Bazel, Turborepo)
- Polyrepo: Repository templates, shared pipeline definitions, automated updates (Dependabot)

---

## Impact on CD Model Stages

| Stage        | Monorepo                                | Polyrepo                                   |
| ------------ | --------------------------------------- | ------------------------------------------ |
| **Stage 3**  | Larger PRs, single review               | Smaller focused PRs, may need coordination |
| **Stage 4**  | Change detection for selective builds   | Independent builds in parallel             |
| **Stage 5**  | Single PLTE with all services           | Version pinning, contract testing          |
| **Stage 8**  | Single orchestrated release event possible for RA                      | Multiple independent release tags          |
| **Stage 10** | Coordinated deployment or feature flags | Independent deployment schedules           |

**Polyrepo Coordination Requirements**: Contract testing for API compatibility, version pinning in PLTE, deployment sequencing for backward compatibility, API versioning for gradual rollout.

---

## Next Steps

- [CD Model Overview](../cd-model/cd-model-overview.md) - Understand how repos integrate with stages
- [Stages 1-6](../cd-model/cd-model-stages-1-6.md) - See repository impact on development
- [Environments](environments.md) - Understand PLTE provisioning strategies
- [Implementation Patterns](../cd-model/implementation-patterns.md) - Choose RA or CDe pattern
- [Testing Strategy Integration](../testing/testing-strategy-integration.md) - Test coordination approaches

## References

- [CD Model Overview](../cd-model/cd-model-overview.md)
- [Trunk-Based Development](../workflow/trunk-based-development.md)
- [Repository Layout](../../reference/continuous-delivery/repository-layout.md)
