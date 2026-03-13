# PROJECT KNOWLEDGE BASE

**Generated:** 2026-03-13 (Europe/Berlin)
**Commit:** de62838
**Branch:** main

## OVERVIEW
Smart Airbase / Bas X: deterministic sortie-turnaround planning with replay and branching.
Monorepo split between Go backend simulation/runtime layers, React/Vite frontend UI, and context docs that must stay aligned with behavior.

## STRUCTURE
```text
mono/
├── backend/          # Go simulation engine, services, HTTP/WS transport, desktop runtime
├── frontend/         # React/Vite operator UI
├── docs/context/     # feature-level context dossiers
├── proto/            # protobuf-related definitions/assets
└── tools/            # small utilities and asset helpers
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Core simulation behavior | `backend/simulation/` | State machine, runner, lifecycle, determinism |
| Simulation orchestration / branches | `backend/services/` | `SimulationService`, broadcaster, branch lifecycle |
| HTTP / WebSocket surface | `backend/api/` | Route registration, handlers, event stream transport |
| Backend contract docs | `backend/docs/context/` | REST contract + live-event integration notes |
| Frontend app shell | `frontend/src/app/`, `frontend/src/pages/` | `App.tsx` + `BaseXOps.tsx` compose the UI entry flow |
| Frontend feature work | `frontend/src/features/` | Map, simulation, status, timeline, shared UI |
| Frontend transport layer | `frontend/src/lib/` | `src/lib/api/*` is the only fetch/WebSocket boundary |
| Feature documentation | `docs/context/README.md`, `docs/context/*/` | Update when behavior or terminology changes |

## DOMAIN LANGUAGE
- Use repo glossary terms exactly: **Base**, **Runway**, **Pad**, **Aircraft**, **Resource**, **Task**, **Event**, **Simulation Run**, **Seed**, **Timeline**, **Plan**, **Branch**, **Critical Path**, **Bottleneck**.
- Prefer “landing -> next sortie time” framing when discussing outcomes.
- “Replay” means reconstruction from immutable event history.
- “Branch” means an alternate run derived from prior state/history, not an ad hoc UI snapshot.

## CONVENTIONS
- Determinism is the governing rule: same inputs + seed => same outputs.
- Frontend must consume backend behavior through `frontend/src/lib/api/*`; no direct transport code in feature components.
- Replay and branching features must remain compatible with immutable event history and reconstructed state.
- `docs/context/*` is feature documentation; `backend/docs/context/*` is backend/API contract documentation.
- There is no root workspace script surface; run commands from `backend/` or `frontend/`.
- Useful local guidance exists in `backend/`, `backend/api/`, `backend/services/`, `backend/simulation/`, `backend/docs/context/`, `frontend/`, `frontend/src/features/`, and `frontend/src/lib/`.

## ANTI-PATTERNS (THIS PROJECT)
- Unseeded randomness or wall-clock-driven simulation decisions.
- Hidden side effects during simulation advancement or replay reconstruction.
- Frontend reimplementing simulation logic instead of using backend outputs.
- UI components calling `fetch` or `WebSocket` directly.
- API / realtime contract changes without matching doc updates.
- Replay views driven by mutable shortcuts instead of reconstructed state.

## UNIQUE STYLES
- Backend is split cleanly into engine (`simulation`), orchestration (`services`), and transport (`api`).
- Frontend is feature-first (`src/features/*`) with transport kept in `src/lib/api/*`.
- Context docs are short, implementation-oriented dossiers; feature folders live under `docs/context/<feature>/`.

## COMMANDS
```bash
cd backend && go test ./...
cd backend && go run .
cd backend && task dev
cd frontend && pnpm dev
cd frontend && pnpm build
cd frontend && pnpm test
cd frontend && pnpm lint
cd frontend && pnpm typecheck
```

## NOTES
- Default AGENTS depth here is capped at 3 from repo root.
- `backend/cmd/basex/app/` is a real hotspot, but it is covered from `backend/` because it sits deeper than the current cap.
- `docs/context/` is governed well by its own `README.md`; child AGENTS there would be redundant at current scale.
