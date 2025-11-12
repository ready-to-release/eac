package testing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverGoTestTags_L0BuildTag(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test file with L0 build tag
	testFile := filepath.Join(tmpDir, "service_test.go")
	content := `//go:build L0
// +build L0

package service

func TestQuickValidation(t *testing.T) {
	// very fast test
}

func TestAnotherQuick(t *testing.T) {
	// another fast test
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Discover tests
	refs, err := DiscoverGoTestTags(tmpDir)
	require.NoError(t, err)

	// Verify
	assert.Len(t, refs, 2)
	for _, ref := range refs {
		assert.Equal(t, "gotest", ref.Type)
		assert.Contains(t, ref.Tags, "@L0")
		assert.Contains(t, ref.FilePath, "service_test.go")
	}

	// Check test names
	testNames := []string{refs[0].TestName, refs[1].TestName}
	assert.Contains(t, testNames, "TestQuickValidation")
	assert.Contains(t, testNames, "TestAnotherQuick")
}

func TestDiscoverGoTestTags_NoTagDefaultsL1(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test file WITHOUT build tag
	testFile := filepath.Join(tmpDir, "service_test.go")
	content := `package service

func TestValidateEmail(t *testing.T) {
	// test implementation
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Discover tests
	refs, err := DiscoverGoTestTags(tmpDir)
	require.NoError(t, err)

	// Verify - no tags yet, inference will add @L1
	assert.Len(t, refs, 1)
	assert.Equal(t, "gotest", refs[0].Type)
	assert.Equal(t, "TestValidateEmail", refs[0].TestName)
	// Tags will be empty here, inference adds @L1 later
	assert.Empty(t, refs[0].Tags)
}

func TestDiscoverGoTestTags_MultipleFiles(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create L0 file
	l0File := filepath.Join(tmpDir, "fast_test.go")
	l0Content := `//go:build L0

package service

func TestFast(t *testing.T) {}
`
	err := os.WriteFile(l0File, []byte(l0Content), 0644)
	require.NoError(t, err)

	// Create L1 file (no tag)
	l1File := filepath.Join(tmpDir, "slow_test.go")
	l1Content := `package service

func TestSlow(t *testing.T) {}
`
	err = os.WriteFile(l1File, []byte(l1Content), 0644)
	require.NoError(t, err)

	// Discover tests
	refs, err := DiscoverGoTestTags(tmpDir)
	require.NoError(t, err)

	// Verify
	assert.Len(t, refs, 2)

	// Find each test
	var fastTest, slowTest *TestReference
	for i := range refs {
		if refs[i].TestName == "TestFast" {
			fastTest = &refs[i]
		}
		if refs[i].TestName == "TestSlow" {
			slowTest = &refs[i]
		}
	}

	require.NotNil(t, fastTest, "TestFast should be found")
	require.NotNil(t, slowTest, "TestSlow should be found")

	assert.Contains(t, fastTest.Tags, "@L0")
	assert.Empty(t, slowTest.Tags) // Will get @L1 from inference
}
