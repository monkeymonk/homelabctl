package inventory

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"homelabctl/internal/paths"
)

// LoadVars loads inventory/vars.yaml
func LoadVars() (map[string]interface{}, error) {
	data, err := os.ReadFile(paths.InventoryVars)
	if err != nil {
		return nil, fmt.Errorf("failed to read inventory/vars.yaml: %w", err)
	}

	var vars map[string]interface{}
	if err := yaml.Unmarshal(data, &vars); err != nil {
		return nil, fmt.Errorf("failed to parse inventory/vars.yaml: %w", err)
	}

	if vars == nil {
		vars = make(map[string]interface{})
	}

	return vars, nil
}

// State represents the tool-managed state
type State struct {
	DisabledServices []string `yaml:"disabled_services"`
}

// LoadState loads inventory/state.yaml
func LoadState() (*State, error) {
	data, err := os.ReadFile(paths.InventoryState)
	if err != nil {
		// If state file doesn't exist, create it
		if os.IsNotExist(err) {
			state := &State{DisabledServices: []string{}}
			if err := writeState(state); err != nil {
				return nil, err
			}
			return state, nil
		}
		return nil, fmt.Errorf("failed to read inventory/state.yaml: %w", err)
	}

	var state State
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse inventory/state.yaml: %w", err)
	}

	if state.DisabledServices == nil {
		state.DisabledServices = []string{}
	}

	return &state, nil
}

// writeState writes the state to inventory/state.yaml
func writeState(state *State) error {
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Use secure permissions (0600) for state file as it may contain sensitive service info
	if err := os.WriteFile(paths.InventoryState, data, paths.SecureFilePermissions); err != nil {
		return fmt.Errorf("failed to write inventory/state.yaml: %w", err)
	}

	return nil
}

// GetDisabledServices returns the list of disabled services from state
func GetDisabledServices() ([]string, error) {
	state, err := LoadState()
	if err != nil {
		return nil, err
	}
	return state.DisabledServices, nil
}

// DisableService adds a service to the disabled_services list in state
func DisableService(serviceName string) error {
	state, err := LoadState()
	if err != nil {
		return err
	}

	// Check if already disabled
	for _, s := range state.DisabledServices {
		if s == serviceName {
			return fmt.Errorf("service '%s' is already disabled", serviceName)
		}
	}

	// Add the service
	state.DisabledServices = append(state.DisabledServices, serviceName)

	return writeState(state)
}

// EnableService removes a service from the disabled_services list in state
func EnableService(serviceName string) error {
	state, err := LoadState()
	if err != nil {
		return err
	}

	// Find and remove the service
	found := false
	newList := make([]string, 0, len(state.DisabledServices))
	for _, s := range state.DisabledServices {
		if s == serviceName {
			found = true
			continue
		}
		newList = append(newList, s)
	}

	if !found {
		return fmt.Errorf("service '%s' is not disabled", serviceName)
	}

	state.DisabledServices = newList

	return writeState(state)
}

// MigrateDisabledServices moves disabled_services from vars.yaml to state.yaml (one-time migration)
func MigrateDisabledServices() error {
	// Load vars
	vars, err := LoadVars()
	if err != nil {
		return err
	}

	// Check if disabled_services exists in vars
	disabled, exists := vars["disabled_services"]
	if !exists {
		return nil // Nothing to migrate
	}

	// Load state
	state, err := LoadState()
	if err != nil {
		return err
	}

	// Convert and migrate
	if disabledList, ok := disabled.([]interface{}); ok {
		for _, item := range disabledList {
			if s, ok := item.(string); ok {
				// Avoid duplicates
				found := false
				for _, existing := range state.DisabledServices {
					if existing == s {
						found = true
						break
					}
				}
				if !found {
					state.DisabledServices = append(state.DisabledServices, s)
				}
			}
		}
	}

	// Write state
	if err := writeState(state); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Migrated disabled_services from vars.yaml to state.yaml\n")
	fmt.Fprintf(os.Stderr, "You can now manually remove 'disabled_services:' from inventory/vars.yaml\n")

	return nil
}
