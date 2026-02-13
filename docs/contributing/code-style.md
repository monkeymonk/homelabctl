# Code Style

Coding standards and best practices for homelabctl development.

## General Principles

### 1. Fail Fast

Return errors immediately, never silently:

```go
// Good
if err != nil {
    return err
}

// Bad
if err != nil {
    log.Printf("Warning: %v\n", err)
    // Continue anyway
}
```

### 2. No Magic

All behavior must be explicit:

```go
// Good
func Enable(stackName string) error {
    return fs.CreateSymlink(stackName)
}

// Bad - implicit behavior
func Enable(stackName string) error {
    // Automatically enables dependencies?
    // Modifies config files?
    // Not clear from signature
}
```

### 3. Deterministic

Same input always produces same output:

```go
// Good
func Generate() error {
    // Reads from stacks/, inventory/, secrets/
    // Always produces same runtime/docker-compose.yml
}

// Bad
func Generate() error {
    // Adds timestamp to output
    // Uses random values
    // Depends on network state
}
```

### 4. Simple Over Clever

```go
// Good
for _, stack := range stacks {
    if err := processStack(stack); err != nil {
        return err
    }
}

// Bad - clever but hard to debug
processStacks := func(s []string) error { /* ... */ }
if err := compose(map, filter, reduce)(stacks, processStacks); err != nil {
    return err
}
```

## Go Style

Follow standard Go conventions:

### Formatting

```bash
# Format all code
go fmt ./...
```

Always run `go fmt` before committing.

### Imports

Group imports in three sections:

```go
import (
    // Standard library
    "fmt"
    "os"
    "path/filepath"

    // External dependencies
    "gopkg.in/yaml.v3"

    // Internal packages
    "github.com/monkeymonk/homelabctl/internal/errors"
    "github.com/monkeymonk/homelabctl/internal/stacks"
)
```

### Naming

**Packages:**
- Lowercase, single word
- Descriptive: `stacks`, `compose`, `render`
- Not: `stack_manager`, `composeFiles`

**Functions:**
- CamelCase for public: `LoadStacks()`, `ValidateDependencies()`
- camelCase for private: `loadStack()`, `validateDeps()`
- Verbs: `Load`, `Validate`, `Create`, `Delete`

**Variables:**
- Short in small scopes: `err`, `i`, `s`
- Descriptive in larger scopes: `stackName`, `enabledStacks`
- No Hungarian notation: `strName` ❌, `name` ✅

**Constants:**
- CamelCase: `DefaultPort`, `MaxStacks`
- Not: `DEFAULT_PORT`, `MAX_STACKS`

### Comments

**Package comments:**
```go
// Package stacks provides stack loading and validation.
package stacks
```

**Function comments:**
```go
// Load reads and parses a stack.yaml file.
// Returns error if file doesn't exist or is invalid.
func Load(stackName string) (*Stack, error) {
    // ...
}
```

**Inline comments:**
```go
// Good - explain why
// Skip cleanup in debug mode to preserve temporary files
if debug {
    return nil
}

// Bad - explain what (code already says this)
// Return nil
return nil
```

## Error Handling

### Always Use Enhanced Errors

```go
import "github.com/monkeymonk/homelabctl/internal/errors"

// Good
return errors.New(
    "stack 'foo' does not exist",
    "Run: homelabctl list",
    "Check stacks/ directory",
)

// Bad
return fmt.Errorf("stack does not exist")
```

### Wrap Errors with Context

```go
// Good
if err := validateDeps(stack); err != nil {
    return errors.Wrap(err,
        "dependency validation failed",
        "Check stack.yaml requires field",
    )
}

// Bad
if err := validateDeps(stack); err != nil {
    return err
}
```

### Use Error Helpers

```go
// File not found
return errors.FileNotFound(path, "Run: homelabctl init")

// Invalid YAML
return errors.InvalidYAML(path, parseErr)

// Dependency cycle
return errors.DependencyCycle(cycle)
```

## Function Design

### Keep Functions Small

```go
// Good - single responsibility
func Load(stackName string) (*Stack, error) {
    path := paths.StackYAML(stackName)
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var stack Stack
    if err := yaml.Unmarshal(data, &stack); err != nil {
        return nil, err
    }

    return &stack, nil
}

// Bad - multiple responsibilities
func LoadAndValidateAndProcess(stackName string) error {
    // 100 lines of mixed concerns
}
```

### Single Return Type

```go
// Good
func Load(name string) (*Stack, error) {
    // ...
}

// Acceptable for boolean checks
func Exists(name string) bool {
    // ...
}

// Bad
func Load(name string) (*Stack, error, bool) {
    // Too many return values
}
```

### Avoid Global State

```go
// Bad
var currentStacks []string

func SetStacks(stacks []string) {
    currentStacks = stacks
}

func GetStacks() []string {
    return currentStacks
}

// Good - pass context explicitly
type Context struct {
    Stacks []string
}

func Process(ctx *Context) error {
    // Use ctx.Stacks
}
```

## Testing

### Table-Driven Tests

```go
func TestValidateDependencies(t *testing.T) {
    tests := []struct {
        name    string
        enabled []string
        wantErr bool
    }{
        {"no dependencies", []string{"standalone"}, false},
        {"satisfied", []string{"core", "app"}, false},
        {"unsatisfied", []string{"app"}, true},
        {"circular", []string{"a", "b"}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateDependencies(tt.enabled)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Use Temporary Directories

```go
func TestGenerate(t *testing.T) {
    tmpDir := t.TempDir()  // Automatically cleaned up
    if err := os.Chdir(tmpDir); err != nil {
        t.Fatal(err)
    }

    // Setup test repository
    setupTestRepo(t, tmpDir)

    // Run test
    err := Generate()
    if err != nil {
        t.Fatalf("Generate() error = %v", err)
    }

    // Verify output
    if !fileExists("runtime/docker-compose.yml") {
        t.Error("output file not created")
    }
}
```

### Helper Functions

```go
// Prefix with "setup" or "create"
func setupTestRepo(t *testing.T, dir string) {
    t.Helper()

    dirs := []string{"stacks", "enabled", "inventory", "runtime"}
    for _, d := range dirs {
        if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
            t.Fatal(err)
        }
    }
}

