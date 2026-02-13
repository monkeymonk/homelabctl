package cmd

import (
	"fmt"

	"homelabctl/internal/errors"
	"homelabctl/internal/fs"
	"homelabctl/internal/stacks"
)

// Validate checks the repository for errors
func Validate() error {
	fmt.Println("Validating homelab configuration...")

	// Verify repository structure
	if err := fs.VerifyRepository(); err != nil {
		return errors.Wrap(
			err,
			"repository structure is invalid",
			"Run: homelabctl init",
			"Check that you're in a homelab repository root",
		)
	}
	fmt.Println("✓ Repository structure valid")

	// Get enabled stacks
	enabled, err := fs.GetEnabledStacks()
	if err != nil {
		return errors.Wrap(
			err,
			"failed to load enabled stacks",
			"Check that enabled/ directory exists",
			"Run: homelabctl list",
		)
	}

	if len(enabled) == 0 {
		return errors.New(
			"no stacks enabled",
			"Run: homelabctl enable <stack>",
			"Example: homelabctl enable core",
		)
	}

	fmt.Printf("Enabled stacks: %d\n", len(enabled))

	// Verify all enabled stacks have stack.yaml
	for _, name := range enabled {
		if _, err := stacks.LoadStack(name); err != nil {
			return errors.Wrap(
				err,
				fmt.Sprintf("invalid stack '%s'", name),
				fmt.Sprintf("Check: stacks/%s/stack.yaml", name),
				fmt.Sprintf("Run: homelabctl disable %s", name),
			)
		}
	}
	fmt.Printf("✓ All %d enabled stacks have valid stack.yaml\n", len(enabled))

	// Verify all enabled stacks have compose.yml.tmpl
	for _, name := range enabled {
		if !stacks.HasComposeTemplate(name) {
			return errors.New(
				fmt.Sprintf("stack '%s' missing compose.yml.tmpl", name),
				fmt.Sprintf("Create: stacks/%s/compose.yml.tmpl", name),
				"See documentation for template format",
			)
		}
	}
	fmt.Println("✓ All enabled stacks have compose.yml.tmpl")

	// Validate dependencies
	if err := stacks.ValidateDependencies(enabled); err != nil {
		return err // Already has enhanced error from stacks package
	}
	fmt.Println("✓ All dependencies satisfied")

	// Validate service definitions
	for _, stackName := range enabled {
		if err := stacks.ValidateServiceDefinitions(stackName); err != nil {
			return errors.Wrap(
				err,
				fmt.Sprintf("invalid service definitions in stack '%s'", stackName),
				fmt.Sprintf("Edit: stacks/%s/stack.yaml", stackName),
				"Ensure all services in 'services:' list have definitions in 'vars:'",
			)
		}
	}
	fmt.Println("✓ All service definitions are valid")

	// Validate category hierarchy
	if err := stacks.ValidateCategoryDependencies(enabled); err != nil {
		return err
	}
	fmt.Println("✓ Category dependencies are valid")

	fmt.Println("\n✓ Validation successful")
	return nil
}
