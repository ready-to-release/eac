// Package testutil provides shared test utilities for command testing.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// GitRepo represents a temporary git repository for testing.
type GitRepo struct {
	t    *testing.T
	root string
}

// NewGitRepo creates a new temporary git repository.
// The repository is initialized with git init and cleaned up after the test.
func NewGitRepo(t *testing.T) *GitRepo {
	t.Helper()

	dir := TempDir(t, "git-test-*")

	repo := &GitRepo{
		t:    t,
		root: dir,
	}

	repo.Git("init")
	repo.Git("config", "user.email", "test@example.com")
	repo.Git("config", "user.name", "Test User")

	return repo
}

// Root returns the root directory of the git repository.
func (r *GitRepo) Root() string {
	return r.root
}

// Git runs a git command in the repository.
func (r *GitRepo) Git(args ...string) string {
	r.t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = r.root
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v failed: %v\nOutput: %s", args, err, out)
	}
	return string(out)
}

// GitMayFail runs a git command that is allowed to fail.
// Returns the output and error.
func (r *GitRepo) GitMayFail(args ...string) (string, error) {
	r.t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = r.root
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// WriteFile creates or overwrites a file in the repository.
func (r *GitRepo) WriteFile(relPath, content string) {
	r.t.Helper()

	fullPath := filepath.Join(r.root, relPath)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		r.t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		r.t.Fatalf("failed to write file %s: %v", fullPath, err)
	}
}

// ReadFile reads a file from the repository.
func (r *GitRepo) ReadFile(relPath string) string {
	r.t.Helper()

	fullPath := filepath.Join(r.root, relPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		r.t.Fatalf("failed to read file %s: %v", fullPath, err)
	}
	return string(content)
}

// Add stages files for commit.
func (r *GitRepo) Add(paths ...string) {
	r.t.Helper()
	args := append([]string{"add"}, paths...)
	r.Git(args...)
}

// AddAll stages all changes.
func (r *GitRepo) AddAll() {
	r.t.Helper()
	r.Git("add", "-A")
}

// Commit creates a commit with the given message.
func (r *GitRepo) Commit(message string) string {
	r.t.Helper()
	return r.Git("commit", "-m", message)
}

// CommitAll stages all changes and commits.
func (r *GitRepo) CommitAll(message string) string {
	r.t.Helper()
	r.AddAll()
	return r.Commit(message)
}

// Tag creates a tag.
func (r *GitRepo) Tag(name string) {
	r.t.Helper()
	r.Git("tag", name)
}

// TagAnnotated creates an annotated tag.
func (r *GitRepo) TagAnnotated(name, message string) {
	r.t.Helper()
	r.Git("tag", "-a", name, "-m", message)
}

// Branch creates and optionally switches to a new branch.
func (r *GitRepo) Branch(name string, checkout bool) {
	r.t.Helper()
	if checkout {
		r.Git("checkout", "-b", name)
	} else {
		r.Git("branch", name)
	}
}

// Checkout switches to a branch or commit.
func (r *GitRepo) Checkout(ref string) {
	r.t.Helper()
	r.Git("checkout", ref)
}

// CurrentBranch returns the current branch name.
func (r *GitRepo) CurrentBranch() string {
	r.t.Helper()
	out := r.Git("rev-parse", "--abbrev-ref", "HEAD")
	// Trim trailing newline
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out
}

// CurrentCommit returns the current commit SHA.
func (r *GitRepo) CurrentCommit() string {
	r.t.Helper()
	out := r.Git("rev-parse", "HEAD")
	// Trim trailing newline
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out
}

// HasUncommittedChanges returns true if there are uncommitted changes.
func (r *GitRepo) HasUncommittedChanges() bool {
	r.t.Helper()
	_, err := r.GitMayFail("diff", "--quiet")
	return err != nil
}

// HasStagedChanges returns true if there are staged changes.
func (r *GitRepo) HasStagedChanges() bool {
	r.t.Helper()
	_, err := r.GitMayFail("diff", "--cached", "--quiet")
	return err != nil
}

// SetupEACConfig creates a minimal EAC config in the git repository.
func (r *GitRepo) SetupEACConfig() {
	r.t.Helper()

	r.WriteFile(".r2r/eac/repository.yml", `
name: test-repo
description: Test repository
versioning:
  scheme: SemVer
modules: []
`)

	r.WriteFile(".r2r/eac/module-types.yml", `
module_types: []
`)
}

// SetupEACConfigWithModules creates an EAC config with the given modules.
func (r *GitRepo) SetupEACConfigWithModules(modules []ModuleSpec) {
	r.t.Helper()

	// Build modules YAML
	modulesYAML := "modules:\n"
	for _, m := range modules {
		modulesYAML += "  - moniker: " + m.Moniker + "\n"
		modulesYAML += "    name: " + m.Name + "\n"
		modulesYAML += "    type: " + m.Type + "\n"
		if m.Description != "" {
			modulesYAML += "    description: " + m.Description + "\n"
		}
	}

	r.WriteFile(".r2r/eac/repository.yml", `
name: test-repo
description: Test repository
versioning:
  scheme: SemVer
`+modulesYAML)

	// Create minimal module-types.yml with referenced types
	typesYAML := "module_types:\n"
	seenTypes := make(map[string]bool)
	for _, m := range modules {
		if !seenTypes[m.Type] {
			typesYAML += "  - name: " + m.Type + "\n"
			typesYAML += "    description: " + m.Type + " module type\n"
			seenTypes[m.Type] = true
		}
	}

	r.WriteFile(".r2r/eac/module-types.yml", typesYAML)
}
