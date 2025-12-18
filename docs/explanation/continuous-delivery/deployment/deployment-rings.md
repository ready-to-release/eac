# Deployment Rings

## Introduction

Deployment rings are a **progressive rollout strategy** that deploys to increasingly larger user groups (rings) over time. Each ring serves as validation for the next, building confidence through production exposure while limiting blast radius.

Unlike technical deployment strategies (Hot Deploy, Rolling, Blue-Green, Canary) that focus on infrastructure-level rollout, deployment rings are **organizational-level** rollouts that focus on **user segments**.

**Key concept**: Don't expose all users to new code simultaneously. Test with internal users first, then early adopters, then broader audiences, gradually increasing until everyone has the new version.

---

## The Rings Model

### Origin: Windows as a Service

Microsoft pioneered deployment rings for Windows 10 updates, deploying to progressively larger user groups:

- **Ring 0**: Internal Microsoft employees
- **Ring 1**: Windows Insiders (beta testers)
- **Ring 2**: Early adopters (fast ring)
- **Ring 3**: Broad deployment (slow ring)
- **Ring 4**: All users (general availability)

This approach prevents widespread issues by catching problems early with smaller, risk-tolerant audiences.

### Four-Ring Structure

**Standard ring structure**:

| Ring | Name                 | Audience                   | Size   | Duration | Traffic % | Risk Tolerance |
| ---- | -------------------- | -------------------------- | ------ | -------- | --------- | -------------- |
| 0    | Canary               | Internal users, developers | Tiny   | Hours    | 1-5%      | High           |
| 1    | Early Adopters       | Beta users, opted-in users | Small  | 1-2 days | 10-25%    | Medium         |
| 2    | Standard             | Regular users              | Medium | 3-7 days | 50-75%    | Low            |
| 3    | General Availability | All users                  | Large  | Complete | 100%      | Very Low       |

**Progressive exposure**:

- Each ring is **larger** than the previous
- Each ring has **more users** affected
- Each ring has **lower risk tolerance**
- Each ring requires **higher confidence** before progression

---

## Ring Structure Explained

### Ring 0: Canary (Internal Validation)

**Purpose**: Early warning system with minimal external user impact

**Audience**:

- Internal developers and engineers
- DevOps and operations teams
- Internal product teams
- QA testers

**Characteristics**:

- **Size**: 1-5% of total capacity
- **Duration**: 1-4 hours
- **Risk tolerance**: High (internal users understand risk)
- **Monitoring**: Intensive (developers actively watching)

**What's validated**:

- Basic functionality works (no immediate crashes)
- Integration with dependencies (APIs, databases, services)
- Health checks passing
- No critical errors in logs

**Rollback triggers**:

- Any critical error
- Crashes or exceptions
- Failed health checks
- Integration failures

**Why internal users first**:

- Developers can immediately debug issues
- Internal users understand beta risks
- Minimal external reputation impact
- Fast feedback loop

**Example**:

- SaaS product: Deploy to `internal.example.com` subdomain
- Mobile app: Deploy to internal TestFlight group
- API service: Route 5% traffic from internal services only

### Ring 1: Early Adopters (Beta Validation)

**Purpose**: Broader validation with real users who accept risk

**Audience**:

- Beta program participants
- Early adopters (opted-in users)
- Power users willing to provide feedback
- Developer community

**Characteristics**:

- **Size**: 10-25% of total capacity
- **Duration**: 24-48 hours
- **Risk tolerance**: Medium (users opted in to beta program)
- **Monitoring**: Active (alerts configured, dashboards reviewed)

**What's validated**:

- Edge cases and varied usage patterns
- Performance under broader load
- User workflows across different user types
- Feature usability and UX

**Rollback triggers**:

- Error rate exceeds threshold (> 1%)
- Performance degradation (> 10%)
- Multiple user complaints
- Critical workflow broken

**Why early adopters**:

- Willing to tolerate issues
- Provide valuable feedback
- Diverse usage patterns
- Represent real users (not internal developers)

**Example**:

- SaaS product: Enable feature flag for "beta" user segment
- Mobile app: Roll out to 10% of users via app store phased rollout
- API service: Route 25% traffic to new version

