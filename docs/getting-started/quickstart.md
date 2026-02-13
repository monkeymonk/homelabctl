# Quick Start

Get homelabctl up and running in 5 minutes!

## 1. Initialize Your Homelab

```bash
mkdir ~/homelab && cd ~/homelab
homelabctl init
```

This creates the required directory structure:

```
homelab/
├── stacks/       # Stack definitions (commit to git)
├── enabled/      # Symlinks to enabled stacks
├── inventory/    # Your configuration
│   └── vars.yaml
├── secrets/      # Encrypted secrets
└── runtime/      # Generated files (gitignored)
```

## 2. Configure Your Domain

Edit `inventory/vars.yaml`:

```yaml
domain: homelab.local
timezone: America/New_York
```

## 3. Create Your First Stack

Create a simple nginx stack:

```bash
mkdir -p stacks/nginx
```

Create `stacks/nginx/stack.yaml`:

```yaml
name: nginx
category: tools
requires: []
services:
  - nginx
vars:
  nginx:
    image: nginx:alpine
    port: 8080
```

Create `stacks/nginx/compose.yml.tmpl`:

```yaml
services:
  nginx:
    image: {{ .vars.nginx.image }}
    container_name: nginx
    ports:
      - "{{ .vars.nginx.port }}:80"
    restart: unless-stopped
```

## 4. Enable and Deploy

```bash
# Enable the stack
homelabctl enable nginx

# Validate configuration
homelabctl validate

# Generate and deploy
homelabctl deploy
```

## 5. Verify It's Running

```bash
# Check service status
homelabctl ps

# View logs
homelabctl logs nginx

# Test the service
curl http://localhost:8080
```

## What Just Happened?

1. ✅ Created a homelab repository structure
2. ✅ Configured global variables
3. ✅ Defined a stack with templates
4. ✅ Enabled and deployed the stack
5. ✅ Verified it's running

## Next Steps

- [Your First Stack](first-stack.md) - Deeper dive into stack creation
- [Commands Reference](../guide/commands.md) - Learn all available commands
- [Variables & Templating](../guide/variables.md) - Advanced configuration

## Common Commands

```bash
# List enabled stacks
homelabctl list

# Disable a stack
homelabctl disable nginx

# Regenerate after changes
homelabctl generate

# Restart services
homelabctl restart nginx

# Stop everything
homelabctl down
```

!!! success "You're ready!"
    You now have a working homelabctl setup. Explore the [User Guide](../guide/commands.md) to learn more!
