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

| File               | Responsibility                                                                             |
| ------------------ | ------------------------------------------------------------------------------------------ |
| handler.go         | `AIToolHandler` implementing `BuildHandler` with AI execution orchestration                |
| handler_loader.go  | Content loading: `loadFilesWithExtensions` and file collection from module component roots |
| handler_prompts.go | Prompt construction: analysis-specific prompt building with prior context and instructions |

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
- None identified

### Pain Points
- handler_loader.go is 203 lines; candidate for splitting by file type (e.g., separate DSL, spec, and source loading logic)

### Optimization Opportunities
- None identified
