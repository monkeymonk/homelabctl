# Stack Structure

Learn how stacks are organized and how to structure your homelab repository.

## Repository Layout

```
homelab/
├── stacks/              # Stack definitions (PUBLIC)
│   ├── traefik/
│   ├── authentik/
│   └── jellyfin/
├── enabled/             # Symlinks to enabled stacks
│   ├── traefik -> ../stacks/traefik
│   └── authentik -> ../stacks/authentik
├── inventory/           # Your configuration (PRIVATE)
│   └── vars.yaml
├── secrets/             # Encrypted secrets (PRIVATE)
│   └── traefik.enc.yaml
└── runtime/             # Generated files (NEVER COMMIT)
    └── docker-compose.yml
```

## Stack Directory Structure

```
stacks/mystack/
├── stack.yaml           # Manifest + default variables
├── compose.yml.tmpl     # Docker Compose template
├── config/              # Configuration file templates (optional)
│   └── app.conf.tmpl
└── contribute/          # Cross-stack contributions (optional)
    └── traefik/
        └── routes.yml.tmpl
```

## stack.yaml Schema

```yaml
name: mystack              # Stack identifier (must match directory name)
category: tools            # Category for deployment ordering
requires:                  # Dependencies
  - core
  - traefik
services:                  # List of all services (REQUIRED)
  - myapp
  - worker
vars:                      # Default variables
  myapp:
    image: myapp:latest
    port: 8080
  worker:
    image: worker:latest
persistence:               # Data persistence
  volumes:
    - myapp_data
  paths:
    - ./runtime/mystack
```

## compose.yml.tmpl

Standard Docker Compose file with template variables:

```yaml
services:
  myapp:
    image: {{ .vars.myapp.image }}
    container_name: myapp
    ports:
      - "{{ .vars.myapp.port }}:8080"
    volumes:
      - myapp_data:/data
    restart: unless-stopped
    {{ if has "traefik" .stacks.enabled }}
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.myapp.rule=Host(`myapp.{{ .vars.domain }}`)"
    {{ end }}

volumes:
  myapp_data:
```

See [Variables & Templating](variables.md) for template syntax.

## Best Practices

### Naming

- **Stack names**: lowercase, hyphen-separated (`my-stack`)
- **Service names**: match container names for clarity
- **Volume names**: prefix with stack name (`mystack_data`)

### Organization

- Group related services in one stack
- Use categories to control deployment order
- Keep stacks self-contained when possible

### Variables

- Provide sensible defaults in `stack.yaml`
- Document all variables with comments
- Use nested structure for multi-service stacks

### Files

- Commit: `stacks/`, `enabled/`
- Private: `inventory/`, `secrets/`
- Generated: `runtime/` (gitignored)

## See Also

- [Variables & Templating](variables.md)
- [Categories](categories.md)
- [Secrets Management](secrets.md)
