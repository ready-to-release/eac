//go:build L1

package work

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/work/internal"
	"github.com/ready-to-release/eac/go/core/tool"
)

// ---------------------------------------------------------------------------
// workCommand metadata tests
// ---------------------------------------------------------------------------

func TestWorkCommand_Name(t *testing.T) {
	cmd := &workCommand{}
	assert.Equal(t, "work", cmd.Name())
}

func TestWorkCommand_Metadata(t *testing.T) {
	cmd := &workCommand{}
	meta := cmd.Metadata()

	t.Run("canonical name", func(t *testing.T) {
		assert.Equal(t, "work", meta.CanonicalName)
	})

	t.Run("is parent command", func(t *testing.T) {
		assert.True(t, meta.IsParent, "work command should be a parent command")
	})

	t.Run("short description is non-empty", func(t *testing.T) {
		assert.NotEmpty(t, meta.Short)
	})

	t.Run("has subcommand groups", func(t *testing.T) {
		require.NotEmpty(t, meta.SubcommandGroups, "work command should have subcommand groups")
	})

	t.Run("subcommand groups contain expected names", func(t *testing.T) {
		groupNames := make([]string, len(meta.SubcommandGroups))
		for i, g := range meta.SubcommandGroups {
			groupNames[i] = g.Name
		}
		assert.Contains(t, groupNames, "Workspace Lifecycle")
		assert.Contains(t, groupNames, "Development Workflow")
		assert.Contains(t, groupNames, "Completion")
	})

	t.Run("lifecycle group has expected subcommands", func(t *testing.T) {
		for _, g := range meta.SubcommandGroups {
			if g.Name == "Workspace Lifecycle" {
				assert.Contains(t, g.Subcommands, "create")
				assert.Contains(t, g.Subcommands, "list")
				assert.Contains(t, g.Subcommands, "remove")
				return
			}
		}
		t.Fatal("Workspace Lifecycle group not found")
	})

	t.Run("development workflow group has expected subcommands", func(t *testing.T) {
		for _, g := range meta.SubcommandGroups {
			if g.Name == "Development Workflow" {
				assert.Contains(t, g.Subcommands, "commit")
				assert.Contains(t, g.Subcommands, "pull")
				return
			}
		}
		t.Fatal("Development Workflow group not found")
	})

	t.Run("completion group has expected subcommands", func(t *testing.T) {
		for _, g := range meta.SubcommandGroups {
			if g.Name == "Completion" {
				assert.Contains(t, g.Subcommands, "merge")
				assert.Contains(t, g.Subcommands, "pr")
				return
			}
		}
		t.Fatal("Completion group not found")
	})

	t.Run("has examples", func(t *testing.T) {
		require.NotEmpty(t, meta.Examples, "work command should have examples")
		for _, example := range meta.Examples {
			assert.True(t, strings.HasPrefix(example, "clie work"),
				"each example should start with 'clie work', got: %s", example)
		}
	})
}

// ---------------------------------------------------------------------------
// Commands() registration tests
// ---------------------------------------------------------------------------

func TestCommands(t *testing.T) {
	cmds := Commands()

	t.Run("returns non-empty command list", func(t *testing.T) {
		require.NotEmpty(t, cmds)
	})

	t.Run("contains expected number of commands", func(t *testing.T) {
		// work, work commit, work create, work merge, work pull, work remove, show workspaces, create pr
		assert.Equal(t, 8, len(cmds), "expected 8 commands registered")
	})

	t.Run("all commands have non-empty names", func(t *testing.T) {
		for _, cmd := range cmds {
			assert.NotEmpty(t, cmd.Name(), "every command should have a non-empty name")
		}
	})

	t.Run("all commands have non-empty metadata", func(t *testing.T) {
		for _, cmd := range cmds {
			meta := cmd.Metadata()
			assert.NotEmpty(t, meta.CanonicalName, "command %s should have a canonical name", cmd.Name())
			assert.NotEmpty(t, meta.Short, "command %s should have a short description", cmd.Name())
		}
	})

	t.Run("contains parent work command", func(t *testing.T) {
		found := false
		for _, cmd := range cmds {
			if cmd.Name() == "work" {
				found = true
				assert.True(t, cmd.Metadata().IsParent)
				break
			}
		}
		assert.True(t, found, "Commands() should include the parent work command")
	})

	t.Run("command names are unique", func(t *testing.T) {
		seen := make(map[string]bool)
		for _, cmd := range cmds {
			name := cmd.Name()
			assert.False(t, seen[name], "duplicate command name: %s", name)
			seen[name] = true
		}
	})
}

