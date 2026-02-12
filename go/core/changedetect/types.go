package changedetect

import (
	"context"
	"time"
)

// GitStateProvider abstracts git operations for testing and flexibility.
type GitStateProvider interface {
	// HeadCommit returns the full SHA of HEAD.
	HeadCommit(ctx context.Context) (string, error)

	// UncommittedFiles returns paths of files with uncommitted changes.
	// Uses git status --porcelain format parsing.
	UncommittedFiles(ctx context.Context) ([]string, error)
}

// FileHasher abstracts file content hashing for testing.
type FileHasher interface {
	// HashFiles computes a deterministic hash of file contents.
	// Files are sorted, and both path and content are included.
	HashFiles(ctx context.Context, workspaceRoot string, files []string) (string, error)

	// HashUncommittedState computes a hash representing dirty working tree state.
	HashUncommittedState(ctx context.Context, workspaceRoot string, files []string) (string, error)
}

// ModuleFileResolver provides file lists for modules.
type ModuleFileResolver interface {
	// GetModuleFiles expands glob patterns and returns files for a module.
	GetModuleFiles(ctx context.Context, workspaceRoot string, moniker string) ([]string, error)
}

// WorkspaceState represents the recorded state at a point in time.
type WorkspaceState struct {
	// Git commit SHA at time of recording
	Commit string `json:"commit"`

	// Hash of uncommitted changes (empty if working tree clean)
	UncommittedHash string `json:"uncommitted_hash,omitempty"`

	// Per-module state
	Modules map[string]ModuleState `json:"modules"`

	// When this state was recorded
	RecordedAt time.Time `json:"recorded_at"`
}

// Clone creates a deep copy of the workspace state.
func (s *WorkspaceState) Clone() *WorkspaceState {
	clone := &WorkspaceState{
		Commit:          s.Commit,
		UncommittedHash: s.UncommittedHash,
		Modules:         make(map[string]ModuleState, len(s.Modules)),
		RecordedAt:      s.RecordedAt,
	}
	for k, v := range s.Modules {
		// Clone the Extra map
		var extraClone map[string]interface{}
		if v.Extra != nil {
			extraClone = make(map[string]interface{}, len(v.Extra))
			for ek, ev := range v.Extra {
				extraClone[ek] = ev
			}
		}
		clone.Modules[k] = ModuleState{
			SourceHash: v.SourceHash,
			Extra:      extraClone,
		}
	}
	return clone
}

// ModuleState represents recorded state for a single module.
type ModuleState struct {
	// Hash of source files
	SourceHash string `json:"source_hash"`

	// Operation-specific fields (stored but not interpreted by changedetect)
	// Build: BuiltAt timestamp
	// Lint: Passed bool, LintedAt timestamp
	// Test: TestHash, BuildID, Passed bool, Dependencies
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// ChangeResult represents detected changes.
type ChangeResult struct {
	// Modules that changed
	Changed []string

	// Modules that are up-to-date
	UpToDate []string

	// Reason for each changed module
	Reasons map[string]string

	// True if no prior state exists
	Fresh bool

	// Detection duration
	Duration time.Duration
}

// DetectOptions configures change detection behavior.
type DetectOptions struct {
	// WorkspaceRoot is the repository root directory
	WorkspaceRoot string

	// PreviousState is the recorded state to compare against (nil = fresh)
	PreviousState *WorkspaceState

	// Modules to check (monikers)
	Modules []string

	// CheckDependencies enables transitive dependency checking (for tests)
	CheckDependencies bool

	// DependencyGraph maps module -> dependencies (for tests)
	DependencyGraph map[string][]string
}

// Detector detects module changes against recorded state.
type Detector struct {
	git    GitStateProvider
	hasher FileHasher
	files  ModuleFileResolver
}

// NewDetector creates a change detector with the given providers.
func NewDetector(git GitStateProvider, hasher FileHasher, files ModuleFileResolver) *Detector {
	return &Detector{
		git:    git,
		hasher: hasher,
		files:  files,
	}
}
