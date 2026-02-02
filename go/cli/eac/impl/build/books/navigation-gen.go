package books

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/core/paths"
	"gopkg.in/yaml.v3"
)

// ensureNavigationStructure ensures .nav.yml exists and is valid in all staging directories
// This function now validates existing .nav.yml files and cleans up broken references.
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

// generateNavForDir creates .nav.yml for a directory.
func (p *Preprocessor) generateNavForDir(dir string) error {
	relDir, relErr := filepath.Rel(p.stagingDir, dir)
	if relErr != nil || relDir == "." {
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
			relPath, relErr := filepath.Rel(p.stagingDir, filePath)
			if relErr != nil {
				relPath = filePath // fallback
			}
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
	if err := os.WriteFile(navPath, data, 0o644); err != nil {
		return err
	}

	genRelPath, genRelErr := filepath.Rel(p.stagingDir, dir)
	if genRelErr != nil || genRelPath == "." {
		genRelPath = "(root)"
	}
	p.log("    Generated: %s/.nav.yml", genRelPath)

	return nil
}

// hasMarkdownContent checks if a directory has any markdown files.
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

// parseNavFile parses a .nav.yml file into structured format.
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

// scanMarkdownFiles returns a map of markdown files in a directory (filename -> true).
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

// scanSubdirectories returns a map of subdirectories with markdown content (dirname -> true).
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

// sortNewFiles sorts new files intelligently using order from command sources, then alphabetically.
func (p *Preprocessor) sortNewFiles(files []string, dirPath string) []string {
	type fileWithOrder struct {
		name  string
		order int
	}

	filesWithOrder := make([]fileWithOrder, 0, len(files))

	for _, file := range files {
		filePath := filepath.Join(dirPath, file)
		relPath, relErr := filepath.Rel(p.stagingDir, filePath)
		if relErr != nil {
			relPath = filePath // fallback
		}
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