### Ring 2: Standard Users (Majority Rollout)

**Purpose**: Majority deployment with continued monitoring

**Audience**:

- Regular, mainstream users
- Production users not in early adopter program
- Standard customer segments

**Characteristics**:

- **Size**: 50-75% of total capacity
- **Duration**: 3-7 days
- **Risk tolerance**: Low (standard production users)
- **Monitoring**: Standard (automated alerts, periodic review)

**What's validated**:

- Stability under full production load
- Long-running performance (multi-day stability)
- Business metrics (conversion rates, engagement)
- Customer satisfaction (support tickets, feedback)

**Rollback triggers**:

- Significant regression (error rate, performance)
- Business metric degradation
- Spike in support tickets
- Negative customer feedback

**Why majority before 100%**:

- Still have 25-50% on previous version (rollback possible)
- Can catch issues that manifest over days (memory leaks, resource exhaustion)
- Business metrics validated at scale

**Example**:

- SaaS product: Feature flag enabled for 75% of users
- Mobile app: 75% phased rollout via app store
- API service: 75% traffic to new version

### Ring 3: General Availability (Complete Rollout)

**Purpose**: Complete deployment to all users

**Audience**:

- All users (100%)
- Conservative user segments (opted out of early access)

**Characteristics**:

- **Size**: 100% of total capacity
- **Duration**: Ongoing (permanent)
- **Risk tolerance**: Very low (full production)
- **Monitoring**: Ongoing (business as usual)

**What's validated**:

- Complete migration successful
- No segments left on old version
- Full decommissioning of old version possible

**Rollback triggers**:

- Critical issues affecting all users
- Major incidents only

**Why final ring**:

- Confidence built through Ring 0, 1, 2 validation
- Issues caught and resolved in earlier rings
- Full user base receives update

**Example**:

- SaaS product: Feature flag enabled for 100% of users, flag removed
- Mobile app: 100% phased rollout, old version deprecated
- API service: 100% traffic, old version decommissioned

---

## Progression Criteria

### Ring 0 → Ring 1

**Criteria**:

- ✅ No critical errors in logs
- ✅ Health checks passing consistently
- ✅ Key integrations functioning
- ✅ Internal users report "works for me"
- ✅ Monitoring dashboards show normal metrics

**Typical wait time**: 1-4 hours

**Decision**: Automated or manual based on metrics

### Ring 1 → Ring 2

**Criteria**:

- ✅ Error rate below threshold (e.g., < 0.5%)
- ✅ P95 latency below threshold (e.g., < 200ms)
- ✅ No critical user-reported issues
- ✅ Positive or neutral user feedback
- ✅ Business metrics stable (conversion, engagement)

**Typical wait time**: 24-48 hours

**Decision**: Usually automated based on objective metrics

### Ring 2 → Ring 3

**Criteria**:

- ✅ All metrics healthy over multi-day period
- ✅ No regressions detected (business or technical)
- ✅ Support ticket volume normal
- ✅ Customer satisfaction scores stable
- ✅ Long-running stability confirmed (no memory leaks, resource exhaustion)

**Typical wait time**: 3-7 days

**Decision**: Usually manual approval (confirms business is ready for 100%)

### Automated Progression

**Example progression logic**:

```python
def should_progress_to_next_ring(current_ring, metrics):
    thresholds = {
        0: {  # Ring 0 → Ring 1
            'min_duration_hours': 2,
            'max_error_rate': 0.01,      # 1%
            'max_p95_latency_ms': 300,
        },
        1: {  # Ring 1 → Ring 2
            'min_duration_hours': 24,
            'max_error_rate': 0.005,     # 0.5%
            'max_p95_latency_ms': 250,
            'min_user_feedback_score': 3.5,
        },
        2: {  # Ring 2 → Ring 3
            'min_duration_hours': 72,
            'max_error_rate': 0.003,     # 0.3%
            'max_p95_latency_ms': 200,
            'max_support_ticket_increase': 0.05,  # 5%
        },
    }

    criteria = thresholds[current_ring]

    if metrics['duration_hours'] < criteria['min_duration_hours']:
        return False  # Haven't waited long enough

    if metrics['error_rate'] > criteria['max_error_rate']:
        return False  # Error rate too high

    if metrics['p95_latency_ms'] > criteria['max_p95_latency_ms']:
        return False  # Latency too high

    # All criteria met
    return True
```

