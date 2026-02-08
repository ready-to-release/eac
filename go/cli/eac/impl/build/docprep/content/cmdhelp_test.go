package content

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHelpOutput(t *testing.T) {
	tests := []struct {
		name         string
		cmdName      string
		input        string
		wantDesc     string
		wantUsage    string
		wantNotes    string
		wantExamples string
		wantArgs     int
		wantFlags    int
	}{
		{
			name:    "basic help output",
			cmdName: "show modules",
			input: `show modules - Display all module contracts in a human-readable table

This command shows all modules defined in the repository.yml file.

Usage: show modules [flags]

Flags:
  --format    Output format (table/json)
  --verbose   Show verbose output

Example:
  show modules
  show modules --format json
`,
			wantDesc:     "show modules - Display all module contracts in a human-readable table\n\nThis command shows all modules defined in the repository.yml file.",
			wantUsage:    "show modules [flags]",
			wantExamples: "show modules\n  show modules --format json",
			wantFlags:    2,
		},
		{
			name:    "help with expected output",
			cmdName: "get modules",
			input: `get modules - Get all module contracts in structured format

Returns all module contracts in YAML/JSON format.

Expected Output:
- modules: array of module contracts
- count: total number of modules

Usage: get modules [flags]

Flags:
  --format    Output format (yaml/json)

Example:
  get modules | jq '.modules[]'
`,
			wantDesc:     "get modules - Get all module contracts in structured format\n\nReturns all module contracts in YAML/JSON format.",
			wantUsage:    "get modules [flags]",
			wantNotes:    "**Expected Output:**\n\n- modules: array of module contracts\n- count: total number of modules\n",
			wantExamples: "get modules | jq '.modules[]'",
			wantFlags:    1,
		},
		{
			name:    "help with arguments",
			cmdName: "build",
			input: `build - Build one or more modules by moniker

Builds the specified modules and their dependencies.

Usage: build [module...] [flags]

Arguments:
  module    Module moniker(s) to build

Flags:
  --dry-run     Show what would be built
  --parallel    Build modules in parallel

Example:
  build core
  build core eac-cli
`,
			wantDesc:     "build - Build one or more modules by moniker\n\nBuilds the specified modules and their dependencies.",
			wantUsage:    "build [module...] [flags]",
			wantExamples: "build core\n  build core eac-cli",
			wantArgs:     1,
			wantFlags:    2,
		},
		{
			name:    "help with plural Examples",
			cmdName: "test",
			input: `test - Test one or more modules

Runs tests for specified modules.

Usage: test [module...] [flags]

Flags:
  --suite    Test suite to run

Examples:
  test core
  test core --suite unit
  test core --suite unit+integration
`,
			wantDesc:     "test - Test one or more modules\n\nRuns tests for specified modules.",
			wantUsage:    "test [module...] [flags]",
			wantExamples: "test core\n  test core --suite unit\n  test core --suite unit+integration",
			wantFlags:    1,
		},
		{
			name:    "multi-paragraph description",
			cmdName: "validate",
			input: `validate - Validate repository contracts

This command validates all contracts in the repository against their
JSON schemas. It ensures data integrity and consistency.

Multiple validation types are supported including module contracts,
dependency contracts, and environment domain.

Usage: validate [flags]

Flags:
  --strict    Enable strict validation

Example:
  validate
`,
			wantDesc:     "validate - Validate repository contracts\n\nThis command validates all contracts in the repository against their JSON schemas. It ensures data integrity and consistency.\n\nMultiple validation types are supported including module contracts, dependency contracts, and environment domain.",
			wantUsage:    "validate [flags]",
			wantExamples: "validate",
			wantFlags:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help, err := ParseHelpOutput(tt.cmdName, tt.input)
			require.NoError(t, err)

			assert.Equal(t, tt.cmdName, help.Name)

			if tt.wantDesc != "" {
				assert.Equal(t, tt.wantDesc, help.Description)
			}
			if tt.wantUsage != "" {
				assert.Equal(t, tt.wantUsage, help.Usage)
			}
			if tt.wantNotes != "" {
				assert.Equal(t, tt.wantNotes, help.Notes)
			}
			if tt.wantExamples != "" {
				assert.Equal(t, strings.TrimSpace(tt.wantExamples), strings.TrimSpace(help.Examples))
			}
			assert.Equal(t, tt.wantArgs, len(help.Arguments))
			assert.Equal(t, tt.wantFlags, len(help.Flags))
		})
	}
}

