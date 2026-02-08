# clibase

The `clibase` module provides shared CLI infrastructure for orchestration, flags, command framework, and rendering.

It is the foundation for command execution across `eac-cli` and other CLI modules.

## System Context

Shows how clibase provides CLI infrastructure to command modules.

<!-- structurizr:clibase:SystemContext -->

## Container Architecture

High-level view of the clibase framework packages.

<!-- structurizr:clibase:Containers -->

## Component Architecture

### Command Framework

Five-phase command lifecycle (Register, Discover, Resolve, Schedule, Execute).

<!-- structurizr:clibase:CommandFramework -->

### Orchestrator Components

Parallel execution engine with dependency-aware scheduling.

<!-- structurizr:clibase:OrchestratorComponents -->

### Flags Components

Composable flag system for consistent CLI interfaces.

<!-- structurizr:clibase:FlagsComponents -->

### Render Components

Structured output rendering (table, JSON, YAML, markdown).

<!-- structurizr:clibase:RenderComponents -->

### Lock Tracker Components

Distributed lock lifecycle tracking with TTL and stale detection.

<!-- structurizr:clibase:LockTrackerComponents -->

## Design File

- **Location**: `specs/clibase/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module clibase`

## Key Responsibilities

| Package       | Purpose                                              |
| ------------- | ---------------------------------------------------- |
| cmdframework  | Five-phase command lifecycle and registration         |
| orchestrator  | Parallel build/test execution with dependency graphs  |
| flags         | Composable flag definitions for consistent CLI        |
| render        | Multi-format output rendering                        |
| registry      | Component-to-handler registry                        |
| services      | Shared service container for dependency injection     |
| locktracker   | Distributed lock lifecycle with TTL cleanup           |
