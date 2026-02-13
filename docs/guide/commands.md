# Commands Reference

Complete reference for all homelabctl commands.

## Setup Commands

### `init`

Initialize a new homelab repository or verify an existing one.

```bash
homelabctl init
```

**What it does:**

- Creates required directories (`stacks/`, `enabled/`, `inventory/`, `secrets/`, `runtime/`)
- Creates starter `.gitignore`
- Creates template `inventory/vars.yaml`

**When to use:** First time setup or structure verification.

---

### `enable`

Enable a stack or re-enable a disabled service.

=== "Enable Stack"

    ```bash
    homelabctl enable <stack>
    ```

    Creates a symlink in `enabled/` pointing to `stacks/<stack>`.

=== "Enable Service"

    ```bash
    homelabctl enable -s <service>
    homelabctl enable --service <service>
    ```

    Removes service from disabled list in `inventory/vars.yaml`.

**Examples:**

```bash
# Enable traefik stack
homelabctl enable traefik

# Re-enable a previously disabled service
homelabctl enable -s scrutiny
```

---

### `disable`

Disable a stack or an individual service.

=== "Disable Stack"

    ```bash
    homelabctl disable <stack>
    ```

    Removes the symlink from `enabled/`.

    !!! warning
        Does not check if other stacks depend on this one.

=== "Disable Service"

    ```bash
    homelabctl disable -s <service>
    homelabctl disable --service <service>
    ```

    Adds service to disabled list without disabling the entire stack.

**Use cases for disabling services:**

- Save resources on constrained systems
- Skip services that don't work on your hardware
- Temporarily disable a service you don't need

---

### `list`

List all enabled stacks and disabled services.

```bash
homelabctl list
```

**Output example:**

```
Enabled stacks (3):
  • core
  • traefik
  • monitoring

Disabled services (1):
  • scrutiny
```

---

### `validate`

Validate your homelab configuration.

```bash
homelabctl validate
```

**Checks:**

- ✅ Repository structure is valid
- ✅ All enabled stacks have `stack.yaml`
- ✅ All enabled stacks have `compose.yml.tmpl`
- ✅ All dependencies are satisfied
- ✅ No circular dependencies
- ✅ Service definitions match compose templates

---

## Deployment Commands

### `generate`

Generate runtime Docker Compose files.

```bash
homelabctl generate [--debug]
```

**Flags:**

- `--debug` - Preserve temporary files for inspection

**What it does:**

1. Load enabled stacks
2. Merge variables (stack defaults < inventory < secrets)
3. Render templates with gomplate
4. Filter disabled services
5. Merge all compose files into `runtime/docker-compose.yml`

**Output:** `runtime/docker-compose.yml`

---

### `deploy`

Generate and deploy stacks.

```bash
homelabctl deploy
```

Equivalent to:

```bash
homelabctl generate
docker compose -f runtime/docker-compose.yml up -d
```

**When to use:** After enabling/disabling stacks or changing configuration.

---

## Operations Commands

### `ps`

Show service status.

```bash
homelabctl ps
```

Shows all running services with their status, ports, and container IDs.

---

### `logs`

View service logs.

```bash
homelabctl logs [service...] [flags]
```

**Examples:**

```bash
# Follow all logs
homelabctl logs

# Follow specific service
homelabctl logs traefik

# Last 100 lines
homelabctl logs -n 100

# Last hour
homelabctl logs --since 1h

# Multiple services
homelabctl logs traefik authentik
```

---

### `restart`

Restart services.

```bash
homelabctl restart [service...]
```

**Examples:**

```bash
# Restart all services
homelabctl restart

# Restart specific service
homelabctl restart traefik
```

---

### `stop`

Stop services (keeps containers).

```bash
homelabctl stop [service...]
```

---

### `down`

Stop and remove containers.

```bash
homelabctl down [--volumes]
```

**Flags:**

- `--volumes` - Also remove volumes (⚠️ data loss!)

---

### `exec`

Execute command in a container.

```bash
homelabctl exec <service> <command> [args...]
```

**Examples:**

```bash
# Open shell
homelabctl exec traefik sh

# Database access
homelabctl exec postgres psql -U postgres

# Run CLI tool
homelabctl exec authentik ak --version
```

---

## Passthrough Commands

Any unrecognized command is passed to `docker compose` with the correct file:

```bash
# Pull latest images
homelabctl pull

# Validate compose file
homelabctl config

# Show running processes
homelabctl top

# List images
homelabctl images

# Check port mapping
homelabctl port traefik 80

# Stream events
homelabctl events

# Pause/unpause
homelabctl pause jellyfin
homelabctl unpause jellyfin
```

All docker compose commands work!

---

## Debugging

### Preserve Temporary Files

```bash
homelabctl generate --debug
```

Keeps files in `runtime/`:

```
runtime/
├── traefik-compose.yml      # Per-stack compose files
├── authentik-compose.yml
└── docker-compose.yml        # Final merged file
```

### Inspect Generated Compose

```bash
homelabctl config
```

Shows the final merged configuration that will be deployed.
