# ai-adapter

The `ai-adapter` module provides AI service integration, wrapping multiple LLM providers behind a unified interface.

## System Context

Shows how the AI adapter connects CLI commands to AI providers.

<!-- structurizr:ai-adapter:SystemContext -->

## Container Architecture

High-level view of the AI adapter packages.

<!-- structurizr:ai-adapter:Containers -->

## Component Architecture

### Provider Components

Provider registry with Anthropic, OpenAI, Gemini, Claude CLI, and Test providers.

<!-- structurizr:ai-adapter:ProvidersComponents -->

## Design File

- **Location**: `specs/ai-adapter/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module ai-adapter`

## Key Responsibilities

| Package    | Purpose                                          |
| ---------- | ------------------------------------------------ |
| executor   | Unified AI execution with retry and rate limiting |
| config     | Environment-substituted AI configuration loading  |
| providers  | Multi-provider registry (Anthropic, OpenAI, etc.) |
