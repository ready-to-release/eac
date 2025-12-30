# Release Toggling

Release toggling controls when features are made available to users **after** code is deployed.

## Deployment vs Release

| Concern    | Stage | Question                        | Tools                               |
| ---------- | ----- | ------------------------------- | ----------------------------------- |
| Deployment | 10    | How does code reach production? | Rolling, Blue-Green, Canary         |
| Release    | 12    | How do features reach users?    | Feature flags, progressive exposure |

**Key principle:** Deploy code with features disabled, enable gradually via runtime configuration.

---

## In This Section

| Topic                                             | Description                                    |
| ------------------------------------------------- | ---------------------------------------------- |
| [Feature Flags](./feature-flags.md)               | Runtime feature control and gradual enablement |
| [Progressive Exposure](./progressive-exposure.md) | Rollout to user segments (rings)               |

---

## Why Separate Deployment from Release?

| Benefit         | Explanation                                     |
| --------------- | ----------------------------------------------- |
| Faster deploys  | Ship code without waiting for feature readiness |
| Safer releases  | Enable for 1% of users, not 100%                |
| Instant disable | Turn off feature without redeployment           |
| A/B testing     | Compare feature variants in production          |

---

## Next Steps

- [Feature Flags](./feature-flags.md) - Start here for runtime feature control
- [CD Model Stage 12](../cd-model/stages.md#stage-12-release-toggling) - Release toggling in context
- [Deployment Strategies](../deployment/deployment-strategies.md) - Stage 10 infrastructure patterns
