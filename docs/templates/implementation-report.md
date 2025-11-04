# Implementation Report {{ cookiecutter.project_name }}

!!! note

    Decision on overall change type (Normal, Standard or Emergency)

**Change Type:** \[Standard|Normal]

<!--{% raw %}-->
**Pipeline ID / Run Number:** {{ dynamic.pipeline_id }}
**Repository / Branch:** {{ dynamic.repository_branch }}
**Build Date/Time:** {{ dynamic.build_date }}
**Triggered By** {{ dynamic.triggered_by }}
<!--{% endraw %}-->  

## Summary

!!! note

    Add an executive summary for the release

### Changed requirements

<!-- This section should be dynamically created by knowing last-release-commit-sha and glob pattern for requirements in repo. the process should the via git log find changes. -->

!!! note

    Add a description of changes since the last deployment includes changes to existing requirements or newly added ones.

<!--{% raw %}-->
{{ dynamic.changed_requirements }}
<!--{% endraw %}-->  

### Conclusion on Fitness for Intended Use

<!-- Here we have a requirement to stop the pipeline and await user input for this field. -->

!!! important

    Provide a conclusion on fitness for intended use

### Impact on Business Process

<!-- Here we have a requirement to stop the pipeline and await user input for this field. -->

!!! note

    Describe the impact on the supported business process

---

## Design Review

<!-- This section should be dynamically created by knowing last-release-commit-sha and glob pattern for URS in repo.  -->

Changes to requirements from Merge Request approvals, each row should contain name of the approver.

<!--{% raw %}-->
{{ dynamic.req_approval_comments }}
<!--{% endraw %}-->

## Change Log

<!-- This section should be dynamically created by knowing last-release-commit-sha - basically git log.  -->

The change log contains changes from all Merge Requests included in the release.

<!--{% raw %}-->
{{ dynamic.release_notes }}
<!--{% endraw %}-->

---

## Requirements Specifications

This list includes all the requirements for the solution.

<!--{% raw %}-->
{{ dynamic.requirements }}
<!--{% endraw %}-->

## Design Documentation

Please refer to the Solution Design Documentation document also generated as part of the audit ready documentation.

---

## Tests Summary

This section shows requirements traceability from features through acceptance criteria to test execution results. Each test scenario is uniquely identifiable, with execution results (automated or manual) linked to the corresponding requirement.

!!! note Example

    **Feature ID**: `api_user_authentication`
    **User Story**: As a user, I want to authenticate securely, so that I can access protected resources
    **Specification**: [acceptance.spec](specs/api/user_authentication/acceptance.spec) | [behavior.feature](specs/api/user_authentication/behavior.feature)

    | Acceptance Criteria                                        | Execution Type | Result    |
    |------------------------------------------------------------|----------------|-----------|
    | AC1: System validates user credentials correctly           | Automated      | 🟢 Passed |
    | AC1: System rejects invalid credentials with clear error   | Automated      | 🟢 Passed |
    | AC1: System handles missing credentials appropriately      | Automated      | 🟢 Passed |
    | AC2: Authentication system installs with secure defaults   | Automated      | 🟢 Passed |
    | AC2: Installation verifies certificate configuration       | Automated      | 🟢 Passed |
    | AC3: Authentication completes within 500ms threshold       | Automated      | 🔴 Failed |
    | AC3: System handles 100 concurrent authentication requests | Automated      | 🟢 Passed |
    | AC3: Authentication maintains memory under 100MB           | Automated      | 🟢 Passed |

<!--{% raw %}-->
{{ dynamic.feature_test_results }}
<!--{% endraw %}-->

---

### Verification (IV/OV/PV)

- **IV (Installation Verification)**: Tests that verify installation, deployment, configuration, and setup
- **OV (Operational Verification)**: Tests that verify functional behavior, business logic, and operational requirements
- **PV (Performance Verification)**: Tests that verify performance requirements, response times, and resource usage

#### Installation Verification - IV

Installation verification is performed automatically using the following methods in a Production-Like Test Environment (PLTE). The test results below display all scenarios tagged with `@IV`.

1. **Execution Logs:** Tools generate logs to document the setup process.
2. **Baseline Checks:** Automated commands establish system and environment baselines.
3. **Version Verification:** Scripts verify that all components are operating on the correct versions.

<!--{% raw %}-->
{{ dynamic.iv_test_traceability_report }}
<!--{% endraw %}-->

---

#### Operational Verification (Traceability) - OV

The test results below display all scenarios **not** tagged with `@IV` or `@PV`.

<!--{% raw %}-->
{{ dynamic.ov_test_traceability_report }}
<!--{% endraw %}-->

---

#### Performance Verification - PV

Performance verification is conducted in a Production-Like Test Environment (PLTE) as well as in the production environment following the release. The test results below display all scenarios tagged with `@PV`.

<!--{% raw %}-->
{{ dynamic.pv_test_traceability_report }}
<!--{% endraw %}-->
