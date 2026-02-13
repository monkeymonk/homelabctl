# Pipeline Details

The `generate` command uses a pipeline pattern for modularity, testability, and debuggability. This document explains the architecture and implementation.

## Overview

### Pipeline Pattern

```
LoadStacks → LoadInventory → MergeVariables → FilterServices →
RenderTemplates → MergeCompose → WriteOutput → Cleanup
```

Each stage:
- Operates on a shared `Context`
- Returns error on failure (stops pipeline)
- Is independently testable
- Can be skipped for debugging

### Benefits

**Modularity:**
- Each stage has single responsibility
- Easy to understand and maintain
- Can be reordered if dependencies allow

**Testability:**
- Test stages independently
- Mock context for isolated tests
- Integration tests compose stages

**Debuggability:**
- Skip cleanup to inspect temp files
- Add logging per stage
- Pause pipeline at any point

**Future-proof:**
- Parallel execution for independent stages
- Progress reporting
- Caching/memoization

## Pipeline Context

Shared state passed through all stages:

```go
type Context struct {
    // Input: Enabled stacks
    EnabledStacks []string

    // Configuration
    InventoryVars map[string]interface{}
    DisabledServices map[string]bool

    // Per-stack configuration
    StackConfigs map[string]*StackConfig

    // Rendering state
    RenderedFiles []string
    RenderedCompose map[string]string

    // Output
    MergedCompose *ComposeFile

    // Options
    Debug bool
}

type StackConfig struct {
    Stack      *Stack
    Secrets    map[string]interface{}
    MergedVars map[string]interface{}
}
```

## Pipeline Stages

### 1. LoadStacks

**Purpose:** Load enabled stacks from `enabled/` symlinks

**Input:** Context (empty)

**Output:** `Context.EnabledStacks`

**Implementation:**

```go
func LoadStacks() Stage {
    return func(ctx *Context) error {
        enabled, err := fs.ListEnabled()
        if err != nil {
            return errors.Wrap(err, "failed to list enabled stacks")
        }

        ctx.EnabledStacks = enabled
        return nil
    }
}
```

**Errors:**
- Repository not initialized
- Invalid symlinks in `enabled/`

### 2. ValidateDependencies

**Purpose:** Ensure all dependencies are satisfied

**Input:** `Context.EnabledStacks`

**Output:** None (validation only)

**Implementation:**

```go
func ValidateDependencies() Stage {
    return func(ctx *Context) error {
        // Load stack definitions
        stacks := make(map[string]*Stack)
        for _, name := range ctx.EnabledStacks {
            s, err := stacks.Load(name)
            if err != nil {
                return err
            }
            stacks[name] = s
        }

        // Validate dependencies
        return stacks.ValidateDependencies(ctx.EnabledStacks)
    }
}
```

**Errors:**
- Missing dependencies
- Circular dependencies
- Invalid category dependencies

### 3. LoadInventory

**Purpose:** Load global configuration from `inventory/vars.yaml`

**Input:** Context

**Output:** `Context.InventoryVars`, `Context.DisabledServices`

**Implementation:**

```go
func LoadInventory() Stage {
    return func(ctx *Context) error {
        vars, err := inventory.Load()
        if err != nil && !os.IsNotExist(err) {
            return errors.Wrap(err, "failed to load inventory")
        }

        ctx.InventoryVars = vars

        // Extract disabled_services list
        if disabled, ok := vars["disabled_services"].([]interface{}); ok {
            ctx.DisabledServices = make(map[string]bool)
            for _, svc := range disabled {
                ctx.DisabledServices[svc.(string)] = true
            }
        }

        return nil
    }
}
```

**Errors:**
- Invalid YAML syntax
- Malformed structure

### 4. MergeVariables

**Purpose:** Merge variables for each stack (stack defaults < inventory < secrets)

**Input:** `Context.EnabledStacks`, `Context.InventoryVars`

**Output:** `Context.StackConfigs`

**Implementation:**

```go
func MergeVariables() Stage {
    return func(ctx *Context) error {
        ctx.StackConfigs = make(map[string]*StackConfig)

        for _, stackName := range ctx.EnabledStacks {
            // Load stack definition
            stack, err := stacks.Load(stackName)
            if err != nil {
                return err
            }

            // Load secrets (optional)
            secrets, err := secrets.Load(stackName)
            if err != nil && !os.IsNotExist(err) {
                return err
            }

            // Merge: stack defaults < inventory < secrets
            merged := mergeVars(
                stack.Vars,           // Lowest priority
                ctx.InventoryVars,    // Medium priority
                secrets,              // Highest priority
            )

            ctx.StackConfigs[stackName] = &StackConfig{
                Stack:      stack,
                Secrets:    secrets,
                MergedVars: merged,
            }
        }

        return nil
    }
}
```

