# Installation Guide

## Prerequisites

- **Go 1.21+** (for building from source)
- **Docker** with Compose plugin v2
- **[gomplate](https://docs.gomplate.ca/installing/)** (template engine)

### Install gomplate

```bash
# Linux
curl -o /usr/local/bin/gomplate -sSL https://github.com/hairyhenderson/gomplate/releases/download/v3.11.6/gomplate_linux-amd64
chmod +x /usr/local/bin/gomplate

# macOS (Homebrew)
brew install gomplate

# Verify
gomplate --version
```

## Installation Methods

### Option 1: Install from Source (Recommended)

```bash
go install github.com/monkeymonk/homelabctl@latest
```

The binary will be installed to `$GOPATH/bin/homelabctl` (usually `~/go/bin/homelabctl`).

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

Download the latest release for your platform from:
https://github.com/monkeymonk/homelabctl/releases

```bash
# Linux AMD64
curl -LO https://github.com/monkeymonk/homelabctl/releases/latest/download/homelabctl-linux-amd64.tar.gz
tar xzf homelabctl-linux-amd64.tar.gz
sudo mv homelabctl /usr/local/bin/

# macOS ARM64 (Apple Silicon)
curl -LO https://github.com/monkeymonk/homelabctl/releases/latest/download/homelabctl-darwin-arm64.tar.gz
tar xzf homelabctl-darwin-arm64.tar.gz
sudo mv homelabctl /usr/local/bin/
```

### Option 4: Package Managers

#### Homebrew (macOS/Linux)

```bash
brew tap monkeymonk/tap
brew install homelabctl
```

#### APT (Debian/Ubuntu)

```bash
# Add repository
echo "deb [trusted=yes] https://apt.monkeymonk.com/ /" | sudo tee /etc/apt/sources.list.d/homelabctl.list
sudo apt update

# Install
sudo apt install homelabctl
```

## Verify Installation

```bash
homelabctl --help
```

You should see the help output with available commands.

## Next Steps

1. **Create a homelab repository:**
   ```bash
   mkdir ~/homelab
   cd ~/homelab
   homelabctl init
   ```

2. **Add stack definitions:**
   - Clone existing stacks or create your own
   - Place in `stacks/` directory

3. **Enable and deploy:**
   ```bash
   homelabctl enable core
   homelabctl deploy
   ```

See [GUIDE.md](GUIDE.md) for complete usage documentation.

## Updating

### From source
```bash
go install github.com/monkeymonk/homelabctl@latest
```

### Homebrew
```bash
brew upgrade homelabctl
```

### APT
```bash
sudo apt update && sudo apt upgrade homelabctl
```

## Uninstalling

```bash
# If installed to /usr/local/bin
sudo rm /usr/local/bin/homelabctl

# If installed via go install
rm $(go env GOPATH)/bin/homelabctl

# If installed via homebrew
brew uninstall homelabctl

# If installed via apt
sudo apt remove homelabctl
```

## Troubleshooting

### "command not found: homelabctl"

- Ensure the binary is in your PATH
- Check installation location: `which homelabctl`
- Add to PATH: `export PATH=$PATH:/usr/local/bin`

### "gomplate: command not found"

homelabctl requires gomplate for template rendering:
```bash
curl -o /usr/local/bin/gomplate -sSL https://github.com/hairyhenderson/gomplate/releases/download/v3.11.6/gomplate_linux-amd64
chmod +x /usr/local/bin/gomplate
```

### "docker compose: command not found"

Install Docker with Compose plugin:
- https://docs.docker.com/compose/install/

### Permission denied when building

Ensure Go is properly installed:
```bash
go version
```

If not installed, visit: https://go.dev/doc/install

## Platform-Specific Notes

### Linux

- Requires Docker with Compose plugin v2
- Binary works on x86_64 and ARM64
- Tested on Ubuntu 22.04+, Debian 11+, Fedora 38+

### macOS

- Requires Docker Desktop or Colima
- Supports Intel (x86_64) and Apple Silicon (ARM64)
- Tested on macOS 12+

### Windows

- Use WSL2 (Windows Subsystem for Linux)
- Install homelabctl inside WSL2 environment
- Docker Desktop with WSL2 backend required

## Support

- **Issues**: https://github.com/monkeymonk/homelabctl/issues
- **Discussions**: https://github.com/monkeymonk/homelabctl/discussions
- **Documentation**: [GUIDE.md](GUIDE.md), [README.md](README.md)
