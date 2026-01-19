// Package testingframework contains godog step implementations for specs/eac-core/testing-framework.
//
// This file contains step definitions for the testing framework meta-tests.
// These tests run meta-tests in isolation to verify the testing framework itself.
package testingframework

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// testContext holds state for testing framework scenarios.
type testContext struct {
	tempDir      string
	testOutput   string
	testExitCode int
}

var tfCtx *testContext

// RegisterSteps registers step definitions for testing framework feature specs.
func RegisterSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	sc.Step(`^the testing framework meta tests exist in testdata$`, theMetaTestsExistInTestdata)
	sc.Step(`^I run the (\w+) meta tests in isolation$`, iRunTheMetaTestsInIsolation)
	sc.Step(`^all (\w+) tests should pass$`, allSpecificTestsShouldPass)
	sc.Step(`^I run all testing framework meta tests in isolation$`, iRunAllMetaTestsInIsolation)
	sc.Step(`^the complete meta test suite should pass$`, allMetaTestsShouldPass)
	sc.Step(`^there should be no test failures$`, thereShouldBeNoTestFailures)
}

func theMetaTestsExistInTestdata() error {
	// Verify that meta test files exist in testdata
	// Path is relative to go/eac/core/testing/testdata from the original repo
	metaTestsDir := filepath.Join(getRepoRoot(), "go", "eac", "core", "testing", "testdata", "meta-tests")
	expectedFiles := []string{
		"discovery_test.go.txt",
		"inference_test.go.txt",
		"reports_test.go.txt",
		"suite_test.go.txt",
		"validation_test.go.txt",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(metaTestsDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("meta test file not found: %s", file)
		}
	}

	return nil
}

func iRunTheMetaTestsInIsolation(testName string) error {
	// Initialize context
	tfCtx = &testContext{}

	// Create temporary directory structure: tempRoot/testing/
	tempRoot, err := os.MkdirTemp("", "meta-tests-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	tfCtx.tempDir = tempRoot

	testingDir := filepath.Join(tempRoot, "testing")
	if err := os.MkdirAll(testingDir, 0o755); err != nil {
		return fmt.Errorf("failed to create testing dir: %w", err)
	}

	// Determine which test file to run
	var testFile string
	switch testName {
	case "discovery":
		testFile = "discovery_test.go.txt"
	case "inference":
		testFile = "inference_test.go.txt"
	case "reports":
		testFile = "reports_test.go.txt"
	case "suite":
		testFile = "suite_test.go.txt"
	case "validation":
		testFile = "validation_test.go.txt"
	default:
		return fmt.Errorf("unknown test type: %s", testName)
	}

	// Copy test file and rename to .go
	if err := copyMetaTest(testFile, testingDir); err != nil {
		return err
	}

	// Copy necessary supporting files (the testing package itself)
	if err := setupTestEnvironment(tempRoot, testingDir); err != nil {
		return err
	}

	// Run go test in the testing directory
	return runGoTest(testingDir)
}

func iRunAllMetaTestsInIsolation() error {
	// Initialize context
	tfCtx = &testContext{}

	// Create temporary directory structure: tempRoot/testing/
	tempRoot, err := os.MkdirTemp("", "meta-tests-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	tfCtx.tempDir = tempRoot

	testingDir := filepath.Join(tempRoot, "testing")
	if err := os.MkdirAll(testingDir, 0o755); err != nil {
		return fmt.Errorf("failed to create testing dir: %w", err)
	}

	// Copy all meta test files
	metaTests := []string{
		"discovery_test.go.txt",
		"inference_test.go.txt",
		"reports_test.go.txt",
		"suite_test.go.txt",
		"validation_test.go.txt",
	}

	for _, testFile := range metaTests {
		if err := copyMetaTest(testFile, testingDir); err != nil {
			return err
		}
	}

	// Copy necessary supporting files
	if err := setupTestEnvironment(tempRoot, testingDir); err != nil {
		return err
	}

	// Run go test in the testing directory
	return runGoTest(testingDir)
}

func getRepoRoot() string {
	// Navigate from go/eac/specs/impl/eac-core/testing-framework to repo root
	// This assumes tests run from the specs module directory
	cwd, _ := os.Getwd()
	// Try to find repo root by looking for .git or go.work
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Fallback to relative path from typical test location
			return filepath.Join(cwd, "..", "..", "..", "..")
		}
		dir = parent
	}
}

func copyMetaTest(testFile, destDir string) error {
	srcPath := filepath.Join(getRepoRoot(), "go", "eac", "core", "testing", "testdata", "meta-tests", testFile)
	destPath := filepath.Join(destDir, strings.TrimSuffix(testFile, ".txt"))

	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", testFile, err)
	}

	if err := os.WriteFile(destPath, content, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", destPath, err)
	}

	return nil
}

func setupTestEnvironment(tempRoot, testingDir string) error {
	repoRoot := getRepoRoot()
	coreDir := filepath.Join(repoRoot, "go", "eac", "core")
	testingPkgDir := filepath.Join(coreDir, "testing")

	// Copy all .go files except *_test.go files
	if err := copyGoFiles(testingPkgDir, testingDir); err != nil {
		return err
	}

	// Copy go.mod and go.sum from go/eac/core to tempRoot
	goModSrc := filepath.Join(coreDir, "go.mod")
	goModDest := filepath.Join(tempRoot, "go.mod")

	if err := copyFile(goModSrc, goModDest); err != nil {
		return fmt.Errorf("failed to copy go.mod: %w", err)
	}

	goSumSrc := filepath.Join(coreDir, "go.sum")
	goSumDest := filepath.Join(tempRoot, "go.sum")

	if err := copyFile(goSumSrc, goSumDest); err != nil {
		return fmt.Errorf("failed to copy go.sum: %w", err)
	}

	// Also need to copy other go/eac/core subdirectories that testing depends on
	if err := copyDependentPackages(coreDir, tempRoot); err != nil {
		return err
	}

	// Copy .r2r/eac config files (required by config.Global() used in testing package)
	if err := copyConfigFiles(repoRoot, tempRoot); err != nil {
		return err
	}

	return nil
}

