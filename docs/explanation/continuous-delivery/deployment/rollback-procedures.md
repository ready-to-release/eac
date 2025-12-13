# Rollback Procedures

{{ page_breadcrumb() }}

How to plan and execute production rollbacks.

## When to Roll Back

| Trigger                    | Action                    | Decision Time |
| -------------------------- | ------------------------- | ------------- |
| Error rate > 1%            | Immediate rollback        | < 2 min       |
| P95 latency > 2x baseline  | Immediate rollback        | < 2 min       |
| Health checks failing      | Immediate rollback        | < 1 min       |
| Critical bug discovered    | Immediate rollback        | < 5 min       |
| Business metrics down 10%+ | Evaluate, likely rollback | < 15 min      |

## Pre-Deployment Checklist

Before every deployment, ensure:

- [ ] Previous version artifacts available
- [ ] Rollback procedure documented
- [ ] Database rollback plan ready (if schema changed)
- [ ] Monitoring dashboards visible
- [ ] On-call team notified
- [ ] Communication channels ready

## Rollback Steps

### Step 1: Detect Issue

```bash
# Check error rates
curl -s https://monitoring.example.com/api/errors | jq '.rate'

# Check health endpoints
curl -s https://api.example.com/health
```

### Step 2: Decide to Roll Back

Decision criteria:

- Error rate exceeds threshold
- Latency exceeds threshold
- Health checks failing
- Critical user impact

### Step 3: Execute Rollback

**For Blue-Green:**

```bash
# Switch traffic back to blue
kubectl patch service app \
  -p '{"spec":{"selector":{"version":"v1.1.0"}}}'
```

**For Rolling Deployment:**

```bash
# Roll back to previous version
kubectl rollout undo deployment/app
```

**For Canary:**

```bash
# Set canary weight to 0
kubectl patch canary app \
  -p '{"spec":{"canaryWeight":0}}'
```

### Step 4: Verify Rollback

```bash
# Check pods are running previous version
kubectl get pods -o jsonpath='{.items[*].spec.containers[0].image}'

# Verify health
curl -s https://api.example.com/health

# Check error rates returning to normal
watch 'curl -s https://monitoring.example.com/api/errors | jq .rate'
```

### Step 5: Communicate

- Update status page
- Notify stakeholders
- Alert support team
- Document incident

## Database Rollback Considerations

| Schema Change              | Rollback Approach                        |
| -------------------------- | ---------------------------------------- |
| Additive (new columns)     | Safe - previous code ignores new columns |
| Destructive (drop columns) | Requires data restore                    |
| Transformative (rename)    | Requires backward-compatible migration   |

**Best Practice:** Use expand-contract migrations

1. **Expand:** Add new column, keep old
2. **Deploy:** New code uses both
3. **Migrate:** Copy data from old to new
4. **Contract:** Remove old column

## Feature Flag Rollback

Fastest rollback when using feature flags:

```bash
# Disable feature instantly
curl -X PUT https://flags.example.com/api/flags/new-checkout \
  -d '{"enabled": false}'
```

No redeployment required.

## Post-Rollback Actions

1. **Stabilize** - Confirm system is healthy
2. **Investigate** - Root cause analysis
3. **Document** - Record incident details
4. **Fix** - Address root cause
5. **Retest** - Validate fix in pre-production
6. **Redeploy** - When ready and tested

## Rollback Time Objectives

| Environment           | Target RTO   |
| --------------------- | ------------ |
| Production (critical) | < 5 minutes  |
| Production (standard) | < 15 minutes |
| Staging               | < 30 minutes |

## Practice Rollbacks

Schedule regular rollback drills:

- Monthly: Execute rollback in staging
- Quarterly: Execute rollback in production (during low-traffic)
- Document learnings and update procedures

## Next Steps

- [Deployment Strategies](./deployment-strategies.md) - Choose the right deployment approach
- [Production Deployment](../cd-model/cd-model-stages-8-12.md#stage-10-production-deployment) - CD Model Stage 10
- [Incident Response](./incident-response.md) - Handle production incidents
- [Feature Flags](./feature-flags.md) - Instant rollback via feature toggles

{{ diataxis_footer() }}
