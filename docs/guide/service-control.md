# Service Control

homelabctl allows disabling individual services without disabling entire stacks, giving fine-grained control over your deployment.

## Overview

**Stack-level control:**
- `homelabctl enable <stack>` - Enable entire stack
- `homelabctl disable <stack>` - Disable entire stack

**Service-level control:**
- `homelabctl disable -s <service>` - Disable specific service (keeps stack enabled)
- `homelabctl enable -s <service>` - Re-enable disabled service

Disabled services are filtered out during `generate` and won't appear in the final `runtime/docker-compose.yml`.

## Why Disable Services?

### Hardware Limitations

Skip services that don't work on your hardware:

```bash
# Disable S.M.A.R.T. monitoring without drives
homelabctl disable -s scrutiny

# Disable GPU-dependent transcoding
homelabctl disable -s plex-transcoder
```

### Resource Constraints

Save resources on constrained systems:

```bash
# Disable resource-intensive services
homelabctl disable -s elasticsearch
homelabctl disable -s grafana-loki
```

### Temporary Changes

Temporarily disable services during maintenance:

```bash
# Disable backup service during migration
homelabctl disable -s backup-runner
# ... perform migration ...
homelabctl enable -s backup-runner
```

### Feature Toggles

Control optional features:

```bash
# Disable optional monitoring
homelabctl disable -s metrics-exporter

# Disable development tools in production
homelabctl disable -s debug-ui
```

## How It Works

### State Storage

Disabled services are stored in `inventory/vars.yaml`:

```yaml
# inventory/vars.yaml
disabled_services:
  - scrutiny
  - elasticsearch
  - debug-ui

vars:
  domain: example.com
  # ... other variables
```

This file is **committed to git** and shared across your team.

### Generation Process

When you run `homelabctl generate`:

1. Load enabled stacks from `enabled/` symlinks
2. Render each stack's `compose.yml.tmpl`
3. **Filter out disabled services** from compose files
4. Merge filtered compose files into `runtime/docker-compose.yml`

### Stack Definitions Preserved

Service definitions remain in `stack.yaml` for documentation:

```yaml
# stacks/monitoring/stack.yaml
name: monitoring
services:
  - prometheus   # ✓ Enabled
  - grafana      # ✓ Enabled
  - loki         # ✗ Disabled via inventory/vars.yaml
  - tempo        # ✓ Enabled
```

The stack knows about all services, but `loki` won't be deployed.

## Usage

### Disable a Service

```bash
homelabctl disable -s scrutiny
```

**Output:**
```
Service 'scrutiny' disabled
Updated inventory/vars.yaml
```

**What happens:**
1. Adds `scrutiny` to `disabled_services` list in `inventory/vars.yaml`
2. Service will be excluded from next `generate` or `deploy`

### Re-enable a Service

```bash
homelabctl enable -s scrutiny
```

**Output:**
```
Service 'scrutiny' re-enabled
Updated inventory/vars.yaml
```

**What happens:**
1. Removes `scrutiny` from `disabled_services` list
2. Service will be included in next `generate` or `deploy`

### View Disabled Services

```bash
homelabctl list
```

**Output:**
```
Enabled stacks:
  core/traefik
  infrastructure/authentik
  monitoring/prometheus
  monitoring/grafana

Disabled services:
  - scrutiny (in monitoring stack)
  - loki (in monitoring stack)
```

### Validate Configuration

```bash
homelabctl validate
```

Checks:
- ✓ All disabled services exist in enabled stacks
- ✓ No dependency issues from disabled services
- ✓ Valid `inventory/vars.yaml` syntax

## Examples

### Skip Hardware-Specific Service

```yaml
# stacks/monitoring/stack.yaml
name: monitoring
services:
  - prometheus
  - grafana
  - scrutiny      # Requires S.M.A.R.T. drives
```

```bash
# System without drives
homelabctl enable monitoring
homelabctl disable -s scrutiny
homelabctl deploy
```

**Result:** Prometheus and Grafana deploy, Scrutiny skipped.

### Conditional Features

```yaml
# stacks/myapp/stack.yaml
services:
  - myapp-api
  - myapp-web
  - myapp-worker
  - myapp-metrics  # Optional metrics endpoint
```

```bash
# Production: All services
homelabctl enable myapp

# Development: Skip metrics
homelabctl enable myapp
homelabctl disable -s myapp-metrics
```