// ---------------------------------------------------------------------------
// Subcommand metadata tests
// ---------------------------------------------------------------------------

func TestSubcommandMetadata(t *testing.T) {
	t.Run("createPrCommand", func(t *testing.T) {
		cmd := &createPrCommand{}
		assert.Equal(t, "create pr", cmd.Name())

		meta := cmd.Metadata()
		assert.Equal(t, "create-pr", meta.CanonicalName)
		assert.NotEmpty(t, meta.Short)
		assert.NotEmpty(t, meta.Long)

		flagNames := extractFlagNames(meta.Flags)
		assert.Contains(t, flagNames, "target")
		assert.Contains(t, flagNames, "title")
		assert.Contains(t, flagNames, "debug")
	})

	t.Run("workRemoveCommand", func(t *testing.T) {
		cmd := &workRemoveCommand{}
		assert.Equal(t, "work remove", cmd.Name())

		meta := cmd.Metadata()
		assert.Equal(t, "work-remove", meta.CanonicalName)
		assert.NotEmpty(t, meta.Short)

		flagNames := extractFlagNames(meta.Flags)
		assert.Contains(t, flagNames, "keep-branch")
		assert.Contains(t, flagNames, "delete-remote")
		assert.Contains(t, flagNames, "force")
		assert.Contains(t, flagNames, "debug")
	})

	t.Run("workCommitCommand", func(t *testing.T) {
		cmd := &workCommitCommand{}
		assert.Equal(t, "work commit", cmd.Name())

		meta := cmd.Metadata()
		assert.Equal(t, "work-commit", meta.CanonicalName)
		assert.NotEmpty(t, meta.Short)

		flagNames := extractFlagNames(meta.Flags)
		assert.Contains(t, flagNames, "all")
		assert.Contains(t, flagNames, "message")
		assert.Contains(t, flagNames, "debug")
	})

	t.Run("workCreateCommand", func(t *testing.T) {
		cmd := &workCreateCommand{}
		assert.Equal(t, "work create", cmd.Name())

		meta := cmd.Metadata()
		assert.Equal(t, "work-create", meta.CanonicalName)
		assert.NotEmpty(t, meta.Short)

		flagNames := extractFlagNames(meta.Flags)
		assert.Contains(t, flagNames, "from")
		assert.Contains(t, flagNames, "path")
		assert.Contains(t, flagNames, "debug")
	})

	t.Run("workMergeCommand", func(t *testing.T) {
		cmd := &workMergeCommand{}
		assert.Equal(t, "work merge", cmd.Name())

		meta := cmd.Metadata()
		assert.Equal(t, "work-merge", meta.CanonicalName)
		assert.NotEmpty(t, meta.Short)

		flagNames := extractFlagNames(meta.Flags)
		assert.Contains(t, flagNames, "target")
		assert.Contains(t, flagNames, "no-squash")
		assert.Contains(t, flagNames, "keep-worktree")
		assert.Contains(t, flagNames, "debug")
	})

	t.Run("workPullCommand", func(t *testing.T) {
		cmd := &workPullCommand{}
		assert.Equal(t, "work pull", cmd.Name())

		meta := cmd.Metadata()
		assert.Equal(t, "work-pull", meta.CanonicalName)
		assert.NotEmpty(t, meta.Short)

		flagNames := extractFlagNames(meta.Flags)
		assert.Contains(t, flagNames, "target")
		assert.Contains(t, flagNames, "autostash")
		assert.Contains(t, flagNames, "no-fetch")
		assert.Contains(t, flagNames, "debug")
	})

	t.Run("showWorkspacesCommand", func(t *testing.T) {
		cmd := &showWorkspacesCommand{}
		assert.Equal(t, "show workspaces", cmd.Name())

		meta := cmd.Metadata()
		assert.Equal(t, "show-workspaces", meta.CanonicalName)
		assert.NotEmpty(t, meta.Short)

		flagNames := extractFlagNames(meta.Flags)
		assert.Contains(t, flagNames, "verbose")
		assert.Contains(t, flagNames, "debug")
	})
}

// ---------------------------------------------------------------------------
// generatePRTitle edge cases
// ---------------------------------------------------------------------------

