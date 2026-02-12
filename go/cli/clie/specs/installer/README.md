# specs/installer

BDD specifications for extension installer behavior using Godog (Cucumber for Go).

## Key Types

_No exported types. This directory contains only test files._

## Patterns

- Godog BDD: Step definitions map Gherkin scenarios to Go test functions
- Test harness: `godog_test.go` provides the `TestFeatures` entry point with Godog options
- Build-tagged: Tests run under specific test suite configurations

## Internal Structure

| File           | Responsibility                                     |
| -------------- | -------------------------------------------------- |
| steps_test.go  | Godog step definitions for installer BDD scenarios |
| godog_test.go  | Test harness with TestFeatures entry point         |

## Dependencies

_Test-only dependencies (Godog, testing infrastructure)._

## Role in System

The specs/installer directory contains BDD acceptance tests for the extension installer workflow. These tests verify the end-to-end behavior of extension installation including image pulling, configuration updates, and error handling scenarios. They are part of the clie module's specification-driven testing approach.

## Code Health

### Tech Debt

- None identified.

### Pain Points

- None identified.

### Optimization Opportunities

- None identified.
