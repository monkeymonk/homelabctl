package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/monkeymonk/homelabctl/internal/paths"
)

// Compose is a passthrough to docker compose for any command
// This allows access to all docker compose commands while using the correct compose file
func Compose(command string, args []string) error {
	// Check if docker-compose.yml exists
	if _, err := os.Stat(paths.DockerCompose); err != nil {
		return fmt.Errorf("no runtime/docker-compose.yml found - run 'generate' first")
	}

	// Build docker compose command
	cmdArgs := []string{"compose", "-f", paths.DockerCompose, command}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("docker", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin // Allow interactive commands

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose %s failed: %w", command, err)
	}

	return nil
}
