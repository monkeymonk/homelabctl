# Architecture

homelabctl is designed as a **compiler** that transforms declarative stack definitions into runtime Docker Compose files. This document explains the design philosophy and code structure.

## Core Philosophy

### 1. Deterministic

Same input always produces same output:

```bash
# Same stacks + same inventory + same secrets = identical output
homelabctl generate  # Run 1
homelabctl generate  # Run 2 → Identical docker-compose.yml
```

No timestamps, random values, or state-dependent behavior.

### 2. Fail-Fast

Loud errors, no silent recovery:

```go
if err != nil {
    return err  // Exit code 1
}
```

**Never:**
- Silently skip errors
- Provide default fallbacks for missing data
- Continue on failure

**Always:**
- Return errors immediately
- Exit with code 1 on any error
- Print actionable error messages

### 3. No Magic

All behavior is explicit and traceable:

- **No** auto-discovery of undefined behavior
- **No** implicit configuration
- **No** hidden state
- Everything is defined in YAML or code

### 4. Compiler-Style

Static input → generated output:

```
stacks/ + inventory/ + secrets/ → runtime/docker-compose.yml
   ↓          ↓          ↓              ↑
 (static)  (static)  (static)      (generated)
```

No runtime state, no database, no API calls (except to Docker).

### 5. Filesystem is Source of Truth

Enabled services represented by symlinks:

```
enabled/
├── traefik -> ../stacks/traefik
└── authentik -> ../stacks/authentik
```

Not variables, not config files, not state.

## Code Structure

```
main.go           # Entry point - simple switch/case
cmd/              # Command implementations (orchestration only)
internal/         # All business logic
```

### Design Decisions

**1. No CLI Framework**

```go
// main.go
switch command {
case "init":
    return cmd.Init()
case "enable":
    return cmd.Enable(args)
// ...
}
```

**Why:** Simple, auditable, no magic. No hidden flags, no auto-generated help text that diverges from docs.

**2. Gomplate as External Binary**

```go
// internal/render/render.go
cmd := exec.Command("gomplate", "-f", template, "-c", context)
```

**Why:**
- Don't reinvent templating
- Gomplate is battle-tested
- Users can test templates standalone
- Clear separation of concerns

**3. Temporary Compose Files**

```
runtime/
├── traefik-compose.yml      # Temporary
├── authentik-compose.yml    # Temporary
└── docker-compose.yml       # Final merged output
```

**Why:**
- Each stack renders independently
- Easy to debug individual stacks
- Parallel rendering possible (future)
- Clean separation before merging

**4. Secrets are Optional**

```go
// Secrets may or may not exist - that's OK
secrets, err := secrets.Load(stackName)
if err != nil && !os.IsNotExist(err) {
    return err  // Only fail on read errors, not missing files
}
```

**Why:** Stacks should work without secrets for development/testing.

**5. Pipeline Pattern**

```go
p := pipeline.New()
p.AddStage(LoadStacks())
p.AddStage(LoadInventory())
p.AddStage(RenderTemplates())
// ...
return p.Execute()
```

**Why:** See [Pipeline Details](pipeline.md) for full explanation.

**6. Docker Compose Passthrough**

```go
// main.go
default:
    // Unknown command → pass to docker compose
    return cmd.ComposePassthrough(os.Args[1:])
```

**Why:** homelabctl acts as complete Docker Compose wrapper.

## Module Organization

### cmd/ - Command Orchestration

**Responsibilities:**
- Parse command-line arguments
- Call internal packages in correct order
- Handle errors and output
- **NO business logic**

**Example:**
```go
func Enable(args []string) error {
    stackName := args[0]

    if err := fs.VerifyRepository(); err != nil {
        return err
    }

    return fs.CreateSymlink(stackName)
}
```

### internal/ - Business Logic

All implementation details:

#### internal/fs - Filesystem Operations

```go
// Verify repository structure
fs.VerifyRepository()

// Manage symlinks
fs.CreateSymlink(stackName)
fs.RemoveSymlink(stackName)
fs.ListEnabled()
```

#### internal/stacks - Stack Management

```go
// Load stack definitions
stack, err := stacks.Load(stackName)

// Validate dependencies
err := stacks.ValidateDependencies(enabled)

// Detect cycles
cycle, err := stacks.DetectCycles(enabled)
```

#### internal/categories - Category System

```go
// Register category (dynamic discovery)
categories.RegisterCategory(categoryName)

// Get category metadata
meta := categories.Get(categoryName)

// Validate dependencies based on order
err := categories.ValidateDependencies(stacks)
```

#### internal/inventory - Global Configuration

```go
// Load inventory vars
vars, err := inventory.Load()

// Access variables
domain := vars.Get("domain")
```

#### internal/secrets - Secret Management

