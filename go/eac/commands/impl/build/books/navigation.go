package books

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// NavFile represents a .nav.yml file structure
type NavFile struct {
	Title string `yaml:"title,omitempty"`
	Nav   []any  `yaml:"nav"`
}

// ensureRootIndex ensures an index.md exists at staging root
// If no index.md exists, generates one with book metadata and TOC
// If index.md exists (from copy source), keeps it as-is
func (p *Preprocessor) ensureRootIndex() error {
	indexPath := filepath.Join(p.stagingDir, "index.md")

	// Check if index.md already exists
	if _, err := os.Stat(indexPath); err == nil {
		p.log("    Root index.md exists (from source)")
		return nil
	}

	// Generate index.md with book metadata and TOC
	content := p.generateRootIndex()
	if err := os.WriteFile(indexPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing index.md: %w", err)
	}

	p.log("    Generated: index.md")
	return nil
}

// generateRootIndex creates index.md content with book metadata and TOC
func (p *Preprocessor) generateRootIndex() string {
	var sb strings.Builder

	// Frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("title: %s\n", p.book.Name))
	if p.book.Description != "" {
		sb.WriteString(fmt.Sprintf("description: %s\n", p.book.Description))
	}
	sb.WriteString("---\n\n")

	// Title
	title := p.book.Name
	if title == "" {
		title = "Documentation"
	}
	sb.WriteString(fmt.Sprintf("# %s\n\n", toTitleCase(title)))

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

// generateTOC creates a markdown TOC from the staging directory
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

// tocEntry represents a TOC item
type tocEntry struct {
	title string
	path  string
	depth int
	order int
}

// collectTOCEntries recursively collects markdown files for TOC
func (p *Preprocessor) collectTOCEntries(dir string, depth int) []tocEntry {
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
		} else if strings.HasSuffix(name, ".md") && name != "index.md" {
			files = append(files, item)
		}
	}

	// Process files first (with ordering from command sources)
	fileEntries := make([]tocEntry, 0, len(files))
	for _, f := range files {
		name := f.Name()
		filePath := filepath.Join(dir, name)
		relPath, _ := filepath.Rel(p.stagingDir, filePath)
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
		relPath, _ := filepath.Rel(p.stagingDir, subdir)
		relPath = filepath.ToSlash(relPath)

		// Check for index.md in subdirectory
		indexPath := filepath.Join(subdir, "index.md")
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

// getTitleFromFile extracts title from markdown frontmatter or first heading
func (p *Preprocessor) getTitleFromFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return filenameToTitle(filepath.Base(path))
	}

	text := string(content)

	// Try frontmatter title
	if strings.HasPrefix(text, "---") {
		endIdx := strings.Index(text[3:], "---")
		if endIdx > 0 {
			frontmatter := text[3 : 3+endIdx]
			titleMatch := regexp.MustCompile(`(?m)^title:\s*["']?([^"'\n]+)["']?`).FindStringSubmatch(frontmatter)
			if len(titleMatch) > 1 {
				return strings.TrimSpace(titleMatch[1])
			}
		}
	}

	// Try first H1
	h1Match := regexp.MustCompile(`(?m)^#\s+(.+)$`).FindStringSubmatch(text)
	if len(h1Match) > 1 {
		return strings.TrimSpace(h1Match[1])
	}

	return filenameToTitle(filepath.Base(path))
}

// getOrderForFile returns the order for a file from command sources
func (p *Preprocessor) getOrderForFile(relPath string) int {
	for _, src := range p.book.Sources {
		if src.Type == "command" && filepath.ToSlash(src.Target) == relPath {
			if src.Order > 0 {
				return src.Order
			}
		}
	}
	return 500 // Default order for non-command files
}

// ensureNavigationStructure ensures .nav.yml exists in all staging directories
func (p *Preprocessor) ensureNavigationStructure() error {
	return filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		// Skip hidden directories and assets
		name := d.Name()
		if strings.HasPrefix(name, ".") || name == "assets" {
			return filepath.SkipDir
		}

		// Check if .nav.yml exists
		navPath := filepath.Join(path, ".nav.yml")
		if _, err := os.Stat(navPath); err == nil {
			return nil // Already has navigation
		}

		// Generate .nav.yml for this directory
		if err := p.generateNavForDir(path); err != nil {
			return err
		}

		return nil
	})
}

