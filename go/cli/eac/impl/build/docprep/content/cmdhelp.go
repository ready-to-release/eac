package content

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/staging"
	"github.com/ready-to-release/eac/go/core/paths"
)

// ErrCommandNotFound indicates a command marker references a non-existent command.
var ErrCommandNotFound = errors.New("command not found")

// CommandHelp represents parsed help output for a command.
type CommandHelp struct {
	Name        string
	Description string
	Usage       string
	Arguments   []FlagArg
	Flags       []FlagArg
	Notes       string
	Examples    string
}

// FlagArg represents a flag or argument with description.
type FlagArg struct {
	Name        string
	Description string
}

// CommandInfo represents a command from get valid-commands.
type CommandInfo struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

// CategoryStats holds category information for documentation.
type CategoryStats struct {
	Name         string
	Description  string
	CommandCount int
}

// flagSplitPattern splits flag lines on 2+ spaces.
var flagSplitPattern = regexp.MustCompile(`\s{2,}`)

// cmdMarkerPatterns matches command markers in markdown.
var cmdMarkerPatterns = map[string]*regexp.Regexp{
	"cmd":               regexp.MustCompile(`<!--\s*book:cmd\s+([a-zA-Z0-9_-]+(?:\s+[a-zA-Z0-9_-]+)*)\s*-->`),
	"cmd-group":         regexp.MustCompile(`<!--\s*book:cmd-group\s+([a-zA-Z0-9_-]+)\s*-->`),
	"cmd-all":           regexp.MustCompile(`<!--\s*book:cmd-all\s*-->`),
	"cmd-toc":           regexp.MustCompile(`<!--\s*book:cmd-toc\s*-->`),
	"categories-table":  regexp.MustCompile(`<!--\s*book:categories-table\s*-->`),
	"categories-list":   regexp.MustCompile(`<!--\s*book:categories-list\s*-->`),
	"category-section":  regexp.MustCompile(`<!--\s*book:category-section\s+([a-zA-Z0-9_-]+)\s*-->`),
	"category-commands": regexp.MustCompile(`<!--\s*book:category-commands\s+([a-zA-Z0-9_-]+)\s*-->`),
	"categories-index":  regexp.MustCompile(`<!--\s*book:categories-index\s*-->`),
}

// ProcessCommandMarkers finds and replaces command help markers in staging markdown files.
func ProcessCommandMarkers(
	ctx context.Context,
	fileIndex *staging.FileIndex,
	stagingDir, workspaceRoot string,
	executor CommandExecutor,
	logf func(string, ...any),
	warnf func(string, ...any),
) error {
	logf("    Processing command help markers...")

	cmdBinary := paths.CommandsBinaryPath(workspaceRoot)
	if _, err := os.Stat(cmdBinary); os.IsNotExist(err) {
		logf("    Warning: eac binary not found at %s, skipping command markers", cmdBinary)
		return nil
	}

	replacements := 0
	filesModified := 0

	for _, path := range fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := original
		fileReplacements := 0

		for markerType, pattern := range cmdMarkerPatterns {
			matches := pattern.FindAllStringSubmatch(modified, -1)
			for _, match := range matches {
				var replacement string
				var err error

				switch markerType {
				case "cmd":
					cmdName := match[1]
					replacement, err = FormatSingleCommand(ctx, workspaceRoot, executor, cmdName)
				case "cmd-group":
					groupName := match[1]
					replacement, err = FormatCommandGroup(ctx, workspaceRoot, executor, groupName, warnf)
				case "cmd-all":
					replacement, err = FormatAllCommands(ctx, workspaceRoot, executor)
				case "cmd-toc":
					replacement, err = FormatCommandTOC(ctx, workspaceRoot, executor)
				case "categories-table":
					replacement, err = FormatCategoriesTable(ctx, workspaceRoot, executor)
				case "categories-list":
					replacement, err = FormatCategoriesList(ctx, workspaceRoot, executor)
				case "category-section":
					categoryName := match[1]
					replacement, err = FormatCategorySection(ctx, workspaceRoot, executor, categoryName)
				case "category-commands":
					categoryName := match[1]
					replacement, err = FormatCategoryCommands(ctx, workspaceRoot, executor, categoryName)
				case "categories-index":
					replacement, err = FormatCategoriesIndex(ctx, workspaceRoot, executor)
				}

				if err != nil {
					if errors.Is(err, ErrCommandNotFound) {
						relPath, relErr := filepath.Rel(stagingDir, path)
						if relErr != nil {
							relPath = path
						}
						return fmt.Errorf("command marker in %s references non-existent command: %w", relPath, err)
					}
					warnf("failed to process %s marker '%s': %v", markerType, match[0], err)
					continue
				}

				wrapped := fmt.Sprintf("<!-- book:generated %s -->\n%s\n<!-- /book:generated -->", match[0][5:len(match[0])-4], replacement)
				modified = strings.Replace(modified, match[0], wrapped, 1)
				fileReplacements++
			}
		}

		if fileReplacements > 0 {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			replacements += fileReplacements
			filesModified++
		}
	}

	logf("    Processed %d command markers in %d files", replacements, filesModified)
	return nil
}