// Cleanup returned as function
func setupTestStack(t *testing.T) func() {
    t.Helper()

    // Setup
    createStack(t, "test")

    // Return cleanup
    return func() {
        os.RemoveAll("stacks/test")
    }
}

// Usage
cleanup := setupTestStack(t)
defer cleanup()
```

## Code Organization

### Package Structure

```
internal/
├── fs/           # One concern: filesystem operations
├── stacks/       # One concern: stack management
├── compose/      # One concern: compose file operations
└── errors/       # One concern: error handling
```

### File Structure

```go
// mypackage.go
package mypackage

// 1. Imports
import (
    "fmt"
    "os"
)

// 2. Constants
const (
    DefaultValue = "default"
)

// 3. Types
type MyStruct struct {
    Field string
}

// 4. Constructors
func New() *MyStruct {
    return &MyStruct{}
}

// 5. Public functions (alphabetical)
func PublicA() error {
    return nil
}

func PublicB() error {
    return nil
}

// 6. Private functions (alphabetical)
func privateA() error {
    return nil
}

func privateB() error {
    return nil
}
```

### Test Files

```go
// mypackage_test.go
package mypackage

// 1. Test helpers
func setupTest(t *testing.T) {
    t.Helper()
    // ...
}

// 2. Tests (alphabetical)
func TestPublicA(t *testing.T) {
    // ...
}

func TestPublicB(t *testing.T) {
    // ...
}
```

## Command Implementation

### cmd/ Pattern

```go
// cmd/mycommand.go
package cmd

import "github.com/monkeymonk/homelabctl/internal/fs"

// Orchestration only, no business logic
func MyCommand(args []string) error {
    // 1. Verify repository
    if err := fs.VerifyRepository(); err != nil {
        return err
    }

    // 2. Parse arguments
    if len(args) < 1 {
        return errors.MissingArgument("name", "mycommand")
    }
    name := args[0]

    // 3. Call internal packages
    return internalPackage.DoWork(name)
}
```

**Rules:**
- No business logic in `cmd/`
- Only orchestration and argument parsing
- Call `internal/` packages for implementation
- Return errors to main()

## Pipeline Stages

### Stage Pattern

```go
func MyStage() Stage {
    return func(ctx *Context) error {
        // 1. Check preconditions
        if len(ctx.EnabledStacks) == 0 {
            return errors.New("no stacks enabled")
        }

        // 2. Do work
        for _, stack := range ctx.EnabledStacks {
            if err := processStack(stack); err != nil {
                return err
            }
        }

        // 3. Update context
        ctx.ProcessedStacks = append(ctx.ProcessedStacks, ...)

        return nil
    }
}
```

## Performance

### Prefer Simplicity

```go
// Good - simple and clear
for _, stack := range stacks {
    if err := process(stack); err != nil {
        return err
    }
}

// Premature optimization
var wg sync.WaitGroup
errCh := make(chan error, len(stacks))
for _, stack := range stacks {
    wg.Add(1)
    go func(s string) {
        defer wg.Done()
        if err := process(s); err != nil {
            errCh <- err
        }
    }(stack)
}
// ... complex error handling
```

Only optimize when:
- Profiling shows bottleneck
- Significant performance impact
- Complexity is justified

## Documentation

### Public API

Every public function, type, and constant needs a comment:

```go
// Stack represents a homelab stack definition.
type Stack struct {
    Name     string
    Category string
}

// Load reads and parses a stack.yaml file.
// Returns error if file doesn't exist or is invalid.
func Load(stackName string) (*Stack, error) {
    // ...
}
```

### Complex Logic

Explain **why**, not **what**:

```go
// Good
// Category defaults have lowest priority to allow stack-specific overrides
merged := merge(categoryDefaults, stackDefaults, inventoryVars, secrets)

// Bad
// Merge all the variables together
merged := merge(categoryDefaults, stackDefaults, inventoryVars, secrets)
```

## Common Patterns

### File Existence

```go
func fileExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}
```

### Directory Creation

```go
if err := os.MkdirAll(dir, 0755); err != nil {
    return err
}
```

### File Reading

```go
data, err := os.ReadFile(path)
if err != nil {
    return errors.FileNotFound(path, "Check path")
}
```

### YAML Parsing

```go
var config Config
if err := yaml.Unmarshal(data, &config); err != nil {
    return errors.InvalidYAML(path, err)
}
```

## Anti-Patterns

### Don't Panic

```go
// Bad
if err != nil {
    panic(err)
}

// Good
if err != nil {
    return err
}
```

### Don't Ignore Errors

```go
// Bad
data, _ := os.ReadFile(path)

// Good
data, err := os.ReadFile(path)
if err != nil {
    return err
}
```

### Don't Use `init()`

```go
// Bad
func init() {
    // Global state modification
}

// Good - explicit initialization
func Setup() error {
    // ...
}
```

## See Also

- [Development Setup](development.md) - Getting started
- [Architecture](../advanced/architecture.md) - System design
- [How to Contribute](how-to-contribute.md) - Contribution workflow