---

## Organizational Considerations

### Building Ring Audiences

**Ring 0 (Internal)**:

- Employees using internal tools/domains
- Developers with debug builds
- QA team with test accounts

**Ring 1 (Early Adopters)**:

- Beta program (users opt-in via settings)
- Power users (identified by usage patterns)
- Developer community (API consumers, integrators)
- Friendly customers (close relationship with your company)

**Ring 2 (Standard)**:

- Regular production users
- Customers without special designation
- Default production traffic

**Ring 3 (Everyone)**:

- Conservative user segments
- Users who opted out of early access
- Final stragglers

### User Communication

**Transparency**:

- Inform users they're in early ring (Ring 0, Ring 1)
- Set expectations about potential issues
- Provide feedback channel

**Example communication**:
> **You're part of our Beta Program!**
>
> You'll receive new features before other users. Occasionally, you might encounter issues - please report them via the feedback button. Thank you for helping us improve!

### Opt-In vs Automatic Assignment

**Opt-in (Ring 1)**:

- Users choose to join beta program
- Explicit consent to early access
- Users understand and accept risk

**Automatic (Ring 2, Ring 3)**:

- Users automatically moved to new version
- Based on progressive rollout percentage
- Users may not notice

### Compliance Considerations

**Regulated industries**:

- Document ring structure and progression criteria
- Maintain audit trail of ring progressions
- Formal approval before Ring 3 (full GA)
- Risk assessments per ring

---

## Implementation Strategies

### SaaS Web Applications

**Feature flag-based**:

```yaml
# Feature flag configuration
new-feature:
  enabled: true
  rollout:
    - ring: 0
      percentage: 100
      audience: internal_users
    - ring: 1
      percentage: 100
      audience: beta_users
    - ring: 2
      percentage: 75
      audience: all_users
    - ring: 3
      percentage: 100
      audience: all_users
```

**Infrastructure-based** (multiple production deployments):

```text
Production cluster:
- Prod-Ring0: 5% capacity, internal traffic only
- Prod-Ring1: 20% capacity, beta user traffic
- Prod-Ring2: 75% capacity, standard user traffic
- Prod-Ring3: 100% capacity (replaces all above)
```

### Mobile Applications

**App store phased rollout**:

- **Day 1**: Release to internal TestFlight (Ring 0)
- **Day 2**: Release to external TestFlight (Ring 1)
- **Day 3**: 10% app store rollout (Ring 1 continues)
- **Day 5**: 50% app store rollout (Ring 2)
- **Day 10**: 100% app store rollout (Ring 3)

**Server-side feature flags**:

- Deploy mobile app with features disabled
- Enable features via server-side flags by user segment
- Instant rollback by disabling flag (no app redeployment)

### API Services

**Traffic-based rings**:

```text
Load balancer routing:
- Ring 0: Route internal API consumers → new version
- Ring 1: Route 25% external traffic → new version
- Ring 2: Route 75% external traffic → new version
- Ring 3: Route 100% traffic → new version (decommission old)
```

### Hybrid Approach

Many organizations use **hybrid** strategies:

- Ring 0: Infrastructure-based (separate internal environment)
- Ring 1-3: Feature flag-based (same infrastructure, runtime control)

This provides:

- Clear separation for internal testing
- Flexible progressive rollout for external users
- Instant rollback capability via flags

---

## Rings vs Canary

### Similarities

Both are progressive rollout strategies:

- Start with small percentage
- Monitor metrics
- Gradually increase
- Rollback on issues

### Differences

| Aspect           | Canary Deployment                     | Deployment Rings                         |
| ---------------- | ------------------------------------- | ---------------------------------------- |
| **Focus**        | Infrastructure/technical rollout      | User segment/organizational rollout      |
| **Duration**     | Hours (fast progression)              | Days to weeks (deliberate pauses)        |
| **Progression**  | Continuous (1% → 5% → 10% → 50%)      | Discrete rings (Ring 0 → 1 → 2 → 3)     |
| **User segments**| Random users (percentage-based)       | Specific audiences (internal, beta, etc.)|
| **Automation**   | Highly automated                      | Mix of automated and manual approvals    |
| **Organization** | Technical (DevOps-focused)            | Organizational (involves product, users) |
| **Rollback**     | Instant (route 0% to canary)          | Slower (may require app updates)         |

