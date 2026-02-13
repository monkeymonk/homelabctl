package errors

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	err := New(
		"test error message",
		"suggestion 1",
		"suggestion 2",
	).WithContext(
		"context line 1",
		"context line 2",
	)

	output := err.Error()

	// Test that output contains all parts
	if !strings.Contains(output, "test error message") {
		t.Error("Error should contain message")
	}
	if !strings.Contains(output, "suggestion 1") {
		t.Error("Error should contain first suggestion")
	}
	if !strings.Contains(output, "suggestion 2") {
		t.Error("Error should contain second suggestion")
	}
	if !strings.Contains(output, "context line 1") {
		t.Error("Error should contain first context line")
	}
	if !strings.Contains(output, "context line 2") {
		t.Error("Error should contain second context line")
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		suggestions []string
		wantNil     bool
	}{
		{
			name:        "with suggestions",
			message:     "error occurred",
			suggestions: []string{"try this", "or that"},
			wantNil:     false,
		},
		{
			name:        "without suggestions",
			message:     "error occurred",
			suggestions: []string{},
			wantNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.message, tt.suggestions...)
			if (err == nil) != tt.wantNil {
				t.Errorf("New() = %v, wantNil %v", err, tt.wantNil)
			}
			if err != nil && !strings.Contains(err.Error(), tt.message) {
				t.Errorf("Error should contain message %q", tt.message)
			}
		})
	}
}

func TestWithContext(t *testing.T) {
	err := New("error message", "suggestion")
	errWithCtx := err.WithContext("context1", "context2")

	output := errWithCtx.Error()

	if !strings.Contains(output, "context1") {
		t.Error("Should contain context1")
	}
	if !strings.Contains(output, "context2") {
		t.Error("Should contain context2")
	}
}

func TestWrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")

	wrappedErr := Wrap(
		originalErr,
		"wrapped message",
		"suggestion 1",
	)

	output := wrappedErr.Error()

	if !strings.Contains(output, "wrapped message") {
		t.Error("Should contain wrapped message")
	}
	if !strings.Contains(output, "suggestion 1") {
		t.Error("Should contain suggestion")
	}
	if !strings.Contains(output, "original error") {
		t.Error("Should contain original error")
	}
}

func TestColorsEnabled(t *testing.T) {
	// Save original value
	originalNoColor := os.Getenv("NO_COLOR")
	defer func() {
		if originalNoColor != "" {
			os.Setenv("NO_COLOR", originalNoColor)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	tests := []struct {
		name    string
		noColor string
	}{
		{
			name:    "NO_COLOR=1 disables colors",
			noColor: "1",
		},
		{
			name:    "NO_COLOR=true disables colors",
			noColor: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NO_COLOR", tt.noColor)

			got := colorsEnabled()
			if got {
				t.Errorf("colorsEnabled() = true when NO_COLOR=%s, want false", tt.noColor)
			}
		})
	}

	// Note: We can't reliably test the terminal detection in a test environment
	// because tests run without a TTY. The actual behavior when NO_COLOR is unset
	// depends on whether stdout is a terminal, which varies by test environment.
}

func TestErrorFormatting(t *testing.T) {
	// Test that error formatting doesn't crash with various inputs
	tests := []struct {
		name        string
		message     string
		suggestions []string
		context     []string
	}{
		{
			name:        "empty message",
			message:     "",
			suggestions: []string{"fix it"},
			context:     []string{},
		},
		{
			name:        "no suggestions",
			message:     "error",
			suggestions: []string{},
			context:     []string{},
		},
		{
			name:        "no context",
			message:     "error",
			suggestions: []string{"fix it"},
			context:     []string{},
		},
		{
			name:        "many suggestions",
			message:     "error",
			suggestions: []string{"s1", "s2", "s3", "s4", "s5"},
			context:     []string{},
		},
		{
			name:        "many context lines",
			message:     "error",
			suggestions: []string{"fix"},
			context:     []string{"c1", "c2", "c3", "c4", "c5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.message, tt.suggestions...)
			if len(tt.context) > 0 {
				err = err.WithContext(tt.context...)
			}

			// Should not panic
			output := err.Error()

			// Should produce non-empty output
			if output == "" {
				t.Error("Error() should return non-empty string")
			}
		})
	}
}

func TestErrorInterface(t *testing.T) {
	// Test that our Error type satisfies the error interface
	var _ error = &Error{}
	var _ error = New("test")
}

func TestIs(t *testing.T) {
	// Test that our enhanced error can be checked with errors.Is
	enhancedErr := New("test error")

	// Should be itself
	if !errors.Is(enhancedErr, enhancedErr) {
		t.Error("errors.Is should return true for same error")
	}
}
