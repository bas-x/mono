# BACKEND KNOWLEDGE BASE

## OVERVIEW
Go package root for simulation engine, orchestration services, HTTP/WebSocket transport, and the Ebiten desktop runtime.

## STRUCTURE
```text
backend/
├── simulation/    # deterministic engine + lifecycle/state tests
├── services/      # SimulationService, broadcaster, branch orchestration
├── api/           # Echo handlers, route registration, websocket transport
├── cmd/           # entry points (`main.go`, `cmd/basex`, `cmd/draw`)
├── docs/context/  # backend contract docs
├── prng/          # deterministic random helpers
└── geometry/      # spatial utilities used by simulation/map data
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| HTTP boot / server wiring | `main.go`, `api/server.go` | `api.Run` is the backend bootstrap |
| Route surface | `api/routes.go` | Single registration hub |
| Branch/start/pause/reset behavior | `services/simulation.go` | Service-layer orchestration boundary |
| Core aircraft/task behavior | `simulation/` | Most domain invariants live here |
| Deterministic helpers | `prng/`, `geometry/` | Utility packages used by engine logic |
| Backend contract docs | `docs/context/api.md`, `docs/context/frontend-live-events.md` | Must stay synced with shipped behavior |

## CONVENTIONS
- Treat `services.SimulationService` as the orchestration boundary around `simulation`.
- Keep transport concerns in `api/`; keep domain/state-machine logic out of handlers.
- `cmd/basex` is the desktop runtime entry; `main.go` is the HTTP server entry.
- When API or event behavior changes, update `backend/docs/context/*` as part of the same change.

## ANTI-PATTERNS
- Calling into deep simulation internals from handlers when a service boundary exists.
- Changing event payloads or route behavior without corresponding tests and doc updates.
- Introducing non-deterministic behavior into `simulation/` through backend convenience code.
- Treating `cmd/` runtime code as the source of truth for domain rules.

## COMMANDS
```bash
go test ./...
go test ./api ./services ./simulation
go run .
task dev
```

## NOTES
- `backend/simulation`, `backend/services`, `backend/api`, and `backend/docs/context` each have local AGENTS because they carry distinct constraints.
- `backend/cmd/basex/app/` is complex but deeper than the current AGENTS depth cap; inspect its files directly when working there.