// generateNavForDir creates .nav.yml for a directory
func (p *Preprocessor) generateNavForDir(dir string) error {
	items, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Collect navigation entries
	type navItem struct {
		name  string
		order int
		isDir bool
	}

	var navItems []navItem

	for _, item := range items {
		name := item.Name()

		// Skip hidden files and non-markdown files
		if strings.HasPrefix(name, ".") {
			continue
		}

		if item.IsDir() {
			// Skip assets directory
			if name == "assets" {
				continue
			}
			// Include directory if it has content
			subdir := filepath.Join(dir, name)
			if hasMarkdownContent(subdir) {
				navItems = append(navItems, navItem{
					name:  name + "/",
					order: 1000, // Directories after files
					isDir: true,
				})
			}
		} else if strings.HasSuffix(name, ".md") {
			filePath := filepath.Join(dir, name)
			relPath, _ := filepath.Rel(p.stagingDir, filePath)
			relPath = filepath.ToSlash(relPath)

			order := p.getOrderForFile(relPath)
			navItems = append(navItems, navItem{
				name:  name,
				order: order,
				isDir: false,
			})
		}
	}

	if len(navItems) == 0 {
		return nil // Empty directory, skip
	}

	// Sort: index.md first, then by order, then alphabetically
	sort.Slice(navItems, func(i, j int) bool {
		// index.md always first
		if navItems[i].name == "index.md" {
			return true
		}
		if navItems[j].name == "index.md" {
			return false
		}

		// Files before directories
		if !navItems[i].isDir && navItems[j].isDir {
			return true
		}
		if navItems[i].isDir && !navItems[j].isDir {
			return false
		}

		// By order
		if navItems[i].order != navItems[j].order {
			return navItems[i].order < navItems[j].order
		}

		// Alphabetically
		return navItems[i].name < navItems[j].name
	})

	// Build nav list
	nav := make([]any, 0, len(navItems))
	for _, item := range navItems {
		nav = append(nav, item.name)
	}

	navFile := NavFile{
		Nav: nav,
	}

	data, err := yaml.Marshal(navFile)
	if err != nil {
		return err
	}

	navPath := filepath.Join(dir, ".nav.yml")
	if err := os.WriteFile(navPath, data, 0644); err != nil {
		return err
	}

	relPath, _ := filepath.Rel(p.stagingDir, dir)
	if relPath == "." {
		relPath = "(root)"
	}
	p.log("    Generated: %s/.nav.yml", relPath)

	return nil
}

// hasMarkdownContent checks if a directory has any markdown files
func hasMarkdownContent(dir string) bool {
	items, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, item := range items {
		if strings.HasSuffix(item.Name(), ".md") {
			return true
		}
		if item.IsDir() && !strings.HasPrefix(item.Name(), ".") && item.Name() != "assets" {
			if hasMarkdownContent(filepath.Join(dir, item.Name())) {
				return true
			}
		}
	}
	return false
}

// filenameToTitle converts a filename to a readable title
func filenameToTitle(filename string) string {
	// Remove extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	// Replace dashes and underscores with spaces
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	return toTitleCase(name)
}

// toTitleCase converts a string to title case
func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// stripNavTitles removes the 'title' field from .nav.yml files
// The awesome-nav plugin warns that title has no effect at top level
func (p *Preprocessor) stripNavTitles() error {
	p.log("    Stripping titles from .nav.yml files...")

	stripped := 0

	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".nav.yml") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Simple regex to remove title line at start of file
		titlePattern := regexp.MustCompile(`(?m)^title:\s*.*\n?`)
		modified := titlePattern.ReplaceAllString(string(content), "")

		if modified != string(content) {
			if err := os.WriteFile(path, []byte(modified), 0644); err != nil {
				return err
			}
			stripped++
		}

		return nil
	})

	if err != nil {
		return err
	}

	p.log("    Stripped titles from %d .nav.yml files", stripped)
	return nil
}
