# BACKEND CONTEXT DOCS KNOWLEDGE BASE

## OVERVIEW
Backend-owned contract docs for shipped HTTP, WebSocket, and frontend-integration behavior.

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| REST / WS contract | `api.md` | Paths, request shapes, response shapes, branch limits |
| Frontend event integration | `frontend-live-events.md` | Live flow, event payloads, reducer expectations |

## CONVENTIONS
- `api.md` documents actual backend contract, not aspirational design.
- `frontend-live-events.md` documents the currently supported hydration + websocket flow.
- Keep docs short, concrete, and version-accurate.
- If API/service behavior changes, update these docs in the same change.

## ANTI-PATTERNS
- Shipping contract changes without doc updates.
- Describing unimplemented workflows as if they are supported.
- Letting event examples drift from real payload fields or branch constraints.

## COMMANDS
```bash
go test ./api ./services
```

## NOTES
- This directory is distinct from `docs/context/*`: these files are backend contract references, not cross-feature planning dossiers.
- Treat examples here as shipped behavior snapshots, not loose sketches.
