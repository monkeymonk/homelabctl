package render

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/monkeymonk/homelabctl/internal/errors"
	"github.com/monkeymonk/homelabctl/internal/paths"
)

// Context represents the template context passed to gomplate
type Context struct {
	Vars   map[string]interface{} `yaml:"vars"`
	Stack  map[string]interface{} `yaml:"stack"`
	Stacks map[string]interface{} `yaml:"stacks"`
}

// RenderTemplate renders a template file using gomplate
func RenderTemplate(templatePath string, context *Context) (string, error) {
	// Check gomplate is available
	if _, err := exec.LookPath("gomplate"); err != nil {
		return "", errors.New(
			"gomplate not found in PATH",
			"Install gomplate: https://docs.gomplate.ca/installing/",
			"On Linux: curl -o /usr/local/bin/gomplate -sSL https://github.com/hairyhenderson/gomplate/releases/download/v3.11.5/gomplate_linux-amd64",
			"On macOS: brew install gomplate",
		)
	}

	// Marshal context to YAML
	contextData, err := yaml.Marshal(context)
	if err != nil {
		return "", fmt.Errorf("failed to marshal context: %w", err)
	}

	// Create temp file for context
	tmpfile, err := os.CreateTemp("", "homelabctl-context-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	// Set secure permissions (0600) to prevent other users from reading context data
	if err := tmpfile.Chmod(paths.SecureFilePermissions); err != nil {
		tmpfile.Close()
		return "", fmt.Errorf("failed to set temp file permissions: %w", err)
	}

	if _, err := tmpfile.Write(contextData); err != nil {
		tmpfile.Close()
		return "", fmt.Errorf("failed to write context: %w", err)
	}
	tmpfile.Close()

	// Run gomplate
	cmd := exec.Command("gomplate",
		"-f", templatePath,
		"-c", ".="+tmpfile.Name(),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Parse gomplate error for better messaging
		stderrStr := stderr.String()

		suggestions := []string{
			fmt.Sprintf("Check template syntax in: %s", templatePath),
			fmt.Sprintf("View context: cat %s", tmpfile.Name()),
			"Run: gomplate -f <template> -c .=<context> to debug",
		}

		return "", errors.New(
			fmt.Sprintf("gomplate failed to render %s", templatePath),
			suggestions...,
		).WithContext(
			"Gomplate error:",
			stderrStr,
		)
	}

	return stdout.String(), nil
}

// RenderToFile renders a template and writes to output file
func RenderToFile(templatePath, outputPath string, context *Context) error {
	content, err := RenderTemplate(templatePath, context)
	if err != nil {
		return err
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), paths.DirPermissions); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), paths.FilePermissions); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
