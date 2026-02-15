# riskprofile

Creates OSCAL profiles from risk assessment documents using AI to map identified risks to NIST 800-53 controls from a catalog.

## Key Types

- **`Config`** -- Holds command configuration: assessment path, catalog URL, output path, force/debug flags, and workspace root
- **`Deps`** -- Injectable dependencies for testing; an `AIResponse` field bypasses AI when non-empty

## Key Functions

- `CreateRiskProfile` -- Entry point for the `create risk-profile` command; orchestrates the full workflow from reading the assessment file through AI generation to writing the OSCAL profile
- `generateProfile` -- Generates an OSCAL profile using AI with retry, or uses injected/mock responses for testing
- `buildProfilePrompt` -- Constructs the AI prompt using a three-tier template priority (command flag, team override, system default) with available controls context
- `callAIWithRetry` -- Calls the AI provider using the two-phase generation-with-retry framework
- `parseProfileFromAI` -- Parses AI response (full OSCAL profile JSON) into an OSCAL Profile struct
- `filterInvalidControls` -- Removes control IDs that don't exist in the catalog, rebuilding the profile with only valid controls
- `extractControlIDs` -- Extracts control IDs from AI response supporting multiple formats (JSON array, JSON object, or plaintext with pattern matching)

## Patterns

- **Command registration via init()**: Uses `registry.Register(CreateRiskProfile)` for automatic command discovery
- **Dependency injection for testing**: The `Deps` struct allows injecting AI responses to bypass real AI calls in tests
- **Three-tier prompt loading**: Prompt templates are loaded with priority: command flag, team override (`.eac/templates/ai/risk-profile/`), system default
- **Post-generation catalog validation**: After AI generates controls, `filterInvalidControls` validates each against the catalog and removes any that don't exist, rebuilding the profile with control info

## Internal Structure

| File | Responsibility |
| --- | --- |
| risk-profile.go | Command entry point, configuration parsing, orchestration workflow, catalog loading, profile writing, and success reporting |
| ai_generation.go | AI interaction: prompt building, retry-based generation, OSCAL profile parsing, and invalid control filtering |
| control_extraction.go | Control ID extraction from AI responses: JSON array/object parsing, markdown code block extraction, and NIST pattern matching fallback |
| deps.go | Injectable Deps struct for test dependency injection |

## Dependencies

- `go/adapters/ai` -- AI executor creation and adapter wrapping
- `go/adapters/ai/providers` -- Built-in AI provider registration
- `go/cli/eac/internal/risk/oscal` -- OSCAL catalog/profile loading, control extraction, validation, and profile document creation
- `go/clibase/flags` -- Shared flag parsing and validation
- `go/clibase/registry` -- Command registration and workspace root discovery
- `go/core/ai` -- Retry framework, AI config loading, contract loader, mock response support
- `go/core/config` -- Loading risk config for default catalog URL and EAC config for specs path
- `go/core/logging` -- Component logger for info, warning, and error output
- `go/core/paths` -- EAC config path resolution
- `github.com/defenseunicorns/go-oscal` -- OSCAL type definitions for profile and catalog structures

## Role in System

This package implements the `create risk-profile` command, which is the first step in the risk/compliance workflow. It takes a risk assessment document, uses AI to identify relevant NIST 800-53 controls, validates them against a catalog, and produces an OSCAL profile JSON file. This profile is then used downstream by `create risk-assess` to map tagged Gherkin scenarios to controls for compliance evidence.

## Code Health

### Tech Debt

- None identified

### Pain Points

- No test coverage for `ai_generation.go`, `control_extraction.go`, `deps.go`, `risk-profile.go`

### Optimization Opportunities

- None identified
