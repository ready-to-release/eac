package validate

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/commands/repository/internal/dependabot"
	"github.com/ready-to-release/eac/go/core/repository"
)

type validateDependabotCommand struct{}

var _ core.SimpleCommandPort = (*validateDependabotCommand)(nil)

func (c *validateDependabotCommand) Name() string { return "validate dependabot" }

func (c *validateDependabotCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "validate-dependabot",
		Short:         "Validate that dependabot.yml covers all dependency sources",
		Long:          "Validates that .github/dependabot.yml includes entries for all\ndependency sources in the repository.\n\nScans for:\n  - Go modules (from go.work)\n  - npm packages (package.json)\n  - Python packages (requirements.txt)\n  - Docker base images (Dockerfile)\n  - GitHub Actions workflows\n\nExpected Output:\n  Shows missing and extra entries. Exit code 0 if all covered, 1 if gaps found.\n\nExample:\n  validate dependabot",
	}
}

func (c *validateDependabotCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ValidateDependabot()
}

// ValidateDependabot checks dependabot.yml against actual dependency sources.
func ValidateDependabot() int {
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:]
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printDependabotUsage()
		return 0
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

	report := dependabot.Compare(config.Updates, discovered)

	printDependabotReport(report, len(config.Updates), len(discovered), len(report.Consolidated))

	if report.HasIssues() {
		return 1
	}
	return 0
}

func printDependabotReport(report *dependabot.ComparisonReport, declared, discovered, consolidated int) {
	var missingItems []string
	for _, m := range report.Missing {
		missingItems = append(missingItems,
			formatBulletWithDetail(
				fmt.Sprintf("%s %s", m.Ecosystem, m.Directory),
				fmt.Sprintf("Add a '%s' entry for directory '%s'", m.Ecosystem, m.Directory),
			))
	}

	var extraItems []string
	for _, e := range report.Extra {
		extraItems = append(extraItems,
			formatBullet(
				fmt.Sprintf("%s %s (no matching source found)", e.PackageEcosystem, e.Directory),
			))
	}

	printValidationReportWithSummary(
		validationReport{
			Title: "Dependabot Coverage Validation Report",
			Sections: []validationSection{
				{
					Icon:    "\u274c",
					Label:   "Missing dependabot entries:",
					Items:   missingItems,
					FixHint: "To fix, run: eac update dependabot",
				},
				{
					Icon:    "\u26a0\ufe0f",
					Label:   "Extra dependabot entries (no source found):",
					Items:   extraItems,
					FixHint: "To fix, run: eac update dependabot --prune",
				},
			},
			SuccessMessage: "All dependency sources are covered by dependabot!",
		},
		summaryLines(declared, discovered, consolidated, report),
	)
}

func summaryLines(declared, discovered, consolidated int, report *dependabot.ComparisonReport) []string {
	lines := []string{
		fmt.Sprintf("Declared entries in dependabot.yml: %d", declared),
		fmt.Sprintf("Discovered dependency sources: %d", discovered),
		fmt.Sprintf("Missing entries: %d", len(report.Missing)),
		fmt.Sprintf("Extra entries: %d", len(report.Extra)),
		fmt.Sprintf("Matched entries: %d", len(report.Matched)),
	}
	if consolidated > 0 {
		lines = append(lines, fmt.Sprintf("Consolidated entries (externally managed): %d", consolidated))
	}
	return lines
}

func printDependabotUsage() {
	log.Info("Validate that dependabot.yml covers all dependency sources")
	log.Info("")
	log.Info("Usage: clie validate dependabot")
	log.Info("")
	log.Info("Checks:")
	log.Info("  - All Go modules from go.work have gomod entries")
	log.Info("  - All package.json files have npm entries")
	log.Info("  - All requirements.txt files have pip entries")
	log.Info("  - All Dockerfiles have docker entries")
	log.Info("  - GitHub Actions has a github-actions entry")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Validate dependabot coverage")
	log.Info("  clie validate dependabot")
}
