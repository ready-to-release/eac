package core

// WorkspacePort provides workspace detection and configuration.
type WorkspacePort interface {
	// Root returns the workspace root path.
	Root() string

	// Source indicates how the workspace was detected.
	// Values: "env:R2R_REPO_ROOT", "env:R2R_DOCKER_MODE", "git"
	Source() string

	// IsContainer returns true when running inside a container.
	IsContainer() bool

	// DistRoot returns the distribution root for tool assets.
	DistRoot() string
}

// WorkspaceDetectorPort detects workspaces.
type WorkspaceDetectorPort interface {
	// Detect finds the workspace root using the full detection chain.
	Detect() (WorkspacePort, error)

	// DetectWithOptions finds the workspace root with custom options.
	DetectWithOptions(opts WorkspaceDetectOptions) (WorkspacePort, error)
}

// WorkspaceDetectOptions configures workspace detection.
type WorkspaceDetectOptions struct {
	Mode       WorkspaceDetectMode
	StartPath  string
	Validate   bool
	RequireGit bool
}

// WorkspaceDetectMode controls how workspace detection behaves.
type WorkspaceDetectMode int

const (
	// WorkspaceModeAuto tries env vars first, then git detection.
	WorkspaceModeAuto WorkspaceDetectMode = iota
	// WorkspaceModeExplicit requires R2R_REPO_ROOT to be set.
	WorkspaceModeExplicit
	// WorkspaceModeGitOnly skips env vars and uses only git detection.
	WorkspaceModeGitOnly
)

// RepositoryPort provides repository operations.
type RepositoryPort interface {
	// Root returns the repository root path.
	Root() string

	// GitRepo returns the git repository interface.
	GitRepo() GitRepositoryPort

	// Files returns all files in the repository.
	Files(trackedOnly, includeIgnored bool) ([]FileInfo, error)
}

// GitRepositoryPort provides git operations.
type GitRepositoryPort interface {
	// RootPath returns the repository root.
	RootPath() string

	// TrackedFiles returns all tracked files.
	TrackedFiles() ([]string, error)

	// StagedFiles returns all staged files.
	StagedFiles() ([]string, error)

	// IsFileTracked checks if a file is tracked by git.
	IsFileTracked(relPath string) bool

	// IsFileIgnored checks if a file is ignored by .gitignore.
	IsFileIgnored(relPath string) bool

	// CurrentBranch returns the current branch name.
	CurrentBranch() (string, error)

	// CurrentCommit returns the current commit SHA.
	CurrentCommit() (string, error)
}

// FileInfo represents information about a repository file.
type FileInfo struct {
	Path         string // Relative path from repository root
	AbsolutePath string // Absolute filesystem path
	IsTracked    bool   // Whether the file is tracked by git
	IsIgnored    bool   // Whether the file is ignored by .gitignore
}

// ExecutionPlanPort provides execution planning.
type ExecutionPlanPort interface {
	// All returns all monikers in execution order.
	All() []string
}
