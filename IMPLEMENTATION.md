# homelabctl Implementation Guide

## Overview

This is the reference implementation of `homelabctl`, a CLI tool for managing homelab Docker stacks.

The tool is designed as a **compiler-style tool** that transforms static stack definitions into runtime artifacts.

## Core Philosophy

- **Deterministic**: Same input always produces same output
- **Fail-fast**: Loud errors, no silent recovery
- **No magic**: All behavior is explicit and traceable
- **Compiler-style**: Static input → generated output

## Source Code Structure

```
homelabctl/
├── main.go                    # Entry point (argument parsing only)
│
├── cmd/                       # Command implementations (orchestration only)
│   ├── init.go               # Verify repository structure
│   ├── enable.go             # Create symlink in enabled/
│   ├── disable.go            # Remove symlink from enabled/
│   ├── list.go               # List enabled stacks
│   ├── validate.go           # Validate configuration
│   ├── generate.go           # Generate runtime files (CORE)
│   └── deploy.go             # Generate + docker compose
│
└── internal/                  # All business logic
    ├── fs/                    # Filesystem operations
    │   └── fs.go             # Repository verification, symlink management
    │
    ├── stacks/                # Stack management
    │   └── stacks.go         # Load stack.yaml, validate dependencies
    │
    ├── inventory/             # Inventory management
    │   └── inventory.go      # Load inventory/vars.yaml
    │
    ├── secrets/               # Secrets management
    │   └── secrets.go        # Load secrets (SOPS integration point)
    │
    ├── render/                # Template rendering
    │   └── render.go         # Gomplate wrapper
    │
    └── compose/               # Docker Compose operations
        └── compose.go        # Merge compose files
```

## Generation Flow (The Heart of homelabctl)

The `generate` command implements the **mandatory generation algorithm**:

### Step-by-Step

```
1. Verify repository structure
   ↓
2. Load enabled stacks (read enabled/ symlinks)
   ↓
3. Validate dependencies (fail if unsatisfied)
   ↓
4. Load inventory/vars.yaml
   ↓
5. For each enabled stack:
   ├─ Load stack.yaml
   ├─ Load secrets/<stack>.yaml (optional)
   ├─ Build merged vars (precedence: stack < inventory < secrets)
   ├─ Build gomplate context (.vars, .stack, .stacks)
   ├─ Render compose.yml.tmpl → runtime/<stack>-compose.yml
   └─ Render contribute/traefik/*.tmpl → runtime/traefik/dynamic/
   ↓
6. Merge all *-compose.yml files into runtime/docker-compose.yml
   ↓
7. Clean up temporary files
   ↓
8. Done (atomic write)
```

### Context Structure

Every template receives:

```yaml
.vars:
  # Merged from:
  # 1. stack.yaml → vars (lowest priority)
  # 2. inventory/vars.yaml
  # 3. secrets/<stack>.yaml (highest priority)

.stack:
  name: <stack-name>
  category: <category>

.stacks:
  enabled: [list, of, enabled, stacks]
```

## Variable Precedence (CRITICAL)

Variables are merged in strict order:

```
Stack defaults (stack.yaml → vars)
  ↓ (overridden by)
Inventory vars (inventory/vars.yaml)
  ↓ (overridden by)
Secrets (secrets/<stack>.yaml)
```

This order is **mandatory** and ensures:
- Stacks provide sane defaults
- Operators can override via inventory
- Secrets take absolute priority

## Key Design Decisions

### 1. No CLI Framework

We use simple `switch` in `main.go` instead of frameworks like Cobra.

**Why**: Reduces dependencies, keeps the tool simple and auditable.

### 2. Gomplate as External Binary

We shell out to `gomplate` instead of using a Go template library.

**Why**:
- Gomplate has rich functions and datasource support
- Same tool can be used standalone for debugging
- Clear separation: homelabctl orchestrates, gomplate renders

### 3. Temporary Compose Files

