package cmd

import (
	"fmt"

	"github.com/monkeymonk/homelabctl/internal/categories"
	"github.com/monkeymonk/homelabctl/internal/fs"
	"github.com/monkeymonk/homelabctl/internal/inventory"
	"github.com/monkeymonk/homelabctl/internal/stacks"
)

// categoryColor returns a colored category badge
func categoryColor(catName string) string {
	cat, _ := categories.Get(catName)
	if cat == nil {
		return catName
	}

	// Simple color mapping
	switch cat.Color {
	case "blue":
		return "\033[34m" + cat.DisplayName + "\033[0m"
	case "magenta":
		return "\033[35m" + cat.DisplayName + "\033[0m"
	case "cyan":
		return "\033[36m" + cat.DisplayName + "\033[0m"
	default:
		return cat.DisplayName
	}
}

// List shows enabled stacks grouped by category
func List() error {
	if err := fs.VerifyRepository(); err != nil {
		return err
	}

	enabled, err := fs.GetEnabledStacks()
	if err != nil {
		return err
	}

	if len(enabled) == 0 {
		fmt.Println("No stacks enabled")
		fmt.Println("\nRun: homelabctl enable <stack>")
		return nil
	}

	// Group by category
	groups, err := stacks.GroupByCategory(enabled)
	if err != nil {
		return err
	}

	// Load disabled services
	disabledServices, err := inventory.GetDisabledServices()
	if err != nil {
		return err
	}

	fmt.Println("Enabled stacks:")
	fmt.Println()

	// Display in category order
	for _, cat := range categories.AllCategories() {
		stacksInCat, exists := groups[cat.Name]
		if !exists || len(stacksInCat) == 0 {
			continue
		}

		// Category header
		fmt.Printf("  %s (%d):\n", categoryColor(cat.Name), len(stacksInCat))

		// List stacks in this category
		for _, stackName := range stacksInCat {
			fmt.Printf("    • %s\n", stackName)

			// Show disabled services for this stack
			stack, _ := stacks.LoadStack(stackName)
			if stack != nil {
				for _, svc := range stack.Services {
					for _, disabled := range disabledServices {
						if svc == disabled {
							fmt.Printf("      ⨯ %s (disabled)\n", svc)
						}
					}
				}
			}
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("Total: %d stack(s) enabled", len(enabled))
	if len(disabledServices) > 0 {
		fmt.Printf(", %d service(s) disabled", len(disabledServices))
	}
	fmt.Println()

	return nil
}
