# toolhandler

AI tool handler that bridges the tool system to the AI adapter for automated module analysis.

## Key Types

- **`AIToolHandler`** -- Implements `BuildHandler` for AI-driven analysis of DSL, specs, source, and docs
- **`AIAnalysisType`** -- Enum for analysis types (dsl, specs, source, docs)

## Key Functions

- `NewAIToolHandler` -- Creates a handler with analysis type inferred from tool ID
- `NewAIToolHandlerWithType` -- Creates a handler with an explicit analysis type

## Patterns

- Tool ID inference: Extracts analysis type from tool ID (e.g., "ai-dsl-analyzer" -> "dsl")
- Multi-stage analysis: Source analysis loads prior DSL and specs results for context
- Content loading: Loads files by extension from module component roots
- Prompt construction: Builds analysis-specific prompts with prior context and instructions
- Retry with generation: Uses `coreai.GenerateWithRetry` for resilient AI execution

## Internal Structure

| File | Responsibility |
| --- | --- |
| handler.go | `AIToolHandler` implementing `BuildHandler` with content loading, prompt building, and AI execution |

## Dependencies

- `contracts/core/0.1.0` -- `ModuleContractPort` for module component access
- `adapters/ai` -- `Executor` and `ExecutorAdapter` for AI provider execution
- `adapters/ai/providers` -- `RegisterBuiltIn` for provider registration
- `core/ai/generation` -- `GenerateWithRetry` and retry configuration
- `core/tool` -- `ToolDefinition` and `BuildOptions`

## Role in System

The toolhandler sub-package enables AI-powered automated analysis of module artifacts (architecture DSL, BDD specs, source code, documentation). It is registered as a build handler in the tool system, allowing the orchestrator to schedule AI analysis as part of build pipelines. Each analysis type loads relevant files, constructs a prompt, and writes the AI-generated analysis to the output directory.

## Code Health

### Tech Debt
- handler.go is ~460 lines combining content loading, prompt building, and execution logic; extracting `contentLoader` and `promptBuilder` helpers would improve clarity

### Pain Points
- `loadFilesWithExtensions` in handler.go:256 silently skips walk errors and limits to 50 files with no user feedback when the limit is hit

### Optimization Opportunities
- None identified