func copyGoFiles(srcDir, destDir string) error {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read dir %s: %w", srcDir, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only copy .go files that are not test files
		if strings.HasSuffix(file.Name(), ".go") && !strings.HasSuffix(file.Name(), "_test.go") {
			srcPath := filepath.Join(srcDir, file.Name())
			destPath := filepath.Join(destDir, file.Name())

			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dest string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", src, err)
	}

	if err := os.WriteFile(dest, content, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", dest, err)
	}

	return nil
}

func copyDependentPackages(coreDir, tempRoot string) error {
	// Copy contracts, environments, config, etc. packages
	// Note: defaults and logging are required by config/modules.go and contracts/retry.go
	packages := []string{"config", "contracts", "defaults", "environments", "git", "logging", "markdown", "repository", "system-deps"}

	for _, pkg := range packages {
		srcPkgDir := filepath.Join(coreDir, pkg)
		if _, err := os.Stat(srcPkgDir); os.IsNotExist(err) {
			continue // Skip if package doesn't exist
		}

		destPkgDir := filepath.Join(tempRoot, pkg)
		if err := os.MkdirAll(destPkgDir, 0o755); err != nil {
			return fmt.Errorf("failed to create %s: %w", destPkgDir, err)
		}

		if err := copyPackageRecursive(srcPkgDir, destPkgDir); err != nil {
			return fmt.Errorf("failed to copy package %s: %w", pkg, err)
		}
	}

	return nil
}

func copyPackageRecursive(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip testdata, tests directories, and test files
		if info.IsDir() && (info.Name() == "testdata" || info.Name() == "tests") {
			return filepath.SkipDir
		}

		if strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		// Only copy .go, .yml/.yaml, and .json files (json needed for embedded schemas)
		if strings.HasSuffix(info.Name(), ".go") ||
			strings.HasSuffix(info.Name(), ".yml") ||
			strings.HasSuffix(info.Name(), ".yaml") ||
			strings.HasSuffix(info.Name(), ".json") {
			return copyFile(path, destPath)
		}

		return nil
	})
}

func copyConfigFiles(repoRoot, tempRoot string) error {
	// Copy .r2r/eac config directory which contains repository.yml and other configs
	// needed by config.Global()
	configSrcDir := filepath.Join(repoRoot, ".r2r", "eac")
	configDestDir := filepath.Join(tempRoot, ".r2r", "eac")

	if err := os.MkdirAll(configDestDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	// Create a fake .git directory so the config loader can find the repo root
	gitDir := filepath.Join(tempRoot, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .git dir: %w", err)
	}

	// Copy essential config files
	// Note: Modules are defined in repository.yml (unified config)
	configFiles := []string{
		"repository.yml",
		"module-types.yml",
		"testing-tags.yml",
		"test-suites.yml",
	}

	for _, file := range configFiles {
		srcPath := filepath.Join(configSrcDir, file)
		destPath := filepath.Join(configDestDir, file)

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue // Skip if file doesn't exist
		}

		if err := copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy config file %s: %w", file, err)
		}
	}

	return nil
}

func runGoTest(dir string) error {
	cmd := exec.Command("go", "test", "-v", "-tags=L1,ov")
	cmd.Dir = dir
	// Set R2R_REPO_ROOT to tempRoot (parent of testing dir) so config.Global() finds configs
	tempRoot := filepath.Dir(dir)
	cmd.Env = append(os.Environ(), "R2R_REPO_ROOT="+tempRoot)
	output, err := cmd.CombinedOutput()
	tfCtx.testOutput = string(output)

	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			tfCtx.testExitCode = exitErr.ExitCode()
		}
		return fmt.Errorf("go test failed: %w\nOutput:\n%s", err, tfCtx.testOutput)
	}

	tfCtx.testExitCode = 0
	return nil
}

func allSpecificTestsShouldPass(testName string) error {
	if tfCtx.testExitCode != 0 {
		return fmt.Errorf("%s tests failed with exit code %d:\n%s", testName, tfCtx.testExitCode, tfCtx.testOutput)
	}

	// Verify output contains PASS
	if !strings.Contains(tfCtx.testOutput, "PASS") {
		return fmt.Errorf("%s tests did not pass:\n%s", testName, tfCtx.testOutput)
	}

	return nil
}

func allMetaTestsShouldPass() error {
	if tfCtx.testExitCode != 0 {
		return fmt.Errorf("meta tests failed with exit code %d:\n%s", tfCtx.testExitCode, tfCtx.testOutput)
	}

	// Verify output contains PASS
	if !strings.Contains(tfCtx.testOutput, "PASS") {
		return fmt.Errorf("meta tests did not pass:\n%s", tfCtx.testOutput)
	}

	return nil
}

func thereShouldBeNoTestFailures() error {
	if strings.Contains(tfCtx.testOutput, "FAIL") {
		return fmt.Errorf("test output contains failures:\n%s", tfCtx.testOutput)
	}
	return nil
}

func cleanupTestContext() {
	if tfCtx != nil && tfCtx.tempDir != "" {
		os.RemoveAll(tfCtx.tempDir)
	}
	tfCtx = nil
}
