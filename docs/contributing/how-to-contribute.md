# How to Contribute

Thank you for your interest in contributing! This document provides guidelines and instructions.

## Code of Conduct

Be respectful, constructive, and professional in all interactions.

## How to Contribute

### Reporting Bugs

1. **Check existing issues** to avoid duplicates
2. **Use the bug report template** with:
   - homelabctl version (`homelabctl version`)
   - Go version (`go version`)
   - Operating system
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant logs or error messages

### Suggesting Features

1. **Check existing feature requests** first
2. **Open a discussion** to gather feedback before implementing
3. Describe:
   - Use case and problem being solved
   - Proposed solution
   - Alternative approaches considered
   - Impact on existing functionality

### Contributing Code

#### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/monkeymonk/homelabctl.git
cd homelabctl

# Install dependencies
go mod download

# Build
go build -o homelabctl

# Run tests
go test ./...
```

See [Development Setup](development.md) for detailed instructions.

#### Development Workflow

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Follow [Code Style](code-style.md) guidelines
   - Add tests for new functionality
   - Update documentation as needed

4. **Test your changes**
   ```bash
   go test ./...
   go build -o homelabctl
   ./homelabctl --help
   ```

5. **Commit with clear messages**
   ```bash
   git commit -m "feat: add support for X"
   git commit -m "fix: resolve issue with Y"
   ```

   Use conventional commits:
   - `feat:` New feature
   - `fix:` Bug fix
   - `docs:` Documentation only
   - `refactor:` Code refactoring
   - `test:` Adding tests
   - `chore:` Maintenance tasks

6. **Push and create pull request**
   ```bash
   git push origin feature/your-feature-name
   ```

## Code Style Guidelines

### Go Code Style

Follow standard Go conventions:

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run
```

**Key principles:**
- Simple, readable code over clever code
- Clear variable names (no abbreviations like `cfg`, use `config`)
- Early returns for error cases
- Fail-fast - don't recover from errors silently
- Use meaningful comments for complex logic only

See [Code Style](code-style.md) for detailed guidelines.

### Error Handling

Use the enhanced error package:

```go
// Good: Enhanced error with suggestions
if !stackExists {
    return errors.New(
        "stack 'foo' does not exist",
        "Run: homelabctl list",
        "Check stacks/ directory",
    )
}

// Bad: Plain error
return fmt.Errorf("stack does not exist")
```

### Testing

- Write tests for all new functionality
- Use table-driven tests where appropriate
- Keep tests focused and readable
- Use `testdata/` for test fixtures

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "foo", "bar", false},
        {"invalid input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Feature(tt.input)
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

### Documentation

- Update README.md for user-facing changes
- Update documentation in `docs/` directory
- Add inline comments for complex logic
- Update CHANGELOG.md following Keep a Changelog format

## Project Structure

```
homelabctl/
├── main.go              # Entry point - argument parsing only
├── cmd/                 # Command implementations (orchestration)
├── internal/            # All business logic
│   ├── fs/              # Filesystem operations
│   ├── stacks/          # Stack management
│   ├── inventory/       # Inventory vars loading
│   ├── secrets/         # Secrets management
│   ├── render/          # Template rendering (gomplate)
│   ├── compose/         # Docker Compose operations
│   ├── errors/          # Enhanced error handling
│   └── pipeline/        # Pipeline pattern for generation
└── testdata/            # Test fixtures
```

### Design Principles

1. **No CLI frameworks** - Keep it simple with switch/case
2. **Fail-fast** - Loud errors, no silent recovery
3. **Deterministic** - Same input = same output
4. **Commands are thin** - All logic in `internal/`
5. **Pipeline pattern** - Modular, testable stages
6. **External tools** - Shell out to gomplate, docker compose

See [Architecture](../advanced/architecture.md) for detailed design documentation.

## Commit Message Guidelines

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**

```
feat(render): add support for nested template contexts

Allow templates to reference nested variables using dot notation.
This enables more flexible template structures.

Closes #123
```

```
fix(validate): resolve circular dependency detection

The cycle detection algorithm was missing self-referencing stacks.
Added test case and fixed the DFS implementation.

Fixes #456
```

## Pull Request Process

1. **Ensure all tests pass**
2. **Update documentation** if needed
3. **Add changelog entry** in CHANGELOG.md (Unreleased section)
4. **Describe your changes** clearly in the PR description
5. **Link related issues** using keywords (Fixes #123, Closes #456)
6. **Request review** from maintainers

### PR Title Format

Use conventional commit format:
```
feat: add XYZ feature
fix: resolve issue with ABC
docs: update installation guide
```

### Review Criteria

PRs are reviewed for:
- **Correctness**: Does it work as intended?
- **Testing**: Are there adequate tests?
- **Code quality**: Is it readable and maintainable?
- **Documentation**: Are changes documented?
- **Compatibility**: Does it break existing functionality?

## Release Process

Maintainers handle releases:

1. Update CHANGELOG.md
2. Tag release: `git tag v0.x.0`
3. Push tag: `git push origin v0.x.0`
4. GitHub Actions builds and publishes binaries
5. Update package managers (Homebrew, APT)

## Getting Help

- **Questions**: Open a [Discussion](https://github.com/monkeymonk/homelabctl/discussions)
- **Bugs**: Open an [Issue](https://github.com/monkeymonk/homelabctl/issues)
- **Documentation**: https://monkeymonk.github.io/homelabctl/

## Recognition

Contributors are recognized in:
- README.md contributors section
- Release notes
- Git commit history
- CHANGELOG.md

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## See Also

- [Development Setup](development.md) - Environment setup guide
- [Code Style](code-style.md) - Detailed coding standards
- [Architecture](../advanced/architecture.md) - System design overview
