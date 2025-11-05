# CD Model: The 12-Stage Framework

The Continuous Delivery Model is a comprehensive framework for delivering software from development through production with quality, traceability, and compliance built in.

This section covers the complete 12-stage model, how to read the visual notation, and the two primary implementation patterns (Release Approval and Continuous Deployment).

---

## [Overview & Visual Notation](cd-model-overview.md)

Introduction to the 12-stage Continuous Delivery Model and how to read the visual diagrams.

**Topics covered:**

- What is the CD Model and why it matters
- Traditional approach vs CD Model approach
- Visual notation and legend explanations
- Overview of all 12 stages
- Key principles: shift-left, fail fast, automation, traceability

**Read this first** to understand the visual language used throughout all CD Model documentation.

---

## [Stages 1-6: Development to Testing](cd-model-stages-1-6.md)

Detailed explanation of development and testing stages from code authoring through extended validation.

**The 12 stages in detail:**

**Stage 1: Authoring Changes**:

- Requirements as Code with Gherkin specifications
- Feature ID tracking and traceability

**Stage 2: Pre-commit**:

- Local validation (5-10 min time-box)
- L0/L1 tests, linting, secret scanning

**Stage 3: Merge Request**:

- Automated checks plus peer review
- Approval gate before trunk integration

**Stage 4: Commit**:

- Integration testing on trunk
- Build immutable artifacts
- L0-L2 tests, Hybrid E2E

**Stage 5: Acceptance Testing**:

- Deploy to PLTE (Production-Like Test Environment)
- L0-L3 end-to-end tests
- Collect verification evidence (IV, OV, PV)

**Stage 6: Extended Testing**:

- Comprehensive validation in PLTE
- Performance testing, security scanning
- Compliance validation

---

## [Stages 7-12: Release to Production](cd-model-stages-7-12.md)

Detailed explanation of release and production stages from stakeholder validation through live monitoring.

**Stage 7: Exploration**:

- Demo environment for stakeholder validation
- L4 exploratory testing and UAT
- Business sign-off

**Stage 8: Start Release**:

- Version tagging and release candidate creation
- RA pattern: Create release branch
- CDE pattern: Tag trunk commit

**Stage 9: Release Approval**:

- RA pattern: Manual approval and sign-off
- CDE pattern: Automated approval via quality gates
- Evidence review and compliance

**Stage 10: Production Deployment**:

- Deploy to production environment
- Blue-green, canary, or rolling deployment strategies
- Rollback procedures

**Stage 11: Live**:

- Production monitoring and observability
- Phased rollout strategies
- Incident response

**Stage 12: Release Toggling**:

- Feature flags for runtime control
- Gradual rollout (1% → 10% → 100%)
- Kill switches for high-risk features

---

## [Implementation Patterns](implementation-patterns.md)

Understanding when and how to use Release Approval (RA) vs Continuous Deployment (CDE) patterns.

The CD Model supports two primary implementation patterns that adjust automation level and manual oversight while maintaining the same 12-stage structure.

**Release Approval (RA) Pattern:**

- Manual approval gates at Stages 3 and 9
- Release branches for validation (Stage 8)
- Comprehensive audit trail and traceability
- **Best for**: Regulated systems, high-risk applications
- **Cycle time**: 1-2 weeks from commit to production

**Continuous Deployment (CDE) Pattern:**

- Automated approval at Stage 9
- No release branches (deploy directly from trunk)
- Feature flags for runtime control
- **Best for**: Non-regulated systems, internal tools
- **Cycle time**: 2-4 hours from commit to production

**Topics covered:**

- When to use each pattern
- Decision tree for pattern selection
- Traceability and audit evidence
- Compliance and signoffs
- Manual vs automated approval gates

---

## How These Articles Work Together

**Start with Overview** to learn:

- The visual notation (legends and diagrams)
- What each stage accomplishes
- Key principles

**Then read Stages 1-6** to understand:

- Development workflow
- Testing levels (L0-L4)
- Quality gates before release

**Continue with Stages 7-12** to understand:

- Release processes
- Production deployment
- Monitoring and control

**Finally read Implementation Patterns** to understand:

- How to adapt the model to your context
- RA vs CDE decision making
- Compliance requirements

---

## Integration with Other Sections

The CD Model is the central framework that integrates with:

**[Core Concepts](../core-concepts/index.md)**:

- Unit of Flow components map to CD Model stages
- Deployable Units progress through all 12 stages

**[Workflow](../workflow/index.md)**:

- Trunk-based development supports Stages 1-4
- Branching strategies differ between RA and CDE patterns

**[Testing](../testing/index.md)**:

- Test levels (L0-L4) execute at specific stages
- Shift-left strategy drives stage timing

**[Architecture](../architecture/index.md)**:

- Environments map to specific stages (DevBox, Build Agents, PLTE, Production)
- Repository patterns affect how pipelines trigger

**[Security](../security/index.md)**:

- Security tools integrate at multiple stages
- Shift-left security principles align with CD Model

---

## Next Steps

- **New to CD Model?** Start with [Overview & Visual Notation](cd-model-overview.md)
- **Want details?** Read [Stages 1-6](cd-model-stages-1-6.md) and [Stages 7-12](cd-model-stages-7-12.md)
- **Choosing a pattern?** See [Implementation Patterns](implementation-patterns.md)
- **Ready to implement?** Explore [Workflow](../workflow/index.md) and [Testing](../testing/index.md)

Return to **[Continuous Delivery Overview](../index.md)** for complete navigation.
