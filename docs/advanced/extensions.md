# Extension Points

homelabctl is designed to be extensible. This document describes the primary extension points and how to add new functionality.

## Overview

homelabctl can be extended in several ways:

1. **SOPS Integration** - Add secret backends
2. **Contribute Patterns** - Add cross-stack contributions (Nginx, etc.)
3. **Validation Rules** - Add custom validation logic
4. **Pipeline Stages** - Add processing stages
5. **Commands** - Add new CLI commands

## 1. SOPS Integration

### Current Implementation

Location: `internal/secrets/secrets.go`

```go
func Load(stackName string) (map[string]interface{}, error) {
    // Try encrypted file first
    encPath := path.Join("secrets", stackName+".enc.yaml")
    if fileExists(encPath) {
        return decryptSOPS(encPath)
    }

    // Fall back to plain YAML
    plainPath := path.Join("secrets", stackName+".yaml")
    if fileExists(plainPath) {
        return loadPlainYAML(plainPath)
    }

    return nil, os.ErrNotExist
}

func decryptSOPS(path string) (map[string]interface{}, error) {
    cmd := exec.Command("sops", "-d", path)
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    var data map[string]interface{}
    if err := yaml.Unmarshal(output, &data); err != nil {
        return nil, err
    }

    return data, nil
}
```

### Adding Alternative Backend

To support a different secret backend (e.g., Vault, AWS Secrets Manager):

```go
// internal/secrets/vault.go
func loadFromVault(stackName string) (map[string]interface{}, error) {
    client, err := vault.NewClient(&vault.Config{
        Address: os.Getenv("VAULT_ADDR"),
    })
    if err != nil {
        return nil, err
    }

    secret, err := client.Logical().Read(fmt.Sprintf("secret/homelabctl/%s", stackName))
    if err != nil {
        return nil, err
    }

    return secret.Data, nil
}

// Update Load() to try multiple backends
func Load(stackName string) (map[string]interface{}, error) {
    // Try SOPS first
    if data, err := loadSOPS(stackName); err == nil {
        return data, nil
    }

    // Try Vault
    if vaultEnabled() {
        if data, err := loadFromVault(stackName); err == nil {
            return data, nil
        }
    }

    // Fall back to plain YAML
    return loadPlainYAML(stackName)
}
```

### Configuration

Add backend configuration to `inventory/vars.yaml`:

```yaml
secrets:
  backend: vault  # or "sops", "aws-secretsmanager"
  vault:
    address: https://vault.example.com
    path_prefix: homelabctl
```

## 2. Contribute Patterns

### Current Implementation

Location: `cmd/generate.go` (around line 100)

```go
// Handle Traefik contributions
traefikDir := filepath.Join("runtime", "traefik", "dynamic")
if err := os.MkdirAll(traefikDir, 0755); err != nil {
    return err
}

for _, stackName := range enabledStacks {
    contributePath := filepath.Join("stacks", stackName, "contribute", "traefik")
    if !dirExists(contributePath) {
        continue
    }

    // Render Traefik contribution templates
    templates, _ := filepath.Glob(filepath.Join(contributePath, "*.tmpl"))
    for _, tmpl := range templates {
        outputName := strings.TrimSuffix(filepath.Base(tmpl), ".tmpl")
        outputPath := filepath.Join(traefikDir, stackName+"-"+outputName)

        if err := render.Template(tmpl, context, outputPath); err != nil {
            return err
        }
    }
}
```

### Adding Nginx Contributions

```go
// internal/pipeline/contributions.go
func RenderContributions() Stage {
    return func(ctx *Context) error {
        // Define contribution handlers
        handlers := map[string]ContributionHandler{
            "traefik": renderTraefikContributions,
            "nginx":   renderNginxContributions,
            "caddy":   renderCaddyContributions,
        }

        for provider, handler := range handlers {
            if err := handler(ctx, provider); err != nil {
                return err
            }
        }

        return nil
    }
}

type ContributionHandler func(*Context, string) error

func renderNginxContributions(ctx *Context, provider string) error {
    outputDir := filepath.Join("runtime", "nginx", "conf.d")
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        return err
    }

    for _, stackName := range ctx.EnabledStacks {
        contributePath := filepath.Join("stacks", stackName, "contribute", provider)
        if !dirExists(contributePath) {
            continue
        }

        templates, _ := filepath.Glob(filepath.Join(contributePath, "*.tmpl"))
        for _, tmpl := range templates {
            outputName := strings.TrimSuffix(filepath.Base(tmpl), ".tmpl")
            outputPath := filepath.Join(outputDir, stackName+"-"+outputName)

            config := ctx.StackConfigs[stackName]
            templateCtx := buildTemplateContext(stackName, config, ctx)

            if err := render.Template(tmpl, templateCtx, outputPath); err != nil {
                return err
            }
        }
    }

    return nil
}
```

