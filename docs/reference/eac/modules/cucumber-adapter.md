# cucumber-adapter

The `cucumber-adapter` module provides Cucumber/Gherkin test runner integration for TypeScript BDD execution.

## System Context

Shows how the Cucumber adapter integrates with the test execution framework.

<!-- structurizr:cucumber-adapter:SystemContext -->

## Container Architecture

High-level view of the Cucumber adapter packages.

<!-- structurizr:cucumber-adapter:Containers -->

## Design File

- **Location**: `specs/cucumber-adapter/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module cucumber-adapter`

## Key Responsibilities

| Package       | Purpose                                     |
| ------------- | ------------------------------------------- |
| runner        | Cucumber test execution via Docker container |
| tag_translator | Translates EAC tags to Cucumber tag format  |
