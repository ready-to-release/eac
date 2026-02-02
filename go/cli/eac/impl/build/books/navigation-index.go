package books

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/core/paths"
)

// tocEntry represents a TOC item.
type tocEntry struct {
	title string
	path  string
	depth int
	order int
}

// ensureRootIndex ensures an index.md exists at staging root
// If no index.md exists, generates one with book metadata and TOC
// If index.md exists (from copy source), keeps it as-is.
func (p *Preprocessor) ensureRootIndex() error {
	indexPath := paths.IndexMarkdownPath(p.stagingDir)

	// Check if index.md already exists
	if _, err := os.Stat(indexPath); err == nil {
		p.log("    Root index.md exists (from source)")
		return nil
	}

	// Generate index.md with book metadata and TOC
	content := p.generateRootIndex()
	if err := os.WriteFile(indexPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing index.md: %w", err)
	}

	p.log("    Generated: index.md")
	return nil
}

// generateRootIndex creates index.md content with book metadata and TOC.
func (p *Preprocessor) generateRootIndex() string {
	var sb strings.Builder

	// Frontmatter - use Title for display
	sb.WriteString("---\n")
	title := p.book.Title
	if title == "" {
		title = p.book.Name
	}
	sb.WriteString(fmt.Sprintf("title: %s\n", title))
	if p.book.Description != "" {
		sb.WriteString(fmt.Sprintf("description: %s\n", p.book.Description))
	}
	sb.WriteString("---\n\n")

	// Title heading
	if title == "" {
		title = "Documentation"
	}
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	// Description
	if p.book.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", p.book.Description))
	}

	// Generate TOC from staging directory structure
	toc := p.generateTOC()
	if toc != "" {
		sb.WriteString("## Contents\n\n")
		sb.WriteString(toc)
	}

	return sb.String()
}

// generateTOC creates a markdown TOC from the staging directory.
func (p *Preprocessor) generateTOC() string {
	var sb strings.Builder

	// Get all markdown files
	entries := p.collectTOCEntries(p.stagingDir, 0)

	for _, entry := range entries {
		// Indent based on depth
		indent := strings.Repeat("  ", entry.depth)
		sb.WriteString(fmt.Sprintf("%s- [%s](%s)\n", indent, entry.title, entry.path))
	}

	return sb.String()
}

// collectTOCEntries recursively collects markdown files for TOC
// Uses .nav.yml order when available to preserve source navigation structure.
func (p *Preprocessor) collectTOCEntries(dir string, depth int) []tocEntry {
	var entries []tocEntry

	// DEBUG: Log directory scanning
	relDir, relErr := filepath.Rel(p.stagingDir, dir)
	if relErr != nil || relDir == "." {
		relDir = "(root)"
	}
	navLog.Debugf("[TOC] Scanning directory: %s", relDir)

	// Try to read .nav.yml for ordering
	navPath := paths.NavigationConfigPath(dir)
	navFile, err := p.parseNavFile(navPath)

	if err == nil && len(navFile.Nav) > 0 {
		// Use .nav.yml order
		entries = p.collectTOCFromNav(dir, depth, navFile.Nav)
	} else {
		// Fallback to filesystem order (for directories without .nav.yml)
		entries = p.collectTOCFromFilesystem(dir, depth)
	}

	return entries
}

