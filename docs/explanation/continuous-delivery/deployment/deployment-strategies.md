# Deployment Strategies

A deployment strategy determines how new code reaches production (Stage 10) and what happens if problems are detected.

For controlling when features reach users, see [Release Toggling](../release-toggling/index.md) (Stage 12).

The following diagram compares the four primary deployment strategies, showing their processes, risk levels, and when to use each approach.

![Deployment Strategies Overview](../../../assets/assisted/09-deployment-strategies.drawio.png){width=1000}

---

## Strategy Comparison

| Strategy   | Downtime     | Rollback | Complexity | Cost      | Best For                   |
| ---------- | ------------ | -------- | ---------- | --------- | -------------------------- |
| Hot Deploy | Brief (<30s) | 1-2 min  | Low        | Low       | Internal tools, batch jobs |
| Rolling    | None         | ~5 min   | Medium     | Low       | Standard web services      |
| Blue-Green | None         | Instant  | Medium     | High (2x) | High-risk deployments      |
| Canary     | None         | Instant  | High       | Medium    | Production validation      |

---

## Hot Deploy

Replaces running instances directly with brief service interruption.

**Process:** Stop instance → Replace files → Start instance → Verify health

**Use when:**

- Single-instance applications
- Brief downtime acceptable
- Fast-starting applications (<10s startup)
- Simple infrastructure

**Avoid when:** Zero downtime required, stateful sessions, high availability needed.

**Variants:** Prepare files in staging folder, switch host binding to minimize downtime.

---

## Rolling Deployment

Updates instances gradually while others continue serving traffic.

**Process:** For each instance: Remove from LB → Update → Health check → Add back to LB

**Use when:**

- Multiple instances behind load balancer
- Zero downtime required
- Changes are backward-compatible

**Avoid when:** Single instance, breaking changes, database migrations requiring same version.

---

## Blue-Green Deployment

Maintains two identical environments. Deploy to inactive, test, switch traffic.

**Process:** Deploy to Green → Test Green → Switch traffic Blue→Green → Blue becomes idle

**Use when:**

- Instant rollback critical
- Large artifacts or slow startup
- Database migrations need full testing
- High-stakes releases

**Tradeoffs:** Higher infrastructure cost (2x for some licensing), database state management complexity.

---

## Canary Deployment

Routes small percentage of traffic to new version, monitors metrics, gradually increases.

**Process:** Deploy canary (1-5%) → Monitor metrics → Increase (10%→25%→50%→100%) or rollback

**Use when:**

- Need production validation before full rollout
- Have robust monitoring and metrics
- Large user base (1% is meaningful sample)
- Risk-averse organizations

**Avoid when:** Small user base, poor observability, breaking changes.

**Key metrics:** Error rate, P95 latency, conversion rate, user engagement.

---

## Decision Framework

**Can you tolerate any downtime?**

- No → Rolling, Blue-Green, or Canary
- Yes (<30s) → Hot Deploy

**Need instant rollback?**

- Yes → Blue-Green or Canary
- No → Rolling or Hot Deploy

**Have robust monitoring?**

- Yes → Canary (metrics-driven)
- No → Blue-Green (instant switch)

---

## External Resources

- [Kubernetes Rolling Updates](https://kubernetes.io/docs/tutorials/kubernetes-basics/update/update-intro/)
- [AWS Blue-Green Deployments](https://docs.aws.amazon.com/whitepapers/latest/overview-deployment-options/bluegreen-deployments.html)
- [Flagger - Canary Deployments](https://flagger.app/)
- [Argo Rollouts](https://argoproj.github.io/rollouts/)

---

## Next Steps

- [Rollback Procedures](./rollback-procedures.md) - Emergency response
- [Release Toggling](../release-toggling/index.md) - Feature flags and progressive exposure (Stage 12)
- [CD Variants](../cd-model/variants.md) - RA vs CDe patterns
