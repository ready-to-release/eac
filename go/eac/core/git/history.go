package git

import (
	"fmt"
	"sort"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// CommitInfo represents minimal commit information for changelog generation
type CommitInfo struct {
	// SHA is the full commit hash
	SHA string

	// ShortSHA is the abbreviated commit hash (7 chars)
	ShortSHA string

	// Message is the full commit message
	Message string

	// Subject is the first line of the commit message
	Subject string

	// Author is the commit author name
	Author string

	// AuthorEmail is the commit author email
	AuthorEmail string

	// Date is the commit timestamp
	Date time.Time

	// Files is the list of files changed in this commit
	Files []string
}

// CommitsBetween returns commits between two references (tag/SHA/branch).
// If fromRef is empty, returns all commits up to toRef.
// Returns commits in reverse chronological order (newest first).
func (r *Repository) CommitsBetween(fromRef, toRef string) ([]CommitInfo, error) {
	// Resolve toRef
	toHash, err := r.resolveRef(toRef)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %q: %w", toRef, err)
	}

	// Resolve fromRef if provided
	var fromHash plumbing.Hash
	if fromRef != "" {
		fromHash, err = r.resolveRef(fromRef)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %q: %w", fromRef, err)
		}
	}

	// Get commit iterator starting from toRef
	commitIter, err := r.repo.Log(&gogit.LogOptions{
		From:  toHash,
		Order: gogit.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}
	defer commitIter.Close()

	var commits []CommitInfo

	err = commitIter.ForEach(func(c *object.Commit) error {
		// Stop if we've reached the fromRef commit
		if fromRef != "" && c.Hash == fromHash {
			return storer.ErrStop // Stop iteration, don't include the fromRef commit
		}

		info := CommitInfo{
			SHA:         c.Hash.String(),
			ShortSHA:    c.Hash.String()[:7],
			Message:     c.Message,
			Subject:     strings.SplitN(c.Message, "\n", 2)[0],
			Author:      c.Author.Name,
			AuthorEmail: c.Author.Email,
			Date:        c.Author.When,
		}

		// Get files changed in this commit
		files, err := r.getCommitFiles(c)
		if err == nil {
			info.Files = files
		}

		commits = append(commits, info)
		return nil
	})

	if err != nil && err != storer.ErrStop {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

// CommitsSince returns all commits since a given tag or reference.
// This is a convenience wrapper around CommitsBetween.
func (r *Repository) CommitsSince(ref string) ([]CommitInfo, error) {
	return r.CommitsBetween(ref, "HEAD")
}

// getCommitFiles returns the list of files changed in a commit
func (r *Repository) getCommitFiles(c *object.Commit) ([]string, error) {
	var files []string

	// Get the tree for this commit
	tree, err := c.Tree()
	if err != nil {
		return nil, err
	}

	// Get parent commit's tree for comparison
	var parentTree *object.Tree
	if c.NumParents() > 0 {
		parent, err := c.Parent(0)
		if err == nil {
			parentTree, _ = parent.Tree()
		}
	}

	// Compare trees to find changed files
	if parentTree == nil {
		// First commit - all files are new
		err = tree.Files().ForEach(func(f *object.File) error {
			files = append(files, f.Name)
			return nil
		})
	} else {
		// Compare with parent
		changes, err := parentTree.Diff(tree)
		if err != nil {
			return nil, err
		}

		for _, change := range changes {
			// Get the file path from either From or To
			name := change.From.Name
			if name == "" {
				name = change.To.Name
			}
			if name != "" {
				files = append(files, name)
			}
		}
	}

	return files, err
}

// TagsMatching returns tags matching a glob pattern (e.g., "r2r-cli/*").
// Tags are returned sorted by version (newest first for semver).
func (r *Repository) TagsMatching(pattern string) ([]string, error) {
	tagIter, err := r.repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer tagIter.Close()

	var tags []string

	err = tagIter.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()

		// Simple glob matching
		if matchTagPattern(tagName, pattern) {
			tags = append(tags, tagName)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	// Sort tags (simple string sort for now - semver sorting can be added)
	sort.Sort(sort.Reverse(sort.StringSlice(tags)))

	return tags, nil
}

// LatestTag returns the most recent tag matching the pattern.
// Returns empty string if no matching tags found.
func (r *Repository) LatestTag(pattern string) (string, error) {
	tags, err := r.TagsMatching(pattern)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", nil
	}

	return tags[0], nil
}

// TagCommit returns the commit SHA that a tag points to.
// Works for both lightweight and annotated tags.
func (r *Repository) TagCommit(tagName string) (string, error) {
	// Try to resolve as a tag reference
	ref, err := r.repo.Reference(plumbing.NewTagReferenceName(tagName), true)
	if err != nil {
		return "", fmt.Errorf("tag %q not found: %w", tagName, err)
	}

	// For annotated tags, we need to dereference to get the commit
	hash := ref.Hash()

	// Try to get the tag object (for annotated tags)
	tagObj, err := r.repo.TagObject(hash)
	if err == nil {
		// Annotated tag - get the commit it points to
		commit, err := tagObj.Commit()
		if err == nil {
			return commit.Hash.String(), nil
		}
	}

	// Lightweight tag or direct reference - hash is the commit
	return hash.String(), nil
}

// TagDate returns the date of a tag.
// For annotated tags, returns the tag date; for lightweight tags, returns the commit date.
func (r *Repository) TagDate(tagName string) (time.Time, error) {
	ref, err := r.repo.Reference(plumbing.NewTagReferenceName(tagName), true)
	if err != nil {
		return time.Time{}, fmt.Errorf("tag %q not found: %w", tagName, err)
	}

	hash := ref.Hash()

	// Try annotated tag first
	tagObj, err := r.repo.TagObject(hash)
	if err == nil {
		return tagObj.Tagger.When, nil
	}

	// Lightweight tag - get commit date
	commit, err := r.repo.CommitObject(hash)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get commit for tag: %w", err)
	}

	return commit.Author.When, nil
}

// TagExists checks if a tag with the given name exists.
func (r *Repository) TagExists(tagName string) (bool, error) {
	_, err := r.repo.Reference(plumbing.NewTagReferenceName(tagName), true)
	if err != nil {
		if err == plumbing.ErrReferenceNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// resolveRef resolves a reference string to a hash.
// Supports: branch names, tag names, SHA, and special refs like HEAD.
func (r *Repository) resolveRef(ref string) (plumbing.Hash, error) {
	if ref == "" || ref == "HEAD" {
		head, err := r.repo.Head()
		if err != nil {
			return plumbing.ZeroHash, err
		}
		return head.Hash(), nil
	}

	// Try as a tag first (e.g., "r2r-cli/0.0.14")
	tagRef, err := r.repo.Reference(plumbing.NewTagReferenceName(ref), true)
	if err == nil {
		// For annotated tags, dereference
		tagObj, err := r.repo.TagObject(tagRef.Hash())
		if err == nil {
			commit, err := tagObj.Commit()
			if err == nil {
				return commit.Hash, nil
			}
		}
		return tagRef.Hash(), nil
	}

	// Try as a branch
	branchRef, err := r.repo.Reference(plumbing.NewBranchReferenceName(ref), true)
	if err == nil {
		return branchRef.Hash(), nil
	}

	// Try as a remote branch (e.g., "origin/main")
	// Check if ref already has remote prefix
	if strings.Contains(ref, "/") {
		parts := strings.SplitN(ref, "/", 2)
		if len(parts) == 2 {
			remoteName := parts[0]
			branchName := parts[1]
			remoteBranchRef, err := r.repo.Reference(plumbing.NewRemoteReferenceName(remoteName, branchName), true)
			if err == nil {
				return remoteBranchRef.Hash(), nil
			}
		}
	} else {
		// Try with "origin/" prefix if no slash in ref
		remoteBranchRef, err := r.repo.Reference(plumbing.NewRemoteReferenceName("origin", ref), true)
		if err == nil {
			return remoteBranchRef.Hash(), nil
		}
	}

	// Try as a direct hash
	if len(ref) >= 7 {
		hash := plumbing.NewHash(ref)
		if hash != plumbing.ZeroHash {
			// Verify it exists
			_, err := r.repo.CommitObject(hash)
			if err == nil {
				return hash, nil
			}
		}
	}

	return plumbing.ZeroHash, fmt.Errorf("could not resolve reference %q", ref)
}

// matchTagPattern matches a tag name against a simple glob pattern.
// Supports: * (matches anything), and prefix matching (e.g., "r2r-cli/*").
func matchTagPattern(tagName, pattern string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}

	// Handle prefix/* pattern
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(tagName, prefix+"/")
	}

	// Handle prefix* pattern
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(tagName, prefix)
	}

	// Exact match
	return tagName == pattern
}
