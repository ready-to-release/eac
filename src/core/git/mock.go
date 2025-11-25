package git

import (
	"errors"
	"fmt"

	gogit "github.com/go-git/go-git/v5"
)

// MockRepository implements GitRepository for testing.
// Use the builder methods to configure behavior.
type MockRepository struct {
	rootPath        string
	remotes         map[string]string
	currentBranch   string
	headSHA         string
	trackedFiles    []string
	stagedFiles     []string
	stagedDiff      string
	stagedDiffStats string
	ignoredFiles    map[string]bool
	configs         map[string]map[string]string
	addedFiles      []string
	commits         []MockCommit

	// Error injection for testing failure paths
	RemoteURLError       error
	CurrentBranchError   error
	HeadSHAError         error
	TrackedFilesError    error
	StagedFilesError     error
	StagedDiffError      error
	StagedDiffStatsError error
	AddError             error
	CommitError          error
	ConfigSetError       error
	AddRemoteError       error
}

// MockCommit records commit information for assertions.
type MockCommit struct {
	Message     string
	AuthorName  string
	AuthorEmail string
	Hash        string
}

// NewMockRepository creates a new mock repository with the given root path.
func NewMockRepository(rootPath string) *MockRepository {
	return &MockRepository{
		rootPath:     rootPath,
		remotes:      make(map[string]string),
		ignoredFiles: make(map[string]bool),
		configs:      make(map[string]map[string]string),
	}
}

// Builder methods for configuring the mock

// WithRemote adds a remote URL.
func (m *MockRepository) WithRemote(name, url string) *MockRepository {
	m.remotes[name] = url
	return m
}

// WithCurrentBranch sets the current branch name.
func (m *MockRepository) WithCurrentBranch(branch string) *MockRepository {
	m.currentBranch = branch
	return m
}

// WithHeadSHA sets the HEAD commit SHA.
func (m *MockRepository) WithHeadSHA(sha string) *MockRepository {
	m.headSHA = sha
	return m
}

// WithTrackedFiles sets the list of tracked files.
func (m *MockRepository) WithTrackedFiles(files []string) *MockRepository {
	m.trackedFiles = files
	return m
}

// WithStagedFiles sets the list of staged files.
func (m *MockRepository) WithStagedFiles(files []string) *MockRepository {
	m.stagedFiles = files
	return m
}

// WithIgnoredFile marks a file as ignored.
func (m *MockRepository) WithIgnoredFile(path string) *MockRepository {
	m.ignoredFiles[path] = true
	return m
}

// WithIgnoredFiles marks multiple files as ignored.
func (m *MockRepository) WithIgnoredFiles(paths []string) *MockRepository {
	for _, p := range paths {
		m.ignoredFiles[p] = true
	}
	return m
}

// WithStagedDiff sets the staged diff content.
func (m *MockRepository) WithStagedDiff(diff string) *MockRepository {
	m.stagedDiff = diff
	return m
}

// WithStagedDiffStats sets the staged diff stats.
func (m *MockRepository) WithStagedDiffStats(stats string) *MockRepository {
	m.stagedDiffStats = stats
	return m
}

// WithError sets an error for a specific operation.
func (m *MockRepository) WithError(operation string, err error) *MockRepository {
	switch operation {
	case "RemoteURL":
		m.RemoteURLError = err
	case "CurrentBranch":
		m.CurrentBranchError = err
	case "HeadSHA":
		m.HeadSHAError = err
	case "TrackedFiles":
		m.TrackedFilesError = err
	case "StagedFiles":
		m.StagedFilesError = err
	case "StagedDiff":
		m.StagedDiffError = err
	case "StagedDiffStats":
		m.StagedDiffStatsError = err
	case "Add":
		m.AddError = err
	case "Commit":
		m.CommitError = err
	case "ConfigSet":
		m.ConfigSetError = err
	case "AddRemote":
		m.AddRemoteError = err
	}
	return m
}

// GitRepository interface implementation

