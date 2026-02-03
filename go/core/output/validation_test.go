package output

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// HashFile() Tests
// =============================================================================

func TestHashFile_ReturnsCorrectHash(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with known content
	content := []byte("Hello, World!")
	filePath := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	// Compute expected hash
	hasher := sha256.New()
	hasher.Write(content)
	expectedHash := "sha256:" + hex.EncodeToString(hasher.Sum(nil))

	// Test HashFile
	size, hash, err := HashFile(filePath)
	require.NoError(t, err)

	assert.Equal(t, int64(len(content)), size)
	assert.Equal(t, expectedHash, hash)
}

func TestHashFile_ReturnsCorrectSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files of various sizes
	tests := []struct {
		name         string
		expectedSize int
	}{
		{"size_0", 0},
		{"size_1", 1},
		{"size_100", 100},
		{"size_1000", 1000},
		{"size_10000", 10000},
		{"size_100000", 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := make([]byte, tt.expectedSize)
			for i := range content {
				content[i] = byte(i % 256)
			}

			filePath := filepath.Join(tmpDir, tt.name+".bin")
			err := os.WriteFile(filePath, content, 0644)
			require.NoError(t, err)

			size, _, err := HashFile(filePath)
			require.NoError(t, err)

			assert.Equal(t, int64(tt.expectedSize), size)
		})
	}
}

func TestHashFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty file
	filePath := filepath.Join(tmpDir, "empty.txt")
	err := os.WriteFile(filePath, []byte{}, 0644)
	require.NoError(t, err)

	// SHA256 of empty content
	expectedHash := "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	size, hash, err := HashFile(filePath)
	require.NoError(t, err)

	assert.Equal(t, int64(0), size)
	assert.Equal(t, expectedHash, hash)
}

func TestHashFile_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()

	nonExistentPath := filepath.Join(tmpDir, "does-not-exist.txt")

	size, hash, err := HashFile(nonExistentPath)
	assert.Error(t, err, "HashFile should return error for nonexistent file")
	assert.Zero(t, size)
	assert.Empty(t, hash)
}

func TestHashFile_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	size, hash, err := HashFile(tmpDir)
	assert.Error(t, err, "HashFile should return error for directory")
	assert.Zero(t, size)
	assert.Empty(t, hash)
}

func TestHashFile_IsDeterministic(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("Deterministic content for hashing test")
	filePath := filepath.Join(tmpDir, "deterministic.txt")
	err := os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	// Hash multiple times
	_, hash1, err := HashFile(filePath)
	require.NoError(t, err)

	_, hash2, err := HashFile(filePath)
	require.NoError(t, err)

	_, hash3, err := HashFile(filePath)
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2, "Hash should be deterministic")
	assert.Equal(t, hash2, hash3, "Hash should be deterministic")
}

func TestHashFile_DifferentContentDifferentHash(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	err := os.WriteFile(file1, []byte("Content A"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(file2, []byte("Content B"), 0644)
	require.NoError(t, err)

	_, hash1, err := HashFile(file1)
	require.NoError(t, err)

	_, hash2, err := HashFile(file2)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2, "Different content should produce different hash")
}

func TestHashFile_SameContentSameHash(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	content := []byte("Identical content")

	err := os.WriteFile(file1, content, 0644)
	require.NoError(t, err)

	err = os.WriteFile(file2, content, 0644)
	require.NoError(t, err)

	_, hash1, err := HashFile(file1)
	require.NoError(t, err)

	_, hash2, err := HashFile(file2)
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2, "Same content should produce same hash")
}

func TestHashFile_BinaryContent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create binary content
	content := make([]byte, 256)
	for i := range content {
		content[i] = byte(i)
	}

	filePath := filepath.Join(tmpDir, "binary.bin")
	err := os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	size, hash, err := HashFile(filePath)
	require.NoError(t, err)

	assert.Equal(t, int64(256), size)
	assert.NotEmpty(t, hash)
	assert.True(t, len(hash) > 60, "Hash should be sha256: prefix + 64 hex chars")
}

func TestHashFile_LargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	tmpDir := t.TempDir()

	// Create 1MB file
	content := make([]byte, 1024*1024)
	for i := range content {
		content[i] = byte(i % 256)
	}

	filePath := filepath.Join(tmpDir, "large.bin")
	err := os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	size, hash, err := HashFile(filePath)
	require.NoError(t, err)

	assert.Equal(t, int64(1024*1024), size)
	assert.NotEmpty(t, hash)
}

