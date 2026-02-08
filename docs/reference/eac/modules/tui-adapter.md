# tui-adapter

The `tui-adapter` module implements the TUI port interface for parallel console rendering using Bubbletea.

## System Context

Shows how the TUI adapter provides terminal UI to CLI commands.

<!-- structurizr:tui-adapter:SystemContext -->

## Container Architecture

High-level view of the TUI adapter packages.

<!-- structurizr:tui-adapter:Containers -->

## Component Architecture

### Parallel Console Components

Bubbletea-based parallel console with phase tracking, tab rendering, and ring buffer output.

<!-- structurizr:tui-adapter:ParallelConsoleComponents -->

## Design File

- **Location**: `specs/tui-adapter/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module tui-adapter`

## Key Responsibilities

| Package  | Purpose                                             |
| -------- | --------------------------------------------------- |
| console  | Parallel Bubbletea model with phase/UoW tracking     |
| render   | Icon, lamp, and style rendering for console output   |
| observer | Progress observation bridge to TUI updates           |
| hooks    | Lifecycle hooks for TUI event handling               |
