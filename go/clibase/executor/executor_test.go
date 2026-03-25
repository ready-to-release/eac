package executor

import (
	"context"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test helpers: mocks for core.CommandRegistryPort, SimpleCommandPort, CommandPort
// ---------------------------------------------------------------------------

// mockRegistry implements core.CommandRegistryPort.
// Only Get is wired; the remaining methods return zero values.
type mockRegistry struct {
	commands map[string]core.CommandPort
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{commands: make(map[string]core.CommandPort)}
}

func (r *mockRegistry) Get(name string) (core.CommandPort, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

func (r *mockRegistry) GetByCanonical(_ string) (core.CommandPort, bool) { return nil, false }
func (r *mockRegistry) All() []core.CommandPort                          { return nil }
func (r *mockRegistry) Names() []string                                  { return nil }
func (r *mockRegistry) Subcommands(_ string) []core.CommandPort                { return nil }
func (r *mockRegistry) SubcommandEntries(_ string) []core.SubcommandEntry { return nil }

// mockSimpleCommand implements core.SimpleCommandPort (executable command).
type mockSimpleCommand struct {
	name     string
	meta     core.CommandMetadata
	exitCode int

	// Captured values from the last Execute call.
	calledWith *core.CommandRequest
	calledCtx  context.Context
}

func (c *mockSimpleCommand) Name() string              { return c.name }
func (c *mockSimpleCommand) Metadata() core.CommandMetadata { return c.meta }
func (c *mockSimpleCommand) Execute(ctx context.Context, req *core.CommandRequest) int {
	c.calledCtx = ctx
	c.calledWith = req
	return c.exitCode
}

// mockParentCommand implements only core.CommandPort (not executable).
type mockParentCommand struct {
	name string
	meta core.CommandMetadata
}

func (c *mockParentCommand) Name() string              { return c.name }
func (c *mockParentCommand) Metadata() core.CommandMetadata { return c.meta }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestExecute_DispatchesToSimpleCommandAndReturnsExitCode(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
	}{
		{name: "returns zero on success", exitCode: 0},
		{name: "returns non-zero on failure", exitCode: 2},
		{name: "returns 42", exitCode: 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := newMockRegistry()
			cmd := &mockSimpleCommand{name: "build", exitCode: tt.exitCode}
			reg.commands["build"] = cmd

			exec := New(reg)
			got := exec.Execute(context.Background(), "build", []string{"--verbose"})

			assert.Equal(t, tt.exitCode, got)
			require.NotNil(t, cmd.calledWith, "command should have been called")
		})
	}
}

func TestExecute_UnknownCommandReturnsOne(t *testing.T) {
	reg := newMockRegistry()
	exec := New(reg)

	got := exec.Execute(context.Background(), "no-such-command", nil)

	assert.Equal(t, 1, got)
}

func TestExecute_ParentOnlyCommandReturnsOne(t *testing.T) {
	reg := newMockRegistry()
	parent := &mockParentCommand{
		name: "get",
		meta: core.CommandMetadata{CanonicalName: "get", IsParent: true},
	}
	reg.commands["get"] = parent

	exec := New(reg)
	got := exec.Execute(context.Background(), "get", nil)

	assert.Equal(t, 1, got)
}

func TestExecute_PassesContextToCommand(t *testing.T) {
	type ctxKey string
	key := ctxKey("test-key")

	reg := newMockRegistry()
	cmd := &mockSimpleCommand{name: "deploy", exitCode: 0}
	reg.commands["deploy"] = cmd

	ctx := context.WithValue(context.Background(), key, "test-value")

	exec := New(reg)
	_ = exec.Execute(ctx, "deploy", nil)

	require.NotNil(t, cmd.calledCtx, "context should have been forwarded")
	assert.Equal(t, "test-value", cmd.calledCtx.Value(key))
}

func TestExecute_PassesArgsToCommandRequest(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "single arg", args: []string{"--module=core"}},
		{name: "multiple args", args: []string{"--format", "json", "--verbose"}},
		{name: "no args", args: nil},
		{name: "empty args", args: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := newMockRegistry()
			cmd := &mockSimpleCommand{name: "info", exitCode: 0}
			reg.commands["info"] = cmd

			exec := New(reg)
			_ = exec.Execute(context.Background(), "info", tt.args)

			require.NotNil(t, cmd.calledWith, "command should have been called")
			assert.Equal(t, tt.args, cmd.calledWith.Args)
		})
	}
}

func TestRegistry_ReturnsInjectedRegistry(t *testing.T) {
	reg := newMockRegistry()
	exec := New(reg)

	assert.Same(t, reg, exec.Registry(), "Registry() should return the same instance passed to New")
}
