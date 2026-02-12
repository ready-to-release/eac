package commitmessage

import (
	"fmt"
	"os"

	commitmessageinternal "github.com/ready-to-release/eac/go/cli/eac/impl/create/commit-message/internal"
	"github.com/ready-to-release/eac/go/cli/eac/impl/create/aiutil"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

var log = logging.C()

// logDebugArtifact delegates to the shared AI utility for debug artifact logging.
func logDebugArtifact(label, content string) {
	aiutil.LogDebugArtifact(log, label, content)
}

// logDebugArtifactf logs debug content with a formatted label.
func logDebugArtifactf(format, content string, args ...interface{}) {
	label := fmt.Sprintf(format, args...)
	logDebugArtifact(label, content)
}

// ValidationError is an alias for commitmessageinternal.ValidationError for external access.
type ValidationError = commitmessageinternal.ValidationError

// VerifyCommitMessageContract validates a commit message against the contract rules.
// This is exposed for testing purposes.
func VerifyCommitMessageContract(commitMessage string, affectedModules []string) []ValidationError {
	return commitmessageinternal.VerifyCommitMessageContract(commitMessage, affectedModules)
}

// AutoCleanup performs automatic fixes on commit message before validation.
// This is exposed for testing purposes.
func AutoCleanup(commitMessage string) string {
	return commitmessageinternal.AutoCleanup(commitMessage)
}

// parseConfig parses command-line flags and returns configuration values.
func parseConfig() (debug, autoCommit bool, workspaceRoot string, err error) {
	args := os.Args[3:] // Skip program name, "create", and "commit-message"

	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		return false, false, "", err
	}

	// Parse flags using shared package
	debug = flags.ParseDebugFlag(args)
	autoCommit = flags.HasFlag(args, "--commit", "-c")

	// Get repository root
	workspaceRoot, err = repository.GetRepositoryRoot("")
	if err != nil {
		return false, false, "", fmt.Errorf("failed to find repository root: %w", err)
	}

	return debug, autoCommit, workspaceRoot, nil
}
