# AGENTS.md – Operating Manual for AI Assistants

Multilingual wedding experience built with Hugo (frontend) plus a Go RSVP API deployed on Fly.io. The site is live at [lauraygerard.wedding](https://lauraygerard.wedding/).

This repository hosts a Hugo-based multilingual wedding site plus a Go RSVP backend. Follow the guardrails below whenever you act in this project.

## Repo Layout

```
assets/      # Hugo Pipes CSS/JS
content/     # Markdown per language (home + traveling)
data/{en,es,ca}/
layouts/     # Hugo templates/partials
backend/     # Go RSVP API + Google Sheets sync
```

## Prerequisites

1. Install [Hermit](https://cashapp.github.io/hermit/):
   ```bash
   curl -fsSL https://github.com/cashapp/hermit/releases/download/stable/install.sh | /bin/bash
   ```
2. Clone and activate the toolchain:
   ```bash
   git clone https://github.com/casassg/wedding.git
   cd wedding
   source bin/activate-hermit
   ```

Hermit provides pinned versions of `hugo`, `go`, `sqlc`, `gh`, `flyctl`, and `gcloud` under `./bin/`.

## Quick Start

### Frontend (Hugo)

```bash
hugo server          # http://localhost:1313/
hugo --gc --minify   # production build into public/
```

Editing tips:
- Update copy in `content/_index.<lang>.md` or `content/traveling/` for long-form sections.
- Add structured items (FAQ, Copán, Honduras, travel tips) in `data/{en,es,ca}/*.yaml`, keeping keys consistent across languages.
- All UI strings live in `i18n/{en,es,ca}.yaml`; never hardcode text in templates.

### Backend (Go RSVP API)

```
cd backend
cp .env.example .env        # configure DB path + Google credentials
./dev.sh                    # launches server on http://localhost:8081

# Manual entry points
go run cmd/server/main.go serve   # start API
go run cmd/server/main.go sync    # one-off Google Sheets sync
go run cmd/server/main.go inspect # inspect sheet structure
```

Testing & quality:

```bash
cd backend
go test ./...                             # full suite
go test -run TestUpsertInvitePreservesSyncedAt ./internal/db
go fmt ./...
```

The backend stores data in SQLite (`backend/tmp/wedding.db` locally, `/data/wedding.db` on Fly.io) and syncs bidirectionally with a Google Sheet when `GOOGLE_SHEET_ID` plus credentials are configured.

## Deployment

- **Frontend:** GitHub Actions (`.github/workflows/deploy.yml`) runs `hugo --gc --minify --baseURL "$BASE_URL/"` on every push to `main` and publishes to GitHub Pages.
- **Backend:** From `backend/`, run `flyctl deploy --ha=false`. Ensure LiteFS volume + secrets (`GOOGLE_SHEET_ID`, `GOOGLE_SHEETS_CREDENTIALS` or `GOOGLE_APPLICATION_CREDENTIALS`) exist first.

## Content Management Cheatsheet

- **FAQ:** Edit `data/{en,es,ca}/faq.yaml`.
- **Copán Places:** `data/{lang}/copan_places.yaml` (id, title, description, icon, maps_url).
- **Honduras Places:** `data/{lang}/honduras_places.yaml` (id, title, description, icon, wiki_url).
- **Travel Tips:** `data/{lang}/travel_tips.yaml`.

Remember to update all three languages before committing so multilingual builds succeed.

## Contributing Workflow

1. Create a feature branch.
2. Activate Hermit and make changes.
3. Run the relevant checks:
   - Frontend: `hugo --gc --minify`.
   - Backend: `cd backend && go fmt ./... && go test ./...`.
4. Document any new commands or conventions inside `AGENTS.md` if they affect other contributors.
5. Open a PR; deployment is handled automatically once `main` is updated.

Questions about the backend internals? Consult the sections below; this document is the single source of truth.

## 2. Build, Run, and Deploy Commands
### Frontend (Hugo)
- `hugo server` → dev server at http://localhost:1313/wedding/ with hot reload.
- `hugo server -D` → include drafts when previewing changes.
- `hugo --gc --minify` → production build into `public/` (garbage-collect unused resources).
- `rm -rf public/ && hugo --gc --minify` → clean rebuild when fingerprints feel stale.

### Backend (Go API)
- `cd backend && ./dev.sh` → local server with `.env` auto-loaded and DB at `backend/tmp/wedding.db`.
- `cd backend && go run cmd/server/main.go serve` → manual execution honoring env vars; same binary exposes `inspect` and `sync` subcommands.
- `cd backend && go build -o server ./cmd/server` → compile production binary; run with `./server serve`.
- `cd backend && ../bin/sqlc generate` → regenerate typed query layer after editing `internal/db/queries.sql`.
- `cd backend && flyctl deploy --ha=false` → push Docker image + config to Fly.io (ensure secrets set first).
- `cd backend && flyctl logs` → stream production logs; add `--region` to scope if necessary.

### Running Tests (backend only)
- `cd backend && go test ./...` → entire suite.
- `cd backend && go test -v ./internal/db` → verbose tests for DB layer.
- `cd backend && go test -run TestUpsertInvitePreservesSyncedAt ./internal/db` → single test focus (preferred way to debug repros).
- `cd backend && go test -cover ./...` → quick coverage snapshot; add `-coverprofile=coverage.out` then `go tool cover -html=coverage.out` for drill-down.

### Quality & Static Checks
- Formatting is handled by `cd backend && go fmt ./...` or `gofmt -w .`; never leave Go files unformatted.
- `cd backend && go vet ./...` before committing to catch suspicious constructs.
- Dependency hygiene: `cd backend && go mod tidy && go mod verify` after touching modules.
- Frontend relies on Hugo template validation; a failing `hugo --gc --minify` run usually means malformed shortcodes/partials.

### Environment Bootstrapping
- Activate Hermit (`source bin/activate-hermit`) so `hugo`, `go`, `sqlc`, `gh`, `flyctl`, and `gcloud` resolve to project-pinned versions.
- `.env.example` under `backend/` shows required variables; copy to `.env` for local work. Never commit secrets.
- Google Sheets sync expects `GOOGLE_SHEET_ID` plus either `GOOGLE_APPLICATION_CREDENTIALS=./credentials.json` or `GOOGLE_SHEETS_CREDENTIALS='...json...'` in `.env`.
- Local DB lives at `backend/tmp/wedding.db`; delete it when schema edits demand a fresh start.

## 3. Deployment Notes
1. Frontend deploys automatically via `.github/workflows/deploy.yml` when `main` receives a push; the job runs `hugo --gc --minify --baseURL "$BASE_URL/"` and uploads `public/` to GitHub Pages.
2. Backend deploys through Fly.io (see `backend/fly.toml`). Only run `flyctl deploy` from `backend/` after confirming LiteFS volume + secrets exist.
3. Health endpoint (`/health`) drives Fly checks; keep it fast and dependency-free.

## 4. Repository Landmarks
- `assets/` → Hugo Pipes-managed CSS/JS (vanilla, no bundler). Fingerprint outputs before referencing in templates.
- `layouts/` → Hugo templates. `layouts/index.html` assembles partials such as `hero.html`, `travel.html`, and `faq.html`.
- `content/` → Markdown content per language for the landing page plus `traveling/` detail pages.
- `data/{en,es,ca}/` → YAML-driven sections (FAQ, Copán/Honduras places, travel tips). Keep keys consistent across locales.
- `i18n/*.yaml` → UI string dictionaries; always update all three languages when adding keys.
- `backend/internal/api/` → HTTP handlers, router, middleware, and DTOs.
- `backend/internal/db/` → SQLite store wrapper, repository helpers, sqlc artifacts, and queries.
- `backend/internal/sheets/` → Google Sheets client + Syncer orchestrating bidirectional sync.
- `backend/cmd/server/` → CLI entrypoints (`serve`, `inspect`, `sync`).
- `backend/scripts/migrate.sh` → invoked by LiteFS before the app boots to ensure schema exists.

## 5. Code Style Guidelines
### Go (backend)
1. **Imports**: Group stdlib, then third-party, then internal packages separated by blank lines; rely on `goimports` behavior.
2. **Formatting**: Always run `go fmt`; use tabs for indentation and keep lines idiomatic (no manual alignment).
3. **Types & Naming**: Exported structs/functions/fields use PascalCase; unexported items use camelCase. SQL columns stay `snake_case` because sqlc expects that mapping.
4. **Error Handling**: Return wrapped errors using `fmt.Errorf("context: %w", err)` inside `backend/internal/db`; API layer should respond with user-friendly messages via `respondError`.
5. **Context**: DB calls instantiate `context.Background()` today; if you introduce request-scoped context, thread it consistently.
6. **Concurrency**: SQLite is single-writer; `db.New` pins `SetMaxOpenConns(1)`. Avoid spinning up goroutines that share the same `*sql.DB` unless you understand LiteFS constraints.
7. **Tests**: Use `testing` + `testify/require`; create temp DBs via `os.CreateTemp` like `internal/db/db_test.go` shows. Clean up files and close handles to avoid leaks.
8. **Logging**: Stick to the stdlib `log` package. Structured logs are not wired up; keep messages concise and informative.
9. **HTTP Semantics**: `internal/api/router.go` delegates to handlers based on path suffix; maintain `GET` for fetch, `POST` for RSVP updates, and use `OPTIONS` for preflight only.

### Hugo Templates
1. Use 4-space indentation; leverage `{{- ... -}}` to trim whitespace when nested loops produce gaps.
2. Never hardcode strings visible to guests—fetch from `i18n` or `data` files instead.
3. Partial naming should reflect purpose (`details.html`, `spain.html`); pass the root context (`.`) unless a narrower scope is required.
4. Asset handling: reference fingerprinted resources via `resources.Get ... | fingerprint` as done in `layouts/partials/head.html`.
5. When adding sections, update `layouts/index.html` order and the navigation partial (`nav.html`) if the section is linkable.

### CSS (`assets/css/custom.css`)
1. Indent with 4 spaces; keep selectors descriptive (e.g., `.glass-nav`, `.card-shadow`).
2. Group related rules under comment headings (Base, Hero, Cards, etc.).
3. Use keyframes for bespoke motion; prefer CSS variables defined through Tailwind config if shared between CSS and templates.
4. Avoid introducing new fonts; `Montserrat` and `Caveat` already load via Google Fonts.

### JavaScript (`assets/js/main.js`)
1. Stick to the existing IIFE + `'use strict'` pattern.
2. Keep functions named and scoped; initialize them inside the DOMContentLoaded callback unless they must run earlier (language detection does).
3. No frameworks or build steps—only browser APIs. Optional chaining, template literals, and other ES6+ features are fine.
4. Guard DOM queries (`if (!element) return;`) to avoid runtime errors on pages that omit a section.
5. Store one-off state in closures (e.g., `isHeartMode`) rather than global variables.

### Content & Data Files
1. Maintain identical key sets across `data/en/*.yaml`, `data/es/*.yaml`, and `data/ca/*.yaml`; missing translations break language builds.
2. YAML lists use `- key: value` style; keep indentation at two spaces.
3. Markdown in values is acceptable (Hugo runs `markdownify`), but avoid raw HTML unless necessary.
4. For large copy edits, favor `content/_index.<lang>.md` so translators can work in Markdown instead of templates.

### Config & Workflow
1. `config.toml` defines language metadata and site params; keep new params nested under `[languages.<lang>.params]` when language-specific.
2. GitHub Actions workflow `.github/workflows/deploy.yml` assumes Hugo cache under `${{ runner.temp }}/hugo_cache`; do not relocate without updating the workflow.
3. No ESLint/Prettier or golangci-lint is configured; if you introduce new tooling, document invocation commands here.

## 6. Collaboration Etiquette
1. Never commit `.env`, `backend/credentials.json`, or other secrets—`.gitignore` already blocks them; verify before staging.
2. When touching multiple areas, explain the coupling in PR descriptions so future agents understand why both frontend and backend changed.
3. Prefer small, reviewable diffs. Update this AGENTS guide whenever you change canonical commands or style rules.
4. After significant backend edits, run `go test ./...` plus at least one targeted `go test -run Name ./path` to ensure focused coverage.
5. After frontend template changes, run `hugo server` locally and sanity-check the `/traveling/` page plus all three languages for rendering regressions.

---

Use this document as the single source of truth for agent operations. Keep it roughly in sync with reality: when commands or conventions drift, fix them here immediately.
