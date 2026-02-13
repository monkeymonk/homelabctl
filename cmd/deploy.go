package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"homelabctl/internal/paths"
)

// Deploy generates runtime files and deploys using docker compose
func Deploy() error {
	// Step 1: Run generate
	if err := Generate(); err != nil {
		return err
	}

	fmt.Println("\nDeploying with docker compose...")

	// Step 2: Run docker compose
	// Check if .env file exists and pass it explicitly
	args := []string{"compose", "-f", paths.DockerCompose}

	// Add --env-file if .env exists in current directory
	if _, err := os.Stat(".env"); err == nil {
		args = append(args, "--env-file", ".env")
	}

	args = append(args, "up", "-d")

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose failed: %w", err)
	}

	fmt.Println("\n✓ Deployment complete")
	return nil
}
