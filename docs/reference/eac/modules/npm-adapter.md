# npm-eac

The `npm-eac` module provides NPM package manager integration for dependency isolation in containerized builds.

## System Context

Shows how the NPM adapter provides dependency isolation to build commands.

<!-- structurizr:adapters:npm:SystemContext -->

## Container Architecture

High-level view of the NPM adapter packages.

<!-- structurizr:adapters:npm:Containers -->

## Component Architecture

### Isolation Components

NPM workspace isolation with directory sync, file copy, and install mutex.

<!-- structurizr:adapters:npm:IsolationComponents -->

## Design File

- **Location**: `specs/npm-eac/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module npm-eac`

## Key Responsibilities

| Package   | Purpose                                           |
| --------- | ------------------------------------------------- |
| isolation | NPM workspace isolation with mutex-protected install |
| sync      | Directory synchronization for isolated workspaces  |
| detector  | Package change detection for incremental installs  |
