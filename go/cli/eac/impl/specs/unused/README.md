# unused

Implements the `specs unused-steps` command that discovers and reports step definitions in godog BDD test files that are not referenced by any feature file scenario.

## Key Types

- **`ImplSpecsPair`** -- Pairs a godog implementation directory with its corresponding specs directory
- **`StepDefinition`** -- Represents a single godog step definition with its pattern, file location, and line number
- **`FeatureStep`** -- Represents a step from a Gherkin feature file with its keyword and text
- **`UnusedStep`** -- A step definition identified as unused (no matching feature step)
- **`AnalysisResult`** -- Full analysis result containing all unused steps across all spec pairs
- **`PairResult`** -- Analysis result for a single impl/specs pair

## Key Functions

- **`SpecsUnusedSteps()`** -- Entry point for the `specs unused-steps` command
- **`runAnalysis()`** -- Coordinate discovery, parsing, and matching across all pairs
- **`printResults()`** -- Format and display unused step analysis results
- **`DiscoverPairs()`** -- Walk the repository to find godog impl/specs directory pairs
- **`ParseStepDefinitions()`** -- Parse Go files to extract godog step registration patterns
- **`ParseFeatureFiles()`** -- Parse Gherkin feature files to extract scenario steps
- **`FindUnusedSteps()`** -- Compare step definitions against feature steps to identify unused ones
- **`NormalizeForMatching()`** -- Normalize step text for fuzzy matching between definitions and usage
- **`MatchesStep()`** -- Check if a step definition pattern matches a feature step text

## Patterns

- Discovery-parse-match pipeline: discovers pairs, parses both sides, then matches to find unused steps
- Regex-based step matching: converts godog step patterns to regex for matching against feature text
- Normalization for fuzzy matching: strips parameters and normalizes whitespace for more accurate matching

## Internal Structure

| File | Responsibility |
| --- | --- |
| unused.go | Command entry point, analysis orchestration, and result formatting |
| discovery.go | Repository walking to discover godog impl/specs directory pairs |
| step_parser.go | Go source file parsing to extract godog step definitions |
| feature_parser.go | Gherkin feature file parsing to extract scenario steps |
| matcher.go | Step definition vs feature step matching and unused step identification |

## Dependencies

- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration and workspace root
- `core/logging` -- structured logging

## Role in System

The `unused` sub-package provides a code hygiene tool for BDD test suites. By identifying step definitions that no feature file references, it helps developers clean up dead test code and maintain a lean, focused test implementation.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- Step matching uses regex conversion which may produce false negatives for complex godog patterns with custom argument types

### Optimization Opportunities
- Cache parsed feature steps across pairs that share the same specs directory (low effort, reduces redundant parsing)
