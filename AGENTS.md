# AGENTS.md

## Layout
- This repo has two runtimes: Go API at repo root and React app in `frontend/`.
- Backend entrypoint: `cmd/api/main.go`; route wiring: `internal/server/server.go` and `internal/server/routes.go`; DB layer: `internal/database/database.go`.
- Frontend entrypoint: `frontend/src/main.tsx`; main screen/data flow: `frontend/src/App.tsx`.
- Go module path is `template` (`go.mod`), so internal imports use `template/internal/...`.

## Trust executable config over README
- `README.md` is stale (mentions old PQC/WebSocket architecture); use `Makefile`, `go.mod`, backend code, and `frontend/package.json`/`frontend/vite.config.ts` as source of truth.
- Frontend-specific agent rules already exist in `frontend/AGENTS.md`.

## Backend workflow (run from repo root)
- `make run` starts the API using `.env` values (default `PORT=8080`).
- `make test` runs `go test ./... -v`.
- Single test example: `go test ./internal/server -run TestHelloWorldHandler -v`.
- `make watch` uses `air`; build command is `make build` and output binary is `./main`.
- SQLite path comes from `BLUEPRINT_DB_URL`; default `./test.db` is relative to repo root.

## Frontend workflow (run from `frontend/`)
- Use Vite+ (`vp`) commands per `frontend/AGENTS.md`; avoid direct npm/pnpm/yarn commands for normal workflow.
- Install deps: `vp install`.
- Dev server: `vp dev` (proxies `/api` to `http://localhost:8080`).
- Validate: `vp check`.
- Build with typecheck step: `vp run build` (this script runs `tsc -b && vp build`; plain `vp build` skips the script's `tsc -b`).

## Integration gotchas
- Backend CORS allowlist is hardcoded to `http://localhost:5173` in `internal/server/routes.go`; if frontend runs on another port, update CORS.
- API startup auto-creates the `entries` table in SQLite (`internal/database/database.go`); there is no separate migration command.

## Verification order
- Backend-only edits: `make test`.
- Frontend-only edits: `vp check && vp run build` (from `frontend/`).
- API contract/full-stack edits: run both checks above.
