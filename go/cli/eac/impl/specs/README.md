# specs

Provides Gherkin specification validation and security utilities for feature file processing. Handles path validation, feature name extraction, and Gherkin content syntax validation.

## Key Types

- **`gherkinValidationState`** -- Tracks state during line-by-line Gherkin parsing (in-scenario, has-given, etc.)

## Key Functions

- **`ExtractFeatureName()`** -- Extract a sanitized feature name from a feature file path
- **`ValidateFeatureLineSecurity()`** -- Validate a feature file path against directory traversal and injection attacks
- **`ValidateWindowsReservedName()`** -- Check for Windows reserved filenames (CON, PRN, NUL, etc.)
- **`ValidateOutputPath()`** -- Validate that an output path is safe and within expected bounds
- **`DetermineOutputPath()`** -- Determine the output path for a generated file based on feature path
- **`ValidateGherkinContent()`** -- Validate Gherkin feature file content for structural correctness
- **`ValidateGherkinLine()`** -- Validate individual Gherkin lines within parsing context

## Patterns

- Security-first path handling: all path operations validate against traversal attacks and reserved names
- Stateful line-by-line parsing: `gherkinValidationState` tracks context across Gherkin lines
- Pure validation functions: no side effects, return errors for invalid input

## Internal Structure

| File | Responsibility |
| --- | --- |
| security.go | Feature name extraction, path security validation, and output path determination |
| validation.go | Gherkin content and line validation with stateful parsing |

## Dependencies

None (standard library only).

## Role in System

The `specs` package provides the validation and security layer for Gherkin specification processing. It is used by spec generation commands to safely handle feature file paths and validate Gherkin content before writing output files.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
