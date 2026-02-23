package validate

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Commands returns all command ports provided by the validate package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&validateCommand{},
		&validateArtifactsCommand{},
		&validateBooksCommand{},
		&validateConfigCommand{},
		&validateContractsCommand{},
		&validateDependabotCommand{},
		&validateControlTagsCommand{},
		&validateDependenciesCommand{},
		&validateDesignCommand{},
		&validateDocsCommand{},
		&validateGoTidyCommand{},
		&validateMarkdownCommand{},
		&validateModuleFilesCommand{},
		&validateModuleHierarchyCommand{},
		&validateReleaseVersionCommand{},
		&validateSpecsCommand{},
		&validateTestTagsCommand{},
		&validateVersionCommand{},
	}
}
