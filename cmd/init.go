package cmd

import (
	"fmt"
	"os"

	"github.com/monkeymonk/homelabctl/internal/fs"
	"github.com/monkeymonk/homelabctl/internal/inventory"
)

// Init initializes a new homelab repository or verifies an existing one
func Init() error {
	// Check if this is already a homelab repository
	if !fs.IsHomelabRepository() {
		fmt.Println("No homelab repository found. Initializing new repository...")

		if err := fs.InitializeRepository(); err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}

		fmt.Println()
		fmt.Println("✓ Repository initialized successfully!")
		fmt.Println()
		fmt.Println("Created structure:")
		fmt.Println("  stacks/           - Place your stack definitions here")
		fmt.Println("  enabled/          - Symlinks to enabled stacks")
		fmt.Println("  inventory/        - Your environment configuration")
		fmt.Println("  secrets/          - Encrypted secrets")
		fmt.Println("  .gitignore        - Protects sensitive files")
		fmt.Println("  README.md         - Getting started guide")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  1. Create stack definitions in stacks/")
		fmt.Println("  2. Enable stacks: homelabctl enable <stack>")
		fmt.Println("  3. Configure: edit inventory/vars.yaml")
		fmt.Println("  4. Deploy: homelabctl deploy")
		fmt.Println()

		return nil
	}

	// Existing repository - verify it
	fmt.Println("Verifying homelab repository structure...")

	if err := fs.VerifyRepository(); err != nil {
		return fmt.Errorf("repository verification failed: %w", err)
	}

	// Migrate disabled services from vars.yaml to state.yaml
	if err := inventory.MigrateDisabledServices(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to migrate disabled services: %v\n", err)
	}

	// Verify enabled symlinks are valid
	enabled, err := fs.GetEnabledStacks()
	if err != nil {
		return fmt.Errorf("failed to read enabled stacks: %w", err)
	}

	fmt.Printf("✓ Repository structure valid\n")
	fmt.Printf("✓ Found %d enabled stack(s)\n", len(enabled))

	return nil
}
