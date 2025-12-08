<!-- EDITOR
# Editor: reference/continuous-delivery/ra-stage-breakdown.md

## Soul

Release Approval pattern 12-stage reference with manual approval gates at Stage 3 and 9, typical cycle time 1-2 weeks.

## Sections

1. Stage Summary
2. Key Characteristics
3. Approval Gates
4. Related
-->

# Release Approval (RA) Stage Breakdown

Stage-by-stage reference for the Release Approval implementation pattern.

## Stage Summary

| Stage                     | Automation Level          | Approval Required     | Duration          |
| ------------------------- | ------------------------- | --------------------- | ----------------- |
| 1. Authoring              | Manual                    | No                    | hours to days     |
| 2. Pre-commit             | Automated                 | No                    | 5-10 min          |
| 3. Merge Request          | Automated + Manual Review | Yes (Peer)            | hours             |
| 4. Commit                 | Automated                 | No                    | 5-10 min          |
| 5. Acceptance Testing     | Automated                 | No                    | minutes to 1 hour |
| 6. Extended Testing       | Automated + Manual        | No                    | hours             |
| 7. Exploration            | Manual                    | No                    | continuous        |
| 8. Start Release          | Automated                 | No                    | minutes           |
| 9. Release Approval       | Manual Review             | Yes (Release Manager) | hours to days     |
| 10. Production Deployment | Automated                 | No                    | minutes           |
| 11. Live                  | Automated Monitoring      | No                    | Ongoing           |
| 12. Release Toggling      | Manual Control            | No                    | As needed         |

## Key Characteristics

- **Typical cycle time**: 1-2 weeks from commit to production
- **Manual approval gates**: Stage 3 (Peer) and Stage 9 (Release Manager)
- **Primary constraints**: Authoring time and Release Approval queue

## Approval Gates

| Gate         | Stage    | Approver        | Purpose              |
| ------------ | -------- | --------------- | -------------------- |
| First-level  | Stage 3  | Peer reviewer   | Code quality, design |
| Second-level | Stage 9  | Release manager | Production readiness |
| Third-level  | Stage 12 | Feature owner   | Feature exposure     |

## Related

- [CDe Stage Breakdown](cde-stage-breakdown.md)
- [Implementation Patterns](../../explanation/continuous-delivery/cd-model/implementation-patterns.md)