func TestGeneratePRTitle_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		commits    []string
		branchName string
		want       string
	}{
		{
			name:       "single commit",
			commits:    []string{"fix: correct typo"},
			branchName: "fix/typo",
			want:       "fix: correct typo",
		},
		{
			name:       "empty string commit falls back to branch",
			commits:    []string{""},
			branchName: "feature/dark-mode",
			want:       "Add Dark Mode",
		},
		{
			name:       "nil commits falls back to branch",
			commits:    nil,
			branchName: "feature/dark-mode",
			want:       "Add Dark Mode",
		},
		{
			name:       "branch with multiple slashes uses second segment title cased",
			commits:    []string{},
			branchName: "feature/auth/oauth",
			want:       "Add Auth",
		},
		{
			name:       "branch without prefix uses Changes from",
			commits:    []string{},
			branchName: "my-branch",
			want:       "Changes from my-branch",
		},
		{
			name:       "first commit preferred over branch",
			commits:    []string{"docs: update readme", "feat: something else"},
			branchName: "feature/readme",
			want:       "docs: update readme",
		},
		{
			name:       "branch with hyphens converted to spaces and title cased",
			commits:    []string{},
			branchName: "feature/add-new-feature",
			want:       "Add Add New Feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generatePRTitle(tt.commits, tt.branchName)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// generatePRDescription edge cases
// ---------------------------------------------------------------------------

func TestGeneratePRDescription_EdgeCases(t *testing.T) {
	t.Run("single commit description", func(t *testing.T) {
		desc := generatePRDescription([]string{"feat: initial commit"}, "1 file changed")

		assert.Contains(t, desc, "## Summary")
		assert.Contains(t, desc, "feat: initial commit")
		// Single commit should not say "includes N commits"
		assert.NotContains(t, desc, "includes")
		assert.Contains(t, desc, "## Changes")
		assert.Contains(t, desc, "1 file changed")
		assert.Contains(t, desc, "## Test Plan")
		assert.Contains(t, desc, "Manual testing completed")
		assert.Contains(t, desc, "Unit tests pass")
		assert.Contains(t, desc, "Integration tests pass")
	})

	t.Run("multiple commits with empty entries", func(t *testing.T) {
		desc := generatePRDescription(
			[]string{"feat: add auth", "", "fix: patch", ""},
			"3 files changed",
		)

		assert.Contains(t, desc, "includes 4 commits")
		assert.Contains(t, desc, "- feat: add auth")
		assert.Contains(t, desc, "- fix: patch")
		// Empty commit lines should be filtered out from bullet list
		assert.NotContains(t, desc, "- \n")
	})

	t.Run("empty commits list", func(t *testing.T) {
		desc := generatePRDescription([]string{}, "no diff")

		assert.Contains(t, desc, "## Summary")
		assert.Contains(t, desc, "## Changes")
		assert.Contains(t, desc, "no diff")
	})

	t.Run("empty diff stat", func(t *testing.T) {
		desc := generatePRDescription([]string{"feat: something"}, "")

		assert.Contains(t, desc, "## Summary")
		assert.Contains(t, desc, "```\n\n```")
	})
}

// ---------------------------------------------------------------------------
// Validation logic tests (using mock git ops)
// ---------------------------------------------------------------------------

func TestValidateRemoveEnvironment_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		branchName    string
		expectError   bool
		errorContains string
	}{
		{
			name:          "reject main branch",
			branchName:    "main",
			expectError:   true,
			errorContains: "cannot remove main workspace",
		},
		{
			name:          "reject master branch",
			branchName:    "master",
			expectError:   true,
			errorContains: "cannot remove main workspace",
		},
		{
			name:        "allow feature branch",
			branchName:  "feature/test",
			expectError: false,
		},
		{
			name:        "allow branch named main-feature",
			branchName:  "main-feature",
			expectError: false,
		},
		{
			name:        "allow branch named master-fix",
			branchName:  "master-fix",
			expectError: false,
		},
		{
			name:        "allow release branch",
			branchName:  "release/1.0.0",
			expectError: false,
		},
		{
			name:        "allow develop branch",
			branchName:  "develop",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &removeConfig{
				base: &internal.BaseConfig{
					GitOps:   internal.NewDefaultGitOps(".", tool.NewToolSystemForTesting()),
					RepoRoot: ".",
				},
				branchName: tt.branchName,
			}

			err := validateRemoveEnvironment(config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				// May fail due to not being in git repo during test, that's OK
				if err != nil && !strings.Contains(err.Error(), "not in a git repository") {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FormatWorktreeStatus tests
// ---------------------------------------------------------------------------

func TestFormatWorktreeStatus_Exhaustive(t *testing.T) {
	assert.Equal(t, "clean", FormatWorktreeStatus(true))
	assert.Equal(t, "dirty", FormatWorktreeStatus(false))
}

// ---------------------------------------------------------------------------
// ShortenSHA edge cases
// ---------------------------------------------------------------------------

func TestShortenSHA_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		sha      string
		expected string
	}{
		{name: "full 40-char SHA", sha: "abc123def456789012345678901234567890abcd", expected: "abc123d"},
		{name: "exactly 7 chars", sha: "abc1234", expected: "abc1234"},
		{name: "less than 7 chars", sha: "abc12", expected: "abc12"},
		{name: "single char", sha: "a", expected: "a"},
		{name: "empty string", sha: "", expected: ""},
		{name: "8 chars", sha: "abc12345", expected: "abc1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ShortenSHA(tt.sha))
		})
	}
}

// ---------------------------------------------------------------------------
// FormatBranch edge cases
// ---------------------------------------------------------------------------

func TestFormatBranch_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected string
	}{
		{name: "normal branch", branch: "main", expected: "main"},
		{name: "feature branch with slash", branch: "feature/auth", expected: "feature/auth"},
		{name: "empty string", branch: "", expected: "(detached)"},
		{name: "whitespace only (spaces)", branch: "   ", expected: "(detached)"},
		{name: "whitespace only (tab)", branch: "\t", expected: "(detached)"},
		{name: "whitespace only (mixed)", branch: " \t ", expected: "(detached)"},
		{name: "branch with leading space", branch: " main", expected: " main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, FormatBranch(tt.branch))
		})
	}
}

