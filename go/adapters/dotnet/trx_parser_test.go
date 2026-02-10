package dotnet

import (
	"testing"
)

func TestConvertTRXToCTRF_PassingTests(t *testing.T) {
	trxData := []byte(`<?xml version="1.0" encoding="utf-8"?>
<TestRun>
  <Results>
    <UnitTestResult testName="MyTest.Add_ReturnsSum" outcome="Passed" duration="00:00:00.0012345" />
    <UnitTestResult testName="MyTest.Divide_ByZero" outcome="Passed" duration="00:00:00.0005678" />
  </Results>
  <ResultSummary>
    <Counters total="2" executed="2" passed="2" failed="0" />
  </ResultSummary>
  <Times start="2024-01-01T00:00:00Z" finish="2024-01-01T00:00:01Z" />
</TestRun>`)

	report := ConvertTRXToCTRF(trxData)
	if report == nil {
		t.Fatal("expected report, got nil")
	}
	if report.Results.Summary.Tests != 2 {
		t.Errorf("Tests = %d, want 2", report.Results.Summary.Tests)
	}
	if report.Results.Summary.Passed != 2 {
		t.Errorf("Passed = %d, want 2", report.Results.Summary.Passed)
	}
	if report.Results.Summary.Failed != 0 {
		t.Errorf("Failed = %d, want 0", report.Results.Summary.Failed)
	}
}

func TestConvertTRXToCTRF_FailingTests(t *testing.T) {
	trxData := []byte(`<?xml version="1.0" encoding="utf-8"?>
<TestRun>
  <Results>
    <UnitTestResult testName="MyTest.Broken" outcome="Failed" duration="00:00:00.0500000">
      <Output>
        <ErrorInfo>
          <Message>Assert.Equal failed</Message>
          <StackTrace>at MyTest.Broken() in test.cs:line 42</StackTrace>
        </ErrorInfo>
      </Output>
    </UnitTestResult>
  </Results>
  <ResultSummary>
    <Counters total="1" executed="1" passed="0" failed="1" />
  </ResultSummary>
</TestRun>`)

	report := ConvertTRXToCTRF(trxData)
	if report == nil {
		t.Fatal("expected report, got nil")
	}
	if report.Results.Summary.Failed != 1 {
		t.Errorf("Failed = %d, want 1", report.Results.Summary.Failed)
	}
	if report.Results.Tests[0].Message != "Assert.Equal failed" {
		t.Errorf("Message = %q, want 'Assert.Equal failed'", report.Results.Tests[0].Message)
	}
	if report.Results.Tests[0].Trace != "at MyTest.Broken() in test.cs:line 42" {
		t.Errorf("Trace = %q, want stack trace", report.Results.Tests[0].Trace)
	}
}

func TestConvertTRXToCTRF_SkippedTests(t *testing.T) {
	trxData := []byte(`<?xml version="1.0" encoding="utf-8"?>
<TestRun>
  <Results>
    <UnitTestResult testName="MyTest.Skipped" outcome="NotExecuted" duration="00:00:00.0000000" />
  </Results>
</TestRun>`)

	report := ConvertTRXToCTRF(trxData)
	if report == nil {
		t.Fatal("expected report, got nil")
	}
	if report.Results.Summary.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", report.Results.Summary.Skipped)
	}
}

func TestConvertTRXToCTRF_MixedResults(t *testing.T) {
	trxData := []byte(`<?xml version="1.0" encoding="utf-8"?>
<TestRun>
  <Results>
    <UnitTestResult testName="MyTest.Pass" outcome="Passed" duration="00:00:00.0010000" />
    <UnitTestResult testName="MyTest.Fail" outcome="Failed" duration="00:00:00.0020000" />
    <UnitTestResult testName="MyTest.Skip" outcome="NotExecuted" duration="00:00:00.0000000" />
    <UnitTestResult testName="MyTest.Pending" outcome="Inconclusive" duration="00:00:00.0000000" />
  </Results>
</TestRun>`)

	report := ConvertTRXToCTRF(trxData)
	if report == nil {
		t.Fatal("expected report, got nil")
	}
	if report.Results.Summary.Tests != 4 {
		t.Errorf("Tests = %d, want 4", report.Results.Summary.Tests)
	}
	if report.Results.Summary.Passed != 1 {
		t.Errorf("Passed = %d, want 1", report.Results.Summary.Passed)
	}
	if report.Results.Summary.Failed != 1 {
		t.Errorf("Failed = %d, want 1", report.Results.Summary.Failed)
	}
	if report.Results.Summary.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", report.Results.Summary.Skipped)
	}
	if report.Results.Summary.Pending != 1 {
		t.Errorf("Pending = %d, want 1", report.Results.Summary.Pending)
	}
}

func TestConvertTRXToCTRF_InvalidXML(t *testing.T) {
	report := ConvertTRXToCTRF([]byte("not xml"))
	if report != nil {
		t.Error("expected nil for invalid XML")
	}
}

func TestConvertTRXToCTRF_EmptyResults(t *testing.T) {
	trxData := []byte(`<?xml version="1.0" encoding="utf-8"?>
<TestRun>
  <Results />
</TestRun>`)

	report := ConvertTRXToCTRF(trxData)
	if report == nil {
		t.Fatal("expected report, got nil")
	}
	if report.Results.Summary.Tests != 0 {
		t.Errorf("Tests = %d, want 0", report.Results.Summary.Tests)
	}
}

func TestParseTRXDuration(t *testing.T) {
	tests := []struct {
		input  string
		wantMs int64
	}{
		{"00:00:00.0012345", 1},
		{"00:00:01.5000000", 1500},
		{"00:01:30.0000000", 90000},
		{"", 0},
	}
	for _, tt := range tests {
		got := parseTRXDuration(tt.input)
		if got != tt.wantMs {
			t.Errorf("parseTRXDuration(%q) = %d, want %d", tt.input, got, tt.wantMs)
		}
	}
}

func TestReformatTRXDuration(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"00:00:01.5000000", "00h00m01.5000000s"},
		{"01:30:00.0000000", "01h30m00.0000000s"},
		{"invalid", "0s"},
	}
	for _, tt := range tests {
		got := reformatTRXDuration(tt.input)
		if got != tt.want {
			t.Errorf("reformatTRXDuration(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMapTRXOutcomeToCTRF(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Passed", "passed"},
		{"Failed", "failed"},
		{"NotExecuted", "skipped"},
		{"Inconclusive", "pending"},
		{"Unknown", "other"},
	}
	for _, tt := range tests {
		got := mapTRXOutcomeToCTRF(tt.input)
		if string(got) != tt.want {
			t.Errorf("mapTRXOutcomeToCTRF(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
