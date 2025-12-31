# Compliance and Evidence

This article consolidates compliance concepts used throughout the CD Model, including verification types, evidence generation, signoff gates, and audit trail requirements.

## Verification Types

The CD Model uses three verification types to validate software readiness, primarily collected during Stage 5 (Acceptance Testing).

### Installation Verification (IV)

Confirms the solution can be installed and configured correctly in the target environment.

**What IV validates:**

- Deployment scripts execute without errors
- Dependencies are satisfied and compatible
- Configuration files are correctly applied
- Database migrations complete successfully
- Services start and respond to health checks

**Evidence generated:**

- Deployment logs
- Configuration validation reports
- Dependency audit results
- Health check responses

### Operational Verification (OV)

Ensures the solution operates as intended under normal conditions.

**What OV validates:**

- Functional requirements are met
- User workflows complete successfully
- Integration points function correctly
- Business rules are enforced
- Error handling works as expected

**Evidence generated:**

- Functional test results (unit, integration, system)
- API test reports
- User workflow validation logs
- Integration test results

### Performance Verification (PV)

Validates that performance benchmarks are met under expected load.

**What PV validates:**

- Response times meet SLAs
- Throughput handles expected load
- Resource utilization stays within limits
- System scales appropriately
- No memory leaks or degradation over time

**Evidence generated:**

- Load test results (p50, p95, p99 latencies)
- Throughput measurements
- Resource utilization metrics (CPU, memory, disk, network)
- Scalability test reports

---

## Evidence Generation

Evidence is automatically generated and stored throughout the pipeline, linked to specific builds for traceability.

### Automated Evidence Collection

| Stage                   | Evidence Type                                   | Storage                |
| ----------------------- | ----------------------------------------------- | ---------------------- |
| Stage 2 (Pre-commit)    | Lint reports, unit test results, security scans | Build artifacts        |
| Stage 3 (Merge Request) | Code review approvals, CI check results         | VCS + Build artifacts  |
| Stage 5 (Acceptance)    | IV/OV/PV reports, screenshots, API logs         | Artifact repository    |
| Stage 6 (Extended)      | Performance reports, security scan results      | Artifact repository    |
| Stage 9 (Approval)      | Approval records, quality metrics snapshot      | Compliance system      |
| Stage 11 (Live)         | Monitoring dashboards, incident reports         | Observability platform |

### Evidence Linking

Every piece of evidence is linked to:

- **Commit SHA**: The exact code version tested
- **Build ID**: The specific build artifact
- **Release version**: The release candidate (if applicable)
- **Requirement IDs**: Traced back to specifications

---

## Signoff Gates

The CD Model includes formal signoff points where human or automated approval is required.

### Stage 3: Peer Review (First-Level)

**Purpose:** Validate code quality and design before integration.

**Approver:** Peer developer(s)

**What's reviewed:**

- Code correctness and logic
- Readability and maintainability
- Test coverage
- Security implications
- Alignment with architecture

**Evidence:** PR approval records, review comments, CI check results

### Stage 9: Release Approval (Second-Level)

**Purpose:** Validate production readiness before deployment.

| Pattern                     | Approver                          | Mechanism                     |
| --------------------------- | --------------------------------- | ----------------------------- |
| RA (Release Approval)       | Release manager or approval board | Manual review and sign-off    |
| CDe (Continuous Deployment) | Automated quality gates           | All gates pass → auto-approve |

**What's validated:**

- All tests passing (100%)
- Code coverage meets threshold
- No critical/high vulnerabilities
- Performance benchmarks met
- Required documentation complete

**Evidence:** Quality metrics snapshot, approval records, release notes

### Stage 12: Feature Toggle (Third-Level)

**Purpose:** Control feature exposure to end users.

**Approver:** Feature owner or product manager

**What's decided:**

- When to enable features
- Rollout percentage (gradual: 1% → 10% → 50% → 100%)
- Kill switch activation if issues arise

**Evidence:** Feature flag configuration history, rollout logs

---

## Audit Trail Requirements

### Requirements Traceability

Every production change must trace back to its origin:

```text
Requirement/User Story
    ↓
Acceptance Criteria (Gherkin specs)
    ↓
Test Cases (automated)
    ↓
Code Changes (commits)
    ↓
Build Artifacts
    ↓
Deployment Records
```

### Change Traceability

For each change, the audit trail captures:

| Question                  | Evidence                               |
| ------------------------- | -------------------------------------- |
| **Who** made the change?  | Commit author, PR creator              |
| **Who** reviewed it?      | PR approvers, release approver         |
| **When** was it made?     | Commit timestamps, approval timestamps |
| **Why** was it made?      | Linked requirement, PR description     |
| **What** was changed?     | Diff, affected files                   |
| **How** was it validated? | Test results, scan reports             |

### Retention Requirements

Evidence must be retained according to organizational and regulatory requirements:

- **Minimum:** Duration of software's production lifecycle
- **Regulatory:** As specified by applicable regulations (FDA, SOX, etc.)
- **Recommended:** At least 3-5 years for audit purposes

---

## Compliance Artifact Management

### RA Pattern (Manual Oversight)

- Artifacts reviewed and approved manually before release
- Stored in designated compliance systems (e.g., regulatory document management)
- Formal approval documented with signatures/timestamps
- Audit trail emphasized throughout

### CDe Pattern (Automated)

- Artifacts generated automatically at each stage
- Stored in artifact repositories (linked to builds)
- Compliance validation automated where possible
- Evidence available for retrospective audit

Both patterns generate the same evidence types; the difference is in approval workflow and oversight level.

---

## References

- [The 12 Stages](stages.md) - Where verification occurs
- [Variants](variants.md) - RA vs CDe compliance differences
- [Testing Strategy](../testing/index.md) - Test levels and coverage
