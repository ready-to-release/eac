package reports

import (
	"fmt"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/releasenotes"
)

// ReleaseNotesReport represents parsed release notes with module context.
type ReleaseNotesReport struct {
	Module       string
	Path         string
	ReleaseNotes *releasenotes.ReleaseNotes
}

// GetReleaseNotes loads and parses a module's RELEASE-NOTES.md file.
func GetReleaseNotes(workspaceRoot, module string) (*ReleaseNotesReport, error) {
	// Load config
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Get module contract to validate it exists
	_, ok := cfg.Repository.GetModule(module)
	if !ok {
		return nil, fmt.Errorf("module not found: %s", module)
	}

	// Use paths package for proper path construction
	releaseNotesPath := paths.ReleaseNotesPath(workspaceRoot, module)

	// Parse release notes
	notes, err := releasenotes.Parse(releaseNotesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse release notes: %w", err)
	}

	return &ReleaseNotesReport{
		Module:       module,
		Path:         releaseNotesPath,
		ReleaseNotes: notes,
	}, nil
}