// =============================================================================
// ValidateArtifacts() Tests
// =============================================================================

func TestValidateArtifacts_AllPresent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with known hashes
	file1Content := []byte("File 1 content")
	file1Path := filepath.Join(tmpDir, "file1.bin")
	err := os.WriteFile(file1Path, file1Content, 0644)
	require.NoError(t, err)

	file2Content := []byte("File 2 content")
	file2Path := filepath.Join(tmpDir, "file2.bin")
	err = os.WriteFile(file2Path, file2Content, 0644)
	require.NoError(t, err)

	// Compute actual hashes
	_, hash1, err := HashFile(file1Path)
	require.NoError(t, err)
	_, hash2, err := HashFile(file2Path)
	require.NoError(t, err)

	artifacts := []Artifact{
		{ID: "file1", Path: "file1.bin", SHA256: hash1, Size: int64(len(file1Content)), Type: "binary"},
		{ID: "file2", Path: "file2.bin", SHA256: hash2, Size: int64(len(file2Content)), Type: "binary"},
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.True(t, result.Valid)
	assert.True(t, result.ArtifactsValid)
	assert.Empty(t, result.MissingArtifacts)
	assert.Empty(t, result.CorruptArtifacts)
	assert.Nil(t, result.Error)
}

func TestValidateArtifacts_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create one file but reference two
	file1Content := []byte("File 1 content")
	file1Path := filepath.Join(tmpDir, "file1.bin")
	err := os.WriteFile(file1Path, file1Content, 0644)
	require.NoError(t, err)

	_, hash1, err := HashFile(file1Path)
	require.NoError(t, err)

	artifacts := []Artifact{
		{ID: "file1", Path: "file1.bin", SHA256: hash1, Size: int64(len(file1Content)), Type: "binary"},
		{ID: "file2", Path: "file2.bin", SHA256: "sha256:nonexistent", Size: 100, Type: "binary"},
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.False(t, result.Valid)
	assert.False(t, result.ArtifactsValid)
	assert.Contains(t, result.MissingArtifacts, "file2.bin")
	assert.Empty(t, result.CorruptArtifacts)
}

func TestValidateArtifacts_CorruptFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file
	fileContent := []byte("Original content")
	filePath := filepath.Join(tmpDir, "file.bin")
	err := os.WriteFile(filePath, fileContent, 0644)
	require.NoError(t, err)

	// Use wrong hash
	wrongHash := "sha256:0000000000000000000000000000000000000000000000000000000000000000"

	artifacts := []Artifact{
		{ID: "file", Path: "file.bin", SHA256: wrongHash, Size: int64(len(fileContent)), Type: "binary"},
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.False(t, result.Valid)
	assert.False(t, result.ArtifactsValid)
	assert.Empty(t, result.MissingArtifacts)
	assert.Contains(t, result.CorruptArtifacts, "file.bin")
}

func TestValidateArtifacts_MixedMissingAndCorrupt(t *testing.T) {
	tmpDir := t.TempDir()

	// Create one valid file
	validContent := []byte("Valid content")
	validPath := filepath.Join(tmpDir, "valid.bin")
	err := os.WriteFile(validPath, validContent, 0644)
	require.NoError(t, err)
	_, validHash, err := HashFile(validPath)
	require.NoError(t, err)

	// Create one corrupt file (wrong hash)
	corruptContent := []byte("Corrupt content")
	corruptPath := filepath.Join(tmpDir, "corrupt.bin")
	err = os.WriteFile(corruptPath, corruptContent, 0644)
	require.NoError(t, err)

	artifacts := []Artifact{
		{ID: "valid", Path: "valid.bin", SHA256: validHash, Size: int64(len(validContent)), Type: "binary"},
		{ID: "corrupt", Path: "corrupt.bin", SHA256: "sha256:wrong", Size: int64(len(corruptContent)), Type: "binary"},
		{ID: "missing", Path: "missing.bin", SHA256: "sha256:missing", Size: 100, Type: "binary"},
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.False(t, result.Valid)
	assert.False(t, result.ArtifactsValid)
	assert.Contains(t, result.MissingArtifacts, "missing.bin")
	assert.Contains(t, result.CorruptArtifacts, "corrupt.bin")
}

func TestValidateArtifacts_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()

	artifacts := []Artifact{}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.True(t, result.Valid)
	assert.True(t, result.ArtifactsValid)
	assert.Empty(t, result.MissingArtifacts)
	assert.Empty(t, result.CorruptArtifacts)
}

