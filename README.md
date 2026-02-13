# homelabctl

**homelabctl** is a CLI tool for managing Docker stacks in your homelab using declarative, template-based configuration.

Think of it as a compiler for your homelab: it transforms static stack definitions into runtime Docker Compose configurations.

## Features

- ✅ **Declarative Configuration** - Infrastructure as code for your homelab
- ✅ **Template-Based** - Powered by gomplate for flexible templates
- ✅ **Dependency Management** - Automatic validation of stack dependencies
- ✅ **Secrets Support** - Automatic SOPS decryption for encrypted secrets
- ✅ **Category System** - Organize stacks with automatic deployment ordering
- ✅ **Variable Precedence** - Stack defaults < Inventory overrides < Secrets
- ✅ **Service-Level Control** - Disable individual services without disabling stacks
- ✅ **Docker Compose Passthrough** - Full access to all docker compose commands
- ✅ **Fail-Fast** - Clear errors with actionable suggestions

## Quick Start

### Prerequisites

- [gomplate](https://docs.gomplate.ca/installing/) - Template rendering engine
- Docker with Compose plugin v2
- Go 1.21+ (for building from source)

### Installation

```bash
# Build from source
cd homelabctl/
go build -o homelabctl
sudo mv homelabctl /usr/local/bin/

# Or use go install
go install
```

### Create Your First Homelab

```bash
# Initialize repository
mkdir ~/homelab && cd ~/homelab
homelabctl init

# Enable and deploy a stack
homelabctl enable core
homelabctl validate
homelabctl deploy

# Check status
homelabctl ps
```

## Usage

All commands must be run from the **root of your homelab repository**.

### Initialize or Verify Repository

```bash
homelabctl init
```

Creates the required directory structure in a new directory, or verifies an existing repository.

### List Enabled Stacks and Disabled Services

```bash
homelabctl list
```

Shows all currently enabled stacks and any disabled services.

### Enable a Stack

```bash
homelabctl enable <stack>
```

Creates a symlink in `enabled/` and validates dependencies.

### Disable a Stack

```bash
homelabctl disable <stack>
```

Removes the symlink from `enabled/`.

⚠️ **Warning**: Does not check if other stacks depend on this one.

### Disable/Enable Individual Services

```bash
# Disable a service (keeps stack enabled)
homelabctl disable -s <service>
homelabctl disable --service <service>

# Re-enable a service
homelabctl enable -s <service>
homelabctl enable --service <service>
```

**Use cases:**
- Temporarily disable a service you don't need
- Skip services that don't work on your hardware (e.g., Scrutiny without S.M.A.R.T. drives)
- Save resources on constrained systems

### Validate Configuration

```bash
homelabctl validate
```

Checks:
- Repository structure
- All enabled stacks have valid `stack.yaml`
- All enabled stacks have `compose.yml.tmpl`
- All dependencies are satisfied

### Generate Runtime Files

```bash
homelabctl generate

# Debug mode (preserves temporary files)
homelabctl generate --debug
```

Renders all templates and creates `runtime/docker-compose.yml`.

**Generation Algorithm:**

1. Load enabled stacks
2. Parse each `stack.yaml`
3. Load `inventory/vars.yaml`
4. Load `secrets/<stack>.yaml` (if exists)
5. Build gomplate context (merge vars by precedence)
6. Render `compose.yml.tmpl` for each stack
7. Filter out disabled services
8. Merge all compose files into `runtime/docker-compose.yml`
9. Render Traefik contributions into `runtime/traefik/dynamic/`

### Deploy

```bash
homelabctl deploy
```

Runs `generate` then executes:

```bash
docker compose -f runtime/docker-compose.yml up -d
```

### Manage Running Services

```bash
# Show service status
homelabctl ps

# View logs (follows by default)
homelabctl logs                # All services
homelabctl logs traefik        # Specific service
homelabctl logs -n 100         # Last 100 lines

# Restart services
homelabctl restart             # All services
homelabctl restart traefik     # Specific service

# Stop services (keeps containers)
homelabctl stop                # All services
homelabctl stop jellyfin       # Specific service

# Stop and remove containers
homelabctl down                # Preserves volumes
homelabctl down --volumes      # Removes volumes too

# Execute commands in containers
homelabctl exec traefik sh     # Open shell
homelabctl exec postgres psql -U postgres  # Run psql
```

### Docker Compose Passthrough

Any command not recognized by homelabctl is automatically passed to `docker compose` with the correct file:

```bash
# All these work automatically:
homelabctl pull                # Pull latest images
homelabctl config              # Validate and view compose config
homelabctl top                 # Display running processes
homelabctl images              # List images
homelabctl port traefik 80     # Show port mapping
homelabctl events              # Stream container events
homelabctl pause traefik       # Pause a service
homelabctl unpause traefik     # Unpause a service

# Any docker compose command works:
homelabctl <compose-command> [args...]
```

This means you have access to **all** docker compose functionality through homelabctl.

## Repository Structure

```
homelab/
├── stacks/              # PUBLIC - reusable stack definitions
│   └── <stack>/
│       ├── stack.yaml           # Manifest + default variables
│       ├── compose.yml.tmpl     # Docker Compose template
│       ├── config/              # Optional config templates
│       └── contribute/          # Cross-stack contributions
│           └── traefik/
│               └── middleware.yml.tmpl
├── enabled/             # INVENTORY - symlinks to enabled stacks
├── inventory/
│   └── vars.yaml        # PRIVATE - global overrides
├── secrets/             # PRIVATE - encrypted per-stack secrets
│   └── <stack>.enc.yaml
└── runtime/             # GENERATED - never committed
    └── docker-compose.yml
```

## Stack Definition Example

**stacks/traefik/stack.yaml:**
```yaml
name: traefik
category: core
requires: []
services:
  - traefik
vars:
  traefik:
    image: traefik:v3.0
    hostname: traefik
    port: 8080
persistence:
  volumes:
    - traefik_data
```

**stacks/traefik/compose.yml.tmpl:**
```yaml
services:
  traefik:
    image: {{ .vars.traefik.image }}
    container_name: traefik
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - traefik_data:/data
      - /var/run/docker.sock:/var/run/docker.sock:ro
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.dashboard.rule=Host(`{{ .vars.traefik.hostname }}.{{ .vars.domain }}`)"
    restart: unless-stopped

volumes:
  traefik_data:
```

## Variable Precedence

Variables are merged in the following order (lowest to highest priority):

1. **Stack defaults** (`stack.yaml → vars`)
2. **Inventory overrides** (`inventory/vars.yaml`)
3. **Secrets** (`secrets/<stack>.yaml`)

**Example:**

Stack default:
```yaml
# stacks/traefik/stack.yaml
vars:
  domain: example.com
  traefik_version: "2.10"
```

Inventory override:
```yaml
# inventory/vars.yaml
domain: myserver.home
```

Secret override:
```yaml
# secrets/traefik.yaml
cloudflare_api_token: "secret123"
```

Final merged variables:
```yaml
domain: myserver.home              # From inventory
traefik_version: "2.10"            # From stack default
cloudflare_api_token: "secret123"  # From secrets
```

## Template Context

Templates are rendered with:

```yaml
.vars:    # Merged variables from all sources
.stack:   # Stack metadata (name, category)
.stacks:  # Global info (enabled: [list, of, stacks])
```

**Template examples:**

```yaml
# Conditional based on enabled stacks
{{ if has "sablier" .stacks.enabled }}
labels:
  - "traefik.http.services.myapp.loadbalancer.server.port=8080"
{{ end }}

# Use variables
domain: {{ .vars.domain }}
image: traefik:{{ .vars.traefik_version }}

# Stack metadata
container_name: {{ .stack.name }}
```

## Categories

Organize stacks into categories with automatic deployment ordering:

- `core` (order 1) - Essential services (blue, security options)
- `infrastructure` (order 2) - Supporting services (cyan, security options)
- `monitoring` (order 3) - Observability (green)
- `automation` (order 4) - Workflows (yellow)
- `media` (order 5) - Media management (magenta, PUID/PGID env vars)
- `tools` (order 6) - User applications (white)

**Custom categories are automatically discovered** - no code changes needed! Simply set a new category in your `stack.yaml`.

## Cross-Stack Contributions

Stacks can contribute configuration to other stacks without modifying them.

**Example:** Your `jellyfin` stack contributes Traefik routes:

```
stacks/jellyfin/
└── contribute/
    └── traefik/
        └── routes.yml.tmpl
```

**routes.yml.tmpl:**
```yaml
http:
  routers:
    jellyfin:
      rule: "Host(`jellyfin.{{ .vars.domain }}`)"
      service: jellyfin
      entryPoints: [websecure]
      tls:
        certResolver: letsencrypt

  services:
    jellyfin:
      loadBalancer:
        servers:
          - url: "http://jellyfin:8096"
```

This gets rendered to `runtime/traefik/dynamic/jellyfin-routes.yml` automatically.

## Common Workflows

### Daily Management

```bash
# Check what's running
homelabctl ps

# View logs for debugging
homelabctl logs traefik
homelabctl logs --since 30m  # Last 30 minutes

# Restart after config change
homelabctl restart traefik

# Update a stack
homelabctl disable mystack
# Edit stacks/mystack/compose.yml.tmpl
homelabctl enable mystack
homelabctl deploy
```

### Troubleshooting a Service

```bash
# Check status
homelabctl ps

# View logs
homelabctl logs problematic-service -n 100

# Restart it
homelabctl restart problematic-service

# Still broken? Recreate it
homelabctl stop problematic-service
homelabctl deploy  # Will recreate stopped containers

# Need to debug inside the container?
homelabctl exec problematic-service sh

# Check resource usage
homelabctl top

# Pull latest images
homelabctl pull
```

### Maintenance Window

```bash
# Stop everything
homelabctl stop

# Perform system updates, backups, etc.
# ...

# Restart everything
homelabctl deploy
```

### Development vs Production

**Development** (unencrypted secrets):

```bash
# Use plain YAML for quick iteration
echo "api_key: test123" > secrets/mystack.yaml
homelabctl generate
```

**Production** (encrypted secrets):

```bash
# Setup SOPS encryption
age-keygen -o ~/.config/sops/age/keys.txt

# Encrypt with SOPS
sops -e secrets/mystack.yaml > secrets/mystack.enc.yaml
rm secrets/mystack.yaml
homelabctl generate  # Decrypts automatically
```

## Troubleshooting

### "missing required path: stacks"

You're not in the homelab repository root. Run commands from there:

```bash
cd ~/homelab
homelabctl init
```

### "stack X requires Y but it is not enabled"

Enable the dependency first:

```bash
homelabctl enable authentik  # Enable dependency
homelabctl enable mystack    # Then enable your stack
```

### "gomplate not found in PATH"

Install gomplate:

```bash
curl -o /usr/local/bin/gomplate -sSL https://github.com/hairyhenderson/gomplate/releases/download/v3.11.6/gomplate_linux-amd64
chmod +x /usr/local/bin/gomplate
```

### Changes not applied

```bash
# Regenerate and redeploy
homelabctl deploy

# Or force recreate containers
homelabctl down
homelabctl deploy
```

## Architecture

homelabctl acts as a **compiler**:

```
Stack Definitions + Inventory + Secrets
          ↓
   homelabctl generate
          ↓
  runtime/docker-compose.yml
          ↓
    docker compose up
```

**Code Structure:**

```
homelabctl/
├── main.go           # Entry point (simple switch/case)
├── cmd/              # Command implementations (orchestration only)
├── internal/
│   ├── fs/           # Filesystem operations
│   ├── stacks/       # Stack loading and validation
│   ├── categories/   # Category system
│   ├── inventory/    # Inventory vars loading
│   ├── secrets/      # Secrets loading
│   ├── render/       # Gomplate rendering
│   ├── compose/      # Compose file merging
│   ├── pipeline/     # Pipeline pattern for generate
│   └── errors/       # Enhanced error formatting
└── go.mod
```

## Extension Points

### SOPS Support

SOPS decryption is automatically handled for `.enc.yaml` files:

1. Place encrypted secrets in `secrets/<stack>.enc.yaml`
2. Ensure `sops` is installed and in PATH
3. Ensure age key is configured (`~/.config/sops/age/keys.txt`)

Files ending in `.enc.yaml` are automatically decrypted during generation.

### Additional Providers

To support providers beyond Traefik:

1. Add contribute patterns in stacks (e.g., `contribute/nginx/`)
2. Update `cmd/generate.go` to render these
3. Mount output directories in respective services

### Validation Rules

To add custom validation:

1. Extend `cmd/validate.go`
2. Add checks in `internal/stacks/`

## Development

### Building from Source

```bash
git clone https://github.com/monkeymonk/homelabctl.git
cd homelabctl
go build -o homelabctl
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific test
go test ./internal/stacks -run TestValidateDependencies

# Run with verbose output
go test -v ./...

# Run integration tests
go test ./cmd -run TestIntegration
```

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Documentation

- **[INSTALL.md](INSTALL.md)** - Installation guide
- **[GUIDE.md](GUIDE.md)** - Complete usage guide with examples
- **[IMPLEMENTATION.md](IMPLEMENTATION.md)** - Implementation details
- **[CHANGELOG.md](CHANGELOG.md)** - Version history
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - How to contribute

## Philosophy

> "If you can understand what is running by reading `enabled/`, `inventory/vars.yaml`, and the generated files, the system is correct."

homelabctl is designed to be **simple, deterministic, and transparent**. Every decision is visible in your filesystem, every change is reproducible, and nothing happens by magic.

## Error Handling

homelabctl follows a **fail-fast** philosophy:

- All errors are loud and explicit
- No silent recovery
- Exit code 1 on any failure
- Enhanced errors with actionable suggestions

## License

MIT License - See LICENSE file for details.

## Acknowledgments

- [gomplate](https://github.com/hairyhenderson/gomplate) - Template engine
- [Docker](https://www.docker.com/) - Container platform
- [SOPS](https://github.com/getsops/sops) - Secrets management
