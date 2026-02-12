# create/risk-assess

Creates OSCAL assessment-results from existing test and security scan evidence, mapping `@control` tags in feature files to OSCAL control IDs and determining satisfied/not-satisfied status. Generates Markdown risk assessment reports with optional AI-powered executive summaries and risk scoring.

## Key Types

- **`AssessConfig`** -- Command configuration with modules, OSCAL profile path, max evidence age, parallel mode, and output directory
- **`ModuleAssessmentResult`** -- Per-module outcome with OSCAL assessment-results, evidence collection, control counts, risk score, and warnings
- **`AIRiskAssessmentInput`** -- Aggregate input for AI analysis with per-module vulnerability findings, control counts, and profile metadata
- **`AIRiskAssessmentOutput`** -- AI-generated executive summary with risk posture, key findings, critical modules, and per-module analysis
- **`RiskAssessmentReportData`** -- Template data structure for rendering Markdown risk assessment reports
- **`ExecutiveSummary`** -- Combined basic statistics and AI-generated content (risk posture, narrative, recommendations)
- **`ModuleReportData`** -- Per-module report data with formatted evidence, control lists, findings, and risk score

## Patterns

- Evidence-only assessment: reads existing test results and security scans without running tests or scans
- OSCAL compliance output: generates standard OSCAL assessment-results JSON with observations, findings, and control status
- Parallel module assessment: concurrent goroutines with indexed results for order-preserving parallel execution
- AI-enhanced risk scoring: AI generates executive summaries and per-module likelihood scores, with basic scoring fallback
- Three-tier template loading: team override (`.eac/templates/`), system default (`templates/`), with container-root awareness
- Evidence staleness detection: warns when test or security evidence exceeds configurable max age

## Internal Structure

| File               | Responsibility                                                                                |
| ------------------ | --------------------------------------------------------------------------------------------- |
| assess.go          | Command entry point, type definitions, and assessment orchestration                           |
| assess_config.go   | CLI flag parsing, module discovery/validation, and workspace root resolution                  |
| assess_evidence.go | Evidence collection and staleness detection, AI input building, and fallback scoring          |
| assess_oscal.go    | Aggregated OSCAL report writing, Markdown report generation, and template loading             |
| ai_analyzer.go     | AI risk assessment generation with retry, prompt building, and executive summary construction |
| evidence.go        | Evidence collection from test manifests and security scan results with age validation         |
| oscal.go           | OSCAL assessment-results document building with observations, findings, and control status    |
| parallel.go        | Parallel and sequential module assessment execution with order-preserving concurrency         |
| report_builder.go  | Template data construction from assessment results with evidence formatting                   |
| reporting.go       | Console summary display with per-module control counts, evidence details, and risk scores     |
| template_data.go   | Data structures for template rendering: report data, executive summary, module data, findings |

## Dependencies

- `adapters/ai` -- AI executor and provider registration for risk assessment generation
- `adapters/ai/providers` -- built-in AI provider registration
- `cli/eac/internal/risk/evidence` -- evidence collection, security result discovery, and control-test mapping
- `cli/eac/internal/risk/oscal` -- OSCAL document creation, profile loading, and state constants
- `cli/eac/internal/risk/scoring` -- risk score computation, vulnerability extraction, and module context
- `cli/eac/impl/internal/manifests/testview` -- test manifest loading for evidence collection
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration and workspace root resolution
- `clibase/template` -- Go template rendering for Markdown report generation
- `core/ai` -- prompt loading, retry generation, and AI config
- `core/config` -- EAC configuration for template paths and module directories
- `core/domain/modules` -- module registry loading for module discovery and validation
- `core/environments` -- environment variable constants for container root detection
- `core/logging` -- structured logging
- `core/paths` -- risk output paths, template directories, and contract schema paths

## Role in System

The `create risk-assess` command provides automated compliance assessment for `eac`, reading existing test and security evidence to produce OSCAL-standard assessment-results and Markdown risk reports. It bridges the gap between test/scan execution and compliance documentation, enabling teams to assess control satisfaction across all modules with AI-enhanced risk analysis and evidence staleness tracking.

## Code Health

### Tech Debt

- `oscal.go` (367 lines), `evidence.go` (328 lines) exceed 300 lines

### Pain Points

- No test coverage for `assess.go`, `assess_config.go`, `assess_evidence.go`, `assess_oscal.go`, `evidence.go`, `oscal.go`, `parallel.go`, `reporting.go`, `template_data.go`

### Optimization Opportunities

- None identified
