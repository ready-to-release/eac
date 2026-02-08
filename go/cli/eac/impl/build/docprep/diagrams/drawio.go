package diagrams

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MaxImageWidthPDF is the maximum width for images in PDF output.
// A4 at 150 DPI is about 1240px wide, so 1200px leaves some margin.
const MaxImageWidthPDF = 1200

// DrawioImage represents a discovered drawio.png image in the docs.
type DrawioImage struct {
	// SourceFile is the absolute path to the drawio.png file
	SourceFile string
	// RelPath is the relative path from docs directory
	RelPath string
	// Hash is the SHA256 hash of the source file content
	Hash string
}

// FindDrawioImages scans the docs directory for all .drawio.png files.
// Returns a list of DrawioImage with source path, relative path, and content hash.
func FindDrawioImages(docsDir string) ([]DrawioImage, error) {
	var images []DrawioImage

	err := filepath.WalkDir(docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".drawio.png") {
			return nil
		}

		hash, err := HashFileContent(path)
		if err != nil {
			return fmt.Errorf("hashing %s: %w", path, err)
		}

		relPath, err := filepath.Rel(docsDir, path)
		if err != nil {
			relPath = path
		}

		images = append(images, DrawioImage{
			SourceFile: path,
			RelPath:    relPath,
			Hash:       hash,
		})

		return nil
	})

	return images, err
}

// HashFileContent returns the SHA256 hash of a file's content.
func HashFileContent(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
