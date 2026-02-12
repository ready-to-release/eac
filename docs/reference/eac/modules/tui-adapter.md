# tui-eac

The `tui-eac` module implements the TUI port interface for parallel console rendering using Bubbletea.

## System Context

Shows how the TUI adapter provides terminal UI to CLI commands.

<!-- structurizr:adapters:tui-eac:SystemContext -->

## Container Architecture

High-level view of the TUI adapter packages.

<!-- structurizr:adapters:tui-eac:Containers -->

## Component Architecture

### Parallel Console Components

Bubbletea-based parallel console with phase tracking, tab rendering, and ring buffer output.

<!-- structurizr:adapters:tui-eac:ParallelConsoleComponents -->

## Design File

- **Location**: `specs/tui-eac/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module tui-eac`

## Key Responsibilities

| Package  | Purpose                                             |
| -------- | --------------------------------------------------- |
| console  | Parallel Bubbletea model with phase/UoW tracking     |
| render   | Icon, lamp, and style rendering for console output   |
| observer | Progress observation bridge to TUI updates           |
| hooks    | Lifecycle hooks for TUI event handling               |
