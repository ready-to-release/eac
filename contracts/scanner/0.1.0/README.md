# scanner

Security scanner contracts defining scanner configuration, risk/compliance
profiles, and embedded default policies.

## Key Types

- **`SecurityConfigPort`** -- Access scanner definitions and policies
- **`ScannerPort`** -- Single scanner identity (image, category, timeout)
- **`RiskConfigPort`** -- OSCAL risk profiles and scoring configuration
- **`ProfilePort`** -- OSCAL profile with control IDs and metadata
- **`RiskScoringPort`** -- Impact and criticality scoring by module type
- **`ScannerDefinition`** -- Concrete scanner config (implements `ScannerPort`)
- **`PoliciesConfig`** -- Component-to-scanner mapping and skip rules

## Patterns

- Hexagonal ports: `SecurityConfigPort` and `RiskConfigPort` are read-only config ports
- Category constants: well-known scanner categories (sbom, vuln, sast, dast, etc.)
- Embedded defaults: policies, risk-config, and scanners YAML in schemas/defaults/

## Internal Structure

| File / Sub-directory             | Responsibility                                                              |
| -------------------------------- | --------------------------------------------------------------------------- |
| ports.go                         | `SecurityConfigPort`, `ScannerPort`, `RiskConfigPort`, `ProfilePort`        |
| types.go                         | `ScannerDefinition`, `ScannersConfig`, `PoliciesConfig`, category constants |
| risk_types.go                    | `RiskConfigPort`, `ProfilePort`, `RiskScoringPort`                          |
| schemas/defaults/scanners.yml    | Default scanner definitions                                                 |
| schemas/defaults/policies.yml    | Default component-to-scanner policies                                       |
| schemas/defaults/risk-config.yml | Default risk scoring configuration                                          |

## Dependencies

None -- this is a leaf contract module with no internal dependencies.

## Role in System

The `scanner` package (moniker: contracts-scanner) provides the contract
layer for the security scanning pipeline. The scan command and its adapters
consume `SecurityConfigPort` for scanner selection and `RiskConfigPort` for
compliance scoring against OSCAL control catalogs.

## Code Health

### Tech Debt

- No test files; add compile-time interface checks for `ScannerDefinition` implementing `ScannerPort`
- `ScannerDefinition.Timeout()` in types.go:55-61 silently swallows parse errors and defaults to 10m -- surface the error or log a warning

### Pain Points

- `ScannerPort` (8 methods) mixes identity (ID, Description) with runtime config (Image, Tag, Command, Timeout) -- consider splitting into identity and execution facets
- Scanner category constants in types.go are untyped strings; a `Category` type alias would enable compile-time validation

### Optimization Opportunities

- Introduce a `ScannerCategory` type with validation -- low effort, prevents typos in policy config
- Add `Validate()` to `ScannerDefinition` to catch missing Image or invalid Timeout at load time -- moderate effort, fails fast
