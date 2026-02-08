package navigation

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
	"gopkg.in/yaml.v3"
)

// NavFile represents a .nav.yml file structure.
type NavFile struct {
	Title string `yaml:"title,omitempty"`
	Nav   []any  `yaml:"nav"`
}

// EnsureNavigationStructure ensures .nav.yml exists and is valid in all staging directories.
func EnsureNavigationStructure(
	book *config.Book,
	stagingDir string,
	logf func(string, ...any),
) error {
	validated := 0
	generated := 0

	err := filepath.WalkDir(stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		if path != stagingDir {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "assets" {
				return filepath.SkipDir
			}
		}

		navPath := paths.NavigationConfigPath(path)
		if _, err := os.Stat(navPath); err == nil {
			if err := ValidateAndCleanNav(book, stagingDir, navPath, path, logf); err != nil {
				return err
			}
			validated++
		} else if os.IsNotExist(err) {
			if err := GenerateNavForDir(book, stagingDir, path, logf); err != nil {
				return err
			}
			generated++
		} else {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	logf("    Validated: %d .nav.yml files", validated)
	logf("    Generated: %d .nav.yml files", generated)

	return nil
}

// GenerateNavForDir creates .nav.yml for a directory.
func GenerateNavForDir(
	book *config.Book,
	stagingDir, dir string,
	logf func(string, ...any),
) error {
	items, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	type navItem struct {
		name  string
		order int
		isDir bool
	}

	var navItems []navItem

	for _, item := range items {
		name := item.Name()

		if strings.HasPrefix(name, ".") {
			continue
		}

		if item.IsDir() {
			if name == "assets" {
				continue
			}
			subdir := filepath.Join(dir, name)
			if HasMarkdownContent(subdir) {
				navItems = append(navItems, navItem{
					name:  name + "/",
					order: 1000,
					isDir: true,
				})
			}
		} else if strings.HasSuffix(name, ".md") {
			filePath := filepath.Join(dir, name)
			relPath, relErr := filepath.Rel(stagingDir, filePath)
			if relErr != nil {
				relPath = filePath
			}
			relPath = filepath.ToSlash(relPath)

			order := GetOrderForFile(book, relPath)
			navItems = append(navItems, navItem{
				name:  name,
				order: order,
				isDir: false,
			})
		}
	}

	if len(navItems) == 0 {
		return nil
	}

	sort.Slice(navItems, func(i, j int) bool {
		if navItems[i].name == "index.md" {
			return true
		}
		if navItems[j].name == "index.md" {
			return false
		}

		if !navItems[i].isDir && navItems[j].isDir {
			return true
		}
		if navItems[i].isDir && !navItems[j].isDir {
			return false
		}

		if navItems[i].order != navItems[j].order {
			return navItems[i].order < navItems[j].order
		}

		return navItems[i].name < navItems[j].name
	})

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

	navPath := paths.NavigationConfigPath(dir)
	if err := os.WriteFile(navPath, data, 0o644); err != nil {
		return err
	}

	genRelPath, genRelErr := filepath.Rel(stagingDir, dir)
	if genRelErr != nil || genRelPath == "." {
		genRelPath = "(root)"
	}
	logf("    Generated: %s/.nav.yml", genRelPath)

	return nil
}

// HasMarkdownContent checks if a directory has any markdown files.
func HasMarkdownContent(dir string) bool {
	items, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, item := range items {
		if strings.HasSuffix(item.Name(), ".md") {
			return true
		}
		if item.IsDir() && !strings.HasPrefix(item.Name(), ".") && item.Name() != "assets" {
			if HasMarkdownContent(filepath.Join(dir, item.Name())) {
				return true
			}
		}
	}
	return false
}

// ParseNavFile parses a .nav.yml file into structured format.
func ParseNavFile(navPath string) (*NavFile, error) {
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

// ScanMarkdownFiles returns a map of markdown files in a directory.
func ScanMarkdownFiles(dir string) map[string]bool {
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

// ScanSubdirectories returns a map of subdirectories with markdown content.
func ScanSubdirectories(dir string) map[string]bool {
	dirs := make(map[string]bool)

	items, err := os.ReadDir(dir)
	if err != nil {
		return dirs
	}

	for _, item := range items {
		if item.IsDir() && !strings.HasPrefix(item.Name(), ".") && item.Name() != "assets" {
			subdir := filepath.Join(dir, item.Name())
			if HasMarkdownContent(subdir) {
				dirs[item.Name()] = true
			}
		}
	}

	return dirs
}

// SortNewFiles sorts new files intelligently using order from command sources.
func SortNewFiles(book *config.Book, stagingDir string, files []string, dirPath string) []string {
	type fileWithOrder struct {
		name  string
		order int
	}

	filesWithOrder := make([]fileWithOrder, 0, len(files))

	for _, file := range files {
		filePath := filepath.Join(dirPath, file)
		relPath, relErr := filepath.Rel(stagingDir, filePath)
		if relErr != nil {
			relPath = filePath
		}
		relPath = filepath.ToSlash(relPath)

		order := GetOrderForFile(book, relPath)
		filesWithOrder = append(filesWithOrder, fileWithOrder{
			name:  file,
			order: order,
		})
	}

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

// GetTitleFromFile extracts title from markdown frontmatter or first heading.
func GetTitleFromFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return FilenameToTitle(filepath.Base(path))
	}

	text := string(content)

	if strings.HasPrefix(text, "---") {
		endIdx := strings.Index(text[3:], "---")
		if endIdx > 0 {
			frontmatter := text[3 : 3+endIdx]
			titleMatch := frontmatterTitlePattern.FindStringSubmatch(frontmatter)
			if len(titleMatch) > 1 {
				return strings.TrimSpace(titleMatch[1])
			}
		}
	}

	h1Match := h1HeadingPattern.FindStringSubmatch(text)
	if len(h1Match) > 1 {
		return strings.TrimSpace(h1Match[1])
	}

	return FilenameToTitle(filepath.Base(path))
}

// GetOrderForFile returns the order for a file from command sources.
func GetOrderForFile(book *config.Book, relPath string) int {
	for _, src := range book.Sources {
		if src.Type == "command" && filepath.ToSlash(src.Target) == relPath {
			if src.Order > 0 {
				return src.Order
			}
		}
	}
	return 500
}

// FilenameToTitle converts a filename to a readable title.
func FilenameToTitle(filename string) string {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	return ToTitleCase(name)
}

// ToTitleCase converts a string to title case.
func ToTitleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if word != "" {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}
