package interfaces

import "time"

// OutputReaderPort provides read access to execution outputs.
// It aggregates UoW (Unit of Work) manifests into higher-level views.
// Implementations compute aggregations on-the-fly from UoW manifest files.
type OutputReaderPort interface {
	// GetUoW loads a single UoW manifest from disk.
	// context is the operation type: "build", "test", "lint", "scan"
	// Returns nil and error if the manifest doesn't exist or is invalid.
	GetUoW(context, module, component, tool string) (UoWManifestPort, error)

	// GetModule computes a module view by aggregating all UoWs for a module.
	// Returns the aggregated status and list of components.
	GetModule(context, module string) (ModuleViewPort, error)

	// ListUoWs returns all UoW manifests for a module in a given context.
	ListUoWs(context, module string) ([]UoWManifestPort, error)

	// ValidateUoW checks if a UoW's output is valid (manifest exists, artifacts present).
	ValidateUoW(context, module, component, tool string) ValidationResultPort

	// ValidateModule checks if all expected UoWs for a module are valid.
	// expectedUoWs contains the list of unit IDs that should exist.
	ValidateModule(context, module string, expectedUoWs []UnitIDPort) ValidationResultPort
}

// UoWTrackerPort provides write access to execution outputs.
// It tracks UoW execution and persists manifests.
type UoWTrackerPort interface {
	// RecordStart marks a UoW as started and creates its output directory.
	// This should be called before executing the work unit.
	RecordStart(unitID UnitIDPort) error

	// RecordComplete records completion and persists the UoW manifest.
	// If RecordStart was called, duration is computed from start time.
	RecordComplete(unitID UnitIDPort, manifest UoWManifestPort) error

	// RecordCacheHit validates and returns an existing UoW manifest.
	// Returns error if manifest is missing, invalid, or artifacts are corrupt.
	RecordCacheHit(unitID UnitIDPort) (UoWManifestPort, error)
}

// UoWManifestPort represents a Unit of Work manifest containing execution metadata.
// Path format: out/{context}/{module}/{component}[-extra1][-extra2]/uow.manifest.json
type UoWManifestPort interface {
	// GetContext returns the operation type: "build", "test", "lint", "scan"
	GetContext() string

	// GetModule returns the module moniker.
	GetModule() string

	// GetComponent returns the component name.
	GetComponent() string

	// GetTool returns the tool/handler name.
	GetTool() string

	// GetExitCode returns the exit code (0=success, >0=failure, <0=cached).
	GetExitCode() int

	// GetInputHash returns the hash of all inputs.
	GetInputHash() string

	// GetExecutedAt returns when execution started.
	GetExecutedAt() time.Time

	// GetDuration returns how long execution took.
	GetDuration() time.Duration

	// GetArtifacts returns the list of output artifacts.
	GetArtifacts() []OutputArtifactPort

	// GetOutputHash returns the hash of all artifact hashes.
	GetOutputHash() string

	// GetExtra returns context-specific discriminators (e.g., testset, category).
	GetExtra() map[string]string
}

// OutputArtifactPort represents a single output artifact from a work unit.
// Named to distinguish from ArtifactPort in config.go which represents build artifact specs.
type OutputArtifactPort interface {
	// GetID returns the unique identifier for this artifact.
	GetID() string

	// GetPath returns the relative path to the artifact file.
	GetPath() string

	// GetSHA256 returns the hash of the artifact content.
	GetSHA256() string

	// GetSize returns the file size in bytes.
	GetSize() int64

	// GetType returns the artifact type (e.g., "binary", "report").
	GetType() string
}

// ModuleViewPort aggregates component results for a single module.
type ModuleViewPort interface {
	// GetModule returns the module moniker.
	GetModule() string

	// GetStatus returns the aggregated status: "pending", "in_progress", "completed", "failed", "cached"
	GetStatus() string

	// GetComponents returns the list of component views.
	GetComponents() []ComponentViewPort

	// GetTotalSize returns the sum of all artifact sizes in bytes.
	GetTotalSize() int64
}

// ComponentViewPort aggregates UoW results for a single component.
type ComponentViewPort interface {
	// GetModule returns the module moniker.
	GetModule() string

	// GetComponent returns the component name.
	GetComponent() string

	// GetStatus returns the aggregated status.
	GetStatus() string

	// GetUoWs returns the list of work unit manifests.
	GetUoWs() []UoWManifestPort

	// GetTotalSize returns the sum of all artifact sizes in bytes.
	GetTotalSize() int64
}

// ValidationResultPort contains the outcome of validating a work unit's output.
type ValidationResultPort interface {
	// IsValid returns true if everything is valid.
	IsValid() bool

	// HasManifest returns true if the manifest file exists.
	HasManifest() bool

	// IsManifestValid returns true if the manifest is valid JSON.
	IsManifestValid() bool

	// AreArtifactsValid returns true if all artifacts exist and hashes match.
	AreArtifactsValid() bool

	// GetMissingArtifacts returns paths of artifacts that don't exist.
	GetMissingArtifacts() []string

	// GetCorruptArtifacts returns paths of artifacts with wrong hashes.
	GetCorruptArtifacts() []string

	// GetError returns any error encountered during validation.
	GetError() error
}
