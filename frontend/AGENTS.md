# FRONTEND KNOWLEDGE BASE

## OVERVIEW
React + TypeScript + Vite operator UI. Feature-first structure with transport isolated behind `src/lib/api/*`.

## STRUCTURE
```text
frontend/
├── src/app/       # app root
├── src/pages/     # page-level composition
├── src/features/  # map, simulation, status, timeline, shared UI
├── src/lib/       # API/config abstraction boundary
└── src/assets/    # static frontend assets
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| App bootstrap | `src/main.tsx`, `src/app/App.tsx` | `ApiProvider` wraps the app |
| Main page composition | `src/pages/BaseXOps.tsx` | Base operations shell |
| Feature work | `src/features/` | UI/domain feature components and hooks |
| Transport/config work | `src/lib/` | `src/lib/api/*` only |
| Env / scripts | `package.json`, `README.md`, `vite.config.ts` | Command and alias surface |

## CONVENTIONS
- `src/lib/api/*` is the only place that implements fetch or WebSocket.
- UI should consume `useApi`, API clients, and feature hooks instead of raw transport calls.
- `ApiProvider` owns remote/mock mode wiring.
- Feature composition favors panels and reusable shells over inline one-off logic.
- Keep accessibility affordances intact in page/shell components.

## ANTI-PATTERNS
- Direct `fetch` / `WebSocket` usage in pages, panels, or feature components.
- Reimplementing backend simulation rules in the UI.
- Mixing transport configuration into presentation components.
- Breaking replay/scrub behavior by treating viewed state as mutable truth.

## COMMANDS
```bash
pnpm dev
pnpm dev:full
pnpm build
pnpm test
pnpm lint
pnpm typecheck
```

## NOTES
- There is no repo-root workspace wrapper here; run frontend commands from `frontend/`.
- Local guidance continues in `frontend/src/features/` and `frontend/src/lib/`.
