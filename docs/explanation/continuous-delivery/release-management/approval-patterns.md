# Approval Patterns

## Overview

Release approval is the **final decision** before production deployment: "Is this code ready for production?"

The CD Model supports two approaches:

| Pattern                         | Who Decides             | When                                | Timeline  |
| ------------------------------- | ----------------------- | ----------------------------------- | --------- |
| **Release Approval (RA)**       | Human (release manager) | Stage 9                             | 1-2 weeks |
| **Continuous Deployment (CDe)** | Automation              | Stage 9 (human approved at Stage 3) | 2-4 hours |

Both patterns validate the same quality criteria - they differ in **who decides** and **when**.

---

## RA Pattern: Manual Approval

### How It Works

**Two-level approval:**

1. **Stage 3 (Merge Request)**: Peer reviewer approves code quality
2. **Stage 9 (Release Approval)**: Release manager approves production deployment

**Flow:**

```text
Code merged (Stage 3) → Build/Test (4-6) → Exploration (7) →
Release candidate (8) → MANUAL APPROVAL (9) → Production (10)
```

### Release Manager Role

**Responsibilities:**

- Review quality metrics (tests, coverage, security)
- Review documentation (release notes, runbooks)
- Assess business risk and timing
- Make go/no-go decision
- Coordinate with stakeholders

**Decision options:**

- **Approve**: Proceed to production
- **Reject**: Fix issues, return to earlier stage
- **Defer**: Quality acceptable, but timing is bad

### When to Use RA Pattern

- Regulated industries (finance, healthcare, government)
- High-risk systems (critical infrastructure)
- Large coordinated releases
- Organizations requiring formal approval
- Teams building deployment maturity

---

## CDe Pattern: Automated Approval

### How It Works

**Single approval point:**

1. **Stage 3 (Merge Request)**: Peer reviewer approves for production (combined approval)
2. **Stage 9**: Automated gate verifies objective criteria

**Flow:**

```text
Code merged for PRODUCTION (Stage 3) → Build/Test (4-6) →
Release candidate (8) → AUTO-APPROVE (9) → Production (10)
```

### Shift-Left Approval

The key difference: approval happens at Stage 3, not Stage 9.

When a peer reviewer approves in CDe, they're saying:

- Code quality is good
- Tests are adequate
- **Ready for production deployment**

Stage 9 becomes verification (did we meet criteria?) not decision (should we deploy?).

### Requirements for CDe

| Requirement              | Why                                         |
| ------------------------ | ------------------------------------------- |
| Robust automated testing | No human gate means tests must catch issues |
| Feature flags            | Decouple deployment from release            |
| Comprehensive monitoring | Detect issues in production fast            |
| Automated rollback       | Recover without human intervention          |
| Team trust in automation | Cultural readiness                          |

### When to Use CDe Pattern

- Fast-moving SaaS products
- Non-regulated industries
- Strong automated testing culture
- Feature flag infrastructure in place
- Independent services (microservices)

---

## Choosing Between Patterns

### Decision Framework

| Question                                  | RA Pattern                   | CDe Pattern             |
| ----------------------------------------- | ---------------------------- | ----------------------- |
| Regulated industry?                       | Yes                          | No                      |
| Robust automated testing (≥80% coverage)? | Optional                     | Required                |
| Feature flags in place?                   | Optional                     | Required                |
| Desired deployment frequency?             | Weekly/monthly               | Multiple per day        |
| Risk tolerance?                           | Low (prefer human oversight) | High (trust automation) |

### Pattern Comparison

| Factor           | RA Pattern               | CDe Pattern                    |
| ---------------- | ------------------------ | ------------------------------ |
| Approval         | Manual (release manager) | Automated (objective criteria) |
| Timeline         | 1-2 weeks                | 2-4 hours                      |
| Stage 3 approves | Code quality only        | Code quality + production      |
| Stage 9          | Human decision           | Automated verification         |
| Release timing   | Business-driven          | Continuous                     |
| Feature flags    | Optional                 | Required                       |
| Risk mitigation  | Human judgment           | Automation + flags             |

---

## Objective vs Subjective Criteria

**Objective (automatable):**

- Test pass rate: 100%
- Code coverage: ≥80%
- Critical bugs: 0
- Performance regression: <5%
- Security vulnerabilities: 0 critical/high

**Subjective (requires human judgment):**

- Business timing: Is now a good time?
- Risk assessment: What's the business impact?
- Stakeholder readiness: Are teams prepared?

**RA pattern**: Uses both objective and subjective criteria

**CDe pattern**: Only objective criteria (subjective handled at Stage 3 or via feature flags)

---

## Next Steps

- [Release Evidence](./release-evidence.md) - What evidence is required for approval
- [Changelog System](./changelog-system.md) - How to release in this repository
- [Quality Gates](../quality-gates/index.md) - Stage-specific quality thresholds
