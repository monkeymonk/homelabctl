# Configuration Reference

Complete reference for homelabctl configuration files and options.

## Repository Structure

```
homelab/
├── stacks/              # PUBLIC - Reusable stack definitions
│   └── mystack/
│       ├── stack.yaml           # Stack manifest
│       ├── compose.yml.tmpl     # Docker Compose template
│       ├── config/              # Optional config templates
│       │   └── app.conf.tmpl
│       └── contribute/          # Optional cross-stack contributions
│           └── traefik/
│               └── routes.yml.tmpl
├── enabled/             # INVENTORY - Symlinks to enabled stacks
│   └── mystack -> ../stacks/mystack
├── inventory/           # PRIVATE - Your configuration
│   └── vars.yaml
├── secrets/             # PRIVATE - Encrypted secrets
│   ├── mystack.enc.yaml
│   └── .sops.yaml
└── runtime/             # GENERATED - Never commit
    └── docker-compose.yml
```

## stack.yaml

Stack manifest defining metadata, dependencies, and default configuration.

### Schema

```yaml
name: string              # Stack identifier (REQUIRED)
category: string          # Deployment category (REQUIRED)
requires: []string        # Stack dependencies (optional)
services: []string        # List of all services (REQUIRED)
vars: map                 # Default variables (optional)
persistence:              # Data persistence (optional)
  volumes: []string       # Docker volumes
  paths: []string         # Host paths
```

### Full Example

```yaml
name: jellyfin
category: media
requires:
  - traefik
  - authentik

services:
  - jellyfin
  - jellyfin-web

vars:
  jellyfin:
    # Image configuration
    image: jellyfin/jellyfin:latest
    tag: latest

    # Network configuration
    host_port: 8096
    expose_ports: false  # Use Traefik instead

    # Media paths
    media_path: /mnt/media
    config_path: /mnt/config/jellyfin

    # Transcoding
    enable_gpu: false
    gpu_vendor: nvidia  # or "amd", "intel"

    # Features
    enable_dlna: true
    enable_discovery: true

  jellyfin-web:
    image: jellyfin/jellyfin-web:latest
    port: 8080

persistence:
  volumes:
    - jellyfin_config
    - jellyfin_cache
  paths:
    - /mnt/media  # Mounted from host
    - /mnt/config/jellyfin
```

### Field Descriptions

**name** (required)
- Must match directory name in `stacks/`
- Lowercase, hyphen-separated
- Used as stack identifier throughout system

**category** (required)
- Controls deployment order
- Built-in: `core`, `infrastructure`, `monitoring`, `automation`, `media`, `tools`
- Custom categories automatically discovered

