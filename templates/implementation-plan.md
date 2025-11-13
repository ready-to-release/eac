# Implementation Plan {{ .ProjectName }}

<!--{% remove %}-->
!!! tip

    The implementation plan must be approved in document management system.
    Electronic signatures are required for the implementation plan per regulatory requirements (FDA 21 CFR Part 11, EU Annex 11).
<!--{% endremove %}-->

**This document is signed electronically in the document management system by the QA and System Owner.**

**Prepared by:** [Insert name]
**Date:** [Insert date]

## Definitions

| **Term** | **Definition**                                                                                                                                                                                                                     |
| -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Artefact | Any document, diagram, file, or other piece of information that is produced, modified, or used during the software development process. This includes items such as requirements documents, design specifications, and test plans. |
| CI       | Continuous Integration is a development practice where developers regularly merge their code changes into a central repository. This process is followed by automated builds and tests.                                                            |
| CD       | Continuous Delivery is the practice of ensuring that the system is always in a deployable state, and that every change could potentially be released to production at any time. Continuous Delivery automates and streamlines the release process. |
| Tool     | Software used to support development, operation or maintenance of the software product. Tools are not part of the finished software product.                                                                                       |

For general abbreviations, refer to organizational quality system documentation {{abbreviations_reference}}.

## Scope

This document is an implementation plan that describes the activities to ensure {{ .ProjectName }} is fit for the intended use,
as required by applicable GxP regulations (FDA 21 CFR Part 11, EU Annex 11, GAMP 5) and organizational quality procedures.

The plan describes a way of working and covers multiple releases that are delivered on a continuous basis.

All releases are assessed to determine if a GxP change control is required according to organizational change control procedures.

Implementation reports will be written for each release, and reporting is always done against the most current version of this implementation plan.

Each implementation report will reference the related change request(s) when required by change control procedures.

The plan is continuously evaluated to always reflect how the team is working. Changes to the original scope will be evaluated for impact on the planning established in this document. All changes are considered and managed as normal changes.

## Roles and responsibilities

The generic responsibilities defined in the organizational quality system procedures apply unless otherwise specified for system management, system ownership, or information protection procedures.

| Role                   | Name | Comments |
| ---------------------- | ---- | -------- |
| System Owner           |      |          |
| Data Owner             |      |          |
| Project/System Manager |      |          |
| QA                     |      |          |

---

## Supported Business Process

<!--{% remove %}-->
!!! tip

    This section should provide a comprehensive overview of the processes that are supported by the system or project.
    This includes detailing which steps are supported by business process.
    It should also reference relevant documentation and System Description in the service management system.

<!--{% endremove %}-->

## System Overview

<!--{% remove %}-->
!!! tip

    This section should provide a high level overview to illustrate the scope and boundary of the system covered by the plan.
    It should also include any interfaces to other systems.
<!--{% endremove %}-->

## Initial Risk Assessment

<!--{% remove %}-->
!!! tip

    If the business application is supporting a GxP regulated process, add a high level statement on software category per GAMP 5 (Category 1-5).
<!--{% endremove %}-->

**GAMP 5 Category:** {{gamp_category}}
**Rationale:** {{gamp_rationale}}

