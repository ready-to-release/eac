package ghost

import (
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// BuildReport scans for ghosts and builds a complete report.
// Uses the provided FileSource - does not create its own.
// The alias parameter specifies the ghost prefix (e.g., "ghost").
func BuildReport(source FileSource, registry *modules.Registry, alias string) (*GhostReport, error) {
	if alias == "" {
		alias = "ghost"
	}

	// Build scan options from config
	opts := ScanOptions{
		Alias: alias,
	}

	// Scan for ghosts using FileSource
	scanner := NewScanner(source, registry, opts)
	ghosts, err := scanner.Scan()
	if err != nil {
		return nil, err
	}

	// Build summary
	summary := BuildSummary(ghosts)

	return &GhostReport{
		Ghosts:  ghosts,
		Summary: summary,
		Config: GhostConfigSummary{
			Alias:    opts.Alias,
			Patterns: []string{opts.Alias + "-*", opts.Alias + ".*", opts.Alias},
		},
	}, nil
}

// BuildSummary calculates summary statistics from a list of ghosts.
// This is exported so commands can rebuild summaries after filtering.
func BuildSummary(ghosts []Ghost) GhostSummary {
	summary := GhostSummary{
		Total:    len(ghosts),
		ByModule: make(map[string]int),
	}

	for _, g := range ghosts {
		switch g.Type {
		case GhostTypeFile:
			summary.Files++
		case GhostTypeDirectory:
			summary.Directories++
		}

		if g.Module == "" {
			summary.Unowned++
		} else {
			summary.ByModule[g.Module]++
		}
	}

	return summary
}
