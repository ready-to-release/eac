package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// testContext holds state for testing framework scenarios
type testContext struct {
	tempDir      string
	testOutput   string
	testExitCode int
}

var ctx *testContext

func InitializeTestingFrameworkSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the testing framework meta tests exist in testdata$`, theMetaTestsExistInTestdata)
	sc.Step(`^I run the (\w+) meta tests in isolation$`, iRunTheMetaTestsInIsolation)
	sc.Step(`^all (\w+) tests should pass$`, allSpecificTestsShouldPass)
	sc.Step(`^I run all testing framework meta tests in isolation$`, iRunAllMetaTestsInIsolation)
	sc.Step(`^the complete meta test suite should pass$`, allMetaTestsShouldPass)
	sc.Step(`^there should be no test failures$`, thereShouldBeNoTestFailures)
}

func theMetaTestsExistInTestdata() error {
	// Verify that meta test files exist in testdata
	metaTestsDir := "../testdata/meta-tests"
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
	ctx = &testContext{}

	// Create temporary directory structure: tempRoot/testing/
	tempRoot, err := os.MkdirTemp("", "meta-tests-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	ctx.tempDir = tempRoot

	testingDir := filepath.Join(tempRoot, "testing")
	if err := os.MkdirAll(testingDir, 0755); err != nil {
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
	ctx = &testContext{}

	// Create temporary directory structure: tempRoot/testing/
	tempRoot, err := os.MkdirTemp("", "meta-tests-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	ctx.tempDir = tempRoot

	testingDir := filepath.Join(tempRoot, "testing")
	if err := os.MkdirAll(testingDir, 0755); err != nil {
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

func copyMetaTest(testFile string, destDir string) error {
	srcPath := filepath.Join("../testdata/meta-tests", testFile)
	destPath := filepath.Join(destDir, strings.TrimSuffix(testFile, ".txt"))

	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", testFile, err)
	}

	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", destPath, err)
	}

	return nil
}

func setupTestEnvironment(tempRoot, testingDir string) error {
	// Copy the testing package source files (non-test .go files)
	testingPkgDir := ".."

	// Copy all .go files except *_test.go files
	if err := copyGoFiles(testingPkgDir, testingDir); err != nil {
		return err
	}

	// Copy go.mod and go.sum from src/core to tempRoot
	coreDir := "../.."
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

	// Also need to copy other src/core subdirectories that testing depends on
	// (contracts, environments, etc.)
	if err := copyDependentPackages(coreDir, tempRoot); err != nil {
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

	if err := os.WriteFile(dest, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", dest, err)
	}

	return nil
}

func copyDependentPackages(coreDir, tempRoot string) error {
	// Copy contracts, environments, etc. packages
	packages := []string{"contracts", "environments", "git", "markdown", "repository", "system-deps"}

	for _, pkg := range packages {
		srcPkgDir := filepath.Join(coreDir, pkg)
		if _, err := os.Stat(srcPkgDir); os.IsNotExist(err) {
			continue // Skip if package doesn't exist
		}

		destPkgDir := filepath.Join(tempRoot, pkg)
		if err := os.MkdirAll(destPkgDir, 0755); err != nil {
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
			return os.MkdirAll(destPath, 0755)
		}

		// Only copy .go and .yml/.yaml files
		if strings.HasSuffix(info.Name(), ".go") ||
		   strings.HasSuffix(info.Name(), ".yml") ||
		   strings.HasSuffix(info.Name(), ".yaml") {
			return copyFile(path, destPath)
		}

		return nil
	})
}

func runGoTest(dir string) error {
	cmd := exec.Command("go", "test", "-v", "-tags=L1,ov")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	ctx.testOutput = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			ctx.testExitCode = exitErr.ExitCode()
		}
		return fmt.Errorf("go test failed: %w\nOutput:\n%s", err, ctx.testOutput)
	}

	ctx.testExitCode = 0
	return nil
}

func allSpecificTestsShouldPass(testName string) error {
	if ctx.testExitCode != 0 {
		return fmt.Errorf("%s tests failed with exit code %d:\n%s", testName, ctx.testExitCode, ctx.testOutput)
	}

	// Verify output contains PASS
	if !strings.Contains(ctx.testOutput, "PASS") {
		return fmt.Errorf("%s tests did not pass:\n%s", testName, ctx.testOutput)
	}

	return nil
}

func allMetaTestsShouldPass() error {
	if ctx.testExitCode != 0 {
		return fmt.Errorf("meta tests failed with exit code %d:\n%s", ctx.testExitCode, ctx.testOutput)
	}

	// Verify output contains PASS
	if !strings.Contains(ctx.testOutput, "PASS") {
		return fmt.Errorf("meta tests did not pass:\n%s", ctx.testOutput)
	}

	return nil
}

func thereShouldBeNoTestFailures() error {
	if strings.Contains(ctx.testOutput, "FAIL") {
		return fmt.Errorf("test output contains failures:\n%s", ctx.testOutput)
	}
	return nil
}

func cleanupTestContext() {
	if ctx != nil && ctx.tempDir != "" {
		os.RemoveAll(ctx.tempDir)
	}
	ctx = nil
}
