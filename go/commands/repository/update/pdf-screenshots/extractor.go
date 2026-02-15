package pdfscreenshots

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ready-to-release/eac/go/adapters/docker"
	"github.com/ready-to-release/eac/go/core/paths"
)

// ensurePDFToolsImage builds the pdf-oci Docker image if needed.
func ensurePDFToolsImage(repoRoot string) error {
	// Check if image exists
	cmd := exec.Command("docker", "image", "inspect", pdfToolsImage)
	if err := cmd.Run(); err == nil {
		return nil // Image exists
	}

	// Build the image
	dockerfilePath := paths.ContainerDockerfilePath(repoRoot, "pdf-cli-oci")
	buildCtx := paths.ContainersPath(repoRoot, "pdf-cli-oci")

	fmt.Println("Building pdf-cli-oci image...")
	cmd = exec.Command("docker", "build",
		"-t", pdfToolsImage,
		"-f", dockerfilePath,
		buildCtx)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// extractPages uses pdftoppm to extract PDF pages as PNG images.
func extractPages(_ docker.DockerClient, pdfPath, outputDir string, dpi int) error {
	pdfDir := filepath.Dir(pdfPath)
	pdfName := filepath.Base(pdfPath)

	ctx := context.Background()

	runConfig := &docker.RunConfig{
		Image: pdfToolsImage,
		Command: []string{
			"pdftoppm",
			"-png",
			"-r", fmt.Sprintf("%d", dpi),
			"/input/" + pdfName,
			"/output/page",
		},
		Mounts: []docker.MountConfig{
			{Source: pdfDir, Target: "/input", ReadOnly: true},
			{Source: outputDir, Target: "/output", ReadOnly: false},
		},
		WorkingDir:    "/workspace",
		Timeout:       10 * time.Minute,
		ContainerName: fmt.Sprintf("pdf-extract-%d", time.Now().UnixNano()),
	}

	result, err := docker.RunContainer(ctx, runConfig)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		logs := result.Stderr
		if logs == "" {
			logs = result.Stdout
		}
		return fmt.Errorf("container exited with code %d: %s", result.ExitCode, logs)
	}

	// pdftoppm outputs files like page-1.png, page-2.png
	// Rename to zero-padded format: page-01.png, page-02.png
	return renameToZeroPadded(outputDir)
}

// renameToZeroPadded renames page-1.png to page-01.png etc for proper sorting.
func renameToZeroPadded(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Find max page number for padding
	maxPage := 0
	pagePattern := rePageNumber

	for _, entry := range entries {
		matches := pagePattern.FindStringSubmatch(entry.Name())
		if matches != nil {
			if num, err := strconv.Atoi(matches[1]); err == nil && num > maxPage {
				maxPage = num
			}
		}
	}

	// Determine padding width
	padWidth := 2
	if maxPage >= 100 {
		padWidth = 3
	}
	if maxPage >= 1000 {
		padWidth = 4
	}

	// Rename files
	for _, entry := range entries {
		matches := pagePattern.FindStringSubmatch(entry.Name())
		if matches != nil {
			num, convErr := strconv.Atoi(matches[1])
			if convErr != nil {
				continue
			}
			newName := fmt.Sprintf("page-%0*d.png", padWidth, num)
			if entry.Name() != newName {
				oldPath := filepath.Join(dir, entry.Name())
				newPath := filepath.Join(dir, newName)
				if err := os.Rename(oldPath, newPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
