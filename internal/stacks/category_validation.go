package stacks

import (
	"fmt"

	"github.com/monkeymonk/homelabctl/internal/categories"
)

// ValidateCategoryDependencies ensures dependency order respects category hierarchy
// Rule: A stack can only depend on stacks in the same or lower-order categories
func ValidateCategoryDependencies(stackNames []string) error {
	for _, stackName := range stackNames {
		stack, err := LoadStack(stackName)
		if err != nil {
			return err
		}

		stackCat, err := categories.Get(stack.Category)
		if err != nil {
			return err
		}

		// Check each dependency
		for _, depName := range stack.Requires {
			depStack, err := LoadStack(depName)
			if err != nil {
				// Dependency doesn't exist - will be caught by normal validation
				continue
			}

			depCat, err := categories.Get(depStack.Category)
			if err != nil {
				continue
			}

			// Violation: depending on higher-order category
			if depCat.Order > stackCat.Order {
				return fmt.Errorf(
					"invalid category dependency in stack '%s': %s (category: %s, order: %d) depends on %s (category: %s, order: %d)\n"+
						"Category order: Infrastructure(1) → Automation(2) → Media(3) → Other(4)\n"+
						"To resolve:\n"+
						"  - Move %s to category '%s' or lower\n"+
						"  - Or move %s to category '%s' or higher\n"+
						"  - Or remove the dependency from stacks/%s/stack.yaml",
					stackName,
					stackName, stackCat.DisplayName, stackCat.Order,
					depName, depCat.DisplayName, depCat.Order,
					depName, stackCat.Name,
					stackName, depCat.Name,
					stackName,
				)
			}
		}
	}

	return nil
}

// SuggestCategoryForStack suggests the best category based on dependencies
func SuggestCategoryForStack(stackName string) (string, error) {
	stack, err := LoadStack(stackName)
	if err != nil {
		return "", err
	}

	if len(stack.Requires) == 0 {
		return "core", nil // No dependencies = core infrastructure
	}

	// Find maximum order among dependencies
	maxOrder := 0
	for _, depName := range stack.Requires {
		depStack, err := LoadStack(depName)
		if err != nil {
			continue
		}
		order := categories.GetOrder(depStack.Category)
		if order > maxOrder {
			maxOrder = order
		}
	}

	// Suggest same order or higher
	for _, cat := range categories.AllCategories() {
		if cat.Order >= maxOrder {
			return cat.Name, nil
		}
	}

	return "tools", nil
}
