package books

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"gopkg.in/yaml.v3"
)

var navLog = logging.C()

// NavFile represents a .nav.yml file structure
type NavFile struct {
	Title string `yaml:"title,omitempty"`
	Nav   []any  `yaml:"nav"`
}

// ensureRootIndex ensures an index.md exists at staging root
// If no index.md exists, generates one with book metadata and TOC
// If index.md exists (from copy source), keeps it as-is
func (p *Preprocessor) ensureRootIndex() error {
	indexPath := paths.IndexMarkdownPath(p.stagingDir)

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
// Uses .nav.yml order when available to preserve source navigation structure
func (p *Preprocessor) collectTOCEntries(dir string, depth int) []tocEntry {
	var entries []tocEntry

	// DEBUG: Log directory scanning
	relDir, _ := filepath.Rel(p.stagingDir, dir)
	if relDir == "." {
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

// collectTOCFromNav collects TOC entries using .nav.yml order
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
					relPath, _ := filepath.Rel(p.stagingDir, subdir)
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
					relPath, _ := filepath.Rel(p.stagingDir, filePath)
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
					relPath, _ := filepath.Rel(p.stagingDir, subdir)
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
						relPath, _ := filepath.Rel(p.stagingDir, filePath)
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

// collectTOCFromFilesystem collects TOC entries from filesystem (fallback)
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

// ensureNavigationStructure ensures .nav.yml exists and is valid in all staging directories
// This function now validates existing .nav.yml files and cleans up broken references
func (p *Preprocessor) ensureNavigationStructure() error {
	validated := 0
	generated := 0

	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		// Skip hidden directories and assets (but not the staging root itself)
		if path != p.stagingDir {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "assets" {
				return filepath.SkipDir
			}
		}

		// Check if .nav.yml exists
		navPath := paths.NavigationConfigPath(path)
		if _, err := os.Stat(navPath); err == nil {
			// EXISTS: Validate and clean
			if err := p.validateAndCleanNav(navPath, path); err != nil {
				return err
			}
			validated++
		} else if os.IsNotExist(err) {
			// MISSING: Generate new
			if err := p.generateNavForDir(path); err != nil {
				return err
			}
			generated++
		} else {
			return err // Real error, not just "not found"
		}

		return nil
	})

	if err != nil {
		return err
	}

	p.log("    Validated: %d .nav.yml files", validated)
	p.log("    Generated: %d .nav.yml files", generated)

	return nil
}

// generateNavForDir creates .nav.yml for a directory
func (p *Preprocessor) generateNavForDir(dir string) error {
	relDir, _ := filepath.Rel(p.stagingDir, dir)
	if relDir == "." {
		relDir = "(root)"
	}

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

	// Don't set title - let the nav system auto-generate from markdown files
	// The awesome-nav plugin extracts titles from frontmatter or first H1
	navFile := NavFile{
		Nav: nav,
	}

	data, err := yaml.Marshal(navFile)
	if err != nil {
		return err
	}

	navPath := paths.NavigationConfigPath(dir)
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

// parseNavFile parses a .nav.yml file into structured format
func (p *Preprocessor) parseNavFile(navPath string) (*NavFile, error) {
	data, err := os.ReadFile(navPath)
	if err != nil {
		return nil, err
	}

	var navFile NavFile
	if err := yaml.Unmarshal(data, &navFile); err != nil {
		return nil, err
	}

	return &navFile, nil
}

// scanMarkdownFiles returns a map of markdown files in a directory (filename -> true)
func (p *Preprocessor) scanMarkdownFiles(dir string) map[string]bool {
	files := make(map[string]bool)

	items, err := os.ReadDir(dir)
	if err != nil {
		return files
	}

	for _, item := range items {
		if !item.IsDir() && strings.HasSuffix(item.Name(), ".md") {
			files[item.Name()] = true
		}
	}

	return files
}

// scanSubdirectories returns a map of subdirectories with markdown content (dirname -> true)
func (p *Preprocessor) scanSubdirectories(dir string) map[string]bool {
	dirs := make(map[string]bool)

	items, err := os.ReadDir(dir)
	if err != nil {
		return dirs
	}

	for _, item := range items {
		if item.IsDir() && !strings.HasPrefix(item.Name(), ".") && item.Name() != "assets" {
			// Only include if it has markdown content
			subdir := filepath.Join(dir, item.Name())
			if hasMarkdownContent(subdir) {
				dirs[item.Name()] = true
			}
		}
	}

	return dirs
}

// sortNewFiles sorts new files intelligently using order from command sources, then alphabetically
func (p *Preprocessor) sortNewFiles(files []string, dirPath string) []string {
	type fileWithOrder struct {
		name  string
		order int
	}

	filesWithOrder := make([]fileWithOrder, 0, len(files))

	for _, file := range files {
		filePath := filepath.Join(dirPath, file)
		relPath, _ := filepath.Rel(p.stagingDir, filePath)
		relPath = filepath.ToSlash(relPath)

		order := p.getOrderForFile(relPath)
		filesWithOrder = append(filesWithOrder, fileWithOrder{
			name:  file,
			order: order,
		})
	}

	// Sort by order, then alphabetically
	sort.Slice(filesWithOrder, func(i, j int) bool {
		if filesWithOrder[i].order != filesWithOrder[j].order {
			return filesWithOrder[i].order < filesWithOrder[j].order
		}
		return filesWithOrder[i].name < filesWithOrder[j].name
	})

	result := make([]string, len(filesWithOrder))
	for i, f := range filesWithOrder {
		result[i] = f.name
	}

	return result
}

// validateNavEntry validates a single nav entry recursively
// Returns (validatedEntry, referencedFiles) or (nil, empty) if invalid
func (p *Preprocessor) validateNavEntry(entry any, dirPath string, actualFiles, actualDirs map[string]bool) (any, map[string]bool) {
	referenced := make(map[string]bool)

	switch v := entry.(type) {
	case string:
		// Simple file or directory reference: "modules.md" or "subdir/" or "subdir/file.md"
		if strings.HasSuffix(v, "/") {
			// Directory reference
			dirName := strings.TrimSuffix(v, "/")
			if actualDirs[dirName] {
				referenced[dirName+"/"] = true
				return v, referenced // Valid directory
			}
			// For relative paths (containing /), check filesystem
			if strings.Contains(dirName, "/") {
				fullPath := filepath.Join(dirPath, dirName)
				if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
					referenced[dirName+"/"] = true
					return v, referenced // Valid directory
				}
			}
			// Directory missing, skip
			return nil, referenced
		} else if strings.HasSuffix(v, ".md") {
			// File reference
			if actualFiles[v] {
				referenced[v] = true
				return v, referenced // Valid file
			}
			// For relative paths (containing /), check filesystem
			if strings.Contains(v, "/") {
				fullPath := filepath.Join(dirPath, v)
				navLog.Debugf("[VALIDATE] Checking relative file: %s -> %s", v, fullPath)
				if _, err := os.Stat(fullPath); err == nil {
					navLog.Debugf("[VALIDATE] ✓ File exists: %s", v)
					referenced[v] = true
					return v, referenced // Valid file
				} else {
					navLog.Debugf("[VALIDATE] ✗ File missing: %s (error: %v)", v, err)
				}
			}
			// File missing, skip
			return nil, referenced
		}
		// Unknown format, keep as-is
		return v, referenced

	case map[string]any:
		// Titled section: {"Section Name": [...]} or {"Title": "file.md"}
		validatedMap := make(map[string]any)

		for title, content := range v {
			switch c := content.(type) {
			case string:
				// {"Title": "file.md"}
				if strings.HasSuffix(c, ".md") {
					if actualFiles[c] {
						validatedMap[title] = c
						referenced[c] = true
					} else if strings.Contains(c, "/") {
						// For relative paths, check filesystem
						fullPath := filepath.Join(dirPath, c)
						if _, err := os.Stat(fullPath); err == nil {
							validatedMap[title] = c
							referenced[c] = true
						}
					}
				} else if strings.HasSuffix(c, "/") {
					// {"Title": "subdir/"}
					dirName := strings.TrimSuffix(c, "/")
					if actualDirs[dirName] {
						validatedMap[title] = c
						referenced[dirName+"/"] = true
					} else if strings.Contains(dirName, "/") {
						// For relative paths, check filesystem
						fullPath := filepath.Join(dirPath, dirName)
						if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
							validatedMap[title] = c
							referenced[dirName+"/"] = true
						}
					}
				} else {
					// {"Title": "subdir"} - directory without trailing slash
					// awesome-nav allows both formats
					if actualDirs[c] {
						validatedMap[title] = c
						referenced[c+"/"] = true
					} else if strings.Contains(c, "/") {
						// For relative paths, check filesystem
						fullPath := filepath.Join(dirPath, c)
						if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
							validatedMap[title] = c
							referenced[c+"/"] = true
						}
					}
				}
				// If file/dir missing, omit this entry

			case []any:
				// {"Section": [...items...]}
				validatedItems := []any{}
				for _, item := range c {
					validated, refs := p.validateNavEntry(item, dirPath, actualFiles, actualDirs)
					if validated != nil {
						validatedItems = append(validatedItems, validated)
						for ref := range refs {
							referenced[ref] = true
						}
					}
				}

				// Only include section if it has valid items
				if len(validatedItems) > 0 {
					validatedMap[title] = validatedItems
				}
			}
		}

		if len(validatedMap) > 0 {
			return validatedMap, referenced
		}
		return nil, referenced
	}

	return entry, referenced // Unknown type, keep as-is
}

