package docs

// UpdateOptions holds parsed flags for the update docs command.
type UpdateOptions struct {
	// Areas specifies which documentation areas to update.
	// An empty set means all areas.
	Areas *AreaSet

	// DryRun shows what would be changed without actually changing.
	DryRun bool

	// Force re-renders/re-optimizes all assets even if cached.
	Force bool

	// Verbose shows detailed progress for each file.
	Verbose bool

	// PruneCache identifies and removes orphaned cache files (mermaid/drawio).
	PruneCache bool
}
