# Implementation Report {{ .ProjectName }}
<!--{% raw %}-->
!!! note

    Decision on overall change type (Normal, Standard or Emergency)
<!--{% endraw %}-->

**Change Type:** {{ .ChangeType }}

**Pipeline ID / Run Number:** {{ .PipelineID }}
**Repository / Branch:** {{ .Repository }}/{{ .Branch }}
**Build Date/Time:** {{ .BuildDate }}
**Triggered By** {{ .TriggeredBy }}  

## Summary

!!! note

    Add an executive summary for the release

### Changed requirements

<!-- This section should be dynamically created by knowing last-release-commit-sha and glob pattern for requirements in repo. the process should the via git log find changes. -->

<!--{% raw %}-->
!!! note

    Add a description of changes since the last deployment includes changes to existing requirements or newly added ones.
<!--{% endraw %}-->  

{{ .changed_requirements }}

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

{{ .req_approval_comments }}

## Change Log

<!-- This section should be dynamically created by knowing last-release-commit-sha - basically git log.  -->

The change log contains changes from all Merge Requests included in the release.

{{ .release_notes }}

---

## Requirements Specifications

This list includes all the requirements for the solution.

{{ .requirements }}

## Design Documentation

Please refer to the Solution Design Documentation document also generated as part of the audit ready documentation.

---

## Tests Summary

This section shows requirements traceability from features through acceptance criteria to test execution results.

<!--
To include multiple test suite results:
1. Run each test suite: `test suite commit`, `test suite acceptance`, etc.
2. Copy the "Tests Summary" section from each suite's test-suite-summary.md
3. Paste below, or use the multi-suite format shown in the example

Example multi-suite format:
-->

{{ if .commit_suite }}

### Commit Tests (L0-L2)

{{ .commit_suite }}

---
{{ end }}

{{ if .acceptance_suite }}

### PLTE Acceptance Tests (IV/OV/PV)

{{ .acceptance_suite }}

---
{{ end }}

{{ if .production_verification_suite }}

### Production Verification Tests (L4 + PIV)

{{ .production_verification_suite }}

---
{{ end }}
