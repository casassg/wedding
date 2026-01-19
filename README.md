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

## Content Management

### FAQ & Places

FAQ questions/answers and places (Copán, Honduras) are stored in the `data/` folder for easy editing:

```
data/
├── en/                    # English
│   ├── faq.yaml          # FAQ questions/answers
│   ├── copan_places.yaml # Things to do in Copán
│   └── honduras_places.yaml # Explore Honduras
├── es/                    # Spanish
└── ca/                    # Catalan
```

**To add/edit FAQ items:** Edit `data/{en,es,ca}/faq.yaml`

```yaml
- question: "Your question?"
  answer: "Answer with [markdown links](https://example.com) supported."
```

**To add/edit Copán places:** Edit `data/{en,es,ca}/copan_places.yaml`

```yaml
- id: unique_id
  title: "Place Name"
  description: "Description with markdown support."
  icon: "fa-solid fa-icon-name"
  maps_url: "https://maps.app.goo.gl/..."
```

**To add/edit Honduras places:** Edit `data/{en,es,ca}/honduras_places.yaml`

```yaml
- id: unique_id
  title: "Place Name"
  description: "Description with markdown support."
  icon: "fa-solid fa-icon-name"
  wiki_url: "https://en.wikipedia.org/wiki/..."
```

**Note:** Remember to update all 3 language files when making changes.
