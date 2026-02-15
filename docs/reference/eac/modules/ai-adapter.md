# ai-eac

The `ai-eac` module provides AI service integration, wrapping multiple LLM providers behind a unified interface.

## System Context

Shows how the AI adapter connects CLI commands to AI providers.

<!-- structurizr:adapters:ai:SystemContext -->

## Container Architecture

High-level view of the AI adapter packages.

<!-- structurizr:adapters:ai:Containers -->

## Component Architecture

### Provider Components

Provider registry with Anthropic, OpenAI, Gemini, Claude CLI, and Test providers.

<!-- structurizr:adapters:ai:ProvidersComponents -->

## Design File

- **Location**: `specs/ai-eac/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module ai-eac`

## Key Responsibilities

| Package    | Purpose                                          |
| ---------- | ------------------------------------------------ |
| executor   | Unified AI execution with retry and rate limiting |
| config     | Environment-substituted AI configuration loading  |
| providers  | Multi-provider registry (Anthropic, OpenAI, etc.) |
