package release

import "github.com/ready-to-release/eac/go/clibase/registry"

func init() {
	registry.RegisterAll(
		ReleaseAwaitDeps,
		ReleaseCalver,
		ReleaseChangelog,
		ReleaseCheckCI,
		ReleaseCheckExists,
		ReleaseCheckPending,
		ReleaseCleanup,
		ReleaseSrcCli,
		ReleaseExtEac,
		ReleaseExecuteLayers,
		ExtractVersion,
		ReleaseGetVersion,
		ReleasePending,
		ReleasePrune,
		ReleasePrunePackages,
		ReleaseTagPending,
		ReleaseThis,
		ReleaseValidate,
	)
}
