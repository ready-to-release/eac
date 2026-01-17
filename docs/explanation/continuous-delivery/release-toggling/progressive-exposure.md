# Progressive Exposure

Progressive exposure releases features to increasingly larger user groups over time.

This is a **Stage 12 (Release Toggling)** concern - controlling who sees features, not how code is deployed.

---

## The Four-Ring Model

| Ring | Name                 | Audience                   | Duration | Exposure |
| ---- | -------------------- | -------------------------- | -------- | -------- |
| 0    | Canary               | Internal users, developers | Hours    | 1-5%     |
| 1    | Early Adopters       | Beta users, opted-in       | 1-2 days | 10-25%   |
| 2    | Standard             | Regular users              | 3-7 days | 50-75%   |
| 3    | General Availability | All users                  | Complete | 100%     |

---

## Ring Definitions

### Ring 0: Canary

Internal users first. Developers can immediately debug issues.

**Validates:** Basic functionality, integrations, health checks.

### Ring 1: Early Adopters

Beta program participants who accept risk and provide feedback.

**Validates:** Edge cases, varied usage patterns, UX.

### Ring 2: Standard

Majority of users. Still have 25-50% on previous behavior for comparison.

**Validates:** Long-running stability, business metrics at scale.

### Ring 3: General Availability

Complete rollout after confidence built through Ring 0-2 validation.

---

## Progression Criteria

| Transition   | Criteria                                           | Ex. SLI/SLA Wait Time |
| ------------ | -------------------------------------------------- | --------------------- |
| Ring 0 → 1   | No critical errors, internal users confirm working | 1-4 hours             |
| Ring 1 → 2   | Error rate < 0.5%, positive user feedback          | 24-48 hours           |
| Ring 2 → 3   | All metrics healthy, no regressions                | 3-7 days              |

---

## Progressive Exposure vs Canary Deployment

| Aspect      | Canary Deployment (Stage 10)   | Progressive Exposure (Stage 12) |
| ----------- | ------------------------------ | ------------------------------- |
| Controls    | Infrastructure/traffic routing | Feature visibility              |
| Mechanism   | Load balancer, service mesh    | Feature flags                   |
| Granularity | Random percentage              | Specific user segments          |
| Scope       | Entire application version     | Individual features             |

**Use canary deployment:** Validate infrastructure changes, new service versions.

**Use progressive exposure:** Validate feature changes, user experience.

---

## Anti-Patterns

| Anti-Pattern         | Problem                         | Solution                     |
| -------------------- | ------------------------------- | ---------------------------- |
| Skipping rings       | No incremental validation       | Respect ring progression     |
| No Ring 0            | External users see issues first | Always test internally first |
| Too-fast progression | Miss time-based issues          | Respect minimum durations    |

---

## External Resources

- [Microsoft Safe Deployment Practices](https://learn.microsoft.com/en-us/devops/operate/safe-deployment-practices)
- [Progressive Exposure (Microsoft)](https://learn.microsoft.com/en-us/azure/devops/migrate/phase-rollout-with-rings)

---

## Next Steps

- [Feature Flags](./feature-flags.md) - Runtime feature control
- [CD Model Stage 12](../cd-model/stages.md#stage-12-release-toggling) - Release toggling in context
- [Deployment Strategies](../deployment/deployment-strategies.md) - Stage 10 infrastructure patterns
