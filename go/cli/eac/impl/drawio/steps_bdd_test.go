// Package drawio contains godog step implementations for eac-cli.
//
// This file contains DrawIO diagram command step definitions.
// These specs require Docker for the drawio-tool container.
package drawio

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/core/paths"
	eacgodog "github.com/ready-to-release/eac/go/godog"
)

// drawioContext holds state for drawio tests.
type drawioContext struct {
	dockerAvailable bool
	lastDecoded     string
}

// registerSteps registers step definitions for drawio command features.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	dCtx := &drawioContext{}

	// Background steps
	sc.Step(`^the drawio-tool Docker image is available$`, func() error {
		return drawioEnsureImage(ctx, dCtx)
	})

	// Given steps - file creation
	sc.Step(`^a valid \.drawio\.png file "([^"]*)"$`, func(filename string) error {
		return drawioCreateTestFile(ctx, filename, "TestPage")
	})
	sc.Step(`^a \.drawio\.png file "([^"]*)" with page name "([^"]*)"$`, func(filename, pageName string) error {
		return drawioCreateTestFile(ctx, filename, pageName)
	})
	sc.Step(`^a file "([^"]*)" with valid mxGraphModel content$`, func(filename string) error {
		return drawioCreateXMLFile(ctx, filename, "")
	})
	sc.Step(`^a file "([^"]*)" with valid mxGraphModel content containing "([^"]*)"$`, func(filename, content string) error {
		return drawioCreateXMLFile(ctx, filename, content)
	})
	sc.Step(`^an encoded XML file "([^"]*)"$`, func(filename string) error {
		return drawioCreateEncodedXMLFile(ctx, filename, "")
	})
	sc.Step(`^an encoded XML file "([^"]*)" with different content$`, func(filename string) error {
		return drawioCreateEncodedXMLFile(ctx, filename, "ModifiedContent")
	})

	// Then steps - file verification
	sc.Step(`^the file "([^"]*)" should be a valid PNG$`, func(filename string) error {
		return drawioVerifyPNG(ctx, filename)
	})
	sc.Step(`^the file "([^"]*)" should contain "([^"]*)"$`, func(filename, text string) error {
		return eacgodog.FileContains(ctx, filename, text)
	})
	sc.Step(`^decoding "([^"]*)" should show the new content$`, func(filename string) error {
		return drawioDecodeAndVerify(ctx, dCtx, filename, "ModifiedContent")
	})
	sc.Step(`^decoding "([^"]*)" should contain "([^"]*)"$`, func(filename, text string) error {
		return drawioDecodeAndVerify(ctx, dCtx, filename, text)
	})
	sc.Step(`^decoding "([^"]*)" should contain '([^']*)'$`, func(filename, text string) error {
		return drawioDecodeAndVerify(ctx, dCtx, filename, text)
	})
}

// drawioEnsureImage builds the drawio-tool Docker image if needed.
func drawioEnsureImage(ctx *eacgodog.TestContext, dCtx *drawioContext) error {
	// Check if Docker is available
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not available: %w", err)
	}

	// Check if image exists
	cmd = exec.Command("docker", "image", "inspect", "cli-drawio-tool:latest")
	if err := cmd.Run(); err != nil {
		// Image doesn't exist, build it using ORIGINAL repo root (not isolated dir)
		repoRoot := ctx.OriginalRepoRoot
		dockerfilePath := filepath.Join(repoRoot, "containers", "drawio-tool", "Dockerfile")
		contextPath := filepath.Join(repoRoot, "containers", "drawio-tool")

		cmd = exec.Command("docker", "build", "-t", "cli-drawio-tool:latest", "-f", dockerfilePath, contextPath)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to build drawio-tool image: %w: %s", err, stderr.String())
		}
	}

	dCtx.dockerAvailable = true
	return nil
}