// validateAndCleanNav validates existing .nav.yml, removes broken refs, adds new files
func (p *Preprocessor) validateAndCleanNav(navPath, dirPath string) error {
	// 1. Parse existing .nav.yml
	navFile, err := p.parseNavFile(navPath)
	if err != nil {
		// If parse fails, regenerate from scratch
		relPath, _ := filepath.Rel(p.stagingDir, navPath)
		p.log("    ⚠️  Failed to parse %s, regenerating", relPath)
		return p.generateNavForDir(dirPath)
	}

	// 2. Scan directory for actual files
	actualFiles := p.scanMarkdownFiles(dirPath)
	actualDirs := p.scanSubdirectories(dirPath)

	// 3. Process nav entries
	validatedNav := make([]any, 0)
	referencedFiles := make(map[string]bool)

	for _, entry := range navFile.Nav {
		validated, referenced := p.validateNavEntry(entry, dirPath, actualFiles, actualDirs)
		if validated != nil {
			validatedNav = append(validatedNav, validated)
			// Track what files are referenced
			for file := range referenced {
				referencedFiles[file] = true
			}
		}
	}

	// 4. Find unreferenced files (new from command sources or other additions)
	newFiles := []string{}
	for file := range actualFiles {
		if !referencedFiles[file] {
			newFiles = append(newFiles, file)
		}
	}

	// 5. Add new files with intelligent ordering
	if len(newFiles) > 0 {
		sortedNew := p.sortNewFiles(newFiles, dirPath)
		for _, file := range sortedNew {
			validatedNav = append(validatedNav, file)
		}

		relDir, _ := filepath.Rel(p.stagingDir, dirPath)
		if relDir == "." {
			relDir = "(root)"
		}
		p.log("    Added %d new files to %s/.nav.yml", len(newFiles), relDir)
	}

	// 6. Write updated .nav.yml
	// Preserve title if set - awesome-nav uses it for navigation display
	navFile.Nav = validatedNav
	data, err := yaml.Marshal(navFile)
	if err != nil {
		return err
	}

	return os.WriteFile(navPath, data, 0644)
}

