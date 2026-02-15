package release

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&releaseAwaitDepsCommand{},
		&releaseGenerateModuleCalverCommand{},
		&releaseChangelogCommand{},
		&releaseCheckCICommand{},
		&releaseCheckExistsCommand{},
		&releaseCheckPendingCommand{},
		&releaseCleanupCommand{},
		&releaseClieCommand{},
		&releaseEacExtCommand{},
		&releaseExecuteLayersCommand{},
		&releaseExtractVersionCommand{},
		&releaseGetVersionCommand{},
		&releasePendingCommand{},
		&releasePruneCommand{},
		&releasePrunePackagesCommand{},
		&releaseTagPendingCommand{},
		&releaseThisCommand{},
		&validateReleaseCommand{},
	}
}
