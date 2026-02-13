package fs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/monkeymonk/homelabctl/internal/paths"
)

// VerifyRepository checks that the homelab repository structure is valid
func VerifyRepository() error {
	required := []string{
		paths.Stacks,
		paths.Enabled,
		paths.Inventory,
		paths.InventoryVars,
	}

	for _, path := range required {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("missing required path: %s", path)
		}

		// For directories
		if path == paths.Stacks || path == paths.Enabled || path == paths.Inventory {
			if !info.IsDir() {
				return fmt.Errorf("%s must be a directory", path)
			}
		}
	}

	return nil
}

// GetEnabledStacks returns list of enabled stack names
func GetEnabledStacks() ([]string, error) {
	entries, err := os.ReadDir(paths.Enabled)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", paths.Enabled, err)
	}

	var stacks []string
	for _, entry := range entries {
		// Skip hidden files (., .., .gitkeep, etc.)
		if entry.Name()[0] == '.' {
			continue
		}

		// Verify it's a valid symlink
		linkPath := paths.EnabledStackLink(entry.Name())
		target, err := os.Readlink(linkPath)
		if err != nil {
			return nil, fmt.Errorf("%s/%s is not a valid symlink: %w", paths.Enabled, entry.Name(), err)
		}

		// Verify target exists (resolve relative to enabled/ directory)
		targetPath := filepath.Clean(filepath.Join(filepath.Dir(linkPath), target))
		if _, err := os.Stat(targetPath); err != nil {
			return nil, fmt.Errorf("%s/%s points to non-existent stack: %s", paths.Enabled, entry.Name(), target)
		}

		stacks = append(stacks, entry.Name())
	}

	return stacks, nil
}

// StackExists checks if a stack exists in stacks/
func StackExists(name string) bool {
	stackPath := paths.StackDir(name)
	info, err := os.Stat(stackPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsStackEnabled checks if a stack is enabled
func IsStackEnabled(name string) bool {
	linkPath := paths.EnabledStackLink(name)
	_, err := os.Lstat(linkPath)
	return err == nil
}

// EnableStack creates a symlink in enabled/
func EnableStack(name string) error {
	if !StackExists(name) {
		return fmt.Errorf("stack does not exist: %s", name)
	}

	if IsStackEnabled(name) {
		return fmt.Errorf("stack already enabled: %s", name)
	}

	linkPath := paths.EnabledStackLink(name)
	targetPath := filepath.Join("..", paths.Stacks, name)

	if err := os.Symlink(targetPath, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// DisableStack removes symlink from enabled/
func DisableStack(name string) error {
	if !IsStackEnabled(name) {
		return fmt.Errorf("stack not enabled: %s", name)
	}

	linkPath := paths.EnabledStackLink(name)
	if err := os.Remove(linkPath); err != nil {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}

	return nil
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, paths.DirPermissions)
}

// GetAvailableStacks returns all stacks in the stacks/ directory
func GetAvailableStacks() ([]string, error) {
	entries, err := os.ReadDir(paths.Stacks)
	if err != nil {
		return nil, fmt.Errorf("failed to read stacks directory: %w", err)
	}

	var stacks []string
	for _, entry := range entries {
		if entry.IsDir() {
			stacks = append(stacks, entry.Name())
		}
	}

	return stacks, nil
}

// IsHomelabRepository checks if current directory looks like a homelab repository
func IsHomelabRepository() bool {
	// Check if at least the stacks directory exists
	_, err := os.Stat(paths.Stacks)
	return err == nil
}

// InitializeRepository creates a fresh homelab repository structure
func InitializeRepository() error {
	// Create required directories
	dirs := []string{
		paths.Stacks,
		paths.Enabled,
		paths.Inventory,
		paths.Secrets,
	}

	for _, dir := range dirs {
		if err := EnsureDir(dir); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	// Create inventory/vars.yaml with defaults
	inventoryContent := `# Homelab Inventory Variables
#
# This file contains environment-specific configuration that overrides
# stack defaults. Variables defined here are available to all templates.
#
# Example variables:
# domain: home.example.com
# timezone: America/New_York
# acme_email: admin@home.example.com

# Add your global variables below:
`

	if err := os.WriteFile(paths.InventoryVars, []byte(inventoryContent), paths.FilePermissions); err != nil {
		return fmt.Errorf("failed to create inventory/vars.yaml: %w", err)
	}

	// Create .gitignore
	gitignoreContent := `# Generated runtime files (never commit)
runtime/

# Secrets (never commit unencrypted)
secrets/*.yaml

# Personal inventory (optional - remove these lines to commit your config)
inventory/

# Binary
homelabctl

# OS files
.DS_Store
Thumbs.db
`

	if err := os.WriteFile(".gitignore", []byte(gitignoreContent), paths.FilePermissions); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	// Create README.md
	readmeContent := `# My Homelab

This repository contains my homelab infrastructure managed by [homelabctl](https://github.com/yourusername/homelabctl).

## Getting Started

1. Create stack definitions in ` + "`stacks/`" + `
2. Enable stacks: ` + "`homelabctl enable <stack>`" + `
3. Configure variables in ` + "`inventory/vars.yaml`" + `
4. Deploy: ` + "`homelabctl deploy`" + `

## Repository Structure

- ` + "`stacks/`" + ` - Stack definitions (commit to git)
- ` + "`enabled/`" + ` - Enabled stacks as symlinks (commit to git)
- ` + "`inventory/`" + ` - Environment configuration (private)
- ` + "`secrets/`" + ` - Encrypted secrets (private)
- ` + "`runtime/`" + ` - Generated files (never commit)

## Commands

` + "```bash" + `
homelabctl init          # Initialize/verify repository
homelabctl list          # List enabled stacks
homelabctl enable <name> # Enable a stack
homelabctl validate      # Validate configuration
homelabctl generate      # Generate runtime files
homelabctl deploy        # Deploy with docker compose
` + "```" + `

For more information, see [GUIDE.md](GUIDE.md) in the homelabctl repository.
`

	if err := os.WriteFile("README.md", []byte(readmeContent), paths.FilePermissions); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	return nil
}
