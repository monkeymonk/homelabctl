# homelabctl - Complete Guide

This guide explains how to set up and use **homelabctl** to manage your homelab Docker infrastructure.

## What is homelabctl?

A declarative, template-based CLI tool for managing single-node homelab Docker stacks. Think of it as a lightweight alternative to Kubernetes for home servers - you get reproducibility, security, and version control without the complexity.

**Key Features:**
- ✅ All stack definitions are public-safe (no secrets in git)
- ✅ Single command deployment
- ✅ Declarative configuration (what you write is what runs)
- ✅ Template-based with gomplate
- ✅ Built-in dependency management

## Prerequisites

Install these tools before starting:

```bash
# Go 1.21+ (for building homelabctl)
# Check: go version

# gomplate (for template rendering)
curl -o /usr/local/bin/gomplate -sSL https://github.com/hairyhenderson/gomplate/releases/download/v3.11.6/gomplate_linux-amd64
chmod +x /usr/local/bin/gomplate

# Docker with Compose plugin
# Check: docker compose version

# Optional: SOPS + age (for encrypted secrets)
# age-keygen -o ~/.config/sops/age/keys.txt
```

## Installation

### 1. Build homelabctl

```bash
cd homelabctl/
go build -o homelabctl
sudo mv homelabctl /usr/local/bin/
```

### 2. Create Your Homelab Repository

```bash
mkdir ~/homelab
cd ~/homelab
homelabctl init
```

This creates the required directory structure:

```
homelab/
├── stacks/       # Stack definitions (commit to git)
├── enabled/      # Symlinks to enabled stacks (commit to git)
├── inventory/    # Your configuration (DO NOT commit)
│   └── vars.yaml
├── secrets/      # Encrypted secrets (DO NOT commit if unencrypted)
├── .gitignore    # Protects sensitive files
└── README.md     # Getting started guide
```

The `init` command also creates a starter `.gitignore` and `README.md` for you.

## Repository Structure Explained

### `stacks/` - Stack Definitions (PUBLIC)

Each stack is a directory containing:

```
stacks/traefik/
├── stack.yaml           # Manifest + default variables
├── compose.yml.tmpl     # Docker Compose template
├── config/              # Optional config templates
│   └── traefik.yml.tmpl
└── contribute/          # Cross-stack contributions
    └── traefik/
        └── middleware.yml.tmpl
```

**Example `stack.yaml`:**

```yaml
name: traefik
category: infra

requires: []  # Dependencies

vars:
  domain: example.com
  acme_email: admin@example.com
  traefik_version: "2.10"

persistence:
  volumes:
    - traefik-acme
  paths:
    - ./runtime/traefik
```

**Example `compose.yml.tmpl`:**

```yaml
services:
  traefik:
    image: traefik:{{ .vars.traefik_version }}
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - traefik-acme:/acme
      - ./runtime/traefik:/etc/traefik:ro
    labels:
      - "traefik.enable=true"

volumes:
  traefik-acme:
```

### `enabled/` - Enabled Stacks (COMMIT)

Contains symlinks to enabled stacks:

```bash
enabled/
├── traefik -> ../stacks/traefik
└── authentik -> ../stacks/authentik
```

**This is how you declare what's running.** The filesystem is the source of truth.

### `inventory/vars.yaml` - Your Configuration (PRIVATE)

Override stack defaults with your environment-specific values:

```yaml
# Global variables available to all stacks
domain: myserver.home
acme_email: me@myserver.home
timezone: America/New_York

# Service-specific overrides
authentik_port: 9000
traefik_dashboard_enabled: true
```

### `secrets/` - Encrypted Secrets (PRIVATE)

One encrypted file per stack (optional):

```bash
secrets/
├── authentik.enc.yaml    # Encrypted with SOPS
└── traefik.enc.yaml
```

**Example unencrypted secret (for development):**

```yaml
# secrets/authentik.yaml
authentik_secret_key: "super-secret-key-here"
database_password: "another-secret"
```

**Encrypt with SOPS:**

```bash
sops -e secrets/authentik.yaml > secrets/authentik.enc.yaml
rm secrets/authentik.yaml
```

### `runtime/` - Generated Files (NEVER COMMIT)

All generated files go here - these are recreated on every `generate`:

```
runtime/
├── docker-compose.yml      # Final merged compose file
└── traefik/
    └── dynamic/            # Traefik dynamic configs
        ├── traefik-middleware.yml
        └── authentik-routes.yml
```

Add to `.gitignore`:

