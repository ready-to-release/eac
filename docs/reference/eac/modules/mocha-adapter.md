# mocha-eac

The `mocha-eac` module provides Mocha test runner integration for TypeScript unit test execution.

## System Context

Shows how the Mocha adapter integrates with the test execution framework.

<!-- structurizr:adapters:mocha:SystemContext -->

## Container Architecture

High-level view of the Mocha adapter packages.

<!-- structurizr:adapters:mocha:Containers -->

## Design File

- **Location**: `specs/mocha-eac/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module mocha-eac`

## Key Responsibilities

| Package | Purpose                                       |
| ------- | --------------------------------------------- |
| runner  | Mocha test execution via Docker container      |
| ctrf    | Converts Mocha JSON output to CTRF format     |
