# Organizing Specifications

> **How to structure and organize Gherkin specification files**

Learn how to organize specifications for maintainability, clarity, and discoverability.

---

## Overview

This section covers:

- **File Structure** - Separation of specifications from implementation
- **Organizing Rules** - Creating measurable acceptance criteria
- **Organizing Scenarios** - Coverage strategy and independence
- **Naming Conventions** - Feature and file naming standards
- **Size Guidelines** - Rule and scenario count limits

---

## In This Section

| Topic                                             | Description                                 |
| ------------------------------------------------- | ------------------------------------------- |
| [File Structure](./file-structure.md)             | Specifications vs implementation separation |
| [Organizing Rules](./organizing-rules.md)         | Rule blocks and acceptance criteria         |
| [Organizing Scenarios](./organizing-scenarios.md) | Scenario structure and independence         |
| [Naming Conventions](./naming-conventions.md)     | Feature and file naming standards           |
| [Size Guidelines](./size-guidelines.md)           | Rule and scenario count limits              |

---

## Key Principles

### 1. Specifications separate from implementation

- Specifications (WHAT) in `specs/` directory
- Implementation (HOW) in `src/` or module directories
- Business can review specs without seeing code

### 2. Features contain Rules, Rules contain Scenarios

```gherkin
Feature: module_feature-name
  Rule: Acceptance Criterion 1
    Scenario: Happy path
    Scenario: Error case
  Rule: Acceptance Criterion 2
    Scenario: Edge case
```

### 3. Size Guidelines

- **2-6 Rules per feature** (optimal readability)
- **2-4 Scenarios per Rule** (focused testing)
- **10-20 Scenarios per feature** (manageable file size)

### 4. Feature naming: `module_feature-name`

- Lowercase, kebab-case
- Module prefix for traceability
- Matches file system organization

---

## Quick Start

### Creating a New Feature

1. **Run Example Mapping workshop** - Discover requirements
2. **Create feature directory** - `specs/<module>/<feature>/`
3. **Copy template** - `templates/specs/specification.feature`
4. **Add Rules** - Blue cards → Rule blocks
5. **Add Scenarios** - Green cards → Scenarios
6. **Add Tags** - Verification tags + dependencies

### Example Feature Structure

```gherkin
@cli @critical
Feature: cli_init-project

  As a developer
  I want to initialize a CLI project
  So that I can quickly start development

  Rule: Creates project directory structure

    @ov
    Scenario: Initialize in empty directory
      Given I am in an empty folder
      When I run "cli init"
      Then a file named "cli.yaml" should be created

    @ov
    Scenario: Initialize in existing project
      Given I am in a directory with "cli.yaml"
      When I run "cli init"
      Then the command should fail
```

---

## Related Documentation

- [BDD Fundamentals](../concepts/bdd-fundamentals.md) - Understanding BDD
- [Three-Layer Approach](../concepts/three-layer-approach.md) - Rules/Scenarios/Unit Tests
- [Example Mapping](../discovery/example-mapping.md) - Discovery workshops
- [Testing Taxonomy](../taxonomy/index.md) - Tags and test levels
