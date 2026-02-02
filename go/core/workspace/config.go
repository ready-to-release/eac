package workspace

// Mode controls workspace detection behavior.
type Mode int

const (
	// ModeAuto uses the full detection chain (env -> docker -> git).
	ModeAuto Mode = iota

	// ModeExplicit requires R2R_REPO_ROOT to be set, fails otherwise.
	// Use for tests that require isolation.
	ModeExplicit

	// ModeGitOnly skips env var checks, only uses git detection.
	// Use when you need the "real" repo root regardless of env.
	ModeGitOnly
)

// Options configures workspace detection behavior.
type Options struct {
	// Mode controls detection strategy (default: ModeAuto).
	Mode Mode

	// StartPath is the directory to start git detection from.
	// Empty means use current working directory.
	StartPath string

	// Validate checks that the detected path is a valid workspace.
	// When true, verifies .git or .eac/repository.yml exists.
	// Default: true
	Validate bool

	// RequireGit requires .git directory to exist (not just .eac/repository.yml).
	// Default: false
	RequireGit bool

	// Logger for debug output (optional).
	Logger interface{ Debug(msg string, keysAndValues ...any) }
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Mode:     ModeAuto,
		Validate: true,
	}
}
