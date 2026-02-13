# Changelog

All notable changes to homelabctl will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial standalone release
- Comprehensive installation guide
- Pre-built binaries for multiple platforms
- Enhanced error messages with actionable suggestions
- Pipeline-based generation architecture
- Dynamic category discovery
- Service-level enable/disable functionality
- Complete documentation site with guides

### Changed

- Extracted from monolithic homelab repository
- Improved documentation for standalone use
- Consolidated CLI documentation
- Module path fixed for `go install` compatibility

### Fixed

- Variable precedence consistency
- Dependency validation edge cases

## [0.1.0] - 2025-01-XX

### Added

- Core CLI functionality (`init`, `enable`, `disable`, `list`, `validate`, `generate`, `deploy`)
- Gomplate-based template rendering
- Docker Compose file merging
- Dependency management and validation
- Category system with defaults
- Secrets loading with automatic SOPS decryption
- Docker Compose command passthrough
- Enhanced error handling with colors and suggestions
- Integration tests

### Documentation

- README with comprehensive usage examples
- Complete user and developer guides
- Architecture and pipeline documentation
- CLI reference and troubleshooting guides

---

## Version History

### v0.1.0 - Initial Release

- First stable release as standalone CLI tool
- Full feature parity with integrated version
- Production-ready for homelab management

---

## Upgrade Notes

### From integrated version (pre-0.1.0)

If upgrading from the integrated homelabctl within a homelab repository:

1. Install standalone version:

   ```bash
   go install github.com/monkeymonk/homelabctl@latest
   ```

2. No changes to your homelab repository structure required
3. All existing commands work identically
4. Enhanced error messages may display differently (improved formatting)

### Breaking Changes

**None** - Full backward compatibility maintained.

---

## Compatibility

- **Go**: 1.21 or higher
- **Docker**: 20.10 or higher
- **Docker Compose**: 2.0 or higher
- **gomplate**: 3.0 or higher (called as external binary)
- **SOPS**: 3.7 or higher (optional, for secrets)

---

## Upcoming Features (Roadmap)

- [ ] Interactive mode for stack selection
- [ ] Stack templates/scaffolding
- [ ] Health check integration
- [ ] Backup/restore functionality
- [ ] Remote repository support (pull stacks from git)
- [ ] Plugin system for custom commands
- [ ] Shell completions (bash, zsh, fish)
- [ ] Parallel rendering for faster generation
- [ ] Config diff viewer
- [ ] Dry-run mode for all commands

---

## Links

- [GitHub Repository](https://github.com/monkeymonk/homelabctl)
- [Documentation](https://monkeymonk.github.io/homelabctl/)
- [Issue Tracker](https://github.com/monkeymonk/homelabctl/issues)
- [Discussions](https://github.com/monkeymonk/homelabctl/discussions)

---

*For detailed version history, see [GitHub Releases](https://github.com/monkeymonk/homelabctl/releases)*
