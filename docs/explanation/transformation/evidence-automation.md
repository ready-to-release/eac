# Evidence Automation

## Introduction

Manual evidence collection is one of the most time-consuming and error-prone aspects of traditional compliance. Teams spend weeks assembling evidence for audits, searching through systems for the right artifacts, and creating packages that auditors can review.

Evidence automation transforms this process by generating evidence automatically as a byproduct of delivery. When auditors arrive, complete evidence packages are already available—generated in minutes, not weeks.

This document explains the architecture for automated evidence collection, the types of evidence collected, and how to implement evidence automation in delivery pipelines.

**Audience**: Engineering leaders, compliance officers, DevOps engineers, and anyone responsible for audit evidence.

**What you'll learn**: You'll understand the evidence automation architecture, see concrete examples of evidence collection, and learn how to implement automated evidence generation in your delivery pipelines.

## The Problem with Manual Evidence

### Challenges

**Time-Consuming**
- Typical audit preparation: 200-400 person-hours
- Evidence collection takes majority of preparation time
- Teams pulled from feature work to support audits
- External auditor billable time wasted waiting for evidence

**Error-Prone**
- Human forgetfulness leads to missing evidence
- Inconsistent collection across teams
- Evidence quality varies
- Gaps discovered mid-audit

**Difficult to Prove Completeness**
- How do you know you collected everything?
- Hard to demonstrate comprehensive coverage
- Manual checklists miss items
- Auditors question completeness

**Not Real-Time**
- Evidence collected retrospectively during audit
- No continuous visibility into compliance posture
- Issues discovered weeks after they occur
- Can't prove compliance between audits

### Cost

Quantified impact for typical mid-size organization (30 teams):

**Audit Preparation Cost**
- 200-400 person-hours per audit cycle
- 4 audits per year (quarterly)
- 800-1,600 person-hours annually
- At $100-150/hour: $80K-$240K per year just for evidence collection

**Audit Cost**
- External auditors wait for evidence: 20-40% of audit time
- Wasted auditor billable hours
- Extended audit duration
- Higher audit fees

**Risk of Findings**
- Evidence gaps lead to audit findings
- Findings require remediation
- Potential compliance violations
- Regulatory scrutiny

## Automated Evidence Architecture

### Core Concept

Evidence generation integrated into delivery pipeline as byproduct:

1. **Pipeline Execution** → Artifacts automatically collected
2. **Git Commits** → Audit trail automatically captured
3. **Test Execution** → Results automatically stored
4. **Scans and Checks** → Reports automatically generated
5. **On Demand** → Evidence package created in minutes

No manual collection needed. Evidence exists before audit begins.

### Evidence Sources

#### 1. Pipeline Artifacts

Collected automatically from every pipeline run:

**Test Results**
- JUnit XML from unit tests
- HTML reports from integration tests
- Godog results from compliance scenarios
- Coverage reports showing test completeness

**Security Scans**
- SAST (Static Application Security Testing) results
- DAST (Dynamic Application Security Testing) results
- Dependency vulnerability scans
- Container image scans
- Secret detection results

**SBOMs (Software Bill of Materials)**
- Complete component inventory
- License information
- Dependency tree
- Supply chain traceability

**Deployment Logs**
- What was deployed
- When it was deployed
- Who approved deployment
- Which environment
- Deployment success/failure status

**Quality Gate Results**
- Which gates passed/failed
- Why gates failed
- Override decisions and justifications

**Security Integration**: See [Security in CD Model](../continuous-delivery/security/index.md) for detailed security tooling.

#### 2. Git History

Version control provides comprehensive audit trail:

**Commit Messages**
- What changed
- Why it changed
- Who made the change
- When it occurred

**Pull Request Records**
- Code review discussions
- Approval trail
- Review comments
- Merge decisions

**Tag History**
- Release tracking
- Version progression
- Release notes

**Contributor Attribution**
- Who wrote which code
- Blame/annotate for traceability
- Team composition over time

**Merge Records**
- When changes integrated
- Merge strategy used
- Conflicts and resolutions

