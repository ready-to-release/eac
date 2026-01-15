# Deployment

Deployment is the process of delivering software from artifacts to production environments (Stage 10).

## Deployment vs Release

| Concern    | Stage | Question                        | Covered In                                       |
| ---------- | ----- | ------------------------------- | ------------------------------------------------ |
| Deployment | 10    | How does code reach production? | This section                                     |
| Release    | 12    | How do features reach users?    | [Release Toggling](../release-toggling/index.md) |

---

## In This Section

| Topic                                               | Description                             |
| --------------------------------------------------- | --------------------------------------- |
| [Deployment Strategies](./deployment-strategies.md) | Hot Deploy, Rolling, Blue-Green, Canary |
| [Rollback Procedures](./rollback-procedures.md)     | Emergency rollback execution            |
| [Incident Response](./incident-response.md)         | Production incident handling            |

---

## Quick Reference

| Strategy   | Downtime | Rollback | Use Case              |
| ---------- | -------- | -------- | --------------------- |
| Hot Deploy | Brief    | Minutes  | Internal tools        |
| Rolling    | None     | Minutes  | Standard services     |
| Blue-Green | None     | Instant  | High-risk releases    |
| Canary     | None     | Instant  | Production validation |

---

## Next Steps

- [Deployment Strategies](./deployment-strategies.md) - Choose the right pattern
- [Release Toggling](../release-toggling/index.md) - Feature flags and progressive exposure
- [CD Model Stages 8-12](../cd-model/stages.md#release-stages) - Deployment in context
