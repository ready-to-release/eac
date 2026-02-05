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

	// Build set of (module, component, handler) tuples that were not successfully executed.
	// ExitCode != 0 means:
	// - ExitCode > 0: failed
	// - ExitCode < 0: skipped/cached (e.g., -1 for cached)
	type unitKey struct{ module, component, handler string }
	skipManifestCheck := make(map[unitKey]bool)
	for _, r := range ctx.UnitResults {
		if r.ExitCode != 0 {
			skipManifestCheck[unitKey{r.Module, r.Component, r.Handler}] = true
		}
	}

	prefix := strings.ToUpper(commandName)
	reader := output.NewReader(ctx.WorkspaceRoot)
	var missing []string
	checked := 0

	for _, spec := range specs {
		key := unitKey{spec.ID.Module, spec.ID.Component, spec.ID.Tool}
		if skipManifestCheck[key] {
			log.Debugf("[%s-ASSERT] Skipping manifest check for failed/cached/skipped UoW: %s", prefix, spec.ID.Longname())
			continue
		}

		checked++
		if _, err := reader.GetUoW(spec.ID); err != nil {
			missing = append(missing, spec.ID.Longname())
			log.Debugf("[%s-ASSERT] Missing manifest for UoW: %s (error: %v)", prefix, spec.ID.Longname(), err)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%s completed but %d UoW manifest(s) are missing: %v\nThis indicates a bug in manifest generation - each UoW must persist its manifest",
			commandName, len(missing), missing)
	}

	log.Debugf("[%s-ASSERT] All %d UoW manifests verified (%d skipped due to non-zero exit codes)", prefix, checked, len(specs)-checked)
	return nil
}
