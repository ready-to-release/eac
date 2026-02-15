package linking

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/core/paths"
)

// LinkTranslation holds path translations for a single file.
type LinkTranslation struct {
	SourceFile   string            // Original source file path (absolute)
	StagingFile  string            // Staging file path (absolute)
	Translations map[string]string // old_path -> new_path
}

// TranslationMap manages link translations for all files.
type TranslationMap struct {
	SourceRoot   string                     // Source root directory (workspace)
	StagingRoot  string                     // Staging root directory
	FileMap      map[string]string          // staging_path -> source_path (absolute)
	DocsRelPath  map[string]string          // source_path -> docs-relative path
	Translations map[string]*LinkTranslation // staging_path -> translation
	StripExternal bool                      // True if building PDF (strips external links)
	Logf         func(string, ...any)       // Logger
}

// NewTranslationMap creates a new link translation map.
func NewTranslationMap(sourceRoot, stagingRoot string, stripExternal bool, logf func(string, ...any)) *TranslationMap {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	return &TranslationMap{
		SourceRoot:    filepath.Clean(sourceRoot),
		StagingRoot:   filepath.Clean(stagingRoot),
		FileMap:       make(map[string]string),
		DocsRelPath:   make(map[string]string),
		Translations:  make(map[string]*LinkTranslation),
		StripExternal: stripExternal,
		Logf:          logf,
	}
}

// AddFileMapping tracks a source -> staging file mapping.
func (m *TranslationMap) AddFileMapping(stagingPath, sourcePath string) {
	m.FileMap[stagingPath] = sourcePath

	docsDir := paths.DocsSourcePath(m.SourceRoot)
	if relPath, err := filepath.Rel(docsDir, sourcePath); err == nil {
		m.DocsRelPath[sourcePath] = filepath.ToSlash(relPath)
	}
}

// GetSourcePath returns the source path for a staging path, if tracked.
func (m *TranslationMap) GetSourcePath(stagingPath string) (string, bool) {
	src, ok := m.FileMap[stagingPath]
	return src, ok
}

// BuildTranslations analyzes all source files and builds translation maps.
func (m *TranslationMap) BuildTranslations(siteURL string) error {
	sourceToStaging := make(map[string]string)
	for stagingPath, sourcePath := range m.FileMap {
		sourceToStaging[sourcePath] = stagingPath
	}

	for stagingFile, sourceFile := range m.FileMap {
		if !strings.HasSuffix(strings.ToLower(sourceFile), ".md") {
			continue
		}

		content, err := os.ReadFile(sourceFile)
		if err != nil {
			return fmt.Errorf("reading source file %s: %w", sourceFile, err)
		}

		links := ExtractRelativeLinks(string(content))

		trans := &LinkTranslation{
			SourceFile:   sourceFile,
			StagingFile:  stagingFile,
			Translations: make(map[string]string),
		}

		sourceDir := filepath.Dir(sourceFile)
		stagingDir := filepath.Dir(stagingFile)

		for _, oldLink := range links {
			if strings.Contains(oldLink, "{{") && strings.Contains(oldLink, "}}") {
				continue
			}

			pathPart := oldLink
			anchor := ""
			if idx := strings.Index(oldLink, "#"); idx >= 0 {
				pathPart = oldLink[:idx]
				anchor = oldLink[idx:]
			}

			absSourcePath := filepath.Clean(filepath.Join(sourceDir, pathPart))

			var newLink string

			targetStagingFile, existsInStaging := sourceToStaging[absSourcePath]

			if !existsInStaging {
				docsDir := paths.DocsSourcePath(m.SourceRoot)

				targetExists := false
				if _, err := os.Stat(absSourcePath); err == nil {
					targetExists = true
				} else {
					if _, err := os.Stat(absSourcePath + ".md"); err == nil {
						targetExists = true
					} else {
						indexPath := filepath.Join(absSourcePath, "index.md")
						if _, err := os.Stat(indexPath); err == nil {
							targetExists = true
						}
					}
				}

				if targetExists && strings.HasPrefix(absSourcePath, docsDir) {
					if m.StripExternal {
						newLink = "##EXTERNAL_STRIPPED##" + oldLink
					} else {
						if siteURL == "" {
							return fmt.Errorf("link '%s' in %s points to file not in staging, but no site_url configured for external links",
								oldLink, sourceFile)
						}

						relPath, err := filepath.Rel(docsDir, absSourcePath)
						if err != nil {
							return fmt.Errorf("calculating docs-relative path for %s: %w", absSourcePath, err)
						}

						relPath = filepath.ToSlash(relPath)
						relPath = strings.TrimSuffix(relPath, ".md")

						baseURL := siteURL
						if !strings.HasSuffix(baseURL, "/") {
							baseURL += "/"
						}
						newLink = baseURL + relPath + anchor
					}
				} else {
					return fmt.Errorf("BROKEN LINK in %s: '%s' points to '%s' which doesn't exist in source",
						sourceFile, oldLink, absSourcePath)
				}
			} else {
				relPath, err := filepath.Rel(stagingDir, targetStagingFile)
				if err != nil {
					return fmt.Errorf("calculating relative path from %s to %s: %w",
						stagingDir, targetStagingFile, err)
				}
				newLink = filepath.ToSlash(relPath) + anchor
			}

			if newLink != oldLink {
				trans.Translations[oldLink] = newLink
			}
		}

		if len(trans.Translations) > 0 {
			m.Translations[stagingFile] = trans
		}
	}

	return nil
}

