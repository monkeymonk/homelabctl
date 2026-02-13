package stacks

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"homelabctl/internal/categories"
	"homelabctl/internal/errors"
	"homelabctl/internal/paths"
)

// Stack represents a stack.yaml manifest
type Stack struct {
	Name        string                 `yaml:"name"`
	Category    string                 `yaml:"category"`
	Requires    []string               `yaml:"requires"`
	Services    []string               `yaml:"services"`
	Vars        map[string]interface{} `yaml:"vars"`
	Persistence struct {
		Volumes []string `yaml:"volumes"`
		Paths   []string `yaml:"paths"`
	} `yaml:"persistence"`
}

// LoadStack reads and parses a stack.yaml file
func LoadStack(name string) (*Stack, error) {
	stackPath := paths.StackYAMLPath(name)

	data, err := os.ReadFile(stackPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read stack.yaml for %s: %w", name, err)
	}

	var stack Stack
	if err := yaml.Unmarshal(data, &stack); err != nil {
		return nil, fmt.Errorf("failed to parse stack.yaml for %s: %w", name, err)
	}

	// Validate
	if stack.Name == "" {
		return nil, fmt.Errorf("stack.yaml for %s missing 'name' field", name)
	}

	if stack.Name != name {
		return nil, fmt.Errorf("stack.yaml name mismatch: directory=%s, name=%s", name, stack.Name)
	}

	// Validate category
	if stack.Category == "" {
		return nil, fmt.Errorf("stack.yaml for %s missing 'category' field", name)
	}

	if !categories.ValidCategoryName(stack.Category) {
		return nil, fmt.Errorf("invalid category '%s' in stack %s (category must be a non-empty string)", stack.Category, name)
	}

	// Register the category for dynamic discovery
	categories.RegisterCategory(stack.Category)

	// Temporary migration: if services list is missing, derive from vars keys
	if len(stack.Services) == 0 && len(stack.Vars) > 0 {
		fmt.Fprintf(os.Stderr, "WARNING: stack %s missing 'services' field, deriving from vars (deprecated)\n", name)
		for key := range stack.Vars {
			stack.Services = append(stack.Services, key)
		}
	}

	// After fallback, still require at least one service
	if len(stack.Services) == 0 {
		return nil, fmt.Errorf("stack.yaml for %s has no services defined", name)
	}

	// Validate dependencies - check for self-dependency
	for _, dep := range stack.Requires {
		if dep == name {
			return nil, errors.New(
				fmt.Sprintf("stack '%s' cannot depend on itself", name),
				fmt.Sprintf("Edit: stacks/%s/stack.yaml", name),
				fmt.Sprintf("Remove '%s' from requires list", name),
			)
		}
	}

	return &stack, nil
}

// ValidateDependencies checks that all dependencies are satisfied and no cycles exist
func ValidateDependencies(enabledStacks []string) error {
	// Build map for quick lookup
	enabled := EnabledStacksMap(enabledStacks)

	// Check each stack's dependencies are enabled
	for _, name := range enabledStacks {
		stack, err := LoadStack(name)
		if err != nil {
			return err
		}

		for _, dep := range stack.Requires {
			if !enabled[dep] {
				return fmt.Errorf("stack %s requires %s but it is not enabled", name, dep)
			}
		}
	}

	// Check for circular dependencies
	detector, err := NewCycleDetector(enabledStacks)
	if err != nil {
		return err
	}

	cycles := detector.DetectCycles()
	if len(cycles) > 0 {
		// Use enhanced error for the first cycle found
		return errors.DependencyCycle(cycles[0])
	}

	return nil
}

// CheckDependenciesForStack checks if enabling a stack would satisfy dependencies
func CheckDependenciesForStack(stackName string, enabledStacks []string) error {
	stack, err := LoadStack(stackName)
	if err != nil {
		return err
	}

	enabled := EnabledStacksMap(enabledStacks)

	var missing []string
	for _, dep := range stack.Requires {
		if !enabled[dep] {
			missing = append(missing, dep)
		}
	}

	if len(missing) > 0 {
		// Build suggestions
		suggestions := make([]string, 0, len(missing)+1)
		for _, dep := range missing {
			suggestions = append(suggestions, fmt.Sprintf("Run: homelabctl enable %s", dep))
		}
		suggestions = append(suggestions, fmt.Sprintf("Then run: homelabctl enable %s", stackName))
		suggestions = append(suggestions, fmt.Sprintf("Or remove dependencies in stacks/%s/stack.yaml", stackName))

		// Build context showing dependency chain
		context := []string{
			"Dependency chain:",
			fmt.Sprintf("  %s requires: %v", stackName, stack.Requires),
			fmt.Sprintf("  Missing: %v", missing),
		}

		return errors.New(
			fmt.Sprintf("stack '%s' has unsatisfied dependencies", stackName),
			suggestions...,
		).WithContext(context...)
	}

	return nil
}

// HasComposeTemplate checks if stack has compose.yml.tmpl
func HasComposeTemplate(name string) bool {
	composePath := paths.StackComposeTemplate(name)
	_, err := os.Stat(composePath)
	return err == nil
}

// GetStackVars returns the vars section from stack.yaml
func GetStackVars(name string) (map[string]interface{}, error) {
	stack, err := LoadStack(name)
	if err != nil {
		return nil, err
	}

	if stack.Vars == nil {
		return make(map[string]interface{}), nil
	}

	return stack.Vars, nil
}

// ValidateServiceDefinitions checks that all services have corresponding var definitions
func ValidateServiceDefinitions(name string) error {
	stack, err := LoadStack(name)
	if err != nil {
		return err
	}

	for _, serviceName := range stack.Services {
		if _, exists := stack.Vars[serviceName]; !exists {
			return fmt.Errorf("service '%s' listed in services but missing from vars section", serviceName)
		}
	}

	return nil
}

// GetServiceNames returns all service names from a stack's explicit services list
func GetServiceNames(name string) ([]string, error) {
	stack, err := LoadStack(name)
	if err != nil {
		return nil, err
	}

	return stack.Services, nil
}

// GetAllServicesFromStacks returns all service names from a list of stacks
func GetAllServicesFromStacks(stackNames []string) (map[string]string, error) {
	// Returns map of service name -> stack name
	services := make(map[string]string)

	for _, stackName := range stackNames {
		stackServices, err := GetServiceNames(stackName)
		if err != nil {
			return nil, err
		}
		for _, svc := range stackServices {
			services[svc] = stackName
		}
	}

	return services, nil
}

// ServiceExists checks if a service exists in any of the given stacks
func ServiceExists(serviceName string, enabledStacks []string) (bool, string) {
	services, err := GetAllServicesFromStacks(enabledStacks)
	if err != nil {
		return false, ""
	}

	if stackName, exists := services[serviceName]; exists {
		return true, stackName
	}

	return false, ""
}
