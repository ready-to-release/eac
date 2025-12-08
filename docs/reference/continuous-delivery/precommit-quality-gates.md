<!-- EDITOR
# Editor: reference/continuous-delivery/precommit-quality-gates.md

## Soul

Stage 2 quality gate specifications with 5-10 minute target including formatting, linting, unit tests, and security scans.

## Sections

1. Gate Summary
2. Time Budget
3. Environment
4. Related
-->

# Pre-commit Quality Gates

Quality gates for Stage 2 (Pre-commit) validation.

## Gate Summary

| Check                   | Pass Criteria                     | Failure Action           |
| ----------------------- | --------------------------------- | ------------------------ |
| Code formatting         | All files formatted per standards | Auto-fix or block commit |
| Linting                 | No high-severity issues           | Block commit             |
| Unit tests              | 100% passing, minimum coverage    | Block commit             |
| Security - secrets      | No hardcoded secrets detected     | Block commit             |
| Security - dependencies | No critical vulnerabilities       | Block commit             |
| Build                   | Successful compilation            | Block commit             |

## Time Budget

**Target**: 5-10 minutes maximum

Strategies to stay within budget:

- Incremental scanning (only changed files)
- Local caching (reuse previous results)
- Fail fast (stop on first critical failure)

## Environment

Pre-commit runs in:

- **DevBox**: Local developer machine
- **Build Agents**: CI/CD pipeline for PR validation

## Related

- [Merge Request Quality Gates](merge-request-quality-gates.md)
- [CD Model Stages 1-6](../../explanation/continuous-delivery/cd-model/cd-model-stages-1-6.md)
