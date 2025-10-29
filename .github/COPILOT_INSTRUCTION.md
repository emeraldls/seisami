## 1. Project Summary

## 1. What This Project Is

Seisami is a voice-first task manager delivered as:
- **Desktop app** (`/app`): Go + Wails shell with a React 19 frontend, SQLite storage, PortAudio capture, GPT-powered automation.
- **Optional server** (`/server`): Go HTTP + TCP service providing auth, PostgreSQL persistence, and cross-device sync.
- **Marketing site** (`/web`): TanStack Start app used for landing page and future web client work.

Default posture is local-first. Cloud sync and the central server are opt-in.

## 2. Tech Snapshot

- **Go 1.25+** (Wails, pgx, SQLC) for desktop and server backends.
- **React 19 + TypeScript + Tailwind** inside Wails and the marketing site.
- **SQLite (desktop) / PostgreSQL (server)** for storage, both driven through SQLC.
- **OpenAI GPT-4.1** for transcription + intent. Local Whisper integration is planned but not finished.

## 3. Files That Matter Most

- `app/main.go`, `app/app.go`: Wails entry and exposed Go API.
- `app/internal/repo/`: SQLite access layer; edit `schema.sql` then run `sqlc generate`.
- `app/frontend/src/`: React UI (views, Zustand stores, shadcn components).
- `server/main.go`: Starts TCP sync server and HTTP auth.
- `server/central/`: Auth handlers and configuration.
- `server/sqlc/`: Postgres schema + SQLC config; regenerate with `sqlc generate`.
- `web/src/`: Marketing site routes and components.

## 4. Day-One Commands

```bash
# Desktop app
cd app
go mod download
npm install --prefix frontend
wails dev             # hot reload desktop client
wails build           # produces binaries in build/bin/

# Server (requires DATABASE_URL, JWT_SECRET)
cd server
go run main.go

# Marketing site
cd web
npm install
npm run dev
```

## 5. Working Guidelines

- Keep desktop logic split: Go handles IO/AI, React manages UI via generated `wailsjs` bindings.
- When exposing new Go features, add a method on `App` (in `app/app.go`), update the repo layer if storage is involved, then sync the generated TypeScript bindings.
- Prefer local data paths and design for offline; treat server interaction as additive, not required.
- Collaboration messages live in `app/types/types.go` (desktop) and `server/types/types.go`; keep payloads small and JSON serializable.
- Follow existing naming: kebab-case filenames in TS, PascalCase components, snake_case in Go when interoperating with native code.
- Never hardcode secrets. Use `.env` or OS env vars.

## 6. Common Playbooks

- **Change the SQLite schema**: edit `app/internal/repo/sqlc/schema.sql` → run `sqlc generate` → update `interfaces.go` and `repo.go` → expose via `App`.
- **Extend GPT tooling**: add tool definition in `app/internal/tools/tools.go`, wire logic in `actions/actions.go`, emit board change events so the UI updates.
- **Add a new board interaction in UI**: create component in `app/frontend/src/components/`, hook into Zustand store (`board-store.ts`), call Go bridge via `wailsjs/go/main/App`.
- **Server-side feature**: edit `server/sqlc/schema.sql` → `sqlc generate` → implement handler in `server/central` or TCP flow in `server/room*` packages.

## 7. Guardrails

- Maintain the event-driven flow: Go emits Wails events, TS stores react; do not bypass with ad-hoc globals.
- Handle async errors explicitly (audio, TCP, AI). Broken streams must fail gracefully.
- Keep cross-platform code clean. macOS Objective-C lives in `app/include`; ensure Windows paths remain untouched unless intended.
- Tests are light today; if you add risky logic, add unit coverage in Go or component tests in TS.

## 8. Reference URLs

- Product updates: https://seisami.com
- Issue tracker: https://github.com/emeraldls/seisami/issues