package compose

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/monkeymonk/homelabctl/internal/paths"
)

// ComposeFile represents a docker-compose.yml structure
type ComposeFile struct {
	Services map[string]interface{} `yaml:"services,omitempty"`
	Volumes  map[string]interface{} `yaml:"volumes,omitempty"`
	Networks map[string]interface{} `yaml:"networks,omitempty"`
}

// MergeComposeFiles merges multiple rendered compose files into one
func MergeComposeFiles(files []string) (*ComposeFile, error) {
	merged := &ComposeFile{
		Services: make(map[string]interface{}),
		Volumes:  make(map[string]interface{}),
		Networks: make(map[string]interface{}),
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", file, err)
		}

		var compose ComposeFile
		if err := yaml.Unmarshal(data, &compose); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", file, err)
		}

		// Merge services
		for name, svc := range compose.Services {
			if _, exists := merged.Services[name]; exists {
				return nil, fmt.Errorf("duplicate service name: %s", name)
			}
			merged.Services[name] = svc
		}

		// Merge volumes
		for name, vol := range compose.Volumes {
			if existing, exists := merged.Volumes[name]; exists {
				// Warn about duplicate volume definitions
				fmt.Fprintf(os.Stderr, "WARNING: Duplicate volume '%s' in %s (using first definition)\n", name, file)

				// Quick check if they might differ (pointer comparison is cheap)
				if fmt.Sprintf("%p", existing) != fmt.Sprintf("%p", vol) {
					// Check if definitions differ
					existingYAML, _ := yaml.Marshal(existing)
					newYAML, _ := yaml.Marshal(vol)
					if string(existingYAML) != string(newYAML) {
						fmt.Fprintf(os.Stderr, "WARNING: Volume '%s' has conflicting definitions:\n  First: %s\n  Ignored: %s\n",
							name, string(existingYAML), string(newYAML))
					}
				}
				continue
			}
			merged.Volumes[name] = vol
		}

		// Merge networks
		// Prefer non-external definitions over external ones
		for name, net := range compose.Networks {
			if existing, exists := merged.Networks[name]; exists {
				// Check if new network is external
				newIsExternal := false
				if netMap, ok := net.(map[string]interface{}); ok {
					if extVal, hasExt := netMap["external"]; hasExt {
						if ext, ok := extVal.(bool); ok {
							newIsExternal = ext
						}
					}
				}

				// Check if existing network is external
				existingIsExternal := false
				if existingMap, ok := existing.(map[string]interface{}); ok {
					if extVal, hasExt := existingMap["external"]; hasExt {
						if ext, ok := extVal.(bool); ok {
							existingIsExternal = ext
						}
					}
				}

				// Handle different cases
				if newIsExternal && !existingIsExternal {
					// New is external, existing creates it - keep existing (expected)
					continue
				} else if !newIsExternal && existingIsExternal {
					// New creates it, existing is external - replace with new (expected)
					merged.Networks[name] = net
					continue
				} else if !newIsExternal && !existingIsExternal {
					// Both trying to create - this is a REAL conflict
					fmt.Fprintf(os.Stderr, "WARNING: Duplicate network '%s' in %s\n", name, file)
					fmt.Fprintf(os.Stderr, "  → Multiple stacks trying to create the same network\n")
					fmt.Fprintf(os.Stderr, "  → Keeping first definition\n")
					continue
				}
				// Both are external - silently keep first (expected)
				continue
			}
			merged.Networks[name] = net
		}
	}

	return merged, nil
}

// WriteComposeFile writes a ComposeFile to disk as YAML
func WriteComposeFile(path string, compose *ComposeFile) error {
	data, err := yaml.Marshal(compose)
	if err != nil {
		return fmt.Errorf("failed to marshal compose file: %w", err)
	}

	if err := os.WriteFile(path, data, paths.FilePermissions); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	return nil
}

// FilterDisabledServices removes disabled services from a ComposeFile
func FilterDisabledServices(compose *ComposeFile, disabledServices []string) []string {
	if len(disabledServices) == 0 {
		return nil
	}

	// Build lookup map
	disabled := make(map[string]bool)
	for _, svc := range disabledServices {
		disabled[svc] = true
	}

	var removed []string
	for name := range compose.Services {
		if disabled[name] {
			delete(compose.Services, name)
			removed = append(removed, name)
		}
	}

	return removed
}
