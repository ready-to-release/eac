# scoring

Provides risk scoring utilities following ISO 27005 methodology. Computes risk scores from vulnerability findings and test evidence, with optional AI-powered analysis for likelihood assessment.

## Key Types

- **`RiskBand`** -- Risk level category (Critical, High, Medium, Low) with score ranges
- **`RiskScore`** -- Computed risk information including likelihood, impact, score, band, reasoning, and confidence
- **`AIRiskAnalysis`** -- AI-generated risk analysis with computed likelihood, reasoning, summary, and recommended controls
- **`AIScorer`** -- AI-powered risk analysis engine with configurable provider and model
- **`AIAnalysisInput`** -- Complete input structure sent to AI for risk analysis
- **`VulnerabilityInput`** -- Vulnerability data for AI analysis (scanner, severity, CVE, CVSS, package)
- **`ModuleContext`** -- Module context for AI analysis (component type, criticality, existing controls)
- **`ControlFinding`** -- Finding information for a specific security control (state, test counts, vulnerability counts)
- **`Deps`** -- Injectable dependencies for testing (AI response bypass)

## Key Functions

- **`CalculateBaseLikelihood()`** -- Compute likelihood from vulnerability severity counts using weighted approach
- **`ApplyControlReductions()`** -- Reduce likelihood based on passing security controls
- **`CalculateRiskScore()`** -- Compute final risk score (likelihood x impact)
- **`DetermineRiskBand()`** -- Map risk score to band (Critical 20-25, High 12-19, Medium 6-11, Low 1-5)
- **`ComputeRiskScore()`** -- Calculate complete risk score from inputs
- **`GetDefaultImpact()`** -- Get default impact rating based on module type from configuration
- **`FormatRiskBandColor()`** -- Return ANSI color code for risk band display
- **`FormatRiskScore()`** -- Format risk score with ANSI color for terminal display
- **`FormatRiskScorePlain()`** -- Format risk score without ANSI colors for markdown/text output
- **`NewAIScorer()`** -- Create a new AI-powered risk scorer
- **`AnalyzeRisk()`** -- Perform AI-powered risk analysis on security findings
- **`ExtractVulnerabilityFindings()`** -- Extract vulnerability data from security scan evidence for AI analysis

## Patterns

- ISO 27005 risk methodology: Risk Score = Likelihood x Impact, with 1-5 scales
- Weighted severity scoring: critical (+4), high (+3), medium (+2), low (+1) vulnerability adjustments
- Control-based reduction: passing security controls reduce likelihood score
- AI-augmented analysis: optional AI provider analyzes findings for nuanced likelihood assessment
- Dependency injection for testing: `Deps.AIResponse` bypasses AI calls with canned responses
- Multi-scanner extraction: parses Trivy, GoSec, and ZAP scan formats for vulnerability input

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | Risk scoring types, calculation functions, band determination, and formatting (211 lines) |
| ai_scorer.go | AI-powered risk analysis with provider abstraction and prompt management |
| evidence_extractor.go | Vulnerability finding extraction from multiple scanner formats (Trivy, GoSec, ZAP) |
| deps.go | Injectable dependency container for test double injection |

## Dependencies

- `adapters/ai` -- AI client interface for risk analysis
- `adapters/ai/providers` -- AI provider implementations
- `cli/eac/internal/risk/evidence` -- evidence collection types for finding extraction
- `core/ai` -- AI mock support for testing
- `core/config` -- risk scoring configuration and default impact ratings
- `core/domain/modules` -- module registry for context building

## Role in System

The `scoring` package is the computation engine of the risk assessment pipeline. It takes vulnerability findings and test evidence as input and produces quantified risk scores following ISO 27005. The optional AI analysis layer provides more nuanced likelihood assessments when an AI provider is configured.

## Code Health

### Tech Debt

- None identified.

### Pain Points

- `evidence_extractor.go` is 254 lines due to handling multiple scanner formats (Trivy, GoSec, ZAP).

### Optimization Opportunities

- None identified.
