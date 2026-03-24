//go:build L1 && ov
// +build L1,ov

// Feature: cli_command_routing
// Unit tests for command routing and registration
package main

import (
	"strings"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/stretchr/testify/assert"
)

func TestCommandRegistryExists(t *testing.T) {
	reg := registry.Global()
	if reg == nil {
		t.Fatal("global registry should not be nil")
	}
}

func TestCommandsRegistered(t *testing.T) {
	reg := registry.Global()
	if reg == nil {
		t.Fatal("global registry is nil")
	}
	expectedCommands := []string{
		"show modules",
		"show files",
		"show files-changed",
		"show files-staged",
	}

	for _, cmdName := range expectedCommands {
		if _, ok := reg.Get(cmdName); !ok {
			t.Errorf("expected command '%s' to be registered", cmdName)
		}
	}
}

func TestGetSubcommands(t *testing.T) {
	reg := registry.Global()
	if reg == nil {
		t.Fatal("global registry is nil")
	}

	// Test that "show" parent returns subcommands
	subs := reg.Subcommands("show")

	if len(subs) == 0 {
		t.Error("expected 'show' to have subcommands")
	}

	// Check for expected subcommands
	expectedSubs := []string{"files", "modules", "dependencies"}
	for _, expected := range expectedSubs {
		found := false
		for _, sub := range subs {
			name := strings.TrimPrefix(sub.Name(), "show ")
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected 'show' to have subcommand '%s'", expected)
		}
	}
}

func TestGetSubcommandsReturnsEmpty(t *testing.T) {
	reg := registry.Global()
	if reg == nil {
		t.Fatal("global registry is nil")
	}

	// Leaf commands should have no subcommands
	subs := reg.Subcommands("show modules")

	if len(subs) != 0 {
		t.Errorf("expected 'show modules' to have no subcommands, got %d", len(subs))
	}
}

func TestGetSubcommandsSorted(t *testing.T) {
	reg := registry.Global()
	if reg == nil {
		t.Fatal("global registry is nil")
	}

	subs := reg.Subcommands("show")

	// Verify alphabetical sorting
	for i := 1; i < len(subs); i++ {
		if subs[i-1].Name() > subs[i].Name() {
			t.Errorf("subcommands not sorted: '%s' should come before '%s'",
				subs[i-1].Name(), subs[i].Name())
		}
	}
}

// --- Test stubs for printCommandHelp ---

type stubCmd struct {
	name string
	meta core.CommandMetadata
}

func (s *stubCmd) Name() string                  { return s.name }
func (s *stubCmd) Metadata() core.CommandMetadata { return s.meta }

type stubReg struct {
	commands map[string]core.CommandPort
}

