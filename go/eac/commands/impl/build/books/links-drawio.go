package books

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func (p *Preprocessor) convertDrawioToLinks() error {
	if p.book.SiteURL == "" {
		p.log("    Skipping drawio conversion (no site_url configured)")
		return nil
	}

	p.log("    Converting .drawio images to links...")

	siteURL := p.book.SiteURL
	if !strings.HasSuffix(siteURL, "/") {
		siteURL += "/"
	}

	processed := 0
	converted := 0

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

		fileDir := filepath.Dir(path)
		relFileDir, relErr := filepath.Rel(p.stagingDir, fileDir)
		if relErr != nil {
			relFileDir = fileDir // fallback to absolute
		}

		original := string(content)
		modified, count := convertDrawioImages(original, relFileDir, siteURL)

		if count > 0 {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			converted += count
		}
		processed++
		return nil
	})
	if err != nil {
		return err
	}

	p.log("    Processed %d files, converted %d .drawio images to links", processed, converted)
	return nil
}

// drawioImagePattern matches markdown images pointing to .drawio files.
var drawioImagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]*\.drawio)\)`)

// convertDrawioImages converts ![alt](path.drawio) to [View interactive diagram](url).
func convertDrawioImages(content, relFileDir, siteURL string) (string, int) {
	count := 0

	result := drawioImagePattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := drawioImagePattern.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		// alt := parts[1] // Not used in link text
		imgPath := parts[2]

		// Build URL path
		var urlPath string
		if strings.HasPrefix(imgPath, "/") {
			urlPath = imgPath[1:]
		} else {
			urlPath = filepath.ToSlash(filepath.Join(relFileDir, imgPath))
		}
		urlPath = filepath.ToSlash(filepath.Clean(urlPath))

		// Handle paths that go outside staging
		// For explanation book, assets are at root level
		// ../assets/x -> assets/x
		urlPath = strings.TrimPrefix(urlPath, "../")

		fullURL, err := url.JoinPath(siteURL, urlPath)
		if err != nil {
			return match
		}

		count++
		return fmt.Sprintf("[View interactive diagram](%s)", fullURL)
	})

	return result, count
}

// fixBrokenInternalLinks converts broken internal links to GitHub Pages URLs
// This is needed because copy-in may not include all linked files, breaking relative links.
//
// For each markdown link [text](path):
//  1. If path is external (http/https/mailto), skip
//  2. If path is an anchor-only link (#section), skip
//  3. Resolve the path relative to the current file
//  4. Check if the target exists in staging
//  5. If not, convert to absolute GitHub Pages URL
