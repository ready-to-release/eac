package validate

import "github.com/ready-to-release/eac/go/clibase/registry"

func init() {
	registry.RegisterAll(
		Validate,
		ValidateArtifacts,
		ValidateBooks,
		ValidateConfigCmd,
		ValidateContracts,
		ControlTags,
		ValidateDependencies,
		ValidateDesign,
		ValidateDocs,
		ValidateGoTidy,
		ValidateMarkdown,
		ValidateModuleFiles,
		ValidateModuleHierarchy,
		ValidateReleaseVersion,
		ValidateSpecs,
		TestTags,
		ValidateVersion,
	)
}
