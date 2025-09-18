package util

import (
	"strings"
)

// SanitizeFilename removes invalid characters from a filename
func SanitizeFilename(filename string) string {
	if filename == "" {
		return "video"
	}

	// Replace common invalid characters with underscores
	replacer := strings.NewReplacer(
		"\\", "_",
		"/", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "'",
		"<", "_",
		">", "_",
		"|", "_",
		" ", " ", // Keep spaces
	)

	// Apply replacements and trim extra spaces
	result := replacer.Replace(filename)
	result = strings.TrimSpace(result)

	// Ensure the filename isn't empty after sanitization
	if result == "" {
		return "video"
	}

	return result
}
