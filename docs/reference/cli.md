# CLI Reference

Complete command-line reference for homelabctl.

## Synopsis

```
homelabctl <command> [arguments] [flags]
```

## Global Flags

None currently. All commands must be run from within a homelab repository.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SOPS_AGE_KEY_FILE` | Path to Age encryption key for SOPS | `~/.config/sops/age/keys.txt` |
| `HOMELAB_ROOT` | Override repository root detection | Current directory |
| `NO_COLOR` | Disable colored output | Not set |

**Examples:**

```bash
# Use different SOPS key
SOPS_AGE_KEY_FILE=/path/to/key.txt homelabctl generate

# Disable colors
NO_COLOR=1 homelabctl list

# Override repository root
HOMELAB_ROOT=/path/to/homelab homelabctl deploy
```

## Commands

### Setup Commands

#### `init`

Initialize a new homelab repository or verify existing structure.

**Syntax:**
```bash
homelabctl init
```

**Behavior:**
- Creates directory structure if missing
- Creates `.gitignore` if missing
- Creates template `inventory/vars.yaml` if missing
- Idempotent (safe to run multiple times)

**Exit codes:**
- `0` - Success
- `1` - Error (permission denied, etc.)

**Example:**
```bash
mkdir ~/homelab
cd ~/homelab
homelabctl init
```

---

#### `enable`

Enable a stack or re-enable a disabled service.

**Syntax:**
```bash
# Enable stack
homelabctl enable <stack>

# Re-enable service
homelabctl enable -s <service>
homelabctl enable --service <service>
```

**Arguments:**
- `<stack>` - Stack name (must exist in `stacks/`)
- `<service>` - Service name (for `-s` flag)

**Flags:**
- `-s, --service` - Enable a previously disabled service

**Behavior:**
- Creates symlink `enabled/<stack> -> ../stacks/<stack>`
- Or removes service from `disabled_services` in `inventory/vars.yaml`

**Exit codes:**
- `0` - Success
- `1` - Stack doesn't exist, or other error

**Examples:**
```bash
# Enable traefik stack
homelabctl enable traefik

# Re-enable a service
homelabctl enable -s scrutiny
```

---

#### `disable`

Disable a stack or an individual service.

**Syntax:**
```bash
# Disable stack
homelabctl disable <stack>

# Disable service
homelabctl disable -s <service>
homelabctl disable --service <service>
```

**Arguments:**
- `<stack>` - Stack name
- `<service>` - Service name (for `-s` flag)

**Flags:**
- `-s, --service` - Disable a single service without disabling the stack

**Behavior:**
- Removes symlink from `enabled/`
- Or adds service to `disabled_services` in `inventory/vars.yaml`

**Exit codes:**
- `0` - Success
- `1` - Stack not enabled, or other error

**Examples:**
```bash
# Disable monitoring stack
homelabctl disable monitoring

# Disable scrutiny service only
homelabctl disable -s scrutiny
```

---

#### `list`

List enabled stacks and disabled services.

**Syntax:**
```bash
homelabctl list
```

**Output:**
```
Enabled stacks (3):
  core/vpn
  infrastructure/traefik
  monitoring/prometheus

Disabled services (2):
  - scrutiny (in monitoring stack)
  - loki (in monitoring stack)
```

**Exit codes:**
- `0` - Success
- `1` - Not in repository, or other error

---

#### `validate`

Validate homelab configuration.

**Syntax:**
```bash
homelabctl validate
```

**Checks:**
- Repository structure
- Stack definitions exist
- Dependencies satisfied
- No circular dependencies
- Category dependencies valid
- Service definitions match templates

**Output:**
```
✓ Repository structure valid
✓ All enabled stacks have stack.yaml
✓ All enabled stacks have compose.yml.tmpl
✓ Dependencies satisfied
✓ No circular dependencies
✓ All validations passed
```

**Exit codes:**
- `0` - All validations passed
- `1` - Validation failed

---

### Deployment Commands

#### `generate`

Generate runtime Docker Compose files.

**Syntax:**
```bash
homelabctl generate [--debug]
```

**Flags:**
- `--debug` - Preserve temporary files for inspection

**Behavior:**
1. Load enabled stacks from `enabled/` symlinks
2. Load `inventory/vars.yaml`
3. For each stack:
   - Load `stack.yaml`
   - Load `secrets/<stack>.enc.yaml` (if exists)
   - Merge variables (stack < inventory < secrets)
   - Render `compose.yml.tmpl` with gomplate
4. Filter disabled services
5. Merge all compose files
6. Write `runtime/docker-compose.yml`
7. Clean up temporary files (unless `--debug`)

**Output:**
```
Generated: runtime/docker-compose.yml
```

**Exit codes:**
- `0` - Success
- `1` - Template error, validation failed, or other error

**Examples:**
```bash
# Normal generation
homelabctl generate

