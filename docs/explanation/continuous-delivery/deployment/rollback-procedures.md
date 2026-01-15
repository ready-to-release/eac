# Rollback Procedures

How to plan and execute production rollbacks when deployments fail (Stage 10).

For disabling features via flags, see [Feature Flags](../release-toggling/feature-flags.md) (Stage 12) - that's not a deployment rollback, just turning something off again.

---

## When to Roll Back

| Trigger                    | Action                    | Ex. SLA/SLI Decision Time |
| -------------------------- | ------------------------- | ------------------------- |
| Error rate > 1%            | Immediate rollback        | < 2 min                   |
| P95 latency > 2x baseline  | Immediate rollback        | < 2 min                   |
| Health checks failing      | Immediate rollback        | < 1 min                   |
| Critical bug discovered    | Immediate rollback        | < 5 min                   |
| Business metrics down 10%+ | Evaluate, likely rollback | < 15 min                  |

---

## Rollback by Strategy

| Strategy   | Rollback Method                             | Time    |
| ---------- | ------------------------------------------- | ------- |
| Blue-Green | Switch traffic back to previous environment | Seconds |
| Rolling    | `kubectl rollout undo deployment/app`       | Minutes |
| Canary     | Set canary weight to 0%                     | Seconds |

---

## Database Rollback

| Schema Change              | Rollback Approach                        |
| -------------------------- | ---------------------------------------- |
| Additive (new columns)     | Safe - previous code ignores new columns |
| Destructive (drop columns) | Requires data restore                    |
| Transformative (rename)    | Requires backward-compatible migration   |

**Best Practice:** Use expand-contract migrations:

1. **Expand:** Add new column, keep old
2. **Deploy:** New code uses both
3. **Migrate:** Copy data
4. **Contract:** Remove old column

---

## Rollback Time Objectives

An example of a team's objectives:

| Environment           | Ex. SLA/SLI Target RTO |
| --------------------- | ---------------------- |
| Production (critical) | < 5 minutes            |
| Production (standard) | < 15 minutes           |
| Staging               | < 30 minutes           |

---

## Post-Rollback Actions

1. **Stabilize** - Confirm system is healthy
2. **Investigate** - Root cause analysis
3. **Document** - Record incident details
4. **Fix** - Address root cause
5. **Retest** - Validate fix in pre-production
6. **Redeploy** - When ready and tested

---

## Practice Rollbacks

Schedule regular rollback drills:

- **Monthly:** Execute rollback in staging
- **Quarterly:** Execute rollback in production (during low-traffic)
- Document learnings and update procedures

---

## External Resources

- [Kubernetes Rollout Undo](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-back-a-deployment)
- [AWS Blue/Green Rollback](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/using-features.rolling-version-deploy.html)

---

## Next Steps

- [Deployment Strategies](./deployment-strategies.md) - Strategy-specific rollback patterns
- [Incident Response](./incident-response.md) - Handle production incidents