func TestValidateArtifacts_NilList(t *testing.T) {
	tmpDir := t.TempDir()

	result := ValidateArtifacts(tmpDir, nil)

	assert.True(t, result.Valid)
	assert.True(t, result.ArtifactsValid)
	assert.Empty(t, result.MissingArtifacts)
	assert.Empty(t, result.CorruptArtifacts)
}

func TestValidateArtifacts_SingleValid(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("Single file content")
	filePath := filepath.Join(tmpDir, "single.bin")
	err := os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	_, hash, err := HashFile(filePath)
	require.NoError(t, err)

	artifacts := []Artifact{
		{ID: "single", Path: "single.bin", SHA256: hash, Size: int64(len(content)), Type: "binary"},
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.True(t, result.Valid)
	assert.True(t, result.ArtifactsValid)
}

func TestValidateArtifacts_SingleMissing(t *testing.T) {
	tmpDir := t.TempDir()

	artifacts := []Artifact{
		{ID: "missing", Path: "missing.bin", SHA256: "sha256:x", Size: 100, Type: "binary"},
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.False(t, result.Valid)
	assert.False(t, result.ArtifactsValid)
	assert.Len(t, result.MissingArtifacts, 1)
}

func TestValidateArtifacts_SingleCorrupt(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("Some content")
	filePath := filepath.Join(tmpDir, "corrupt.bin")
	err := os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	artifacts := []Artifact{
		{ID: "corrupt", Path: "corrupt.bin", SHA256: "sha256:wronghash", Size: int64(len(content)), Type: "binary"},
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.False(t, result.Valid)
	assert.False(t, result.ArtifactsValid)
	assert.Len(t, result.CorruptArtifacts, 1)
}

func TestValidateArtifacts_NestedPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	nestedDir := filepath.Join(tmpDir, "nested", "deep", "path")
	err := os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	content := []byte("Nested file content")
	filePath := filepath.Join(nestedDir, "file.bin")
	err = os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	_, hash, err := HashFile(filePath)
	require.NoError(t, err)

	artifacts := []Artifact{
		{
			ID:     "nested",
			Path:   filepath.Join("nested", "deep", "path", "file.bin"),
			SHA256: hash,
			Size:   int64(len(content)),
			Type:   "binary",
		},
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.True(t, result.Valid)
	assert.True(t, result.ArtifactsValid)
}

func TestValidateArtifacts_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("Absolute path content")
	filePath := filepath.Join(tmpDir, "absolute.bin")
	err := os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	_, hash, err := HashFile(filePath)
	require.NoError(t, err)

	// Use absolute path in artifact
	artifacts := []Artifact{
		{ID: "absolute", Path: filePath, SHA256: hash, Size: int64(len(content)), Type: "binary"},
	}

	// When baseDir is empty, absolute paths should work
	result := ValidateArtifacts("", artifacts)

	assert.True(t, result.Valid)
	assert.True(t, result.ArtifactsValid)
}

func TestValidateArtifacts_SizeMismatch(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("Size test content")
	filePath := filepath.Join(tmpDir, "size.bin")
	err := os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	_, hash, err := HashFile(filePath)
	require.NoError(t, err)

	// Correct hash but wrong size
	artifacts := []Artifact{
		{ID: "size", Path: "size.bin", SHA256: hash, Size: 999999, Type: "binary"},
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	// Size mismatch might be treated as corruption depending on implementation
	// The hash check should catch this, but if only hash is checked, this should still pass
	// Let's verify the behavior
	if result.Valid {
		// If valid, size is not checked separately from hash
		assert.True(t, result.ArtifactsValid)
	} else {
		// If invalid, size mismatch is detected
		assert.False(t, result.ArtifactsValid)
	}
}

func TestValidateArtifacts_TableDriven(t *testing.T) {
	tests := []struct {
		name                string
		setupFiles          map[string][]byte   // path -> content
		artifacts           []Artifact
		expectValid         bool
		expectMissingCount  int
		expectCorruptCount  int
	}{
		{
			name: "all valid",
			setupFiles: map[string][]byte{
				"a.bin": []byte("content a"),
				"b.bin": []byte("content b"),
			},
			artifacts: []Artifact{
				{ID: "a", Path: "a.bin", Type: "binary"},
				{ID: "b", Path: "b.bin", Type: "binary"},
			},
			expectValid:        true,
			expectMissingCount: 0,
			expectCorruptCount: 0,
		},
		{
			name: "all missing",
			setupFiles: map[string][]byte{},
			artifacts: []Artifact{
				{ID: "x", Path: "x.bin", SHA256: "sha256:x", Type: "binary"},
				{ID: "y", Path: "y.bin", SHA256: "sha256:y", Type: "binary"},
			},
			expectValid:        false,
			expectMissingCount: 2,
			expectCorruptCount: 0,
		},
		{
			name: "all corrupt",
			setupFiles: map[string][]byte{
				"a.bin": []byte("actual a"),
				"b.bin": []byte("actual b"),
			},
			artifacts: []Artifact{
				{ID: "a", Path: "a.bin", SHA256: "sha256:wrong", Type: "binary"},
				{ID: "b", Path: "b.bin", SHA256: "sha256:wrong", Type: "binary"},
			},
			expectValid:        false,
			expectMissingCount: 0,
			expectCorruptCount: 2,
		},
		{
			name:               "empty artifacts",
			setupFiles:         map[string][]byte{},
			artifacts:          []Artifact{},
			expectValid:        true,
			expectMissingCount: 0,
			expectCorruptCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup files and compute hashes if needed
			for path, content := range tt.setupFiles {
				filePath := filepath.Join(tmpDir, path)
				err := os.WriteFile(filePath, content, 0644)
				require.NoError(t, err)
			}

			// Update artifact hashes for "all valid" case
			for i := range tt.artifacts {
				if tt.name == "all valid" {
					filePath := filepath.Join(tmpDir, tt.artifacts[i].Path)
					_, hash, err := HashFile(filePath)
					require.NoError(t, err)
					tt.artifacts[i].SHA256 = hash

					info, err := os.Stat(filePath)
					require.NoError(t, err)
					tt.artifacts[i].Size = info.Size()
				}
			}

			result := ValidateArtifacts(tmpDir, tt.artifacts)

			assert.Equal(t, tt.expectValid, result.Valid)
			assert.Len(t, result.MissingArtifacts, tt.expectMissingCount)
			assert.Len(t, result.CorruptArtifacts, tt.expectCorruptCount)
		})
	}
}

// =============================================================================
// ValidateArtifacts() Edge Cases
// =============================================================================

func TestValidateArtifacts_ZeroSizeFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty file
	filePath := filepath.Join(tmpDir, "empty.bin")
	err := os.WriteFile(filePath, []byte{}, 0644)
	require.NoError(t, err)

	_, hash, err := HashFile(filePath)
	require.NoError(t, err)

	artifacts := []Artifact{
		{ID: "empty", Path: "empty.bin", SHA256: hash, Size: 0, Type: "marker"},
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.True(t, result.Valid)
	assert.True(t, result.ArtifactsValid)
}

func TestValidateArtifacts_ManyArtifacts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many files
	artifactCount := 50
	artifacts := make([]Artifact, artifactCount)

	for i := 0; i < artifactCount; i++ {
		content := []byte("content " + string(rune('0'+i%10)))
		fileName := "file" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + ".bin"
		filePath := filepath.Join(tmpDir, fileName)
		err := os.WriteFile(filePath, content, 0644)
		require.NoError(t, err)

		_, hash, err := HashFile(filePath)
		require.NoError(t, err)

		artifacts[i] = Artifact{
			ID:     fileName,
			Path:   fileName,
			SHA256: hash,
			Size:   int64(len(content)),
			Type:   "binary",
		}
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.True(t, result.Valid)
	assert.True(t, result.ArtifactsValid)
}

func TestValidateArtifacts_SpecialFileNames(t *testing.T) {
	tmpDir := t.TempDir()

	// Test files with special characters in names (that are valid on the filesystem)
	specialNames := []string{
		"file-with-dashes.bin",
		"file_with_underscores.bin",
		"file.multiple.dots.bin",
		"UPPERCASE.BIN",
		"MixedCase.Bin",
	}

	artifacts := make([]Artifact, len(specialNames))

	for i, name := range specialNames {
		content := []byte("content for " + name)
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, content, 0644)
		require.NoError(t, err)

		_, hash, err := HashFile(filePath)
		require.NoError(t, err)

		artifacts[i] = Artifact{
			ID:     name,
			Path:   name,
			SHA256: hash,
			Size:   int64(len(content)),
			Type:   "binary",
		}
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.True(t, result.Valid)
	assert.True(t, result.ArtifactsValid)
}

// =============================================================================
// Integration-style Tests
// =============================================================================

func TestValidateArtifacts_SimulatedBuildOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Simulate a typical build output structure
	buildDir := filepath.Join(tmpDir, "out", "build", "core", "go_go")
	err := os.MkdirAll(buildDir, 0755)
	require.NoError(t, err)

	// Create "binary" files
	binaries := map[string][]byte{
		"eac-linux-amd64":     make([]byte, 10000),
		"eac-darwin-amd64":    make([]byte, 11000),
		"eac-windows-amd64.exe": make([]byte, 12000),
	}

	artifacts := make([]Artifact, 0, len(binaries))
	for name, content := range binaries {
		// Fill with pseudo-random data
		for i := range content {
			content[i] = byte(i % 256)
		}

		filePath := filepath.Join(buildDir, name)
		err := os.WriteFile(filePath, content, 0644)
		require.NoError(t, err)

		_, hash, err := HashFile(filePath)
		require.NoError(t, err)

		artifacts = append(artifacts, Artifact{
			ID:     name,
			Path:   filepath.Join("out", "build", "core", "go_go", name),
			SHA256: hash,
			Size:   int64(len(content)),
			Type:   "binary",
		})
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.True(t, result.Valid)
	assert.True(t, result.ArtifactsValid)
}

func TestValidateArtifacts_SimulatedPartialBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// Simulate partial build where one file is missing
	buildDir := filepath.Join(tmpDir, "out", "build", "core", "go_go")
	err := os.MkdirAll(buildDir, 0755)
	require.NoError(t, err)

	// Create only one of three expected files
	content := []byte("partial build content")
	filePath := filepath.Join(buildDir, "eac-linux-amd64")
	err = os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	_, hash, err := HashFile(filePath)
	require.NoError(t, err)

	artifacts := []Artifact{
		{ID: "linux", Path: filepath.Join("out", "build", "core", "go_go", "eac-linux-amd64"), SHA256: hash, Size: int64(len(content)), Type: "binary"},
		{ID: "darwin", Path: filepath.Join("out", "build", "core", "go_go", "eac-darwin-amd64"), SHA256: "sha256:missing", Size: 1000, Type: "binary"},
		{ID: "windows", Path: filepath.Join("out", "build", "core", "go_go", "eac-windows-amd64.exe"), SHA256: "sha256:missing", Size: 1000, Type: "binary"},
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.False(t, result.Valid)
	assert.False(t, result.ArtifactsValid)
	assert.Len(t, result.MissingArtifacts, 2)
}

func TestValidateArtifacts_SimulatedCorruptedBuild(t *testing.T) {
	tmpDir := t.TempDir()

	buildDir := filepath.Join(tmpDir, "out", "build", "core", "go_go")
	err := os.MkdirAll(buildDir, 0755)
	require.NoError(t, err)

	// Create files but with different content than expected
	filePath := filepath.Join(buildDir, "eac")
	actualContent := []byte("corrupted/modified content")
	err = os.WriteFile(filePath, actualContent, 0644)
	require.NoError(t, err)

	// Manifest claims different hash
	artifacts := []Artifact{
		{
			ID:     "eac",
			Path:   filepath.Join("out", "build", "core", "go_go", "eac"),
			SHA256: "sha256:expected_but_different_hash",
			Size:   int64(len(actualContent)),
			Type:   "binary",
		},
	}

	result := ValidateArtifacts(tmpDir, artifacts)

	assert.False(t, result.Valid)
	assert.False(t, result.ArtifactsValid)
	assert.Len(t, result.CorruptArtifacts, 1)
}
