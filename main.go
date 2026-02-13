package main

import (
	"fmt"
	"os"

	"github.com/monkeymonk/homelabctl/cmd"
	"github.com/monkeymonk/homelabctl/internal/errors"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse debug flag
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--debug" {
			os.Setenv("HOMELAB_DEBUG", "1")
			// Remove flag from args
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			break
		}
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error

	switch command {
	case "init":
		err = cmd.Init()
	case "enable":
		err = cmd.Enable(args)
	case "disable":
		err = cmd.Disable(args)
	case "list":
		err = cmd.List()
	case "validate":
		err = cmd.Validate()
	case "generate":
		err = cmd.Generate()
	case "deploy":
		err = cmd.Deploy()
	default:
		// Pass through to docker compose for all other commands
		// This allows ps, logs, restart, stop, down, exec, pull, config, etc.
		err = cmd.Compose(command, args)
	}

	if err != nil {
		// Check if it's our enhanced error type
		if enhancedErr, ok := err.(*errors.Error); ok {
			// Already formatted with suggestions
			fmt.Fprint(os.Stderr, enhancedErr.Error())
		} else {
			// Standard error
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("homelabctl - Homelab Stack Runtime CLI")
	fmt.Println()
	fmt.Println("Setup:")
	fmt.Println("  homelabctl init                            Initialize new repository or verify existing")
	fmt.Println("  homelabctl enable <stack> [--suggest-category]  Enable a stack")
	fmt.Println("  homelabctl enable -s <service>             Re-enable a disabled service")
	fmt.Println("  homelabctl disable <stack>        Disable a stack")
	fmt.Println("  homelabctl disable -s <service>   Disable a service (keeps stack enabled)")
	fmt.Println("  homelabctl list                   List enabled stacks and disabled services")
	fmt.Println("  homelabctl validate               Validate configuration")
	fmt.Println()
	fmt.Println("Deployment:")
	fmt.Println("  homelabctl generate               Generate runtime files")
	fmt.Println("  homelabctl deploy                 Generate and deploy")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --debug                           Enable debug mode (preserve temporary files)")
	fmt.Println()
	fmt.Println("Operations:")
	fmt.Println("  homelabctl ps                     Show service status")
	fmt.Println("  homelabctl logs [service...]      Show logs (default: follow all)")
	fmt.Println("  homelabctl restart [service...]   Restart services (default: all)")
	fmt.Println("  homelabctl stop [service...]      Stop services (default: all)")
	fmt.Println("  homelabctl down [--volumes]       Stop and remove containers")
	fmt.Println("  homelabctl exec <service> <cmd>   Execute command in container")
	fmt.Println()
	fmt.Println("Passthrough:")
	fmt.Println("  Any other command is passed to docker compose with the correct file:")
	fmt.Println("  homelabctl pull             # docker compose pull")
	fmt.Println("  homelabctl config           # docker compose config")
	fmt.Println("  homelabctl top              # docker compose top")
	fmt.Println()
	fmt.Println("Get started:")
	fmt.Println("  mkdir homelab && cd homelab")
	fmt.Println("  homelabctl init")
}
