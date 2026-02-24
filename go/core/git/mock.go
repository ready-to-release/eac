package git

import (
	"errors"
	"fmt"
	"time"

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

	// Changelog/release related fields
	commitHistory []CommitInfo
	tags          map[string]MockTag

	// Branch comparison fields
	branchCommits   []CommitInfo
	branchDiff      string
	branchDiffStats string
	branchFiles     []string
	upstreamBranch  string

	// Uncommitted files for change detection
	uncommittedFiles []string

	// Error injection for testing failure paths
	RemoteURLError          error
	CurrentBranchError      error
	HeadSHAError            error
	HeadCommitError         error
	UncommittedFilesError   error
	TrackedFilesError       error
	StagedFilesError        error
	StagedDiffError         error
	StagedDiffStatsError    error
	AddError                error
	CommitError             error
	ConfigSetError          error
	AddRemoteError          error
	CommitsBetweenError     error
	TagsMatchingError       error
	TagCommitError          error
	TagDateError            error
	GetBranchCommitsError   error
	GetBranchDiffError      error
	GetBranchDiffStatsError error
	GetBranchFilesError     error
	UpstreamBranchError     error
}

// MockTag represents a mock tag for testing.
type MockTag struct {
	CommitSHA string
	Date      time.Time
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
		tags:         make(map[string]MockTag),
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

// WithUncommittedFiles sets the list of uncommitted files.
func (m *MockRepository) WithUncommittedFiles(files []string) *MockRepository {
	m.uncommittedFiles = files
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
	case "HeadCommit":
		m.HeadCommitError = err
	case "UncommittedFiles":
		m.UncommittedFilesError = err
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

func (m *MockRepository) HeadCommit() (string, error) {
	if m.HeadCommitError != nil {
		return "", m.HeadCommitError
	}
	if m.headSHA == "" {
		return "", errors.New("no HEAD commit")
	}
	return m.headSHA, nil
}

func (m *MockRepository) UncommittedFiles() ([]string, error) {
	if m.UncommittedFilesError != nil {
		return nil, m.UncommittedFilesError
	}
	return m.uncommittedFiles, nil
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

// --- Changelog/Release related mock implementations ---

// WithCommitHistory sets the commit history for CommitsBetween/CommitsSince.
func (m *MockRepository) WithCommitHistory(commits []CommitInfo) *MockRepository {
	m.commitHistory = commits
	return m
}

// WithTag adds a mock tag.
func (m *MockRepository) WithTag(name, commitSHA string, date time.Time) *MockRepository {
	m.tags[name] = MockTag{CommitSHA: commitSHA, Date: date}
	return m
}

func (m *MockRepository) CommitsBetween(fromRef, toRef string) ([]CommitInfo, error) {
	if m.CommitsBetweenError != nil {
		return nil, m.CommitsBetweenError
	}
	return m.commitHistory, nil
}

func (m *MockRepository) CommitsSince(ref string) ([]CommitInfo, error) {
	return m.CommitsBetween(ref, "HEAD")
}

func (m *MockRepository) TagsMatching(pattern string) ([]string, error) {
	if m.TagsMatchingError != nil {
		return nil, m.TagsMatchingError
	}
	var result []string
	for name := range m.tags {
		if matchTagPattern(name, pattern) {
			result = append(result, name)
		}
	}
	return result, nil
}

func (m *MockRepository) LatestTag(pattern string) (string, error) {
	tags, err := m.TagsMatching(pattern)
	if err != nil {
		return "", err
	}
	if len(tags) == 0 {
		return "", nil
	}
	return tags[0], nil
}

func (m *MockRepository) TagCommit(tagName string) (string, error) {
	if m.TagCommitError != nil {
		return "", m.TagCommitError
	}
	tag, ok := m.tags[tagName]
	if !ok {
		return "", fmt.Errorf("tag %q not found", tagName)
	}
	return tag.CommitSHA, nil
}

func (m *MockRepository) TagDate(tagName string) (time.Time, error) {
	if m.TagDateError != nil {
		return time.Time{}, m.TagDateError
	}
	tag, ok := m.tags[tagName]
	if !ok {
		return time.Time{}, fmt.Errorf("tag %q not found", tagName)
	}
	return tag.Date, nil
}

func (m *MockRepository) TagExists(tagName string) (bool, error) {
	_, ok := m.tags[tagName]
	return ok, nil
}

// GetBranchCommits returns the mock branch commits.
func (m *MockRepository) GetBranchCommits(baseBranch string) ([]CommitInfo, error) {
	if m.GetBranchCommitsError != nil {
		return nil, m.GetBranchCommitsError
	}
	if len(m.branchCommits) == 0 {
		return nil, fmt.Errorf("no commits ahead of %s", baseBranch)
	}
	return m.branchCommits, nil
}

// GetBranchDiff returns the mock branch diff.
func (m *MockRepository) GetBranchDiff(baseBranch string) (string, error) {
	if m.GetBranchDiffError != nil {
		return "", m.GetBranchDiffError
	}
	return m.branchDiff, nil
}

// GetBranchDiffStats returns the mock branch diff stats.
func (m *MockRepository) GetBranchDiffStats(baseBranch string) (string, error) {
	if m.GetBranchDiffStatsError != nil {
		return "", m.GetBranchDiffStatsError
	}
	return m.branchDiffStats, nil
}

// GetBranchFiles returns the mock branch files.
func (m *MockRepository) GetBranchFiles(baseBranch string) ([]string, error) {
	if m.GetBranchFilesError != nil {
		return nil, m.GetBranchFilesError
	}
	return m.branchFiles, nil
}

// Builder methods for branch comparison

// WithBranchCommits sets the mock branch commits.
func (m *MockRepository) WithBranchCommits(commits []CommitInfo) *MockRepository {
	m.branchCommits = commits
	return m
}

// WithBranchDiff sets the mock branch diff.
func (m *MockRepository) WithBranchDiff(diff string) *MockRepository {
	m.branchDiff = diff
	return m
}

// WithBranchDiffStats sets the mock branch diff stats.
func (m *MockRepository) WithBranchDiffStats(stats string) *MockRepository {
	m.branchDiffStats = stats
	return m
}

// WithBranchFiles sets the mock branch files.
func (m *MockRepository) WithBranchFiles(files []string) *MockRepository {
	m.branchFiles = files
	return m
}

// WithUpstreamBranch sets the mock upstream branch.
func (m *MockRepository) WithUpstreamBranch(branch string) *MockRepository {
	m.upstreamBranch = branch
	return m
}

// UpstreamBranch returns the mock upstream tracking branch.
func (m *MockRepository) UpstreamBranch() (string, error) {
	if m.UpstreamBranchError != nil {
		return "", m.UpstreamBranchError
	}
	if m.upstreamBranch == "" {
		return "", fmt.Errorf("no upstream configured")
	}
	return m.upstreamBranch, nil
}

// Ensure MockRepository implements GitRepository (also verified in interface.go).
var _ GitRepository = (*MockRepository)(nil)
