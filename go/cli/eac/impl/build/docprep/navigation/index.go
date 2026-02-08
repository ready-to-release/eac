package navigation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
)

// Patterns for title extraction.
var (
	frontmatterTitlePattern = regexp.MustCompile(`(?m)^title:\s*["']?([^"'\n]+)["']?`)
	h1HeadingPattern        = regexp.MustCompile(`(?m)^#\s+(.+)$`)
)

// tocEntry represents a TOC item.
type tocEntry struct {
	title string
	path  string
	depth int
	order int
}

// EnsureRootIndex ensures an index.md exists at staging root.
func EnsureRootIndex(
	book *config.Book,
	stagingDir string,
	logf func(string, ...any),
) error {
	indexPath := paths.IndexMarkdownPath(stagingDir)

	if _, err := os.Stat(indexPath); err == nil {
		logf("    Root index.md exists (from source)")
		return nil
	}

	content := GenerateRootIndex(book, stagingDir)
	if err := os.WriteFile(indexPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing index.md: %w", err)
	}

	logf("    Generated: index.md")
	return nil
}

// GenerateRootIndex creates index.md content with book metadata and TOC.
func GenerateRootIndex(book *config.Book, stagingDir string) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	title := book.Title
	if title == "" {
		title = book.Name
	}
	sb.WriteString(fmt.Sprintf("title: %s\n", title))
	if book.Description != "" {
		sb.WriteString(fmt.Sprintf("description: %s\n", book.Description))
	}
	sb.WriteString("---\n\n")

	if title == "" {
		title = "Documentation"
	}
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	if book.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", book.Description))
	}

	toc := GenerateTOC(book, stagingDir)
	if toc != "" {
		sb.WriteString("## Contents\n\n")
		sb.WriteString(toc)
	}

	return sb.String()
}

// GenerateTOC creates a markdown TOC from the staging directory.
func GenerateTOC(book *config.Book, stagingDir string) string {
	var sb strings.Builder

	entries := CollectTOCEntries(book, stagingDir, stagingDir, 0)

	for _, entry := range entries {
		indent := strings.Repeat("  ", entry.depth)
		sb.WriteString(fmt.Sprintf("%s- [%s](%s)\n", indent, entry.title, entry.path))
	}

	return sb.String()
}

// CollectTOCEntries recursively collects markdown files for TOC.
func CollectTOCEntries(book *config.Book, stagingDir, dir string, depth int) []tocEntry {
	var entries []tocEntry

	navPath := paths.NavigationConfigPath(dir)
	navFile, err := ParseNavFile(navPath)

	if err == nil && len(navFile.Nav) > 0 {
		entries = collectTOCFromNav(book, stagingDir, dir, depth, navFile.Nav)
	} else {
		entries = collectTOCFromFilesystem(book, stagingDir, dir, depth)
	}

	return entries
}

