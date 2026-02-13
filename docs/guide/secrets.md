# Secrets Management

homelabctl supports encrypted secrets using [SOPS](https://github.com/getsops/sops) for secure storage of sensitive configuration.

## Overview

Secrets are stored in `secrets/<stack>.yaml` or `secrets/<stack>.enc.yaml` files and automatically merged into template variables with **highest priority**.

```
Stack defaults < Inventory vars < Secrets
                                  ↑ highest priority
```

## Quick Start

### 1. Install SOPS

```bash
# Using package manager
apt install sops        # Debian/Ubuntu
brew install sops       # macOS

# Or download binary
wget https://github.com/getsops/sops/releases/download/v3.8.1/sops-v3.8.1.linux.amd64
chmod +x sops-v3.8.1.linux.amd64
sudo mv sops-v3.8.1.linux.amd64 /usr/local/bin/sops
```

### 2. Configure Encryption

SOPS supports multiple backends. Choose one:

**Age (Recommended):**

```bash
# Generate key
age-keygen -o ~/.config/sops/age/keys.txt

# Create .sops.yaml in repository root
cat > .sops.yaml <<EOF
creation_rules:
  - path_regex: secrets/.*\.enc\.yaml$
    age: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
EOF
```

**GPG:**

```bash
# List GPG keys
gpg --list-secret-keys

# Create .sops.yaml
cat > .sops.yaml <<EOF
creation_rules:
  - path_regex: secrets/.*\.enc\.yaml$
    pgp: YOUR_GPG_FINGERPRINT
EOF
```

### 3. Create Encrypted Secrets

```bash
# Create/edit encrypted file
sops secrets/traefik.enc.yaml
```

Add your secrets:

```yaml
vars:
  traefik:
    cloudflare_api_token: "your-secret-token"
    admin_password: "secure-password"
```

Save and exit. File is automatically encrypted.

### 4. Use in Templates

```yaml
# stacks/traefik/compose.yml.tmpl
services:
  traefik:
    environment:
      - CF_API_TOKEN={{ .vars.traefik.cloudflare_api_token }}
```

## How It Works

### File Detection

homelabctl automatically detects and decrypts secrets:

- `secrets/<stack>.enc.yaml` → Decrypted with SOPS
- `secrets/<stack>.yaml` → Loaded as plain YAML (not recommended)

### Decryption Process

When you run `homelabctl generate` or `homelabctl deploy`:

1. homelabctl checks for `secrets/<stack>.enc.yaml`
2. Runs `sops -d secrets/<stack>.enc.yaml`
3. Merges decrypted values into template context
4. Renders templates
5. Decrypted values never touch disk unencrypted

### Variable Merging

Secrets override all other sources:

```yaml
# stack.yaml
vars:
  myapp:
    api_key: "default-key"
    debug: true

# inventory/vars.yaml
vars:
  myapp:
    debug: false

# secrets/myapp.enc.yaml
vars:
  myapp:
    api_key: "production-key"

# Result:
# api_key: "production-key"  (from secrets)
# debug: false               (from inventory)
```

## File Structure

### Encrypted File Format

```yaml
# secrets/mystack.enc.yaml (before encryption)
vars:
  mystack:
    # Database credentials
    db_password: "super-secret"
    db_user: "admin"

    # API tokens
    api_key: "sk_live_..."

    # OAuth secrets
    oauth_client_secret: "..."
```

After encryption with SOPS, it looks like:

```yaml
vars:
    mystack:
        db_password: ENC[AES256_GCM,data:...,iv:...,tag:...,type:str]
        db_user: ENC[AES256_GCM,data:...,iv:...,tag:...,type:str]
        # ... encrypted values
sops:
    kms: []
    gcp_kms: []
    azure_kv: []
    age:
        - recipient: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
          enc: |
            -----BEGIN AGE ENCRYPTED FILE-----
            ...
            -----END AGE ENCRYPTED FILE-----
    lastmodified: "2025-01-15T10:30:00Z"
```

### SOPS Configuration

`.sops.yaml` in repository root:

```yaml
# Encrypt all files in secrets/ ending in .enc.yaml
creation_rules:
  - path_regex: secrets/.*\.enc\.yaml$
    age: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p

  # Optional: Different keys for different stacks
  - path_regex: secrets/production-.*\.enc\.yaml$
    age: age1production...
```

## Common Workflows

### Edit Existing Secrets

```bash
sops secrets/mystack.enc.yaml
```

SOPS automatically decrypts, opens editor, re-encrypts on save.

### View Secrets

```bash
# Decrypt to stdout
sops -d secrets/mystack.enc.yaml

# Pretty print
sops -d secrets/mystack.enc.yaml | yq
```

### Rotate Keys

```bash
# Update .sops.yaml with new key
# Then rotate all files:
sops updatekeys secrets/*.enc.yaml
```

### Share Secrets with Team

```yaml
# .sops.yaml - Multiple recipients
creation_rules:
  - path_regex: secrets/.*\.enc\.yaml$
    age: >-
      age1person1...,
      age1person2...,
      age1person3...
```

```bash
# Re-encrypt with new recipients
sops updatekeys secrets/*.enc.yaml
```

### Backup Secrets

```bash
# Export decrypted (CAREFUL!)
sops -d secrets/mystack.enc.yaml > backup.yaml

# Better: Backup encrypted files
cp -r secrets/ secrets-backup/
```

## Security Best Practices

### Key Management

- **Never commit `.age/keys.txt`** or private GPG keys
- Store keys in password manager or hardware token
- Use separate keys for different environments (dev/staging/prod)
- Rotate keys periodically

### File Permissions

```bash
# Restrict access to key file
chmod 600 ~/.config/sops/age/keys.txt

# Keep secrets directory restricted
chmod 700 secrets/
```

### Git Configuration

```gitignore
# .gitignore - Never commit decrypted secrets
secrets/*.yaml
!secrets/*.enc.yaml  # Only encrypted files

# Never commit SOPS keys
.age/
*.key
```

### Environment Separation

```
secrets/
├── dev-mystack.enc.yaml      # Development secrets
├── staging-mystack.enc.yaml  # Staging secrets
└── prod-mystack.enc.yaml     # Production secrets
```

Use different keys per environment in `.sops.yaml`.

## Optional: Plain Secrets

For non-sensitive configuration, plain YAML is supported:

```yaml
# secrets/mystack.yaml (unencrypted)
vars:
  mystack:
    feature_flag: true
    log_level: debug
```

**Not recommended** for sensitive data. Use `.enc.yaml` instead.

## Troubleshooting

### "failed to get the data key"

**Cause:** SOPS cannot decrypt (wrong key or no access)

**Fix:**
```bash
# Verify key is available
age-keygen -y ~/.config/sops/age/keys.txt

# Check .sops.yaml matches key
cat .sops.yaml
```

### "no key found"

**Cause:** Missing `.sops.yaml` or key not configured

**Fix:**
```bash
# Create .sops.yaml
cat > .sops.yaml <<EOF
creation_rules:
  - path_regex: secrets/.*\.enc\.yaml$
    age: $(age-keygen -y ~/.config/sops/age/keys.txt)
EOF
```

### "MAC mismatch"

**Cause:** File corrupted or tampered with

**Fix:**
```bash
# Verify file integrity
sops -d secrets/mystack.enc.yaml

# If corrupted, restore from backup
git checkout secrets/mystack.enc.yaml
```

### Templates not using secrets

**Cause:** Variable path mismatch

**Fix:**
```yaml
# Wrong:
vars:
  different_name:
    api_key: "..."

# Correct: Must match stack name
vars:
  mystack:
    api_key: "..."
```

## See Also

- [Variables & Templating](variables.md) - Variable precedence
- [SOPS Documentation](https://github.com/getsops/sops) - Full SOPS guide
- [Age Encryption](https://github.com/FiloSottile/age) - Modern encryption tool
- [Stack Structure](stack-structure.md) - Repository organization
