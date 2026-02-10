# risk

Extracts control tag evidence from Gherkin feature files by scanning `@control:<id>` and `@controls:<id1>,<id2>` tags on scenarios.

## Key Types

- **`ControlEvidence`** -- Represents a control with its list of scenario evidence entries
- **`ScenarioEvidence`** -- Represents evidence from a single scenario, including feature path, feature name, scenario name, line number, tags, and test status

## Key Functions

- `ExtractControlEvidence` -- Scans all `.feature` files for a module and extracts control tag evidence, returning a map from control ID to ControlEvidence

## Patterns

- **Tag pattern matching**: Uses precompiled regexes to extract control IDs from `@control:<id>` (single) and `@controls:<id1>,<id2>` (multiple) tag formats
- **Line-by-line Gherkin parsing**: Lightweight parser that tracks feature name, collects scenario tags, and resets tag accumulation at scenario boundaries

## Internal Structure

| File | Responsibility |
| --- | --- |
| tag_extractor.go | All package functionality: ControlEvidence and ScenarioEvidence types, ExtractControlEvidence, file scanning, tag extraction, control ID parsing, and module feature file discovery |

## Dependencies

- `go/core/config` -- Loading repository configuration to resolve specs directory path
- `go/core/logging` -- Component logger for warnings on file extraction failures

## Role in System

This package supports the risk assessment workflow by extracting evidence of which Gherkin scenarios provide test coverage for specific NIST 800-53 (or custom catalog) controls. The extracted evidence is used by the `create risk-assess` command to build compliance reports showing which controls have passing test evidence.

## Code Health

### Tech Debt
- None

### Pain Points
- None identified

### Optimization Opportunities
- None identified
