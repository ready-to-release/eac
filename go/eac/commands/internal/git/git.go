// Package git provides git-related utilities
package git

import (
	"os/exec"
	"strings"
)

// GetCommitSHA returns the current git commit SHA for the workspace
func GetCommitSHA(workspaceRoot string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = workspaceRoot
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