Relevant risk assessments are performed, and the resulting risk controls are stored as executable specifications in Git, enabling direct traceability to the corresponding requirements. See [Specifications](#specifications) for more details.

---

## Development and Implementation Strategy

### Delivery Workflow

Software will be developed, deployed and maintained with a 12-stage workflow, that is based on the principles of Continuous Integration (CI) and Continuous Delivery (CD).

At each stage in the workflow, more extensive and higher-level tests are performed, providing increased confidence in the committed changes as the software progresses through the workflow.

| **#**  | **Stage Name**        | **Description**                                                                                                                                                                                                                                                                                                                                                                                                             | **Primary** **Roles**         |
| ------ | --------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------- |
| **1**  | Authoring Change      | A member of the development team initiates changes within a local development environment. Code and configuration are updated in parallel with requirements, test cases, risk assessments, and other documentation. The intent is to keep the system in a continuously compliant state, by integrating quality and risk management activities for each small change. The changes should be no larger than two days of work. | Team member                   |
| **2**  | Pre-Commit            | A fully automated stage, where the changes made by the author are subjected to automated unit tests and other static testing. This includes running an automated tool that verifies that the changed code adheres to solution standards and best practices (see [Code Review](#code-review)).                                                                                                                               | Automated                     |
| **3**  | Merge Request         | The proposed changes are reviewed and approved by a peer in the development team. For changes that affect user requirements, System Owner (or Delegate) and QA also approves. For other changes, any member of the development team can act as approver.                                                                                                                                                                    | Team member                   |
| **4**  | Commit                | A fully automated stage with no manual activities. It serves as an automated gate that gives fast feedback on the viability of the changes, not to execute complete test portfolios. An automated build process generates immutable artifacts, such as PDF documentation and software. This step is time-boxed (few minutes) and executes as many unit and integration tests as possible within the time-box.               | Automated                     |
| **5**  | Acceptance            | Execution of automated tests that cover unit and integration tests that were not able to be completed within time-limits of previous step. Acceptance also includes system level end-to-end testing, and is performed in a production-like test-environment.                                                                                                                                                                | Automated                     |
| **6**  | Extended Testing      | Optional stage for horizontal system level testing including performance testing, security scanning, and additional compliance validation when required.                                                                                                                                                                                                                                                                    | Automated                     |
| **7**  | Exploration           | Ongoing exploratory testing of the current release candidate, in a demonstration environment.                                                                                                                                                                                                                                                                                                                               | Can vary                      |
| **8**  | Start Release         | A separate branch is created with a locked feature set for the next deployment to production.                                                                                                                                                                                                                                                                                                                               | Team member                   |
| **9**  | Release Approval      | The Release Branch is approved for **Deployment** and **Release** by QA and System Owner (or delegate). A set of reports are generated to inform the decision to release (see [Release Approval](#release-approval)). All releases will go through the **Release Approval** step, where a complete set of updated documentation will be generated and approved.                                                             | QA and System Owner |
| **10** | Production Deployment | Manually initiated automated deployment to the production environment.                                                                                                                                                                                                                                                                                                                                                      | Team member                   |
| **11** | Live                  | Continuous monitoring of the software for optimal performance and reliability.                                                                                                                                                                                                                                                                                                                                              | Team member                   |
| **12** | Release Toggling      | The team can decide for every piece of functionality to separate deployment and release with feature flags. With feature flags changes can be activated in the live production system after release, without changing the system code. The timing can be decided by the system owner (or delegate), since authorization to release is formally given by approval of the release report in **Release Approval**.             | System Owner   |

Throughout the workflow, close cooperation between all stakeholders is expected.

### Approval Gates

Three stages of the workflow (**Merge Request**, **Release Approval**, **Release Toggling**) are manually controlled gates that are used for the approvals required by the quality system.

- **Pull request review** ensures that every change to the IT solution and the related documentation is traceable to an author and approved by a Peer.
  When the changes are related to User Requirements, the pull request must be approved by QA and System Owner (or delegate).
- The Final Design Review, Performance verification, Implementation Report and authorization to release is approved by QA and System Owner (or delegate) in the **Release Approval.**
- The timing of the **Release** is decided by the System Owner.

### Configuration management

Git has been selected as the underlying technology for the version control system. It is used to manage application code, infrastructure (as code), specifications, documentation, and other configuration items.

Every change made in a Git repository includes a clear description explaining the purpose of the change.

This ensures full traceability — providing visibility into what changes were made, by whom, and why.

Git maintains a complete history of all software revisions, making it possible to recreate the system’s state at any point in time when needed.

Sensitive configurations, such as passwords, are confidential and are not stored unencrypted in the Git repository.
Instead, they are kept in a secure secret management system accessible only to authorized personnel.

If the quality system requires certain artifacts to be controlled in specific IT systems, they are transferred to their designated locations after the Release Approval step.

### Environments

| **Environment name**                    | **Activities**                                                                                                                                                                                                                                                                                                                                        |
| --------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Development                             | Local Development Environment. This is where developers initiate changes, write code, and perform pre-integration testing. Used in: Step 1 (Authoring Change) and Step 2 (Pull request).                                                                                                                                                              |
| Production-Like-Test-Environment (PLTE) | Production-like Test Environment (PLTE). This environment mimics the production setup as closely as possible. It's used for more comprehensive automated testing on the system level. If some production data sources or dependencies are not available, suitable mockups should be employed. Used in: Step 5 (Acceptance) and Step 6 (Extended test) |
| Demonstration Environment               | Demonstration Environment. This environment also mimics the production setup as closely as possible but is used for ongoing manual exploration of the software. Used in: Step 7 (Exploration)                                                                                                                                                         |
| Production Environment                  | Purpose: This is the live environment where the software runs and serves end users. Used in: Step 10 (Deployment), Step 11 (Live), and Step 12 (Release)                                                                                                                                                                                              |

Apart from production and demonstration, the environments are created and removed on-demand, as part of the delivery workflow.

The infrastructure will be entirely described using code, ensuring that the same code is utilized across all environments. This approach guarantees consistency, repeatability, and minimizes the risk of errors caused by environment discrepancies. The only variation across environments will be the configuration, which is defined separately and tailored for each specific environment. A risk-based approach will be employed to determine and justify the different configuration options, prioritizing critical functionalities while balancing compliance, efficiency, and operational needs.

### Code Review

To ensure the software code is of high quality and meets predefined standards, the business application will use a combination of automated tools and a traditional code review performed by a Peer.

A static code analysis tool will analyze the source code to identify potential errors, bugs, stylistic errors, and other issues.

The code analysis tool is used to enforce coding standards and best practices, making the code more readable, maintainable, and less prone to errors.

The code analysis tool is automatically activated during **Pull request**.

The tools configuration and the result of the analysis will be retained and summarized in a development report (see [Release Approval](#release-approval)).

In addition to the code analysis tool, a Peer will perform a code review as part of the pull request approval to ensure that code meets the specified requirements,
design constraints and correctly implements the intended functionality.

Approval of the pull request constitutes the documentation for the code review.

---

## Specifications

<!--{% remove %}-->
!!! tip

    Please utilize the requirements templates provided alongside the O&M template.
<!--{% endremove %}-->

The solution uses several specification types to ensure requirements, design, configuration, and risk are clearly documented and traceable:

- **User Requirement Specification (URS):** Feature files written in Gherkin or Markdown format. The feature name itself (format: `<module>_<feature-name>`) serves as the URS identifier. Used to systematically capture requirements, acceptance criteria, specifications, and test instructions for each feature.
- **Functional Specification (FS):** Defines the functional behavior of the system. It is documented in feature files, represented as scenarios that outline the intended functionality.
- **Design Specification (DS):** Outlines the technical design and architecture of the system. It is documented as a standalone artifact called the Solution Design Documentation.
- **Configuration Specification (CS):** Captured as scripts, Infrastructure as Code (IaC), and versioned configuration files.

Feature files should be created collaboratively by QA, Business SME, team members, and the System Owner. This ensures that technical, compliance, and business concerns are all considered.

User Requirements are the highest-level specifications and must reflect the intended use (the problem to be solved) for the complete software product. Each requirement must be tagged if its source is GxP or a Critical Aspect (for GMP).

Feature files are plain text files stored in the same repository as the system code. The following tagging taxonomy is used:

**Testing Taxonomy Tags:**

| Tag Type                                  | Format                              | Purpose                                |
| ----------------------------------------- | ----------------------------------- | -------------------------------------- |
| Test Level                                | `@L0` - `@L4`                       | Define execution environment and scope |
| Verification (REQUIRED for all scenarios) | `@ov`, `@iv`, `@pv`, `@piv`, `@ppv` | Categorize validation type             |
| Test Execution Control                    | `@ignore`, `@Manual`                | Control test execution behavior        |
| System Dependencies                       | `@dep:docker`, `@dep:git`, etc.     | Declare required tooling               |
| Risk Controls                             | `@risk-control:<name>-<id>`         | Link to compliance requirements        |

**Regulatory Tags (for GxP environments):**

| Tag Type                   | Format                     | Purpose                                 |
| -------------------------- | -------------------------- | --------------------------------------- |
| GxP-Related Requirement    | `@gxp`                     | Identify GxP-controlled aspects         |
| Critical Aspect (GmP only) | `@critical-aspect`         | Mark critical aspects for GmP products  |
| GxP Risk Control           | `@risk-control:gxp-<name>` | Link to GxP risk control specifications |

If manual test cases are required, they will be stored alongside the automated test scenarios and marked with the tag `@Manual`. Test results will be fed back into the repository and stored alongside the automated test results, enabling visualization within a unified traceability matrix.

## Functional Risk Assessment

A functional risk assessment is required whenever a requirement in a feature file is tagged with `@gxp`. In these cases, a risk control specification must be created in and explicitly linked to the relevant requirement using `@risk-control:gxp-<name>`.

The risk control specification should include:

- **Risk Description:** What could go wrong with the feature.
- **Root Cause and Likelihood:** Identify the cause(s) and estimate the likelihood (Unlikely <30%, Possible 30-70%, Likely >70%).
- **Impact:** Assess the impact on the supported process (Insignificant, Moderate, or Critical impact on Product, Patient, or Data Integrity per GAMP 5 and ICH Q9 guidelines).
- **Risk Controls:** List the controls or mitigations, documented as test scenarios within the risk control specification.
- **Risk Classification:** Classify both gross and net risk as High, Medium, or Low.

Risk controls are implemented as test scenarios that verify the mitigations are effective. These may include additional requirements, negative test scenarios, challenge tests, or validation of generic controls.

For all requirements tagged as `@gxp`, there must be test scenarios for negative testing and/or challenge tests, or a clear justification for performing only positive tests.

---

## Verification

The delivery workflow (see [Development and Implementation Strategy](#development-and-implementation-strategy)) relies on test automation to provide developers with rapid feedback on their code changes.

The scenarios in the feature files are translated into automated test cases, using the same given/when/then format.

The scenarios that reflect user requirements are approved by the Quality Unit and the System Owner in the stage **pull request review**.

Additional, lower-level test cases, such as unit tests, are created and maintained by the team without involvement of QA or System Owner.

The delivery workflow enforces that the software is built only once (during **Commit**), and then progressively tested more extensively at each stage.

This means that the software being tested during development activities is identical to the software subsequently deployed to production, except for environment-specific configurations.

The high degree of control over the software development allows test activities from the software development process to be re-used as the verification tests (IV/OV/PV), as supported by GAMP 5 guidance and regulatory expectations for computerized systems validation.

The evidence for the successful deployment to the production-like test environments and production environment will be retained as evidence of installation verification (IV). [^1]

[^1]: Concretely, successful deployment outputs serve as Installation Verification evidence, because being able to successfully deploy implicitly verifies correct installation.

## Defects

If the test fails during any of the test stages, this will prevent the change from proceeding to the next stage, and thereby automatically prevent it from being released.

Any failure prompts an investigation to find and resolve the defect, for example updating a test case. Since all specifications and tests are controlled in the repository, the defect investigation and resolution are documented in the step **Authoring Change.**

When an author submits a fix for a defect to the controlled repository (trunk) the change description must at least include the following information:

- The word "defect" must be part of the title (to allow searching later)
- describe what happened
- investigate and describe the cause
- describe the actions taken

If a test of a requirement marked as `@CriticalAspect` fails after **Production Deployment**,
it must be managed as a validation deviation according to organizational quality procedures.

## Release Approval

At the release approval, a set of information must be available in a format that enables outsiders to understand the history of the development process, e.g., a non-technical, human-readable PDF.

- **Implementation Report**
  - Summary of changes since the last deployment
  - Updates to requirements (new or changed)
  - Conclusion on Fitness for Intended Use
  - Impact on Business Process
  - List of updated external dependencies
  - References to Change Requests
  - Final Design Review (comments from Merge Request approvals)
  - Change log (comments from all Merge Requests)
  - Requirements Specifications (URS/FS) and related risk assessments
  - Tests summary with a traceability matrix linking requirements to test results.
- **Design Documentation**
- **Supplier Information**
  - List of external suppliers and internal service providers
  - Supplier assessments and interface agreements per GAMP 5 supplier assessment requirements

The IT Solution will be **released for deployment and use**, when release approval is approved by QA and System Owner.

At the release approval all deliverables and documentation can be inspected as needed.

The release report and all other deliverables defined in this section must be uploaded and digital signed not later than one week after **Release Approval**.

Approval of the implementation report in does not block release, since intermediate approval is provided by QA and System Owner as part of **Release Approval** per organizational quality procedures.

## Roll back plan

If issues are detected after deployment, the rollback process can be initiated. The decision to initiate a rollback must be taken by the System Owner.

QA must be informed as soon as possible, although documentation for the information is not required since QA approves the release (of the roll back).

- Identify the last stable version (typically last release) in the repository.
- The code is built and tested to ensure it is still valid.
- If all tests pass, the previous version is deployed to the production environment.

After rollback, the system must be closely monitored to ensure the previous version is functioning correctly.

## IT Infrastructure

The software product may rely on IT Infrastructure components that are not considered as part of the software product itself and managed by another business unit or a supplier.

Such services are out of the scope of this plan and are qualified separately as defined in organizational IT infrastructure management procedures.

All infrastructure that is part of the software product is managed according to organizational IT system management procedures.
