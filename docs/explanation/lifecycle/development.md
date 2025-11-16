# Development

This phase implements the [Continuous Delivery Model](../continuous-delivery/index.md), taking the solution from design through to production deployment.

---

## Implementation Plan

The implementation plan provides a framework for successful delivery, covering security, compliance, and data management. It ensures:

- Team and stakeholder alignment
- Clear goals and activities
- Risk mitigation
- Quality assurance
- Change management
- Deployment planning

**Regulatory note:** Some regulated processes require digital signature.

Templates: [Implementation Plan](../../../templates/compliance/implementation-plan.md)

---

## Specifications

Follow the guidance in the [Specifications section](../specifications/index.md).

---

## Risk Controls

Follow the guidance in the [Risk Controls article](../specifications/risk-controls.md).

---

## Implementation Report

At release approval, provide clear, non-technical documentation as a human-readable PDF so anyone can understand the development history.

### Required Content

**Implementation Report**:

- Summary of changes since last deployment
- Specification updates (new or changed)
- Conclusion on Fitness for Intended Use (regulatory)
- Business process impact
- Updated external dependencies
- Merge request approval comments
- Change log from all merge requests
- Specifications and related risk controls
- Test summary:
  - Installation Verification (IV)
  - Operational Verification (OV)
  - Performance Verification (PV)

**Design Documentation**:

- C4 model diagrams
- Architecture drawings
- Technologies used

**Supplier Information (regulatory)**:

- External suppliers and internal service providers
- Supplier assessments and interface agreements

The solution can be released when the Release Approval stage is approved.

**Regulatory note:** Some regulated processes require digital signature.

Templates: [Implementation Report](../../../templates/compliance/implementation-report.md)

---

## Testing

All tests should be automated and run in the pipeline. Manual testing occurs during the [Exploration stage](../continuous-delivery/cd-model/cd-model-stages-7-12.md#exploratory-testing-approach). Collect test evidence and save it in Git for inclusion in the implementation report.

---

## Documenting Changes

This is a fully automated process handled by the pipeline. It prepares all artifacts needed for [Release Approval](../continuous-delivery/cd-model/cd-model-stages-7-12.md#stage-9-release-approval).
