# Your First Stack

This guide walks you through creating a complete stack from scratch.

## What is a Stack?

A stack is a collection of related Docker services defined using:

- **`stack.yaml`** - Manifest with metadata and default variables
- **`compose.yml.tmpl`** - Docker Compose template
- **`config/`** (optional) - Configuration file templates
- **`contribute/`** (optional) - Cross-stack contributions

## Example: Traefik Reverse Proxy

Let's create a Traefik stack as a real-world example.

### 1. Create Stack Directory

```bash
mkdir -p stacks/traefik
cd stacks/traefik
```

### 2. Create stack.yaml

```yaml title="stacks/traefik/stack.yaml"
name: traefik
category: core
requires: []

services:
  - traefik

vars:
  traefik:
    image: traefik:v3.0
    dashboard_enabled: true
    acme_email: admin@example.com

persistence:
  volumes:
    - traefik_acme
```

### 3. Create compose.yml.tmpl

```yaml title="stacks/traefik/compose.yml.tmpl"
services:
  traefik:
    image: {{ .vars.traefik.image }}
    container_name: traefik
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"  # Dashboard
    command:
      - "--api.dashboard={{ .vars.traefik.dashboard_enabled }}"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.email={{ .vars.traefik.acme_email }}"
      - "--certificatesresolvers.letsencrypt.acme.storage=/acme/acme.json"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - traefik_acme:/acme
    restart: unless-stopped
    labels:
      - "traefik.enable=true"

volumes:
  traefik_acme:
```

### 4. Override Variables in Inventory

```yaml title="inventory/vars.yaml"
domain: myserver.home
traefik:
  acme_email: me@myserver.home
  dashboard_enabled: false  # Disable in production
```

### 5. Enable and Deploy

```bash
cd ~/homelab
homelabctl enable traefik
homelabctl validate
homelabctl deploy
```

### 6. Verify

```bash
# Check status
homelabctl ps

# View logs
homelabctl logs traefik

# Access dashboard (if enabled)
curl http://localhost:8080/dashboard/
```

## Stack Components Explained

### stack.yaml

```yaml
name: traefik              # Stack identifier
category: core             # Deployment order
requires: []               # Dependencies
services:                  # List of services (must match compose)
  - traefik
vars:                      # Default variables
  traefik:
    image: traefik:v3.0
persistence:               # Data that survives restarts
  volumes:
    - traefik_acme
```

### compose.yml.tmpl

Standard Docker Compose with templating:

```yaml
services:
  service-name:
    image: {{ .vars.service.image }}  # Variable reference
    ports:
      - "{{ .vars.service.port }}:80"
```

**Template Context:**

- `.vars` - Merged variables (stack defaults < inventory < secrets)
- `.stack.name` - Stack name
- `.stack.category` - Stack category
- `.stacks.enabled` - List of enabled stacks

## Best Practices

!!! tip "Keep It Simple"
    Start with minimal configuration. Add complexity only when needed.

!!! tip "Use Sensible Defaults"
    Put common values in `stack.yaml`, let users override in inventory.

!!! tip "Document Your Vars"
    Add comments in `stack.yaml` explaining what each variable does.

!!! warning "Don't Hardcode Secrets"
    Use the secrets system for sensitive data.

## Next Steps

- [Variables & Templating](../guide/variables.md) - Deep dive into templating
- [Secrets Management](../guide/secrets.md) - Handle sensitive data
- [Categories](../guide/categories.md) - Organize your stacks