func (m *MockRepository) RootPath() string {
	return m.rootPath
}

func (m *MockRepository) RemoteURL(remoteName string) (string, error) {
	if m.RemoteURLError != nil {
		return "", m.RemoteURLError
	}
	url, ok := m.remotes[remoteName]
	if !ok {
		return "", fmt.Errorf("remote %q not found", remoteName)
	}
	return url, nil
}

func (m *MockRepository) AddRemote(name, url string) error {
	if m.AddRemoteError != nil {
		return m.AddRemoteError
	}
	if _, exists := m.remotes[name]; exists {
		return fmt.Errorf("remote %q already exists", name)
	}
	m.remotes[name] = url
	return nil
}

func (m *MockRepository) CurrentBranch() (string, error) {
	if m.CurrentBranchError != nil {
		return "", m.CurrentBranchError
	}
	if m.currentBranch == "" {
		return "", errors.New("no branch set")
	}
	return m.currentBranch, nil
}

func (m *MockRepository) HeadShortSHA() (string, error) {
	if m.HeadSHAError != nil {
		return "", m.HeadSHAError
	}
	if m.headSHA == "" {
		return "", errors.New("no HEAD commit")
	}
	if len(m.headSHA) > 7 {
		return m.headSHA[:7], nil
	}
	return m.headSHA, nil
}

func (m *MockRepository) TrackedFiles() ([]string, error) {
	if m.TrackedFilesError != nil {
		return nil, m.TrackedFilesError
	}
	return m.trackedFiles, nil
}

func (m *MockRepository) StagedFiles() ([]string, error) {
	if m.StagedFilesError != nil {
		return nil, m.StagedFilesError
	}
	return m.stagedFiles, nil
}

func (m *MockRepository) StagedDiff() (string, error) {
	if m.StagedDiffError != nil {
		return "", m.StagedDiffError
	}
	return m.stagedDiff, nil
}

func (m *MockRepository) StagedDiffStats() (string, error) {
	if m.StagedDiffStatsError != nil {
		return "", m.StagedDiffStatsError
	}
	return m.stagedDiffStats, nil
}

func (m *MockRepository) IsFileTracked(relPath string) bool {
	for _, f := range m.trackedFiles {
		if f == relPath {
			return true
		}
	}
	return false
}

func (m *MockRepository) IsFileIgnored(relPath string) bool {
	return m.ignoredFiles[relPath]
}

func (m *MockRepository) Add(path string) error {
	if m.AddError != nil {
		return m.AddError
	}
	m.addedFiles = append(m.addedFiles, path)
	return nil
}

func (m *MockRepository) Commit(message, authorName, authorEmail string) (string, error) {
	if m.CommitError != nil {
		return "", m.CommitError
	}
	hash := fmt.Sprintf("mock%040d", len(m.commits)+1)
	m.commits = append(m.commits, MockCommit{
		Message:     message,
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
		Hash:        hash,
	})
	return hash, nil
}

func (m *MockRepository) ConfigSet(section, key, value string) error {
	if m.ConfigSetError != nil {
		return m.ConfigSetError
	}
	if m.configs[section] == nil {
		m.configs[section] = make(map[string]string)
	}
	m.configs[section][key] = value
	return nil
}

func (m *MockRepository) GoGitRepo() *gogit.Repository {
	return nil // Mock doesn't have underlying go-git repo
}

// Assertion helpers for tests

// AddedFiles returns the list of files that were added via Add().
func (m *MockRepository) AddedFiles() []string {
	return m.addedFiles
}

// Commits returns the list of commits made via Commit().
func (m *MockRepository) Commits() []MockCommit {
	return m.commits
}

// Config returns a config value that was set.
func (m *MockRepository) Config(section, key string) (string, bool) {
	if m.configs[section] == nil {
		return "", false
	}
	val, ok := m.configs[section][key]
	return val, ok
}

// Ensure MockRepository implements GitRepository
var _ GitRepository = (*MockRepository)(nil)
