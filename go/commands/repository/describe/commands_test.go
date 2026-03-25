package describe

import (
	"slices"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCmd implements core.CommandPort for testing.
type mockCmd struct {
	name     string
	metadata core.CommandMetadata
}

func (c *mockCmd) Name() string                  { return c.name }
func (c *mockCmd) Metadata() core.CommandMetadata { return c.metadata }

func buildTestRegistry(t *testing.T) *registry.CommandRegistry {
	t.Helper()
	reg := registry.NewCommandRegistry()
	cmds := []core.CommandPort{
		&mockCmd{name: "show", metadata: core.CommandMetadata{
			CanonicalName: "show", Short: "Show information", IsParent: true,
		}},
		&mockCmd{name: "show workspaces", metadata: core.CommandMetadata{
			CanonicalName: "show-workspaces",
			Short:         "List all workspaces and their status",
			Aliases:       []string{"work list"},
		}},
		&mockCmd{name: "work", metadata: core.CommandMetadata{
			CanonicalName: "work", Short: "Workspace management", IsParent: true,
		}},
		&mockCmd{name: "work create", metadata: core.CommandMetadata{
			CanonicalName: "work-create", Short: "Create a workspace",
		}},
		&mockCmd{name: "create", metadata: core.CommandMetadata{
			CanonicalName: "create", Short: "Create resources", IsParent: true,
		}},
		&mockCmd{name: "create pr", metadata: core.CommandMetadata{
			CanonicalName: "create-pr",
			Short:         "Create pull request",
			Aliases:       []string{"work pr"},
		}},
	}
	require.NoError(t, reg.RegisterAll(cmds...))
	return reg
}

func TestBuildCommandTree_AliasAppearsInTree(t *testing.T) {
	reg := buildTestRegistry(t)
	tree := buildCommandTreeFrom(reg)

	// "list" and "pr" should be injected into tree["work"] from aliases
	workChildren := tree.Tree["work"]
	assert.Contains(t, workChildren, "list",
		"alias leaf 'list' must be present in tree[\"work\"]")
	assert.Contains(t, workChildren, "pr",
		"alias leaf 'pr' must be present in tree[\"work\"]")
	// Primary child "create" should also be there
	assert.Contains(t, workChildren, "create",
		"primary leaf 'create' must be present in tree[\"work\"]")
}

func TestBuildCommandTree_AliasNotDuplicatedInCommands(t *testing.T) {
	reg := buildTestRegistry(t)
	tree := buildCommandTreeFrom(reg)

	names := make([]string, 0, len(tree.Commands))
	for _, c := range tree.Commands {
		names = append(names, c.Name)
	}
	assert.NotContains(t, names, "work list",
		"alias path must not appear as a command entry")
	assert.NotContains(t, names, "work pr",
		"alias path must not appear as a command entry")
	// Primary names should be present
	assert.Contains(t, names, "show workspaces")
	assert.Contains(t, names, "create pr")
}

func TestBuildCommandTree_PrimaryTreeUnchanged(t *testing.T) {
	reg := buildTestRegistry(t)
	tree := buildCommandTreeFrom(reg)

	assert.Contains(t, tree.Tree["show"], "workspaces",
		"primary 'show workspaces' leaf must remain under 'show'")
	assert.Contains(t, tree.Tree["create"], "pr",
		"primary 'create pr' leaf must remain under 'create'")
}

func TestBuildCommandTree_AliasLeafNotDuplicated(t *testing.T) {
	reg := buildTestRegistry(t)
	tree := buildCommandTreeFrom(reg)

	// "create" is both a primary child of "work" AND could be confused.
	// Count occurrences of each leaf under "work"
	workChildren := tree.Tree["work"]
	listCount := 0
	for _, child := range workChildren {
		if child == "list" {
			listCount++
		}
	}
	assert.Equal(t, 1, listCount, "alias leaf 'list' should appear exactly once under 'work'")
}

func TestBuildCommandTree_TreeSortable(t *testing.T) {
	reg := buildTestRegistry(t)
	tree := buildCommandTreeFrom(reg)

	// Verify tree["work"] can be sorted and contains expected items
	workChildren := make([]string, len(tree.Tree["work"]))
	copy(workChildren, tree.Tree["work"])
	slices.Sort(workChildren)
	assert.Contains(t, workChildren, "create")
	assert.Contains(t, workChildren, "list")
	assert.Contains(t, workChildren, "pr")
}
