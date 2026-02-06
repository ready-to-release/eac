package console

import "strings"

// extractToolFromMoniker extracts the tool name from a moniker.
// Moniker format: context:module:component:tool[:extra...]
func extractToolFromMoniker(moniker string) string {
	parts := strings.Split(moniker, ":")
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}
