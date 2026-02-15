# godog-eac

The `godog-eac` module provides BDD test infrastructure for Godog integration, including test context, caching, and shared step definitions.

## System Context

Shows how godog-eac provides BDD infrastructure to test modules.

<!-- structurizr:adapters:godog:SystemContext -->

## Container Architecture

High-level view of the godog-eac packages.

<!-- structurizr:adapters:godog:Containers -->

## Component Architecture

### Test Context Components

Scenario initialization and mock configuration for BDD tests.

<!-- structurizr:adapters:godog:TestContextComponents -->

### Test Cache Components

Thread-safe test result caching with CI optimization.

<!-- structurizr:adapters:godog:TestCacheComponents -->

## Design File

- **Location**: `specs/godog-eac/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module godog-eac`

## Key Responsibilities

| Package    | Purpose                                             |
| ---------- | --------------------------------------------------- |
| context    | BDD test context with scenario init and mock config  |
| cache      | Thread-safe test cache with RWMutex protection       |
| steps      | Shared step definitions for cross-module BDD tests   |
| fixtures   | Test fixture pool and module setup helpers           |
| dispatcher | In-process command dispatcher for BDD tests          |
