# Installation

## Prerequisites

Before installing homelabctl, ensure you have:

- **Go 1.21+** (for building from source)
- **Docker** with Compose plugin v2
- **[gomplate](https://docs.gomplate.ca/installing/)** (template engine)

### Install gomplate

=== "Linux"

    ```bash
    curl -o /usr/local/bin/gomplate -sSL \
      https://github.com/hairyhenderson/gomplate/releases/download/v3.11.6/gomplate_linux-amd64
    chmod +x /usr/local/bin/gomplate
    ```

=== "macOS"

    ```bash
    brew install gomplate
    ```

=== "Verify"

    ```bash
    gomplate --version
    ```

## Installation Methods

### Option 1: Install from Source (Recommended)

```bash
go install github.com/monkeymonk/homelabctl@latest
```

The binary will be installed to `$GOPATH/bin/homelabctl` (usually `~/go/bin/homelabctl`).

!!! tip "Add to PATH"
    Make sure `$GOPATH/bin` is in your PATH:
    ```bash
    export PATH=$PATH:$(go env GOPATH)/bin
    ```

### Option 2: Build from Source

```bash
# Clone repository
git clone https://github.com/monkeymonk/homelabctl.git
cd homelabctl

# Build
go build -o homelabctl

# Install
sudo mv homelabctl /usr/local/bin/
```

### Option 3: Download Pre-built Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/monkeymonk/homelabctl/releases).

=== "Linux AMD64"

    ```bash
    curl -LO https://github.com/monkeymonk/homelabctl/releases/latest/download/homelabctl-linux-amd64.tar.gz
    tar xzf homelabctl-linux-amd64.tar.gz
    sudo mv homelabctl /usr/local/bin/
    ```

=== "macOS ARM64"

    ```bash
    curl -LO https://github.com/monkeymonk/homelabctl/releases/latest/download/homelabctl-darwin-arm64.tar.gz
    tar xzf homelabctl-darwin-arm64.tar.gz
    sudo mv homelabctl /usr/local/bin/
    ```

## Verify Installation

```bash
homelabctl --help
```

You should see the help output with available commands.

## Next Steps

- [Quick Start Guide](quickstart.md) - Get up and running in 5 minutes
- [Your First Stack](first-stack.md) - Create your first stack

## Updating

### From source

```bash
go install github.com/monkeymonk/homelabctl@latest
```

### Homebrew

```bash
brew upgrade homelabctl
```

## Troubleshooting

!!! warning "Command not found"
    If you get "command not found", ensure the binary is in your PATH:
    ```bash
    which homelabctl
    # Should show the path to the binary
    ```

!!! warning "gomplate not found"
    Install gomplate as shown in the prerequisites section.

!!! warning "Docker compose not found"
    Install Docker with Compose plugin: [docs.docker.com/compose/install](https://docs.docker.com/compose/install/)