#### 3. Automated Reports

Generated automatically from collected data:

**Traceability Matrix**
- Requirement → Test → Result linkage
- @risk tag coverage analysis
- Implementation completeness
- Evidence references

**Coverage Reports**
- Which requirements have tests
- Which tests are passing
- Evidence availability per requirement
- Gaps and missing coverage

**Compliance Status Dashboard**
- Real-time compliance posture
- Passing vs failing scenarios
- Trend analysis over time
- Alerts for regressions

**Trend Analysis**
- Compliance improving or degrading
- New requirements vs implementation rate
- Test success rate over time
- Evidence collection completeness trend

#### 4. Runtime Logs

Collected from production systems:

**Access Logs**
- Authentication events
- Authorization decisions
- Failed access attempts
- Session management

**Audit Trails**
- Data changes (old value → new value)
- User actions
- Administrative operations
- Configuration changes

**Security Events**
- Alert notifications
- Potential incidents
- Threat detection results
- Response actions

**Performance Metrics**
- Response times
- Resource utilization
- Availability metrics
- SLA compliance

### Evidence Storage

**Immutable Storage**
- S3 with versioning enabled
- Azure Blob Storage with immutability policies
- Write-once-read-many (WORM) storage
- Prevents tampering, provides integrity

**Retention Policy**
- Per regulatory requirements (often 7 years for GxP)
- Lifecycle rules for automatic archival
- Cost optimization through storage tiers
- Legal hold capabilities

**Organization**
- By date/time period (2025-Q1)
- By release version (v2.3.0)
- By evidence type (test-results, security-scans)
- Consistent directory structure

**Access Control**
- Read-only access for auditors
- Write access only for automated systems
- Audit logging on evidence access
- Encryption at rest and in transit

## Evidence Collection Flow

```
Delivery Pipeline Execution
         │
         ├─> Test Execution → test-results.xml
         ├─> Security Scan → scan-results.json
         ├─> SBOM Generation → sbom.spdx
         ├─> Deployment → deployment.log
         ├─> Quality Gates → gate-results.json
         │
         ▼
Evidence Store (S3 / Azure Blob / Artifact Repository)
  │
  │  [Evidence accumulates over time]
  │  [Organized by date, release, type]
  │
  ▼
On-Demand Evidence Package Generator
  │
  ├─> Query evidence store for time period
  ├─> Collect relevant artifacts
  ├─> Extract Git history for period
  ├─> Generate traceability matrix
  ├─> Create package README with instructions
  ├─> Bundle all evidence into archive
  │
  ▼
audit-evidence-2025-Q1.zip
  │
  └─> Delivered to auditors
```

**Automation Note**: Evidence collection and packaging requires automation layer. Ready-to-Release (r2r) CLI tries to help with these operations.

## Traceability Matrix Generation

### What It Links

Creates comprehensive requirement-to-evidence linkage:

```
Requirement (@risk tag)
    ↓
Test Scenarios (Gherkin)
    ↓
Test Execution (test results)
    ↓
Evidence Artifacts (logs, reports)
    ↓
Code Implementation (Git commits)
```

### Example Matrix

