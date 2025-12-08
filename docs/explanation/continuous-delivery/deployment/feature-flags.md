# Feature Flags

How to implement, operate, and retire feature flags.

## Flag Lifecycle

```
Create → Enable (gradual) → Stabilize → Remove
```

| Phase | Duration | Actions |
|-------|----------|---------|
| Create | Day 0 | Add flag, deploy code OFF |
| Enable | Days 1-14 | Gradual rollout (10% → 50% → 100%) |
| Stabilize | Days 15-30 | Monitor, collect feedback |
| Remove | Day 30-90 | Delete flag and old code paths |

## Creating a Flag

### Step 1: Define the Flag

```yaml
# flags.yaml
features:
  new-checkout:
    description: "New checkout flow with one-click purchase"
    owner: "checkout-team"
    created: "2024-01-15"
    cleanup_deadline: "2024-04-15"
    enabled: false
```

### Step 2: Implement in Code

```go
func processCheckout(user User) {
    if featureFlags.IsEnabled("new-checkout", user) {
        newCheckoutFlow(user)
    } else {
        legacyCheckoutFlow(user)
    }
}
```

### Step 3: Deploy

Deploy with flag OFF. Code reaches production but feature is not active.

## Enabling a Flag (Rollout)

### Simple Boolean

```yaml
new-checkout:
  enabled: true
```

### Percentage Rollout

```yaml
new-checkout:
  enabled: true
  rollout_percentage: 10  # Start with 10%
```

Increase gradually:

| Day | Percentage | Validation |
|-----|------------|------------|
| 1 | 10% | Monitor errors |
| 3 | 25% | Check business metrics |
| 7 | 50% | Validate performance |
| 14 | 100% | Full rollout |

### User-Targeted Rollout

```yaml
new-checkout:
  enabled_for_users:
    - "beta@example.com"
    - "tester@example.com"
  enabled_for_groups:
    - "beta-testers"
    - "internal-users"
```

### Time-Based Flag

```yaml
black-friday-sale:
  enabled_after: "2024-11-29T00:00:00Z"
  enabled_before: "2024-12-02T00:00:00Z"
```

## Monitoring Flags

Track key metrics per flag:

| Metric | Flag OFF | Flag ON | Action |
|--------|----------|---------|--------|
| Error rate | 0.1% | 0.5% | Investigate |
| Latency P95 | 100ms | 150ms | Acceptable |
| Conversion | 2.5% | 2.8% | Success! |

## Emergency Disable

When issues occur:

```bash
# Instant disable
curl -X PUT https://flags.example.com/api/flags/new-checkout \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"enabled": false}'
```

Or through UI dashboard.

## Removing Flags

### When to Remove

- Feature stable for 30+ days
- 100% rollout complete
- No issues or rollbacks needed
- Business metrics validated

### Cleanup Steps

1. **Verify stability** - No issues for 30 days
2. **Remove old code path** - Delete legacy implementation
3. **Remove flag checks** - Delete conditional logic
4. **Remove flag definition** - Delete from configuration
5. **Deploy cleanup** - Single commit removing all traces

### Example Cleanup Commit

```text
chore(checkout): remove new-checkout feature flag

Feature has been stable at 100% for 45 days.
Removing flag and legacy checkout code.

- Remove feature flag checks from checkout.go
- Delete legacy_checkout.go
- Remove flag from flags.yaml
- Update tests

Relates to #456
```

## Flag Hygiene

### Track Flag Age

| Flag | Created | Age | Status |
|------|---------|-----|--------|
| new-checkout | 2024-01-15 | 45 days | Ready for cleanup |
| beta-dashboard | 2024-02-01 | 30 days | Monitoring |
| experiment-a | 2023-10-01 | 120 days | OVERDUE |

### Enforce Cleanup Deadlines

```yaml
new-checkout:
  cleanup_deadline: "2024-04-15"
  owner: "checkout-team"
```

Alert when deadline approaches or passes.

## Anti-Patterns to Avoid

| Anti-Pattern | Problem | Solution |
|--------------|---------|----------|
| Nested flags | Complex logic, hard to test | Flatten or combine |
| Permanent flags | Technical debt accumulates | Enforce cleanup deadlines |
| Missing owners | Nobody responsible | Require owner in definition |
| No monitoring | Can't detect issues | Add metrics per flag |

## Next Steps

- [Release Toggling](../cd-model/cd-model-stages-7-12.md#stage-12-release-toggling) - Control feature exposure at runtime
- [Deployment Rings](./deployment-rings.md) - Progressive rollout to user groups
- [Rollback Procedures](./rollback-procedures.md) - Emergency response procedures
- [Deployment Strategies](./deployment-strategies.md) - Decoupling deployment from release

{{ diataxis_footer() }}