// collectTOCFromNav collects TOC entries using .nav.yml order.
func (p *Preprocessor) collectTOCFromNav(dir string, depth int, nav []any) []tocEntry {
	var entries []tocEntry

	for _, item := range nav {
		switch v := item.(type) {
		case string:
			// Simple entry: "file.md", "subdir/", or "subdir" (directory without slash)
			name := v
			// Check if it's a directory (with or without trailing slash)
			subdirName := strings.TrimSuffix(name, "/")
			subdir := filepath.Join(dir, subdirName)

			if info, err := os.Stat(subdir); err == nil && info.IsDir() {
				// Directory reference
				// Add directory entry if it has index.md
				indexPath := paths.IndexMarkdownPath(subdir)
				if _, err := os.Stat(indexPath); err == nil {
					title := p.getTitleFromFile(indexPath)
					relPath, relErr := filepath.Rel(p.stagingDir, subdir)
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
				// Recurse into subdirectory
				subEntries := p.collectTOCEntries(subdir, depth+1)
				entries = append(entries, subEntries...)
			} else if strings.HasSuffix(name, ".md") && name != "index.md" {
				// Markdown file
				filePath := filepath.Join(dir, name)
				if _, err := os.Stat(filePath); err == nil {
					title := p.getTitleFromFile(filePath)
					relPath, relErr := filepath.Rel(p.stagingDir, filePath)
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
			// Map entry: "Title": "file.md" or "Title": "subdir/"
			for title, target := range v {
				targetStr, ok := target.(string)
				if !ok {
					continue
				}
				// Check if it's a directory
				subdirName := strings.TrimSuffix(targetStr, "/")
				subdir := filepath.Join(dir, subdirName)

				if info, err := os.Stat(subdir); err == nil && info.IsDir() {
					// Directory with custom title
					relPath, relErr := filepath.Rel(p.stagingDir, subdir)
					if relErr != nil {
						relPath = subdir
					}
					relPath = filepath.ToSlash(relPath)
					entries = append(entries, tocEntry{
						title: title,
						path:  relPath + "/",
						depth: depth,
					})
					// Recurse into subdirectory
					subEntries := p.collectTOCEntries(subdir, depth+1)
					entries = append(entries, subEntries...)
				} else if strings.HasSuffix(targetStr, ".md") && targetStr != "index.md" {
					// File with custom title
					filePath := filepath.Join(dir, targetStr)
					if _, err := os.Stat(filePath); err == nil {
						relPath, relErr := filepath.Rel(p.stagingDir, filePath)
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
func (p *Preprocessor) collectTOCFromFilesystem(dir string, depth int) []tocEntry {
	var entries []tocEntry

	items, err := os.ReadDir(dir)
	if err != nil {
		return entries
	}

	// Separate files and directories
	var files, dirs []os.DirEntry
	for _, item := range items {
		name := item.Name()
		// Skip hidden files, assets, and index.md (it's the current page)
		if strings.HasPrefix(name, ".") || name == "assets" {
			continue
		}
		if item.IsDir() {
			dirs = append(dirs, item)
			navLog.Debugf("[TOC]   Found directory: %s/", name)
		} else if strings.HasSuffix(name, ".md") && name != "index.md" {
			files = append(files, item)
			navLog.Debugf("[TOC]   Found markdown: %s", name)
		}
	}

	// Process files first (with ordering from command sources)
	fileEntries := make([]tocEntry, 0, len(files))
	for _, f := range files {
		name := f.Name()
		filePath := filepath.Join(dir, name)
		relPath, relErr := filepath.Rel(p.stagingDir, filePath)
		if relErr != nil {
			relPath = filePath
		}
		relPath = filepath.ToSlash(relPath)

		title := p.getTitleFromFile(filePath)
		order := p.getOrderForFile(relPath)

		fileEntries = append(fileEntries, tocEntry{
			title: title,
			path:  relPath,
			depth: depth,
			order: order,
		})
	}

	// Sort files by order, then by title
	sort.Slice(fileEntries, func(i, j int) bool {
		if fileEntries[i].order != fileEntries[j].order {
			return fileEntries[i].order < fileEntries[j].order
		}
		return fileEntries[i].title < fileEntries[j].title
	})

	entries = append(entries, fileEntries...)

	// Process directories
	for _, d := range dirs {
		subdir := filepath.Join(dir, d.Name())
		relPath, relErr := filepath.Rel(p.stagingDir, subdir)
		if relErr != nil {
			relPath = subdir
		}
		relPath = filepath.ToSlash(relPath)

		// Check for index.md in subdirectory
		indexPath := paths.IndexMarkdownPath(subdir)
		if _, err := os.Stat(indexPath); err == nil {
			title := p.getTitleFromFile(indexPath)
			entries = append(entries, tocEntry{
				title: title,
				path:  relPath + "/",
				depth: depth,
				order: 1000, // Directories come after files
			})
		}

		// Recurse into subdirectory
		subEntries := p.collectTOCEntries(subdir, depth+1)
		entries = append(entries, subEntries...)
	}

	return entries
}
