package books

import (
	"strings"
	"testing"
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
			help, err := parseHelpOutput(tt.cmdName, tt.input)
			if err != nil {
				t.Fatalf("parseHelpOutput() error = %v", err)
			}

			if help.Name != tt.cmdName {
				t.Errorf("Name = %q, want %q", help.Name, tt.cmdName)
			}

			if tt.wantDesc != "" && help.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", help.Description, tt.wantDesc)
			}

			if tt.wantUsage != "" && help.Usage != tt.wantUsage {
				t.Errorf("Usage = %q, want %q", help.Usage, tt.wantUsage)
			}

			if tt.wantNotes != "" && help.Notes != tt.wantNotes {
				t.Errorf("Notes = %q, want %q", help.Notes, tt.wantNotes)
			}

			if tt.wantExamples != "" && strings.TrimSpace(help.Examples) != strings.TrimSpace(tt.wantExamples) {
				t.Errorf("Examples = %q, want %q", strings.TrimSpace(help.Examples), strings.TrimSpace(tt.wantExamples))
			}

			if len(help.Arguments) != tt.wantArgs {
				t.Errorf("Arguments count = %d, want %d", len(help.Arguments), tt.wantArgs)
			}

			if len(help.Flags) != tt.wantFlags {
				t.Errorf("Flags count = %d, want %d", len(help.Flags), tt.wantFlags)
			}
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
			got := formatDescription(tt.lines)
			if got != tt.want {
				t.Errorf("formatDescription() = %q, want %q", got, tt.want)
			}
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
			got := parseFlagLine(tt.line)
			if tt.wantNil {
				if got != nil {
					t.Errorf("parseFlagLine() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("parseFlagLine() = nil, want non-nil")
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", got.Description, tt.wantDesc)
			}
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
			// Note: TrimSpace removes leading spaces from first line before processing,
			// so the minimum indent is calculated from remaining lines only
			name:  "consistent indent after trim",
			input: "  build core\n  build eac-cli",
			want:  "build core\n  build eac-cli", // First line trimmed, second unchanged (minIndent=0)
		},
		{
			name:  "mixed indent after trim",
			input: "  build core\n    build --dry-run",
			want:  "build core\n    build --dry-run", // First line trimmed, second unchanged (minIndent=0)
		},
		{
			// Realistic example: help output typically has no leading space on first line
			name:  "typical help example output",
			input: "build core\n  build core --dry-run",
			want:  "build core\n  build core --dry-run", // No dedent needed (minIndent=0)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dedentExamples(tt.input)
			if got != tt.want {
				t.Errorf("dedentExamples() = %q, want %q", got, tt.want)
			}
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
			got := escapeJinja2(tt.input)
			if got != tt.want {
				t.Errorf("escapeJinja2() = %q, want %q", got, tt.want)
			}
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
			got := escapeTableCell(tt.input)
			if got != tt.want {
				t.Errorf("escapeTableCell() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatCommandHelp_NoTitle(t *testing.T) {
	help := &CommandHelp{
		Name:        "show modules",
		Description: "First paragraph.\n\nSecond paragraph.",
		Usage:       "show modules [flags]",
	}

	// Create a minimal preprocessor (doesn't need workspaceRoot for this test)
	p := &Preprocessor{}

	result := p.formatCommandHelp(help, 2, false)

	// Should NOT contain "## show modules" since includeTitle=false
	if strings.Contains(result, "## show modules") {
		t.Errorf("formatCommandHelp with includeTitle=false should not include title heading, got:\n%s", result)
	}

	// Should contain the description with paragraph breaks
	if !strings.Contains(result, "First paragraph.\n\nSecond paragraph.") {
		t.Errorf("formatCommandHelp should preserve paragraph breaks in description, got:\n%s", result)
	}

	// Should contain usage
	if !strings.Contains(result, "**Usage:** `show modules [flags]`") {
		t.Errorf("formatCommandHelp should include usage, got:\n%s", result)
	}
}

func TestFormatCommandHelp_WithTitle(t *testing.T) {
	help := &CommandHelp{
		Name:        "show modules",
		Description: "Test description.",
	}

	p := &Preprocessor{}
	result := p.formatCommandHelp(help, 2, true)

	// Should contain "## show modules" since includeTitle=true
	if !strings.Contains(result, "## show modules") {
		t.Errorf("formatCommandHelp with includeTitle=true should include title heading, got:\n%s", result)
	}
}
