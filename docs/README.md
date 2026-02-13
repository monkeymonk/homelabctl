# Documentation

This directory contains the source files for the homelabctl documentation website.

## Local Development

### Prerequisites

```bash
pip install -r requirements.txt
```

### Serve Locally

```bash
mkdocs serve
```

Visit http://localhost:8000

### Build Static Site

```bash
mkdocs build
```

Output in `site/` directory.

## Deployment

Documentation is automatically deployed to GitHub Pages when changes are pushed to `main` branch.

Visit: https://monkeymonk.github.io/homelabctl/

## Structure

- `index.md` - Landing page
- `getting-started/` - Installation and quick start guides
- `guide/` - User guide and tutorials
- `advanced/` - Advanced topics and architecture
- `contributing/` - Contributing guidelines
- `reference/` - API and configuration reference
