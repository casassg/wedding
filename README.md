# Laura & Gerard Wedding Website

A multilingual wedding website built with Hugo, Tailwind CSS, and vanilla JavaScript.

**Live site:** [gerard.space/wedding](https://gerard.space/wedding/)

## Quick Start

### Prerequisites

Install [Hermit](https://cashapp.github.io/hermit/) (manages Hugo and GitHub CLI):

```bash
curl -fsSL https://github.com/cashapp/hermit/releases/download/stable/install.sh | /bin/bash
```

### Setup

```bash
# Clone the repo
git clone https://github.com/casassg/wedding.git
cd wedding

# Activate Hermit (installs Hugo automatically)
source bin/activate-hermit

# Start development server
hugo server
```

Visit [http://localhost:1313/wedding/](http://localhost:1313/wedding/)

## Commands

```bash
hugo server        # Development server with hot reload
hugo server -D     # Include drafts
hugo --gc --minify # Production build (outputs to public/)
```

## Deployment

Pushes to `main` automatically deploy to GitHub Pages via GitHub Actions.

## Languages

- English (default)
- Spanish (`/es/`)
- Catalan (`/ca/`)
