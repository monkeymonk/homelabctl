package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"homelabctl/internal/testutil"
)

// Integration tests for CLI commands
// These test the full command flow end-to-end

func TestEnableCommand(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create test stacks
	testutil.CreateStack(t, "core", []string{}, []string{"traefik"})
	testutil.CreateStack(t, "monitoring", []string{"core"}, []string{"grafana"})

	// Enable core stack
	err := Enable([]string{"core"})
	if err != nil {
		t.Fatalf("Enable(core) failed: %v", err)
	}

	// Verify symlink exists
	linkPath := filepath.Join("enabled", "core")
	if _, err := os.Lstat(linkPath); err != nil {
		t.Errorf("Symlink not created: %v", err)
	}

	// Try to enable stack with unsatisfied dependencies
	err = Enable([]string{"monitoring"})
	if err != nil {
		// This should succeed since core is now enabled
		t.Errorf("Enable(monitoring) should succeed: %v", err)
	}

	// Try to enable non-existent stack
	err = Enable([]string{"nonexistent"})
	if err == nil {
		t.Error("Enable(nonexistent) should fail")
	}

	// Try to enable already enabled stack
	err = Enable([]string{"core"})
	if err == nil {
		t.Error("Enable(core) again should fail")
	}
}

func TestDisableCommand(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create and enable test stack
	testutil.CreateStack(t, "core", []string{}, []string{"traefik"})
	testutil.EnableStack(t, "core")

	// Disable the stack
	err := Disable([]string{"core"})
	if err != nil {
		t.Fatalf("Disable(core) failed: %v", err)
	}

	// Verify symlink was removed
	linkPath := filepath.Join("enabled", "core")
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Error("Symlink should have been removed")
	}

	// Try to disable non-existent stack
	err = Disable([]string{"nonexistent"})
	if err == nil {
		t.Error("Disable(nonexistent) should fail")
	}

	// Try to disable already disabled stack
	err = Disable([]string{"core"})
	if err == nil {
		t.Error("Disable(core) again should fail")
	}
}

func TestListCommand(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create and enable test stacks
	testutil.CreateStack(t, "core", []string{}, []string{"traefik"})
	testutil.CreateStack(t, "monitoring", []string{"core"}, []string{"grafana"})
	testutil.EnableStack(t, "core")
	testutil.EnableStack(t, "monitoring")

	// List should succeed
	err := List()
	if err != nil {
		t.Errorf("List() failed: %v", err)
	}

	// Test with no enabled stacks
	os.Remove("enabled/core")
	os.Remove("enabled/monitoring")

	err = List()
	if err != nil {
		t.Errorf("List() should succeed with no stacks: %v", err)
	}
}

func TestValidateCommand(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create test stacks
	testutil.CreateStack(t, "core", []string{}, []string{"traefik"})
	testutil.CreateStack(t, "monitoring", []string{"core"}, []string{"grafana"})

	// Enable stacks
	testutil.EnableStack(t, "core")
	testutil.EnableStack(t, "monitoring")

	// Validate should succeed
	err := Validate()
	if err != nil {
		t.Errorf("Validate() failed: %v", err)
	}

	// Test validation with unsatisfied dependencies
	testutil.CreateStack(t, "broken", []string{"nonexistent"}, []string{"app"})
	testutil.EnableStack(t, "broken")

	err = Validate()
	if err == nil {
		t.Error("Validate() should fail with unsatisfied dependencies")
	}
}

func TestValidateCommand_NoCycle(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create stacks with circular dependency
	testutil.CreateStack(t, "stack-a", []string{"stack-b"}, []string{"app-a"})
	testutil.CreateStack(t, "stack-b", []string{"stack-a"}, []string{"app-b"})

	testutil.EnableStack(t, "stack-a")
	testutil.EnableStack(t, "stack-b")

	// Validate should detect cycle
	err := Validate()
	if err == nil {
		t.Error("Validate() should detect circular dependency")
	}
}

func TestValidateCommand_NoStacks(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Validate with no enabled stacks should fail
	err := Validate()
	if err == nil {
		t.Error("Validate() should fail with no enabled stacks")
	}
}

func TestValidateCommand_MissingStackYaml(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create stack directory without stack.yaml
	testutil.MkdirAll(t, "stacks/broken")

	// Create symlink
	target := filepath.Join("..", "stacks", "broken")
	link := filepath.Join("enabled", "broken")
	testutil.CreateSymlink(t, target, link)

	// Validate should fail
	err := Validate()
	if err == nil {
		t.Error("Validate() should fail with missing stack.yaml")
	}
}

func TestValidateCommand_MissingComposeTemplate(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create stack with stack.yaml but no compose.yml.tmpl
	testutil.CreateStack(t, "incomplete", []string{}, []string{"app"})

	// Remove compose template
	os.Remove("stacks/incomplete/compose.yml.tmpl")

	testutil.EnableStack(t, "incomplete")

	// Validate should fail
	err := Validate()
	if err == nil {
		t.Error("Validate() should fail with missing compose.yml.tmpl")
	}
}

