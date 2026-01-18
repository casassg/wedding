# AGENTS.md - AI Coding Agent Instructions

This document provides guidelines for AI coding agents working on this Hugo-based wedding website.

## Project Overview

- **Type:** Static wedding website (single-page, multilingual)
- **Framework:** Hugo 0.152.2 (extended)
- **Styling:** Tailwind CSS (via CDN) + custom CSS
- **JavaScript:** Vanilla ES6 (no build step, no dependencies)
- **Deployment:** GitHub Pages via GitHub Actions
- **Languages:** English (default), Spanish, Catalan

## Build & Development Commands

```bash
# Activate Hermit environment (loads hugo and gh CLI)
source bin/activate-hermit

# Or use binaries directly
./bin/hugo [command]
./bin/gh [command]

# Development server (http://localhost:1313/wedding/)
hugo server

# Development server with drafts
hugo server -D

# Production build
hugo --gc --minify

# Clean build
rm -rf public/ && hugo --gc --minify
```

## Deployment

Deployment is automatic via GitHub Actions on push to `main`:

```bash
# Push triggers deployment
git push origin main

# Watch deployment status
./bin/gh run list --limit 1
./bin/gh run watch <run-id>
```

## Project Structure

```
├── assets/                    # Processed by Hugo (fingerprinted)
│   ├── css/custom.css        # Custom styles
│   └── js/main.js            # Vanilla JavaScript
├── content/
│   └── _index.{en,es,ca}.md  # Homepage content per language
├── i18n/
│   └── {en,es,ca}.yaml       # Translation strings
├── layouts/
│   ├── _default/baseof.html  # Base HTML template
│   ├── index.html            # Homepage template
│   └── partials/             # Reusable template components
├── config.toml               # Hugo configuration
└── bin/                      # Hermit-managed binaries
```

## Code Style Guidelines

### Hugo Templates (HTML)

- **Indentation:** 4 spaces
- **Whitespace trimming:** Use `{{- ... -}}` to control whitespace
- **i18n:** Always use `{{ i18n "key" }}` for user-facing text, never hardcode
- **Partials:** One responsibility per partial, descriptive names
- **Comments:** Use HTML comments for section headers: `<!-- Section Name -->`
- **Assets:** Use Hugo pipes for fingerprinting:
  ```html
  {{ $css := resources.Get "css/custom.css" | fingerprint }}
  <link rel="stylesheet" href="{{ $css.RelPermalink }}">
  ```

### CSS (custom.css)

- **Indentation:** 4 spaces
- **Organization:** Group by function with comment headers
- **Naming:** Descriptive, BEM-inspired (e.g., `card-shadow`, `glass-nav`, `hero-gradient`)
- **Tailwind:** Extend via config in `head.html`, use utility classes in HTML

```css
/* Section comment format */
.class-name {
    property: value;
}
```

### JavaScript (main.js)

- **Pattern:** IIFE with strict mode
- **Indentation:** 4 spaces
- **Functions:** Named functions, initialized in `DOMContentLoaded`
- **Null safety:** Use optional chaining (`?.`) and null checks
- **No dependencies:** Vanilla ES6 only

```javascript
(function() {
    'use strict';
    
    // ===================
    // Section Name
    // ===================
    function initFeature() {
        const element = document.querySelector('.selector');
        if (!element) return;
        // ...
    }
    
    document.addEventListener('DOMContentLoaded', () => {
        initFeature();
    });
})();
```

### i18n/YAML Files

- **Keys:** snake_case (e.g., `meta_description`, `nav_home`)
- **Organization:** Group by section with comment headers
- **Consistency:** Same keys across all language files

```yaml
# Section Name
key_name: "Translation text"
```

## Tailwind CSS Configuration

Custom colors and fonts are defined in `layouts/partials/head.html`:

| Color     | Hex       | Usage                    |
|-----------|-----------|--------------------------|
| sage      | #E9EFE9   | Soft background          |
| cream     | #FFFEFA   | Main background          |
| leaf      | #8FA876   | Success/nature accents   |
| rose      | #E06C75   | Primary accent (pink)    |
| lavender  | #9D8EB5   | Secondary accent         |
| marigold  | #F2A93B   | Highlight/CTA            |
| clay      | #D97757   | Warm accent              |
| ocean     | #4A6FA5   | Cool contrast            |

| Font      | Class       | Usage                    |
|-----------|-------------|--------------------------|
| Montserrat| `font-sans` | Body text (default)      |
| Caveat    | `font-hand` | Headings, decorative     |

## Common Patterns

### Adding a New Section

1. Create partial in `layouts/partials/newsection.html`
2. Include in `layouts/index.html`: `{{- partial "newsection.html" . -}}`
3. Add translation keys to all `i18n/*.yaml` files
4. Add navigation link in `layouts/partials/nav.html`

### Adding Translations

1. Add key to `i18n/en.yaml` (English first)
2. Add same key to `i18n/es.yaml` (Spanish)
3. Add same key to `i18n/ca.yaml` (Catalan)
4. Use in template: `{{ i18n "new_key" }}`

### Card Component Pattern

```html
<div class="card-shadow bg-white p-8 rounded-3xl group hover:border-color/30 transition-all duration-300">
    <div class="w-12 h-12 bg-color/20 text-color rounded-2xl flex items-center justify-center mb-6 group-hover:rotate-12 transition-transform">
        <i class="fa-solid fa-icon text-xl"></i>
    </div>
    <h3 class="font-hand text-3xl mb-4 text-gray-800">{{ i18n "title" }}</h3>
    <p class="text-gray-600 font-sans">{{ i18n "description" }}</p>
</div>
```

## Important Notes

- **Base URL:** Site is deployed at `/wedding/` subdirectory - use `relURL` for assets
- **No tests:** This is a simple static site with no test framework
- **No linting:** No ESLint/Prettier configured
- **Hermit:** Use `./bin/hugo` or activate environment for correct Hugo version
- **Fingerprinting:** CSS/JS files get content hashes for cache busting
- **Mobile-first:** Always consider responsive design with Tailwind breakpoints (md:, lg:)
