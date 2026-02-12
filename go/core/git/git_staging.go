package git

import (
	"fmt"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// ConfigSet sets a Git configuration value.
func (r *Repository) ConfigSet(section, key, value string) error {
	cfg, err := r.repo.Config()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Handle user.name and user.email specially
	if section == "user" {
		if cfg.User.Name == "" && key == "name" {
			cfg.User.Name = value
		} else if cfg.User.Email == "" && key == "email" {
			cfg.User.Email = value
		}

		switch key {
		case "name":
			cfg.User.Name = value
		case "email":
			cfg.User.Email = value
		}
	} else {
		// For other config, use raw sections
		if cfg.Raw.Section(section) == nil {
			cfg.Raw.AddOption(section, "", key, value)
		} else {
			cfg.Raw.Section(section).SetOption(key, value)
		}
	}

	return r.repo.SetConfig(cfg)
}

// Add stages a file for commit.
func (r *Repository) Add(path string) error {
	wt, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = wt.Add(path)
	if err != nil {
		return fmt.Errorf("failed to add file %q: %w", path, err)
	}

	return nil
}

// Commit creates a new commit with the staged changes.
func (r *Repository) Commit(message, authorName, authorEmail string) (string, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	sig := &object.Signature{
		Name:  authorName,
		Email: authorEmail,
		When:  r.now(),
	}

	hash, err := wt.Commit(message, &gogit.CommitOptions{
		Author: sig,
	})
	if err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}

	return hash.String(), nil
}

// AddRemote adds a new remote to the repository.
func (r *Repository) AddRemote(name, url string) error {
	_, err := r.repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})
	if err != nil {
		return fmt.Errorf("failed to add remote %q: %w", name, err)
	}
	return nil
}