```
Compliance Traceability Matrix
Generated: 2025-11-06 10:00:00 UTC
Period: 2025-Q1 (January 1 - March 31, 2025)

================================================================================
@risk1: Authentication required for all access (RC-001)
================================================================================

Source: ISO 27001:2022 A.8.5, GDPR Article 32
Assessment: RISK-ASSESS-2025-01

Implementation Scenarios: 3

[1] ✅ Valid credentials grant access
    Specification: specs/cli/user-authentication/specification.feature:15
    Test Result: PASS (2025-03-15 14:32:00)
    Test File: test-results/l3-plte-2025-Q1.xml:line-45
    Evidence: logs/authentication-2025-Q1.log
    Code: src/cli/auth/authentication.go
    Commit: abc123def (2025-01-15, "Implement user authentication")
    Author: jane.doe@example.com

[2] ✅ Invalid credentials deny access
    Specification: specs/cli/user-authentication/specification.feature:23
    Test Result: PASS (2025-03-15 14:32:05)
    Test File: test-results/l3-plte-2025-Q1.xml:line-67
    Evidence: logs/authentication-2025-Q1.log

[3] ✅ No credentials block access
    Specification: specs/cli/user-authentication/specification.feature:31
    Test Result: PASS (2025-03-15 14:32:08)
    Test File: test-results/l3-plte-2025-Q1.xml:line-89
    Evidence: logs/authentication-2025-Q1.log

Coverage: 100% (3/3 scenarios passing)
Evidence Completeness: 100%
Last Validation: 2025-03-15 14:32:08
Status: COMPLIANT

================================================================================
@risk2: Multi-factor authentication for privileged access (RC-002)
================================================================================

Source: ISO 27001:2022 A.8.5
Assessment: RISK-ASSESS-2025-01

Implementation Scenarios: 2

[1] ✅ Admin access requires MFA
[2] ✅ MFA failure blocks access

Coverage: 100% (2/2 scenarios passing)
Status: COMPLIANT

================================================================================

Summary:
- Total Requirements: 42
- Total Scenarios: 127
- Passing: 125 (98.4%)
- Failing: 2 (1.6%)
- Evidence Complete: 98.4%
- Overall Status: COMPLIANT (with 2 findings)
```

## Evidence Package Structure

### Directory Layout

```
audit-evidence-2025-Q1/
├── README.md                          # Package overview and navigation
├── traceability-matrix.md             # Complete requirement → evidence mapping
├── compliance-summary.md              # Executive summary of compliance status
│
├── test-results/                      # All test execution results
│   ├── l0-unit-tests-2025-01.xml
│   ├── l0-unit-tests-2025-02.xml
│   ├── l0-unit-tests-2025-03.xml
│   ├── l2-integration-tests-2025-Q1.xml
│   ├── l3-plte-tests-2025-Q1.xml
│   └── test-coverage-report.html
│
├── security-scans/                    # Security scanning results
│   ├── sast-scan-2025-01-15.json
│   ├── sast-scan-2025-02-12.json
│   ├── sast-scan-2025-03-18.json
│   ├── dependency-scan-2025-Q1.json
│   ├── container-scan-2025-Q1.json
│   └── secret-scan-results-2025-Q1.log
│
├── sbom/                              # Software Bill of Materials
│   ├── sbom-v1.0.0.spdx
│   ├── sbom-v1.1.0.spdx
│   ├── sbom-v1.2.0.spdx
│   └── dependency-licenses.md
│
├── deployment-logs/                   # Deployment history
│   ├── prod-deployment-2025-01-15.log
│   ├── prod-deployment-2025-02-12.log
│   ├── prod-deployment-2025-03-18.log
│   └── deployment-approvals-2025-Q1.json
│
├── access-logs/                       # Authentication and authorization
│   ├── authentication-2025-01.log
│   ├── authentication-2025-02.log
│   ├── authentication-2025-03.log
│   └── authorization-audit-2025-Q1.log
│
├── audit-trails/                      # Data change tracking
│   ├── data-changes-2025-01.log
│   ├── data-changes-2025-02.log
│   ├── data-changes-2025-03.log
│   └── admin-actions-2025-Q1.log
│
├── policies/                          # Policies and procedures
│   ├── information-security-policy.md
│   ├── change-management-procedure.md
│   ├── access-control-policy.md
│   └── policy-change-log.md
│
├── risk-controls/                     # Risk control specifications
│   ├── authentication-controls.feature
│   ├── data-protection-controls.feature
│   ├── audit-trail-controls.feature
│   └── implementation-status.md
│
└── git-history/                       # Version control audit trail
    ├── commits-2025-Q1.log
    ├── pull-requests-2025-Q1.json
    ├── releases-2025-Q1.json
    └── contributors-2025-Q1.md
```

### README.md Structure