**Errors:**
- Failed to load stack.yaml
- Failed to decrypt secrets
- YAML parse errors

### 5. RenderTemplates

**Purpose:** Render `compose.yml.tmpl` for each stack

**Input:** `Context.StackConfigs`

**Output:** `Context.RenderedCompose`, `Context.RenderedFiles`

**Implementation:**

```go
func RenderTemplates() Stage {
    return func(ctx *Context) error {
        ctx.RenderedCompose = make(map[string]string)

        for stackName, config := range ctx.StackConfigs {
            // Build template context
            templateCtx := map[string]interface{}{
                "vars": config.MergedVars,
                "stack": map[string]string{
                    "name":     stackName,
                    "category": config.Stack.Category,
                },
                "stacks": map[string]interface{}{
                    "enabled": ctx.EnabledStacks,
                },
            }

            // Render template
            templatePath := paths.StackTemplate(stackName)
            outputPath := paths.RuntimeCompose(stackName)

            if err := render.Template(templatePath, templateCtx, outputPath); err != nil {
                return errors.Wrap(err, fmt.Sprintf("failed to render %s", stackName))
            }

            ctx.RenderedCompose[stackName] = outputPath
            ctx.RenderedFiles = append(ctx.RenderedFiles, outputPath)
        }

        return nil
    }
}
```

**Errors:**
- Template syntax errors
- Missing variables
- Gomplate execution failure

### 6. FilterServices

**Purpose:** Remove disabled services from composed files

**Input:** `Context.RenderedCompose`, `Context.DisabledServices`

**Output:** Modified compose files (in-place)

**Implementation:**

```go
func FilterServices() Stage {
    return func(ctx *Context) error {
        if len(ctx.DisabledServices) == 0 {
            return nil  // Nothing to filter
        }

        for stackName, composePath := range ctx.RenderedCompose {
            // Load compose file
            composeFile, err := compose.Load(composePath)
            if err != nil {
                return err
            }

            // Filter disabled services
            for serviceName := range composeFile.Services {
                if ctx.DisabledServices[serviceName] {
                    delete(composeFile.Services, serviceName)
                }
            }

            // Write filtered file
            if err := compose.Write(composePath, composeFile); err != nil {
                return err
            }
        }

        return nil
    }
}
```

**Errors:**
- Failed to load compose file
- Failed to write filtered file

### 7. MergeCompose

**Purpose:** Merge all stack compose files into single `docker-compose.yml`

**Input:** `Context.RenderedCompose`

**Output:** `Context.MergedCompose`

**Implementation:**

```go
func MergeCompose() Stage {
    return func(ctx *Context) error {
        var files []string
        for _, path := range ctx.RenderedCompose {
            files = append(files, path)
        }

        merged, err := compose.Merge(files...)
        if err != nil {
            return errors.Wrap(err, "failed to merge compose files")
        }

        ctx.MergedCompose = merged
        return nil
    }
}
```

**Errors:**
- Duplicate service names
- Invalid compose syntax
- Missing required fields

### 8. WriteOutput

**Purpose:** Write final `runtime/docker-compose.yml`

**Input:** `Context.MergedCompose`

**Output:** File on disk

**Implementation:**

```go
func WriteOutput() Stage {
    return func(ctx *Context) error {
        outputPath := paths.RuntimeComposeFile()

        if err := compose.Write(outputPath, ctx.MergedCompose); err != nil {
            return errors.Wrap(err, "failed to write output")
        }

        fmt.Printf("Generated: %s\n", outputPath)
        return nil
    }
}
```

**Errors:**
- Permission denied
- Disk full
- Invalid path

### 9. Cleanup

**Purpose:** Remove temporary files (unless debug mode)

**Input:** `Context.RenderedFiles`, `Context.Debug`

**Output:** Cleanup actions

**Implementation:**

```go
func Cleanup() Stage {
    return func(ctx *Context) error {
        if ctx.Debug {
            fmt.Println("Debug mode: Preserving temporary files")
            return nil
        }

        for _, file := range ctx.RenderedFiles {
            if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
                // Log but don't fail on cleanup errors
                fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", file, err)
            }
        }

        return nil
    }
}
```

