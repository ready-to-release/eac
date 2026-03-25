package helprender

import (
	"regexp"
	"strings"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/stretchr/testify/assert"
)

// stubCommand implements core.CommandPort for testing.
type stubCommand struct {
	name string
	meta core.CommandMetadata
}

func (s *stubCommand) Name() string                  { return s.name }
func (s *stubCommand) Metadata() core.CommandMetadata { return s.meta }

// stubRegistry implements core.CommandRegistryPort for testing.
type stubRegistry struct {
	commands map[string]core.CommandPort
}

func (r *stubRegistry) Get(name string) (core.CommandPort, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

func (r *stubRegistry) GetByCanonical(name string) (core.CommandPort, bool) {
	return nil, false
}

func (r *stubRegistry) All() []core.CommandPort {
	var cmds []core.CommandPort
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (r *stubRegistry) Names() []string {
	var names []string
	for name := range r.commands {
		names = append(names, name)
	}
	return names
}

func (r *stubRegistry) Subcommands(parentName string) []core.CommandPort {
	var subs []core.CommandPort
	for name, cmd := range r.commands {
		if strings.HasPrefix(name, parentName+" ") {
			subs = append(subs, cmd)
		}
	}
	return subs
}

func (r *stubRegistry) SubcommandEntries(parentName string) []core.SubcommandEntry {
	var entries []core.SubcommandEntry
	for name, cmd := range r.commands {
		if strings.HasPrefix(name, parentName+" ") {
			entries = append(entries, core.SubcommandEntry{Key: name, Cmd: cmd})
		}
	}
	return entries
}

func TestRenderMarkdownHelp_NilCommand(t *testing.T) {
	result := RenderMarkdownHelp(nil, nil)
	assert.Equal(t, "", result)
}

func TestRenderMarkdownHelp_BasicCommand(t *testing.T) {
	cmd := &stubCommand{
		name: "show modules",
		meta: core.CommandMetadata{
			Short: "Display all module contracts",
			Long:  "This command shows all modules defined in the repository.yml file.",
			Flags: []core.FlagSpec{
				{Name: "format", Type: "string", DefaultValue: "table", Usage: "Output format"},
				{Name: "verbose", Shorthand: "v", Type: "bool", Usage: "Show verbose output"},
			},
			Examples: []string{
				"show modules",
				"show modules --format json",
			},
		},
	}

	result := RenderMarkdownHelp(cmd, nil)

	// Description
	assert.Contains(t, result, "This command shows all modules defined in the repository.yml file.")

	// Usage/Synopsis
	assert.Contains(t, result, "**Usage:** `show modules [flags]`")

	// Flags — no heading, just separator + table
	assert.NotContains(t, result, "### Flags")
	assert.Contains(t, result, "---")
	assert.Contains(t, result, "`--format <string> (optional, default: table)`")
	assert.Contains(t, result, "`-v, --verbose (optional)`")

	// Examples — no heading, just separator + code fence
	assert.NotContains(t, result, "### Examples")
	assert.Contains(t, result, "```bash")
	assert.Contains(t, result, "show modules\nshow modules --format json")
}

func TestRenderMarkdownHelp_WithArgs(t *testing.T) {
	cmd := &stubCommand{
		name: "build",
		meta: core.CommandMetadata{
			Short: "Build modules",
			Long:  "Build one or more modules by moniker.",
			Args:  "modules",
			Flags: []core.FlagSpec{
				{Name: "dry-run", Type: "bool", Usage: "Show what would be built"},
			},
		},
	}

	result := RenderMarkdownHelp(cmd, nil)

	// Synopsis includes args
	assert.Contains(t, result, "**Usage:** `build [flags] <modules>`")

	// Arguments table — no heading
	assert.NotContains(t, result, "### Arguments")
	assert.Contains(t, result, "`modules`")

	// Flags — no heading
	assert.NotContains(t, result, "### Flags")
	assert.Contains(t, result, "`--dry-run (optional)`")
}

func TestRenderMarkdownHelp_WithNotes(t *testing.T) {
	cmd := &stubCommand{
		name: "build",
		meta: core.CommandMetadata{
			Short: "Build modules",
			Long:  "Build one or more modules by moniker.",
			Notes: "Expected Output:\n- Build logs written to 'out/build/<module>/build.log'\n- Build manifest at 'out/build/<module>/<component>/uow.manifest.json'",
		},
	}

	result := RenderMarkdownHelp(cmd, nil)

	// Description preserved
	assert.Contains(t, result, "Build one or more modules by moniker.")

	// Notes content appears verbatim (no **Expected Output:** bold wrapper)
	assert.Contains(t, result, "Expected Output:")
	assert.Contains(t, result, "Build logs written to")

	// Notes appears after a --- separator
	notesIdx := strings.Index(result, "Expected Output:")
	if assert.Greater(t, notesIdx, 0, "Notes content should appear in output") {
		sepIdx := strings.LastIndex(result[:notesIdx], "---")
		assert.Greater(t, sepIdx, 0, "separator should precede Notes")
	}
}

func TestRenderMarkdownHelp_EmptyNotesOmitsSection(t *testing.T) {
	cmd := &stubCommand{
		name: "init",
		meta: core.CommandMetadata{Short: "Initialize"},
	}
	result := RenderMarkdownHelp(cmd, nil)
	// No separators when no sections present
	assert.NotContains(t, result, "---")
}

func TestRenderMarkdownHelp_SeparatorsOnlyWhenDataPresent(t *testing.T) {
	cmd := &stubCommand{
		name: "show modules",
		meta: core.CommandMetadata{
			Short: "Show modules",
			Long:  "Shows all modules.",
			Flags: []core.FlagSpec{{Name: "format", Type: "string"}},
		},
	}
	result := RenderMarkdownHelp(cmd, nil)
	// Exactly one horizontal rule separator: before the flags table
	// (table separator rows also contain "---" so we count "\n---\n" specifically)
	assert.Equal(t, 1, strings.Count(result, "\n---\n"))
}

func TestRenderMarkdownHelp_Jinja2Escaping(t *testing.T) {
	cmd := &stubCommand{
		name: "get config",
		meta: core.CommandMetadata{
			Short: "Get config",
			Long:  "Returns config with {{ variable }} syntax.",
			Flags: []core.FlagSpec{
				{Name: "template", Type: "string", Usage: "Use {{ template }} syntax"},
			},
		},
	}

	result := RenderMarkdownHelp(cmd, nil)

	assert.Contains(t, result, "{ { variable } }")
	assert.Contains(t, result, "{ { template } }")
	assert.NotContains(t, result, "{{ variable }}")
}

func TestRenderMarkdownHelp_RequiredFlag(t *testing.T) {
	cmd := &stubCommand{
		name: "deploy",
		meta: core.CommandMetadata{
			Short: "Deploy service",
			Flags: []core.FlagSpec{
				{Name: "target", Type: "string", Required: true, Usage: "Deploy target"},
			},
		},
	}

	result := RenderMarkdownHelp(cmd, nil)

	assert.Contains(t, result, "`--target <string> (required)`")
}

func TestRenderMarkdownHelp_FallsBackToShort(t *testing.T) {
	cmd := &stubCommand{
		name: "init",
		meta: core.CommandMetadata{
			Short: "Initialize repository",
		},
	}

	result := RenderMarkdownHelp(cmd, nil)
	assert.Contains(t, result, "Initialize repository")
}

func TestRenderMarkdownHelp_WithSubcommands(t *testing.T) {
	parent := &stubCommand{
		name: "create",
		meta: core.CommandMetadata{
			Short:    "Create project artifacts",
			IsParent: true,
		},
	}

	reg := &stubRegistry{
		commands: map[string]core.CommandPort{
			"create": parent,
			"create pr": &stubCommand{
				name: "create pr",
				meta: core.CommandMetadata{Short: "Create a pull request"},
			},
			"create spec": &stubCommand{
				name: "create spec",
				meta: core.CommandMetadata{Short: "Create specification"},
			},
		},
	}

	result := RenderMarkdownHelp(parent, reg)

	// No heading for commands section
	assert.NotContains(t, result, "### Commands")
	// But table content is there
	assert.Contains(t, result, "| `pr`")
	assert.Contains(t, result, "| `spec`")
	// Separator present
	assert.Contains(t, result, "---")
}

func TestRenderMarkdownHelp_PipeEscaping(t *testing.T) {
	cmd := &stubCommand{
		name: "get config",
		meta: core.CommandMetadata{
			Short: "Get config",
			Flags: []core.FlagSpec{
				{Name: "format", Type: "string", Usage: "Output format: json|yaml"},
			},
		},
	}

	result := RenderMarkdownHelp(cmd, nil)

	// Pipe in usage should be escaped within the table
	assert.Contains(t, result, "json\\|yaml")
}

func TestRenderMarkdownHelp_MatchesExpectedSections(t *testing.T) {
	// Comprehensive test matching the plan's expected output format
	cmd := &stubCommand{
		name: "serve docs",
		meta: core.CommandMetadata{
			Short: "Serve documentation locally",
			Long:  "Serves the MkDocs documentation site in a local Docker container.",
			Notes: "Expected Output:\n- Launches MkDocs dev server on localhost",
			Args:  "site",
			Flags: []core.FlagSpec{
				{Name: "port", Type: "number", Usage: "Port to bind the server to"},
				{Name: "open", Type: "bool", Usage: "Open browser after starting"},
			},
			Examples: []string{
				"serve docs",
				"serve docs --port 8080",
				"serve docs --open",
			},
		},
	}

	result := RenderMarkdownHelp(cmd, nil)

	// No title heading
	assert.NotContains(t, result, "## serve docs")

	// Description
	assert.Contains(t, result, "Serves the MkDocs documentation site in a local Docker container.")

	// Usage
	assert.Contains(t, result, "**Usage:** `serve docs [flags] <site>`")

	// Arguments — no heading, table present
	assert.NotContains(t, result, "### Arguments")
	assert.Contains(t, result, "`site`")

	// Flags — no heading, table present
	assert.NotContains(t, result, "### Flags")
	assert.Contains(t, result, "`--port <number> (optional)`")
	assert.Contains(t, result, "`--open (optional)`")

	// Notes — verbatim content
	assert.Contains(t, result, "Expected Output:")

	// Examples — no heading, code fence present
	assert.NotContains(t, result, "### Examples")
	assert.Contains(t, result, "```bash")
	assert.Contains(t, result, "serve docs\nserve docs --port 8080\nserve docs --open")
}

func TestRenderMarkdownHelp_NoHashHeadings(t *testing.T) {
	// Any command with all sections populated should produce zero '#' chars
	// in heading position
	cmd := &stubCommand{
		name: "serve docs",
		meta: core.CommandMetadata{
			Short:    "Serve docs",
			Long:     "Serves docs.",
			Notes:    "Output goes to stdout.",
			Args:     "site",
			Flags:    []core.FlagSpec{{Name: "port", Type: "number"}},
			Examples: []string{"serve docs"},
		},
	}
	result := RenderMarkdownHelp(cmd, nil)
	// No line should start with '#'
	for _, line := range strings.Split(result, "\n") {
		assert.False(t, strings.HasPrefix(line, "#"),
			"line should not start with '#': %q", line)
	}
}

func TestRenderMarkdownHelp_NoStandaloneBold(t *testing.T) {
	// No line should consist entirely of **text** (MD036 pattern)
	cmd := &stubCommand{
		name: "build",
		meta: core.CommandMetadata{
			Short:    "Build modules",
			Long:     "Build one or more modules.",
			Notes:    "Logs written to out/build/",
			Flags:    []core.FlagSpec{{Name: "format", Type: "string"}},
			Examples: []string{"build"},
		},
	}
	result := RenderMarkdownHelp(cmd, nil)
	standaloneBoldRe := regexp.MustCompile(`^\*\*[^*]+\*\*:?$`)
	for _, line := range strings.Split(result, "\n") {
		assert.False(t, standaloneBoldRe.MatchString(strings.TrimSpace(line)),
			"line matches MD036 standalone bold pattern: %q", line)
	}
}

func TestFormatFlagDisplay(t *testing.T) {
	tests := []struct {
		name string
		flag core.FlagSpec
		want string
	}{
		{
			name: "simple bool flag",
			flag: core.FlagSpec{Name: "dry-run", Type: "bool"},
			want: "--dry-run (optional)",
		},
		{
			name: "string flag with default",
			flag: core.FlagSpec{Name: "output", Type: "string", DefaultValue: "json"},
			want: "--output <string> (optional, default: json)",
		},
		{
			name: "required flag",
			flag: core.FlagSpec{Name: "target", Type: "string", Required: true},
			want: "--target <string> (required)",
		},
		{
			name: "flag with shorthand",
			flag: core.FlagSpec{Name: "verbose", Shorthand: "v", Type: "bool"},
			want: "-v, --verbose (optional)",
		},
		{
			name: "int flag required with default",
			flag: core.FlagSpec{Name: "count", Type: "int", Required: true, DefaultValue: "10"},
			want: "--count <int> (required, default: 10)",
		},
		{
			name: "flag with empty type treated as bool",
			flag: core.FlagSpec{Name: "force"},
			want: "--force (optional)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatFlagDisplay(tt.flag)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildSynopsis(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		meta core.CommandMetadata
		want string
	}{
		{
			name: "no flags no args",
			cmd:  "init",
			meta: core.CommandMetadata{},
			want: "init",
		},
		{
			name: "with flags",
			cmd:  "build",
			meta: core.CommandMetadata{
				Flags: []core.FlagSpec{{Name: "dry-run"}},
			},
			want: "build [flags]",
		},
		{
			name: "with flags and args",
			cmd:  "build",
			meta: core.CommandMetadata{
				Flags: []core.FlagSpec{{Name: "dry-run"}},
				Args:  "modules",
			},
			want: "build [flags] <modules>",
		},
		{
			name: "with args only",
			cmd:  "help",
			meta: core.CommandMetadata{Args: "command"},
			want: "help <command>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSynopsis(tt.cmd, tt.meta)
			assert.Equal(t, tt.want, got)
		})
	}
}
