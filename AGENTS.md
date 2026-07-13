# AGENTS.md

Multilingual Hugo wedding site with a Go RSVP API, SQLite storage, Google Sheets sync, GitHub Pages, and Fly.io.

## Layout

- `assets/`, Hugo Pipes CSS and JavaScript.
- `content/`, Markdown pages per language.
- `data/{en,es,ca}/`, structured localized content.
- `i18n/`, UI strings for English, Spanish, and Catalan.
- `layouts/`, Hugo templates and partials.
- `backend/`, Go API, SQLite store, and Sheets sync.

## Tooling

Hermit provides pinned tools under `./bin/`. Run them directly, activation is not required.

```bash
./bin/hugo version
./bin/go version
./bin/sqlc version
```

Never commit `.env`, credentials, tokens, or generated secrets.

## Frontend

```bash
./bin/hugo server
./bin/hugo server -D
./bin/hugo --gc --minify
./bin/uv run scripts/check-i18n.py
./bin/node --check assets/js/main.js
```

- Put guest-visible UI strings in all three `i18n/*.yaml` files.
- Keep data keys aligned across `data/en`, `data/es`, and `data/ca`.
- Use 4-space template indentation and Hugo whitespace trimming where useful.
- Use fingerprinted resources through Hugo Pipes.
- Keep JavaScript in the existing IIFE and guard optional DOM elements.
- Keep CSS grouped by component in `assets/css/custom.css`.

## Backend

Copy `backend/.env.example` to `backend/.env` for local configuration.

```bash
./bin/go -C backend run ./cmd/server serve
./bin/go -C backend run ./cmd/server sync
./bin/go -C backend run ./cmd/server inspect
./bin/go -C backend test ./...
./bin/go -C backend vet ./...
./bin/go -C backend fmt ./...
./bin/go -C backend mod tidy
./bin/go -C backend mod verify
```

After editing `backend/internal/store/queries.sql`, regenerate sqlc output:

```bash
(cd backend && ../bin/sqlc generate)
```

- Use stdlib errors or wrap with `%w`; API errors must remain user-friendly.
- Thread request contexts through database calls.
- SQLite is single-writer; use the existing store transaction helpers.
- Use `testing` and `testify/require`, with schema-backed temporary databases.
- Keep HTTP methods and routes consistent with `internal/api/router.go`.
- Do not manually edit sqlc-generated files.

## Verification

Run the narrowest relevant checks, then the full checks before committing substantial changes:

```bash
./bin/go -C backend test ./...
./bin/go -C backend vet ./...
./bin/hugo --gc --minify
./bin/node --check assets/js/main.js
./bin/uv run scripts/check-i18n.py
```

## Deployment

- `.github/workflows/pr.yaml` validates PRs and manages Fly review apps.
- `.github/workflows/deploy.yml` deploys `main` to GitHub Pages and Fly.io.
- Run Fly commands with the pinned binary, for example `(cd backend && ../bin/flyctl deploy --ha=false)`.
- The backend requires `GOOGLE_SHEET_ID` and Google credentials for Sheets sync.
- Keep `/health` fast and dependency-free.

## Collaboration

- Prefer the smallest correct diff and reuse existing patterns.
- Do not revert unrelated worktree changes.
- Regenerate generated files from their source instead of editing them.
- Explain cross-frontend/backend coupling in PR descriptions.
- Update this guide when canonical commands or conventions change.