// ApplyAllTranslations applies path translations to all staging files.
func (m *TranslationMap) ApplyAllTranslations() error {
	for stagingFile, trans := range m.Translations {
		if err := m.applyTranslation(stagingFile, trans); err != nil {
			return fmt.Errorf("applying translations to %s: %w", stagingFile, err)
		}
	}
	return nil
}

func (m *TranslationMap) applyTranslation(stagingFile string, trans *LinkTranslation) error {
	content, err := os.ReadFile(stagingFile)
	if err != nil {
		return fmt.Errorf("reading staging file: %w", err)
	}

	modified := string(content)

	type pathPair struct {
		old string
		new string
	}
	var pairs []pathPair
	for old, new := range trans.Translations {
		pairs = append(pairs, pathPair{old: old, new: new})
	}

	// Sort by length (longest first) to avoid partial replacements
	sort.Slice(pairs, func(i, j int) bool {
		return len(pairs[i].old) > len(pairs[j].old)
	})

	for _, pair := range pairs {
		if strings.HasPrefix(pair.new, "##EXTERNAL_STRIPPED##") {
			modified = ReplaceMarkdownLinks(modified, pair.old, true)
		} else {
			modified = ReplaceOutsideCodeBlocks(modified, pair.old, pair.new)
		}
	}

	if modified != string(content) {
		if err := os.WriteFile(stagingFile, []byte(modified), 0o644); err != nil {
			return fmt.Errorf("writing staging file: %w", err)
		}
	}

	return nil
}

// CalculateRelativePath calculates the relative path from a staging markdown file
// to a target absolute path.
func CalculateRelativePath(stagingFile, targetAbsPath string) (string, error) {
	absStaging, err := filepath.Abs(stagingFile)
	if err != nil {
		return "", fmt.Errorf("making staging path absolute: %w", err)
	}

	absTarget, err := filepath.Abs(targetAbsPath)
	if err != nil {
		return "", fmt.Errorf("making target path absolute: %w", err)
	}

	stagingDir := filepath.Dir(absStaging)
	relPath, err := filepath.Rel(stagingDir, absTarget)
	if err != nil {
		return "", fmt.Errorf("calculating relative path: %w", err)
	}

	return filepath.ToSlash(relPath), nil
}
