// markdown-commands.go - Build handler for executing EAC commands that produce markdown fragments.
// These fragments are consumed by the docs preprocessing pipeline (book, docs-site, docs-pdf).
// Caching is handled by the build framework at the UoW level (module source hash).
package builders

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/content"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/tool"
	"gopkg.in/yaml.v3"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&MarkdownCommandsHandler{})
}

// MarkdownCommandsHandler executes EAC commands that produce markdown fragments.
// When this handler is called, it executes all commands unconditionally.
// The build framework decides whether to call it (module source hash comparison).
type MarkdownCommandsHandler struct{}

func (h *MarkdownCommandsHandler) Name() string          { return "markdown-commands" }
func (h *MarkdownCommandsHandler) Requirements() []string { return nil }
func (h *MarkdownCommandsHandler) IsContainer() bool      { return false }
func (h *MarkdownCommandsHandler) IsHostInstalled() bool   { return true }

func (h *MarkdownCommandsHandler) ValidateModule(
	module core.ModuleContractPort, workspaceRoot, component string,
) error {
	return nil
}

func (h *MarkdownCommandsHandler) ListArtifacts(
	module core.ModuleContractPort, workspaceRoot string,
) []string {
	return []string{"markdown-commands/"}
}

func (h *MarkdownCommandsHandler) Build(
	module core.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, opts BuildOptions,
) int {
	Logln(logWriter, "=== Executing markdown commands ===")

	// Discover command sources from book config
	commands, err := discoverCommandSources(module, workspaceRoot, opts.Component)
	if err != nil {
		Logln(logWriter, "Error discovering commands: %v", err)
		return 1
	}

	outputCmdsDir := filepath.Join(outputDir, "markdown-commands")
	if err := os.MkdirAll(outputCmdsDir, 0o755); err != nil {
		Logln(logWriter, "Error creating output dir: %v", err)
		return 1
	}

	if len(commands) == 0 {
		Logln(logWriter, "No markdown commands found")
		return 0
	}

	Logln(logWriter, "Found %d markdown command(s)", len(commands))

	// Execute each command and write fragment
	executor := content.ToolCommandExecutor{}
	failed := 0
	for _, cmd := range commands {
		output, err := executor.Run(context.Background(), workspaceRoot, strings.Fields(cmd.Command))
		if err != nil {
			Logln(logWriter, "  FAIL: %s -> %v", cmd.Command, err)
			failed++
			continue
		}

		fragment := formatCommandFragment(cmd, output)
		destPath := filepath.Join(outputCmdsDir, cmd.Target)
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			Logln(logWriter, "  FAIL: mkdir %s -> %v", filepath.Dir(destPath), err)
			failed++
			continue
		}
		if err := os.WriteFile(destPath, []byte(fragment), 0o644); err != nil {
			Logln(logWriter, "  FAIL: write %s -> %v", destPath, err)
			failed++
			continue
		}

		Logln(logWriter, "  OK: %s -> %s", cmd.Command, cmd.Target)
	}

	if failed > 0 {
		Logln(logWriter, "%d/%d commands failed", failed, len(commands))
		return 1
	}

	Logln(logWriter, "All %d commands executed successfully", len(commands))
	return 0
}

// CommandSource holds a command definition from book config.
type CommandSource struct {
	Command     string
	Target      string
	Frontmatter map[string]any
	Order       int
}

// discoverCommandSources extracts command sources from the book associated with this component.
func discoverCommandSources(
	module core.ModuleContractPort, workspaceRoot, component string,
) ([]CommandSource, error) {
	cfg := config.Global()
	if cfg == nil {
		return nil, fmt.Errorf("global config not loaded")
	}

	// Resolve which book this component serves
	bookName := resolveBookNameForMarkdownCommands(module, component)
	book := cfg.GetBookByName(bookName)
	if book == nil {
		return nil, nil // No book = no commands
	}

	// Extract type=command sources
	var commands []CommandSource
	for _, src := range book.GetCommandSources() {
		if src.Command == "" {
			continue
		}
		commands = append(commands, CommandSource{
			Command:     src.Command,
			Target:      src.Target,
			Frontmatter: src.Frontmatter,
			Order:       src.Order,
		})
	}

	// Sort by order for determinism
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Order < commands[j].Order
	})

	return commands, nil
}

// resolveBookNameForMarkdownCommands derives book name from component config or naming convention.
// Priority:
// 1. config.book field (explicit override)
// 2. Strip "-markdown-commands" or "-mdcmds" suffix
// 3. Use component name as book name
// 4. Default to "site" if component name is empty
func resolveBookNameForMarkdownCommands(module core.ModuleContractPort, componentName string) string {
	if componentName == "" {
		return "site"
	}

	// Try to get book name from component config
	concrete := adapters.UnwrapModule(module)
	if concrete != nil {
		if comp, ok := concrete.Components[componentName]; ok && comp != nil {
			if bookName, ok := comp.Config["book"]; ok && bookName != "" {
				return bookName
			}
		}
	}

	// Convention: strip "-markdown-commands" or "-mdcmds" suffix
	if strings.HasSuffix(componentName, "-markdown-commands") {
		return strings.TrimSuffix(componentName, "-markdown-commands")
	}
	if strings.HasSuffix(componentName, "-mdcmds") {
		return strings.TrimSuffix(componentName, "-mdcmds")
	}

	// Fall back to component name as book name
	return componentName
}

// formatCommandFragment formats command output with optional frontmatter.
// Unlike content.FormatCommandOutput, this omits generation timestamps since
// the build framework handles caching at the UoW level.
func formatCommandFragment(cmd CommandSource, output string) string {
	var buf bytes.Buffer

	if len(cmd.Frontmatter) > 0 {
		buf.WriteString("---\n")
		frontmatterBytes, _ := yaml.Marshal(cmd.Frontmatter)
		buf.Write(frontmatterBytes)
		buf.WriteString("---\n\n")
	}

	buf.WriteString(strings.TrimSpace(output))
	buf.WriteString("\n")

	return buf.String()
}
