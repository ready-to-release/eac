package tests

// SharedContext contains state that needs to be shared across test modules
// This is defined here for external packages (src/commands/tests) to use
type SharedContext struct {
	RepoRoot      string
	CurrentBranch string
	ExitCode      int
	Output        string
}
