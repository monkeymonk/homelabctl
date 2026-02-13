package categories

import (
	"fmt"
	"sort"
	"strings"
)

// Category represents a stack category
type Category struct {
	Name        string
	DisplayName string
	Order       int                    // Deployment order (lower = earlier)
	Color       string                 // Terminal color
	Defaults    map[string]interface{} // Category-wide defaults
}

// defaultMetadata provides default metadata for known categories
// This allows categories to be discovered dynamically while still having sensible defaults
var defaultMetadata = map[string]*Category{
	"core": {
		Name:        "core",
		DisplayName: "Core",
		Order:       1,
		Color:       "blue",
		Defaults: map[string]interface{}{
			"restart": "unless-stopped",
			"security_opt": []string{
				"no-new-privileges:true",
			},
		},
	},
	"infrastructure": {
		Name:        "infrastructure",
		DisplayName: "Infrastructure",
		Order:       2,
		Color:       "cyan",
		Defaults: map[string]interface{}{
			"restart": "unless-stopped",
			"security_opt": []string{
				"no-new-privileges:true",
			},
		},
	},
	"monitoring": {
		Name:        "monitoring",
		DisplayName: "Monitoring",
		Order:       3,
		Color:       "green",
		Defaults: map[string]interface{}{
			"restart": "unless-stopped",
		},
	},
	"automation": {
		Name:        "automation",
		DisplayName: "Automation",
		Order:       4,
		Color:       "yellow",
		Defaults: map[string]interface{}{
			"restart": "unless-stopped",
		},
	},
	"media": {
		Name:        "media",
		DisplayName: "Media",
		Order:       5,
		Color:       "magenta",
		Defaults: map[string]interface{}{
			"restart": "unless-stopped",
			"environment": map[string]string{
				"PUID": "1000",
				"PGID": "1000",
			},
		},
	},
	"tools": {
		Name:        "tools",
		DisplayName: "Tools",
		Order:       6,
		Color:       "white",
		Defaults:    map[string]interface{}{},
	},
}

// discoveredCategories stores categories found by scanning stacks
var discoveredCategories = make(map[string]*Category)

// RegisterCategory adds a category to the registry
// If the category has default metadata, it will be used, otherwise sensible defaults are applied
func RegisterCategory(name string) {
	if _, exists := discoveredCategories[name]; exists {
		return // Already registered
	}

	// Check if we have default metadata for this category
	if meta, ok := defaultMetadata[name]; ok {
		discoveredCategories[name] = meta
	} else {
		// Create category with sensible defaults
		discoveredCategories[name] = &Category{
			Name:        name,
			DisplayName: toDisplayName(name),
			Order:       999, // Unknown categories go last
			Color:       "white",
			Defaults:    map[string]interface{}{},
		}
	}
}

// toDisplayName converts "my-category" to "My Category"
func toDisplayName(name string) string {
	words := strings.Split(name, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// Reset clears all discovered categories (useful for testing)
func Reset() {
	discoveredCategories = make(map[string]*Category)
}

// Get returns a category by name
// The category must have been registered via RegisterCategory first
func Get(name string) (*Category, error) {
	cat, exists := discoveredCategories[name]
	if !exists {
		return nil, fmt.Errorf("unknown category: %s (has it been registered via stack discovery?)", name)
	}
	return cat, nil
}

// GetOrDefault returns a category by name, registering it if not found
func GetOrDefault(name string) *Category {
	if cat, err := Get(name); err == nil {
		return cat
	}
	RegisterCategory(name)
	return discoveredCategories[name]
}

// AllCategories returns all registered categories sorted by order
func AllCategories() []*Category {
	categories := make([]*Category, 0, len(discoveredCategories))
	for _, cat := range discoveredCategories {
		categories = append(categories, cat)
	}

	// Sort by order, then by name
	sort.Slice(categories, func(i, j int) bool {
		if categories[i].Order != categories[j].Order {
			return categories[i].Order < categories[j].Order
		}
		return categories[i].Name < categories[j].Name
	})

	return categories
}

// ValidCategoryName checks if a category name is valid
// With dynamic categories, all non-empty names are valid
func ValidCategoryName(name string) bool {
	return name != ""
}

// GetOrder returns the deployment order for a category
func GetOrder(name string) int {
	if cat := GetOrDefault(name); cat != nil {
		return cat.Order
	}
	return 999 // Fallback
}