```go
// Load secrets (SOPS integration)
secrets, err := secrets.Load(stackName)

// Automatically detects .enc.yaml and decrypts
```

#### internal/render - Template Rendering

```go
// Render template with gomplate
output, err := render.Template(templatePath, context)

// Context structure:
// .vars   - merged variables
// .stack  - stack metadata
// .stacks - global info
```

#### internal/compose - Compose File Merging

```go
// Merge multiple compose files
merged, err := compose.Merge(files...)

// Services: must be unique (error on duplicate)
// Volumes: can duplicate (silent skip)
// Networks: can duplicate (silent skip)
```

#### internal/pipeline - Pipeline Pattern

```go
// Create pipeline
p := pipeline.New()

// Add stages
p.AddStage(stage1)
p.AddStage(stage2)

// Execute (stops on first error)
err := p.Execute()
```

#### internal/errors - Enhanced Errors

```go
// Create error with suggestions
return errors.New(
    "stack 'foo' does not exist",
    "Run: homelabctl list",
    "Check stacks/ directory",
)

// Wrap existing error
return errors.Wrap(err,
    "failed to load stack",
    "Check stack.yaml syntax",
)
```

#### internal/paths - Path Resolution

```go
// Get repository root
root := paths.Root()

// Build path relative to root
stackPath := paths.Stack(stackName)
```

## Data Flow

### Generate Command (Simplified)

```
1. VerifyRepository()
   ↓
2. ListEnabled() → [traefik, authentik]
   ↓
3. ValidateDependencies([traefik, authentik])
   ↓
4. LoadInventory() → vars.yaml
   ↓
5. For each stack:
   a. LoadStack(name) → stack.yaml
   b. LoadSecrets(name) → secrets/name.enc.yaml
   c. MergeVars(stack, inventory, secrets)
   d. BuildContext(.vars, .stack, .stacks)
   e. RenderTemplate(compose.yml.tmpl, context)
   ↓
6. FilterServices(disabled_services)
   ↓
7. MergeCompose([...compose files])
   ↓
8. WriteOutput(docker-compose.yml)
   ↓
9. Cleanup(temp files)
```

See [Pipeline Details](pipeline.md) for full implementation.

## Error Handling

### Principle: Fail Fast

```go
if err != nil {
    return err  // Immediately propagate
}
```

### Actionable Errors

Use enhanced errors for clarity:

```go
// Bad
return fmt.Errorf("file not found")

// Good
return errors.FileNotFound(path,
    "Initialize repository with: homelabctl init",
)
```

### Error Propagation

```go
// cmd/ returns errors to main()
func Generate() error {
    if err := verifyRepo(); err != nil {
        return err
    }
    // ...
}

// main() handles output and exit
func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

## Testing Strategy

### Unit Tests

Test individual functions in isolation:

```go
// internal/stacks/stacks_test.go
func TestValidateDependencies(t *testing.T) {
    // Test dependency validation logic
}
```

### Integration Tests

Test full workflows:

```go
// cmd/integration_test.go
func TestGenerateWorkflow(t *testing.T) {
    // Test: init → enable → generate
}
```

### Test Patterns

**Table-Driven Tests:**
```go
tests := []struct {
    name    string
    input   string
    wantErr bool
}{
    {"valid", "traefik", false},
    {"invalid", "missing", true},
}
```

**Temporary Directories:**
```go
tmpDir := t.TempDir()
os.Chdir(tmpDir)
// Test runs in isolation
```

## Performance Considerations

### Current: Serial Processing

Stacks processed one at a time:

```go
for _, stack := range enabled {
    if err := renderStack(stack); err != nil {
        return err
    }
}
```

**Why:** Simpler to implement and debug.

### Future: Parallel Processing

Independent stacks can render in parallel:

```go
// Future optimization
var wg sync.WaitGroup
for _, stack := range enabled {
    wg.Add(1)
    go func(s string) {
        defer wg.Done()
        renderStack(s)
    }(stack)
}
wg.Wait()
```

**Tradeoffs:**
- Faster for many stacks
- More complex error handling
- Harder to debug

## Extension Points

### Adding New Commands

1. Add command handler in `cmd/`:

```go
// cmd/mycommand.go
func MyCommand(args []string) error {
    // Implementation
}
```

2. Add to switch in `main.go`:

```go
case "mycommand":
    return cmd.MyCommand(args[1:])
```

### Adding Pipeline Stages

See [Pipeline Details](pipeline.md#adding-stages)

### Adding Secret Backends

See [Extension Points](extensions.md#sops-integration)

## See Also

- [Pipeline Details](pipeline.md) - Generation pipeline architecture
- [Extension Points](extensions.md) - How to extend homelabctl
- [Contributing](../contributing/development.md) - Development setup
