# FRONTEND LIB KNOWLEDGE BASE

## OVERVIEW
Frontend infrastructure boundary. At current depth cap this file covers `src/lib/api/*`, the only approved transport implementation layer.

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Public export surface | `index.ts`, `api/index.ts` | Re-export boundary used by the app |
| Provider / context | `api/useApi.tsx` | `ApiProvider`, `useApi`, mode switching |
| Client assembly | `api/clients.ts`, `api/services/*` | HTTP service clients |
| Realtime | `api/realtime/*`, `api/useSimulationStream.ts` | WebSocket wrapper + stream hooks |
| Mock parity | `api/mock/*` | Remote/mock behavior should stay aligned |
| Shared types/config | `api/types.ts`, `api/config.ts` | Contract-facing TypeScript layer |

## CONVENTIONS
- Keep all fetch/WebSocket implementation details inside `src/lib/api/*`.
- `ApiProvider` centralizes config parsing, mode selection, and client creation.
- Shared types are the contract boundary between transport code and features.
- Mock, realtime, and service clients should remain behaviorally aligned where interfaces overlap.
- Use `@/lib/api` exports from the rest of the app instead of deep ad hoc imports when possible.

## ANTI-PATTERNS
- Scattering config parsing or client construction into feature code.
- Letting mock and remote clients drift from shared types or expected semantics.
- Exposing raw transport primitives as the public UI-facing API.
- Updating endpoint behavior without matching type and doc changes.

## COMMANDS
```bash
pnpm test
pnpm typecheck
pnpm lint
```

## NOTES
- If backend contract changes, cross-check `backend/docs/context/api.md` and `backend/docs/context/frontend-live-events.md` while updating this layer.
