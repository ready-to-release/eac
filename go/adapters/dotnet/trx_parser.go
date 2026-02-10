package dotnet

import (
	"encoding/xml"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/core/ctrf"
)

// ConvertTRXToCTRF parses TRX XML data and converts it to a CTRF report.
func ConvertTRXToCTRF(trxData []byte) *ctrf.Report {
	var trx TRXTestRun
	if err := xml.Unmarshal(trxData, &trx); err != nil {
		return nil
	}

	report := ctrf.NewReport("dotnet-xunit")

	for _, result := range trx.Results.UnitTestResults {
		test := ctrf.Test{
			Name:     result.TestName,
			Status:   mapTRXOutcomeToCTRF(result.Outcome),
			Duration: parseTRXDuration(result.Duration),
		}

		if result.Output.ErrorInfo.Message != "" {
			test.Message = result.Output.ErrorInfo.Message
			test.Trace = result.Output.ErrorInfo.StackTrace
		}

		report.AddTest(test)
	}

	if trx.Times.Start != "" && trx.Times.Finish != "" {
		if startTime, err := time.Parse(time.RFC3339, trx.Times.Start); err == nil {
			if endTime, err := time.Parse(time.RFC3339, trx.Times.Finish); err == nil {
				report.SetTimes(startTime.UnixMilli(), endTime.UnixMilli())
			}
		}
	} else {
		report.Finalize()
	}

	return report
}

// mapTRXOutcomeToCTRF maps TRX outcome strings to CTRF status.
func mapTRXOutcomeToCTRF(outcome string) ctrf.Status {
	switch strings.ToLower(outcome) {
	case "passed":
		return ctrf.StatusPassed
	case "failed":
		return ctrf.StatusFailed
	case "notexecuted":
		return ctrf.StatusSkipped
	case "inconclusive":
		return ctrf.StatusPending
	default:
		return ctrf.StatusOther
	}
}

// parseTRXDuration converts TRX duration format (HH:MM:SS.FFFFFFF) to milliseconds.
func parseTRXDuration(durationStr string) int64 {
	if durationStr == "" {
		return 0
	}
	d, err := time.ParseDuration(reformatTRXDuration(durationStr))
	if err != nil {
		return 0
	}
	return d.Milliseconds()
}

// reformatTRXDuration converts "HH:MM:SS.FFFFFFF" to Go duration string "HhMmS.FFFFFFs".
func reformatTRXDuration(trxDuration string) string {
	parts := strings.Split(trxDuration, ":")
	if len(parts) != 3 {
		return "0s"
	}
	return parts[0] + "h" + parts[1] + "m" + parts[2] + "s"
}
