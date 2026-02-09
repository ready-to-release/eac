# gotest-eac

The `gotest-eac` module provides Go test runner integration for unit and integration test execution.

## System Context

Shows how the GoTest adapter integrates with the test execution framework.

<!-- structurizr:gotest-eac:SystemContext -->

## Container Architecture

High-level view of the GoTest adapter packages.

<!-- structurizr:gotest-eac:Containers -->

## Design File

- **Location**: `specs/gotest-eac/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module gotest-eac`

## Key Responsibilities

| Package    | Purpose                                        |
| ---------- | ---------------------------------------------- |
| runner     | Go test execution with tag filtering and output |
| helpers    | Test helper utilities for runner configuration  |
| ctrf       | Converts Go test output to CTRF format          |