// stripMacros removes Jinja2 macro calls from markdown files (PDF only)
// The macros plugin is only enabled for site builds, not PDF builds
// Common macros: {{ diataxis_footer() }}, {{ page_breadcrumb() }}
func (p *Preprocessor) stripMacros() error {
	p.log("    Stripping macros from markdown files...")

	// Pattern matches {{ macro_name() }} or {{ macro_name(args) }}
	macroPattern := regexp.MustCompile(`\{\{\s*\w+\([^)]*\)\s*\}\}`)

	stripped := 0
	filesModified := 0

	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := macroPattern.ReplaceAllString(original, "")

		if modified != original {
			// Count how many macros were stripped
			matches := macroPattern.FindAllString(original, -1)
			stripped += len(matches)
			filesModified++

			if err := os.WriteFile(path, []byte(modified), 0644); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	p.log("    Stripped %d macros from %d files", stripped, filesModified)
	return nil
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

// diátaxisSections are the four documentation sections that should have navigation macros
var diataxisSections = map[string]bool{
	"tutorials":     true,
	"how-to-guides": true,
	"explanation":   true,
	"reference":     true,
}

// injectMacros adds Jinja2 navigation macros to markdown files (site builds only)
// Injects {{ page_breadcrumb() }} after the title and {{ diataxis_footer() }} at the end
// Only processes files in Diátaxis sections (tutorials, how-to-guides, explanation, reference)
func (p *Preprocessor) injectMacros() error {
	p.log("    Injecting navigation macros into markdown files...")

	// Pattern to find the first H1 title line
	titlePattern := regexp.MustCompile(`(?m)^#\s+.+$`)

	// Pattern to check if macros already exist (for idempotency)
	breadcrumbPattern := regexp.MustCompile(`\{\{\s*page_breadcrumb\(\)\s*\}\}`)
	footerPattern := regexp.MustCompile(`\{\{\s*diataxis_footer\(\)\s*\}\}`)

	injected := 0
	filesModified := 0

	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Get relative path from staging root to determine section
		relPath, err := filepath.Rel(p.stagingDir, path)
		if err != nil {
			return err
		}

		// Normalize path separators and get first directory component
		relPath = filepath.ToSlash(relPath)
		parts := strings.Split(relPath, "/")
		if len(parts) == 0 {
			return nil
		}

		// Check if file is in a Diátaxis section
		section := parts[0]
		if !diataxisSections[section] {
			return nil // Skip files not in a Diátaxis section
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := original
		macrosAdded := 0

		// Check if breadcrumb already exists
		if !breadcrumbPattern.MatchString(modified) {
			// Find first H1 title and insert breadcrumb after it
			loc := titlePattern.FindStringIndex(modified)
			if loc != nil {
				titleEnd := loc[1]
				// Insert breadcrumb macro after the title line
				modified = modified[:titleEnd] + "\n\n{{ page_breadcrumb() }}" + modified[titleEnd:]
				macrosAdded++
			}
		}

		// Check if footer already exists
		if !footerPattern.MatchString(modified) {
			// Append footer at end of file
			modified = strings.TrimRight(modified, "\n\r\t ") + "\n\n{{ diataxis_footer() }}\n"
			macrosAdded++
		}

		if modified != original {
			injected += macrosAdded
			filesModified++

			if err := os.WriteFile(path, []byte(modified), 0644); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	p.log("    Injected %d macros into %d files", injected, filesModified)
	return nil
}
