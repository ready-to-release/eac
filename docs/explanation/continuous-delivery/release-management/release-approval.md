# Release Approval Patterns

{{ page_breadcrumb() }}

## Introduction

Release approval is the **final decision** before production deployment: "Is this code ready for production?" The CD Model supports two fundamentally different approaches to this decision:

**Release Approval (RA) Pattern**: Human makes the decision (Stage 9 manual approval)

**Continuous Deployment (CDe) Pattern**: Automation makes the decision (Stage 9 automated approval, human approved at Stage 3)

Both patterns validate the same quality criteria - they differ in **who decides**, **when they decide**, and **how long it takes**.

---

## RA Pattern: Manual Approval

### How It Works

**Stage 3 (Merge Request)**: First-level approval

- **Who**: Peer reviewer
- **What**: Code quality, design, tests adequate
- **NOT approving**: Production deployment (that's Stage 9)

**Stage 9 (Release Approval)**: Second-level approval

- **Who**: Release manager
- **What**: Production readiness, business timing, risk assessment
- **Approving**: Production deployment

**Process**:

1. Code merged at Stage 3 (peer-approved code quality)
2. Automated validation in Stages 4-6 (build, test, PLTE)
3. Exploration in Stage 7 (stakeholder validation)
4. Stage 8: Release candidate created, documentation prepared
5. **Stage 9: Release manager reviews evidence and approves (MANUAL)**
6. Stage 10: Production deployment proceeds

**Timeline**: 1-2 weeks from commit to production

### Release Manager Role

**Responsibilities**:

- Review quality metrics (tests, coverage, performance, security)
- Review documentation (release notes, runbooks, rollback procedures)
- Assess business risk (what's the impact if this fails?)
- Evaluate timing (is now a good time to deploy?)
- Make go/no-go decision
- Coordinate with stakeholders

**Skills required**:

- Technical understanding (can read metrics, understand risks)
- Business judgment (balance technical quality with business needs)
- Risk assessment (identify and mitigate risks)
- Communication (coordinate with teams, stakeholders)

**Decision authority**:

- Can approve (proceed to Stage 10)
- Can reject (fix issues, return to earlier stage)
- Can defer (quality acceptable, but timing is bad)

### Approval Checklist

```markdown
## Stage 9 Approval Checklist

### Quality Metrics
- [ ] All tests passing (100%)
- [ ] Code coverage ≥ 80%
- [ ] Zero critical/high bugs
- [ ] Performance regression < 5%
- [ ] Zero critical/high security vulnerabilities

### Documentation
- [ ] Release notes complete and reviewed
- [ ] Deployment runbook complete
- [ ] Rollback procedure documented and tested
- [ ] Risk assessment reviewed
- [ ] Stakeholder sign-offs obtained

### Business Considerations
- [ ] Deployment timing acceptable (not before holiday, not during peak hours)
- [ ] Dependent systems ready (if coordination required)
- [ ] On-call team prepared
- [ ] Customer communication planned (if user-facing changes)

### Decision
- [ ] **APPROVE**: Proceed to Stage 10 (production deployment)
- [ ] **REJECT**: Fix issues (return to Stage X)
- [ ] **DEFER**: Acceptable quality, deploy later (reschedule)
```

### When to Use RA Pattern

**Best for**:

- Regulated industries (finance, healthcare, government)
- High-risk systems (medical devices, aviation, critical infrastructure)
- Large coordinated releases (multiple systems must update together)
- Organizations requiring formal approval (compliance, audit)
- Teams building deployment maturity

**Characteristics**:

- Manual approval provides human judgment
- Business-driven release timing
- Formal decision-making process
- Clear accountability (release manager approved)

---

## CDe Pattern: Automated Approval

### How It Works

**Stage 3 (Merge Request)**: Combined first and second-level approval

- **Who**: Peer reviewer
- **What**: Code quality, design, tests adequate, AND production approval
- **Approving**: Production deployment (by merging)

**Stage 9 (Release Approval)**: Automated gate

- **Who**: Automated system
- **What**: Objective quality thresholds
- **Process**: If thresholds met, auto-approve (proceed to Stage 10)

**Process**:

1. **Code merged at Stage 3 (peer-approved for PRODUCTION)**
2. Automated validation in Stages 4-6 (build, test, PLTE)
3. Exploration in Stage 7 (optional, may be skipped)
4. Stage 8: Release candidate created automatically
5. **Stage 9: Automated threshold evaluation (SECONDS)**
6. Stage 10: Production deployment proceeds automatically

**Timeline**: 2-4 hours from commit to production

### Automated Gate Evaluation

**Pseudocode**:

```python
def evaluate_stage_9_approval():
    """Automated Stage 9 approval for CDe pattern"""

    # Check quality thresholds
    if not all_tests_pass():
        return REJECT("Tests failing - cannot proceed")

    if code_coverage < 80:
        return REJECT(f"Coverage {code_coverage}% < 80%")

    if critical_bugs > 0:
        return REJECT(f"{critical_bugs} critical bugs open")

    if high_bugs > 0:
        return REJECT(f"{high_bugs} high bugs open")

    if performance_regression_percent > 5:
        return REJECT(f"Performance regression {performance_regression_percent}%")

    if critical_vulnerabilities > 0:
        return REJECT(f"{critical_vulnerabilities} critical vulnerabilities")

    if not documentation_complete():
        return REJECT("Release documentation incomplete")

    # All criteria met - auto-approve
    log_approval("Automated approval: all quality thresholds met")
    notify_team("Proceeding to Stage 10 deployment")
    return APPROVE
```

### Shift-Left Approval

**Key difference**: Approval happens at Stage 3, not Stage 9

When peer reviewer approves merge request in CDe pattern, they're saying:

- ✅ Code quality good
- ✅ Tests adequate
- ✅ **Ready for production deployment**

Stage 9 becomes a **verification** (did we meet criteria?) not a **decision** (should we deploy?).

**Implications**:

- Peer reviewer has more responsibility
- Team must trust automated testing
- Quality gates enforce minimum standards (can't be overridden)
- Feature flags decouple deployment from release

### Requirements for CDe Pattern

**1. Robust Automated Testing**:

- Comprehensive unit tests (≥80% coverage)
- Integration tests for service interactions
- Acceptance tests (IV, OV, PV) in PLTE
- Performance tests with realistic load
- Security scans (SAST, DAST, dependency scanning)

**2. Feature Flags**:

- Deploy code with features disabled
- Enable features via flags (Stage 12)
- Instant disable capability
- Gradual rollout per feature

**3. Comprehensive Monitoring**:

- Application metrics (errors, latency, throughput)
- Business metrics (conversion, revenue)
- Automated alerting on threshold breaches
- Real-time dashboards

**4. Automated Rollback**:

- Automated triggers (error rate > 1%, P95 > 500ms)
- Fast rollback (< 5 minutes)
- Tested rollback procedures
- Minimal human intervention

**5. Team Culture**:

- Trust in automation
- Ownership of quality (developers own quality, not QA)
- Fast feedback loops valued
- Continuous improvement mindset

### When to Use CDe Pattern

**Best for**:

- Fast-moving SaaS products
- Non-regulated industries (or light regulation)
- Teams with strong automated testing culture
- Organizations with feature flag infrastructure
- Independent services (microservices with clear boundaries)

**Characteristics**:

- Automation provides consistency
- Fast feedback (2-4 hours to production)
- Continuous flow (no batching)
- Forces quality discipline

---

## Choosing Between Patterns

### Decision Framework

**Question 1: Are you in a regulated industry?**

- Yes (finance, healthcare, government) → Likely **RA pattern**
- No → Consider CDe pattern

**Question 2: Do you have robust automated testing?**

- Yes (≥80% coverage, comprehensive acceptance/performance tests) → Consider **CDe pattern**
- No → Start with **RA pattern**, build testing maturity

**Question 3: Do you have feature flags?**

- Yes (infrastructure in place, team uses them) → Consider **CDe pattern**
- No → **RA pattern**, or invest in feature flags

**Question 4: What's your deployment frequency?**

- Multiple times per day desired → **CDe pattern**
- Weekly/monthly acceptable → **RA pattern**

**Question 5: What's your team's risk tolerance?**

- High (trust automation, willing to learn from production issues) → **CDe pattern**
- Low (prefer human oversight, avoid production issues) → **RA pattern**

### Pattern Comparison

| Factor | RA Pattern | CDe Pattern |
|--------|-----------|-------------|
| **Approval** | Manual (release manager) | Automated (objective criteria) |
| **Timeline** | 1-2 weeks | 2-4 hours |
| **Stage 3 approves** | Code quality only | Code quality + production |
| **Stage 9 approves** | Production deployment | Automated validation |
| **Release timing** | Business-driven | Continuous |
| **Feature control** | Optional | Required (feature flags) |
| **Testing requirements** | Standard | Comprehensive |
| **Risk mitigation** | Human judgment | Automation + feature flags |
| **Best for** | Regulated, high-risk | Fast-moving, mature testing |

---

## Approval Decision Framework

### Objective vs Subjective Criteria

**Objective (measurable, automatable)**:

- Test pass rate: 100%
- Code coverage: ≥80%
- Critical bugs: 0
- Performance regression: <5%
- Security vulnerabilities: 0 critical/high

**Subjective (requires human judgment)**:

- Business timing: Is now a good time?
- Risk assessment: What's the business impact?
- Stakeholder readiness: Are teams prepared?
- Market conditions: Customer sentiment, competitive landscape

**RA pattern**: Uses both objective and subjective criteria

**CDe pattern**: Only objective criteria (subjective handled at Stage 3 or via feature flags)

### Making the Go/No-Go Decision

**RA Pattern decision tree**:

```text
1. Objective quality criteria met?
   - No → REJECT (fix issues)
   - Yes → Continue to 2

2. Documentation complete?
   - No → REJECT (complete documentation)
   - Yes → Continue to 3

3. Business timing acceptable?
   - No → DEFER (deploy later)
   - Yes → Continue to 4

4. Risk acceptable?
   - No → REJECT or DEFER (mitigate or delay)
   - Yes → APPROVE
```

**CDe Pattern decision (automated)**:

```text
1. All objective criteria met?
   - No → REJECT (automated block)
   - Yes → APPROVE (proceed automatically)
```

---

## Organizational Considerations

### Team Structure

**RA Pattern**:

- Dedicated release manager role (or rotation)
- Clear escalation path
- Release committee (for high-risk changes)

**CDe Pattern**:

- No dedicated release manager
- Peer reviewers share responsibility
- On-call team handles production issues

### Communication

**RA Pattern**:

- Formal approval communication (email, ticket)
- Stakeholder notifications
- Scheduled deployment windows
- Change advisory board (if required)

**CDe Pattern**:

- Automated notifications (Slack, email)
- Continuous deployment announcements
- No scheduled windows (deploy anytime)
- Lightweight communication

### Accountability

**RA Pattern**:

- Release manager accountable for approval decision
- Clear audit trail (who approved, when, why)
- Documented decision-making

**CDe Pattern**:

- Peer reviewer accountable for Stage 3 approval
- Team collectively accountable for quality
- Automated audit trail (quality gates passed)

---

## Transitioning Between Patterns

### RA → CDe (Building Maturity)

**Phase 1: Build automated testing**:

- Increase unit test coverage (target ≥80%)
- Add integration tests
- Implement acceptance tests (IV, OV, PV)
- Add performance tests

**Phase 2: Implement feature flags**:

- Choose feature flag system
- Integrate into application
- Train team on flag usage
- Practice gradual rollouts

**Phase 3: Enhance monitoring**:

- Implement comprehensive metrics
- Add automated alerting
- Create dashboards
- Practice rollback procedures

**Phase 4: Pilot CDe on low-risk service**:

- Choose non-critical service
- Implement automated Stage 9 approval
- Monitor closely
- Iterate based on learnings

**Phase 5: Expand gradually**:

- Apply CDe to additional services
- Refine automated criteria
- Build team confidence
- Eventually: all services on CDe

### CDe → RA (Increasing Control)

**Rare, but may happen when**:

- Entering regulated industry
- High-profile incidents require more oversight
- Organizational mandate for formal approval

**Transition**:

- Implement release manager role
- Add manual approval at Stage 9
- Keep automated validation
- Maintain feature flags (still valuable)

---

## Best Practices Summary

**RA Pattern**:

1. Clear approval criteria (document what "production ready" means)
2. Timely approvals (< 24 hours SLA)
3. Rotate release managers (spread knowledge, avoid bottleneck)
4. Document decisions (why approved, why rejected)
5. Continuous improvement (review what went wrong/right)

**CDe Pattern**:

1. Comprehensive automated testing (can't fake quality)
2. Feature flags required (decouple deployment from release)
3. Robust monitoring (detect issues fast)
4. Fast rollback (< 5 minutes)
5. Team ownership (everyone responsible for production quality)

**Both Patterns**:

1. Objective quality criteria (define thresholds)
2. Complete documentation (before approval, not after)
3. Risk assessment (identify what could go wrong)
4. Rollback procedures (test before needing them)
5. Continuous improvement (learn from deployments)

---

## Next Steps

- [Release Documentation](./release-documentation.md) - What documentation is required
- [Release Quality Gates](../quality-gates/release-gates.md) - Stage 9 thresholds explained
- [Implementation Patterns](../cd-model/implementation-patterns.md) - RA vs CDe patterns in full CD Model context
- [CD Model Stages 7-12](../cd-model/cd-model-stages-8-12.md) - Release stages detailed

## Quick Reference

- [Release Quality Thresholds Reference](../quality-gates/release-gates.md) - Approval threshold specifications
- [CDE Stage Breakdown Reference](../cd-model/implementation-patterns.md) - CDe pattern stage details
- [RA Stage Breakdown Reference](../cd-model/implementation-patterns.md) - RA pattern stage details

{{ diataxis_footer() }}
