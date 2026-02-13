# Template Context Reference

Complete reference for the template context available in `compose.yml.tmpl` and contribution templates.

## Context Structure

Every template receives a context object with three top-level keys:

```go
{
  "vars":   map[string]interface{},  // Merged variables
  "stack":  StackMetadata,            // Current stack info
  "stacks": GlobalInfo,               // All stacks info
}
```

## `.vars` - Merged Variables

Merged configuration from all sources with precedence order:

```
Stack defaults (stack.yaml) < Inventory (inventory/vars.yaml) < Secrets (secrets/<stack>.yaml)
```

### Structure

```yaml
vars:
  domain: "example.com"            # Global variable
  timezone: "UTC"                  # Global variable

  mystack:                         # Stack-specific variables
    image: "myapp:latest"
    port: 8080
    features:
      - auth
      - metrics

  another-stack:                   # Another stack's variables
    # ... stack configuration
```

### Access Pattern

```yaml
# Global variables (top-level in vars)
domain: {{ .vars.domain }}

# Stack-specific variables (nested under stack name)
image: {{ .vars.mystack.image }}
port: {{ .vars.mystack.port }}

# Nested properties
{{ range .vars.mystack.features }}
- {{ . }}
{{ end }}

# With defaults
image: {{ .vars.mystack.image | default "nginx:latest" }}
```

### Common Global Variables

These are typically defined in `inventory/vars.yaml`:

```yaml
vars:
  # Domain configuration
  domain: example.com
  email: admin@example.com

  # Timezone
  timezone: America/New_York

  # User/Group IDs (for media stacks)
  puid: 1000
  pgid: 1000

  # Paths
  data_root: /mnt/data
  config_root: /mnt/config

  # Disabled services
  disabled_services:
    - service1
    - service2
```

## `.stack` - Stack Metadata

Information about the current stack being rendered.

### Structure

```go
{
  "name":     string,  // Stack name (e.g., "traefik")
  "category": string,  // Stack category (e.g., "infrastructure")
}
```

### Access Pattern

```yaml
# Stack name
container_name: {{ .stack.name }}
labels:
  - "stack={{ .stack.name }}"

# Stack category
labels:
  - "category={{ .stack.category }}"

# Conditional based on stack
{{ if eq .stack.name "traefik" }}
  # Traefik-specific config
{{ end }}
```

### Example

```yaml
services:
  {{ .stack.name }}:
    container_name: {{ .stack.name }}
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.{{ .stack.name }}.rule=Host(`{{ .stack.name }}.{{ .vars.domain }}`)"
    logging:
      driver: json-file
      options:
        tag: "{{ .stack.category }}/{{ .stack.name }}"
```

## `.stacks` - Global Information

Information about all stacks in the deployment.

### Structure

```go
{
  "enabled": []string,  // List of enabled stack names
}
```

### Access Pattern

```yaml
# Check if specific stack is enabled
{{ if has "traefik" .stacks.enabled }}
labels:
  - "traefik.enable=true"
{{ end }}

# Check multiple stacks
{{ if and (has "traefik" .stacks.enabled) (has "authentik" .stacks.enabled) }}
environment:
  - OAUTH_ENABLED=true
{{ end }}

# Iterate enabled stacks
{{ range .stacks.enabled }}
  # {{ . }} is stack name
{{ end }}
```

### Common Patterns

**Conditional Traefik Integration:**

```yaml
{{ if has "traefik" .stacks.enabled }}
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.{{ .stack.name }}.rule=Host(`{{ .stack.name }}.{{ .vars.domain }}`)"
  - "traefik.http.routers.{{ .stack.name }}.entrypoints=websecure"
  - "traefik.http.routers.{{ .stack.name }}.tls.certresolver=letsencrypt"
{{ else }}
ports:
  - "{{ .vars.myapp.port }}:8080"
{{ end }}
```

**Conditional OAuth:**

```yaml
{{ if has "authentik" .stacks.enabled }}
environment:
  - OAUTH_ENABLED=true
  - OAUTH_ISSUER=https://auth.{{ .vars.domain }}
{{ end }}
```

**Dependency Check:**

```yaml
services:
  myapp:
    {{ if not (has "postgres" .stacks.enabled) }}
    # Use SQLite if postgres not enabled
    environment:
      - DATABASE_TYPE=sqlite
    {{ else }}
    # Use PostgreSQL if enabled
    environment:
      - DATABASE_TYPE=postgres
      - DATABASE_HOST=postgres
    {{ end }}
```

## Complete Example

### Template: `stacks/myapp/compose.yml.tmpl`

