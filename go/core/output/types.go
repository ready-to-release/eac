package output

import (
	"sort"
	"time"

	"github.com/ready-to-release/eac/go/core/workunit"
)

// Status represents the execution status of a work unit.
type Status string

const (
	// StatusPending indicates the work unit is waiting to start.
	StatusPending Status = "pending"

	// StatusInProgress indicates the work unit is currently executing.
	StatusInProgress Status = "in_progress"

	// StatusCompleted indicates the work unit finished successfully.
	StatusCompleted Status = "completed"

	// StatusFailed indicates the work unit finished with errors.
	StatusFailed Status = "failed"

	// StatusCached indicates the work unit result was reused from cache.
	StatusCached Status = "cached"
)

// Artifact represents a single output artifact from a work unit.
type Artifact struct {
	// ID is a unique identifier for this artifact within the work unit.
	ID string `json:"id"`

	// Path is the relative path to the artifact file from workspace root.
	Path string `json:"path"`

	// SHA256 is the hash of the artifact content, prefixed with "sha256:".
	SHA256 string `json:"sha256"`

	// Size is the file size in bytes.
	Size int64 `json:"size"`

	// Type describes the artifact type (e.g., "binary", "report", "coverage").
	Type string `json:"type"`
}

// UoWManifest represents a Unit of Work manifest containing execution metadata.
// This is written to out/{context}/{module}/{dirname}/uow.manifest.json
// where dirname = component[-extra1][-extra2]... for uniqueness.
type UoWManifest struct {
	// Context is the operation type: build, test, lint, or scan.
	Context workunit.Context `json:"context"`

	// Module is the module moniker (e.g., "core").
	Module string `json:"module"`

	// Component is the component name (e.g., "go", "docker").
	Component string `json:"component"`

	// Tool is the handler/provider/scanner name (e.g., "go", "gotest", "trivy-vuln").
	Tool string `json:"tool"`

	// Extra contains context-specific discriminating fields (e.g., testset, category).
	// These are included in the directory name for uniqueness.
	Extra map[string]string `json:"extra,omitempty"`

	// ExitCode is the exit code of the tool execution (0 = success).
	ExitCode int `json:"exit_code"`

	// InputHash is the hash of all inputs that affect this work unit.
	InputHash string `json:"input_hash"`

	// ExecutedAt is the timestamp when execution started.
	ExecutedAt time.Time `json:"executed_at"`

	// Duration is how long the execution took.
	Duration time.Duration `json:"duration"`

	// Artifacts is the list of output artifacts produced.
	Artifacts []Artifact `json:"artifacts"`

	// OutputHash is a hash of all artifact hashes, representing total output.
	OutputHash string `json:"output_hash"`

	// Version is the manifest schema version.
	Version string `json:"version"`

	// NoOp indicates this UoW intentionally produced no output.
	// Used for placeholder UoWs like build:templates:none modules.
	// When true, empty InputHash/OutputHash/Artifacts are expected and valid.
	// Cache detection treats NoOp UoWs as always up-to-date if the manifest exists.
	NoOp bool `json:"noop,omitempty"`

	// Metadata contains optional context-specific key-value pairs.
	// Used to store testset, build_id, and other UoW-specific information.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Tags carries classified tag data for test UoWs.
	// Populated pre-execution from source-level TestReference data.
	Tags workunit.TagSummary `json:"tags,omitempty"`
}

// DirName returns the unique directory name for this manifest.
// Format: component-tool[-extraVal1][-extraVal2]...
// Tool and extra values are appended with dashes in sorted key order for uniqueness.
func (m *UoWManifest) DirName() string {
	dirName := m.Component
	if m.Tool != "" {
		dirName += "-" + m.Tool
	}

	// Append all Extra values with dashes in sorted key order
	if len(m.Extra) > 0 {
		keys := make([]string, 0, len(m.Extra))
		for k := range m.Extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if v := m.Extra[k]; v != "" {
				dirName += "-" + v
			}
		}
	}
	return dirName
}

// GetExtra returns context-specific discriminators (e.g., testset, category).
// Implements interfaces.UoWManifestPort.
func (m *UoWManifest) GetExtra() map[string]string {
	return m.Extra
}

// ValidationResult contains the outcome of validating a work unit's output.
type ValidationResult struct {
	// Valid is true if the manifest exists, is valid, and all artifacts are present and uncorrupted.
	Valid bool

	// ManifestExists is true if the manifest file exists on disk.
	ManifestExists bool

	// ManifestValid is true if the manifest file contains valid JSON.
	ManifestValid bool

	// ArtifactsValid is true if all artifacts exist and their hashes match.
	ArtifactsValid bool

	// MissingArtifacts lists paths of artifacts that don't exist on disk.
	MissingArtifacts []string

	// CorruptArtifacts lists paths of artifacts whose hashes don't match.
	CorruptArtifacts []string

	// Error contains any error encountered during validation.
	Error error
}

// ComponentView aggregates UoW results for a single component.
type ComponentView struct {
	// Module is the module moniker.
	Module string

	// Component is the component name.
	Component string

	// Status is the aggregated status of all UoWs.
	Status Status

	// UoWs is the list of work unit manifests for this component.
	UoWs []UoWManifest

	// TotalSize is the sum of all artifact sizes in bytes.
	TotalSize int64
}

// ModuleView aggregates component results for a single module.
type ModuleView struct {
	// Module is the module moniker.
	Module string

	// Status is the aggregated status of all components.
	Status Status

	// Components is the list of component views for this module.
	Components []ComponentView

	// TotalSize is the sum of all artifact sizes in bytes.
	TotalSize int64
}

// StatusFromExitCode converts an exit code to a Status.
// Convention:
//   - exitCode == 0: StatusCompleted (success)
//   - exitCode < 0:  StatusCached (cached/skipped)
//   - exitCode > 0:  StatusFailed (error)
func StatusFromExitCode(exitCode int) Status {
	switch {
	case exitCode == 0:
		return StatusCompleted
	case exitCode < 0:
		return StatusCached
	default:
		return StatusFailed
	}
}

// NewNoOpManifest creates a manifest for UoWs that intentionally produce no output.
// Used for placeholder UoWs like modules with no buildable components, no tests,
// no lintable code, or non-scannable components.
//
// NoOp manifests have:
//   - NoOp: true (always up-to-date in cache detection if ExitCode == 0)
//   - ExitCode: 0 (successful)
//   - Empty InputHash/OutputHash/Artifacts (expected and valid)
//   - Metadata["reason"] describing why it's NoOp
func NewNoOpManifest(ctx workunit.Context, module, component, tool, reason string) *UoWManifest {
	return &UoWManifest{
		Context:    ctx,
		Module:     module,
		Component:  component,
		Tool:       tool,
		ExitCode:   0,
		InputHash:  "",
		ExecutedAt: time.Now().UTC(),
		Duration:   0,
		Artifacts:  []Artifact{},
		OutputHash: "",
		Version:    "1.0.0",
		NoOp:       true,
		Metadata: map[string]string{
			"reason": reason,
		},
	}
}
