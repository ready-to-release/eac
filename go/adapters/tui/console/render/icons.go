package render

// UoWStatusIcon returns the appropriate icon for a UoW status.
// Status values: 0=Pending, 1=Running, 2=Complete, 3=Skipped, 4=Failed.
func UoWStatusIcon(status int, asciiMode bool) string {
	if asciiMode {
		switch status {
		case 0: // Pending
			return "o"
		case 1: // Running
			return ">"
		case 2: // Complete
			return "V"
		case 3: // Skipped
			return "="
		case 4: // Failed
			return "X"
		default:
			return "?"
		}
	}
	switch status {
	case 0:
		return "◦"
	case 1:
		return "▶"
	case 2:
		return "✓"
	case 3:
		return "⏭"
	case 4:
		return "✗"
	default:
		return "?"
	}
}

// PhaseIcon returns the appropriate icon for a phase status.
// Status values: 0=Pending, 1=Active, 2=Complete, 3=Failed.
func PhaseIcon(status int, asciiMode bool) string {
	if asciiMode {
		switch status {
		case 0: // Pending
			return "o"
		case 1: // Active
			return ">"
		case 2: // Complete
			return "V"
		case 3: // Failed
			return "X"
		default:
			return "?"
		}
	}
	switch status {
	case 0:
		return "○"
	case 1:
		return "▶"
	case 2:
		return "✓"
	case 3:
		return "✗"
	default:
		return "?"
	}
}

// LineIcons returns fail, warn, info icons for log line prefixes.
func LineIcons(asciiMode bool) (fail, warn, info string) {
	if asciiMode {
		return "X", "!", " "
	}
	return "✗", "!", " "
}
