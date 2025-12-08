# Deployment Strategies

Quick reference for production deployment strategy selection.

## Strategy Comparison

| Strategy   | Rollback Speed | Downtime      | Complexity | Resource Cost |
| ---------- | -------------- | ------------- | ---------- | ------------- |
| Hot Deploy | Fast (1-2 min) | Brief (< 30s) | Low        | Low           |
| Rolling    | Fast (< 1 min) | None          | Medium     | Low           |
| Blue-Green | Instant        | None          | Medium     | High (2x)     |
| Canary     | Instant        | None          | High       | Medium        |
| Rings      | Gradual        | None          | High       | Medium        |

## Pattern Recommendations

**For RA pattern**: Blue-Green or Rolling recommended

- Manual approval before full rollout
- Clear cut-over point

**For CDe pattern**: Canary or Feature Flags recommended

- Automated with gradual validation
- Metrics-driven progression

## Strategy Details

### Hot Deploy

- Replace running instances directly
- Fastest deployment (seconds)
- Acceptable for brief downtime (< 30s)

### Rolling

- Update instances one/few at a time
- Zero downtime
- Built-in health checks stop rollout on issues

### Blue-Green

- Two identical environments
- Instant traffic switch
- Instant rollback (switch back)
- 2x resource cost

### Canary

- Route 1-5% traffic to new version
- Monitor metrics, gradually increase
- Instant rollback to 0%

### Rings

- Deploy to user groups progressively
- Ring 0: Internal users (hours)
- Ring 1: Early adopters (1-2 days)
- Ring 2: Standard users (3-7 days)
- Ring 3: All users

## Related

- [Environments](../../explanation/continuous-delivery/architecture/environments.md)
- [CD Model Stages 7-12](../../explanation/continuous-delivery/cd-model/cd-model-stages-7-12.md)
