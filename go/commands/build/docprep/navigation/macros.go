package navigation

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/commands/build/docprep/staging"
)

// DiataxisSections are the four documentation sections that should have navigation macros.
var DiataxisSections = map[string]bool{
	"tutorials":     true,
	"how-to-guides": true,
	"explanation":   true,
	"reference":     true,
}

// navRegex consolidates all regex patterns used across the navigation package.
// Previously these were spread across macros.go and index.go as separate package-level vars.
var navRegex = struct {
	// Macro patterns
	Macro      *regexp.Regexp // matches {{ macro_name() }} or {{ macro_name(args) }}
	Breadcrumb *regexp.Regexp // matches {{ page_breadcrumb() }}
	Footer     *regexp.Regexp // matches {{ diataxis_footer() }}

	// Navigation patterns
	NavTitle        *regexp.Regexp // matches title line in .nav.yml files
	H1Title         *regexp.Regexp // matches first H1 heading in markdown
	FrontmatterTitle *regexp.Regexp // matches title in YAML frontmatter with optional quotes
}{
	Macro:            regexp.MustCompile(`\{\{\s*\w+\([^)]*\)\s*\}\}`),
	Breadcrumb:       regexp.MustCompile(`\{\{\s*page_breadcrumb\(\)\s*\}\}`),
	Footer:           regexp.MustCompile(`\{\{\s*diataxis_footer\(\)\s*\}\}`),
	NavTitle:         regexp.MustCompile(`(?m)^title:\s*.*\n?`),
	H1Title:          regexp.MustCompile(`(?m)^#\s+(.+)$`),
	FrontmatterTitle: regexp.MustCompile(`(?m)^title:\s*["']?([^"'\n]+)["']?`),
}

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
		modified := navRegex.Macro.ReplaceAllString(original, "")

		if modified != original {
			matches := navRegex.Macro.FindAllString(original, -1)
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

		modified := navRegex.NavTitle.ReplaceAllString(string(content), "")

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

		if !navRegex.Breadcrumb.MatchString(modified) {
			loc := navRegex.H1Title.FindStringIndex(modified)
			if loc != nil {
				titleEnd := loc[1]
				modified = modified[:titleEnd] + "\n\n{{ page_breadcrumb() }}" + modified[titleEnd:]
				macrosAdded++
			}
		}

		if !navRegex.Footer.MatchString(modified) {
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
