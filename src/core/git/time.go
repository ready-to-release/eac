package git

import "time"

// timeNow is a package-level function that returns the current time.
// It can be overridden in tests to provide deterministic timestamps.
var timeNow = time.Now
