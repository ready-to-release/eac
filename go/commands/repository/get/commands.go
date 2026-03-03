package get

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Commands returns all get command implementations.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&getCommand{},
		&getApprovalCommentsCommand{},
		&getArtifactsCommand{},
		&getBinarySizesCommand{},
		&getBookDescriptionCommand{},
		&getBuildDepsCommand{},
		&getBuildTimesCommand{},
		&getChangedContainersCommand{},
		&getChangedModulesCommand{},
		&getChangedModulesCICommand{},
		&getChangedModulesLocalCommand{},
		&getChangelogCommand{},
		&getCIConfigCommand{},
		&getCIDispatchCommand{},
		&getCIResultsCommand{},
		&getCIWorkflowsCommand{},
		&getCLIReleaseNotesCommand{},
		&getComponentsCommand{},
		&getConfigCommand{},
		&getCurrentSHACommand{},
		&getDependenciesModuleCommand{},
		&getDocumentedCommandsCommand{},
		&getEnvironmentsCommand{},
		&getEvidenceCIRunsCommand{},
		&getFilesCommand{},
		&getFilesByModuleCommand{},
		&getGhostsCommand{},
		&getModuleCIWorkflowCommand{},
		&getModulesCommand{},
		&getModuleTriggerReasonCommand{},
		&getReleaseBundleCommand{},
		&getReleaseConfigCommand{},
		&getReleaseNotesCommand{},
		&getReleaseStatusCommand{},
		&getSpecsCommand{},
		&getSuiteCommand{},
		&getTestResultsCommand{},
		&getTestsCommand{},
		&getTestTimingsCommand{},
		&getTokenSizeCommand{},
		&getUnitsCommand{},
		&getValidCommandsCommand{},
	}
}
