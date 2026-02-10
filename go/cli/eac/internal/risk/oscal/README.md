# oscal

Provides OSCAL document loading, writing, and manipulation utilities. Handles OSCAL profiles, catalogs, and assessment results using go-oscal types with file locking for concurrent access.

## Key Types

- **Type aliases from go-oscal**: `AssessmentResults`, `Result`, `Observation`, `Finding`, `Prop`, `RelevantEvidence`, `ReviewedControls`, `ControlSelection`, `ControlRef`, `Target`, `RelatedObservation`

## Key Functions

- **`LoadProfile()`** -- Load an OSCAL profile from a JSON file using go-oscal types
- **`LoadAssessmentResults()`** -- Load OSCAL assessment results from a JSON file
- **`LoadCatalog()`** -- Load an OSCAL catalog from a URL or local file path (supports NIST 800-53)
- **`WriteProfile()`** -- Write an OSCAL profile to a JSON file with metadata timestamp update
- **`WriteAssessmentResults()`** -- Write OSCAL assessment results with file locking for concurrent access
- **`GetProfileControlIDs()`** -- Extract all control IDs from a profile's import declarations
- **`GetControlIDsFromProfile()`** -- Extract control IDs from profile imports (alternative implementation)
- **`ProfileHasControl()`** -- Check if a profile includes a specific control ID
- **`NewProfileDocument()`** -- Create a new OSCAL profile document with required metadata
- **`UpdateProfileMetadata()`** -- Update profile metadata timestamps

## Patterns

- go-oscal type aliases: provides convenient aliases for verbose go-oscal types
- File locking: uses `gofrs/flock` for concurrent write access to OSCAL documents
- URL and file loading: `LoadCatalog()` transparently handles both HTTP URLs and local file paths
- JSON envelope wrapping: wraps profile/catalog in `OscalModels` for correct OSCAL JSON structure
- NIST constants: defines standard OSCAL finding states and observation methods

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | Type aliases for go-oscal types and target type constants |
| constants.go | NIST catalog URL, OSCAL finding states, and observation method constants |
| loader.go | Profile, assessment results, and catalog loading from files or URLs |
| writer.go | Profile and assessment results writing with file locking |
| oscal_helpers.go | Profile control ID extraction and new document creation |
| profile_helpers.go | Profile control ID extraction and control membership checking |
| catalog_helpers.go | OSCAL catalog loading from URLs and local files |

## Dependencies

- `github.com/defenseunicorns/go-oscal` -- OSCAL 1.1.3 type definitions
- `github.com/gofrs/flock` -- file locking for concurrent write access
- `github.com/google/uuid` -- UUID generation for OSCAL document identifiers
- `core/config` -- configuration for evidence paths
- `core/logging` -- structured logging
- `core/paths` -- OSCAL file path resolution

## Role in System

The `oscal` package provides the OSCAL document manipulation layer for the risk assessment pipeline. It handles reading and writing OSCAL profiles, catalogs, and assessment results -- the standard format for communicating security control information in compliance workflows.

## Code Health

### Tech Debt
- `GetProfileControlIDs()` in `oscal_helpers.go` and `GetControlIDsFromProfile()` in `profile_helpers.go` are near-identical functions extracting control IDs from profiles

### Pain Points
- None identified.

### Optimization Opportunities
- Consolidate the duplicate control ID extraction functions into a single implementation (low effort)
