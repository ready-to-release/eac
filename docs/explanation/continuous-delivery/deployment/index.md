# Deployment

## Introduction

Deployment is the process of delivering software from artifacts to running production environments. How you deploy directly impacts risk, downtime, and ability to respond to problems.

The CD Model treats deployment as a deliberate, controlled process with multiple strategies available depending on risk tolerance, infrastructure capabilities, and organizational maturity.

### Deployment vs Release

**Critical distinction**:

- **Deployment** (Stage 10): Code reaches production environment
- **Release** (Stage 12): Features become available to users

Modern CD practices decouple these:

- Deploy code to production (with features disabled via feature flags)
- Release features gradually (enable via flags for specific user segments)

This decoupling enables:

- Continuous deployment without continuous release
- Risk reduction through gradual rollout
- Instant feature disablement without redeployment
- A/B testing and experimentation

### Deployment in the CD Model

**Stage 10: Production Deployment**:

- Deploy artifacts to production infrastructure
- Choose deployment strategy based on risk and infrastructure
- Execute deployment via Deploy Agents (Zone C)
- Verify health checks and smoke tests

**Stage 11: Live Monitoring**:

- Monitor application health in production
- Track key metrics (errors, latency, throughput)
- Automated rollback on threshold breaches
- Manual rollback procedures available

**Stage 12: Release Toggling**:

- Control feature exposure via feature flags
- Gradual rollout to user segments
- Instant disable if issues detected
- Independent of deployment

### Deployment Strategy Selection

Different situations call for different strategies:

**Hot Deploy**: Fast, simple, brief downtime acceptable

**Rolling**: Zero downtime, gradual instance updates

**Blue-Green**: Instant rollback, higher infrastructure cost

**Canary**: Gradual production validation, metrics-driven

**Rings**: Progressive user group rollout, organizational-scale

**Feature Flags**: Decouple deployment from release, runtime control

The right strategy depends on:

- Risk tolerance
- Infrastructure capabilities
- Downtime requirements
- Team maturity
- Cost constraints

---

## In This Section

| Topic                                             | Description                                                                    |
| ------------------------------------------------- | ------------------------------------------------------------------------------ |
| [Deployment Strategies](./deployment-strategies.md) | Comprehensive explanation of deployment patterns with tradeoffs and use cases |
| [Deployment Rings](./deployment-rings.md)          | Progressive user group rollout pattern explained in depth                     |
| [Feature Flags](./feature-flags.md)                | Feature flag lifecycle management from creation to cleanup                    |
| [Rollback Procedures](./rollback-procedures.md)    | Emergency rollback execution guide with decision criteria and procedures      |
| [Incident Response](./incident-response.md)        | Production incident response playbook with severity levels and templates      |

---

## Next Steps

- [Deployment Strategies](./deployment-strategies.md) - Deep dive into each pattern
- [CD Model Stages 7-12](../cd-model/cd-model-stages-7-12.md) - See deployment in CD Model context
- [Environments Architecture](../architecture/environments.md) - Environment types and network zones

## Quick Reference

- [Deployment Strategies Reference](./deployment-strategies.md) - Strategy comparison table
- [Deployment Rings Reference](./deployment-rings.md) - Ring structure table

{{ diataxis_footer() }}