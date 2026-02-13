package errors

import "fmt"

// CommandNotFound creates an error for unknown commands
func CommandNotFound(command string, availableCommands []string) *Error {
	suggestions := []string{
		"Run: homelabctl help",
	}

	context := []string{
		"Available commands:",
	}

	for _, cmd := range availableCommands {
		context = append(context, fmt.Sprintf("  - %s", cmd))
	}

	return New(
		fmt.Sprintf("unknown command: %s", command),
		suggestions...,
	).WithContext(context...)
}

// MissingArgument creates an error for missing required arguments
func MissingArgument(argName, command string) *Error {
	return New(
		fmt.Sprintf("missing required argument: %s", argName),
		fmt.Sprintf("Run: homelabctl %s --help", command),
		fmt.Sprintf("Usage: homelabctl %s <%s>", command, argName),
	)
}

// FileNotFound creates an error for missing files
func FileNotFound(path, purpose string) *Error {
	return New(
		fmt.Sprintf("file not found: %s", path),
		fmt.Sprintf("Purpose: %s", purpose),
		"Check that the file path is correct",
		"Run: homelabctl init (if in a new repository)",
	)
}

// InvalidYAML creates an error for YAML parsing failures
func InvalidYAML(path string, parseError error) *Error {
	return New(
		fmt.Sprintf("invalid YAML in %s", path),
		"Check YAML syntax (indentation, colons, dashes)",
		"Use a YAML validator: https://www.yamllint.com/",
		fmt.Sprintf("Edit: %s", path),
	).WithContext(
		"Parse error:",
		parseError.Error(),
	)
}

// DependencyCycle creates an error for circular dependencies
func DependencyCycle(cycle []string) *Error {
	context := []string{
		"Dependency cycle detected:",
	}

	for i := 0; i < len(cycle); i++ {
		if i == len(cycle)-1 {
			context = append(context, fmt.Sprintf("  %s → %s (cycle!)", cycle[i], cycle[0]))
		} else {
			context = append(context, fmt.Sprintf("  %s → %s", cycle[i], cycle[i+1]))
		}
	}

	suggestions := []string{
		"Remove one of the dependencies to break the cycle",
	}

	// Add specific file suggestions for each stack in the cycle
	for _, stack := range cycle {
		suggestions = append(suggestions, fmt.Sprintf("Edit: stacks/%s/stack.yaml", stack))
	}

	return New(
		"circular dependency detected",
		suggestions...,
	).WithContext(context...)
}