func TestGenerateCommand(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create test stacks with full compose templates
	testutil.CreateStack(t, "core", []string{}, []string{"traefik"})

	// Update compose template to be valid
	composeContent := `services:
  traefik:
    image: {{ .vars.traefik.image }}
    container_name: traefik
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"

volumes:
  traefik_data:

networks:
  default:
`
	testutil.WriteFile(t, "stacks/core/compose.yml.tmpl", composeContent)

	testutil.EnableStack(t, "core")

	// Generate should succeed
	// Note: This requires gomplate to be installed
	// In CI, we might want to skip or mock this
	t.Skip("Skipping generate test - requires gomplate binary")

	err := Generate()
	if err != nil {
		t.Errorf("Generate() failed: %v", err)
	}

	// Verify runtime/docker-compose.yml was created
	outputPath := "runtime/docker-compose.yml"
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("Generate() should create %s: %v", outputPath, err)
	}
}

func TestGenerateCommand_InvalidRepository(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	// Don't create repository structure

	// Generate should fail with repository validation error
	err := Generate()
	if err == nil {
		t.Error("Generate() should fail in invalid repository")
	}
}

func TestEnableService(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create stack with multiple services
	testutil.CreateStack(t, "core", []string{}, []string{"traefik", "sablier"})
	testutil.EnableStack(t, "core")

	// First disable a service
	err := Disable([]string{"-s", "traefik"})
	if err != nil {
		t.Fatalf("Disable service failed: %v", err)
	}

	// Then enable it back
	err = Enable([]string{"-s", "traefik"})
	if err != nil {
		t.Errorf("Enable service failed: %v", err)
	}

	// Try to enable non-existent service
	err = Enable([]string{"-s", "nonexistent"})
	if err == nil {
		t.Error("Enable(nonexistent service) should fail")
	}

	// Try to enable already enabled service
	err = Enable([]string{"-s", "sablier"})
	if err == nil {
		t.Error("Enable(already enabled service) should fail")
	}
}

func TestDisableService(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create stack with multiple services
	testutil.CreateStack(t, "core", []string{}, []string{"traefik", "sablier"})
	testutil.EnableStack(t, "core")

	// Disable a service
	err := Disable([]string{"-s", "traefik"})
	if err != nil {
		t.Errorf("Disable service failed: %v", err)
	}

	// Verify service is in disabled list
	stateFile := "inventory/state.yaml"
	if _, err := os.Stat(stateFile); err != nil {
		t.Error("State file should be created")
	}

	// Try to disable non-existent service
	err = Disable([]string{"-s", "nonexistent"})
	if err == nil {
		t.Error("Disable(nonexistent service) should fail")
	}

	// Try to disable already disabled service
	err = Disable([]string{"-s", "traefik"})
	if err == nil {
		t.Error("Disable(already disabled service) should fail")
	}
}

func TestEnableWithDependencyCheck(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create stacks with dependencies
	testutil.CreateStack(t, "core", []string{}, []string{"traefik"})
	testutil.CreateStack(t, "databases", []string{"core"}, []string{"postgres"})
	testutil.CreateStack(t, "app", []string{"core", "databases"}, []string{"webapp"})

	// Try to enable app without dependencies
	err := Enable([]string{"app"})
	if err == nil {
		t.Error("Enable(app) should fail without dependencies")
	}

	// Enable in correct order
	if err := Enable([]string{"core"}); err != nil {
		t.Fatalf("Enable(core) failed: %v", err)
	}

	if err := Enable([]string{"databases"}); err != nil {
		t.Fatalf("Enable(databases) failed: %v", err)
	}

	if err := Enable([]string{"app"}); err != nil {
		t.Errorf("Enable(app) should succeed with dependencies: %v", err)
	}
}

func TestDisableWithDependents(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create stacks with dependencies
	testutil.CreateStack(t, "core", []string{}, []string{"traefik"})
	testutil.CreateStack(t, "app", []string{"core"}, []string{"webapp"})

	// Enable both
	testutil.EnableStack(t, "core")
	testutil.EnableStack(t, "app")

	// Disable core (has dependent)
	// Current implementation shows a warning but doesn't fail
	err := Disable([]string{"core"})
	if err != nil {
		t.Errorf("Disable(core) should succeed with warning: %v", err)
	}

	// Verify core is disabled
	linkPath := filepath.Join("enabled", "core")
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Error("Symlink should have been removed")
	}
}

func TestValidateServiceDefinitions(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	testutil.CreateRepoStructure(t)

	// Create stack with service in list but not in vars
	stackContent := `name: broken
category: other
requires: []
services:
  - app
  - missing
vars:
  app:
    image: nginx:latest
    hostname: app
    port: 80
`
	testutil.WriteFile(t, "stacks/broken/stack.yaml", stackContent)
	testutil.WriteFile(t, "stacks/broken/compose.yml.tmpl", "services:\n")
	testutil.EnableStack(t, "broken")

	// Validate should fail
	err := Validate()
	if err == nil {
		t.Error("Validate() should fail with missing service definition")
	}
}