// drawioCreateTestFile creates a test .drawio.png file directly (no Docker needed).
func drawioCreateTestFile(ctx *eacgodog.TestContext, filename, pageName string) error {
	outputPath := eacgodog.ResolvePath(ctx, filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create minimal .drawio.png file with embedded XML
	// This is a 1x1 transparent PNG with mxfile metadata in a tEXt chunk
	pngData := createMinimalDrawioPNG(pageName)

	if err := os.WriteFile(outputPath, pngData, 0o644); err != nil {
		return fmt.Errorf("failed to write drawio.png: %w", err)
	}

	return nil
}

// drawioCreateXMLFile creates a test XML file with mxGraphModel content.
func drawioCreateXMLFile(ctx *eacgodog.TestContext, filename, extraContent string) error {
	content := fmt.Sprintf(`<mxfile host="test" agent="test">
  <diagram name="Test" id="test-diagram">
    <mxGraphModel dx="1426" dy="758" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="1654" pageHeight="1169" background="#CFCFCF" shadow="1">
      <root>
        <mxCell id="0"/>
        <mxCell id="1" parent="0"/>
        <mxCell id="test-cell" value="%s" style="rounded=1" vertex="1" parent="1">
          <mxGeometry x="100" y="100" width="100" height="50" as="geometry"/>
        </mxCell>
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>`, extraContent)

	return eacgodog.CreateFile(ctx, filename, content)
}

// drawioCreateEncodedXMLFile creates an encoded XML file (same as decoded for test purposes).
// The actual encoding is done by the drawio commands, not by the test setup.
func drawioCreateEncodedXMLFile(ctx *eacgodog.TestContext, filename, extraContent string) error {
	// For test purposes, we just create the decoded XML format
	// The commands under test will handle proper encoding
	return drawioCreateXMLFile(ctx, filename, extraContent)
}

// drawioVerifyPNG checks if a file is a valid PNG.
func drawioVerifyPNG(ctx *eacgodog.TestContext, filename string) error {
	path := eacgodog.ResolvePath(ctx, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Check PNG magic bytes
	if len(data) < 8 {
		return fmt.Errorf("file too small to be a PNG")
	}

	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if !bytes.Equal(data[:8], pngMagic) {
		return fmt.Errorf("file does not have PNG magic bytes")
	}

	return nil
}

// drawioDecodeAndVerify runs drawio decode command and checks for content in the decoded XML.
func drawioDecodeAndVerify(ctx *eacgodog.TestContext, dCtx *drawioContext, filename, expectedContent string) error {
	// Run the actual decode command to properly decode base64+deflate content
	// We need to run this without affecting the main command context state
	binaryPath := paths.CommandsBinaryPath(ctx.OriginalRepoRoot)

	cmd := exec.Command(binaryPath, "drawio", "decode", "--input", filename)

	// Set working directory to isolated dir if in isolation
	if ctx.IsolatedDir != "" {
		cmd.Dir = ctx.IsolatedDir
	}

	// Build environment with R2R_REPO_ROOT for proper path resolution
	env := os.Environ()
	if ctx.IsolatedDir != "" {
		env = append(env, fmt.Sprintf("R2R_REPO_ROOT=%s", ctx.IsolatedDir))
		env = append(env, fmt.Sprintf("R2R_PWD=%s", ctx.IsolatedDir))
	}
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("decode failed: %w: %s", err, stderr.String())
	}

	dCtx.lastDecoded = stdout.String()

	if !strings.Contains(dCtx.lastDecoded, expectedContent) {
		preview := dCtx.lastDecoded
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return fmt.Errorf("decoded content does not contain %q, got: %s", expectedContent, preview)
	}

	return nil
}

// createMinimalDrawioPNG creates a minimal .drawio.png file with embedded mxfile XML.
func createMinimalDrawioPNG(pageName string) []byte {
	// Minimal mxfile XML
	mxfile := fmt.Sprintf(`<mxfile host="test" agent="test">
  <diagram name="%s" id="test">
    <mxGraphModel dx="1426" dy="758" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="1654" pageHeight="1169" background="#CFCFCF" shadow="1">
      <root>
        <mxCell id="0"/>
        <mxCell id="1" parent="0"/>
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>`, pageName)

	return buildDrawioPNG(mxfile)
}

// buildDrawioPNG creates a PNG file with embedded mxfile in a tEXt chunk.
func buildDrawioPNG(mxfile string) []byte {
	var buf bytes.Buffer

	// PNG signature
	buf.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})

	// IHDR chunk (1x1 RGBA image)
	ihdr := []byte{
		0x00, 0x00, 0x00, 0x01, // width: 1
		0x00, 0x00, 0x00, 0x01, // height: 1
		0x08, // bit depth: 8
		0x06, // color type: RGBA
		0x00, // compression: deflate
		0x00, // filter: adaptive
		0x00, // interlace: none
	}
	writeChunk(&buf, "IHDR", ihdr)

	// tEXt chunk with mxfile
	textData := append([]byte("mxfile\x00"), []byte(mxfile)...)
	writeChunk(&buf, "tEXt", textData)

	// IDAT chunk (minimal compressed image data - 1 transparent pixel)
	// Pre-computed deflate of filter byte + 4 zero bytes (transparent RGBA)
	idat := []byte{0x78, 0x9c, 0x62, 0x60, 0x60, 0x60, 0x60, 0x00, 0x00, 0x00, 0x05, 0x00, 0x01}
	writeChunk(&buf, "IDAT", idat)

	// IEND chunk
	writeChunk(&buf, "IEND", nil)

	return buf.Bytes()
}

// writeChunk writes a PNG chunk to the buffer.
func writeChunk(buf *bytes.Buffer, chunkType string, data []byte) {
	// Length (4 bytes, big-endian)
	length := uint32(len(data)) //nolint:gosec // PNG chunk data is always small
	buf.WriteByte(byte(length >> 24))
	buf.WriteByte(byte(length >> 16))
	buf.WriteByte(byte(length >> 8))
	buf.WriteByte(byte(length))

	// Type (4 bytes)
	buf.WriteString(chunkType)

	// Data
	buf.Write(data)

	// CRC (4 bytes) - CRC32 of type + data
	crcData := append([]byte(chunkType), data...)
	crc := crc32Checksum(crcData)
	buf.WriteByte(byte(crc >> 24))
	buf.WriteByte(byte(crc >> 16))
	buf.WriteByte(byte(crc >> 8))
	buf.WriteByte(byte(crc))
}

// crc32Checksum computes CRC32 for PNG chunks.
func crc32Checksum(data []byte) uint32 {
	// PNG uses CRC-32 with polynomial 0xedb88320
	var crcTable [256]uint32
	for i := 0; i < 256; i++ {
		crc := uint32(i) //nolint:gosec // Loop index is always 0-255
		for j := 0; j < 8; j++ {
			if crc&1 != 0 {
				crc = 0xedb88320 ^ (crc >> 1)
			} else {
				crc >>= 1
			}
		}
		crcTable[i] = crc
	}

	crc := uint32(0xffffffff)
	for _, b := range data {
		crc = crcTable[(crc^uint32(b))&0xff] ^ (crc >> 8)
	}
	return crc ^ 0xffffffff
}
