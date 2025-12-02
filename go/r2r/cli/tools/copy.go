// +build ignore

package main

import (
	"io"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 3 {
		panic("Usage: go run copy.go <source> <dest>")
	}

	src := os.Args[1]
	dst := os.Args[2]

	// Create destination directory if needed
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		panic(err)
	}

	// Copy file
	sourceFile, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		panic(err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		panic(err)
	}
}
