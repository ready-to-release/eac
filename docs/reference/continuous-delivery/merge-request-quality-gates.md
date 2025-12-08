# Merge Request Quality Gates

Quality gates for Stage 3 (Merge Request) validation.

## Gate Summary

| Gate               | Requirement                                |
| ------------------ | ------------------------------------------ |
| Peer approval      | At least one approver (configurable)       |
| Automated tests    | All tests passing                          |
| Code coverage      | Meets minimum threshold (e.g., 80%)        |
| No merge conflicts | Branch is up-to-date with target           |
| Security scans     | No critical vulnerabilities                |
| Required reviewers | Domain experts have approved (if required) |

## Automated Checks

Run in parallel with peer review:

- All pre-commit checks (repeated in CI/CD)
- Integration tests
- SAST security scans
- Dependency scanning
- Container image scanning (if applicable)
- Documentation builds
- Preview deployments (optional)

## Compliance Artifacts

In regulated environments, Stage 3 may include:

- Change control documentation
- Risk assessments
- Test plans and evidence
- Security reviews

## Environment

Stage 3 executes on **Build Agents** - dedicated CI/CD pipeline runners.

## Related

- [Pre-commit Quality Gates](precommit-quality-gates.md)
- [Release Quality Thresholds](release-quality-thresholds.md)
- [CD Model Stages 1-6](../../explanation/continuous-delivery/cd-model/cd-model-stages-1-6.md)
