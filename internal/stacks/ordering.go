package stacks

import (
	"sort"

	"homelabctl/internal/categories"
)

// StackWithCategory pairs a stack name with its category info
type StackWithCategory struct {
	Name     string
	Category string
	Order    int
}

// SortByCategory sorts stack names by their category deployment order
func SortByCategory(stackNames []string) ([]string, error) {
	// Load category info for each stack
	stacksInfo := make([]StackWithCategory, 0, len(stackNames))

	for _, name := range stackNames {
		stack, err := LoadStack(name)
		if err != nil {
			return nil, err
		}

		stacksInfo = append(stacksInfo, StackWithCategory{
			Name:     name,
			Category: stack.Category,
			Order:    categories.GetOrder(stack.Category),
		})
	}

	// Sort by category order, then alphabetically within category
	sort.Slice(stacksInfo, func(i, j int) bool {
		if stacksInfo[i].Order != stacksInfo[j].Order {
			return stacksInfo[i].Order < stacksInfo[j].Order
		}
		return stacksInfo[i].Name < stacksInfo[j].Name
	})

	// Extract sorted names
	sorted := make([]string, len(stacksInfo))
	for i, s := range stacksInfo {
		sorted[i] = s.Name
	}

	return sorted, nil
}

// GroupByCategory groups stack names by their category
func GroupByCategory(stackNames []string) (map[string][]string, error) {
	groups := make(map[string][]string)

	for _, name := range stackNames {
		stack, err := LoadStack(name)
		if err != nil {
			return nil, err
		}

		groups[stack.Category] = append(groups[stack.Category], name)
	}

	// Sort within each group
	for cat := range groups {
		sort.Strings(groups[cat])
	}

	return groups, nil
}
