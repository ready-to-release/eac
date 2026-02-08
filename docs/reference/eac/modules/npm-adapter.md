# npm-adapter

The `npm-adapter` module provides NPM package manager integration for dependency isolation in containerized builds.

## System Context

Shows how the NPM adapter provides dependency isolation to build commands.

<!-- structurizr:npm-adapter:SystemContext -->

## Container Architecture

High-level view of the NPM adapter packages.

<!-- structurizr:npm-adapter:Containers -->

## Component Architecture

### Isolation Components

NPM workspace isolation with directory sync, file copy, and install mutex.

<!-- structurizr:npm-adapter:IsolationComponents -->

## Design File

- **Location**: `specs/npm-adapter/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module npm-adapter`

## Key Responsibilities

| Package   | Purpose                                           |
| --------- | ------------------------------------------------- |
| isolation | NPM workspace isolation with mutex-protected install |
| sync      | Directory synchronization for isolated workspaces  |
| detector  | Package change detection for incremental installs  |
