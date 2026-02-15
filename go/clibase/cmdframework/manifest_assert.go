package cmdframework

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// AssertManifestsExist verifies that all executed UoWs have valid manifests.
// Units that failed, were skipped (dependency failed), cached, or timed out are excluded
// since they never ran (or were killed) and therefore never generated manifests.
//
// commandName is used for log prefixes and error messages (e.g., "build", "test").
// specs is the full set of expected UoW specs for this command.
func AssertManifestsExist(ctx *ExecutionContext, commandName string, specs []workunit.UnitSpec) error {
	if len(specs) == 0 {
		return nil
	}

	// Build set of UoW longnames that were not successfully executed.
	// ExitCode != 0 means:
	// - ExitCode > 0: failed
	// - ExitCode < 0: skipped/cached (e.g., -1 for cached)
	// Uses full Longname() to distinguish between UoWs that share the same
	// module/component/tool but differ by Extra fields (e.g., testname).
	skipManifestCheck := make(map[string]bool)
	for _, r := range ctx.UnitResults {
		if r.ExitCode != 0 {
			skipManifestCheck[r.Longname] = true
		}
	}

	prefix := strings.ToUpper(commandName)
	reader := output.NewReader(ctx.WorkspaceRoot)
	var missing []string
	checked := 0

	for _, spec := range specs {
		longname := spec.ID.Longname()
		if skipManifestCheck[longname] {
			log.Debugf("[%s-ASSERT] Skipping manifest check for failed/cached/skipped UoW: %s", prefix, longname)
			continue
		}

		checked++
		if _, err := reader.GetUoW(spec.ID); err != nil {
			missing = append(missing, longname)
			log.Debugf("[%s-ASSERT] Missing manifest for UoW: %s (error: %v)", prefix, longname, err)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%s completed but %d UoW manifest(s) are missing: %v\nThis indicates a bug in manifest generation - each UoW must persist its manifest",
			commandName, len(missing), missing)
	}

	log.Debugf("[%s-ASSERT] All %d UoW manifests verified (%d skipped due to non-zero exit codes)", prefix, checked, len(specs)-checked)
	return nil
}
