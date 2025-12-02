package assessment

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/git"
)

// FileInfo represents information about a changed file
type FileInfo struct {
	Path   string
	Status string // added, modified, deleted
}

// ============================================================================
// Mock Support for Testing
// ============================================================================

// gitRepo holds the git repository instance for git operations.
// In production, this is initialized lazily. For tests, it can be injected via SetGitRepo.
var gitRepo git.GitRepository

// getGitRepo returns the git repository, creating one if needed
func getGitRepo(workspaceRoot string) (git.GitRepository, error) {
	if gitRepo != nil {
		return gitRepo, nil
	}
	return git.Open(workspaceRoot)
}

// SetGitRepo allows tests to inject a mock repository.
func SetGitRepo(repo git.GitRepository) {
	gitRepo = repo
}

// ResetGitRepo clears the mock git repository.
func ResetGitRepo() {
	gitRepo = nil
}

// ============================================================================

// getFilesInScope returns files based on the specified scope
func getFilesInScope(scope string, workspaceRoot string) ([]FileInfo, error) {
	switch scope {
	case "staged":
		return getStagedFiles(workspaceRoot)
	case "changed":
		return getChangedFiles(workspaceRoot)
	case "all":
		return getAllFiles(workspaceRoot)
	default:
		return nil, fmt.Errorf("invalid scope: %s", scope)
	}
}

// getStagedFiles returns files in the git staging area
func getStagedFiles(workspaceRoot string) ([]FileInfo, error) {
	repo, err := getGitRepo(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	stagedFiles, err := repo.StagedFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	files := make([]FileInfo, 0, len(stagedFiles))
	for _, file := range stagedFiles {
		files = append(files, FileInfo{
			Path:   file,
			Status: "modified", // Simplified - could determine A/M/D from diff
		})
	}

	return files, nil
}

// getChangedFiles returns modified but unstaged files
func getChangedFiles(workspaceRoot string) ([]FileInfo, error) {
	repo, err := getGitRepo(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get diff of working tree vs index to find unstaged changes
	// For now, return all tracked files as a simplified implementation
	// TODO: Enhance to detect actual changed files
	trackedFiles, err := repo.TrackedFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get tracked files: %w", err)
	}

	files := make([]FileInfo, 0, len(trackedFiles))
	for _, file := range trackedFiles {
		files = append(files, FileInfo{
			Path:   file,
			Status: "tracked",
		})
	}

	return files, nil
}

// getAllFiles returns all tracked files in the repository
func getAllFiles(workspaceRoot string) ([]FileInfo, error) {
	repo, err := getGitRepo(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	trackedFiles, err := repo.TrackedFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get all files: %w", err)
	}

	files := make([]FileInfo, 0, len(trackedFiles))
	for _, file := range trackedFiles {
		files = append(files, FileInfo{
			Path:   file,
			Status: "tracked",
		})
	}

	return files, nil
}

// parseGitStatus parses git diff output (e.g., "M\tfile.go")
func parseGitStatus(output string) []FileInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	files := make([]FileInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		status := parts[0]
		path := parts[1]

		statusName := "modified"
		switch status {
		case "A":
			statusName = "added"
		case "M":
			statusName = "modified"
		case "D":
			statusName = "deleted"
		case "R":
			statusName = "renamed"
		}

		files = append(files, FileInfo{
			Path:   path,
			Status: statusName,
		})
	}

	return files
}

// readFileContent reads the content of a file
func readFileContent(path string, workspaceRoot string) (string, error) {
	fullPath := path
	if !strings.HasPrefix(path, workspaceRoot) {
		fullPath = workspaceRoot + string(os.PathSeparator) + path
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return string(content), nil
}
