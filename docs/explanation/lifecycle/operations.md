# Operations

Operations is not a separate phase but an integral part of the continuous delivery lifecycle. In a DevOps model, the team that builds the system also runs and maintains it.

[](#operation--maintenance-plan)

---

## Operation & Maintenance Plan

Defines how systems are operated, maintained, and kept compliant and secure throughout their lifecycle. Ensures systems meet operational, regulatory, and quality standards.

**Regulatory note:** Some regulated processes require digital signature.

Templates: [Operation & Maintenance Plan](../../../templates/compliance/operations/operation-maintenance.md)

### Problem Management

Use blameless post-mortems for all significant incidents. Document findings for future learning and prevention.

Templates: [Blameless Post Mortem](../../../templates/compliance/operations/blameless-post-mortem.md)

---

## Operational Procedures

### Periodic Evaluation

Regular assessment of system health, compliance, and alignment with business needs.

**Activities:**

- Review system performance against SLAs
- Assess technical debt and improvement opportunities
- Validate continued fitness for intended use
- Update risk assessments based on usage changes

**Frequency:** Annually (mandatory for regulated systems, recommended for others)

### User Access Review

Validates that system access remains appropriate and secure.

**Activities:**

- Review all active user accounts
- Verify access levels match current roles
- Remove inactive or unnecessary accounts
- Document changes and justifications

**Frequency:** Annually minimum; more frequent for high-risk systems

---

## Defects & Incidents

| Term         | Definition                                                                                        | Examples                                                          |
|--------------|---------------------------------------------------------------------------------------------------|-------------------------------------------------------------------|
| **Defect**   | A flaw in software that causes incorrect behavior or failure to meet requirements                 | Bug in code, incorrect data display, security vulnerability       |
| **Incident** | An unplanned interruption or reduction in quality of IT services requiring immediate attention    | System outage, application crash, network connectivity issue      |

### Defects

Document all defects detected after [Production Deployment](../continuous-delivery/cd-model/cd-model-stages-7-12.md#stage-10-production-deployment) by adding items to the backlog.

**Requirements:**

- Each defect must have a unique identifier
- Link fixes to defect reports
- Document root cause analysis for critical defects
- Track defect trends for quality improvements

Templates: [Defect](../../../templates/compliance/operations/defect.md)

---

### Incidents

Record all production incidents in the IT service management system. Complete the incident template to ensure traceability.

**Process:**

- Create incident reports using template
- Link incident resolution to code changes
- Conduct [Blameless Post Mortem](../../../templates/compliance/operations/blameless-post-mortem.md) for significant incidents

Templates: [Incident](../../../templates/compliance/operations/incident.md)
