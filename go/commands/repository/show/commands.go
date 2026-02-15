package show

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Commands returns all show command implementations.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&showParent{},
		&showApprovalCommentsCommand{},
		&showApproveSummaryCommand{},
		&showArtifactsCommand{},
		&showBooksCommand{},
		&showBuildSummaryCommand{},
		&showBuildTimesCommand{},
		&showCIResultsCommand{},
		&showCISummaryCommand{},
		&showChangelogCommand{},
		&showComponentTypesCommand{},
		&showComponentsCommand{},
		&showConfigCommand{},
		&showDependenciesCommand{},
		&showDependencyCISummaryCommand{},
		&showDepsSetupSummaryCommand{},
		&showEnvironmentsCommand{},
		&showFilesCommand{},
		&showFilesChangedCommand{},
		&showFilesStagedCommand{},
		&showGhostsCommand{},
		&showLintSummaryCommand{},
		&showModulesCommand{},
		&showReleaseNotesCommand{},
		&showReleaseSummaryCommand{},
		&showScanSummaryCommand{},
		&showSpecsCommand{},
		&showSuiteCommand{},
		&showTestResultsCommand{},
		&showTestSummaryCommand{},
		&showTestTimingsCommand{},
		&showTestsCommand{},
		&showTriggerSummaryCommand{},
		&showUnitsCommand{},
		&showValidCommandsCommand{},
	}
}
