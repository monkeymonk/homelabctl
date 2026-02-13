package stacks

import (
	"homelabctl/internal/categories"
)

// MergeVariables merges variables according to the precedence rules:
// stack defaults < inventory vars < secrets (highest priority)
func MergeVariables(stackVars, inventoryVars, secrets map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Start with stack defaults (lowest priority)
	for k, v := range stackVars {
		merged[k] = v
	}

	// Override with inventory vars
	for k, v := range inventoryVars {
		merged[k] = v
	}

	// Override with secrets (highest priority)
	for k, v := range secrets {
		merged[k] = v
	}

	return merged
}

// MergeWithCategoryDefaults applies category-level defaults before stack defaults
func MergeWithCategoryDefaults(stackName string, stackVars, inventoryVars, secrets map[string]interface{}) (map[string]interface{}, error) {
	// Load stack to get category
	stack, err := LoadStack(stackName)
	if err != nil {
		return nil, err
	}

	cat, err := categories.Get(stack.Category)
	if err != nil {
		return nil, err
	}

	merged := make(map[string]interface{})

	// Start with category defaults (lowest priority)
	for k, v := range cat.Defaults {
		merged[k] = v
	}

	// Apply stack defaults
	for k, v := range stackVars {
		merged[k] = v
	}

	// Apply inventory vars
	for k, v := range inventoryVars {
		merged[k] = v
	}

	// Apply secrets (highest priority)
	for k, v := range secrets {
		merged[k] = v
	}

	return merged, nil
}

// EnabledStacksMap converts a list of enabled stacks to a map for quick lookup
func EnabledStacksMap(stacks []string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range stacks {
		m[s] = true
	}
	return m
}
