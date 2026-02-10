package show

import "github.com/ready-to-release/eac/go/clibase/registry"

func init() {
	registry.RegisterAll(
		Show,
		ShowApprovalComments,
		ShowApproveSummary,
		ShowArtifacts,
		ShowBooks,
		ShowBuildSummary,
		ShowBuildTimes,
		ShowCIResults,
		ShowCISummary,
		ShowChangelog,
		ShowComponentTypes,
		ShowComponents,
		ShowConfig,
		ShowDependencies,
		ShowDependencyCISummary,
		ShowDepsSetupSummary,
		ShowEnvironments,
		ShowFiles,
		ShowFilesChanged,
		ShowFilesStaged,
		ShowGhosts,
		ShowLintSummary,
		ShowModules,
		ShowReleaseNotes,
		ShowReleaseSummary,
		ShowScanSummary,
		ShowSpecs,
		ShowSuite,
		ShowTestResults,
		ShowTestSummary,
		ShowTestTimings,
		ShowTests,
		ShowTriggerSummary,
		ShowUnits,
		ShowValidCommands,
	)
}