```gitignore
runtime/
secrets/*.yaml
secrets/*.enc.yaml
inventory/
```

## Workflow

### 1. Initialize Repository

From your homelab directory:

```bash
homelabctl init
```

If this is a new directory, `init` will create the required structure. If it's an existing repository, it will verify the structure is valid.

### 2. Browse Available Stacks

```bash
ls stacks/
# traefik  authentik  borgmatic  jellyfin  ...
```

### 3. Enable a Stack

```bash
homelabctl enable traefik
```

This:
- Creates symlink in `enabled/traefik -> ../stacks/traefik`
- Validates dependencies are satisfied
- Does NOT deploy yet

### 4. Configure Variables

Edit `inventory/vars.yaml`:

```yaml
domain: home.example.com
acme_email: admin@home.example.com
```

### 5. Add Secrets (if needed)

Create `secrets/traefik.yaml`:

```yaml
cloudflare_api_token: "your-token-here"
```

### 6. Validate Configuration

```bash
homelabctl validate
```

Checks:
- All enabled stacks have valid `stack.yaml`
- All dependencies are satisfied
- All required templates exist

### 7. Generate Runtime Files

```bash
homelabctl generate
```

This renders all templates and creates `runtime/docker-compose.yml`.

**What happens:**
1. Loads all enabled stacks
2. Merges variables (stack defaults < inventory < secrets)
3. Renders each `compose.yml.tmpl` with gomplate
4. Merges all compose files into one
5. Renders cross-stack contributions (e.g., Traefik routes)

### 8. Deploy

```bash
homelabctl deploy
```

Runs `generate` then executes:

```bash
docker compose -f runtime/docker-compose.yml up -d
```

### 9. Inspect What's Running

```bash
homelabctl list
# Enabled stacks:
#   - traefik
#   - authentik

docker compose -f runtime/docker-compose.yml ps
```

### 10. Monitor and Manage Services

```bash
# Check what's running
homelabctl ps

# View logs
homelabctl logs              # Follow all logs
homelabctl logs traefik      # Follow specific service
homelabctl logs -n 50        # Last 50 lines
homelabctl logs --since 1h   # Last hour

# Restart services
homelabctl restart           # Restart all
homelabctl restart authentik # Restart specific service

# Execute commands in containers
homelabctl exec traefik sh                    # Open shell
homelabctl exec postgres psql -U postgres     # Database access
homelabctl exec authentik ak --version        # Run CLI tools
```

### 11. Stop Services

```bash
# Stop services (keeps containers for restart)
homelabctl stop              # Stop all
homelabctl stop jellyfin     # Stop specific service

# Stop and remove containers (preserves volumes)
homelabctl down

# Full cleanup including volumes
homelabctl down --volumes
```

### 12. Disable a Stack

```bash
homelabctl disable jellyfin
homelabctl generate  # Regenerate without jellyfin
homelabctl deploy    # Deploy changes
```

## Variable Precedence

Variables are merged in this order (lowest to highest priority):

```
1. Stack defaults (stacks/mystack/stack.yaml → vars)
   ↓ overridden by
2. Inventory vars (inventory/vars.yaml)
   ↓ overridden by
3. Secrets (secrets/mystack.yaml)
```

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

Templates receive this context:

```yaml
.vars:
  # All merged variables (see precedence above)
  domain: myserver.home
  traefik_version: "2.10"

.stack:
  name: traefik
  category: infra

.stacks:
  enabled: [traefik, authentik, borgmatic]
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

## Cross-Stack Contributions

Stacks can contribute configuration to other stacks without modifying them.

**Example:** Your `jellyfin` stack contributes Traefik routes:

```
stacks/jellyfin/
└── contribute/
    └── traefik/
        └── routes.yml.tmpl
```

**`routes.yml.tmpl`:**

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

## Common Operations

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

### Updating Enabled Stacks

```bash
# Edit templates or configs
vim stacks/traefik/compose.yml.tmpl

# Regenerate and redeploy
homelabctl generate
homelabctl deploy
# Or combined: homelabctl deploy (runs generate automatically)
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

### Docker Compose Passthrough

**Any unknown command is passed to docker compose automatically:**

```bash
homelabctl config              # Validate compose config
homelabctl images              # List images used
homelabctl port traefik 80     # Check port mapping
homelabctl events              # Stream events
homelabctl pause jellyfin      # Pause a service
homelabctl unpause jellyfin    # Unpause it

# Literally any docker compose command works
homelabctl <any-compose-command> [args...]
```

This means **homelabctl is a complete docker compose wrapper** - you never need to type `docker compose -f runtime/docker-compose.yml` again.

