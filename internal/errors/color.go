package errors

import (
	"os"
)

var (
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
)

// colorsEnabled checks if color output is supported
func colorsEnabled() bool {
	// Disable colors if NO_COLOR is set
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if output is a terminal
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}

	return false
}

// colorize applies color if supported
func colorize(color, text string) string {
	if colorsEnabled() {
		return color + text + colorReset
	}
	return text
}

// Red returns red colored text
func Red(text string) string {
	return colorize(colorRed, text)
}

// Yellow returns yellow colored text
func Yellow(text string) string {
	return colorize(colorYellow, text)
}

// Green returns green colored text
func Green(text string) string {
	return colorize(colorGreen, text)
}

// Bold returns bold text
func Bold(text string) string {
	return colorize(colorBold, text)
}
