package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates a temporary directory for testing and returns it along with a cleanup function
func TempDir(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "homelabctl-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

// WriteFile writes content to a file in tests, creating parent directories if needed
func WriteFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("Failed to create directory for %s: %v", path, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

// CreateSymlink creates a symlink for tests
func CreateSymlink(t *testing.T, target, link string) {
	t.Helper()

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(link), 0755); err != nil {
		t.Fatalf("Failed to create directory for symlink %s: %v", link, err)
	}

	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Failed to create symlink %s -> %s: %v", link, target, err)
	}
}

// Chdir changes to a directory and returns a cleanup function to restore the original
func Chdir(t *testing.T, dir string) func() {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change to %s: %v", dir, err)
	}

	return func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}
}

// MkdirAll creates a directory and all parent directories
func MkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", path, err)
	}
}

// CreateRepoStructure creates a basic homelab repository structure for testing
func CreateRepoStructure(t *testing.T) {
	t.Helper()

	dirs := []string{
		"stacks",
		"enabled",
		"inventory",
		"secrets",
		"runtime",
	}

	for _, dir := range dirs {
		MkdirAll(t, dir)
	}

	// Create required inventory/vars.yaml
	WriteFile(t, "inventory/vars.yaml", "domain: test.local\ntimezone: UTC\n")
}

// CreateStack creates a test stack with the given name and dependencies
func CreateStack(t *testing.T, name string, requires []string, services []string) {
	t.Helper()

	stackDir := filepath.Join("stacks", name)
	MkdirAll(t, stackDir)

	// Build stack.yaml content
	content := "name: " + name + "\n"
	content += "category: other\n"

	if len(requires) > 0 {
		content += "requires:\n"
		for _, req := range requires {
			content += "  - " + req + "\n"
		}
	} else {
		content += "requires: []\n"
	}

	if len(services) == 0 {
		services = []string{"app"}
	}

	content += "services:\n"
	for _, svc := range services {
		content += "  - " + svc + "\n"
	}

	content += "vars:\n"
	for _, svc := range services {
		content += "  " + svc + ":\n"
		content += "    image: nginx:latest\n"
		content += "    hostname: " + svc + "\n"
		content += "    port: 80\n"
	}

	WriteFile(t, filepath.Join(stackDir, "stack.yaml"), content)

	// Create minimal compose template
	WriteFile(t, filepath.Join(stackDir, "compose.yml.tmpl"), "services:\n")
}

// EnableStack creates a symlink to enable a stack
func EnableStack(t *testing.T, name string) {
	t.Helper()

	target := filepath.Join("..", "stacks", name)
	link := filepath.Join("enabled", name)

	CreateSymlink(t, target, link)
}
