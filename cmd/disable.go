package cmd

import (
	"fmt"

	"github.com/monkeymonk/homelabctl/internal/errors"
	"github.com/monkeymonk/homelabctl/internal/fs"
	"github.com/monkeymonk/homelabctl/internal/inventory"
	"github.com/monkeymonk/homelabctl/internal/stacks"
)

// Disable disables a stack or service
func Disable(args []string) error {
	// Parse flags
	isService := false
	var name string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-s", "--service":
			isService = true
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
			return fmt.Errorf("usage: homelabctl disable -s <service>")
		}
		return fmt.Errorf("usage: homelabctl disable <stack>")
	}

	if err := fs.VerifyRepository(); err != nil {
		return err
	}

	if isService {
		return disableService(name)
	}
	return disableStack(name)
}

func disableStack(stackName string) error {
	// Disable the stack
	if err := fs.DisableStack(stackName); err != nil {
		return err
	}

	fmt.Printf("✓ Disabled stack: %s\n", stackName)
	fmt.Println("  Warning: This does not check if other stacks depend on this one")
	return nil
}

func disableService(serviceName string) error {
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

	// Disable the service (add to disabled list)
	if err := inventory.DisableService(serviceName); err != nil {
		// Handle case where service is already disabled
		if err.Error() == "service already disabled" || err.Error() == fmt.Sprintf("service '%s' is already disabled", serviceName) {
			return errors.New(
				fmt.Sprintf("service '%s' is already disabled", serviceName),
				"Use 'homelabctl list' to see disabled services",
				fmt.Sprintf("Run: homelabctl enable -s %s", serviceName),
			)
		}
		return err
	}

	fmt.Printf("✓ Disabled service: %s (from stack: %s)\n", serviceName, stackName)
	fmt.Println("  Run 'homelabctl deploy' to apply changes")
	return nil
}
