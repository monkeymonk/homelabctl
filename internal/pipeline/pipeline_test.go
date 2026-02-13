package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func setupPipelineTest(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "homelabctl-pipeline-test-*")
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

	// Create basic structure
	dirs := []string{"stacks", "enabled", "inventory", "secrets", "runtime"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create %s: %v", dir, err)
		}
	}

	// Create inventory/vars.yaml
	varsContent := "domain: test.local\ntimezone: UTC\n"
	if err := os.WriteFile("inventory/vars.yaml", []byte(varsContent), 0644); err != nil {
		t.Fatalf("Failed to create vars.yaml: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.Chdir(originalDir)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func createTestStack(t *testing.T, name string, requires []string) {
	t.Helper()

	stackDir := filepath.Join("stacks", name)
	if err := os.MkdirAll(stackDir, 0755); err != nil {
		t.Fatalf("Failed to create stack dir: %v", err)
	}

	// Create stack.yaml
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
	content += "services:\n  - app\n"
	content += "vars:\n  app:\n    image: nginx\n"

	stackFile := filepath.Join(stackDir, "stack.yaml")
	if err := os.WriteFile(stackFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write stack.yaml: %v", err)
	}
}

func TestNewPipeline(t *testing.T) {
	p := New()

	if p == nil {
		t.Fatal("New() returned nil")
	}

	if p.ctx == nil {
		t.Error("Pipeline context should not be nil")
	}
}

func TestPipeline_AddStage(t *testing.T) {
	p := New()

	stageCalled := false
	testStage := func(ctx *Context) error {
		stageCalled = true
		return nil
	}

	p.AddStage(testStage)

	// Execute pipeline to verify stage was added
	if err := p.Execute(); err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !stageCalled {
		t.Error("Stage was not called")
	}
}

func TestPipeline_Run_StagesInOrder(t *testing.T) {
	p := New()

	var order []int

	stage1 := func(ctx *Context) error {
		order = append(order, 1)
		return nil
	}

	stage2 := func(ctx *Context) error {
		order = append(order, 2)
		return nil
	}

	stage3 := func(ctx *Context) error {
		order = append(order, 3)
		return nil
	}

	p.AddStage(stage1)
	p.AddStage(stage2)
	p.AddStage(stage3)

	if err := p.Execute(); err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if len(order) != 3 {
		t.Errorf("Expected 3 stages to run, got %d", len(order))
	}

	for i, v := range order {
		expected := i + 1
		if v != expected {
			t.Errorf("Stage %d ran out of order, got %d", expected, v)
		}
	}
}

func TestPipeline_Run_StopOnError(t *testing.T) {
	p := New()

	var executed []int

	stage1 := func(ctx *Context) error {
		executed = append(executed, 1)
		return nil
	}

	stage2 := func(ctx *Context) error {
		executed = append(executed, 2)
		return os.ErrNotExist // Return error
	}

	stage3 := func(ctx *Context) error {
		executed = append(executed, 3)
		return nil
	}

	p.AddStage(stage1)
	p.AddStage(stage2)
	p.AddStage(stage3)

	err := p.Execute()
	if err == nil {
		t.Error("Execute() should return error when stage fails")
	}

	// Only stages 1 and 2 should have executed
	if len(executed) != 2 {
		t.Errorf("Expected 2 stages to execute, got %d", len(executed))
	}

	if executed[0] != 1 || executed[1] != 2 {
		t.Errorf("Unexpected execution order: %v", executed)
	}
}

func TestContext_SharedState(t *testing.T) {
	p := New()

	// Stage 1 sets some state
	stage1 := func(ctx *Context) error {
		ctx.EnabledStacks = []string{"stack1", "stack2"}
		return nil
	}

	// Stage 2 reads that state
	var readStacks []string
	stage2 := func(ctx *Context) error {
		readStacks = ctx.EnabledStacks
		return nil
	}

	p.AddStage(stage1)
	p.AddStage(stage2)

	if err := p.Execute(); err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if len(readStacks) != 2 {
		t.Errorf("Expected 2 stacks, got %d", len(readStacks))
	}
}

func TestLoadInventoryStage(t *testing.T) {
	_, cleanup := setupPipelineTest(t)
	defer cleanup()

	// Add custom vars
	varsContent := "domain: example.com\ncustom_var: test_value\n"
	if err := os.WriteFile("inventory/vars.yaml", []byte(varsContent), 0644); err != nil {
		t.Fatalf("Failed to write vars.yaml: %v", err)
	}

	p := New()
	p.AddStage(LoadInventoryStage())

	if err := p.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Check that inventory vars were loaded
	if p.ctx.InventoryVars == nil {
		t.Fatal("InventoryVars should not be nil")
	}

	if domain, ok := p.ctx.InventoryVars["domain"]; !ok {
		t.Error("domain should be in inventory vars")
	} else if domain != "example.com" {
		t.Errorf("domain = %v, want example.com", domain)
	}
}

func TestFilterServicesStage(t *testing.T) {
	p := New()

	// Setup context with disabled services
	p.ctx.DisabledServices = map[string]bool{
		"disabled1": true,
		"disabled2": true,
	}

	p.ctx.StackConfigs = map[string]*StackConfig{
		"stack1": {
			Name:     "stack1",
			Services: []string{"service1", "disabled1", "service2"},
			MergedVars: map[string]interface{}{
				"service1":  map[string]interface{}{"image": "nginx"},
				"disabled1": map[string]interface{}{"image": "redis"},
				"service2":  map[string]interface{}{"image": "postgres"},
			},
		},
		"stack2": {
			Name:     "stack2",
			Services: []string{"disabled2", "service3"},
			MergedVars: map[string]interface{}{
				"disabled2": map[string]interface{}{"image": "mysql"},
				"service3":  map[string]interface{}{"image": "mongo"},
			},
		},
	}

	p.AddStage(FilterServicesStage())

	if err := p.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Check that disabled services were filtered out from FilteredVars
	stack1Config := p.ctx.StackConfigs["stack1"]
	if stack1Config.FilteredVars != nil {
		if _, exists := stack1Config.FilteredVars["disabled1"]; exists {
			t.Error("disabled1 should have been filtered out")
		}

		if _, exists := stack1Config.FilteredVars["service1"]; !exists {
			t.Error("service1 should still exist in FilteredVars")
		}
	}

	stack2Config := p.ctx.StackConfigs["stack2"]
	if stack2Config.FilteredVars != nil {
		if _, exists := stack2Config.FilteredVars["disabled2"]; exists {
			t.Error("disabled2 should have been filtered out")
		}

		if _, exists := stack2Config.FilteredVars["service3"]; !exists {
			t.Error("service3 should still exist in FilteredVars")
		}
	}
}

func TestPipeline_EmptyPipeline(t *testing.T) {
	p := New()

	// Running empty pipeline should succeed
	if err := p.Execute(); err != nil {
		t.Errorf("Empty pipeline should not error, got: %v", err)
	}
}

func TestContext_Initialization(t *testing.T) {
	p := New()

	ctx := p.ctx

	if ctx == nil {
		t.Fatal("Context should not be nil")
	}

	// Check fields that are initialized in New()
	if ctx.DisabledServices == nil {
		t.Error("DisabledServices should be initialized")
	}

	if ctx.StackConfigs == nil {
		t.Error("StackConfigs should be initialized")
	}

	if ctx.RenderedFiles == nil {
		t.Error("RenderedFiles should be initialized")
	}

	if ctx.RenderedCompose == nil {
		t.Error("RenderedCompose should be initialized")
	}

	// EnabledStacks and InventoryVars are populated by stages, not in New()
}