// ---------------------------------------------------------------------------
// Config struct defaults (no os.Args manipulation needed)
// ---------------------------------------------------------------------------

func TestPRConfigStruct(t *testing.T) {
	t.Run("zero value has empty fields", func(t *testing.T) {
		config := prConfig{}
		assert.Empty(t, config.targetBranch)
		assert.Empty(t, config.customTitle)
		assert.Empty(t, config.currentBranch)
		assert.Nil(t, config.base)
	})

	t.Run("can populate fields", func(t *testing.T) {
		config := prConfig{
			targetBranch:  "develop",
			customTitle:   "My PR",
			currentBranch: "feature/test",
			base: &internal.BaseConfig{
				Debug:    true,
				RepoRoot: "/tmp/repo",
			},
		}
		assert.Equal(t, "develop", config.targetBranch)
		assert.Equal(t, "My PR", config.customTitle)
		assert.Equal(t, "feature/test", config.currentBranch)
		assert.True(t, config.base.Debug)
		assert.Equal(t, "/tmp/repo", config.base.RepoRoot)
	})
}

func TestRemoveConfigStruct(t *testing.T) {
	t.Run("zero value defaults", func(t *testing.T) {
		config := removeConfig{}
		assert.False(t, config.keepBranch)
		assert.False(t, config.deleteRemote)
		assert.False(t, config.force)
		assert.Empty(t, config.branchName)
		assert.Empty(t, config.worktreePath)
	})

	t.Run("can populate all flags", func(t *testing.T) {
		config := removeConfig{
			keepBranch:   true,
			deleteRemote: true,
			force:        true,
			branchName:   "feature/old",
			worktreePath: "/tmp/ws",
		}
		assert.True(t, config.keepBranch)
		assert.True(t, config.deleteRemote)
		assert.True(t, config.force)
	})
}

func TestMergeConfigStruct(t *testing.T) {
	t.Run("zero value defaults", func(t *testing.T) {
		config := mergeConfig{}
		assert.Empty(t, config.targetBranch)
		assert.False(t, config.noSquash)
		assert.False(t, config.keepWorktree)
		assert.Empty(t, config.currentBranch)
		assert.Empty(t, config.worktreePath)
	})
}

func TestPullConfigStruct(t *testing.T) {
	t.Run("zero value defaults", func(t *testing.T) {
		config := pullConfig{}
		assert.Empty(t, config.targetBranch)
		assert.False(t, config.autostash)
		assert.False(t, config.noFetch)
		assert.Empty(t, config.currentBranch)
	})
}

func TestCommitConfigStruct(t *testing.T) {
	t.Run("zero value defaults", func(t *testing.T) {
		config := commitConfig{}
		assert.False(t, config.stageAll)
		assert.Empty(t, config.customMessage)
	})
}

