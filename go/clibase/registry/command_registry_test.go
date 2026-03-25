package registry

import (
	"fmt"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time check: CommandRegistry implements core.CommandRegistryPort.
var _ core.CommandRegistryPort = (*CommandRegistry)(nil)

// mockCommand implements core.CommandPort for testing.
type mockCommand struct {
	name     string
	metadata core.CommandMetadata
}

func (c *mockCommand) Name() string                  { return c.name }
func (c *mockCommand) Metadata() core.CommandMetadata { return c.metadata }

func newMockCommand(name, canonical, short string) *mockCommand {
	return &mockCommand{
		name: name,
		metadata: core.CommandMetadata{
			CanonicalName: canonical,
			Short:         short,
		},
	}
}

func newMockParentCommand(name, canonical, short string) *mockCommand {
	return &mockCommand{
		name: name,
		metadata: core.CommandMetadata{
			CanonicalName: canonical,
			Short:         short,
			IsParent:      true,
		},
	}
}

func TestCommandRegistry_RegisterAndGet(t *testing.T) {
	reg := NewCommandRegistry()
	cmd := newMockCommand("get modules", "get-modules", "Get all modules")

	err := reg.Register(cmd)
	require.NoError(t, err)

	got, ok := reg.Get("get modules")
	require.True(t, ok, "expected Get to find registered command")
	assert.Equal(t, "get modules", got.Name())
	assert.Equal(t, "get-modules", got.Metadata().CanonicalName)
	assert.Equal(t, "Get all modules", got.Metadata().Short)
}

func TestCommandRegistry_RegisterDuplicate(t *testing.T) {
	reg := NewCommandRegistry()
	cmd1 := newMockCommand("build", "build", "Build the project")
	cmd2 := newMockCommand("build", "build", "Build again")

	err := reg.Register(cmd1)
	require.NoError(t, err)

	err = reg.Register(cmd2)
	require.Error(t, err)

	var dupErr *ErrDuplicateCommand
	require.ErrorAs(t, err, &dupErr)
	assert.Equal(t, "build", dupErr.Name)
	assert.Equal(t, `duplicate command definition: "build"`, dupErr.Error())
}

func TestCommandRegistry_MustRegisterPanicsOnDuplicate(t *testing.T) {
	reg := NewCommandRegistry()
	cmd1 := newMockCommand("test", "test", "Run tests")
	cmd2 := newMockCommand("test", "test", "Run tests again")

	reg.MustRegister(cmd1) // should not panic

	assert.Panics(t, func() {
		reg.MustRegister(cmd2)
	})
}

func TestCommandRegistry_MustRegisterNoPanicOnFirst(t *testing.T) {
	reg := NewCommandRegistry()
	cmd := newMockCommand("lint", "lint", "Lint code")

	assert.NotPanics(t, func() {
		reg.MustRegister(cmd)
	})

	got, ok := reg.Get("lint")
	require.True(t, ok)
	assert.Equal(t, "lint", got.Name())
}

func TestCommandRegistry_RegisterAll(t *testing.T) {
	reg := NewCommandRegistry()
	cmds := []core.CommandPort{
		newMockCommand("get modules", "get-modules", "Get modules"),
		newMockCommand("get files", "get-files", "Get files"),
		newMockCommand("get dependencies", "get-dependencies", "Get deps"),
	}

	err := reg.RegisterAll(cmds...)
	require.NoError(t, err)

	for _, cmd := range cmds {
		got, ok := reg.Get(cmd.Name())
		require.True(t, ok, "expected to find %q", cmd.Name())
		assert.Equal(t, cmd.Name(), got.Name())
	}
}

func TestCommandRegistry_RegisterAllStopsOnError(t *testing.T) {
	reg := NewCommandRegistry()
	cmds := []core.CommandPort{
		newMockCommand("alpha", "alpha", "Alpha"),
		newMockCommand("beta", "beta", "Beta"),
		newMockCommand("alpha", "alpha", "Alpha duplicate"), // duplicate
		newMockCommand("gamma", "gamma", "Gamma"),
	}

	err := reg.RegisterAll(cmds...)
	require.Error(t, err)

	var dupErr *ErrDuplicateCommand
	require.ErrorAs(t, err, &dupErr)
	assert.Equal(t, "alpha", dupErr.Name)

	// "alpha" and "beta" should be registered (before the error)
	_, ok := reg.Get("alpha")
	assert.True(t, ok, "alpha should be registered before the duplicate")
	_, ok = reg.Get("beta")
	assert.True(t, ok, "beta should be registered before the duplicate")

	// "gamma" should NOT be registered (after the error)
	_, ok = reg.Get("gamma")
	assert.False(t, ok, "gamma should not be registered after the error")
}

func TestCommandRegistry_GetNonExistent(t *testing.T) {
	reg := NewCommandRegistry()

	got, ok := reg.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestCommandRegistry_GetByCanonical(t *testing.T) {
	reg := NewCommandRegistry()
	cmd := newMockCommand("get modules", "get-modules", "Get all modules")

	err := reg.Register(cmd)
	require.NoError(t, err)

	got, ok := reg.GetByCanonical("get-modules")
	require.True(t, ok, "expected GetByCanonical to find command by kebab-case name")
	assert.Equal(t, "get modules", got.Name())
	assert.Equal(t, "get-modules", got.Metadata().CanonicalName)
}

func TestCommandRegistry_GetByCanonicalNonExistent(t *testing.T) {
	reg := NewCommandRegistry()

	got, ok := reg.GetByCanonical("does-not-exist")
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestCommandRegistry_GetByCanonicalSingleWord(t *testing.T) {
	reg := NewCommandRegistry()
	cmd := newMockCommand("build", "build", "Build the project")

	err := reg.Register(cmd)
	require.NoError(t, err)

	got, ok := reg.GetByCanonical("build")
	require.True(t, ok)
	assert.Equal(t, "build", got.Name())
}

func TestCommandRegistry_All(t *testing.T) {
	reg := NewCommandRegistry()
	cmds := []core.CommandPort{
		newMockCommand("get", "get", "Get things"),
		newMockCommand("show", "show", "Show things"),
		newMockCommand("build", "build", "Build things"),
	}

	err := reg.RegisterAll(cmds...)
	require.NoError(t, err)

	all := reg.All()
	assert.Len(t, all, 3)

	// Verify all registered commands are present
	names := make(map[string]bool)
	for _, cmd := range all {
		names[cmd.Name()] = true
	}
	assert.True(t, names["get"])
	assert.True(t, names["show"])
	assert.True(t, names["build"])
}

func TestCommandRegistry_AllEmpty(t *testing.T) {
	reg := NewCommandRegistry()

	all := reg.All()
	assert.Empty(t, all)
}

func TestCommandRegistry_Names(t *testing.T) {
	reg := NewCommandRegistry()
	// Register in non-alphabetical order to verify sorting
	cmds := []core.CommandPort{
		newMockCommand("show modules", "show-modules", "Show modules"),
		newMockCommand("get files", "get-files", "Get files"),
		newMockCommand("build", "build", "Build"),
		newMockCommand("get modules", "get-modules", "Get modules"),
	}

	err := reg.RegisterAll(cmds...)
	require.NoError(t, err)

	names := reg.Names()
	expected := []string{"build", "get files", "get modules", "show modules"}
	assert.Equal(t, expected, names, "Names() should return sorted command names")
}

func TestCommandRegistry_NamesEmpty(t *testing.T) {
	reg := NewCommandRegistry()

	names := reg.Names()
	assert.Empty(t, names)
}

func TestCommandRegistry_Subcommands(t *testing.T) {
	reg := NewCommandRegistry()

	cmds := []core.CommandPort{
		newMockParentCommand("get", "get", "Retrieve data"),
		newMockCommand("get modules", "get-modules", "Get modules"),
		newMockCommand("get files", "get-files", "Get files"),
		newMockCommand("get dependencies", "get-dependencies", "Get deps"),
		newMockParentCommand("show", "show", "Show information"),
		newMockCommand("show modules", "show-modules", "Show modules"),
	}

	err := reg.RegisterAll(cmds...)
	require.NoError(t, err)

	subs := reg.Subcommands("get")
	assert.Len(t, subs, 3, "expected 3 direct children of 'get'")

	subNames := make([]string, len(subs))
	for i, s := range subs {
		subNames[i] = s.Name()
	}
	assert.Contains(t, subNames, "get modules")
	assert.Contains(t, subNames, "get files")
	assert.Contains(t, subNames, "get dependencies")
	// Must not include commands from other parents
	assert.NotContains(t, subNames, "show modules")
}

func TestCommandRegistry_SubcommandsExcludesNested(t *testing.T) {
	reg := NewCommandRegistry()

	cmds := []core.CommandPort{
		newMockParentCommand("get", "get", "Retrieve data"),
		newMockCommand("get files", "get-files", "Get files"),
		newMockCommand("get files tree", "get-files-tree", "Get files as tree"),
		newMockCommand("get files changed", "get-files-changed", "Get changed files"),
	}

	err := reg.RegisterAll(cmds...)
	require.NoError(t, err)

	subs := reg.Subcommands("get")
	assert.Len(t, subs, 1, "expected only direct children, not nested subcommands")
	assert.Equal(t, "get files", subs[0].Name())
}

func TestCommandRegistry_SubcommandsNonExistentParent(t *testing.T) {
	reg := NewCommandRegistry()

	subs := reg.Subcommands("nonexistent")
	assert.Empty(t, subs, "expected empty slice for non-existent parent")
}

func TestCommandRegistry_SubcommandsExcludesParentItself(t *testing.T) {
	reg := NewCommandRegistry()

	cmds := []core.CommandPort{
		newMockParentCommand("get", "get", "Retrieve data"),
		newMockCommand("get modules", "get-modules", "Get modules"),
	}

	err := reg.RegisterAll(cmds...)
	require.NoError(t, err)

	subs := reg.Subcommands("get")
	for _, sub := range subs {
		assert.NotEqual(t, "get", sub.Name(), "parent should not appear in its own subcommands")
	}
}

// --- Alias tests ---

func newMockCommandWithAliases(name, canonical, short string, aliases []string) *mockCommand {
	return &mockCommand{
		name: name,
		metadata: core.CommandMetadata{
			CanonicalName: canonical,
			Short:         short,
			Aliases:       aliases,
		},
	}
}

func TestCommandRegistry_RegisterAlias(t *testing.T) {
	reg := NewCommandRegistry()
	cmd := newMockCommandWithAliases("show workspaces", "show-workspaces",
		"List workspaces", []string{"work list"})

	require.NoError(t, reg.Register(cmd))

	// Primary name works
	got, ok := reg.Get("show workspaces")
	require.True(t, ok)
	assert.Equal(t, "show workspaces", got.Name())

	// Alias resolves and returns the same command
	got2, ok := reg.Get("work list")
	require.True(t, ok)
	assert.Equal(t, "show workspaces", got2.Name()) // Name() is always primary
	assert.Same(t, got, got2)                       // same pointer
}

func TestCommandRegistry_AliasesExcludedFromAll(t *testing.T) {
	reg := NewCommandRegistry()
	cmd := newMockCommandWithAliases("show workspaces", "show-workspaces",
		"List workspaces", []string{"work list"})
	require.NoError(t, reg.Register(cmd))

	all := reg.All()
	assert.Len(t, all, 1, "alias must not appear as a separate entry in All()")
	assert.Equal(t, "show workspaces", all[0].Name())
}

func TestCommandRegistry_AliasesExcludedFromNames(t *testing.T) {
	reg := NewCommandRegistry()
	cmd := newMockCommandWithAliases("show workspaces", "show-workspaces",
		"List workspaces", []string{"work list"})
	require.NoError(t, reg.Register(cmd))

	names := reg.Names()
	assert.Equal(t, []string{"show workspaces"}, names)
	assert.NotContains(t, names, "work list")
}

func TestCommandRegistry_SubcommandEntries_AliasKey(t *testing.T) {
	reg := NewCommandRegistry()
	parent := newMockParentCommand("work", "work", "Workspace management")
	cmd := newMockCommandWithAliases("show workspaces", "show-workspaces",
		"List workspaces", []string{"work list"})
	require.NoError(t, reg.RegisterAll(parent, cmd))

	entries := reg.SubcommandEntries("work")
	require.Len(t, entries, 1)
	assert.Equal(t, "work list", entries[0].Key,
		"entry key must be the alias key, not the primary name")
	assert.Equal(t, "show workspaces", entries[0].Cmd.Name(),
		"entry Cmd.Name() is still the primary name")
}

func TestCommandRegistry_SubcommandEntries_PrimaryKey(t *testing.T) {
	reg := NewCommandRegistry()
	parent := newMockParentCommand("show", "show", "Show information")
	cmd := newMockCommandWithAliases("show workspaces", "show-workspaces",
		"List workspaces", []string{"work list"})
	require.NoError(t, reg.RegisterAll(parent, cmd))

	entries := reg.SubcommandEntries("show")
	require.Len(t, entries, 1)
	assert.Equal(t, "show workspaces", entries[0].Key,
		"primary key is returned when looking up under 'show'")
}

func TestCommandRegistry_SubcommandEntries_NoDuplicates(t *testing.T) {
	reg := NewCommandRegistry()
	parent1 := newMockParentCommand("show", "show", "Show things")
	parent2 := newMockParentCommand("work", "work", "Work things")
	cmd := newMockCommandWithAliases("show workspaces", "show-workspaces",
		"List workspaces", []string{"work list"})
	require.NoError(t, reg.RegisterAll(parent1, parent2, cmd))

	showSubs := reg.SubcommandEntries("show")
	workSubs := reg.SubcommandEntries("work")
	assert.Len(t, showSubs, 1, "show has one child: 'show workspaces'")
	assert.Len(t, workSubs, 1, "work has one child: 'work list' (alias)")
	// Both entries resolve to the same command instance
	assert.Same(t, showSubs[0].Cmd, workSubs[0].Cmd)
}

func TestCommandRegistry_DuplicateAliasRejected(t *testing.T) {
	reg := NewCommandRegistry()
	cmd1 := newMockCommandWithAliases("show workspaces", "show-workspaces",
		"List workspaces", []string{"work list"})
	cmd2 := newMockCommandWithAliases("other command", "other-command",
		"Another command", []string{"work list"}) // collision

	require.NoError(t, reg.Register(cmd1))
	err := reg.Register(cmd2)
	require.Error(t, err)

	var dupErr *ErrDuplicateCommand
	require.ErrorAs(t, err, &dupErr)
	assert.Equal(t, "work list", dupErr.Name)
}

func TestCommandRegistry_AliasSameAsPrimaryRejected(t *testing.T) {
	reg := NewCommandRegistry()
	cmd := newMockCommandWithAliases("show workspaces", "show-workspaces",
		"List workspaces", []string{"show workspaces"}) // alias = primary

	err := reg.Register(cmd)
	require.Error(t, err, "alias that equals primary name must be rejected")
}

func TestCommandRegistry_AliasResolvedByGetByCanonical(t *testing.T) {
	reg := NewCommandRegistry()
	cmd := newMockCommandWithAliases("show workspaces", "show-workspaces",
		"List workspaces", []string{"work list"})
	require.NoError(t, reg.Register(cmd))

	// "work-list" → spaces → "work list" → found via alias
	got, ok := reg.GetByCanonical("work-list")
	require.True(t, ok)
	assert.Equal(t, "show workspaces", got.Name())
}

func TestErrDuplicateCommand_Error(t *testing.T) {
	err := &ErrDuplicateCommand{Name: "build"}
	assert.Equal(t, `duplicate command definition: "build"`, err.Error())
}

func TestErrDuplicateCommand_ErrorFormat(t *testing.T) {
	tests := []struct {
		name     string
		cmdName  string
		expected string
	}{
		{
			name:     "simple command",
			cmdName:  "build",
			expected: `duplicate command definition: "build"`,
		},
		{
			name:     "two-word command",
			cmdName:  "get modules",
			expected: `duplicate command definition: "get modules"`,
		},
		{
			name:     "three-word command",
			cmdName:  "get files tree",
			expected: `duplicate command definition: "get files tree"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ErrDuplicateCommand{Name: tt.cmdName}
			assert.Equal(t, tt.expected, err.Error())
			assert.Equal(t, tt.expected, fmt.Sprintf("%s", err))
		})
	}
}
