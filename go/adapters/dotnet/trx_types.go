package dotnet

import "encoding/xml"

// TRXTestRun is the root element of a TRX file.
type TRXTestRun struct {
	XMLName  xml.Name   `xml:"TestRun"`
	Results  TRXResults `xml:"Results"`
	Counters TRXCounters
	Times    TRXTimes
}

// TRXResults contains the list of unit test results.
type TRXResults struct {
	UnitTestResults []TRXUnitTestResult `xml:"UnitTestResult"`
}

// TRXUnitTestResult represents a single test result in TRX format.
type TRXUnitTestResult struct {
	TestName  string `xml:"testName,attr"`
	Outcome   string `xml:"outcome,attr"`
	Duration  string `xml:"duration,attr"`
	StartTime string `xml:"startTime,attr"`
	EndTime   string `xml:"endTime,attr"`
	Output    struct {
		StdOut    string `xml:"StdOut"`
		StdErr    string `xml:"StdErr"`
		ErrorInfo struct {
			Message    string `xml:"Message"`
			StackTrace string `xml:"StackTrace"`
		} `xml:"ErrorInfo"`
	} `xml:"Output"`
}

// TRXCounters provides aggregate test counts from TRX.
type TRXCounters struct {
	Total        int `xml:"total,attr"`
	Executed     int `xml:"executed,attr"`
	Passed       int `xml:"passed,attr"`
	Failed       int `xml:"failed,attr"`
	Error        int `xml:"error,attr"`
	Timeout      int `xml:"timeout,attr"`
	NotRunnable  int `xml:"notRunnable,attr"`
	NotExecuted  int `xml:"notExecuted,attr"`
	Inconclusive int `xml:"inconclusive,attr"`
}

// TRXTimes contains timing information from TRX.
type TRXTimes struct {
	Creation string `xml:"creation,attr"`
	Queuing  string `xml:"queuing,attr"`
	Start    string `xml:"start,attr"`
	Finish   string `xml:"finish,attr"`
}
