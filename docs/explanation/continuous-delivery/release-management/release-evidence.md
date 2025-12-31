# Release Evidence

## Overview

Release evidence is the **proof of quality** demonstrating a release meets production readiness criteria. This evidence is collected throughout the pipeline and reviewed at Stage 9 (Release Approval).

---

## Evidence Types

### Test Execution Reports

Evidence from Stages 2-6 proving code works correctly.

| Evidence                             | Source      | Format          |
| ------------------------------------ | ----------- | --------------- |
| Unit test results                    | CI pipeline | JUnit XML, HTML |
| Integration test results             | CI pipeline | JUnit XML       |
| Acceptance test results (IV, OV, PV) | PLTE        | JUnit XML, HTML |
| Code coverage reports                | CI pipeline | Cobertura, HTML |

### Security Scan Results

Evidence from Stages 2, 3, and 6 proving code is secure.

| Evidence                   | Source         | Format      |
| -------------------------- | -------------- | ----------- |
| SAST results               | Semgrep, Gosec | SARIF, HTML |
| Dependency vulnerabilities | Trivy          | JSON, HTML  |
| Container image scans      | Trivy          | JSON        |
| DAST results               | OWASP ZAP      | HTML, JSON  |
| Compliance checks          | Audit tools    | PDF, HTML   |

### Performance Metrics

Evidence from Stage 6 proving code performs acceptably.

| Evidence                    | Source          | Format        |
| --------------------------- | --------------- | ------------- |
| Load test results           | JMeter, Gatling | HTML, CSV     |
| Response time distributions | Load testing    | P50, P95, P99 |
| Throughput metrics          | Load testing    | Requests/sec  |
| Resource utilization        | Monitoring      | CPU, memory   |
| Regression analysis         | Comparison      | Before/after  |

---

## Approval Checklist

### Quality Metrics

| Metric                        | Threshold | Stage |
| ----------------------------- | --------- | ----- |
| Test pass rate                | 100%      | 4-6   |
| Code coverage                 | ≥80%      | 4     |
| Critical/high bugs            | 0         | 4-6   |
| Performance regression        | <5%       | 6     |
| Critical/high vulnerabilities | 0         | 6     |

### Documentation

| Document                    | Owner         | Stage |
| --------------------------- | ------------- | ----- |
| Release notes complete      | Product Owner | 8     |
| Changelog updated           | Developer     | 8     |
| Breaking changes documented | Developer     | 8     |
| Migration guide (if needed) | Developer     | 8     |

### Business Considerations (RA Pattern)

| Consideration                  | Reviewer        |
| ------------------------------ | --------------- |
| Deployment timing acceptable   | Release Manager |
| Dependent systems ready        | Release Manager |
| On-call team prepared          | Operations      |
| Customer communication planned | Product Owner   |

---

## Stakeholder Sign-offs

| Role          | Responsibility       | Required      |
| ------------- | -------------------- | ------------- |
| Product Owner | Feature completeness | Yes           |
| QA Lead       | Quality validation   | Yes           |
| Security      | Security approval    | If applicable |
| Compliance    | Regulatory approval  | If applicable |
| Operations    | Deployment readiness | Yes           |

### Sign-off Process

**Stage 7 (Exploration):**

- Product owner validates features work as expected
- QA confirms quality meets standards
- Security reviews (for security-sensitive changes)

**Stage 9 (Release Approval):**

- Release manager (RA) or automated gate (CDe) confirms all sign-offs obtained
- Final go/no-go decision

---

## Evidence Collection

### Automated Collection

Most evidence is automatically collected by the CI/CD pipeline:

- Test results uploaded as artifacts
- Coverage reports generated and stored
- Security scan results captured
- Performance metrics recorded

### Manual Collection

Some evidence requires human judgment:

- Stakeholder sign-offs (approval comments)
- Risk assessment review
- Business timing evaluation

---

## Evidence Storage

Evidence is stored with the release for audit and compliance:

| Evidence            | Storage Location                 |
| ------------------- | -------------------------------- |
| Test results        | CI artifacts, build manifest     |
| Security scans      | CI artifacts, security dashboard |
| Performance results | CI artifacts, monitoring system  |
| Sign-offs           | PR comments, approval records    |
| Release notes       | Changelog, GitHub release        |

---

## Next Steps

- [Approval Patterns](./approval-patterns.md) - How evidence is reviewed (RA vs CDe)
- [Release Notes](./release-notes.md) - How to document changes
- [Quality Gates](../quality-gates/index.md) - Stage-specific thresholds
