package reporter

import "time"

// TestSuiteReportData contains all data needed for test suite report generation.
type TestSuiteReportData struct {
	// Suite information
	SuiteName    string
	SuiteMoniker string
	StartTime    string
	EndTime      string
	Duration     float64
	Status       string
	GeneratedAt  string

	// Test counts
	TestsDiscovered        int
	ProductionTests        int
	FrameworkTestsExcluded int
	TestsTotal             int
	TestsPassed            int
	TestsFailed            int
	TestsSkipped           int
	TestPassRate           float64

	// Package counts
	PackagesTotal   int
	PackagesPassed  int
	PackagesFailed  int
	PackagePassRate float64

	// Module breakdown
	Modules []ModuleReportData

	// File paths
	ResultsDirectory string
}

// ModuleReportData contains test results for a single module.
type ModuleReportData struct {
	ModuleName     string
	TestResultDir  string
	Features       []FeatureReportData
	UnitTests      []UnitTestReportData
	TotalScenarios int
	TotalUnitTests int
}

// FeatureReportData contains information about a BDD feature.
type FeatureReportData struct {
	Name          string
	URI           string
	NormalizedURI string
	Tags          string
	TestCount     int
	PassedCount   int
	Status        string
}

// UnitTestReportData contains information about Go unit test packages.
type UnitTestReportData struct {
	Package   string
	TestCount int
	Status    string
}

// BuildReportData creates a TestSuiteReportData from raw test results.
func BuildReportData(
	suiteName, suiteMoniker string,
	startTime, endTime time.Time,
	testsDiscovered, productionTests, frameworkTestsExcluded int,
	testsTotal, testsPassed, testsFailed, testsSkipped int,
	packagesTotal, packagesPassed, packagesFailed int,
	modules []ModuleReportData,
	resultsDirectory string,
) TestSuiteReportData {
	duration := endTime.Sub(startTime)

	// Calculate rates
	testPassRate := 0.0
	if testsTotal > 0 {
		testPassRate = float64(testsPassed) / float64(testsTotal) * 100
	}

	packagePassRate := 0.0
	if packagesTotal > 0 {
		packagePassRate = float64(packagesPassed) / float64(packagesTotal) * 100
	}

	// Determine final status
	finalStatus := "✅ PASSED"
	if packagesFailed > 0 {
		finalStatus = "❌ FAILED"
	}

	return TestSuiteReportData{
		SuiteName:              suiteName,
		SuiteMoniker:           suiteMoniker,
		StartTime:              startTime.Format("2006-01-02 15:04:05"),
		EndTime:                endTime.Format("2006-01-02 15:04:05"),
		Duration:               duration.Seconds(),
		Status:                 finalStatus,
		GeneratedAt:            endTime.Format("2006-01-02 15:04:05"),
		TestsDiscovered:        testsDiscovered,
		ProductionTests:        productionTests,
		FrameworkTestsExcluded: frameworkTestsExcluded,
		TestsTotal:             testsTotal,
		TestsPassed:            testsPassed,
		TestsFailed:            testsFailed,
		TestsSkipped:           testsSkipped,
		TestPassRate:           testPassRate,
		PackagesTotal:          packagesTotal,
		PackagesPassed:         packagesPassed,
		PackagesFailed:         packagesFailed,
		PackagePassRate:        packagePassRate,
		Modules:                modules,
		ResultsDirectory:       resultsDirectory,
	}
}