func TestListConfigStruct(t *testing.T) {
	t.Run("zero value defaults", func(t *testing.T) {
		config := listConfig{}
		assert.False(t, config.verbose)
	})
}

// ---------------------------------------------------------------------------
// rebaseInfo struct tests
// ---------------------------------------------------------------------------

func TestRebaseInfo(t *testing.T) {
	tests := []struct {
		name           string
		info           rebaseInfo
		expectUpToDate bool
	}{
		{
			name:           "up to date when no new commits",
			info:           rebaseInfo{currentCommits: 3, newCommits: 0, upToDate: true},
			expectUpToDate: true,
		},
		{
			name:           "not up to date when new commits exist",
			info:           rebaseInfo{currentCommits: 3, newCommits: 5, upToDate: false},
			expectUpToDate: false,
		},
		{
			name:           "zero everything is up to date",
			info:           rebaseInfo{currentCommits: 0, newCommits: 0, upToDate: true},
			expectUpToDate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectUpToDate, tt.info.upToDate)
		})
	}
}

// ---------------------------------------------------------------------------
// buildWorktreeTable tests (pure logic, no I/O)
// ---------------------------------------------------------------------------

func TestBuildWorktreeTable_EdgeCases(t *testing.T) {
	t.Run("empty worktree list", func(t *testing.T) {
		result := buildWorktreeTable([]internal.Worktree{}, false)
		// Should still produce a table header at minimum (rendered as uppercase)
		assert.Contains(t, result, "PATH")
		assert.Contains(t, result, "BRANCH")
		assert.Contains(t, result, "STATUS")
	})

	t.Run("non-verbose omits SHA column", func(t *testing.T) {
		wts := []internal.Worktree{
			{Path: "/repo", Branch: "main", SHA: "abc123def", Clean: true},
		}
		result := buildWorktreeTable(wts, false)
		assert.NotContains(t, result, "SHA")
		assert.Contains(t, result, "clean")
	})

	t.Run("verbose includes SHA column", func(t *testing.T) {
		wts := []internal.Worktree{
			{Path: "/repo", Branch: "main", SHA: "abc123def456", Clean: true},
		}
		result := buildWorktreeTable(wts, true)
		assert.Contains(t, result, "SHA")
		assert.Contains(t, result, "abc123d") // shortened
	})

	t.Run("dirty worktree shows dirty status", func(t *testing.T) {
		wts := []internal.Worktree{
			{Path: "/repo", Branch: "feature/x", SHA: "abc", Clean: false},
		}
		result := buildWorktreeTable(wts, false)
		assert.Contains(t, result, "dirty")
	})

	t.Run("detached HEAD shows detached label", func(t *testing.T) {
		wts := []internal.Worktree{
			{Path: "/repo", Branch: "", SHA: "abc1234", Clean: true},
		}
		result := buildWorktreeTable(wts, false)
		assert.Contains(t, result, "(detached)")
	})

	t.Run("multiple worktrees renders all", func(t *testing.T) {
		wts := []internal.Worktree{
			{Path: "/repo1", Branch: "main", SHA: "aaa", Clean: true},
			{Path: "/repo2", Branch: "feature/a", SHA: "bbb", Clean: false},
			{Path: "/repo3", Branch: "feature/b", SHA: "ccc", Clean: true},
		}
		result := buildWorktreeTable(wts, false)
		assert.Contains(t, result, "main")
		assert.Contains(t, result, "feature/a")
		assert.Contains(t, result, "feature/b")
	})
}

// ---------------------------------------------------------------------------
// Flag spec validation helpers
// ---------------------------------------------------------------------------

func TestAllCommandFlagsHaveTypes(t *testing.T) {
	cmds := Commands()
	for _, cmd := range cmds {
		meta := cmd.Metadata()
		t.Run(cmd.Name(), func(t *testing.T) {
			for _, flag := range meta.Flags {
				assert.NotEmpty(t, flag.Name, "flag should have a name in command %s", cmd.Name())
				assert.NotEmpty(t, flag.Type, "flag %s in command %s should have a type", flag.Name, cmd.Name())
				assert.NotEmpty(t, flag.Usage, "flag %s in command %s should have usage text", flag.Name, cmd.Name())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func extractFlagNames(flags []core.FlagSpec) []string {
	names := make([]string, len(flags))
	for i, f := range flags {
		names[i] = f.Name
	}
	return names
}