**requires** (optional)
- List of stack names that must be enabled
- Dependencies must form DAG (no cycles)
- Category-aware (can't depend on higher-order categories)

**services** (required)
- Explicit list of all service names
- Used for service-level control
- Must match services in `compose.yml.tmpl`

**vars** (optional)
- Default configuration values
- Lowest priority (overridden by inventory and secrets)
- Nested structure recommended

**persistence** (optional)
- Documents volumes and paths
- Not enforced, purely documentation

## inventory/vars.yaml

Global configuration overriding stack defaults.

### Structure

```yaml
# Disabled services list
disabled_services:
  - service1
  - service2

# Global variables
vars:
  # Common configuration
  domain: example.com
  email: admin@example.com
  timezone: America/New_York

  # Paths
  data_root: /mnt/storage/data
  config_root: /mnt/storage/config
  media_root: /mnt/media

  # User/Group IDs
  puid: 1000
  pgid: 1000

  # Stack-specific overrides
  jellyfin:
    host_port: 8920  # Override default 8096
    enable_gpu: true
    media_path: /mnt/media/jellyfin

  traefik:
    acme_email: letsencrypt@example.com
    cloudflare_api_token: cf_token_here
```

### Variable Precedence

```
Stack vars (stack.yaml) < Inventory vars (inventory/vars.yaml) < Secrets (secrets/*.enc.yaml)
```

**Example:**

```yaml
# stacks/app/stack.yaml
vars:
  app:
    port: 8080  # Default

# inventory/vars.yaml
vars:
  app:
    port: 9000  # Overrides stack default

# secrets/app.enc.yaml
vars:
  app:
    port: 9999  # Overrides everything
```

Result: `port = 9999`

## secrets/<stack>.enc.yaml

Encrypted secrets using SOPS.

### Structure

```yaml
vars:
  mystack:
    # Database credentials
    db_password: "secret-password"
    db_user: "admin"

    # API keys
    api_key: "sk_live_..."
    secret_key: "..."

    # OAuth
    oauth_client_id: "..."
    oauth_client_secret: "..."
```

### SOPS Configuration

`.sops.yaml` in repository root:

```yaml
creation_rules:
  # Encrypt all .enc.yaml files in secrets/
  - path_regex: secrets/.*\.enc\.yaml$
    age: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p

  # Different key for production
  - path_regex: secrets/prod-.*\.enc\.yaml$
    age: age1productionkey...

  # Multiple recipients
  - path_regex: secrets/shared-.*\.enc\.yaml$
    age: >-
      age1person1...,
      age1person2...,
      age1person3...
```

### Working with Secrets

```bash
# Create/edit encrypted file
sops secrets/mystack.enc.yaml

# View decrypted content
sops -d secrets/mystack.enc.yaml

# Rotate encryption keys
sops updatekeys secrets/*.enc.yaml
```

## compose.yml.tmpl

Docker Compose template with gomplate syntax.

### Template Context

Available in all templates:

```yaml
.vars:     # Merged variables (stack < inventory < secrets)
.stack:    # Stack metadata (name, category)
.stacks:   # Global info (enabled: [list])
```

### Example Template

```yaml
services:
  {{ .stack.name }}:
    image: {{ .vars.myapp.image | default "myapp:latest" }}
    container_name: {{ .stack.name }}

    {{ if has "traefik" .stacks.enabled }}
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.{{ .stack.name }}.rule=Host(`{{ .stack.name }}.{{ .vars.domain }}`)"
    {{ else }}
    ports:
      - "{{ .vars.myapp.port }}:8080"
    {{ end }}

    environment:
      - TZ={{ .vars.timezone | default "UTC" }}
      - DEBUG={{ .vars.myapp.debug | default "false" }}

    volumes:
      - {{ .vars.data_root }}/{{ .stack.name }}:/data
      - {{ .vars.config_root }}/{{ .stack.name }}:/config

    restart: unless-stopped

volumes:
  {{ .stack.name }}_data:
```

See [Template Context Reference](template-context.md) for details.

## Category Configuration

Categories are dynamically discovered from stack definitions. No configuration file needed.

### Built-in Categories

Defined in `internal/categories/categories.go`:

```go
defaultMetadata = map[string]Metadata{
    "core": {
        Name:        "core",
        DisplayName: "Core Services",
        Order:       1,
        Color:       "blue",
        Defaults: map[string]interface{}{
            "security_opt": []string{"no-new-privileges:true"},
        },
    },
    "infrastructure": {
        Name:        "infrastructure",
        DisplayName: "Infrastructure",
        Order:       2,
        Color:       "cyan",
        Defaults: map[string]interface{}{
            "security_opt": []string{"no-new-privileges:true"},
        },
    },
    "monitoring": {
        Name:        "monitoring",
        DisplayName: "Monitoring",
        Order:       3,
        Color:       "green",
    },
    "automation": {
        Name:        "automation",
        DisplayName: "Automation",
        Order:       4,
        Color:       "yellow",
    },
    "media": {
        Name:        "media",
        DisplayName: "Media",
        Order:       5,
        Color:       "magenta",
        Defaults: map[string]interface{}{
            "environment": map[string]string{
                "PUID": "1000",
                "PGID": "1000",
            },
        },
    },
    "tools": {
        Name:        "tools",
        DisplayName: "Tools",
        Order:       6,
        Color:       "white",
    },
}
```

### Custom Categories

Just use new category name in `stack.yaml`:

```yaml
name: myapp
category: custom-category  # Automatically discovered
```

To customize metadata, edit `internal/categories/categories.go`.

## Directory Structure Requirements

### Minimal Repository

```
homelab/
├── stacks/
│   └── mystack/
│       ├── stack.yaml
│       └── compose.yml.tmpl
├── enabled/
├── inventory/
│   └── vars.yaml
└── runtime/
```

### Full Repository

```
homelab/
├── stacks/                   # Stack definitions
│   ├── traefik/
│   ├── authentik/
│   └── jellyfin/
├── enabled/                  # Symlinks
│   ├── traefik -> ../stacks/traefik
│   ├── authentik -> ../stacks/authentik
│   └── jellyfin -> ../stacks/jellyfin
├── inventory/                # Configuration
│   └── vars.yaml
├── secrets/                  # Encrypted secrets
│   ├── traefik.enc.yaml
│   ├── authentik.enc.yaml
│   └── .sops.yaml
└── runtime/                  # Generated (gitignored)
    ├── docker-compose.yml
    └── traefik/
        └── dynamic/
```

## Git Configuration

### .gitignore

```gitignore
# Runtime directory (generated files)
runtime/

# Decrypted secrets
secrets/*.yaml
!secrets/*.enc.yaml

# SOPS keys
.age/
*.key

# Temporary files
*.tmp
*.swp
.DS_Store
```

### .gitattributes

```gitattributes
# Treat encrypted files as binary
secrets/*.enc.yaml binary
```

## Environment Variables

homelabctl reads these environment variables:

| Variable | Purpose | Default |
|----------|---------|---------|
| `SOPS_AGE_KEY_FILE` | SOPS age key location | `~/.config/sops/age/keys.txt` |
| `HOMELAB_ROOT` | Repository root override | Current directory |
| `NO_COLOR` | Disable colored output | not set |

### Usage

```bash
# Use different SOPS key
export SOPS_AGE_KEY_FILE=/path/to/key.txt
homelabctl generate

# Disable colors
NO_COLOR=1 homelabctl list

# Override repository root
HOMELAB_ROOT=/path/to/homelab homelabctl deploy
```

## Best Practices

### Stack Configuration

**DO:**
- Provide sensible defaults in `stack.yaml`
- Document all variables with comments
- Use nested variable structure
- List all services explicitly

**DON'T:**
- Hardcode secrets in `stack.yaml`
- Use flat variable structure
- Omit service list

### Inventory Configuration

**DO:**
- Commit `inventory/vars.yaml` to git
- Use for environment-specific config
- Override stack defaults as needed
- Document disabled services

**DON'T:**
- Put secrets in `inventory/vars.yaml`
- Override everything (trust stack defaults)
- Use for stack-specific config (use stack.yaml)

### Secrets

**DO:**
- Encrypt all secrets with SOPS
- Use `.enc.yaml` extension
- Commit encrypted files to git
- Rotate keys periodically
- Test decryption on new systems

**DON'T:**
- Commit decrypted secrets
- Put secrets in inventory or stack.yaml
- Share private keys in version control
- Use weak encryption (GPG key size, Age key generation)

### Templates

**DO:**
- Provide defaults for all variables
- Check stack dependencies with `has`
- Keep logic simple
- Use consistent formatting

**DON'T:**
- Assume variables exist
- Use complex nested logic
- Hardcode values

## See Also

- [Stack Structure](../guide/stack-structure.md) - Repository organization
- [Variables & Templating](../guide/variables.md) - Variable system
- [Secrets Management](../guide/secrets.md) - Secret handling
- [Template Context](template-context.md) - Template context reference
