# Deployment Rings

Reference for ring-based deployment rollout strategy.

## Ring Structure

| Ring | Name                 | Audience                   | Duration | Traffic % |
| ---- | -------------------- | -------------------------- | -------- | --------- |
| 0    | Canary               | Internal users, developers | Hours    | 1-5%      |
| 1    | Early Adopters       | Beta users, opted-in users | 1-2 days | 10-25%    |
| 2    | Standard             | Regular users              | 3-7 days | 50-75%    |
| 3    | General Availability | All users                  | Complete | 100%      |

## Ring 0 - Canary

**Purpose:** Early warning detection with minimal blast radius.

| Attribute          | Value                      |
| ------------------ | -------------------------- |
| Audience           | Internal users, developers |
| Traffic percentage | 1-5%                       |
| Monitoring period  | 1-4 hours                  |
| Rollback trigger   | Any critical issue         |

**Validation Focus:**

- Basic functionality works
- No immediate crashes or errors
- Key integrations functioning

## Ring 1 - Early Adopters

**Purpose:** Broader validation with users who accept some risk.

| Attribute          | Value                    |
| ------------------ | ------------------------ |
| Audience           | Beta users, early access |
| Traffic percentage | 10-25%                   |
| Monitoring period  | 24 hours                 |
| Rollback trigger   | Error rate > threshold   |

**Validation Focus:**

- Edge cases and varied usage patterns
- Performance under broader load
- User feedback collection

## Ring 2 - Standard Users

**Purpose:** Majority rollout with continued monitoring.

| Attribute          | Value                  |
| ------------------ | ---------------------- |
| Audience           | Regular users          |
| Traffic percentage | 50-75%                 |
| Monitoring period  | 3-7 days               |
| Rollback trigger   | Significant regression |

**Validation Focus:**

- Full production load handling
- Business metrics validation
- Long-running stability

## Ring 3 - General Availability

**Purpose:** Complete rollout to all users.

| Attribute          | Value                |
| ------------------ | -------------------- |
| Audience           | All users            |
| Traffic percentage | 100%                 |
| Monitoring period  | Ongoing              |
| Rollback trigger   | Critical issues only |

## Progression Criteria

| From Ring | To Ring                                   | Criteria |
| --------- | ----------------------------------------- | -------- |
| 0 → 1     | No critical errors, metrics stable        |          |
| 1 → 2     | Error rate < threshold, positive feedback |          |
| 2 → 3     | All metrics healthy, no regressions       |          |

## When to Use Rings

**Use rings when:**

- Large user base
- Need production feedback before full rollout
- Compliance requires phased approach
- Different user segments exist

**Consider alternatives when:**

- Small user base
- Low-risk changes
- Time-critical deployments

## Related

- [Deployment Strategies](deployment-strategies.md)
- [Live Monitoring](../../explanation/continuous-delivery/cd-model/cd-model-stages-7-12.md#stage-11-live)
- [Feature Flags](../../explanation/continuous-delivery/cd-model/cd-model-stages-7-12.md#stage-12-release-toggling)