# Debug mode (preserves temp files)
homelabctl generate --debug
```

---

#### `deploy`

Generate and deploy stacks.

**Syntax:**
```bash
homelabctl deploy
```

**Behavior:**
1. Run `homelabctl generate`
2. Run `docker compose -f runtime/docker-compose.yml up -d`

**Exit codes:**
- `0` - Success
- `1` - Generation or deployment failed

**Example:**
```bash
homelabctl deploy
```

Equivalent to:
```bash
homelabctl generate
docker compose -f runtime/docker-compose.yml up -d
```

---

### Operations Commands

#### `ps`

Show service status.

**Syntax:**
```bash
homelabctl ps
```

**Behavior:**
Runs `docker compose -f runtime/docker-compose.yml ps`

**Output:**
```
NAME         IMAGE              STATUS        PORTS
traefik      traefik:latest     Up 2 hours    80->80/tcp, 443->443/tcp
authentik    authentik:latest   Up 2 hours    9000->9000/tcp
```

---

#### `logs`

View service logs.

**Syntax:**
```bash
homelabctl logs [service...] [flags]
```

**Arguments:**
- `[service...]` - Service names (optional, defaults to all)

**Common flags** (passed to docker compose):
- `-f, --follow` - Follow log output
- `-n, --tail <lines>` - Number of lines to show
- `--since <time>` - Show logs since timestamp
- `--until <time>` - Show logs before timestamp
- `-t, --timestamps` - Show timestamps

**Examples:**
```bash
# Follow all logs
homelabctl logs -f

# Last 100 lines from traefik
homelabctl logs traefik -n 100

# Logs from last hour
homelabctl logs --since 1h

# Multiple services
homelabctl logs traefik authentik
```

---

#### `restart`

Restart services.

**Syntax:**
```bash
homelabctl restart [service...]
```

**Arguments:**
- `[service...]` - Service names (optional, defaults to all)

**Examples:**
```bash
# Restart all services
homelabctl restart

# Restart traefik only
homelabctl restart traefik
```

---

#### `stop`

Stop services without removing containers.

**Syntax:**
```bash
homelabctl stop [service...]
```

**Arguments:**
- `[service...]` - Service names (optional, defaults to all)

---

#### `down`

Stop and remove containers.

**Syntax:**
```bash
homelabctl down [flags]
```

**Common flags:**
- `--volumes` - Remove volumes (⚠️ **DATA LOSS**)
- `--remove-orphans` - Remove containers for services not in compose file

**Examples:**
```bash
# Stop and remove containers
homelabctl down

# Also remove volumes (careful!)
homelabctl down --volumes
```

---

#### `exec`

Execute command in a running container.

**Syntax:**
```bash
homelabctl exec <service> <command> [args...]
```

**Arguments:**
- `<service>` - Service name (required)
- `<command>` - Command to execute
- `[args...]` - Command arguments

**Examples:**
```bash
# Open shell
homelabctl exec traefik sh

# Run command
homelabctl exec postgres psql -U postgres

# Check version
homelabctl exec authentik ak --version
```

---

### Docker Compose Passthrough

Any unrecognized command is automatically passed to `docker compose` with the correct file path.

**Syntax:**
```bash
homelabctl <docker-compose-command> [args...]
```

**Common commands:**

```bash
# Validate compose file
homelabctl config

# Pull latest images
homelabctl pull

# Show running processes
homelabctl top

# List images
homelabctl images

# Show port mapping
homelabctl port <service> <port>

# Stream events
homelabctl events

# Pause/unpause
homelabctl pause <service>
homelabctl unpause <service>

# Scale services
homelabctl scale <service>=<replicas>
```

All `docker compose` commands work!

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error (validation failed, file not found, etc.) |

## File Locations

### Repository Structure

```
homelab/
├── stacks/              # Stack definitions
├── enabled/             # Symlinks to enabled stacks
├── inventory/           # Configuration
│   └── vars.yaml
├── secrets/             # Encrypted secrets
│   └── *.enc.yaml
└── runtime/             # Generated files (gitignored)
    └── docker-compose.yml
```

### Configuration Files

- `stacks/<name>/stack.yaml` - Stack manifest
- `stacks/<name>/compose.yml.tmpl` - Compose template
- `inventory/vars.yaml` - Global variables
- `secrets/<name>.enc.yaml` - Encrypted secrets

### Generated Files

- `runtime/docker-compose.yml` - Final compose file
- `runtime/<stack>-compose.yml` - Temporary (debug mode only)

## Command Chaining

Commands can be chained with `&&`:

```bash
# Enable, generate, and deploy
homelabctl enable traefik && homelabctl deploy

# Validate before deploying
homelabctl validate && homelabctl deploy

# Generate in debug mode and inspect
homelabctl generate --debug && cat runtime/docker-compose.yml
```

## Debugging

### Verbose Output

Docker Compose supports verbose output:

```bash
homelabctl --verbose ps
homelabctl config --verbose
```

### Inspect Generated Files

```bash
# Generate with debug mode
homelabctl generate --debug

# Inspect temporary files
ls runtime/
cat runtime/traefik-compose.yml

# Validate final output
homelabctl config
```

### Common Issues

**"Not in a homelab repository"**
```bash
# Run from repository root
cd /path/to/homelab
homelabctl <command>

# Or override root
HOMELAB_ROOT=/path/to/homelab homelabctl <command>
```

**"Stack does not exist"**
```bash
# List available stacks
ls stacks/

# Check enabled stacks
homelabctl list
```

**Template errors**
```bash
# Generate with debug mode
homelabctl generate --debug

# Check template syntax
cat stacks/mystack/compose.yml.tmpl
```

## See Also

- [Commands Guide](../guide/commands.md) - Detailed command guide with examples
- [Architecture](../advanced/architecture.md) - How commands work internally
- [Troubleshooting](../advanced/troubleshooting.md) - Common issues and solutions
