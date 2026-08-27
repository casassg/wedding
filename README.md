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

1. Install [Hermit](https://cashapp.github.io/hermit/).
2. Clone repo `git clone https://github.com/casassg/wedding.git && cd wedding`
3. Activate hermit `source ./bin/activate-hermit`
4. Configure environment `cd backend && cp .env.example .env && cd ..`
5. Start server `./dev.sh`

The local database lives at `backend/tmp/wedding.db`. Delete it if you need a fresh state. Google sync requires `GOOGLE_SHEET_ID` plus credentials configured in `.env`.

## Deployment

Automatic via [deploy.yml](.github/workflows/deploy.yml). Frontend is served from `main` as GitHub Pages, backend is served on [fly.io](https://fly.io).

Ensure the required secrets are set on your fly.io deployment (`GOOGLE_SHEET_ID`, `GOOGLE_SHEETS_CREDENTIALS` or `GOOGLE_APPLICATION_CREDENTIALS`).
**Note:** Remember to update all 3 language files when making changes.
