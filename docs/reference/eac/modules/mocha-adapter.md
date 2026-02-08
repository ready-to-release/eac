# mocha-adapter

The `mocha-adapter` module provides Mocha test runner integration for TypeScript unit test execution.

## System Context

Shows how the Mocha adapter integrates with the test execution framework.

<!-- structurizr:mocha-adapter:SystemContext -->

## Container Architecture

High-level view of the Mocha adapter packages.

<!-- structurizr:mocha-adapter:Containers -->

## Design File

- **Location**: `specs/mocha-adapter/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module mocha-adapter`

## Key Responsibilities

| Package | Purpose                                       |
| ------- | --------------------------------------------- |
| runner  | Mocha test execution via Docker container      |
| ctrf    | Converts Mocha JSON output to CTRF format     |
