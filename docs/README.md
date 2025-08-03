# UDDIN-LANG Documentation

This directory contains the complete documentation for UDDIN-LANG (Unified Dynamic Decision Interpreter Notation), built with Docusaurus 2 for a modern, fast, and responsive documentation experience.

## 🚀 About UDDIN-LANG

UDDIN-LANG is a specialized Flexible Rule Logic Platform that resembles a programming language, offering high expressiveness, full flow control, and runtime programmable capabilities for complex business decision support systems.

## 📁 Documentation Structure

```
docs/
├── src/                        # Docusaurus source files
│   ├── components/            # React components
│   ├── css/                   # Custom styles
│   └── pages/                 # Custom pages
├── docs/                      # Documentation content
│   ├── intro.md              # Getting started
│   ├── examples/             # Code examples
│   └── reference/            # Language reference
├── static/                    # Static assets
├── docusaurus.config.js      # Docusaurus configuration
├── package.json              # Node.js dependencies
└── README.md                 # This file
```

## � Quick Start

### Prerequisites

-   Node.js 16.14 or higher
-   npm or yarn package manager

### Local Development

1. **Clone the repository:**

    ```bash
    git clone https://github.com/bonkzero404/uddin-lang.git
    cd uddin-lang/docs
    ```

2. **Install dependencies:**

    ```bash
    npm install
    # or
    yarn install
    ```

3. **Start development server:**

    ```bash
    npm start
    # or
    yarn start
    ```

4. **Open in browser:**
    ```
    http://localhost:3000
    ```

### Building for Production

```bash
npm run build
# or
yarn build
```

The built site will be in the `build/` directory.

## 🌐 Deployment

The documentation can be deployed to various platforms:

### GitHub Pages

```bash
npm run deploy
# or
yarn deploy
```

### Manual Deployment

```bash
npm run build
# Then deploy the build/ directory to your hosting platform
```

## 📝 Contributing to Documentation

### Adding New Content

1. **Documentation Pages:**

    - Add new `.md` or `.mdx` files in the `docs/` directory
    - Use front matter to configure metadata

2. **Blog Posts:**

    - Add posts to `blog/` directory (if enabled)
    - Follow the naming convention: `YYYY-MM-DD-title.md`

3. **Custom Pages:**
    - Create React components in `src/pages/`
    - Use `.js`, `.jsx`, or `.mdx` extensions

### Writing Guidelines

1. **Markdown Format:**

    ```markdown
    ---
    id: page-id
    title: Page Title
    sidebar_position: 1
    ---

    # Page Title

    Content here...
    ```

2. **Code Blocks:**

    ````markdown
    ```uddin
    // UDDIN-LANG code example
    fun detectFraud(transaction):
        if (transaction.amount > 10000) then:
            return "HIGH_RISK"
        end
    end
    ```
    ````

    ```

    ```

3. **Internal Links:**
    ```markdown
    [Link Text](./other-page)
    [Tutorial](../tutorial/intro)
    ```

### Content Standards

-   **Clear and Concise:** Write for developers and business analysts
-   **Complete Examples:** Provide working UDDIN-LANG code
-   **Consistent Style:** Follow established documentation patterns
-   **Up-to-date:** Ensure examples work with current version

## 🔧 Configuration

### Docusaurus Configuration

Key settings in `docusaurus.config.js`:

-   **Site metadata:** Title, description, favicon
-   **Navigation:** Navbar and footer configuration
-   **Themes:** Classic theme with customizations
-   **Plugins:** Additional functionality
-   **Deployment:** GitHub Pages configuration

### Customization

1. **Styling:** Modify `src/css/custom.css`
2. **Components:** Create React components in `src/components/`
3. **Layout:** Override theme components using swizzling
4. **Configuration:** Update `docusaurus.config.js`

## 🎨 Theme Features

-   **Dark/Light Mode:** Built-in theme switching
-   **Search:** Integrated search functionality
-   **Mobile Responsive:** Works on all devices
-   **Fast Performance:** Optimized for speed
-   **SEO Friendly:** Built-in SEO optimizations

## 🐛 Troubleshooting

### Common Issues

1. **Install fails:**

    ```bash
    rm -rf node_modules package-lock.json
    npm install
    ```

2. **Build fails:**

    ```bash
    npm run clear
    npm run build
    ```

3. **Port already in use:**
    ```bash
    npm start -- --port 3001
    ```

### Development Tips

-   **Hot reload:** Changes are reflected immediately
-   **ESLint:** Run `npm run lint` to check code quality
-   **Clear cache:** Use `npm run clear` if experiencing issues

## 📚 Resources

-   **Docusaurus Documentation:** [https://docusaurus.io/docs](https://docusaurus.io/docs)
-   **Markdown Guide:** [https://www.markdownguide.org/](https://www.markdownguide.org/)
-   **React Documentation:** [https://reactjs.org/docs](https://reactjs.org/docs)
-   **UDDIN-LANG Repository:** [https://github.com/bonkzero404/uddin-lang](https://github.com/bonkzero404/uddin-lang)

## 📄 License

This documentation is part of the UDDIN-LANG project and is licensed under the MIT License.

---

**Need help?** Open an issue in the [main repository](https://github.com/bonkzero404/uddin-lang/issues) with the `documentation` label.
