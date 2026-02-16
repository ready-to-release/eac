# Defect Prevention

## Introduction

The CD Model doesn't just detect defects—it systematically **prevents** them through architectural design.

By distributing validation across 12 purpose-built stages, the CD Model creates multiple defensive layers that catch different types of defects at the point where they're cheapest and easiest to fix.

This approach transforms quality from a gate at the end to a continuous property throughout the delivery pipeline.

Understanding which defects each stage prevents helps teams design effective pipelines that shift quality left, fail fast, and maintain high confidence in production deployments.

## Defect Prevention Philosophy

The CD Model prevents defects through three key principles:

**Prevention Over Detection**: Create architectural constraints that make defects difficult to introduce, rather than just detecting them after the fact.

**Shift-Left**: Catch defects as early as possible. Fixing a defect at pre-commit takes minutes; in production it takes weeks.

**Systemic Corrections**: Change processes and architecture to prevent defect classes from recurring (trunk-based development, feature flags, automated gates, contract tests).

## Defect Categories Overview

Software defects fall into **8 categories**, each addressed at specific stages of the CD Model:

- **Product & Discovery** (Stages 1, 7, 12): Building the wrong thing, over-engineering, prioritizing low-value work
- **Integration & Boundaries** (Stages 3, 5, 6): Interface mismatches, race conditions, distributed state issues
- **Knowledge & Communication** (Stages 1, 2, 3): Implicit domain knowledge, ambiguous requirements, divergent mental models
- **Change & Complexity** (Stages 2, 3, 6): Unintended side effects, technical debt, configuration drift
- **Testing & Observability** (Stages 2, 5, 6, 11): Untested edge cases, missing contract tests, insufficient monitoring
- **Process & Deployment** (Stages 2, 3, 4, 9, 10): Long-lived branches, manual pipeline steps, inadequate rollback capability
- **Data & State** (Stages 2, 5, 6): Schema migration failures, null pointer issues, cache invalidation errors
- **Dependency & Infrastructure** (Stages 2, 6): Third-party breaking changes, environment drift, network partition handling

The CD Model prevents defects through **shift-left validation** - catching issues at the earliest stage where detection is practical and cost-effective.

## How This Documentation Is Organized

This section provides three complementary views of defect prevention in the CD Model:

**[Stage-by-Stage Prevention Guide](defect-stages-mapping.md)**:

Walk through CD Model stages 1-12, learning what defects each stage prevents, detection methods, time budgets, best practices, and anti-patterns. Use this guide when designing or implementing pipelines.

**[Comprehensive Defect Catalog](https://bdfinst.github.io/ai-patterns/defect-detection-and-fixes/)**:

Comprehensive catalog of 35 software defects with detection methods, AI improvement opportunities, and systemic corrections. This authoritative source is maintained by Bryan Finster and the continuous delivery community.

> [Local CSV Reference](../../../../assets/cd-model/cd-defect-detection-remediation.csv): Download the catalog as CSV for offline reference.

**[AI-Assisted Detection Strategies](ai-detection.md)**:

Focus on high-value AI opportunities for defect detection, including requirements analysis, test generation, code review, and risk assessment. Use this when exploring AI integration into your CD pipeline.

---

## Related Documentation

**Quality Gates:**

- [Quality Gates Overview](../index.md) - Philosophy and gate design principles
- [Pre-commit Quality Gates](../precommit-gates.md) - Stage 2 defect prevention
- [Merge Request Quality Gates](../merge-request-gates.md) - Stage 3 defect prevention
- [Release Quality Gates](../release-gates.md) - Stage 9 defect prevention

**CD Model:**

- [CD Model Overview](../../cd-model/overview.md) - Understanding the 12-stage model
- [The 12 Stages](../../cd-model/stages.md) - Detailed stage descriptions
- [Key Principles](../../cd-model/principles.md) - Shift-left, fail fast, automation

**Testing:**

- [Testing Strategy](../../testing/index.md) - Test levels and verification types
- [Test Levels](../../testing/test-levels.md) - L0-L6 testing across stages
- [Verification Types](../../testing/verification-types.md) - IV/OV/PV/QV definitions

**Security:**

- [Security Integration](../../security/index.md) - Security throughout the CD Model
- [Shift-Left Security](../../security/shift-left.md) - Early security validation
