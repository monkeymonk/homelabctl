# Categories

Categories organize stacks and control deployment order. They're **automatically discovered** from stack definitions—no code changes needed.

## Overview

Every stack belongs to a category defined in `stack.yaml`:

```yaml
name: traefik
category: infrastructure  # Determines deployment order
```

Categories ensure dependencies deploy before dependents (e.g., core infrastructure before applications).

## Built-in Categories

homelabctl includes predefined categories with optimized defaults:

| Category | Order | Purpose | Default Settings |
|----------|-------|---------|------------------|
| `core` | 1 | Essential services (DNS, auth) | Security options |
| `infrastructure` | 2 | Network, proxy, storage | Security options |
| `monitoring` | 3 | Metrics, logs, alerts | - |
| `automation` | 4 | CI/CD, task automation | - |
| `media` | 5 | Media servers, downloaders | PUID/PGID |
| `tools` | 6 | Utilities, management | - |

**Deployment order:** Core → Infrastructure → Monitoring → Automation → Media → Tools

## Dynamic Discovery

Categories are **discovered automatically** from stack files. No registration required.

### Creating Custom Categories

Just use a new category name in `stack.yaml`:

```yaml
name: myapp
category: custom-category  # Automatically discovered
```

**Default behavior for unknown categories:**
- DisplayName: Auto-capitalized (`custom-category` → `Custom Category`)
- Order: 999 (deploys last)
- Color: white
- Defaults: none

### Customizing Category Metadata

To add predefined settings for a custom category, edit `internal/categories/categories.go`:

```go
"custom-category": {
    Name:        "custom-category",
    DisplayName: "My Custom Category",
    Order:       7,  // After tools (6)
    Color:       "cyan",
    Defaults: map[string]interface{}{
        "restart": "unless-stopped",
        "logging": map[string]interface{}{
            "driver": "json-file",
            "options": map[string]string{
                "max-size": "10m",
                "max-file": "3",
            },
        },
    },
}
```

## Category Defaults

Categories can provide default variables applied to all stacks in that category.

### Example: Media Category

```yaml
# Predefined in code
category: media
defaults:
  environment:
    PUID: 1000
    PGID: 1000
  restart: unless-stopped
```

All media stacks inherit these defaults:

```yaml
# stacks/jellyfin/stack.yaml
name: jellyfin
category: media  # Inherits PUID/PGID automatically
```

### Variable Precedence

Category defaults have **lowest priority**:

```
Category defaults < Stack defaults < Inventory < Secrets
```

Example:

```yaml
# Category (media)
PUID: 1000

# Stack defaults
PUID: 1001  # Overrides category

# Inventory
PUID: 1002  # Overrides stack

# Secrets
PUID: 1003  # Overrides inventory (highest)
```

## Deployment Order

Stacks deploy in **category order**, then **alphabetically** within each category.

### Example Deployment Sequence

```
1. core/
   - dns
   - vpn

2. infrastructure/
   - traefik
   - authentik

3. monitoring/
   - grafana
   - prometheus

4. media/
   - jellyfin
   - sonarr

5. custom-category/  (order: 999)
   - myapp
```

### Controlling Order

**Option 1: Use predefined categories**
```yaml
category: infrastructure  # Deploys at order 2
```

**Option 2: Define custom order** (requires code edit)
```go
"myapp-category": {
    Order: 2.5,  // Between infrastructure (2) and monitoring (3)
}
```

**Option 3: Use dependencies**
```yaml
# Even within same category, dependencies control order
requires:
  - other-stack
```

## Category Validation

homelabctl validates category dependencies to prevent ordering issues.

### Valid Dependencies

Higher-order categories can depend on lower-order:

```yaml
# ✓ Valid: media (5) depends on infrastructure (2)
name: jellyfin
category: media
requires:
  - traefik  # category: infrastructure
```

### Invalid Dependencies

Lower-order cannot depend on higher-order:

```yaml
# ✗ Invalid: infrastructure (2) cannot depend on media (5)
name: traefik
category: infrastructure
requires:
  - jellyfin  # category: media → ERROR
```

This prevents circular deployment issues.

## Listing Categories

View all discovered categories:

```bash
# List enabled stacks grouped by category
homelabctl list

# Validate category dependencies
homelabctl validate
```

## Best Practices

### Choose Appropriate Categories

Match category to function:

```yaml
# Good: Reverse proxy is infrastructure
name: traefik
category: infrastructure

# Bad: Reverse proxy as media
name: traefik
category: media  # Wrong order, may not be available
```

### Leverage Category Defaults

Use categories to reduce boilerplate:

```yaml
# Media category provides PUID/PGID automatically
name: sonarr
category: media  # Inherits user/group IDs

# No need to specify in every stack
```

### Custom Categories for Grouping

Group related custom stacks:

```yaml
# All your internal tools
name: mytool
category: internal-tools

name: myapp
category: internal-tools

# Deploy together, alphabetically
```

### Avoid Over-categorization

Don't create too many categories:

```yaml
# Bad: Too specific
category: media-downloaders
category: media-servers
category: media-management

# Good: Use single category
category: media
# Let dependencies control order within category
```

## Use Cases

### Environment Separation

Categories can help separate environments:

```yaml
# Production services
category: prod-apps

# Development services
category: dev-apps
```

Enable selectively:

```bash
# Only production
homelabctl enable traefik authentik myapp

# Add development
homelabctl enable dev-myapp dev-database
```

### Resource Tiers

Organize by resource requirements:

```yaml
# High-priority, always-on
category: core

# Medium priority, restartable
category: services

# Low priority, optional
category: optional
```

### Maintenance Windows

Deploy by category in maintenance windows:

```bash
# Update core first
homelabctl restart $(homelabctl list | grep "core/" | awk '{print $1}')

# Then infrastructure
homelabctl restart $(homelabctl list | grep "infrastructure/" | awk '{print $1}')
```

## See Also

- [Stack Structure](stack-structure.md) - How stacks are organized
- [Commands](commands.md) - Managing stacks
- [Dependencies](stack-structure.md#dependencies) - Stack dependency system
