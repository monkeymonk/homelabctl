# GitHub Pages Setup Instructions

## Enabling GitHub Pages

1. **Push the documentation to GitHub:**

   ```bash
   git add docs/ mkdocs.yml requirements.txt .github/workflows/docs.yml
   git commit -m "docs: add GitHub Pages documentation site"
   git push origin main
   ```

2. **Enable GitHub Pages in repository settings:**

   - Go to your repository on GitHub
   - Navigate to **Settings** → **Pages**
   - Under **Source**, select **Deploy from a branch**
   - Select branch: **gh-pages**
   - Click **Save**

3. **Wait for deployment:**

   - The GitHub Action will automatically build and deploy
   - Check the **Actions** tab to monitor progress
   - Site will be available at: `https://monkeymonk.github.io/homelabctl/`

## Local Development

### First Time Setup

```bash
# Install dependencies
pip install -r requirements.txt
```

### Serve Locally

```bash
# Start development server
mkdocs serve

# Visit http://localhost:8000
```

### Build Static Site

```bash
# Build to site/ directory
mkdocs build

# Preview the built site
cd site && python -m http.server 8000
```

## Adding New Pages

1. Create markdown file in `docs/`
2. Add to navigation in `mkdocs.yml`
3. Commit and push (auto-deploys)

## Customization

Edit `mkdocs.yml` to customize:

- Site name and description
- Theme colors
- Navigation structure
- Plugins and extensions

## Troubleshooting

### Build Fails

Check the GitHub Actions logs for errors. Common issues:

- Missing dependencies in `requirements.txt`
- Invalid markdown syntax
- Broken internal links

### Site Not Updating

- Ensure workflow ran successfully (check Actions tab)
- Clear browser cache
- Check that gh-pages branch exists

## Resources

- [MkDocs Documentation](https://www.mkdocs.org/)
- [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/)
- [GitHub Pages Docs](https://docs.github.com/en/pages)
