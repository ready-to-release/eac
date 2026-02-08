package navigation

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/staging"
)

// DiataxisSections are the four documentation sections that should have navigation macros.
var DiataxisSections = map[string]bool{
	"tutorials":     true,
	"how-to-guides": true,
	"explanation":   true,
	"reference":     true,
}

// macroPattern matches {{ macro_name() }} or {{ macro_name(args) }}.
var macroPattern = regexp.MustCompile(`\{\{\s*\w+\([^)]*\)\s*\}\}`)

// navTitlePattern matches title line in .nav.yml files.
var navTitlePattern = regexp.MustCompile(`(?m)^title:\s*.*\n?`)

// Patterns for macro injection.
var (
	macroH1TitlePattern = regexp.MustCompile(`(?m)^#\s+.+$`)
	breadcrumbPattern   = regexp.MustCompile(`\{\{\s*page_breadcrumb\(\)\s*\}\}`)
	footerPattern       = regexp.MustCompile(`\{\{\s*diataxis_footer\(\)\s*\}\}`)
)

// StripMacros removes Jinja2 macro calls from markdown files (PDF only).
func StripMacros(fileIndex *staging.FileIndex, logf func(string, ...any)) error {
	logf("    Stripping macros from markdown files...")

	stripped := 0
	filesModified := 0

	for _, path := range fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := macroPattern.ReplaceAllString(original, "")

		if modified != original {
			matches := macroPattern.FindAllString(original, -1)
			stripped += len(matches)
			filesModified++

			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
		}
	}

	logf("    Stripped %d macros from %d files", stripped, filesModified)
	return nil
}

// StripNavTitles removes the 'title' field from .nav.yml files.
func StripNavTitles(fileIndex *staging.FileIndex, logf func(string, ...any)) error {
	logf("    Stripping titles from .nav.yml files...")

	stripped := 0

	for _, path := range fileIndex.GetAllFiles() {
		if !strings.HasSuffix(path, ".nav.yml") {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		modified := navTitlePattern.ReplaceAllString(string(content), "")

		if modified != string(content) {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			stripped++
		}
	}

	logf("    Stripped titles from %d .nav.yml files", stripped)
	return nil
}

// InjectMacros adds Jinja2 navigation macros to markdown files (site builds only).
func InjectMacros(fileIndex *staging.FileIndex, stagingDir string, logf func(string, ...any)) error {
	logf("    Injecting navigation macros into markdown files...")

	injected := 0
	filesModified := 0

	for _, path := range fileIndex.GetMarkdownFiles() {
		relPath, err := filepath.Rel(stagingDir, path)
		if err != nil {
			return err
		}

		relPath = filepath.ToSlash(relPath)
		parts := strings.Split(relPath, "/")
		if len(parts) == 0 {
			continue
		}

		section := parts[0]
		if !DiataxisSections[section] {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := original
		macrosAdded := 0

		if !breadcrumbPattern.MatchString(modified) {
			loc := macroH1TitlePattern.FindStringIndex(modified)
			if loc != nil {
				titleEnd := loc[1]
				modified = modified[:titleEnd] + "\n\n{{ page_breadcrumb() }}" + modified[titleEnd:]
				macrosAdded++
			}
		}

		if !footerPattern.MatchString(modified) {
			modified = strings.TrimRight(modified, "\n\r\t ") + "\n\n{{ diataxis_footer() }}\n"
			macrosAdded++
		}

		if modified != original {
			injected += macrosAdded
			filesModified++

			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
		}
	}

	logf("    Injected %d macros into %d files", injected, filesModified)
	return nil
}