```markdown
# Audit Evidence Package: 2025 Q1

**Period**: January 1, 2025 - March 31, 2025
**Generated**: 2025-04-01 09:00:00 UTC
**Organization**: Example Corp
**System**: Production Application Platform

## Package Contents

This evidence package contains automated compliance evidence for Q1 2025:

- **Traceability Matrix**: Complete requirement → evidence mapping
- **Test Results**: All compliance test executions
- **Security Scans**: SAST, DAST, dependency, container scans
- **SBOMs**: Software bill of materials for all releases
- **Deployment Logs**: All production deployment records
- **Access Logs**: Authentication and authorization audit trail
- **Audit Trails**: Data change tracking
- **Policies**: Current policies and change history
- **Risk Controls**: Compliance requirement specifications

## Quick Start for Auditors

1. Read `compliance-summary.md` for executive overview
2. Review `traceability-matrix.md` for requirement coverage
3. Examine `test-results/` for validation evidence
4. Check `security-scans/` for security posture
5. Review `deployment-logs/` for change management
6. Examine `access-logs/` and `audit-trails/` for access control

## Evidence Generation

All evidence in this package was generated automatically:
- Test results from CI/CD pipeline execution
- Security scans from automated scanning tools
- Deployment logs from deployment automation
- Access logs from production systems
- Git history from version control system

No manual evidence collection was performed.

## Compliance Status Summary

- Total Requirements: 42
- Implementation Coverage: 100%
- Test Pass Rate: 98.4% (125/127 scenarios)
- Evidence Completeness: 98.4%
- Overall Status: COMPLIANT (with 2 findings)

See `compliance-summary.md` for detailed status.

## Contact

For questions about this evidence package:
- Compliance Officer: compliance@example.com
- Engineering Lead: engineering@example.com
- DevOps Team: devops@example.com
```

## Integration with CD Model

### Evidence by Stage

| CD Stage | Evidence Collected |
|----------|-------------------|
| **Stage 2: Pre-commit** | Secret scan results, policy check logs |
| **Stage 4: Commit** | Unit/integration test results, SAST scans, dependency scans, SBOM |
| **Stage 5: Acceptance** | L3 PLTE test results, DAST reports, integration validation |
| **Stage 9: Release Approval** | Approval records, compliance checklist validation, release notes |
| **Stage 11: Live** | Runtime logs, access logs, audit trails, performance metrics |

**Reference**: [CD Model Documentation](../continuous-delivery/cd-model/cd-model-overview.md) for complete stage descriptions.

### Pipeline Integration Example

```yaml
# .github/workflows/compliance-evidence.yml
name: Compliance Evidence Collection

on: [push, pull_request]

jobs:
  collect-evidence:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Run Compliance Tests
        run: go test -v ./src/*/tests > test-results.txt

      - name: Run Security Scans
        run: |
          trivy image --format json myapp:latest > security-scan.json

      - name: Generate SBOM
        run: |
          syft myapp:latest -o spdx > sbom.spdx

      - name: Collect Evidence Artifacts
        if: always()
        run: |
          mkdir -p evidence-artifacts
          cp test-results.txt evidence-artifacts/
          cp security-scan.json evidence-artifacts/
          cp sbom.spdx evidence-artifacts/

      - name: Upload to Evidence Store
        if: always()
        run: |
          aws s3 cp evidence-artifacts/ \
            s3://compliance-evidence/$(date +%Y-%m)/$(git rev-parse HEAD)/ \
            --recursive
```

## Implementation Patterns

### Pattern 1: Pipeline Artifact Collection

```yaml
- name: Collect Compliance Evidence
  if: always()  # Collect even if tests fail
  uses: actions/upload-artifact@v3
  with:
    name: compliance-evidence-${{ github.sha }}
    path: |
      test-results/
      security-scans/
      sbom.spdx
      deployment.log
    retention-days: 2555  # 7 years for GxP compliance
```

### Pattern 2: Immutable Storage Upload