### When to Use Each

**Use Canary when**:

- Technical validation in production
- Fast rollout desired (hours, not days)
- Random user sampling acceptable
- Full automation required

**Use Rings when**:

- Organizational rollout needed (internal first, then external)
- Specific user segments required (beta programs)
- Longer validation periods desired (multi-day stability)
- User communication and consent important

**Use Both**:

- Ring 0: Internal deployment (organizational ring)
- Ring 1-3: Canary rollout within each ring (technical progression)

Example:

1. Deploy to Ring 0 (internal users) - 100% of internal users
2. Deploy to Ring 1 (beta users) - 100% of beta users
3. Deploy to Ring 2 (standard users) - Canary rollout: 1% → 10% → 50% → 100%
4. Deploy to Ring 3 (remaining users) - 100% of remaining users

---

## Anti-Patterns

### Anti-Pattern 1: Skipping Rings

**Problem**: Jumping from Ring 0 directly to Ring 3 (100%)

**Impact**: No incremental validation, all users affected if issues arise

**Solution**: Respect ring progression, validate at each stage

### Anti-Pattern 2: No Ring 0

**Problem**: External users are first to see new code

**Impact**: External reputation damage if critical issues

**Solution**: Always test with internal users first

### Anti-Pattern 3: Too-Fast Progression

**Problem**: Progressing to next ring after 5 minutes

**Impact**: Issues that manifest over time (memory leaks, resource exhaustion) not caught

**Solution**: Respect minimum duration for each ring (hours to days)

### Anti-Pattern 4: Ignoring Ring Metrics

**Problem**: Progressing despite elevated error rates

**Impact**: Problems propagate to larger audiences

**Solution**: Enforce progression criteria, automated halt on threshold breaches

### Anti-Pattern 5: No User Communication

**Problem**: Early ring users don't know they're beta testing

**Impact**: User frustration when encountering issues, poor feedback

**Solution**: Inform users they're in early access, provide feedback channel

---

## Best Practices Summary

1. **Always start with Ring 0**: Internal users first, every time
2. **Define progression criteria**: Objective metrics, not gut feel
3. **Respect minimum durations**: Hours for Ring 0, days for Ring 1-2
4. **Communicate with users**: Inform early rings they're beta testing
5. **Monitor actively**: Watch metrics, don't assume "no news is good news"
6. **Automate progression**: Where possible, use metrics-driven automation
7. **Manual approval for Ring 3**: Final 100% rollout is business decision
8. **Combine with canary**: Use rings for user segments, canary for percentage rollout
9. **Document rollback**: Each ring should have rollback procedure
10. **Build beta program**: Cultivate engaged Ring 1 audience

---

## Next Steps

- [Deployment Strategies](./deployment-strategies.md) - Technical deployment patterns
- [CD Model Stage 11](../cd-model/cd-model-stages-8-12.md#stage-11-live) - Live monitoring during rings
- [CD Model Stage 12](../cd-model/cd-model-stages-8-12.md#stage-12-release-toggling) - Feature flag integration
- [Implementation Patterns](../cd-model/implementation-patterns.md) - RA vs CDe pattern usage

## Quick Reference

### Ring Structure

| Ring | Name                 | Audience                   | Duration | Traffic % |
| ---- | -------------------- | -------------------------- | -------- | --------- |
| 0    | Canary               | Internal users, developers | Hours    | 1-5%      |
| 1    | Early Adopters       | Beta users, opted-in users | 1-2 days | 10-25%    |
| 2    | Standard             | Regular users              | 3-7 days | 50-75%    |
| 3    | General Availability | All users                  | Complete | 100%      |

### Progression Criteria

| From Ring | To Ring | Criteria                                  |
| --------- | ------- | ----------------------------------------- |
| 0         | 1       | No critical errors, metrics stable        |
| 1         | 2       | Error rate < threshold, positive feedback |
| 2         | 3       | All metrics healthy, no regressions       |
