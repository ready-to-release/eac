package cleanup

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/commands/build/docprep/staging"
)

// LinkWithContextPattern matches markdown links with optional preceding character.
// This allows detecting and skipping image links (preceded by !).
// Captures: [1]=preceding char or empty, [2]=link text, [3]=path.
var LinkWithContextPattern = regexp.MustCompile(`(^|[^!])\[([^\]]+)\]\(([^)]+)\)`)

// FixBrokenInternalLinks converts broken internal links to site URLs.
func FixBrokenInternalLinks(
	fileIndex *staging.FileIndex,
	stagingDir, siteURL string,
	logf func(string, ...any),
) error {
	if siteURL == "" {
		logf("    Skipping link fix (no site_url configured)")
		return nil
	}

	logf("    Fixing broken internal links...")

	if !strings.HasSuffix(siteURL, "/") {
		siteURL += "/"
	}

	linksFixed := 0
	mdFiles := fileIndex.GetMarkdownFiles()

	for _, path := range mdFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fileDir := filepath.Dir(path)
		relFileDir, relErr := filepath.Rel(stagingDir, fileDir)
		if relErr != nil {
			relFileDir = fileDir
		}

		original := string(content)
		modified, count := FixLinksInContent(original, relFileDir, stagingDir, siteURL)

		if count > 0 {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			linksFixed += count
		}
	}

	logf("    Processed %d files, fixed %d broken links", len(mdFiles), linksFixed)
	return nil
}

// FixLinksInContent processes a single file's content and fixes broken links.
func FixLinksInContent(content, relFileDir, stagingDir, siteURL string) (string, int) {
	count := 0

	result := LinkWithContextPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := LinkWithContextPattern.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}

		prefix := parts[1]
		text := parts[2]
		linkPath := parts[3]

		if strings.HasPrefix(linkPath, "http://") ||
			strings.HasPrefix(linkPath, "https://") ||
			strings.HasPrefix(linkPath, "mailto:") ||
			strings.HasPrefix(linkPath, "//") {
			return match
		}

		if strings.HasPrefix(linkPath, "#") {
			return match
		}

		pathPart := linkPath
		anchor := ""
		if anchorIdx := strings.Index(linkPath, "#"); anchorIdx >= 0 {
			pathPart = linkPath[:anchorIdx]
			anchor = linkPath[anchorIdx:]
		}

		var targetPath string
		if strings.HasPrefix(pathPart, "/") {
			targetPath = filepath.Join(stagingDir, pathPart)
		} else {
			targetPath = filepath.Join(stagingDir, relFileDir, pathPart)
		}

		targetPath = filepath.Clean(targetPath)

		ext := strings.ToLower(filepath.Ext(pathPart))
		isNonMarkdownFile := ext != "" && ext != ".md"

		if !isNonMarkdownFile {
			if _, err := os.Stat(targetPath); err == nil {
				return match
			}

			if !strings.HasSuffix(targetPath, ".md") {
				if _, err := os.Stat(targetPath + ".md"); err == nil {
					return match
				}
				if _, err := os.Stat(filepath.Join(targetPath, "index.md")); err == nil {
					return match
				}
			}
		}

		urlPath := pathPart
		if strings.HasPrefix(urlPath, "/") {
			urlPath = urlPath[1:]
		} else {
			urlPath = filepath.ToSlash(filepath.Join(relFileDir, pathPart))
		}

		urlPath = filepath.ToSlash(filepath.Clean(urlPath))
		if strings.HasPrefix(urlPath, "../") {
			urlPath = filepath.Base(pathPart)
		}

		urlPath = strings.TrimSuffix(urlPath, ".md")

		fullURL, err := url.JoinPath(siteURL, urlPath)
		if err != nil {
			return match
		}

		count++
		return fmt.Sprintf("%s[%s](%s%s)", prefix, text, fullURL, anchor)
	})

	return result, count
}
