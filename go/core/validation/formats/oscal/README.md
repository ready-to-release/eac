# validation/formats/oscal

Validates OSCAL (Open Security Controls Assessment Language) documents for catalog and profile types.

Uses the `go-oscal` library for type-safe parsing and validates required fields, UUID presence, metadata completeness,
control imports, and NIST 800-53 control ID formats.

## Key Types

| Type               | Purpose                                                                                            |
| ------------------ | -------------------------------------------------------------------------------------------------- |
| `Validator`        | Generic OSCAL validator that delegates to type-specific validators based on `DocumentType`         |
| `DocumentType`     | Enum for OSCAL document types (`catalog`, `profile`)                                               |
| `CatalogValidator` | Validates OSCAL catalog documents (UUID, metadata, groups, controls)                               |
| `ProfileValidator` | Validates OSCAL profile documents (UUID, metadata, imports, control selections, control ID format) |

## Key Functions

| Function           | Purpose                                                        |
| ------------------ | -------------------------------------------------------------- |
| `NewValidator`     | Creates a validator for the specified OSCAL document type      |
| `IsValidControlID` | Validates NIST 800-53 control ID format (e.g., `ac-1`, `au-2`) |

## Patterns

- **Delegation pattern**: `Validator` delegates to `CatalogValidator` or `ProfileValidator` based on document type
- **go-oscal integration**: Uses `defenseunicorns/go-oscal` types for type-safe OSCAL parsing
- **Enhanced error reporting**: Uses `validation.ErrorFormatter` for structured errors with expected format examples and fix hints
- **Hierarchical validation**: Profile validation checks imports, then control selections within each import

## Internal Structure

| File           | Purpose                                                                                           |
| -------------- | ------------------------------------------------------------------------------------------------- |
| `validator.go` | `Validator` facade, `DocumentType` enum, `NewValidator` factory                                   |
| `catalog.go`   | `CatalogValidator` with UUID, metadata, groups, and controls validation                           |
| `profile.go`   | `ProfileValidator` with UUID, metadata, imports, control ID format validation, `IsValidControlID` |

## Dependencies

| Package           | Purpose                                          |
| ----------------- | ------------------------------------------------ |
| `core/validation` | `ValidationError`, error codes, `ErrorFormatter` |

## Role in System

Used by the AI generation pipeline (`ai/generation`) to validate OSCAL output from risk-related commands. Ensures generated OSCAL catalogs and profiles are structurally valid before being saved, with detailed error messages guiding AI retry attempts toward correct output.

## Code Health

### Tech Debt
- `profile.go` is 299 lines, approaching refactor threshold

### Pain Points
- None identified

### Optimization Opportunities
- None identified