```yaml
services:
  {{ .stack.name }}:
    image: {{ .vars.myapp.image | default "myapp:latest" }}
    container_name: {{ .stack.name }}

    {{ if has "traefik" .stacks.enabled }}
    # Traefik integration
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.{{ .stack.name }}.rule=Host(`{{ .stack.name }}.{{ .vars.domain }}`)"
      - "traefik.http.routers.{{ .stack.name }}.entrypoints=websecure"
      - "traefik.http.routers.{{ .stack.name }}.tls.certresolver=letsencrypt"
      {{ if has "authentik" .stacks.enabled }}
      - "traefik.http.routers.{{ .stack.name }}.middlewares=authentik@file"
      {{ end }}
    {{ else }}
    # Direct port mapping
    ports:
      - "{{ .vars.myapp.port | default "8080" }}:8080"
    {{ end }}

    environment:
      - TZ={{ .vars.timezone | default "UTC" }}
      - LOG_LEVEL={{ .vars.myapp.log_level | default "info" }}
      {{ if has "authentik" .stacks.enabled }}
      - OAUTH_ENABLED=true
      - OAUTH_ISSUER=https://auth.{{ .vars.domain }}
      {{ end }}

    volumes:
      - {{ .vars.data_root | default "/mnt/data" }}/{{ .stack.name }}:/data
      - {{ .vars.config_root | default "/mnt/config" }}/{{ .stack.name }}:/config

    restart: unless-stopped

    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
        tag: "{{ .stack.category }}/{{ .stack.name }}"

volumes:
  {{ .stack.name }}_data:
```

### Context Data

```yaml
vars:
  domain: example.com
  timezone: America/New_York
  data_root: /mnt/storage/data
  config_root: /mnt/storage/config
  myapp:
    image: myapp:v1.2.3
    port: 8080
    log_level: debug

stack:
  name: myapp
  category: tools

stacks:
  enabled:
    - traefik
    - authentik
    - myapp
```

### Generated Output

```yaml
services:
  myapp:
    image: myapp:v1.2.3
    container_name: myapp

    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.myapp.rule=Host(`myapp.example.com`)"
      - "traefik.http.routers.myapp.entrypoints=websecure"
      - "traefik.http.routers.myapp.tls.certresolver=letsencrypt"
      - "traefik.http.routers.myapp.middlewares=authentik@file"

    environment:
      - TZ=America/New_York
      - LOG_LEVEL=debug
      - OAUTH_ENABLED=true
      - OAUTH_ISSUER=https://auth.example.com

    volumes:
      - /mnt/storage/data/myapp:/data
      - /mnt/storage/config/myapp:/config

    restart: unless-stopped

    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
        tag: "tools/myapp"

volumes:
  myapp_data:
```

## Variable Precedence Example

Given:

```yaml
# stacks/myapp/stack.yaml
vars:
  myapp:
    image: myapp:latest
    port: 8080
    debug: false

# inventory/vars.yaml
vars:
  myapp:
    port: 9000
    replicas: 2

# secrets/myapp.enc.yaml
vars:
  myapp:
    debug: true
    api_key: secret123
```

Context contains:

```yaml
vars:
  myapp:
    image: myapp:latest      # From stack.yaml
    port: 9000               # From inventory (overrides stack)
    debug: true              # From secrets (overrides stack)
    replicas: 2              # From inventory
    api_key: secret123       # From secrets
```

## Best Practices

### 1. Always Provide Defaults

```yaml
# Good
image: {{ .vars.myapp.image | default "nginx:latest" }}
port: {{ .vars.myapp.port | default "8080" }}

# Bad - fails if not defined
image: {{ .vars.myapp.image }}
```

### 2. Use Consistent Variable Naming

```yaml
# Good - hierarchical structure
vars:
  myapp:
    server:
      port: 8080
    database:
      host: postgres

# Bad - flat structure
vars:
  myapp_server_port: 8080
  myapp_database_host: postgres
```

### 3. Check Stack Dependencies

```yaml
# Good - verify dependency exists
{{ if has "postgres" .stacks.enabled }}
depends_on:
  - postgres
{{ end }}

# Bad - assume dependency exists
depends_on:
  - postgres  # Fails if not enabled
```

### 4. Document Variables

```yaml
# In stack.yaml - comment variable purpose
vars:
  myapp:
    # Container image tag (default: latest)
    image: myapp:latest

    # HTTP port (1024-65535)
    port: 8080

    # Enable debug logging (true/false)
    debug: false
```

## See Also

- [Variables & Templating](../guide/variables.md) - Variable system overview
- [Stack Structure](../guide/stack-structure.md) - Stack organization
- [Gomplate Documentation](https://docs.gomplate.ca/) - Template function reference
