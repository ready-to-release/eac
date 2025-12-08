# Release Quality Thresholds

Automated quality thresholds for Stage 9 (Release Approval) validation.

## Threshold Summary

| Metric                                   | Threshold |
| ---------------------------------------- | --------- |
| Test pass rate                           | 100%      |
| Code coverage                            | ≥ 80%     |
| Critical bugs                            | 0         |
| High bugs                                | 0         |
| Performance regression                   | < 5%      |
| Security vulnerabilities (Critical/High) | 0         |

## Documentation Requirements

Required documentation for release approval:

- Release notes
- Deployment runbook
- Rollback procedure
- Test evidence
- Security scan reports
- Performance test results
- Stakeholder sign-offs
- Risk assessment

## Approval Modes

**RA Pattern (Manual)**:

- Release manager reviews evidence
- Formal approval documented
- May take hours to days

**CDe Pattern (Automated)**:

- Quality gates auto-evaluate
- Pipeline proceeds if all pass
- Takes seconds

## Environment

Validation occurs in:

- **PLTE**: Automated evidence collection
- **Demo**: Exploratory testing evidence

## Related

- [Pre-commit Quality Gates](precommit-quality-gates.md)
- [Merge Request Quality Gates](merge-request-quality-gates.md)
- [Live Monitoring](../../explanation/continuous-delivery/cd-model/cd-model-stages-7-12.md#stage-11-live)
