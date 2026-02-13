package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "homelabctl-fs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Save original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		_ = os.Chdir(originalDir)
		_ = os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func createRepoStructure(t *testing.T) {
	t.Helper()

	dirs := []string{
		"stacks",
		"enabled",
		"inventory",
		"secrets",
		"runtime",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create %s: %v", dir, err)
		}
	}

	// Create inventory/vars.yaml (required file)
	varsContent := "domain: test.local\n"
	if err := os.WriteFile("inventory/vars.yaml", []byte(varsContent), 0644); err != nil {
		t.Fatalf("Failed to create inventory/vars.yaml: %v", err)
	}
}

func TestVerifyRepository_Valid(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	createRepoStructure(t)

	err := VerifyRepository()
	if err != nil {
		t.Errorf("VerifyRepository() should pass for valid repo, got: %v", err)
	}
}

func TestVerifyRepository_MissingDirectories(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create only some directories
	_ = os.MkdirAll("stacks", 0755)
	_ = os.MkdirAll("enabled", 0755)
	// Missing: inventory, secrets, runtime

	err := VerifyRepository()
	if err == nil {
		t.Error("VerifyRepository() should fail for incomplete repo")
	}
}

func TestStackExists(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	createRepoStructure(t)

	// Create a test stack
	stackDir := "stacks/test-stack"
	if err := os.MkdirAll(stackDir, 0755); err != nil {
		t.Fatalf("Failed to create stack dir: %v", err)
	}

	tests := []struct {
		name      string
		stackName string
		want      bool
	}{
		{
			name:      "existing stack",
			stackName: "test-stack",
			want:      true,
		},
		{
			name:      "non-existent stack",
			stackName: "nonexistent",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StackExists(tt.stackName)
			if got != tt.want {
				t.Errorf("StackExists(%s) = %v, want %v", tt.stackName, got, tt.want)
			}
		})
	}
}

func TestGetAvailableStacks(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	createRepoStructure(t)

	// Create test stacks
	testStacks := []string{"stack1", "stack2", "stack3"}
	for _, stack := range testStacks {
		stackDir := filepath.Join("stacks", stack)
		if err := os.MkdirAll(stackDir, 0755); err != nil {
			t.Fatalf("Failed to create stack %s: %v", stack, err)
		}
	}

	// Create a file (should be ignored)
	if err := os.WriteFile("stacks/not-a-stack.txt", []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	stacks, err := GetAvailableStacks()
	if err != nil {
		t.Fatalf("GetAvailableStacks() error = %v", err)
	}

	if len(stacks) != len(testStacks) {
		t.Errorf("GetAvailableStacks() returned %d stacks, want %d", len(stacks), len(testStacks))
	}

	// Check that all test stacks are present
	stackMap := make(map[string]bool)
	for _, stack := range stacks {
		stackMap[stack] = true
	}

	for _, expected := range testStacks {
		if !stackMap[expected] {
			t.Errorf("Expected stack %q not found in results", expected)
		}
	}

	// Verify the file was not included
	if stackMap["not-a-stack.txt"] {
		t.Error("File should not be included in stacks list")
	}
}

func TestGetEnabledStacks(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	createRepoStructure(t)

	// Create test stacks
	_ = os.MkdirAll("stacks/stack1", 0755)
	_ = os.MkdirAll("stacks/stack2", 0755)

	// Create symlinks in enabled/
	_ = os.Symlink("../stacks/stack1", "enabled/stack1")
	_ = os.Symlink("../stacks/stack2", "enabled/stack2")

	stacks, err := GetEnabledStacks()
	if err != nil {
		t.Fatalf("GetEnabledStacks() error = %v", err)
	}

	if len(stacks) != 2 {
		t.Errorf("GetEnabledStacks() returned %d stacks, want 2", len(stacks))
	}

	// Check that both stacks are present
	stackMap := make(map[string]bool)
	for _, stack := range stacks {
		stackMap[stack] = true
	}

	if !stackMap["stack1"] || !stackMap["stack2"] {
		t.Errorf("Expected stacks not found, got: %v", stacks)
	}
}

func TestEnableStack(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	createRepoStructure(t)

	// Create a test stack
	stackName := "test-stack"
	_ = os.MkdirAll(filepath.Join("stacks", stackName), 0755)

	// Enable it
	err := EnableStack(stackName)
	if err != nil {
		t.Fatalf("EnableStack() error = %v", err)
	}

	// Verify symlink was created
	linkPath := filepath.Join("enabled", stackName)
	if _, err := os.Lstat(linkPath); err != nil {
		t.Errorf("Symlink not created: %v", err)
	}

	// Verify it's actually a symlink
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Errorf("Not a symlink: %v", err)
	}

	if target == "" {
		t.Error("Symlink target is empty")
	}
}

func TestDisableStack(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	createRepoStructure(t)

	// Create and enable a test stack
	stackName := "test-stack"
	_ = os.MkdirAll(filepath.Join("stacks", stackName), 0755)
	_ = os.Symlink("../stacks/"+stackName, filepath.Join("enabled", stackName))

	// Disable it
	err := DisableStack(stackName)
	if err != nil {
		t.Fatalf("DisableStack() error = %v", err)
	}

	// Verify symlink was removed
	linkPath := filepath.Join("enabled", stackName)
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Error("Symlink should have been removed")
	}
}

func TestEnableStack_AlreadyEnabled(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	createRepoStructure(t)

	stackName := "test-stack"
	_ = os.MkdirAll(filepath.Join("stacks", stackName), 0755)

	// Enable once
	_ = EnableStack(stackName)

	// Try to enable again
	err := EnableStack(stackName)
	if err == nil {
		t.Error("EnableStack() should return error when already enabled")
	}
}

func TestDisableStack_NotEnabled(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	createRepoStructure(t)

	stackName := "test-stack"
	_ = os.MkdirAll(filepath.Join("stacks", stackName), 0755)

	// Try to disable without enabling first
	err := DisableStack(stackName)
	if err == nil {
		t.Error("DisableStack() should return error when not enabled")
	}
}

func TestEnableStack_NonExistent(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	createRepoStructure(t)

	// Try to enable non-existent stack
	err := EnableStack("nonexistent")
	if err == nil {
		t.Error("EnableStack() should return error for non-existent stack")
	}
}
