package extensions

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}

func TruncateStringStart(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:0+3] + "..."
}