### Usage in Stacks

```
stacks/myapp/
├── stack.yaml
├── compose.yml.tmpl
└── contribute/
    ├── traefik/
    │   └── routes.yml.tmpl
    ├── nginx/
    │   └── server.conf.tmpl
    └── caddy/
        └── site.caddy.tmpl
```

## 3. Validation Rules

### Current Implementation

Location: `cmd/validate.go`

```go
func Validate() error {
    // Verify repository
    if err := fs.VerifyRepository(); err != nil {
        return err
    }

    // Load enabled stacks
    enabled, err := fs.ListEnabled()
    if err != nil {
        return err
    }

    // Validate dependencies
    if err := stacks.ValidateDependencies(enabled); err != nil {
        return err
    }

    fmt.Println("✓ All validations passed")
    return nil
}
```

### Adding Custom Validation

```go
// internal/validation/rules.go
type Rule interface {
    Name() string
    Validate(*Context) error
}

type MandatoryStacksRule struct {
    Required []string
}

func (r *MandatoryStacksRule) Name() string {
    return "mandatory-stacks"
}

func (r *MandatoryStacksRule) Validate(ctx *Context) error {
    enabled := make(map[string]bool)
    for _, stack := range ctx.EnabledStacks {
        enabled[stack] = true
    }

    var missing []string
    for _, required := range r.Required {
        if !enabled[required] {
            missing = append(missing, required)
        }
    }

    if len(missing) > 0 {
        return fmt.Errorf("missing mandatory stacks: %v", missing)
    }

    return nil
}

// Add to validation
func Validate() error {
    ctx := &Context{}
    // ... load context ...

    rules := []Rule{
        &MandatoryStacksRule{Required: []string{"traefik"}},
        &SecretsRequiredRule{Stacks: []string{"authentik"}},
        &CategoryOrderRule{},
    }

    for _, rule := range rules {
        if err := rule.Validate(ctx); err != nil {
            return fmt.Errorf("%s validation failed: %w", rule.Name(), err)
        }
    }

    return nil
}
```

### Configuration

```yaml
# inventory/vars.yaml
validation:
  mandatory_stacks:
    - traefik
    - authentik
  require_secrets:
    - authentik
    - nextcloud
  max_stacks: 50
```

## 4. Pipeline Stages

