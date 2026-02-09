# risk

Risk scoring subsystem that collects security and test evidence, manages OSCAL compliance documents, and computes ISO 27005 risk scores with optional AI analysis.

## Key Types

- **`EvidenceCollection`** -- Aggregated test and security evidence for a module
- **`SecurityResults`** -- Paths to scan result files for a module
- **`RiskScore`** -- Computed risk with likelihood, impact, and band
- **`AIScorer`** -- Performs AI-powered risk likelihood analysis
- **`ControlInfo`** -- Catalog control metadata for profile generation
- **`ControlTestEvidence`** -- Test pass/fail evidence for a specific control
- **`BehaviorGroup`** -- Groups enable/disable flag pairs for behavior

## Patterns

- Best-effort evidence collection: Missing scanners produce partial results rather than errors
- ISO 27005 scoring: Risk = Likelihood x Impact, mapped to Critical/High/Medium/Low bands
- OSCAL compliance: Uses go-oscal types with aliases for profiles and assessment results
- File locking: Concurrent assessment-results writes use exclusive file locks with atomic rename

## Internal Structure

| Sub-package | Responsibility |
| --- | --- |
| evidence/ | Discovers and loads test results and security scan files |
| oscal/ | Loads, writes, and creates OSCAL profiles, catalogs, and assessment results |
| scoring/ | Computes risk scores, extracts vulnerability findings, and invokes AI analysis |

## Dependencies

- `adapters/ai` -- AI provider execution for risk analysis
- `core/config` -- evidence age policies and risk scoring defaults
- `core/paths` -- repository path conventions for risk and spec directories
- `core/domain/modules` -- module registry for building AI context
- `core/logging` -- structured debug logging
- `core/ai` -- mock AI provider support and contract-based prompt loading

## Role in System

The risk sub-packages provide the domain logic behind the `validate risk-profile` and `validate risk-catalog` commands in `eac`. Evidence collection discovers scan outputs and test manifests, the OSCAL layer manages compliance documents against NIST 800-53 catalogs, and scoring computes quantitative risk ratings. The AI scorer optionally enriches likelihood estimates using LLM analysis of vulnerability findings.

## Code Health

### Tech Debt
- `scoring/ai_scorer.go:22`: `mockAIResponse` is package-level mutable global state used for testing; prefer dependency injection or test doubles
- `oscal/oscal_helpers.go` `GetProfileControlIDs` and `oscal/profile_helpers.go` `GetControlIDsFromProfile` duplicate control-ID extraction logic with nearly identical implementations
- `oscal/catalog_helpers.go:17`: `catalogLog` is a package-level global logger instance

### Pain Points
- `scoring/ai_scorer.go`: `parseAIResponse` silently falls back to a default likelihood of 3 when JSON parsing fails, which may mask malformed AI responses

### Optimization Opportunities
- Consolidate the two profile control-ID extraction functions into one, removing the duplicate -- low effort
- Replace global mock state in `ai_scorer.go` with an interface-based AI provider to improve testability -- moderate effort
