# Commit Validation

> **Feature ID**: src_mcp_vscode_commit_validation
> **BDD Scenarios**: See feature files below
> **Module**: src-mcp-vscode
> **Tags**: mcp, vscode, validation, critical

## BDD Feature Files

This feature is split across multiple focused files for maintainability:

- [format_validation.feature](./format_validation.feature) - Heading, semantic format, line lengths
- [completeness_validation.feature](./completeness_validation.feature) - File and module completeness
- [contract_validation.feature](./contract_validation.feature) - YAML blocks, contracts, error handling

## User Story

* As a commit generation system
* I want to validate generated commit messages against quality standards
* So that all commits are consistent, well-formatted, and complete

## Acceptance Criteria

* Validates top-level heading exists (MD041)
* Validates semantic format: `<module>: <type>: <description>`
* Validates commit subject length (≤72 characters)
* Validates body line wrapping (≤72 characters per line)
* Validates file completeness (1-1 mapping with git context)
* Validates module completeness (all modules represented)
* Validates YAML code blocks are properly closed
* Validates module contracts are followed
* Returns array of validation errors with details
* Returns empty array when validation passes

## Acceptance Tests

### AC1: Validates top-level heading exists (MD041)
**Validated by**: format_validation.feature -> @ac1 scenarios

Tags: critical, validation, markdown

* Parse commit message structure
* Check if first non-empty line is a heading
* Verify heading uses single # character
* Return MD041 error if no top-level heading
* Pass validation if heading is present

### AC2: Validates semantic format
**Validated by**: format_validation.feature -> @ac2 scenarios

Tags: critical, validation, semantic

* Parse module section headings
* Extract format: `<module>: <type>: <description>`
* Verify type is valid (feat, fix, docs, refactor, test, chore, build, ci, perf, style)
* Verify colon separators are present
* Return semantic format error if invalid
* Pass validation if format is correct

### AC3: Validates commit subject length
**Validated by**: format_validation.feature -> @ac3 scenarios

Tags: validation, formatting

* Extract top-level heading text (without # character)
* Count character length
* Verify length is ≤72 characters
* Return subject length error if too long
* Pass validation if within limit

### AC4: Validates body line wrapping
**Validated by**: format_validation.feature -> @ac4 scenarios

Tags: validation, formatting

* Parse each line in commit body
* Count characters per line
* Exclude code blocks from wrapping check
* Exclude markdown tables from wrapping check
* Return line wrapping error for lines >72 characters
* Include line number in error
* Pass validation if all lines within limit

### AC5: Validates file completeness
**Validated by**: completeness_validation.feature -> @ac5 scenarios

Tags: critical, validation, completeness

* Extract files from commit message file tables
* Compare with files in GitContext.Changes
* Detect missing files (in git but not in commit)
* Detect extra files (in commit but not in git)
* Return file completeness errors with file paths
* Pass validation if 1-1 mapping exists

### AC6: Validates module completeness
**Validated by**: completeness_validation.feature -> @ac6 scenarios

Tags: critical, validation, completeness

* Extract modules from GitContext.Changes
* Extract module sections from commit message
* Verify all git modules have corresponding sections
* Return module completeness error if modules missing
* Pass validation if all modules represented

### AC7: Validates YAML blocks are closed
**Validated by**: contract_validation.feature -> @ac7 scenarios

Tags: validation, formatting

* Scan commit for code block markers (```)
* Track opening and closing of YAML blocks
* Verify all ```yaml blocks have closing ```
* Return YAML block error if unclosed
* Pass validation if all blocks properly closed

### AC8: Validates module contracts
**Validated by**: contract_validation.feature -> @ac8 scenarios

Tags: validation, contracts

* Load module contract from contracts/deployable-units/
* Parse contract requirements for module
* Verify commit adheres to contract rules
* Return contract violation errors with details
* Pass validation if contract is followed

### AC9: Returns array of validation errors
**Validated by**: contract_validation.feature -> @ac9 scenarios

Tags: critical, validation, errors

* Collect all validation errors during scan
* Create CommitValidationError object for each error
* Include error code, message, line number
* Return array of all errors found
* Preserve error order

### AC10: Returns empty array when valid
**Validated by**: contract_validation.feature -> @ac10 scenarios

Tags: critical, validation

* Run all validation checks on compliant commit
* Verify no errors are detected
* Return empty errors array
* Indicate validation passed