// collectTOCFromNav collects TOC entries using .nav.yml order.
func collectTOCFromNav(book *config.Book, stagingDir, dir string, depth int, nav []any) []tocEntry {
	var entries []tocEntry

	for _, item := range nav {
		switch v := item.(type) {
		case string:
			name := v
			subdirName := strings.TrimSuffix(name, "/")
			subdir := filepath.Join(dir, subdirName)

			if info, err := os.Stat(subdir); err == nil && info.IsDir() {
				indexPath := paths.IndexMarkdownPath(subdir)
				if _, err := os.Stat(indexPath); err == nil {
					title := GetTitleFromFile(indexPath)
					relPath, relErr := filepath.Rel(stagingDir, subdir)
					if relErr != nil {
						relPath = subdir
					}
					relPath = filepath.ToSlash(relPath)
					entries = append(entries, tocEntry{
						title: title,
						path:  relPath + "/",
						depth: depth,
					})
				}
				subEntries := CollectTOCEntries(book, stagingDir, subdir, depth+1)
				entries = append(entries, subEntries...)
			} else if strings.HasSuffix(name, ".md") && name != "index.md" {
				filePath := filepath.Join(dir, name)
				if _, err := os.Stat(filePath); err == nil {
					title := GetTitleFromFile(filePath)
					relPath, relErr := filepath.Rel(stagingDir, filePath)
					if relErr != nil {
						relPath = filePath
					}
					relPath = filepath.ToSlash(relPath)
					entries = append(entries, tocEntry{
						title: title,
						path:  relPath,
						depth: depth,
					})
				}
			}
		case map[string]any:
			for title, target := range v {
				targetStr, ok := target.(string)
				if !ok {
					continue
				}
				subdirName := strings.TrimSuffix(targetStr, "/")
				subdir := filepath.Join(dir, subdirName)

				if info, err := os.Stat(subdir); err == nil && info.IsDir() {
					relPath, relErr := filepath.Rel(stagingDir, subdir)
					if relErr != nil {
						relPath = subdir
					}
					relPath = filepath.ToSlash(relPath)
					entries = append(entries, tocEntry{
						title: title,
						path:  relPath + "/",
						depth: depth,
					})
					subEntries := CollectTOCEntries(book, stagingDir, subdir, depth+1)
					entries = append(entries, subEntries...)
				} else if strings.HasSuffix(targetStr, ".md") && targetStr != "index.md" {
					filePath := filepath.Join(dir, targetStr)
					if _, err := os.Stat(filePath); err == nil {
						relPath, relErr := filepath.Rel(stagingDir, filePath)
						if relErr != nil {
							relPath = filePath
						}
						relPath = filepath.ToSlash(relPath)
						entries = append(entries, tocEntry{
							title: title,
							path:  relPath,
							depth: depth,
						})
					}
				}
			}
		}
	}

	return entries
}

// collectTOCFromFilesystem collects TOC entries from filesystem (fallback).
func collectTOCFromFilesystem(book *config.Book, stagingDir, dir string, depth int) []tocEntry {
	var entries []tocEntry

	items, err := os.ReadDir(dir)
	if err != nil {
		return entries
	}

	var files, dirs []os.DirEntry
	for _, item := range items {
		name := item.Name()
		if strings.HasPrefix(name, ".") || name == "assets" {
			continue
		}
		if item.IsDir() {
			dirs = append(dirs, item)
		} else if strings.HasSuffix(name, ".md") && name != "index.md" {
			files = append(files, item)
		}
	}

	fileEntries := make([]tocEntry, 0, len(files))
	for _, f := range files {
		name := f.Name()
		filePath := filepath.Join(dir, name)
		relPath, relErr := filepath.Rel(stagingDir, filePath)
		if relErr != nil {
			relPath = filePath
		}
		relPath = filepath.ToSlash(relPath)

		title := GetTitleFromFile(filePath)
		order := GetOrderForFile(book, relPath)

		fileEntries = append(fileEntries, tocEntry{
			title: title,
			path:  relPath,
			depth: depth,
			order: order,
		})
	}

	sort.Slice(fileEntries, func(i, j int) bool {
		if fileEntries[i].order != fileEntries[j].order {
			return fileEntries[i].order < fileEntries[j].order
		}
		return fileEntries[i].title < fileEntries[j].title
	})

	entries = append(entries, fileEntries...)

	for _, d := range dirs {
		subdir := filepath.Join(dir, d.Name())
		relPath, relErr := filepath.Rel(stagingDir, subdir)
		if relErr != nil {
			relPath = subdir
		}
		relPath = filepath.ToSlash(relPath)

		indexPath := paths.IndexMarkdownPath(subdir)
		if _, err := os.Stat(indexPath); err == nil {
			title := GetTitleFromFile(indexPath)
			entries = append(entries, tocEntry{
				title: title,
				path:  relPath + "/",
				depth: depth,
				order: 1000,
			})
		}

		subEntries := CollectTOCEntries(book, stagingDir, subdir, depth+1)
		entries = append(entries, subEntries...)
	}

	return entries
}