**Errors:**
- None (warnings only)

## Pipeline Execution

### Basic Usage

```go
// cmd/generate.go
func Generate(debug bool) error {
    ctx := &pipeline.Context{
        Debug: debug,
    }

    p := pipeline.New(ctx)
    p.AddStage(pipeline.LoadStacks())
    p.AddStage(pipeline.ValidateDependencies())
    p.AddStage(pipeline.LoadInventory())
    p.AddStage(pipeline.MergeVariables())
    p.AddStage(pipeline.RenderTemplates())
    p.AddStage(pipeline.FilterServices())
    p.AddStage(pipeline.MergeCompose())
    p.AddStage(pipeline.WriteOutput())
    p.AddStage(pipeline.Cleanup())

    return p.Execute()
}
```

### Pipeline Implementation

```go
// internal/pipeline/pipeline.go
type Pipeline struct {
    ctx    *Context
    stages []Stage
}

type Stage func(*Context) error

func New(ctx *Context) *Pipeline {
    return &Pipeline{
        ctx:    ctx,
        stages: []Stage{},
    }
}

func (p *Pipeline) AddStage(stage Stage) {
    p.stages = append(p.stages, stage)
}

func (p *Pipeline) Execute() error {
    for i, stage := range p.stages {
        if err := stage(p.ctx); err != nil {
            return fmt.Errorf("stage %d failed: %w", i+1, err)
        }
    }
    return nil
}
```

## Debug Mode

Preserve temporary files for inspection:

```bash
homelabctl generate --debug
```

**Effect:**
- Cleanup stage skips deletion
- Temporary files remain in `runtime/`

**Inspect:**
```bash
# View per-stack rendered files
cat runtime/traefik-compose.yml
cat runtime/authentik-compose.yml

# View final merged output
cat runtime/docker-compose.yml

# Compare before/after filtering
diff runtime/traefik-compose.yml \
     runtime/docker-compose.yml
```

## Adding New Stages

### 1. Define Stage Function

```go
// internal/pipeline/stages.go
func MyCustomStage() Stage {
    return func(ctx *Context) error {
        // Access context
        for _, stackName := range ctx.EnabledStacks {
            // Do work...
        }

        // Update context
        ctx.SomeNewField = "value"

        return nil
    }
}
```

### 2. Add to Pipeline

```go
// cmd/generate.go
p.AddStage(pipeline.MyCustomStage())
```

### 3. Update Context

If stage needs new state:

```go
// internal/pipeline/context.go
type Context struct {
    // ... existing fields ...
    SomeNewField string  // New field
}
```

### 4. Test Stage

```go
// internal/pipeline/stages_test.go
func TestMyCustomStage(t *testing.T) {
    ctx := &Context{
        EnabledStacks: []string{"test"},
    }

    stage := MyCustomStage()
    err := stage(ctx)

    assert.NoError(t, err)
    assert.Equal(t, "expected", ctx.SomeNewField)
}
```

## Future Enhancements

### Parallel Execution

Independent stages can run in parallel:

```go
// Parallel rendering
var wg sync.WaitGroup
for stackName := range ctx.StackConfigs {
    wg.Add(1)
    go func(name string) {
        defer wg.Done()
        renderStack(name)
    }(stackName)
}
wg.Wait()
```

**Challenges:**
- Error aggregation
- Progress reporting
- Race conditions in context

### Progress Reporting

Show stage progress:

```go
func (p *Pipeline) Execute() error {
    for i, stage := range p.stages {
        fmt.Printf("[%d/%d] Running stage...\n", i+1, len(p.stages))
        if err := stage(p.ctx); err != nil {
            return err
        }
    }
    return nil
}
```

### Caching

Cache rendered templates if inputs unchanged:

```go
func RenderTemplates() Stage {
    return func(ctx *Context) error {
        for stackName := range ctx.StackConfigs {
            // Check cache
            if cached, ok := checkCache(stackName); ok {
                ctx.RenderedCompose[stackName] = cached
                continue
            }

            // Render and cache
            rendered := render(stackName)
            saveCache(stackName, rendered)
        }
    }
}
```

## See Also

- [Architecture](architecture.md) - Overall system design
- [Extension Points](extensions.md) - How to extend homelabctl
- [Contributing](../contributing/development.md) - Development guidelines
