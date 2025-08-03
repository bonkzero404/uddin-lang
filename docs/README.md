# Uddin Programming Language Documentation

This directory contains the complete documentation for Uddin Programming Language, built with Jekyll and designed to be deployed on GitHub Pages.

## 📁 Documentation Structure

```
docs/
├── index.md                    # Main documentation homepage
├── tutorial/                   # Step-by-step tutorials
│   ├── index.md               # Tutorial overview
│   └── 01-basics/             # Basic concepts
│       ├── 01-introduction.md # Getting started
│       └── 02-basic-syntax.md # Language syntax
├── reference/                  # Language reference
│   └── index.md               # Complete language reference
├── builtin-functions/          # Built-in functions documentation
│   └── index.md               # All built-in functions
├── examples/                   # Code examples
│   └── index.md               # Example programs
├── advanced/                   # Advanced topics
│   └── index.md               # Performance, patterns, best practices
├── _config.yml                 # Jekyll configuration
├── Gemfile                     # Ruby dependencies
└── README.md                   # This file
```

## 🚀 Quick Start

### Prerequisites

- Ruby 3.1 or higher
- Bundler gem
- Git

### Local Development

1. **Clone the repository:**
   ```bash
   git clone https://github.com/bonkzero404/uddin-lang.git
   cd uddin-lang/docs
   ```

2. **Install dependencies:**
   ```bash
   bundle install
   ```

3. **Serve locally:**
   ```bash
   bundle exec jekyll serve
   ```

4. **Open in browser:**
   ```
   http://localhost:4000/uddin-lang/
   ```

### Building for Production

```bash
bundle exec jekyll build
```

The built site will be in the `_site` directory.

## 🌐 Deployment

### Automatic Deployment (Recommended)

The documentation is automatically deployed to GitHub Pages using GitHub Actions when:

- Changes are pushed to the `main` or `master` branch
- Changes are made to files in the `docs/` directory
- The workflow is manually triggered

**Deployment URL:** `https://bonkzero404.github.io/uddin-lang/`

### Manual Deployment

If you need to deploy manually:

1. **Build the site:**
   ```bash
   bundle exec jekyll build
   ```

2. **Deploy to GitHub Pages:**
   - Push the `docs/` directory to your repository
   - Enable GitHub Pages in repository settings
   - Set source to "GitHub Actions"

## 📝 Contributing to Documentation

### Adding New Content

1. **Tutorial Pages:**
   - Add new `.md` files in `tutorial/` subdirectories
   - Follow the existing naming convention
   - Include front matter with title and layout

2. **Examples:**
   - Add code examples to `examples/index.md`
   - Include complete, runnable code
   - Provide clear explanations

3. **Reference Documentation:**
   - Update `reference/index.md` for language features
   - Update `builtin-functions/index.md` for new functions

### Writing Guidelines

1. **Markdown Format:**
   ```markdown
   ---
   layout: default
   title: Page Title
   ---
   
   # Page Title
   
   Content here...
   ```

2. **Code Blocks:**
   ```markdown
   ```uddin
   // Uddin code example
   func hello() {
       print("Hello, World!")
   }
   ```
   ```

3. **Internal Links:**
   ```markdown
   [Link Text](../other-page/)
   [Tutorial](../tutorial/)
   ```

4. **External Links:**
   ```markdown
   [GitHub Repository](https://github.com/tmc/uddin-lang)
   ```

### Content Standards

- **Clear and Concise:** Write for developers of all skill levels
- **Complete Examples:** Provide working code that users can run
- **Consistent Style:** Follow the established documentation style
- **Up-to-date:** Ensure examples work with the current version

## 🔧 Configuration

### Jekyll Configuration (`_config.yml`)

Key settings:

- **Site Information:** Title, description, URL
- **Navigation:** Header pages and menu structure
- **Plugins:** Jekyll plugins for enhanced functionality
- **Syntax Highlighting:** Rouge with GitHub theme
- **Collections:** Organized content grouping

### GitHub Actions Workflow

The deployment workflow (`.github/workflows/deploy-docs.yml`):

- **Triggers:** Push to main/master, manual dispatch
- **Build:** Uses Ruby 3.1 and Jekyll
- **Deploy:** Automatic deployment to GitHub Pages
- **Permissions:** Configured for Pages deployment

## 🎨 Customization

### Theme Customization

The documentation uses the Minima theme. To customize:

1. **Override layouts:** Create files in `_layouts/`
2. **Custom CSS:** Add styles in `_sass/`
3. **Custom includes:** Add reusable components in `_includes/`

### Adding New Sections

1. **Create directory:** `mkdir new-section`
2. **Add index file:** `touch new-section/index.md`
3. **Update navigation:** Add to `_config.yml` header_pages
4. **Link from homepage:** Update `index.md`

## 🐛 Troubleshooting

### Common Issues

1. **Bundle install fails:**
   ```bash
   gem install bundler
   bundle install
   ```

2. **Jekyll serve fails:**
   ```bash
   bundle update
   bundle exec jekyll serve --trace
   ```

3. **GitHub Pages build fails:**
   - Check the Actions tab for error details
   - Ensure all files are properly formatted
   - Verify Gemfile dependencies

4. **Links not working:**
   - Use relative paths: `../other-page/`
   - Ensure proper directory structure
   - Check for typos in filenames

### Local Development Tips

- **Live reload:** Jekyll automatically rebuilds on file changes
- **Draft posts:** Use `--drafts` flag to include draft content
- **Incremental builds:** Use `--incremental` for faster rebuilds
- **Verbose output:** Use `--verbose` for detailed build information

## 📚 Resources

- **Jekyll Documentation:** [https://jekyllrb.com/docs/](https://jekyllrb.com/docs/)
- **GitHub Pages:** [https://pages.github.com/](https://pages.github.com/)
- **Markdown Guide:** [https://www.markdownguide.org/](https://www.markdownguide.org/)
- **Kramdown Syntax:** [https://kramdown.gettalong.org/syntax.html](https://kramdown.gettalong.org/syntax.html)

## 📄 License

This documentation is part of the Uddin Programming Language project and is licensed under the same terms as the main project.

---

**Need help?** Open an issue in the [main repository](https://github.com/bonkzero404/uddin-lang/issues) with the `documentation` label.