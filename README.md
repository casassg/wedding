# Laura & Gerard Wedding Website

A multilingual wedding website built with Hugo, Tailwind CSS, and vanilla JavaScript.

**Live site:** [lauraygerard.wedding](https://lauraygerard.wedding/)

# Project Overview

```
assets/      # Hugo Pipes CSS/JS managed via Hugo Pipes
content/     # Markdown landing pages per language
data/{en,es,ca}/ # Structured FAQ + travel info
i18n/*.yaml  # UI strings per locale
layouts/     # Hugo templates/partials
backend/     # Go RSVP API + Google Sheets sync
```

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

## Frontend Commands

```bash
hugo server        # Development server with hot reload
hugo server -D     # Include drafts
hugo --gc --minify # Production build (outputs to public/)
```

## Backend API

The RSVP backend is a Go service that syncs with Google Sheets and runs on Fly.io.

```bash
cd backend
cp .env.example .env        # configure DB path + Google credentials
./dev.sh                    # start local API on http://localhost:8081

# Manual entry points
go run cmd/server/main.go serve   # run API
go run cmd/server/main.go sync    # one-off Google Sheets sync
go run cmd/server/main.go inspect # print sheet schema

# Tests & formatting
go test ./...                             # full suite
go test -run TestUpsertInvitePreservesSyncedAt ./internal/db
go fmt ./...
```

The local database lives at `backend/tmp/wedding.db`. Delete it if you need a fresh state. Google sync requires `GOOGLE_SHEET_ID` plus credentials configured in `.env`.

## Deployment

- Frontend: pushes to `main` run `hugo --gc --minify --baseURL "$BASE_URL/"` in GitHub Actions and publish to GitHub Pages automatically.
- Backend: deploy from `backend/` with `flyctl deploy --ha=false`; ensure the LiteFS volume and required secrets (`GOOGLE_SHEET_ID`, `GOOGLE_SHEETS_CREDENTIALS` or `GOOGLE_APPLICATION_CREDENTIALS`) are set first.

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
