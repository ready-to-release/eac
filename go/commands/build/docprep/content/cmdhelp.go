package content

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/commands/build/docprep/staging"
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
// Files are processed in parallel (up to 8 goroutines) to reduce wall-clock time when
// there are many markers (e.g. 200+ book:cmd markers each spawning a subprocess).
// A cachingExecutor deduplicates calls such as GetValidCommands across all files.
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

	// Wrap with caching to deduplicate subprocess calls across files.
	// GetValidCommands is called once per category/group marker but always
	// returns the same result; caching reduces it to a single subprocess call.
	cached := newCachingExecutor(executor)

	// Thread-safe warn wrapper for parallel file processing.
	var warnMu sync.Mutex
	safeWarnf := func(format string, args ...any) {
		warnMu.Lock()
		warnf(format, args...)
		warnMu.Unlock()
	}

	files := fileIndex.GetMarkdownFiles()
	type fileResult struct {
		replacements int
		err          error
	}
	results := make([]fileResult, len(files))

	const parallelism = 8
	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup

	for i, p := range files {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, filePath string) {
			defer wg.Done()
			defer func() { <-sem }()
			count, err := processFileMarkers(ctx, filePath, stagingDir, workspaceRoot, cached, safeWarnf)
			results[idx] = fileResult{replacements: count, err: err}
		}(i, p)
	}
	wg.Wait()

	replacements := 0
	filesModified := 0
	for _, r := range results {
		if r.err != nil {
			return r.err
		}
		if r.replacements > 0 {
			replacements += r.replacements
			filesModified++
		}
	}

	logf("    Processed %d command markers in %d files", replacements, filesModified)
	return nil
}

// processFileMarkers processes command help markers in a single markdown file.
// Returns the number of replacements made.
func processFileMarkers(
	ctx context.Context,
	path, stagingDir, workspaceRoot string,
	executor CommandExecutor,
	warnf func(string, ...any),
) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	modified := string(data)
	fileReplacements := 0

	for markerType, pattern := range cmdMarkerPatterns {
		matches := pattern.FindAllStringSubmatch(modified, -1)
		for _, match := range matches {
			var replacement string
			var replErr error

			switch markerType {
			case "cmd":
				cmdName := match[1]
				replacement, replErr = FormatSingleCommand(ctx, workspaceRoot, executor, cmdName)
			case "cmd-group":
				groupName := match[1]
				replacement, replErr = FormatCommandGroup(ctx, workspaceRoot, executor, groupName, warnf)
			case "cmd-all":
				replacement, replErr = FormatAllCommands(ctx, workspaceRoot, executor)
			case "cmd-toc":
				replacement, replErr = FormatCommandTOC(ctx, workspaceRoot, executor)
			case "categories-table":
				replacement, replErr = FormatCategoriesTable(ctx, workspaceRoot, executor)
			case "categories-list":
				replacement, replErr = FormatCategoriesList(ctx, workspaceRoot, executor)
			case "category-section":
				categoryName := match[1]
				replacement, replErr = FormatCategorySection(ctx, workspaceRoot, executor, categoryName)
			case "category-commands":
				categoryName := match[1]
				replacement, replErr = FormatCategoryCommands(ctx, workspaceRoot, executor, categoryName)
			case "categories-index":
				replacement, replErr = FormatCategoriesIndex(ctx, workspaceRoot, executor)
			}

			if replErr != nil {
				if errors.Is(replErr, ErrCommandNotFound) {
					relPath, relErr := filepath.Rel(stagingDir, path)
					if relErr != nil {
						relPath = path
					}
					return 0, fmt.Errorf("command marker in %s references non-existent command: %w", relPath, replErr)
				}
				warnf("failed to process %s marker '%s': %v", markerType, match[0], replErr)
				continue
			}

			wrapped := fmt.Sprintf("<!-- book:generated %s -->\n%s\n<!-- /book:generated -->", match[0][5:len(match[0])-4], replacement)
			modified = strings.Replace(modified, match[0], wrapped, 1)
			fileReplacements++
		}
	}

	if fileReplacements > 0 {
		if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
			return 0, err
		}
	}
	return fileReplacements, nil
}
