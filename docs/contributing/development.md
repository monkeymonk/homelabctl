# Development Setup

Guide for contributing to homelabctl development.

## Prerequisites

### Required

- **Go 1.21+**
  ```bash
  # Check version
  go version

  # Install from https://go.dev/dl/
  ```

- **Git**
  ```bash
  git --version
  ```

- **Docker & Docker Compose**
  ```bash
  docker --version
  docker compose version
  ```

### Optional (for testing)

- **gomplate** (for template testing)
  ```bash
  # Install
  go install github.com/hairyhenderson/gomplate/v3/cmd/gomplate@latest

  # Verify
  gomplate --version
  ```

- **SOPS** (for secrets testing)
  ```bash
  # Linux
  wget https://github.com/getsops/sops/releases/latest/download/sops-linux-amd64
  chmod +x sops-linux-amd64
  sudo mv sops-linux-amd64 /usr/local/bin/sops

  # macOS
  brew install sops

  # Verify
  sops --version
  ```

## Getting Started

### 1. Clone Repository

```bash
git clone https://github.com/monkeymonk/homelabctl.git
cd homelabctl
```

### 2. Install Dependencies

```bash
# Download Go modules
go mod download

# Verify dependencies
go mod verify
```

### 3. Build

```bash
# Build binary
go build -o homelabctl

# Run from source
go run . --help
```

### 4. Install Locally

```bash
# Install to $GOPATH/bin
go install

# Verify installation
homelabctl --version
```

## Development Workflow

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/stacks

# Run specific test
go test ./internal/stacks -run TestValidateDependencies

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Check for common issues
go vet ./...
```

### Local Testing

Create test repository:

```bash
# Create test directory
mkdir -p /tmp/test-homelab
cd /tmp/test-homelab

# Initialize
/path/to/homelabctl init

# Create test stack
mkdir -p stacks/test
cat > stacks/test/stack.yaml <<EOF
name: test
category: tools
services:
  - test
vars:
  test:
    image: nginx:latest
    port: 8080
EOF

cat > stacks/test/compose.yml.tmpl <<EOF
services:
  test:
    image: {{ .vars.test.image }}
    container_name: test
    ports:
      - "{{ .vars.test.port }}:80"
EOF

# Enable and test
/path/to/homelabctl enable test
/path/to/homelabctl generate
/path/to/homelabctl deploy
```

## Project Structure

```
homelabctl/
├── main.go           # Entry point
├── cmd/              # Command implementations
│   ├── init.go
│   ├── enable.go
│   ├── generate.go
│   └── ...
├── internal/         # Internal packages
│   ├── fs/           # Filesystem operations
│   ├── stacks/       # Stack management
│   ├── categories/   # Category system
│   ├── inventory/    # Inventory loading
│   ├── secrets/      # Secret management
│   ├── render/       # Template rendering
│   ├── compose/      # Compose file merging
│   ├── pipeline/     # Pipeline pattern
│   ├── errors/       # Enhanced errors
│   ├── paths/        # Path utilities
│   └── testutil/     # Test utilities
├── docs/             # Documentation
├── testdata/         # Test fixtures
├── go.mod            # Go module definition
└── go.sum            # Dependency checksums
```

## Coding Standards

See [Code Style](code-style.md) for detailed guidelines.

### Quick Reference

**File Organization:**
```go
package mypackage

import (
    "fmt"
    "os"

    "github.com/monkeymonk/homelabctl/internal/errors"
)

// Public functions
func PublicFunction() error {
    return privateFunction()
}

// Private functions
func privateFunction() error {
    return nil
}
```

**Error Handling:**
```go
// Good
if err != nil {
    return errors.Wrap(err, "operation failed",
        "Check configuration",
        "Run: homelabctl validate",
    )
}

// Bad
if err != nil {
    return fmt.Errorf("error: %v", err)
}
```

**Testing:**
```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid", "input", "output", false},
        {"invalid", "", "", true},
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

## Making Changes

### 1. Create Branch

```bash
git checkout -b feature/my-feature
```

### 2. Make Changes

Follow [Code Style](code-style.md) guidelines.

### 3. Write Tests

```go
// internal/mypackage/myfile_test.go
func TestMyFunction(t *testing.T) {
    // Test implementation
}
```

### 4. Run Tests

```bash
go test ./...
```

### 5. Commit Changes

```bash
git add .
git commit -m "Add feature: description"
```

### 6. Push and Create PR

```bash
git push origin feature/my-feature
```

Open pull request on GitHub.

## Common Development Tasks

### Adding a New Command

1. **Create command handler:**

```go
// cmd/mycommand.go
package cmd

func MyCommand(args []string) error {
    // Implementation
    return nil
}
```

2. **Register in main.go:**

```go
// main.go
switch command {
case "mycommand":
    err = cmd.MyCommand(args)
// ...
}
```

3. **Add tests:**

```go
// cmd/mycommand_test.go
func TestMyCommand(t *testing.T) {
    // Test cases
}
```

4. **Update documentation:**

- Add to `docs/guide/commands.md`
- Update `docs/reference/cli.md`

### Adding a Pipeline Stage

See [Pipeline Details](../advanced/pipeline.md#adding-new-stages).

### Adding Validation Rule

```go
// internal/validation/rules.go
type MyRule struct {
    // Configuration
}

func (r *MyRule) Name() string {
    return "my-rule"
}

func (r *MyRule) Validate(ctx *Context) error {
    // Validation logic
    return nil
}
```

### Debugging

**Add logging:**

```go
import "log"

log.Printf("Debug: %+v\n", value)
```

**Use delve debugger:**

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug test
dlv test ./internal/stacks -- -test.run TestMyFunction

# Debug binary
dlv exec ./homelabctl -- generate
```

**Print intermediate values:**

```go
// Temporary debugging
fmt.Fprintf(os.Stderr, "DEBUG: value=%+v\n", value)
```

## Release Process

### Versioning

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

### Creating Release

1. **Update version:**

```go
// main.go
const version = "v1.2.3"
```

2. **Update CHANGELOG:**

```markdown
## [1.2.3] - 2025-02-13

### Added
- New feature description

### Fixed
- Bug fix description
```

3. **Create tag:**

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

4. **Build releases:**

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o homelabctl-linux-amd64

# macOS
GOOS=darwin GOARCH=amd64 go build -o homelabctl-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o homelabctl-darwin-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o homelabctl-windows-amd64.exe
```

## Troubleshooting

### Go Module Issues

```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download

# Tidy dependencies
go mod tidy
```

### Test Failures

```bash
# Run with verbose output
go test -v ./...

# Run specific test
go test -v ./internal/stacks -run TestMyFunction

# Show test coverage
go test -cover ./...
```

### Build Issues

```bash
# Clean build cache
go clean -cache

# Rebuild
go build -v
```

## Getting Help

- **Issues:** https://github.com/monkeymonk/homelabctl/issues
- **Discussions:** https://github.com/monkeymonk/homelabctl/discussions
- **Documentation:** https://monkeymonk.github.io/homelabctl/

## See Also

- [Code Style](code-style.md) - Coding standards
- [Architecture](../advanced/architecture.md) - System design
- [How to Contribute](how-to-contribute.md) - Contribution guidelines
