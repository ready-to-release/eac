// Package git provides git-related utilities
package git

import (
	"strings"

	"github.com/ready-to-release/eac/go/clibase/gitexec"
)

// GetCommitSHA returns the current git commit SHA for the workspace.
func GetCommitSHA(workspaceRoot string) string {
	output, err := gitexec.Run(workspaceRoot, "rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
