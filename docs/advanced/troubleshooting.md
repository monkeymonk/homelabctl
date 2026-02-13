# Troubleshooting

Common issues and solutions when using homelabctl.

## Repository Issues

### "Not in a homelab repository"

**Symptom:**
```
Error: not in a homelab repository
Run: homelabctl init
```

**Cause:** Command run outside repository root

**Fix:**
```bash
# Navigate to repository root
cd /path/to/homelab

# Or initialize new repository
homelabctl init
```

### "Missing required directories"

**Symptom:**
```
Error: repository structure invalid
Missing: stacks/, enabled/
```

**Cause:** Incomplete repository initialization

**Fix:**
```bash
homelabctl init  # Creates missing directories
```

## Stack Issues

### "Stack not found"

**Symptom:**
```
Error: stack 'mystack' does not exist

Check stacks/ directory
Run: homelabctl list
```

**Cause:** Stack directory doesn't exist

**Fix:**
```bash
# List available stacks
ls stacks/

# Create stack
mkdir -p stacks/mystack
# Add stack.yaml and compose.yml.tmpl
```

### "Dependency not satisfied"

**Symptom:**
```
Error: stack 'app' has unsatisfied dependencies

Dependency chain:
app requires: [traefik]
Missing: [traefik]

To resolve:
  → Run: homelabctl enable traefik
```

**Cause:** Required stack not enabled

**Fix:**
```bash
# Enable dependency first
homelabctl enable traefik

# Then enable dependent stack
homelabctl enable app
```

### "Circular dependency detected"

**Symptom:**
```
Error: circular dependency detected

Cycle: traefik → authentik → traefik

Dependencies must form a directed acyclic graph (DAG)
```

**Cause:** Stack dependencies form a loop

**Fix:**
```yaml
# Remove circular dependency in stack.yaml
# Either traefik shouldn't require authentik,
# or authentik shouldn't require traefik

# Wrong:
# stacks/traefik/stack.yaml
requires:
  - authentik

# stacks/authentik/stack.yaml
requires:
  - traefik

# Fixed:
# stacks/authentik/stack.yaml
requires:
  - traefik  # Only one direction
```

## Template Issues

### "Template syntax error"

**Symptom:**
```
Error: failed to render traefik
template: :10:5: unexpected "}" in command
```

**Cause:** Invalid gomplate syntax

**Fix:**
```yaml
# Check template around line 10
# Common issues:

# Missing {{ end }}
{{ if .vars.ssl }}
environment:
  - SSL=true
# Missing: {{ end }}

# Extra braces
image: {{ .vars.image }}}}  # One too many }

# Wrong function syntax
{{ has .stacks.enabled "traefik" }}  # Wrong order
{{ has "traefik" .stacks.enabled }}  # Correct
```

### "Missing variable"

**Symptom:**
```
template: :5:10: executing "" at <.vars.missing>: map has no entry for key "missing"
```

**Cause:** Variable used in template but not defined

**Fix:**
```yaml
# Add to stack.yaml
vars:
  myapp:
    missing: "default value"

# Or inventory/vars.yaml
vars:
  missing: "global value"

# Or use default in template
image: {{ .vars.image | default "nginx:latest" }}
```

### "Type mismatch"

**Symptom:**
```
error calling eq: incompatible types for comparison
```

**Cause:** Comparing different types (string vs int, etc.)

**Fix:**
```yaml
# Wrong: Comparing string to int
{{ if eq .vars.port 8080 }}

# Fixed: Convert to same type
{{ if eq .vars.port "8080" }}
# or
{{ if eq (.vars.port | int) 8080 }}
```

## Secrets Issues

### "Failed to decrypt SOPS"

**Symptom:**
```
Error: failed to decrypt secrets/traefik.enc.yaml
sops: failed to get the data key
```

**Cause:** SOPS key not available or wrong key

**Fix:**
```bash
# Check if key exists
age-keygen -y ~/.config/sops/age/keys.txt

# Verify .sops.yaml configuration
cat .sops.yaml

# Check file was encrypted with correct key
sops -d secrets/traefik.enc.yaml
```

### "SOPS not found"

**Symptom:**
```
Error: exec: "sops": executable file not found in $PATH
```

**Cause:** SOPS not installed

**Fix:**
```bash
# Install SOPS
apt install sops  # Debian/Ubuntu
brew install sops  # macOS

# Or download binary
wget https://github.com/getsops/sops/releases/latest/download/sops-linux-amd64
chmod +x sops-linux-amd64
sudo mv sops-linux-amd64 /usr/local/bin/sops
```

## Generation Issues

### "Duplicate service name"

**Symptom:**
```
Error: failed to merge compose files
duplicate service name: myapp
```

**Cause:** Multiple stacks define same service name

**Fix:**
```yaml
# Use unique service names per stack

# stacks/stack1/compose.yml.tmpl
services:
  stack1-myapp:  # Prefix with stack name
    # ...

# stacks/stack2/compose.yml.tmpl
services:
  stack2-myapp:  # Different name
    # ...
```

### "Invalid compose syntax"

**Symptom:**
```
Error: failed to parse compose file
yaml: line 10: mapping values are not allowed in this context
```

**Cause:** Invalid YAML in generated compose file

**Fix:**
```bash
# Generate with debug mode
homelabctl generate --debug

# Check problematic file
cat runtime/problematic-stack-compose.yml

# Look for:
# - Missing quotes around strings with special chars
# - Incorrect indentation
# - Invalid port syntax
```

