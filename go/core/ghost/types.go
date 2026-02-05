// Package ghost provides ghost tracking functionality for dark launch code discovery.
// Ghosts are files, directories, or code constructs following a configurable naming
// convention (default: ghost-* prefix) that enable dark launching, monitoring probes,
// feature toggles, and quick discovery of hidden code.
package ghost

// GhostType represents the kind of ghost entity discovered.
type GhostType string

const (
	// GhostTypeFile indicates a ghost file (e.g., ghost-feature.go)
	GhostTypeFile GhostType = "file"

	// GhostTypeDirectory indicates a ghost directory (e.g., ghost-monitoring/)
	GhostTypeDirectory GhostType = "directory"

	// Future: GhostTypeFunction, GhostTypeVariable, GhostTypeStruct
)

// Ghost represents a discovered ghost entity in the repository.
type Ghost struct {
	// Path is the relative path from repository root (forward slashes)
	Path string `json:"path" yaml:"path"`

	// Name is the base name of the file/directory
	Name string `json:"name" yaml:"name"`

	// Type indicates whether this is a file or directory
	Type GhostType `json:"type" yaml:"type"`

	// Module is the owning module moniker (empty if unowned)
	Module string `json:"module,omitempty" yaml:"module,omitempty"`

	// GhostName is the name without the ghost prefix
	GhostName string `json:"ghost_name" yaml:"ghost_name"`
}

// GhostReport contains all discovered ghosts with summary statistics.
type GhostReport struct {
	// Ghosts is the list of all discovered ghost entities
	Ghosts []Ghost `json:"ghosts" yaml:"ghosts"`

	// Summary contains aggregate statistics
	Summary GhostSummary `json:"summary" yaml:"summary"`

	// Config shows the effective configuration used
	Config GhostConfigSummary `json:"config" yaml:"config"`
}

// GhostSummary contains aggregate statistics about ghosts.
type GhostSummary struct {
	// Total is the total number of ghosts found
	Total int `json:"total" yaml:"total"`

	// Files is the count of ghost files
	Files int `json:"files" yaml:"files"`

	// Directories is the count of ghost directories
	Directories int `json:"directories" yaml:"directories"`

	// ByModule maps module monikers to ghost counts
	ByModule map[string]int `json:"by_module" yaml:"by_module"`

	// Unowned is the count of ghosts not owned by any module
	Unowned int `json:"unowned" yaml:"unowned"`
}

// GhostConfigSummary shows the effective configuration.
type GhostConfigSummary struct {
	// Alias is the ghost prefix used (e.g., "ghost")
	Alias string `json:"alias" yaml:"alias"`

	// Patterns lists the patterns that match ghosts
	Patterns []string `json:"patterns" yaml:"patterns"`
}

// FilterOptions specifies criteria for filtering ghosts.
type FilterOptions struct {
	Type    string // Filter by type: "file" or "directory"
	Module  string // Filter to ghosts in specific module
	Unowned bool   // Only show unowned ghosts
}

// Filter applies filter options to a list of ghosts.
func Filter(ghosts []Ghost, opts FilterOptions) []Ghost {
	if opts.Type == "" && opts.Module == "" && !opts.Unowned {
		return ghosts
	}

	var result []Ghost
	for _, g := range ghosts {
		if opts.Type != "" && string(g.Type) != opts.Type {
			continue
		}
		if opts.Module != "" && g.Module != opts.Module {
			continue
		}
		if opts.Unowned && g.Module != "" {
			continue
		}
		result = append(result, g)
	}
	return result
}
