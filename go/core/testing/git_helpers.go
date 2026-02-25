// Package testing provides git helper functions using go-git (no subprocess).
//
// These helpers replace exec.Command("git", ...) calls in test code, eliminating
// subprocess overhead (~200-500ms per call on Windows). They use the go/core/git
// package which wraps go-git/v5.
package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	coregit "github.com/ready-to-release/eac/go/core/git"
)

// GitInit initializes a new git repository at the given path using go-git.
// Configures user.name/email and creates an initial commit on "main" branch.
// Returns the Repository for subsequent operations.
func GitInit(repoPath string) (*coregit.Repository, error) {
	mgr := coregit.NewManager(nil)
	repo, err := mgr.Init(repoPath)
	if err != nil {
		return nil, fmt.Errorf("git init failed: %w", err)
	}

	if err := repo.ConfigSet("user", "name", "Test User"); err != nil {
		return nil, fmt.Errorf("git config user.name failed: %w", err)
	}
	if err := repo.ConfigSet("user", "email", "test@example.com"); err != nil {
		return nil, fmt.Errorf("git config user.email failed: %w", err)
	}

	// Disable autocrlf by writing directly to .git/config.
	// go-git's ConfigSet may not produce format that subprocess git reads reliably.
	// On Windows, global core.autocrlf=true causes worktree checkouts to apply
	// CRLF conversion, making all files appear modified (` M` in git status).
	gitConfigPath := filepath.Join(repoPath, ".git", "config")
	configData, err := os.ReadFile(gitConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read .git/config: %w", err)
	}
	configStr := string(configData)
	// Append autocrlf=false to [core] section if not already present
	if !strings.Contains(configStr, "autocrlf") {
		configStr = strings.Replace(configStr, "[core]", "[core]\n\tautocrlf = false", 1)
		if err := os.WriteFile(gitConfigPath, []byte(configStr), 0o644); err != nil {
			return nil, fmt.Errorf("failed to write .git/config: %w", err)
		}
	}

	// Create initial commit so HEAD exists
	readmePath := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repository\n"), 0o644); err != nil {
		return nil, fmt.Errorf("failed to create README.md: %w", err)
	}

	// Create .gitignore matching the real repo's ignore patterns.
	// This prevents subprocess log/build artifacts (out/) from appearing
	// as uncommitted changes in isolated test environments.
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("out/*\nout/**\n!out/.gitkeep\n"), 0o644); err != nil {
		return nil, fmt.Errorf("failed to create .gitignore: %w", err)
	}

	// Create .gitattributes to disable ALL text processing / line ending conversion.
	// On Windows, global core.autocrlf=true causes worktree checkouts to apply
	// CRLF conversion, making all files appear modified (` M` in git status).
	// Using `* -text` at the repo level overrides global config reliably.
	gitattrsPath := filepath.Join(repoPath, ".gitattributes")
	if err := os.WriteFile(gitattrsPath, []byte("* -text\n"), 0o644); err != nil {
		return nil, fmt.Errorf("failed to create .gitattributes: %w", err)
	}

	if err := repo.Add("README.md"); err != nil {
		return nil, fmt.Errorf("git add README.md failed: %w", err)
	}
	if err := repo.Add(".gitignore"); err != nil {
		return nil, fmt.Errorf("git add .gitignore failed: %w", err)
	}
	if err := repo.Add(".gitattributes"); err != nil {
		return nil, fmt.Errorf("git add .gitattributes failed: %w", err)
	}

	if _, err := repo.Commit("Initial commit", "Test User", "test@example.com"); err != nil {
		return nil, fmt.Errorf("git initial commit failed: %w", err)
	}

	// Ensure we're on "main" branch (go-git default may be "master")
	goRepo := repo.GoGitRepo()
	head, err := goRepo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	if head.Name() != plumbing.NewBranchReferenceName("main") {
		// Rename current branch to "main"
		ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), head.Hash())
		if err := goRepo.Storer.SetReference(ref); err != nil {
			return nil, fmt.Errorf("failed to create main branch ref: %w", err)
		}
		// Update HEAD to point to main
		symRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName("main"))
		if err := goRepo.Storer.SetReference(symRef); err != nil {
			return nil, fmt.Errorf("failed to update HEAD to main: %w", err)
		}
		// Remove old branch reference
		if err := goRepo.Storer.RemoveReference(head.Name()); err != nil {
			return nil, fmt.Errorf("failed to remove old branch ref: %w", err)
		}
	}

	return repo, nil
}

// GitAddAll stages all new and modified files in the worktree (equivalent to git add -A).
func GitAddAll(repoPath string) error {
	mgr := coregit.NewManager(nil)
	repo, err := mgr.Open(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	goRepo := repo.GoGitRepo()
	wt, err := goRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// AddWithOptions with All:true handles new, modified, and deleted files (like git add -A)
	if err := wt.AddWithOptions(&gogit.AddOptions{All: true}); err != nil {
		return fmt.Errorf("git add -A failed: %w", err)
	}

	return nil
}

// GitCommit creates a commit with all staged changes using go-git.
// Returns the commit SHA.
func GitCommit(repoPath, message string) (string, error) {
	mgr := coregit.NewManager(nil)
	repo, err := mgr.Open(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}
	return repo.Commit(message, "Test User", "test@example.com")
}

// GitHeadSHA returns the HEAD commit SHA using go-git.
func GitHeadSHA(repoPath string) (string, error) {
	mgr := coregit.NewManager(nil)
	repo, err := mgr.Open(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}
	return repo.HeadCommit()
}

// GitCheckoutPath restores a file or directory from HEAD using go-git worktree checkout.
func GitCheckoutPath(repoPath, path string) error {
	mgr := coregit.NewManager(nil)
	repo, err := mgr.Open(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	goRepo := repo.GoGitRepo()
	wt, err := goRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := wt.Checkout(&gogit.CheckoutOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}

	_ = path // Full worktree checkout restores all tracked files
	return nil
}

// GitAddRemote adds a remote to the repository.
func GitAddRemote(repoPath, name, url string) error {
	mgr := coregit.NewManager(nil)
	repo, err := mgr.Open(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	return repo.AddRemote(name, url)
}
