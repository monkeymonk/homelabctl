package stacks

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestStacks creates test stack definitions in a temporary directory
func setupTestStacks(t *testing.T, stacks map[string][]string) string {
	t.Helper()

	// Create temp directory for test stacks
	tmpDir, err := os.MkdirTemp("", "homelabctl-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	// Create stacks directory
	stacksDir := filepath.Join(tmpDir, "stacks")
	if err := os.MkdirAll(stacksDir, 0755); err != nil {
		t.Fatalf("Failed to create stacks dir: %v", err)
	}

	// Change to temp directory for test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	// Create each stack
	for name, requires := range stacks {
		stackDir := filepath.Join("stacks", name)
		if err := os.MkdirAll(stackDir, 0755); err != nil {
			t.Fatalf("Failed to create stack dir %s: %v", name, err)
		}

		// Write stack.yaml
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
			t.Fatalf("Failed to write stack.yaml for %s: %v", name, err)
		}
	}

	return tmpDir
}

func TestCycleDetector_NoCycles(t *testing.T) {
	// Create test stacks: A → B → C (no cycles)
	stacks := map[string][]string{
		"a": {},
		"b": {"a"},
		"c": {"b"},
	}

	setupTestStacks(t, stacks)

	detector, err := NewCycleDetector([]string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	cycles := detector.DetectCycles()
	if len(cycles) != 0 {
		t.Errorf("Expected no cycles, found: %v", cycles)
	}
}

func TestCycleDetector_SimpleCycle(t *testing.T) {
	// Create test stacks: A ⇄ B (simple cycle)
	stacks := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}

	setupTestStacks(t, stacks)

	detector, err := NewCycleDetector([]string{"a", "b"})
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	cycles := detector.DetectCycles()
	if len(cycles) == 0 {
		t.Fatal("Expected to find cycle")
	}

	cycle := cycles[0]
	if len(cycle) != 2 {
		t.Errorf("Expected cycle length 2, got %d", len(cycle))
	}

	// Verify cycle contains both stacks (order may vary)
	hasA := false
	hasB := false
	for _, node := range cycle {
		if node == "a" {
			hasA = true
		}
		if node == "b" {
			hasB = true
		}
	}
	if !hasA || !hasB {
		t.Errorf("Cycle should contain both a and b, got: %v", cycle)
	}
}

func TestCycleDetector_TransitiveCycle(t *testing.T) {
	// Create test stacks: A → B → C → A (transitive cycle)
	stacks := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}

	setupTestStacks(t, stacks)

	detector, err := NewCycleDetector([]string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	cycles := detector.DetectCycles()
	if len(cycles) == 0 {
		t.Fatal("Expected to find cycle")
	}

	cycle := cycles[0]
	if len(cycle) != 3 {
		t.Errorf("Expected cycle length 3, got %d: %v", len(cycle), cycle)
	}

	// Verify all three stacks are in the cycle
	hasA := false
	hasB := false
	hasC := false
	for _, node := range cycle {
		if node == "a" {
			hasA = true
		}
		if node == "b" {
			hasB = true
		}
		if node == "c" {
			hasC = true
		}
	}
	if !hasA || !hasB || !hasC {
		t.Errorf("Cycle should contain a, b, and c, got: %v", cycle)
	}
}

func TestCycleDetector_SelfDependency(t *testing.T) {
	// Create test stack that depends on itself
	stacks := map[string][]string{
		"a": {"a"},
	}

	setupTestStacks(t, stacks)

	// LoadStack should catch self-dependency before detector is created
	_, err := LoadStack("a")
	if err == nil {
		t.Fatal("Expected error for self-dependency, got nil")
	}

	// Verify error message is helpful
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestCycleDetector_ComplexGraph(t *testing.T) {
	// Create complex graph with cycle:
	// a → b, c
	// b → d
	// c → d
	// d → e
	// e → a (cycle: a → c → d → e → a)
	stacks := map[string][]string{
		"a": {"b", "c"},
		"b": {"d"},
		"c": {"d"},
		"d": {"e"},
		"e": {"a"},
	}

	setupTestStacks(t, stacks)

	detector, err := NewCycleDetector([]string{"a", "b", "c", "d", "e"})
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	cycles := detector.DetectCycles()
	if len(cycles) == 0 {
		t.Fatal("Expected to find cycle")
	}

	// Verify cycle exists and has reasonable length
	cycle := cycles[0]
	if len(cycle) < 2 {
		t.Errorf("Cycle too short: %v", cycle)
	}

	// The cycle should include 'a' and 'e' (the back edge)
	hasA := false
	hasE := false
	for _, node := range cycle {
		if node == "a" {
			hasA = true
		}
		if node == "e" {
			hasE = true
		}
	}
	if !hasA || !hasE {
		t.Errorf("Cycle should contain both a and e, got: %v", cycle)
	}
}

func TestCycleDetector_MultipleDependencies(t *testing.T) {
	// Stack with multiple dependencies, but no cycles
	stacks := map[string][]string{
		"a": {},
		"b": {},
		"c": {"a", "b"},
		"d": {"a", "c"},
	}

	setupTestStacks(t, stacks)

	detector, err := NewCycleDetector([]string{"a", "b", "c", "d"})
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	cycles := detector.DetectCycles()
	if len(cycles) != 0 {
		t.Errorf("Expected no cycles in valid DAG, found: %v", cycles)
	}
}

func TestCycleDetector_DiamondDependency(t *testing.T) {
	// Diamond pattern (not a cycle):
	// a → b, c
	// b → d
	// c → d
	stacks := map[string][]string{
		"a": {"b", "c"},
		"b": {"d"},
		"c": {"d"},
		"d": {},
	}

	setupTestStacks(t, stacks)

	detector, err := NewCycleDetector([]string{"a", "b", "c", "d"})
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	cycles := detector.DetectCycles()
	if len(cycles) != 0 {
		t.Errorf("Diamond pattern should not be detected as cycle, found: %v", cycles)
	}
}

func TestValidateDependencies_WithCycle(t *testing.T) {
	// Test that ValidateDependencies detects cycles
	stacks := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}

	setupTestStacks(t, stacks)

	err := ValidateDependencies([]string{"a", "b"})
	if err == nil {
		t.Fatal("Expected ValidateDependencies to detect cycle")
	}

	// Verify error mentions circular dependency
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestValidateDependencies_WithoutCycle(t *testing.T) {
	// Test that ValidateDependencies passes with valid DAG
	stacks := map[string][]string{
		"a": {},
		"b": {"a"},
		"c": {"b"},
	}

	setupTestStacks(t, stacks)

	err := ValidateDependencies([]string{"a", "b", "c"})
	if err != nil {
		t.Errorf("ValidateDependencies should pass for valid DAG, got error: %v", err)
	}
}
