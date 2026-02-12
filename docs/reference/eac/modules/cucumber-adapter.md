# cucumber-eac

The `cucumber-eac` module provides Cucumber/Gherkin test runner integration for TypeScript BDD execution.

## System Context

Shows how the Cucumber adapter integrates with the test execution framework.

<!-- structurizr:adapters:cucumber-eac:SystemContext -->

## Container Architecture

High-level view of the Cucumber adapter packages.

<!-- structurizr:adapters:cucumber-eac:Containers -->

## Design File

- **Location**: `specs/cucumber-eac/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module cucumber-eac`

## Key Responsibilities

| Package       | Purpose                                     |
| ------------- | ------------------------------------------- |
| runner        | Cucumber test execution via Docker container |
| tag_translator | Translates EAC tags to Cucumber tag format  |