### "Permission denied writing output"

**Symptom:**
```
Error: failed to write runtime/docker-compose.yml
permission denied
```

**Cause:** Insufficient permissions

**Fix:**
```bash
# Check permissions
ls -la runtime/

# Fix ownership
sudo chown -R $USER:$USER runtime/

# Or run with sudo (not recommended)
sudo homelabctl generate
```

## Deployment Issues

### "Container fails to start"

**Symptom:**
```
homelabctl ps
myapp    Exit 1
```

**Cause:** Various (check logs)

**Fix:**
```bash
# Check service logs
homelabctl logs myapp

# Common issues:
# 1. Missing environment variables
# 2. Port already in use
# 3. Volume mount path doesn't exist
# 4. Network issues

# Check generated compose
cat runtime/docker-compose.yml

# Validate compose syntax
docker compose -f runtime/docker-compose.yml config
```

### "Port already allocated"

**Symptom:**
```
Error: failed to deploy
Bind for 0.0.0.0:80 failed: port is already allocated
```

**Cause:** Another service using same port

**Fix:**
```bash
# Find what's using the port
sudo lsof -i :80
# or
sudo ss -tulpn | grep :80

# Stop conflicting service
sudo systemctl stop apache2

# Or change port in inventory/vars.yaml
vars:
  myapp:
    port: 8080  # Use different port
```

### "Network not found"

**Symptom:**
```
Error: network "traefik_default" not found
```

**Cause:** Required network doesn't exist

**Fix:**
```bash
# Deploy dependency stack first
homelabctl deploy traefik

# Or create network manually
docker network create traefik_default
```

## Service Control Issues

### "Service not disabled"

**Symptom:** Service still in runtime/docker-compose.yml after `homelabctl disable -s service`

**Cause:** Name mismatch or not regenerated

**Fix:**
```bash
# Check exact service name
homelabctl list

# Disable with exact name
homelabctl disable -s exact-service-name

# Regenerate compose file
homelabctl generate

# Verify
grep service-name runtime/docker-compose.yml
```

### "Service not found error"

**Symptom:**
```
Error: service 'myservice' not found in any enabled stack
```

**Cause:** Service doesn't exist or stack not enabled

**Fix:**
```bash
# List all enabled stacks and services
homelabctl list

# Enable stack containing service
homelabctl enable mystack

# Then disable specific service
homelabctl disable -s myservice
```

## Validation Issues

### "Category dependency validation failed"

**Symptom:**
```
Error: invalid category dependency
infrastructure (order 2) depends on media (order 5)
```

**Cause:** Lower-order category depending on higher-order

**Fix:**
```yaml
# Fix category in stack.yaml
# Either:
# 1. Change category to higher order
name: mystack
category: media  # Match dependency category

# 2. Or remove invalid dependency
requires:
  # - media  # Remove this
  - core     # OK: lower order
```

## Docker Compose Passthrough Issues

### "Unknown docker compose command"

**Symptom:**
```
Error: unknown command
```

**Cause:** Invalid docker compose command

**Fix:**
```bash
# Check valid docker compose commands
docker compose --help

# homelabctl passes commands directly to docker compose
homelabctl config   # Actually: docker compose -f runtime/docker-compose.yml config
homelabctl pull     # Actually: docker compose -f runtime/docker-compose.yml pull
```

## Debug Techniques

### Enable Debug Mode

```bash
homelabctl generate --debug
```

Preserves temporary files for inspection.

### Inspect Temporary Files

```bash
# After debug generate
ls runtime/

# View per-stack compose files
cat runtime/traefik-compose.yml
cat runtime/authentik-compose.yml

# View final merged output
cat runtime/docker-compose.yml
```

### Test Templates Manually

```bash
# Create test context
cat > /tmp/test-context.yaml <<EOF
vars:
  myapp:
    image: nginx:latest
    port: 8080
stack:
  name: myapp
  category: tools
stacks:
  enabled:
    - traefik
    - myapp
EOF

# Test template
gomplate -f stacks/myapp/compose.yml.tmpl -c .=/tmp/test-context.yaml
```

### Validate Compose File

```bash
# Generate compose
homelabctl generate

# Validate syntax
docker compose -f runtime/docker-compose.yml config

# Check for errors without starting
docker compose -f runtime/docker-compose.yml up --dry-run
```

### Check Dependencies

```bash
# Visualize dependency tree
homelabctl list

# Validate dependencies
homelabctl validate
```

## Getting Help

### Check Documentation

```bash
# View help
homelabctl --help

# Check specific command
homelabctl generate --help
```

### Enable Verbose Logging

```bash
# Show more output
set -x  # In bash scripts
homelabctl generate
set +x
```

### Report Issues

When reporting issues, include:

1. **homelabctl version:**
   ```bash
   homelabctl --version
   ```

2. **Command that failed:**
   ```bash
   homelabctl generate
   ```

3. **Full error message**

4. **Minimal reproduction:**
   - Stack configuration
   - Inventory vars
   - Steps to reproduce

5. **Environment:**
   - OS version
   - Docker version
   - Go version (if building from source)

## See Also

- [Architecture](architecture.md) - Understanding homelabctl design
- [Variables](../guide/variables.md) - Template variable system
- [Secrets](../guide/secrets.md) - Secret management
- [Commands](../guide/commands.md) - Command reference
