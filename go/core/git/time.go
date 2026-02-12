package git

import "time"

// Clock is a function that returns the current time.
// It is used by Repository to timestamp commits, enabling deterministic
// tests via constructor injection rather than mutable package-level state.
type Clock func() time.Time
