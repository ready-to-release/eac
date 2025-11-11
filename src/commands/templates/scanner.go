package templates

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

// PlaceholderInfo contains information about a placeholder and where it's used
type PlaceholderInfo struct {
	Name  string
	Files []string
}

// PlaceholderScanner scans template files and extracts placeholder variables
type PlaceholderScanner struct {
	templateDir  string
	placeholders map[string]map[string]bool // placeholder -> set of files
}

// NewPlaceholderScanner creates a new scanner for the given template directory
func NewPlaceholderScanner(templateDir string) *PlaceholderScanner {
	return &PlaceholderScanner{
		templateDir:  templateDir,
		placeholders: make(map[string]map[string]bool),
	}
}

// Scan walks the template directory and extracts all placeholder variables
func (s *PlaceholderScanner) Scan() ([]string, error) {
	infos, err := s.ScanWithLocations()
	if err != nil {
		return nil, err
	}

	// Extract just the names for backward compatibility
	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name
	}
	return names, nil
}

// ScanWithLocations walks the template directory and extracts all placeholders with their file locations
func (s *PlaceholderScanner) ScanWithLocations() ([]PlaceholderInfo, error) {
	// Verify template directory exists
	info, err := os.Stat(s.templateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template directory does not exist: %s", s.templateDir)
		}
		return nil, fmt.Errorf("cannot access template directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("template path is not a directory: %s", s.templateDir)
	}

	// Walk the directory and scan files
	err = filepath.WalkDir(s.templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Scan file for placeholders
		return s.scanFile(path)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan templates: %w", err)
	}

	// Convert map to sorted slice of PlaceholderInfo
	placeholderInfos := make([]PlaceholderInfo, 0, len(s.placeholders))
	for placeholder, filesMap := range s.placeholders {
		// Convert file set to sorted slice
		files := make([]string, 0, len(filesMap))
		for file := range filesMap {
			files = append(files, file)
		}
		sort.Strings(files)

		placeholderInfos = append(placeholderInfos, PlaceholderInfo{
			Name:  placeholder,
			Files: files,
		})
	}

	// Sort by placeholder name
	sort.Slice(placeholderInfos, func(i, j int) bool {
		return placeholderInfos[i].Name < placeholderInfos[j].Name
	})

	return placeholderInfos, nil
}

// scanFile extracts placeholders from a single file
func (s *PlaceholderScanner) scanFile(filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Get relative path for display
	relPath, err := filepath.Rel(s.templateDir, filePath)
	if err != nil {
		relPath = filePath
	}

	// Extract placeholders from content
	s.extractPlaceholders(string(content), relPath)

	// Also extract from file path (file/directory names can have placeholders)
	s.extractPlaceholders(relPath, relPath)

	return nil
}

// extractPlaceholders finds all {{ .VariableName }} patterns in text
func (s *PlaceholderScanner) extractPlaceholders(text string, filePath string) {
	// Regular expression to match {{ .VariableName }}
	// Matches: {{ .Name }}, {{.Name}}, {{ .Some_Name123 }}
	re := regexp.MustCompile(`\{\{\s*\.([A-Za-z][A-Za-z0-9_]*)\s*\}\}`)

	matches := re.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) > 1 {
			// match[0] is the full match, match[1] is the captured variable name
			placeholder := match[1]

			// Initialize file set if needed
			if s.placeholders[placeholder] == nil {
				s.placeholders[placeholder] = make(map[string]bool)
			}

			// Add file to this placeholder's set
			s.placeholders[placeholder][filePath] = true
		}
	}
}

// GetPlaceholderCount returns the number of unique placeholders found
func (s *PlaceholderScanner) GetPlaceholderCount() int {
	return len(s.placeholders)
}
