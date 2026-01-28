// Command: update lint
// Short: [DEPRECATED] Lint one or more modules with auto-fix
// Long: This command is deprecated. Use 'lint --fix' instead.
// Long:
// Long: The 'update lint' command is now an alias for 'lint --fix'.
// Long: All arguments are passed through to the lint command.
// Args: modules
package lint

import (
	"os"

	reglint "github.com/ready-to-release/eac/go/eac/commands/impl/lint"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(UpdateLint)
}

// UpdateLint command entry point - calls regular lint with --fix.
// Deprecated: Use the top-level "lint --fix" command instead.
func UpdateLint() int {
	// Show deprecation warning
	log.Warnf("⚠️  'update lint' is deprecated. Use 'lint --fix' instead.")

	// Check for help flag
	args := os.Args[3:] // Skip program name, "update", and "lint"
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		log.Info("DEPRECATED: Use 'lint --fix' instead.")
		log.Info("")
		log.Info("Usage: lint --fix [flags] [module1] [module2] ...")
		log.Info("")
		log.Info("Run 'lint --help' for full options.")
		return 0
	}

	// Inject --fix into arguments and call regular lint
	// Reconstruct os.Args for lint command: [program, "lint", "--fix", ...rest]
	newArgs := []string{os.Args[0], "lint", "--fix"}
	newArgs = append(newArgs, os.Args[3:]...) // Skip "update" and "lint"
	os.Args = newArgs

	return reglint.Lint()
}
