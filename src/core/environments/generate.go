// +build ignore

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	// This generator copies environment contracts from the repo config directory
	// to be embedded in the binary. The path is relative because this runs as
	// a standalone script via go:generate from this file's location.
	// See contracts.EACConfigRelPath for the canonical path constant.

	src := filepath.Join("..", "..", "..", ".r2r", "eac", "repository", "environments")
	dst := filepath.Join("repository", "environments")

	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create directory %s: %v\n", dst, err)
		os.Exit(1)
	}

	// Find all .yml files in source
	matches, err := filepath.Glob(filepath.Join(src, "*.yml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to glob files: %v\n", err)
		os.Exit(1)
	}

	// Copy each file
	for _, srcFile := range matches {
		dstFile := filepath.Join(dst, filepath.Base(srcFile))
		if err := copyFile(srcFile, dstFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to copy %s to %s: %v\n", srcFile, dstFile, err)
			os.Exit(1)
		}
		fmt.Printf("Copied %s -> %s\n", srcFile, dstFile)
	}
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