func (r *stubReg) Get(name string) (core.CommandPort, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

func (r *stubReg) GetByCanonical(string) (core.CommandPort, bool) { return nil, false }

func (r *stubReg) All() []core.CommandPort {
	var cmds []core.CommandPort
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (r *stubReg) Names() []string {
	var names []string
	for name := range r.commands {
		names = append(names, name)
	}
	return names
}

func (r *stubReg) Subcommands(parentName string) []core.CommandPort {
	var subs []core.CommandPort
	prefix := parentName + " "
	for name, cmd := range r.commands {
		if strings.HasPrefix(name, prefix) && !strings.Contains(name[len(prefix):], " ") {
			subs = append(subs, cmd)
		}
	}
	return subs
}

func TestPrintCommandHelp_Nil(t *testing.T) {
	var buf strings.Builder
	printCommandHelp(&buf, nil, &stubReg{commands: map[string]core.CommandPort{}})
	assert.Equal(t, "No help available.\n", buf.String())
}

func TestPrintCommandHelp_Sections(t *testing.T) {
	cmd := &stubCmd{
		name: "build",
		meta: core.CommandMetadata{
			Short: "Build modules",
			Long:  "Build one or more modules by moniker.",
			Notes: "Logs written to out/build/",
			Flags: []core.FlagSpec{
				{Name: "dry-run", Type: "bool", Usage: "Show what would be built"},
			},
			Examples: []string{"eac build", "eac build --dry-run"},
		},
	}
	reg := &stubReg{commands: map[string]core.CommandPort{"build": cmd}}

	var buf strings.Builder
	printCommandHelp(&buf, cmd, reg)
	out := buf.String()

	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "build - Build modules")
	assert.Contains(t, out, "SYNOPSIS")
	assert.Contains(t, out, "DESCRIPTION")
	assert.Contains(t, out, "Build one or more modules by moniker.")
	assert.Contains(t, out, "NOTES")
	assert.Contains(t, out, "Logs written to out/build/")
	assert.Contains(t, out, "FLAGS")
	assert.Contains(t, out, "--dry-run")
	assert.Contains(t, out, "EXAMPLES")
	assert.Contains(t, out, "eac build")
}

func TestPrintCommandHelp_FlatSubcommands(t *testing.T) {
	parent := &stubCmd{
		name: "create",
		meta: core.CommandMetadata{
			Short:    "Create project artifacts",
			IsParent: true,
		},
	}
	reg := &stubReg{commands: map[string]core.CommandPort{
		"create":      parent,
		"create pr":   &stubCmd{name: "create pr", meta: core.CommandMetadata{Short: "Create a pull request"}},
		"create spec": &stubCmd{name: "create spec", meta: core.CommandMetadata{Short: "Create specification"}},
	}}

	var buf strings.Builder
	printCommandHelp(&buf, parent, reg)
	out := buf.String()

	assert.Contains(t, out, "COMMANDS")
	assert.Contains(t, out, "pr")
	assert.Contains(t, out, "Create a pull request")
	assert.Contains(t, out, "spec")
	assert.Contains(t, out, "Create specification")
	// No group headers in flat mode
	assert.NotContains(t, out, ":\n        ")
}

func TestPrintCommandHelp_GroupedSubcommands(t *testing.T) {
	cmd := &stubCmd{
		name: "work",
		meta: core.CommandMetadata{
			Short:    "Manage workspaces",
			IsParent: true,
			SubcommandGroups: []core.SubcommandGroup{
				{Name: "Lifecycle", Subcommands: []string{"create", "remove"}},
				{Name: "Workflow", Subcommands: []string{"commit", "pull"}},
			},
		},
	}
	reg := &stubReg{commands: map[string]core.CommandPort{
		"work":        cmd,
		"work create": &stubCmd{name: "work create", meta: core.CommandMetadata{Short: "Create workspace"}},
		"work remove": &stubCmd{name: "work remove", meta: core.CommandMetadata{Short: "Remove workspace"}},
		"work commit": &stubCmd{name: "work commit", meta: core.CommandMetadata{Short: "Commit changes"}},
		"work pull":   &stubCmd{name: "work pull", meta: core.CommandMetadata{Short: "Pull updates"}},
	}}

	var buf strings.Builder
	printCommandHelp(&buf, cmd, reg)
	out := buf.String()

	assert.Contains(t, out, "COMMANDS")
	assert.Contains(t, out, "Lifecycle:")
	assert.Contains(t, out, "Workflow:")
	assert.Contains(t, out, "create")
	assert.Contains(t, out, "Create workspace")
	assert.Contains(t, out, "commit")
	assert.Contains(t, out, "Commit changes")

	// Group header precedes its subcommands
	lifecycleIdx := strings.Index(out, "Lifecycle:")
	createIdx := strings.Index(out, "create")
	assert.Greater(t, createIdx, lifecycleIdx, "create should appear after Lifecycle: header")

	workflowIdx := strings.Index(out, "Workflow:")
	commitIdx := strings.Index(out, "commit")
	assert.Greater(t, commitIdx, workflowIdx, "commit should appear after Workflow: header")
}

func TestPrintCommandHelp_FallbackToFlat(t *testing.T) {
	// Parent command with subcommands but no SubcommandGroups defined
	parent := &stubCmd{
		name: "get",
		meta: core.CommandMetadata{
			Short:    "Get information",
			IsParent: true,
			// SubcommandGroups intentionally empty
		},
	}
	reg := &stubReg{commands: map[string]core.CommandPort{
		"get":        parent,
		"get config": &stubCmd{name: "get config", meta: core.CommandMetadata{Short: "Get configuration"}},
		"get files":  &stubCmd{name: "get files", meta: core.CommandMetadata{Short: "Get file list"}},
	}}

	var buf strings.Builder
	printCommandHelp(&buf, parent, reg)
	out := buf.String()

	assert.Contains(t, out, "COMMANDS")
	assert.Contains(t, out, "config")
	assert.Contains(t, out, "Get configuration")
	// Should NOT contain group header syntax (indented colon)
	assert.NotRegexp(t, `    \w+:$`, out)
}