### Multi-Service Stack Management

```yaml
# stacks/media/stack.yaml
services:
  - jellyfin        # Core media server
  - sonarr          # TV management
  - radarr          # Movie management
  - prowlarr        # Indexer management
  - transmission    # Downloader
```

```bash
# Enable stack
homelabctl enable media

# Only want Jellyfin + Sonarr
homelabctl disable -s radarr
homelabctl disable -s prowlarr
homelabctl disable -s transmission

# Deploy minimal media setup
homelabctl deploy
```

## Dependency Handling

### Service Dependencies Within Stack

If services depend on each other, disabling may cause issues:

```yaml
services:
  myapp:
    depends_on:
      - database
  database:
```

```bash
# ✗ Bad: Disable dependency
homelabctl disable -s database  # myapp will fail to start
```

**Solution:** Disable both or keep dependency enabled.

### Cross-Stack Dependencies

Disabling services doesn't affect stack-level dependencies:

```yaml
# stacks/myapp/stack.yaml
name: myapp
requires:
  - traefik  # Stack dependency

services:
  - myapp-api  # Individual service
```

```bash
# ✓ Safe: Stack dependencies still enforced
homelabctl disable -s myapp-api  # Traefik still required by stack
```

## Best Practices

### Document Disabled Services

Add comments in `inventory/vars.yaml`:

```yaml
# Disabled services
disabled_services:
  - scrutiny       # No S.M.A.R.T. drives on this host
  - loki           # Too resource-intensive for this setup
  - debug-ui       # Production environment
```

### Use Service Naming Convention

Name services clearly in `stack.yaml`:

```yaml
# Good: Clear service names
services:
  - myapp-api          # Core API
  - myapp-worker       # Background jobs
  - myapp-metrics      # Optional metrics (can disable)

# Bad: Vague names
services:
  - app
  - service1
  - service2
```

### Validate After Changes

Always validate after disabling services:

```bash
homelabctl disable -s myservice
homelabctl validate  # Check for issues
homelabctl generate  # Verify compose output
```

### Prefer Service Control Over Stack Control

```bash
# Good: Granular control
homelabctl enable monitoring
homelabctl disable -s loki  # Skip one service

# Less flexible: All or nothing
homelabctl disable monitoring  # Loses all monitoring
```

### Commit Configuration

Disabled services persist in `inventory/vars.yaml`:

```bash
git add inventory/vars.yaml
git commit -m "Disable Scrutiny on systems without drives"
```

Team members inherit the same configuration.

## Troubleshooting

### Service Not Disabled

**Symptom:** Service still appears after `homelabctl generate`

**Causes:**
1. Typo in service name
2. Service from different stack
3. Changes not applied

**Fix:**
```bash
# Verify disabled services list
cat inventory/vars.yaml | grep disabled_services -A 10

# Re-disable with exact name
homelabctl list  # Find exact service name
homelabctl disable -s exact-service-name

# Regenerate
homelabctl generate
```

### "Service not found" Error

**Symptom:** `homelabctl disable -s myservice` fails

**Cause:** Service not in any enabled stack

**Fix:**
```bash
# List enabled stacks and their services
homelabctl list

# Enable stack first
homelabctl enable mystack

# Then disable service
homelabctl disable -s myservice
```

### Dependent Service Fails

**Symptom:** Service fails to start after disabling another

**Cause:** Disabled service was a dependency

**Fix:**
```bash
# Check compose file for depends_on
cat runtime/docker-compose.yml | grep -A5 "myservice:"

# Re-enable dependency or disable both
homelabctl enable -s dependency-service
# OR
homelabctl disable -s dependent-service
```

## Comparison: Stack vs Service Control

| Action | Scope | Command | Use Case |
|--------|-------|---------|----------|
| Disable stack | All services | `homelabctl disable mystack` | Don't need entire functionality |
| Disable service | Single service | `homelabctl disable -s myservice` | Skip one optional/incompatible service |
| Enable stack | All services | `homelabctl enable mystack` | Add functionality |
| Enable service | Single service | `homelabctl enable -s myservice` | Re-enable previously disabled service |

## See Also

- [Commands](commands.md) - Full command reference
- [Stack Structure](stack-structure.md) - How services are defined
- [Variables](variables.md) - Configuration management