```bash
#!/bin/bash
# Upload evidence to immutable S3 bucket

EVIDENCE_DATE=$(date +%Y-%m-%d)
COMMIT_SHA=$(git rev-parse HEAD)

# Upload with versioning and lifecycle policy
aws s3 cp evidence-package.zip \
  s3://compliance-evidence/${EVIDENCE_DATE}/${COMMIT_SHA}/ \
  --storage-class GLACIER \
  --metadata "commit=${COMMIT_SHA},date=${EVIDENCE_DATE}"

# Enable object lock for immutability
aws s3api put-object-legal-hold \
  --bucket compliance-evidence \
  --key ${EVIDENCE_DATE}/${COMMIT_SHA}/evidence-package.zip \
  --legal-hold Status=ON
```

### Pattern 3: On-Demand Package Generation

```bash
#!/bin/bash
# Generate evidence package for specific time period

START_DATE="2025-01-01"
END_DATE="2025-03-31"
PACKAGE_NAME="audit-evidence-2025-Q1"

# Automation layer needed for:
# - Collecting artifacts from evidence store
# - Extracting relevant Git history
# - Generating traceability matrix
# - Bundling into evidence package

r2r evidence generate \
  --start-date ${START_DATE} \
  --end-date ${END_DATE} \
  --output ${PACKAGE_NAME}.zip

# Package ready for auditors in minutes
```

*Note: Ready-to-Release (r2r) CLI tries to help with evidence generation*

## Benefits by Stakeholder

### For Compliance Teams

- **95%+ automation**: Manual collection eliminated
- **Real-time audit readiness**: Always prepared for audits
- **Complete and consistent**: Automation ensures nothing is missed
- **Easy to prove coverage**: Traceability matrix shows everything

### For Auditors

- **Evidence provided upfront**: No waiting, audit starts immediately
- **Traceability matrix**: Clear linkage from requirement to evidence
- **Immutable storage**: Proves integrity, no tampering possible
- **30-50% reduced audit time**: Less time searching, more time validating

### For Development Teams

- **Zero manual work**: Evidence generated automatically
- **No audit preparation**: No scrambling when audit announced
- **Focus on features**: Time spent building, not collecting evidence

### For Leadership

- **Continuous visibility**: Real-time compliance dashboard
- **Reduced costs**: 80-90% reduction in audit prep time
- **Lower risk**: Comprehensive evidence reduces audit findings
- **Confidence**: Know compliance posture at any moment

## Success Metrics

Track these metrics to measure evidence automation success:

**Evidence Automation Rate**
- Target: 95%+ automatically generated
- Measure: Percent of evidence requiring no manual collection
- Baseline: Typically 0-10% in traditional approaches

**Audit Preparation Time**
- Target: 80% reduction from baseline
- Measure: Person-hours to prepare for audit
- Baseline: Typically 200-400 person-hours

**Evidence Completeness**
- Target: 100% coverage for critical requirements
- Measure: Percent of requirements with complete evidence
- Baseline: Often 70-85% in manual approaches

**Audit Duration**
- Target: 40% reduction in external audit hours
- Measure: Billable hours for external auditors
- Baseline: Varies by organization and framework

## Next Steps

To implement evidence automation:

1. **Understand shift-left** - Read [Shift-Left Compliance](shift-left-compliance.md) to see when evidence is generated
2. **Learn risk controls** - Read [Risk Control Specifications](risk-control-specifications.md) to understand what evidence proves
3. **Review framework** - Read [Transformation Framework](transformation-framework.md) for implementation approach
4. **Check CD Model** - Review [CD Model](../continuous-delivery/cd-model/cd-model-overview.md) for pipeline integration points

## Related Documentation

- [Compliance as Code](compliance-as-code.md) - Principle 4: Automated Evidence Collection
- [Risk Control Specifications](risk-control-specifications.md) - What evidence must prove
- [Shift-Left Compliance](shift-left-compliance.md) - When and where evidence is generated
- [Transformation Framework](transformation-framework.md) - How to implement evidence automation
- [CD Model](../continuous-delivery/cd-model/cd-model-overview.md) - Pipeline integration points
- [Security in CD Model](../continuous-delivery/security/index.md) - Security tooling and evidence
