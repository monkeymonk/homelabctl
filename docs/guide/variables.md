# Variables & Templating

homelabctl uses [gomplate](https://docs.gomplate.ca/) for templating, providing powerful variable substitution and logic in your Docker Compose templates.

## Variable Precedence

Variables are merged from multiple sources with a strict precedence order:

```
Stack defaults (stack.yaml → vars)     [lowest priority]
  ↓ overridden by
Inventory vars (inventory/vars.yaml)
  ↓ overridden by
Secrets (secrets/<stack>.yaml)          [highest priority]
```

**Example:**

```yaml
# stacks/jellyfin/stack.yaml
vars:
  jellyfin:
    image: jellyfin/jellyfin:latest
    port: 8096
    theme: default

# inventory/vars.yaml
vars:
  jellyfin:
    port: 8080  # Overrides 8096

# secrets/jellyfin.yaml
vars:
  jellyfin:
    theme: dark  # Overrides "default"
```

**Result:** `port: 8080`, `theme: dark`, `image: jellyfin/jellyfin:latest`

## Template Context

Every template receives a context with three top-level keys:

```yaml
.vars:     # Merged variables (see precedence above)
.stack:    # Stack metadata (name, category)
.stacks:   # Global info (enabled: [list, of, stacks])
```

### `.vars` - Merged Variables

Access your configuration:

```yaml
services:
  myapp:
    image: {{ .vars.myapp.image }}
    ports:
      - "{{ .vars.myapp.port }}:8080"
    environment:
      - DOMAIN={{ .vars.domain }}
```

### `.stack` - Stack Metadata

Current stack information:

```yaml
# {{ .stack.name }} = "jellyfin"
# {{ .stack.category }} = "media"

services:
  {{ .stack.name }}:
    container_name: {{ .stack.name }}
```

### `.stacks` - Global Information

Check what stacks are enabled:

```yaml
{{ if has "traefik" .stacks.enabled }}
labels:
  - "traefik.enable=true"
{{ end }}

{{ if has "authentik" .stacks.enabled }}
environment:
  - OAUTH_ENABLED=true
{{ end }}
```

## Template Syntax

### Variable Substitution

```yaml
# Simple
image: {{ .vars.myapp.image }}

# With default
image: {{ .vars.myapp.image | default "nginx:latest" }}

# Nested variables
replicas: {{ .vars.myapp.scaling.replicas }}
```

### Conditional Logic

```yaml
{{ if .vars.myapp.ssl_enabled }}
environment:
  - SSL_CERT=/certs/cert.pem
  - SSL_KEY=/certs/key.pem
{{ end }}

{{ if eq .vars.myapp.mode "production" }}
restart: always
{{ else }}
restart: unless-stopped
{{ end }}
```

### Loops

```yaml
{{ range .vars.myapp.domains }}
labels:
  - "traefik.http.routers.myapp.rule=Host(`{{ . }}`)"
{{ end }}
```

### Functions

Gomplate provides many built-in functions:

```yaml
# String manipulation
- HOSTNAME={{ .vars.myapp.name | strings.ToUpper }}

# Lists
{{ if has "feature-x" .vars.myapp.features }}
environment:
  - FEATURE_X=enabled
{{ end }}

# JSON
{{ $config := .vars.myapp.config | data.ToJSON }}
- CONFIG={{ $config }}

# Files (advanced)
{{ $cert := file.Read "/path/to/cert" }}
```

See [gomplate documentation](https://docs.gomplate.ca/functions/) for full function reference.

## Common Patterns

### Environment Variables

```yaml
environment:
  - PUID={{ .vars.media.puid | default "1000" }}
  - PGID={{ .vars.media.pgid | default "1000" }}
  - TZ={{ .vars.timezone | default "UTC" }}
```

### Traefik Labels

```yaml
{{ if has "traefik" .stacks.enabled }}
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.{{ .stack.name }}.rule=Host(`{{ .stack.name }}.{{ .vars.domain }}`)"
  - "traefik.http.routers.{{ .stack.name }}.entrypoints=websecure"
  - "traefik.http.routers.{{ .stack.name }}.tls.certresolver=letsencrypt"
{{ end }}
```

### Port Mapping

```yaml
{{ if .vars.myapp.expose_ports }}
ports:
  - "{{ .vars.myapp.port }}:8080"
{{ end }}
```

### Multi-Service Stacks

```yaml
services:
  app:
    image: {{ .vars.app.image }}
    depends_on:
      - db

  db:
    image: {{ .vars.db.image }}
    environment:
      - POSTGRES_PASSWORD={{ .vars.db.password }}
```

## Debugging Templates

### Generate with Debug Mode

```bash
homelabctl generate --debug
```

This preserves temporary files in `runtime/` for inspection.

### Test Standalone

Extract context to test templates manually:

```bash
# Generate creates context files temporarily
# You can capture and use them:
gomplate -f stacks/mystack/compose.yml.tmpl -c .=/tmp/context.yaml
```

### Common Errors

**Missing variable:**
```
template: :5:10: executing "" at <.vars.missing>: map has no entry for key "missing"
```
→ Add the variable to `stack.yaml`, `inventory/vars.yaml`, or secrets

**Type mismatch:**
```
error calling eq: incompatible types for comparison
```
→ Check variable types match (string vs int, etc.)

**Syntax error:**
```
template: :10:5: unexpected "}" in command
```
→ Check template syntax (missing `end`, extra braces, etc.)

## Best Practices

### Provide Defaults

Always provide sensible defaults in `stack.yaml`:

```yaml
vars:
  myapp:
    image: myapp:latest
    port: 8080
    replicas: 1
    debug: false
```

### Document Variables

Use comments to document configuration:

```yaml
vars:
  myapp:
    # Container image (default: latest)
    image: myapp:latest

    # External port (1024-65535)
    port: 8080

    # Enable debug logging (true/false)
    debug: false
```

### Organize Hierarchically

Use nested structures for clarity:

```yaml
vars:
  myapp:
    server:
      host: 0.0.0.0
      port: 8080
    database:
      host: postgres
      port: 5432
    features:
      - auth
      - metrics
```

### Keep Templates Simple

Avoid complex logic in templates:

```yaml
# Good: Simple conditional
{{ if .vars.myapp.ssl }}
- SSL_ENABLED=true
{{ end }}

# Bad: Complex nested logic
{{ if and (has "traefik" .stacks.enabled) (eq .vars.myapp.mode "production") (ne .vars.myapp.ssl false) }}
...
{{ end }}
```

Move complex logic to variables or preprocessing.

## See Also

- [Secrets Management](secrets.md) - Sensitive variable handling
- [Template Context Reference](../reference/template-context.md) - Full context structure
- [Stack Structure](stack-structure.md) - How stacks are organized
- [Gomplate Documentation](https://docs.gomplate.ca/) - Complete template syntax
