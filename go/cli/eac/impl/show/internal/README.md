# internal

Provides shared formatting utilities for the show command group, specifically artifact display formatting used by artifact-related show commands.

## Key Types

None (utility functions only).

## Key Functions

- **`FormatArtifactTable()`** -- Format a list of resolved artifacts as an aligned table for terminal display
- **`FormatMetadataOverrides()`** -- Format metadata override information for display
- **`isArtifactMetadata()`** -- Check whether a metadata key relates to artifact configuration
- **`FormatArtifactDetails()`** -- Format detailed artifact information including paths and verification status
- **`FormatArtifactSummaryHeader()`** -- Format the summary header line showing artifact counts

## Patterns

- Pure formatting functions: stateless utilities that transform data structures into display strings
- Shared across show commands: used by multiple show sub-commands that display artifact information

## Internal Structure

| File | Responsibility |
| --- | --- |
| artifact_formatter.go | Artifact table formatting, metadata override display, and summary header generation |

## Dependencies

- `cli/eac/impl/internal` -- `ResolvedArtifact` and `ArtifactResolutionSummary` types

## Role in System

The `show/internal` package provides display formatting utilities shared across artifact-related show commands. By centralizing artifact formatting logic here, it prevents duplication across `show artifacts`, `show build-summary`, and other commands that display artifact information.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
