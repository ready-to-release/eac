# Implementation Report {{ .ProjectName }}

!!! note

    Decision on overall change type (Normal, Standard or Emergency)

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

!!! note

    Add a description of changes since the last deployment includes changes to existing requirements or newly added ones.

<!--{% raw %}-->
{{ .changed_requirements }}
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
{{ .req_approval_comments }}
<!--{% endraw %}-->

## Change Log

<!-- This section should be dynamically created by knowing last-release-commit-sha - basically git log.  -->

The change log contains changes from all Merge Requests included in the release.

<!--{% raw %}-->
{{ .release_notes }}
<!--{% endraw %}-->

---

## Requirements Specifications

This list includes all the requirements for the solution.

<!--{% raw %}-->
{{ .requirements }}
<!--{% endraw %}-->

## Design Documentation

Please refer to the Solution Design Documentation document also generated as part of the audit ready documentation.

---

## Tests Summary

This section shows requirements traceability from features through acceptance criteria to test execution results. Each test scenario is uniquely identifiable, with execution results (automated or manual) linked to the corresponding requirement.

<!--{% remove %}-->
!!! note "Example: Single Module Release"

    **Feature ID**: `acme-inventory_order-processing`
    **User Story**: As a warehouse operator, I want to process orders efficiently, so that I can fulfill customer requests on time
    **Specification**: [specification.feature](specs/acme-inventory/order-processing/specification.feature)

    **Rule 1: System validates order data correctly** (AC1)

    | Scenario | Tags | Result |
    |----------|------|--------|
    | Valid order is accepted | @success @ac1 @OV | 🟢 Passed |
    | Invalid order is rejected with clear error | @error @ac1 @OV | 🟢 Passed |
    | Missing product SKU returns 400 Bad Request | @error @ac1 @OV | 🟢 Passed |

    **Rule 2: Order processing system installs with default configuration** (AC2)

    | Scenario | Tags | Result |
    |----------|------|--------|
    | Installation configures database connection | @success @ac2 @IV | 🟢 Passed |
    | Installation validates configuration parameters | @success @ac2 @IV | 🟢 Passed |

    **Rule 3: Order processing meets performance requirements** (AC3)

    | Scenario | Tags | Result |
    |----------|------|--------|
    | Order processing completes within 2 seconds | @success @ac3 @PV | 🔴 Failed |
    | System handles 50 concurrent orders | @success @ac3 @PV | 🟢 Passed |
    | Processing maintains memory under 500MB | @success @ac3 @PV | 🟢 Passed |

!!! note "Example: Multi-Module Release"

    ### Module: acme-api

    **Feature ID**: `acme-api_product-search`
    **User Story**: As a customer, I want to search for products, so that I can find items to purchase
    **Specification**: [specification.feature](#product-search)

    **Rule 1** (AC1)

    | Scenario | Tags | Result |
    |----------|------|--------|
    | Search returns matching products | @acme-api @success @ac1 | 🟢 Passed |
    | Empty search returns all products | @acme-api @success @ac1 | 🟢 Passed |

    ---

    **Feature ID**: `acme-api_inventory-check`
    **User Story**: As a customer, I want to check product availability, so that I know if items are in stock
    **Specification**: [specification.feature](#inventory-check)

    **Rule 1** (AC1)

    | Scenario | Tags | Result |
    |----------|------|--------|
    | Available product shows in-stock status | @acme-api @success @ac1 | 🟢 Passed |
    | Out-of-stock product shows unavailable | @acme-api @error @ac1 | 🟢 Passed |

    ### Module: acme-reports

    **Feature ID**: `acme-reports_sales-summary`
    **User Story**: As a manager, I want to view sales summaries, so that I can track business performance
    **Specification**: [specification.feature](#sales-summary)

    **Rule 1** (AC1)

    | Scenario | Tags | Result |
    |----------|------|--------|
    | Daily summary displays total sales | @acme-reports @success @ac1 | 🟢 Passed |

<!--{% endremove %}-->

<!--{% raw %}-->
{{ .feature_test_results }}
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
{{ .iv_test_traceability_report }}
<!--{% endraw %}-->

---

#### Operational Verification (Traceability) - OV

The test results below display all scenarios **not** tagged with `@IV` or `@PV`.

<!--{% raw %}-->
{{ .ov_test_traceability_report }}
<!--{% endraw %}-->

---

#### Performance Verification - PV

Performance verification is conducted in a Production-Like Test Environment (PLTE) as well as in the production environment following the release. The test results below display all scenarios tagged with `@PV`.

<!--{% raw %}-->
{{ .pv_test_traceability_report }}
<!--{% endraw %}-->

---

## Appendix A: Specifications and Test Results

<!--{% remove %}-->
!!! note Example

    ### product-search

    ```gherkin
    @acme-api @inventory @search @critical
    Feature: Product Search

    As a customer
    I want to search for products
    So that I can find items to purchase

    Rule: Search must handle empty and non-empty queries

        @success @ac1
        Scenario: Search with keyword returns matching products
        When I run "acme search widget"
        Then I should see "Product found" or "No matches"

        @success @ac1
        Scenario: Empty search returns all products
        When I run "acme search"
        Then I should see "All products" or "Product catalog"

    Rule: Search results must be formatted for display

        @success @ac2
        Scenario: Results contain product details
        When I run "acme search widget"
        Then I should see "SKU" or "Price" or "No matches"

        @success @ac2
        Scenario: Results include availability status
        When I run "acme search widget"
        Then I should see "In Stock" or "Out of Stock" or "No matches"

    #### Test result for product-search

    For test results see cucumber.json in release artifact.

    ```
<!--{% endremove %}-->

<!--{% raw %}-->
{{ .specs_and_test_results }}
<!--{% endraw %}-->
