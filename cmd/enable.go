package cmd

import (
	"fmt"

	"homelabctl/internal/errors"
	"homelabctl/internal/fs"
	"homelabctl/internal/inventory"
	"homelabctl/internal/stacks"
)

// Enable enables a stack or service
func Enable(args []string) error {
	// Parse flags
	isService := false
	suggestCategory := false
	var name string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-s", "--service":
			isService = true
		case "--suggest-category":
			suggestCategory = true
		default:
			if name == "" {
				name = args[i]
			} else {
				return fmt.Errorf("unexpected argument: %s", args[i])
			}
		}
	}

	if name == "" {
		if isService {
			return fmt.Errorf("usage: homelabctl enable -s <service>")
		}
		return fmt.Errorf("usage: homelabctl enable <stack> [--suggest-category]")
	}

	if err := fs.VerifyRepository(); err != nil {
		return err
	}

	if isService {
		return enableService(name)
	}
	return enableStack(name, suggestCategory)
}

func enableStack(stackName string, suggestCategory bool) error {
	// Check if stack exists
	if !fs.StackExists(stackName) {
		// Get available stacks
		availableStacks, _ := fs.GetAvailableStacks()

		suggestions := []string{
			"Run: homelabctl list",
			"Check stacks/ directory for available stacks",
		}

		// Add similar stack suggestions if possible
		if len(availableStacks) > 0 {
			context := []string{
				"Available stacks:",
			}
			for _, s := range availableStacks {
				context = append(context, fmt.Sprintf("  - %s", s))
			}

			return errors.New(
				fmt.Sprintf("stack '%s' does not exist", stackName),
				suggestions...,
			).WithContext(context...)
		}

		return errors.New(
			fmt.Sprintf("stack '%s' does not exist", stackName),
			suggestions...,
		)
	}

	// Get currently enabled stacks
	enabled, err := fs.GetEnabledStacks()
	if err != nil {
		return err
	}

	// Check dependencies
	if err := stacks.CheckDependenciesForStack(stackName, enabled); err != nil {
		return err
	}

	// Suggest category if requested
	if suggestCategory {
		suggestion, err := stacks.SuggestCategoryForStack(stackName)
		if err != nil {
			return err
		}

		stack, _ := stacks.LoadStack(stackName)
		if stack != nil && stack.Category != suggestion {
			fmt.Printf("⚠ Current category: %s\n", stack.Category)
			fmt.Printf("⚠ Suggested category: %s (based on dependencies)\n", suggestion)
			fmt.Printf("  Consider updating stacks/%s/stack.yaml\n\n", stackName)
		}
	}

	// Enable the stack
	if err := fs.EnableStack(stackName); err != nil {
		return err
	}

	fmt.Printf("✓ Enabled stack: %s\n", stackName)
	return nil
}

func enableService(serviceName string) error {
	// Get enabled stacks
	enabled, err := fs.GetEnabledStacks()
	if err != nil {
		return err
	}

	// Check if service exists in any enabled stack
	exists, stackName := stacks.ServiceExists(serviceName, enabled)
	if !exists {
		// Get all available services
		allServices, err := stacks.GetAllServicesFromStacks(enabled)
		if err != nil {
			return err
		}

		suggestions := []string{
			"Run: homelabctl list",
			"Check that the service's stack is enabled",
		}

		context := []string{
			"Available services in enabled stacks:",
		}
		for svc, stack := range allServices {
			context = append(context, fmt.Sprintf("  - %s (from %s)", svc, stack))
		}

		return errors.New(
			fmt.Sprintf("service '%s' not found in enabled stacks", serviceName),
			suggestions...,
		).WithContext(context...)
	}

	// Re-enable the service (remove from disabled list)
	if err := inventory.EnableService(serviceName); err != nil {
		// Handle case where service is not disabled
		if err.Error() == "service not disabled" || err.Error() == fmt.Sprintf("service '%s' is not disabled", serviceName) {
			return errors.New(
				fmt.Sprintf("service '%s' is not disabled", serviceName),
				fmt.Sprintf("Service is already enabled in stack '%s'", stackName),
				"Use 'homelabctl list' to see disabled services",
			)
		}
		return err
	}

	fmt.Printf("✓ Enabled service: %s (from stack: %s)\n", serviceName, stackName)
	fmt.Println("  Run 'homelabctl deploy' to apply changes")
	return nil
}
