# Intended Use {{ .ProjectName }}

**Prepared by:** [Insert name]
**Date:** [Insert date]

**Reviewed by:** [Insert name]
**Date:** [Insert date]

---

<!--{% remove %}-->
!!! tip

    - Review all sections when system capabilities change
    - Update integration details when adding new platform support
    - Maintain consistent terminology throughout the document
    - Ensure technical accuracy while keeping language accessible
    - Focus on WHAT the system is and does, not HOW it's implemented
    - This document serves as the foundation for understanding a system's purpose and scope

    The intended use document describes the system from the user/business perspective. Technical implementation details belong in the Implementation Plan.
<!--{% endremove %}-->

## System/Service Overview

<!--{% remove %}-->
!!! note

    - Begin with a concise 1-2 sentence description of what the system or service is
    - Specify its primary function and where it sits in the technology landscape
    - Keep language accessible to non-technical stakeholders

<!--{% endremove %}-->

## System Criticality

<!--{% remove %}-->
!!! note

    Classify the system as one of the following:

    - **GxP-critical**: Systems that directly impact product quality, patient safety, or regulatory compliance in pharmaceutical/medical contexts
    - **Other business-critical**: Systems essential for business operations but not subject to GxP regulations
    - **Non-business-critical**: Systems that support operations but whose failure would not significantly impact business or compliance

    This classification determines the level of validation, documentation, and change control required.
    Include justification for the selected criticality level.
<!--{% endremove %}-->

**Criticality Classification:** [Select: GxP-critical | Business-critical | Non-business-critical]

**Justification:**

---

## Target Users and Stakeholders

<!--{% remove %}-->
!!! note

    Identify who will use the system and who has a stake in its operation:

    - **Primary Users**: Who directly interacts with the system
    - **Secondary Users**: Who benefits from or is affected by the system
    - **Stakeholders**: Who has oversight or compliance responsibilities

    Include roles, departments, or organizational units.
<!--{% endremove %}-->

**Primary Users:**

**Secondary Users:**

**Stakeholders:**

---

## Core Functionality

<!--{% remove %}-->
!!! note

    Describe the main capabilities in plain language. Focus on what the system does from the user's perspective, not technical implementation details.

    Use clear, active language:
    - "The system enables users to..."
    - "The system provides..."
    - "The system automates..."
<!--{% endremove %}-->

## Use Cases and Scenarios

<!--{% remove %}-->
!!! note

    Provide concrete examples of how the system is used in practice. Each use case should describe:

    - The user or actor
    - The goal or objective
    - The expected outcome

<!--{% endremove %}-->

### Use Case 1: [Title]

**Actor:** [User role]
**Goal:** [What they want to accomplish]
**Outcome:** [Expected result]

---

## Integration Details

<!--{% remove %}-->
!!! note

    Document how the system interacts with other platforms, systems, or services:

    - List integrated systems and platforms
    - Explain the nature of each integration (authentication, data exchange, orchestration)
    - Describe where this system sits in the overall workflow or ecosystem
    - Specify protocols or standards used for integration
<!--{% endremove %}-->

## Data and Metadata Handling

<!--{% remove %}-->
!!! note

    Create structured lists of what data the system processes:

    - Types of data handled (personal data, GxP data, operational data)
    - Data categories and classifications
    - Metadata tracked or managed
    - Data sensitivity levels
    - Data flow (inputs and outputs)

    Use bullet points for clarity. Be specific but avoid implementation details.
<!--{% endremove %}-->

## System Boundaries and Scope

<!--{% remove %}-->
!!! note

    Clearly state what the system DOES and DOES NOT do. This prevents scope creep and clarifies responsibilities.

    Define:
    - **In Scope**: What the system is responsible for
    - **Out of Scope**: What the system explicitly does NOT handle
    - **Handoff Points**: Where responsibility transfers to/from other systems

<!--{% endremove %}-->

### In Scope

### Out of Scope

### Handoff Points

---

## Prerequisites and Dependencies

<!--{% remove %}-->
!!! note

    Document what must be in place for the system to function:

    - Infrastructure dependencies (cloud platforms, networks)
    - External systems that must be available
    - User prerequisites (accounts, training, permissions)
    - Data prerequisites (existing data sources, configurations)
<!--{% endremove %}-->

## Expected Benefits and Value

<!--{% remove %}-->
!!! note

    Articulate the value proposition and benefits:

    - Business benefits (efficiency, cost savings, compliance)
    - User benefits (ease of use, time savings)
    - Organizational benefits (standardization, risk reduction)
    - Quantify where possible
<!--{% endremove %}-->

## Limitations and Constraints

<!--{% remove %}-->
!!! note

    Be transparent about known limitations:

    - Functional limitations (what it cannot do)
    - Performance constraints (scale, throughput, latency)
    - Operational constraints (maintenance windows, availability)
    - Known issues or workarounds
<!--{% endremove %}-->

## Performance Expectations

<!--{% remove %}-->
!!! note

    Define performance expectations from a user perspective (not technical SLAs):

    - Response time expectations
    - Availability expectations
    - Scalability expectations (number of users, volume of data)
    - Acceptable performance degradation scenarios
<!--{% endremove %}-->

## Compliance and Security Considerations

<!--{% remove %}-->
!!! note

    Document support for regulatory and security requirements:

    - Regulatory frameworks supported (GxP, GDPR, HIPAA, etc.)
    - Data sensitivity levels supported
    - Security features relevant to users (encryption, audit trails, access controls)
    - Compliance responsibilities (user vs. system)
    - Validation or qualification status
<!--{% endremove %}-->

**Regulatory Frameworks:**

**Security Features:**

**User Compliance Responsibilities:**

**System Validation Status:**

---

## Training and Competency Requirements

<!--{% remove %}-->
!!! note

    Describe what training or competencies users need:

    - Required training programs
    - Prerequisite knowledge or skills
    - Certification requirements
    - Ongoing training needs
    - Reference materials and documentation
<!--{% endremove %}-->

## Success Criteria

<!--{% remove %}-->
!!! note

    Define how to measure if the system meets its intended use:

    - Functional success criteria (capabilities delivered)
    - User adoption metrics
    - Performance targets
    - Compliance metrics
    - Business outcomes
<!--{% endremove %}-->

## Lifecycle and Evolution

<!--{% remove %}-->
!!! note

    Set expectations for system lifecycle:

    - Expected lifespan or review period
    - Update and enhancement frequency
    - End-of-life considerations
    - Migration or transition plans (if applicable)
<!--{% endremove %}-->