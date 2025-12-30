# Vulnerability Remediation

Workflow for handling security findings from detection to resolution.

## Remediation Workflow

```text
Detection → Triage → Prioritize → Remediate → Verify → Document
```

### 1. Detection

- Automated scanning finds vulnerability
- Alert sent to team
- Issue created in tracking system

### 2. Triage (within 24 hours)

- Assess severity and impact
- Determine exploitability
- Identify affected systems

### 3. Prioritize

| Priority | Severity | Fix Timeline examples |
| -------- | -------- | --------------------- |
| P0       | Critical | 24 hours              |
| P1       | High     | 7 days                |
| P2       | Medium   | 30 days               |
| P3       | Low      | Next sprint           |

### 4. Remediate

- Update dependency to patched version
- Apply security patch
- Implement workaround if patch unavailable
- Add compensating controls

### 5. Verify

- Re-run security scans
- Validate fix is effective
- Ensure no regression

### 6. Document

- Record in vulnerability tracking system
- Update knowledge base
- Share learnings with team

## Blocking Strategy

Configure pipelines to block on security findings based on severity.

| Severity      | Action          | Pipeline Impact   |
| ------------- | --------------- | ----------------- |
| Critical/High | Block pipeline  | Deployment halted |
| Medium        | Warn and review | Requires approval |
| Low           | Informational   | Track in backlog  |

The `eac scan` command supports severity thresholds to control pipeline behavior.

## Tuning False Positives

### Suppression File

```yaml
# .trivyignore
# Suppress specific CVE (with justification)
CVE-2023-12345  # Not exploitable in our context - no user input reaches this code
```

### Best Practices

- Document why issues are suppressed
- Review suppressions regularly
- Set expiration dates on suppressions

## Security Metrics

Track these metrics to measure security posture:

| Metric              | Description                  |
| ------------------- | ---------------------------- |
| MTTD                | Mean Time To Detect          |
| MTTR                | Mean Time To Remediate       |
| Vulnerability Count | Open issues by severity      |
| Scan Coverage       | % of code/containers scanned |

## Evidence Collection

For compliance and audit:

- Store scan reports as artifacts
- Link findings to commits
- Maintain audit trail
- Generate compliance reports

## Best Practices

### Tool Maintenance

- Update Trivy vulnerability database daily
- Keep OWASP ZAP updated
- Enable Dependabot for all repositories and all modules

### Developer Education

- Secure coding training
- Understanding OWASP Top 10
- How to interpret scan results
- How to remediate vulnerabilities

### Continuous Improvement

- Review blocking thresholds quarterly
- Analyze false positive rates
- Update scan configurations
- Share learnings across teams