Each stack's compose.yml.tmpl is rendered to `runtime/<stack>-compose.yml`, then merged.

**Why**:
- Allows inspection of per-stack output during debugging
- Merge logic is simple and explicit
- Easy to parallelize in future

### 4. Secrets Are Optional

Stacks work without secrets files.

**Why**:
- Not all stacks need secrets
- Simplifies initial setup
- Production deployments add secrets incrementally

## Extension Points

### SOPS Integration

Location: `internal/secrets/secrets.go`

SOPS decryption is automatically handled:
- Files ending in `.enc.yaml` are decrypted using `sops -d`
- Plain `.yaml` files are read directly
- Both formats are supported simultaneously

The implementation checks for SOPS availability and provides clear error messages if not found.

### Additional Contribute Patterns

Location: `cmd/generate.go` (around line 90)

Current implementation handles `contribute/traefik/`. To add more:

```go
// Render Nginx contributions
contributeDir := filepath.Join("stacks", stackName, "contribute", "nginx")
if _, err := os.Stat(contributeDir); err == nil {
    // Render templates to runtime/nginx/conf.d/
}
```

### Validation Rules

Location: `cmd/validate.go`

Add custom checks:

```go
// Check mandatory stacks
mandatoryStacks := []string{"traefik", "authentik"}
for _, required := range mandatoryStacks {
    if !enabled[required] {
        return fmt.Errorf("mandatory stack not enabled: %s", required)
    }
}
```

## Error Handling Strategy

Every function returns `error`.

Commands in `cmd/` return errors to `main()`, which:
- Prints to stderr
- Exits with code 1

**No exceptions. No silent failures.**

## Testing Strategy (Future)

Recommended test structure:

```
homelabctl/
├── testdata/
│   ├── valid-repo/
│   ├── missing-stack/
│   └── broken-dependency/
│
└── internal/
    ├── fs/fs_test.go
    ├── stacks/stacks_test.go
    └── ...
```

Use table-driven tests for validation logic.

## Performance Considerations

Current implementation is **serial** by design:

- Stacks processed one at a time
- Templates rendered sequentially

This is intentional for v1:
- Simpler to debug
- Easier to understand
- Performance is acceptable for single-node homelabs

**Future optimization**: Render stacks in parallel using goroutines.

## Docker Compose Integration

The `deploy` command runs:

```bash
docker compose -f runtime/docker-compose.yml up -d
```

This is intentionally **not abstracted**:
- Users can run it manually
- Easy to add flags (e.g., `--pull`, `--force-recreate`)
- Clear what's happening

## Debugging Tips

### 1. Inspect Context

Modify `internal/render/render.go` to print context:

```go
contextData, _ := yaml.Marshal(context)
fmt.Println(string(contextData))
```

### 2. Test Gomplate Separately

```bash
gomplate -f stacks/traefik/compose.yml.tmpl -c .=/tmp/context.yaml
```

### 3. Check Temporary Files

Before merging, temporary files are at:

```
runtime/traefik-compose.yml
runtime/authentik-compose.yml
```

Modify `cmd/generate.go` to skip cleanup for debugging.

### 4. Validate Repository Manually

```bash
homelabctl validate -v  # (add verbose flag in future)
```

## Future Enhancements

### Planned

- [ ] `homelabctl scaffold <stack>` - Create new stack boilerplate
- [ ] `homelabctl diff` - Show what will change
- [ ] `homelabctl doctor` - Advanced diagnostics
- [ ] SOPS integration
- [ ] Parallel rendering

### Out of Scope

- Multi-node support (use Kubernetes instead)
- Service orchestration beyond Docker Compose
- Runtime state management (Docker Compose handles this)

## Summary

homelabctl is intentionally **small, simple, and deterministic**.

It does **one thing well**: transform static stack definitions into runtime Docker Compose files.

All complexity lives in:
- Stack definitions (user-controlled)
- Templates (user-controlled)
- Docker Compose (external tool)

The CLI itself is just orchestration.
