package updatedependabot

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/internal/dependabot"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

type updateDependabotCommand struct{}

var _ core.SimpleCommandPort = (*updateDependabotCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&updateDependabotCommand{},
	}
}

func (c *updateDependabotCommand) Name() string { return "update dependabot" }

func (c *updateDependabotCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "update-dependabot",
		Short:         "Generate or update .github/dependabot.yml for all dependency sources",
		Long:          "Scans the repository for all dependency sources and ensures\n.github/dependabot.yml is up to date.\n\nBehavior:\n  - Preserves existing entries and their customizations (schedule, labels, etc.)\n  - Adds new entries for discovered but uncovered dependency sources\n  - With --prune, removes entries with no matching source\n\nThis is the fix counterpart to 'validate dependabot' (which only checks).\n\nExamples:\n  update dependabot              Add missing entries\n  update dependabot --prune      Also remove stale entries\n  update dependabot --dry-run    Show what would change",
		Flags: []core.FlagSpec{
			{Name: "dry-run", Type: "bool", DefaultValue: "false",
				Usage: "Show what would change without modifying the file"},
			{Name: "prune", Type: "bool", DefaultValue: "false",
				Usage: "Remove entries that have no matching dependency source"},
			{Name: "verbose", Shorthand: "v", Type: "bool", DefaultValue: "false",
				Usage: "Show detailed discovery output"},
		},
	}
}

func (c *updateDependabotCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return UpdateDependabot()
}

var log = logging.C()

// UpdateDependabot scans the repository and updates dependabot.yml.
func UpdateDependabot() int {
	args := os.Args[3:]

	dryRun := false
	prune := false
	verbose := false

	for _, arg := range args {
		switch arg {
		case "--dry-run":
			dryRun = true
		case "--prune":
			prune = true
		case "-v", "--verbose":
			verbose = true
		case "-h", "--help":
			return 0
		}
	}

	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to get repository root: %v", err)
		return 1
	}

	config, err := dependabot.Parse(repoRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	discovered, err := dependabot.DiscoverAll(repoRoot, dependabot.DefaultDiscoverOptions())
	if err != nil {
		log.Errorf("Error discovering dependency sources: %v", err)
		return 1
	}

	if verbose {
		fmt.Printf("Discovered %d dependency sources\n", len(discovered))
		for _, e := range discovered {
			fmt.Printf("  %s %s\n", e.Ecosystem, e.Directory)
		}
		fmt.Println()
	}

	report := dependabot.Compare(config.Updates, discovered)

	if !report.HasIssues() {
		fmt.Println("dependabot.yml is already up to date")
		return 0
	}

	// Build new updates list
	existingByKey := make(map[string]dependabot.UpdateEntry)
	for _, u := range config.Updates {
		existingByKey[u.Key()] = u
	}

	var newUpdates []dependabot.UpdateEntry

	// Keep consolidated entries unchanged (externally managed)
	for _, u := range config.Updates {
		if u.IsConsolidated() {
			newUpdates = append(newUpdates, u)
		}
	}

	// Keep matched entries (preserves user customizations)
	for _, m := range report.Matched {
		if existing, ok := existingByKey[m.Key()]; ok {
			newUpdates = append(newUpdates, existing)
		}
	}

	// Keep extra entries unless pruning
	if !prune {
		for _, e := range report.Extra {
			newUpdates = append(newUpdates, e)
		}
	}

	// Add missing entries with defaults
	added := 0
	for _, m := range report.Missing {
		newUpdates = append(newUpdates, dependabot.DefaultUpdateEntry(m))
		added++
	}

	pruned := 0
	if prune {
		pruned = len(report.Extra)
	}

	dependabot.SortUpdates(newUpdates)
	config.Updates = newUpdates

	if dryRun {
		fmt.Printf("Would add %d entries, prune %d entries\n", added, pruned)
		if added > 0 {
			fmt.Println("\nNew entries:")
			for _, m := range report.Missing {
				fmt.Printf("  + %s %s\n", m.Ecosystem, m.Directory)
			}
		}
		if prune && pruned > 0 {
			fmt.Println("\nPruned entries:")
			for _, e := range report.Extra {
				fmt.Printf("  - %s %s\n", e.PackageEcosystem, e.Directory)
			}
		}
		return 0
	}

	if err := dependabot.Write(repoRoot, config); err != nil {
		log.Errorf("Error writing dependabot.yml: %v", err)
		return 1
	}

	fmt.Printf("Updated dependabot.yml: added %d, pruned %d, total %d entries\n",
		added, pruned, len(newUpdates))

	return 0
}