See [Pipeline Details](pipeline.md#adding-new-stages) for comprehensive guide.

### Quick Example

```go
// internal/pipeline/stages.go
func BackupStage() Stage {
    return func(ctx *Context) error {
        backupDir := filepath.Join("runtime", "backups")
        timestamp := time.Now().Format("20060102-150405")

        // Backup previous compose file
        if fileExists("runtime/docker-compose.yml") {
            backupPath := filepath.Join(backupDir, fmt.Sprintf("docker-compose-%s.yml", timestamp))
            if err := copyFile("runtime/docker-compose.yml", backupPath); err != nil {
                return err
            }
        }

        return nil
    }
}

// Add to pipeline in cmd/generate.go
p.AddStage(pipeline.BackupStage())
p.AddStage(pipeline.RenderTemplates())
// ...
```

## 5. New Commands

### 1. Create Command Handler

```go
// cmd/backup.go
package cmd

import (
    "fmt"
    "github.com/monkeymonk/homelabctl/internal/fs"
)

func Backup(args []string) error {
    if err := fs.VerifyRepository(); err != nil {
        return err
    }

    // Implementation
    fmt.Println("Creating backup...")

    return nil
}
```

### 2. Register Command

```go
// main.go
func main() {
    // ...

    switch command {
    case "init":
        err = cmd.Init()
    case "backup":
        err = cmd.Backup(args)
    // ... other commands
    }

    // ...
}
```

### 3. Add Documentation

Update help text and documentation to include new command.

## Common Extension Patterns

### Template Functions

Add custom gomplate functions:

```go
// internal/render/functions.go
func customFunctions() template.FuncMap {
    return template.FuncMap{
        "generatePassword": func(length int) string {
            // Generate random password
            return password
        },
        "sha256": func(input string) string {
            h := sha256.New()
            h.Write([]byte(input))
            return hex.EncodeToString(h.Sum(nil))
        },
    }
}

// Update render.go to include functions
func Template(templatePath string, context interface{}, outputPath string) error {
    // Add custom functions to gomplate context
    funcMap := customFunctions()
    // ... pass to gomplate
}
```

### Stack Hooks

Execute hooks before/after stack operations:

```go
// internal/hooks/hooks.go
type Hook func(stackName string) error

var (
    PreRenderHooks  []Hook
    PostRenderHooks []Hook
)

func RegisterPreRender(hook Hook) {
    PreRenderHooks = append(PreRenderHooks, hook)
}

func ExecutePreRender(stackName string) error {
    for _, hook := range PreRenderHooks {
        if err := hook(stackName); err != nil {
            return err
        }
    }
    return nil
}

// Usage in pipeline
func RenderTemplates() Stage {
    return func(ctx *Context) error {
        for stackName := range ctx.StackConfigs {
            if err := hooks.ExecutePreRender(stackName); err != nil {
                return err
            }

            // Render template...

            if err := hooks.ExecutePostRender(stackName); err != nil {
                return err
            }
        }
        return nil
    }
}
```

### Custom Variable Sources

Load variables from external sources:

```go
// internal/inventory/sources.go
type VariableSource interface {
    Load() (map[string]interface{}, error)
}

type ConsulSource struct {
    Address string
    Prefix  string
}

func (s *ConsulSource) Load() (map[string]interface{}, error) {
    // Load from Consul KV
    return vars, nil
}

// Merge multiple sources
func LoadAll() (map[string]interface{}, error) {
    sources := []VariableSource{
        &FileSource{Path: "inventory/vars.yaml"},
        &ConsulSource{Address: "consul:8500"},
        &EnvSource{Prefix: "HOMELAB_"},
    }

    merged := make(map[string]interface{})
    for _, source := range sources {
        vars, err := source.Load()
        if err != nil {
            return nil, err
        }

        // Deep merge
        deepMerge(merged, vars)
    }

    return merged, nil
}
```

## Testing Extensions

### Unit Tests

```go
// internal/myextension/extension_test.go
func TestMyExtension(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid", "input", "output", false},
        {"invalid", "bad", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

```go
// cmd/integration_test.go
func TestExtensionIntegration(t *testing.T) {
    tmpDir := t.TempDir()
    os.Chdir(tmpDir)

    // Setup test repository
    setupTestRepo(t)

    // Test extension
    err := cmd.MyExtendedCommand([]string{"arg"})
    if err != nil {
        t.Fatalf("command failed: %v", err)
    }

    // Verify results
    // ...
}
```

## Best Practices

### 1. Fail Fast

```go
if err != nil {
    return err  // Don't continue on errors
}
```

### 2. Actionable Errors

```go
return errors.New(
    "extension failed",
    "Check configuration in inventory/vars.yaml",
    "Run: homelabctl validate",
)
```

### 3. Configuration Over Code

Prefer configuration files over hardcoded values:

```yaml
# inventory/vars.yaml
extensions:
  my_extension:
    enabled: true
    option: value
```

### 4. Preserve Determinism

Avoid:
- Timestamps (unless explicitly requested)
- Random values (unless seeded)
- Network calls (cache results)

### 5. Document Extensions

Add documentation for:
- What the extension does
- How to configure it
- Example usage
- Troubleshooting

## See Also

- [Architecture](architecture.md) - System design
- [Pipeline Details](pipeline.md) - Pipeline architecture
- [Contributing](../contributing/development.md) - Development guide
