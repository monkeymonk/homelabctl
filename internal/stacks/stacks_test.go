package stacks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestStacksForDeps creates test stack definitions for dependency testing
func setupTestStacksForDeps(t *testing.T) func() {
	t.Helper()

	// Save original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "homelabctl-deps-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create stacks directory
	stacksDir := "stacks"
	if err := os.MkdirAll(stacksDir, 0755); err != nil {
		t.Fatalf("Failed to create stacks dir: %v", err)
	}

	// Create test stacks
	stacks := map[string]struct {
		requires []string
	}{
		"standalone": {requires: []string{}},
		"core":       {requires: []string{}},
		"monitoring": {requires: []string{"core"}},
		"databases":  {requires: []string{"core"}},
		"app":        {requires: []string{"core", "databases"}},
	}

	for name, config := range stacks {
		stackDir := filepath.Join("stacks", name)
		if err := os.MkdirAll(stackDir, 0755); err != nil {
			t.Fatalf("Failed to create stack dir %s: %v", name, err)
		}

		// Write stack.yaml
		content := "name: " + name + "\n"
		content += "category: other\n"
		if len(config.requires) > 0 {
			content += "requires:\n"
			for _, req := range config.requires {
				content += "  - " + req + "\n"
			}
		} else {
			content += "requires: []\n"
		}
		content += "services:\n  - app\n"
		content += "vars:\n  app:\n    image: nginx\n"

		stackFile := filepath.Join(stackDir, "stack.yaml")
		if err := os.WriteFile(stackFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write stack.yaml for %s: %v", name, err)
		}
	}

	// Return cleanup function
	return func() {
		os.Chdir(originalDir)
		os.RemoveAll(tmpDir)
	}
}

func TestValidateDependencies_Valid(t *testing.T) {
	cleanup := setupTestStacksForDeps(t)
	defer cleanup()

	tests := []struct {
		name    string
		enabled []string
		wantErr bool
	}{
		{
			name:    "no dependencies",
			enabled: []string{"standalone"},
			wantErr: false,
		},
		{
			name:    "dependencies satisfied",
			enabled: []string{"core", "monitoring"},
			wantErr: false,
		},
		{
			name:    "multiple dependencies satisfied",
			enabled: []string{"core", "databases", "app"},
			wantErr: false,
		},
		{
			name:    "all stacks enabled",
			enabled: []string{"standalone", "core", "monitoring", "databases", "app"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDependencies(tt.enabled)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDependencies() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDependencies_Missing(t *testing.T) {
	cleanup := setupTestStacksForDeps(t)
	defer cleanup()

	tests := []struct {
		name        string
		enabled     []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "missing single dependency",
			enabled:     []string{"monitoring"},
			wantErr:     true,
			errContains: "requires",
		},
		{
			name:        "missing multiple dependencies",
			enabled:     []string{"app"},
			wantErr:     true,
			errContains: "requires",
		},
		{
			name:        "partial dependencies satisfied",
			enabled:     []string{"core", "app"},
			wantErr:     true,
			errContains: "databases",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDependencies(tt.enabled)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDependencies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidateDependencies() error = %v, should contain %q", err, tt.errContains)
			}
		})
	}
}

func TestCheckDependenciesForStack(t *testing.T) {
	cleanup := setupTestStacksForDeps(t)
	defer cleanup()

	tests := []struct {
		name        string
		stackName   string
		enabled     []string
		wantErr     bool
		errContains string
	}{
		{
			name:      "no dependencies",
			stackName: "standalone",
			enabled:   []string{},
			wantErr:   false,
		},
		{
			name:      "dependencies satisfied",
			stackName: "monitoring",
			enabled:   []string{"core"},
			wantErr:   false,
		},
		{
			name:        "missing dependency",
			stackName:   "monitoring",
			enabled:     []string{},
			wantErr:     true,
			errContains: "core",
		},
		{
			name:        "partial dependencies",
			stackName:   "app",
			enabled:     []string{"core"},
			wantErr:     true,
			errContains: "databases",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckDependenciesForStack(tt.stackName, tt.enabled)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckDependenciesForStack() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("CheckDependenciesForStack() error = %v, should contain %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestValidateServiceDefinitions_Valid(t *testing.T) {
	cleanup := setupTestStacksForDeps(t)
	defer cleanup()

	tests := []struct {
		name      string
		stackName string
		wantErr   bool
	}{
		{
			name:      "valid stack with service",
			stackName: "core",
			wantErr:   false,
		},
		{
			name:      "another valid stack",
			stackName: "monitoring",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServiceDefinitions(tt.stackName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateServiceDefinitions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateServiceDefinitions_Missing(t *testing.T) {
	// Create a test stack with missing service vars
	cleanup := setupTestStacksForDeps(t)
	defer cleanup()

	// Create a stack with service in list but not in vars
	stackDir := "stacks/badstack"
	if err := os.MkdirAll(stackDir, 0755); err != nil {
		t.Fatalf("Failed to create stack dir: %v", err)
	}

	content := `name: badstack
category: other
services:
  - missing-service
  - another-missing
vars:
  different-service:
    image: nginx
`
	stackFile := filepath.Join(stackDir, "stack.yaml")
	if err := os.WriteFile(stackFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write stack.yaml: %v", err)
	}

	err := ValidateServiceDefinitions("badstack")
	if err == nil {
		t.Error("ValidateServiceDefinitions() should return error for missing service vars")
	}

	if !strings.Contains(err.Error(), "missing-service") && !strings.Contains(err.Error(), "another-missing") {
		t.Errorf("Error should mention missing service, got: %v", err)
	}
}

func TestServiceExists(t *testing.T) {
	cleanup := setupTestStacksForDeps(t)
	defer cleanup()

	tests := []struct {
		name        string
		serviceName string
		enabled     []string
		wantExists  bool
		wantStack   string
	}{
		{
			name:        "service exists in enabled stack",
			serviceName: "app",
			enabled:     []string{"core"},
			wantExists:  true,
			wantStack:   "core",
		},
		{
			name:        "service does not exist",
			serviceName: "nonexistent",
			enabled:     []string{"core"},
			wantExists:  false,
			wantStack:   "",
		},
		{
			name:        "no stacks enabled",
			serviceName: "app",
			enabled:     []string{},
			wantExists:  false,
			wantStack:   "",
		},
		{
			name:        "service in one of multiple stacks",
			serviceName: "app",
			enabled:     []string{"core", "monitoring", "databases"},
			wantExists:  true,
			wantStack:   "", // Don't check which stack - order not guaranteed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, stack := ServiceExists(tt.serviceName, tt.enabled)
			if exists != tt.wantExists {
				t.Errorf("ServiceExists() exists = %v, want %v", exists, tt.wantExists)
			}
			// Only check stack name if we expect a specific one
			if tt.wantStack != "" && stack != tt.wantStack {
				t.Errorf("ServiceExists() stack = %v, want %v", stack, tt.wantStack)
			}
			// If service should exist but wantStack is empty, just verify stack is not empty
			if tt.wantExists && tt.wantStack == "" && stack == "" {
				t.Error("ServiceExists() returned empty stack but service should exist")
			}
		})
	}
}
