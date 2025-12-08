<!-- EDITOR
# Editor: reference/continuous-delivery/cde-stage-breakdown.md

## Soul

Continuous Deployment pattern 12-stage reference with automated progression, feature flags required, typical cycle time 2-4 hours.

## Sections

1. Stage Summary
2. Key Characteristics
3. Approval Gates
4. Note
5. Related
-->

# Continuous Deployment (CDe) Stage Breakdown

Stage-by-stage reference for the Continuous Deployment implementation pattern.

## Stage Summary

| Stage                     | Automation Level          | Approval Required  | Duration          |
| ------------------------- | ------------------------- | ------------------ | ----------------- |
| 1. Authoring              | Manual                    | No                 | hours to days     |
| 2. Pre-commit             | Automated                 | No                 | 5-10 min          |
| 3. Merge Request          | Automated + Manual Review | Yes (Peer)         | hours             |
| 4. Commit                 | Automated                 | No                 | 5-10 min          |
| 5. Acceptance Testing     | Automated                 | No                 | minutes to 1 hour |
| 6. Extended Testing       | Automated                 | No                 | hours             |
| 7. Exploration            | Automated (or skipped)    | No                 | continuous        |
| 8. Start Release          | Automated                 | No                 | seconds           |
| 9. Release Approval       | Automated                 | No (Auto-approved) | seconds           |
| 10. Production Deployment | Automated                 | No                 | minutes           |
| 11. Live                  | Automated Monitoring      | No                 | Ongoing           |
| 12. Release Toggling      | Automated Control         | No                 | Real-time         |

## Key Characteristics

- **Typical cycle time**: 2-4 hours from commit to production
- **Single approval gate**: Stage 3 serves as both first-level and second-level sign-off
- **Feature flags**: Required for runtime control of feature exposure

## Approval Gates

| Gate | Stage | Approver | Purpose |
|------|-------|----------|---------|
| Combined first/second-level | Stage 3 | Peer reviewer | Code quality + production approval |
| Third-level | Stage 12 | Feature owner | Feature exposure (via flags) |

## Note

In CDe, Stage 3 approval also approves production deployment. By merging, the reviewer approves the change going to production.

## Related

- [RA Stage Breakdown](ra-stage-breakdown.md)
- [Implementation Patterns](../../explanation/continuous-delivery/cd-model/implementation-patterns.md)