func TestFormatDescription(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{
			name:  "single line",
			lines: []string{"This is a description."},
			want:  "This is a description.",
		},
		{
			name:  "multiple lines same paragraph",
			lines: []string{"This is line one.", "This is line two."},
			want:  "This is line one. This is line two.",
		},
		{
			name:  "two paragraphs",
			lines: []string{"Paragraph one.", "", "Paragraph two."},
			want:  "Paragraph one.\n\nParagraph two.",
		},
		{
			name:  "multiple paragraphs",
			lines: []string{"Para one line one.", "Para one line two.", "", "Para two.", "", "Para three."},
			want:  "Para one line one. Para one line two.\n\nPara two.\n\nPara three.",
		},
		{
			name:  "empty input",
			lines: []string{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDescription(tt.lines)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseFlagLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantName string
		wantDesc string
		wantNil  bool
	}{
		{
			name:     "standard flag",
			line:     "  --dry-run    Show what would be done",
			wantName: "--dry-run",
			wantDesc: "Show what would be done",
		},
		{
			name:     "short flag",
			line:     "  -v    Verbose output",
			wantName: "-v",
			wantDesc: "Verbose output",
		},
		{
			name:     "flag with value placeholder",
			line:     "  --format <type>    Output format",
			wantName: "--format <type>",
			wantDesc: "Output format",
		},
		{
			name:    "non-flag line",
			line:    "This is not a flag",
			wantNil: true,
		},
		{
			name:    "empty line",
			line:    "",
			wantNil: true,
		},
		{
			name:    "only whitespace",
			line:    "    ",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseFlagLine(tt.line)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, tt.wantDesc, got.Description)
		})
	}
}

func TestDedentExamples(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no indent",
			input: "build core\nbuild eac-cli",
			want:  "build core\nbuild eac-cli",
		},
		{
			name:  "consistent indent after trim",
			input: "  build core\n  build eac-cli",
			want:  "build core\n  build eac-cli",
		},
		{
			name:  "mixed indent after trim",
			input: "  build core\n    build --dry-run",
			want:  "build core\n    build --dry-run",
		},
		{
			name:  "typical help example output",
			input: "build core\n  build core --dry-run",
			want:  "build core\n  build core --dry-run",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DedentExamples(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEscapeJinja2(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "double braces",
			input: "Use {{ variable }} in templates",
			want:  "Use { { variable } } in templates",
		},
		{
			name:  "statement tags",
			input: "{% if condition %}",
			want:  "{ % if condition % }",
		},
		{
			name:  "no templates",
			input: "Normal text without templates",
			want:  "Normal text without templates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeJinja2(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEscapeTableCell(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "pipe character",
			input: "format: json|yaml",
			want:  "format: json\\|yaml",
		},
		{
			name:  "newline",
			input: "Line one\nLine two",
			want:  "Line one Line two",
		},
		{
			name:  "both",
			input: "Option: a|b\nDefault: a",
			want:  "Option: a\\|b Default: a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeTableCell(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatCommandHelp_NoTitle(t *testing.T) {
	help := &CommandHelp{
		Name:        "show modules",
		Description: "First paragraph.\n\nSecond paragraph.",
		Usage:       "show modules [flags]",
	}

	result := FormatCommandHelp(help, 2, false)

	assert.NotContains(t, result, "## show modules")
	assert.Contains(t, result, "First paragraph.\n\nSecond paragraph.")
	assert.Contains(t, result, "**Usage:** `show modules [flags]`")
}

func TestFormatCommandHelp_WithTitle(t *testing.T) {
	help := &CommandHelp{
		Name:        "show modules",
		Description: "Test description.",
	}

	result := FormatCommandHelp(help, 2, true)

	assert.Contains(t, result, "## show modules")
}

func TestFormatCommandHelp_WithFlags(t *testing.T) {
	help := &CommandHelp{
		Name:        "build",
		Description: "Build modules.",
		Usage:       "build [flags]",
		Flags: []FlagArg{
			{Name: "--dry-run", Description: "Show what would be built"},
			{Name: "--parallel", Description: "Build in parallel"},
		},
	}

	result := FormatCommandHelp(help, 2, false)

	assert.Contains(t, result, "### Flags")
	assert.Contains(t, result, "| `--dry-run` | Show what would be built |")
	assert.Contains(t, result, "| `--parallel` | Build in parallel |")
}

func TestFormatCommandHelp_WithExamples(t *testing.T) {
	help := &CommandHelp{
		Name:     "build",
		Examples: "build core\nbuild eac-cli",
	}

	result := FormatCommandHelp(help, 2, false)

	assert.Contains(t, result, "```bash\n")
	assert.Contains(t, result, "build core\nbuild eac-cli")
}

func TestFormatCommandHelp_WithArguments(t *testing.T) {
	help := &CommandHelp{
		Name: "build",
		Arguments: []FlagArg{
			{Name: "module", Description: "Module to build"},
		},
	}

	result := FormatCommandHelp(help, 3, true)

	assert.Contains(t, result, "#### Arguments")
	assert.Contains(t, result, "| `module` | Module to build |")
}
