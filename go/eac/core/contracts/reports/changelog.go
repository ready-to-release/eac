package reports

import (
	"fmt"

	"github.com/ready-to-release/eac/go/eac/core/changelog"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// ChangelogReport represents a parsed changelog with module context.
type ChangelogReport struct {
	Module    string
	Path      string
	Changelog *changelog.Changelog
}

// GetChangelog loads and parses a module's CHANGELOG.md file.
func GetChangelog(workspaceRoot, module string) (*ChangelogReport, error) {
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
	changelogPath := paths.ChangelogPath(workspaceRoot, module)

	// Parse changelog
	log, err := changelog.Parse(changelogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse changelog: %w", err)
	}

	log.Module = module

	return &ChangelogReport{
		Module:    module,
		Path:      changelogPath,
		Changelog: log,
	}, nil
}