### Maintenance Window

```bash
# Stop everything
homelabctl stop

# Perform system updates, backups, etc.
# ...

# Restart everything
homelabctl deploy
```

### Clean Shutdown

```bash
# Stop and remove containers (keeps volumes)
homelabctl down

# Or remove everything including volumes
homelabctl down --volumes  # DANGER: Deletes data!
```

## Common Patterns

### Development vs Production

**Development** (unencrypted secrets):

```bash
# Use plain YAML for quick iteration
echo "api_key: test123" > secrets/mystack.yaml
homelabctl generate
```

**Production** (encrypted secrets):

```bash
# Encrypt with SOPS
sops -e secrets/mystack.yaml > secrets/mystack.enc.yaml
rm secrets/mystack.yaml
homelabctl generate  # Decrypts automatically
```

### Multiple Environments

```bash
homelab/
├── inventory/
│   ├── vars.yaml       # Development
│   ├── vars.prod.yaml  # Production
│   └── vars.test.yaml  # Testing
```

Switch environments:

```bash
cp inventory/vars.prod.yaml inventory/vars.yaml
homelabctl deploy
```

### Debugging Templates

Test gomplate rendering manually:

```bash
# Create test context
cat > /tmp/context.yaml <<EOF
vars:
  domain: test.home
  version: "2.10"
stack:
  name: traefik
stacks:
  enabled: [traefik, authentik]
EOF

# Test template
gomplate -f stacks/traefik/compose.yml.tmpl -c .=/tmp/context.yaml
```

### Backup Strategy

**What to commit:**
- `stacks/` - All stack definitions
- `enabled/` - Symlinks (what's running)
- `.gitignore` - Protect secrets

**What to backup (not commit):**
- `inventory/vars.yaml` - Your config
- `secrets/*.enc.yaml` - Encrypted secrets
- Age key (`~/.config/sops/age/keys.txt`)
- Docker volumes (use borgmatic stack)

**Never commit:**
- `runtime/` - Regenerated files
- `secrets/*.yaml` - Plain text secrets

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

### Template rendering fails

1. Check syntax with gomplate directly
2. Verify variable names match
3. Check `.stacks.enabled` conditionals

### Changes not applied

```bash
# Regenerate and redeploy
homelabctl deploy

# Or force recreate containers
homelabctl down
homelabctl deploy
```

### Check service status

```bash
homelabctl ps                     # Show all services
homelabctl logs <service>         # Debug specific service
homelabctl restart <service>      # Restart problematic service
```

## Example: Complete Setup

Here's a complete example setting up Traefik + Authentik:

```bash
# 1. Initialize
mkdir ~/homelab && cd ~/homelab
homelabctl init

# 2. Create Traefik stack
mkdir -p stacks/traefik
cat > stacks/traefik/stack.yaml <<EOF
name: traefik
category: infra
requires: []
vars:
  traefik_version: "2.10"
  domain: example.com
EOF

cat > stacks/traefik/compose.yml.tmpl <<EOF
services:
  traefik:
    image: traefik:{{ .vars.traefik_version }}
    container_name: traefik
    ports:
      - "80:80"
      - "443:443"
    command:
      - "--api.dashboard=true"
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    restart: unless-stopped
EOF

# 3. Configure inventory (init already created this file)
cat >> inventory/vars.yaml <<EOF
domain: home.example.com
timezone: America/New_York
EOF

# 4. Enable and deploy
homelabctl enable traefik
homelabctl validate
homelabctl deploy

# 5. Verify
docker compose -f runtime/docker-compose.yml ps
```

## Next Steps

1. **Create your stack definitions** - Start with essential services (Traefik, reverse proxy)
2. **Enable incrementally** - Don't enable everything at once
3. **Test in development** - Use plain YAML secrets first
4. **Add encryption** - Set up SOPS + age for production
5. **Version control** - Commit your stack definitions and enabled/ directory
6. **Automate backups** - Use the borgmatic stack

## Resources

- [gomplate documentation](https://docs.gomplate.ca/)
- [Docker Compose specification](https://docs.docker.com/compose/compose-file/)
- [SOPS documentation](https://github.com/mozilla/sops)
- [Traefik documentation](https://doc.traefik.io/traefik/)

## Philosophy

> "If you can understand what is running by reading `enabled/`, `inventory/vars.yaml`, and the generated files, the system is correct."

homelabctl is designed to be **simple, deterministic, and transparent**. Every decision is visible in your filesystem, every change is reproducible, and nothing happens by magic.
