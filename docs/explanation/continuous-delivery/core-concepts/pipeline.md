# Deployment Pipeline

## Introduction

The Deployment Pipeline is the automated, repeatable process that takes code from commit to production (or artifact registry).

Every change flows through the pipeline, which validates quality, builds artifacts, and delivers software to its destination.

![Deployment Pipeline Legend](../../../assets/branching/legend-deployment-pipeline.drawio.png){width=250}

---

## Definition

**Deployment Pipeline:** The automated process that validates and delivers changes through the CD Model stages.

**Key Characteristics:**

- Triggered by changes to trunk
- Implements the 12 stages of the CD Model
- Provides automated quality gates
- Generates immutable artifacts
- Collects evidence and metrics

---

## One Pipeline Per Deployable Module

Each [deployable module](deployable-modules.md) has its own pipeline instance:

- **Polyrepo**: One pipeline per repository
- **Monorepo**: Multiple pipelines in one repository (one per deployable module)

Pipelines are triggered when changes affect their deployable module. In a monorepo, change detection determines which pipelines run.

```text
Trunk Commit
    │
    ├─► Module A changed? ─► Run Pipeline A
    ├─► Module B changed? ─► Run Pipeline B
    └─► Module C changed? ─► Run Pipeline C
```

---

## Pipeline Triggers

Pipelines can be triggered by:

| Trigger | When | Purpose |
|---------|------|---------|
| Commit | Code merged to trunk | Main validation and artifact creation |
| Pull Request | PR opened/updated | Pre-merge validation |
| Schedule | Cron (daily, weekly) | Dependency updates, security scans |
| Manual | User-initiated | Releases, hotfixes |
| Tag | Git tag created | Release candidate creation |

---

## The 12 Stages

The pipeline implements the CD Model's 12 stages:

**Development Stages (1-7):**

1. **Authoring** - Developer creates changes on topic branch
2. **Pre-commit** - Local validation before push
3. **Merge Request** - PR validation and review
4. **Commit** - Build artifacts on trunk merge
5. **Acceptance Testing** - Validate in PLTE
6. **Extended Testing** - Performance, security, compliance
7. **Exploration** - Stakeholder validation in Demo

**Release Stages (8-12):**

8. **Start Release** - Create release candidate
9. **Release Approval** - Manual or automated gate
10. **Production Deployment** - Deploy to live
11. **Live** - Monitor in production
12. **Release Toggling** - Feature flag control

See [CD Model Stages](../cd-model/stages.md) for detailed stage descriptions.

---

## Evidence Collection

The pipeline collects evidence at each stage for traceability and compliance:

| Stage | Evidence Collected |
|-------|-------------------|
| Pre-commit | Lint results, unit test results |
| Commit | Build logs, artifact checksums, SBOM |
| Acceptance Testing | Test reports, coverage metrics |
| Extended Testing | Performance benchmarks, security scan results |
| Production Deployment | Deployment logs, health check results |
| Live | Monitoring dashboards, incident records |

Evidence enables:

- Audit trails for compliance
- Debugging production issues
- Release quality metrics
- Continuous improvement

---

## Pipeline Variants

The pipeline structure varies based on the CD Model variant:

**Release Assurance (RA):**

- Manual approval gate at Stage 9
- Release branches for stabilization
- Explicit release candidate management

**Continuous Deployment (CDe):**

- Automated approval based on quality gates
- Direct trunk-to-production flow
- Feature flags for release control

See [CD Variants](../cd-model/variants.md) for detailed comparison.

---

## Pipeline Implementation

Pipelines are typically implemented using:

- **GitHub Actions** - `.github/workflows/`
- **Azure DevOps** - `azure-pipelines.yml`
- **GitLab CI** - `.gitlab-ci.yml`
- **Jenkins** - `Jenkinsfile`

The pipeline definition is stored in the trunk alongside the code it builds, following "Everything as Code" principles.

---

## Next Steps

- [CD Model Stages](../cd-model/stages.md) - Detailed stage descriptions
- [CD Variants](../cd-model/variants.md) - RA vs CDe patterns
- [Deployable Modules](deployable-modules.md) - What pipelines build
- [Live](live.md) - Where pipelines deliver

## References

- [CD Model Overview](../cd-model/overview.md)
- [Trunk-Based Development](../workflow/trunk-based-development.md)
