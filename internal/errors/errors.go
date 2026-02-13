package errors

import (
	"fmt"
	"strings"
)

// Error represents an error with actionable suggestions
type Error struct {
	Message     string   // The error message
	Suggestions []string // Actionable suggestions to fix the issue
	Context     []string // Additional context (optional)
}

// Error implements the error interface
func (e *Error) Error() string {
	var b strings.Builder

	// Main error message (red)
	b.WriteString(Red("Error: "))
	b.WriteString(e.Message)
	b.WriteString("\n")

	// Context (yellow, if any)
	if len(e.Context) > 0 {
		b.WriteString("\n")
		for _, ctx := range e.Context {
			b.WriteString(Yellow("  " + ctx))
			b.WriteString("\n")
		}
	}

	// Suggestions (green arrows)
	if len(e.Suggestions) > 0 {
		b.WriteString("\n")
		b.WriteString(Bold("To resolve:"))
		b.WriteString("\n")
		for _, suggestion := range e.Suggestions {
			b.WriteString(Green("  → "))
			b.WriteString(suggestion)
			b.WriteString("\n")
		}
	}

	return b.String()
}

// New creates a new error with suggestions
func New(message string, suggestions ...string) *Error {
	return &Error{
		Message:     message,
		Suggestions: suggestions,
	}
}

// WithContext adds context to an error
func (e *Error) WithContext(context ...string) *Error {
	e.Context = append(e.Context, context...)
	return e
}

// Wrap wraps an existing error with suggestions
func Wrap(err error, message string, suggestions ...string) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		Message:     fmt.Sprintf("%s: %v", message, err),
		Suggestions: suggestions,
	}
}
